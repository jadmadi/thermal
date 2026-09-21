package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	testThermalBin  string
	testFixtureHome string
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "thermal-sim-user-*")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testThermalBin = filepath.Join(tmpDir, "thermal")
	cmd := exec.Command("go", "build", "-o", testThermalBin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("failed to build thermal binary: %v\n%s", err, out)
	}

	// Setup a clean fixture home with mock session data so simulated tests
	// run deterministically across all environments (including blank CI).
	testFixtureHome = filepath.Join(tmpDir, "mockhome")
	setupMockHome(testFixtureHome)

	os.Exit(m.Run())
}

func setupMockHome(homeDir string) {
	// Create mock git repo
	mockRepo := filepath.Join(homeDir, "projects", "repo-alpha")
	_ = os.MkdirAll(filepath.Join(mockRepo, ".git"), 0755)

	// Create CodeWhale session
	cwDir := filepath.Join(homeDir, ".codewhale", "sessions")
	_ = os.MkdirAll(cwDir, 0755)
	cwSession := fmt.Sprintf(`{
		"session_id": "sim-sess-1",
		"timestamp": 1726000000000,
		"metadata": {
			"total_tokens": 524000,
			"cost": { "session_cost_usd": 1.25 },
			"model": "claude-3-5-sonnet",
			"mode": "chat",
			"workspace": %q
		}
	}`, mockRepo)
	_ = os.WriteFile(filepath.Join(cwDir, "session_1.json"), []byte(cwSession), 0644)

	// Create Claude project transcript
	claudeDir := filepath.Join(homeDir, ".claude", "projects", "proj1")
	_ = os.MkdirAll(claudeDir, 0755)
	claudeTranscript := fmt.Sprintf(`{"type":"message","timestamp":"2026-09-18T10:00:00Z","cwd":%q,"message":{"usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000},"model":"claude-3-5-sonnet"}}`+"\n", mockRepo)
	_ = os.WriteFile(filepath.Join(claudeDir, "session.jsonl"), []byte(claudeTranscript), 0644)
}

func runSim(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(testThermalBin, args...)
	// If running on host with real tool data, use host environment; otherwise fallback to testFixtureHome
	if _, err := os.Stat(os.ExpandEnv("$HOME/.codewhale")); os.IsNotExist(err) {
		if _, err2 := os.Stat(os.ExpandEnv("$HOME/.gemini/antigravity-cli")); os.IsNotExist(err2) {
			cmd.Env = append(os.Environ(), "HOME="+testFixtureHome)
		}
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

// 1. Output & Formatting Checks: no ANSI when --no-color, clean table headers, compact numbers.
func TestSimulatedUser_LeaderboardFormattingAndLanguage(t *testing.T) {
	stdout, stderr, code := runSim(t, "--no-color")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Errorf("stdout contains ANSI escape sequences with --no-color")
	}
	if !strings.Contains(stdout, "THERMAL") {
		t.Errorf("stdout missing THERMAL banner")
	}
	if !strings.Contains(stdout, "Tool") || !strings.Contains(stdout, "Days") {
		t.Errorf("stdout missing standard table headers (Tool, Days)")
	}

	// Verify language: no duplicate words or NaN
	checkLanguageSanity(t, stdout)
}

// 2. JSON Semantics: Valid JSON, expected fields, no stdout pollution.
func TestSimulatedUser_LeaderboardJSON(t *testing.T) {
	stdout, stderr, code := runSim(t, "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr)
	}
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nOutput:\n%s", err, stdout)
	}
	if _, ok := doc["results"]; !ok {
		t.Errorf("JSON output missing 'results' key")
	}
	if _, ok := doc["generatedAt"]; !ok {
		t.Errorf("JSON output missing 'generatedAt' key")
	}
}

// 3. Reports: Daily, Weekly, Monthly with --chart, --json, --last.
func TestSimulatedUser_Reports(t *testing.T) {
	reports := []string{"daily", "weekly", "monthly"}
	for _, rep := range reports {
		t.Run(rep, func(t *testing.T) {
			// Plain text report with --no-color
			stdout, stderr, code := runSim(t, rep, "--no-color", "--last", "4")
			if code != 0 {
				t.Fatalf("%s failed with exit code %d. stderr: %s", rep, code, stderr)
			}
			if strings.Contains(stdout, "\x1b[") {
				t.Errorf("%s contains ANSI with --no-color", rep)
			}
			checkLanguageSanity(t, stdout)

			// JSON report
			jsonOut, jsonErr, jCode := runSim(t, rep, "--json", "--last", "4")
			if jCode != 0 {
				t.Fatalf("%s --json failed with exit code %d. stderr: %s", rep, jCode, jsonErr)
			}
			var jDoc map[string]interface{}
			if err := json.Unmarshal([]byte(jsonOut), &jDoc); err != nil {
				t.Fatalf("%s --json is not valid JSON: %v", rep, err)
			}
			if _, ok := jDoc["data"]; !ok {
				t.Errorf("%s --json missing 'data' array", rep)
			}
			if _, ok := jDoc["totals"]; !ok {
				t.Errorf("%s --json missing 'totals' object", rep)
			}
		})
	}
}

// 4. Projects and Models: Sorting, breakdown, top limits, JSON schema.
func TestSimulatedUser_ProjectsAndModels(t *testing.T) {
	t.Run("projects", func(t *testing.T) {
		stdout, stderr, code := runSim(t, "projects", "--no-color", "--top", "5")
		if code != 0 {
			t.Fatalf("projects failed with exit code %d. stderr: %s", code, stderr)
		}
		if strings.Contains(stdout, "\x1b[") {
			t.Errorf("projects contains ANSI with --no-color")
		}
		checkLanguageSanity(t, stdout)

		jsonOut, _, jCode := runSim(t, "projects", "--json", "--top", "5")
		if jCode != 0 {
			t.Fatalf("projects --json failed with code %d", jCode)
		}
		var jDoc map[string]interface{}
		if err := json.Unmarshal([]byte(jsonOut), &jDoc); err != nil {
			t.Fatalf("projects --json is not valid JSON: %v", err)
		}
	})

	t.Run("models", func(t *testing.T) {
		stdout, stderr, code := runSim(t, "models", "--no-color", "--sort", "cost")
		if code != 0 {
			t.Fatalf("models failed with exit code %d. stderr: %s", code, stderr)
		}
		checkLanguageSanity(t, stdout)

		jsonOut, _, jCode := runSim(t, "models", "--json")
		if jCode != 0 {
			t.Fatalf("models --json failed with code %d", jCode)
		}
		var jDoc map[string]interface{}
		if err := json.Unmarshal([]byte(jsonOut), &jDoc); err != nil {
			t.Fatalf("models --json is not valid JSON: %v", err)
		}
	})
}

// 5. Analytics: Mix, Stats, Trend.
func TestSimulatedUser_Analytics(t *testing.T) {
	cmds := [][]string{
		{"mix", "--by", "model", "--grain", "day", "--no-color"},
		{"mix", "--metric", "cost", "--no-color"},
		{"stats", "--no-color"},
		{"stats", "--metric", "cost", "--no-color"},
		{"trend", "--no-color"},
		{"trend", "--metric", "cost", "--no-color"},
	}

	for _, cmd := range cmds {
		name := strings.Join(cmd, "_")
		t.Run(name, func(t *testing.T) {
			stdout, stderr, code := runSim(t, cmd...)
			if code != 0 {
				t.Fatalf("command %v failed with code %d: %s", cmd, code, stderr)
			}
			checkLanguageSanity(t, stdout)
		})
	}
}

// 6. Replay: Workload simulation against subscriptions and API pricing.
func TestSimulatedUser_ReplaySimulation(t *testing.T) {
	t.Run("default replay", func(t *testing.T) {
		stdout, stderr, code := runSim(t, "replay", "--no-color")
		if code != 0 {
			t.Fatalf("replay failed with code %d: %s", code, stderr)
		}
		if !strings.Contains(stdout, "Thermal") || !strings.Contains(stdout, "replay") {
			t.Errorf("replay header missing")
		}
		if !strings.Contains(stdout, "Capacity Verdict") || !strings.Contains(stdout, "Cost/Mo") {
			t.Errorf("replay table columns missing (Capacity Verdict, Cost/Mo)")
		}
		checkLanguageSanity(t, stdout)
	})

	t.Run("replay against specific targets", func(t *testing.T) {
		targets := []string{"claude-pro", "deepseek-v3", "chatgpt-plus"}
		for _, target := range targets {
			stdout, stderr, code := runSim(t, "replay", "--against", target, "--no-color")
			if code != 0 {
				t.Fatalf("replay --against %s failed with code %d: %s", target, code, stderr)
			}
			// Verify verdict is one of PASS, DEGRADED, FAIL
			if !strings.Contains(stdout, "PASS") && !strings.Contains(stdout, "DEGRADED") && !strings.Contains(stdout, "FAIL") {
				t.Errorf("replay output missing valid verdict (PASS/DEGRADED/FAIL)")
			}
			checkLanguageSanity(t, stdout)
		}
	})

	t.Run("replay JSON schema", func(t *testing.T) {
		jsonOut, stderr, code := runSim(t, "replay", "--json", "--compare", "all")
		if code != 0 {
			t.Fatalf("replay --json failed with code %d: %s", code, stderr)
		}
		var jDoc map[string]interface{}
		if err := json.Unmarshal([]byte(jsonOut), &jDoc); err != nil {
			t.Fatalf("replay --json is not valid JSON: %v", err)
		}
		if _, ok := jDoc["workload"]; !ok {
			t.Errorf("replay --json missing 'workload'")
		}
		if _, ok := jDoc["plans"]; !ok {
			t.Errorf("replay --json missing 'plans'")
		}

		// Verify throttle rate range [0.0, 1.0] and valid verdicts
		plansRaw, _ := json.Marshal(jDoc["plans"])
		var plans []map[string]interface{}
		_ = json.Unmarshal(plansRaw, &plans)
		for _, plan := range plans {
			verdict, _ := plan["capacityVerdict"].(string)
			if verdict != "PASS" && verdict != "DEGRADED" && verdict != "FAIL" {
				t.Errorf("unexpected verdict %q in plan %v", verdict, plan["id"])
			}
			rate, _ := plan["throttleRate"].(float64)
			if rate < 0.0 || rate > 1.0 {
				t.Errorf("throttleRate %v out of range [0, 1]", rate)
			}
		}
	})
}

// 7. Validation Negative Tests: Incorrect flags produce exit code 1 with clear, descriptive messages.
func TestSimulatedUser_ValidationErrors(t *testing.T) {
	testCases := []struct {
		args        []string
		expectedMsg string
	}{
		{[]string{"weekly", "--top", "5"}, "--top only applies to the projects and models commands"},
		{[]string{"replay", "--by", "model"}, "--metric, --by, and --grain only apply to the trend, mix, and stats commands"},
		{[]string{"projects", "--against", "claude-pro"}, "--against and --compare only apply to the replay command"},
		{[]string{"replay", "--chart"}, "--breakdown and --chart do not apply to the replay command"},
		{[]string{"daily", "--last", "-5"}, "--last cannot be negative"},
	}

	for _, tc := range testCases {
		name := strings.Join(tc.args, "_")
		t.Run(name, func(t *testing.T) {
			stdout, stderr, code := runSim(t, tc.args...)
			combined := stdout + stderr
			if code != 1 {
				t.Errorf("expected exit code 1 for args %v, got %d. Output: %s", tc.args, code, combined)
			}
			if !strings.Contains(combined, tc.expectedMsg) {
				t.Errorf("expected error message containing %q, got: %s", tc.expectedMsg, combined)
			}
		})
	}
}

// Helper: verifies no raw NaN, Inf, or weird formatting artifacts.
func checkLanguageSanity(t *testing.T, output string) {
	t.Helper()
	if strings.Contains(output, "NaN") {
		t.Errorf("output contains 'NaN': %s", output)
	}
	if strings.Contains(output, "+Inf") || strings.Contains(output, "-Inf") {
		t.Errorf("output contains 'Inf': %s", output)
	}
	// Check for floating point serialization flaws like 0.00000000001
	weirdFloatRegex := regexp.MustCompile(`\d+\.\d{8,}`)
	if match := weirdFloatRegex.FindString(output); match != "" {
		t.Errorf("output contains unformatted raw float (%s)", match)
	}
	// Check for common typo double words (e.g. "the the", "to to")
	words := strings.Fields(strings.ToLower(output))
	for i := 0; i < len(words)-1; i++ {
		w1 := strings.Trim(words[i], ",.:;\"'()[]")
		w2 := strings.Trim(words[i+1], ",.:;\"'()[]")
		if w1 != "" && w1 == w2 && (w1 == "the" || w1 == "to" || w1 == "in" || w1 == "of" || w1 == "on" || w1 == "is") {
			t.Errorf("output contains duplicated word (%s %s)", w1, w2)
		}
	}
}
