// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
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

	binName := "thermal"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	testThermalBin = filepath.Join(tmpDir, binName)
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

	now := time.Now()
	recentISO := now.AddDate(0, 0, -1).Format(time.RFC3339)

	// Create CodeWhale session
	cwDir := filepath.Join(homeDir, ".codewhale", "sessions")
	_ = os.MkdirAll(cwDir, 0755)
	cwSession := fmt.Sprintf(`{
		"session_id": "sim-sess-1",
		"metadata": {
			"created_at": %q,
			"updated_at": %q,
			"message_count": 10,
			"total_tokens": 524000,
			"cost": { "session_cost_usd": 1.25 },
			"model": "claude-3-5-sonnet",
			"mode": "chat",
			"workspace": %q
		}
	}`, recentISO, recentISO, mockRepo)
	_ = os.WriteFile(filepath.Join(cwDir, "session_1.json"), []byte(cwSession), 0644)

	// Create Claude project transcript
	claudeDir := filepath.Join(homeDir, ".claude", "projects", "proj1")
	_ = os.MkdirAll(claudeDir, 0755)
	claudeTranscript := fmt.Sprintf(`{"type":"assistant","timestamp":%q,"cwd":%q,"message":{"id":"msg-sim-1","model":"claude-3-5-sonnet","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000}}}`+"\n", recentISO, mockRepo)
	_ = os.WriteFile(filepath.Join(claudeDir, "session.jsonl"), []byte(claudeTranscript), 0644)
}

func runSim(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(testThermalBin, args...)

	// Always filter and explicitly isolate environment so subprocesses see testFixtureHome
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "HOME=") &&
			!strings.HasPrefix(e, "USERPROFILE=") &&
			!strings.HasPrefix(e, "GROK_HOME=") &&
			!strings.HasPrefix(e, "DSH_HOME=") &&
			!strings.HasPrefix(e, "HERMES_HOME=") &&
			!strings.HasPrefix(e, "CODEX_HOME=") &&
			!strings.HasPrefix(e, "OPENCODE_HOME=") &&
			!strings.HasPrefix(e, "NO_COLOR=") &&
			!strings.HasPrefix(e, "CLICOLOR_FORCE=") {
			env = append(env, e)
		}
	}
	env = append(env,
		"HOME="+testFixtureHome,
		"USERPROFILE="+testFixtureHome,
		"GROK_HOME="+filepath.Join(testFixtureHome, ".grok"),
		"DSH_HOME="+filepath.Join(testFixtureHome, ".dsh"),
		"HERMES_HOME="+filepath.Join(testFixtureHome, ".hermes"),
		"NO_COLOR=",
		"CLICOLOR_FORCE=0",
	)
	cmd.Env = env
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

// 7. Validation Negative Tests: incorrect flags use the documented exit code
// and a clear, descriptive message (2 for usage, 1 for retired commands).
func TestSimulatedUser_ValidationErrors(t *testing.T) {
	testCases := []struct {
		args         []string
		expectedMsg  string
		expectedCode int
	}{
		{[]string{"stats", "--top", "5"}, "--top does not apply", 2},
		{[]string{"replay"}, "subscription replay simulation has been retired", 1},
		{[]string{"yield"}, "token yield efficiency reporting has been retired", 1},
		{[]string{"daily", "--last", "-5"}, "--last must be at least 1", 2},
	}

	for _, tc := range testCases {
		name := strings.Join(tc.args, "_")
		t.Run(name, func(t *testing.T) {
			stdout, stderr, code := runSim(t, tc.args...)
			combined := stdout + stderr
			if code != tc.expectedCode {
				t.Errorf("expected exit code %d for args %v, got %d. Output: %s", tc.expectedCode, tc.args, code, combined)
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

func TestSimulatedUser_LicenseCommand(t *testing.T) {
	// 1. Text mode via positional
	stdout, stderr, code := runSim(t, "license")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal license', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "AGPL-3.0") || !strings.Contains(stdout, "Jad Madi") || !strings.Contains(stdout, "Commercial") {
		t.Errorf("'thermal license' missing expected content: %s", stdout)
	}
	if !strings.Contains(stdout, "$30 per developer") || !strings.Contains(stdout, "COMMERCIAL-LICENSE.md") {
		t.Errorf("'thermal license' missing the $30 commercial offer: %s", stdout)
	}
	if strings.Contains(stdout, "contact@jadmadi.com") {
		t.Errorf("'thermal license' contains incorrect contact@jadmadi.com")
	}
	if !strings.Contains(stdout, "contact@jadmadi.net") {
		t.Errorf("'thermal license' missing contact@jadmadi.net: %s", stdout)
	}

	// 2. Text mode via flag
	stdout, stderr, code = runSim(t, "--license")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal --license', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "AGPL-3.0") {
		t.Errorf("'thermal --license' missing AGPL-3.0: %s", stdout)
	}

	// 3. JSON mode
	stdout, stderr, code = runSim(t, "license", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal license --json', got %d. stderr: %s", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal license --json': %v. stdout: %s", err, stdout)
	}
	if res["license"] != "AGPL-3.0-only" {
		t.Errorf("expected license == AGPL-3.0-only, got: %v", res["license"])
	}
	if res["author"] != "Jad Madi" {
		t.Errorf("expected author == Jad Madi, got: %v", res["author"])
	}
	if res["dual_license_available"] != true {
		t.Errorf("expected dual_license_available == true, got: %v", res["dual_license_available"])
	}
	comm, _ := res["commercial_licensing"].(map[string]any)
	if comm["contact"] != "contact@jadmadi.net" {
		t.Errorf("expected commercial_licensing.contact == contact@jadmadi.net, got: %v", comm["contact"])
	}
	commercialUse, _ := comm["commercial_use"].(map[string]any)
	if commercialUse["price_usd"] != float64(30) {
		t.Errorf("expected commercial_licensing.commercial_use.price_usd == 30, got: %v", commercialUse["price_usd"])
	}
	if commercialUse["pricing_page"] != "https://thermal.jadmadi.net/pricing.html" {
		t.Errorf("expected commercial_licensing.commercial_use.pricing_page, got: %v", commercialUse["pricing_page"])
	}
	enterprise, _ := comm["enterprise_oem"].(map[string]any)
	if enterprise["contact"] != "contact@jadmadi.net" {
		t.Errorf("expected commercial_licensing.enterprise_oem.contact == contact@jadmadi.net, got: %v", enterprise["contact"])
	}
}

func TestSimulatedUser_AuditCommand(t *testing.T) {
	// 1. Terminal text mode
	stdout, stderr, code := runSim(t, "audit", "--no-color")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal audit', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Thermal") || !strings.Contains(stdout, "audit · local setup & context health") {
		t.Errorf("'thermal audit' missing header: %s", stdout)
	}
	if !strings.Contains(stdout, "Score:") || !strings.Contains(stdout, "Context Tax:") {
		t.Errorf("'thermal audit' missing score / tax: %s", stdout)
	}

	// 2. JSON mode
	stdout, stderr, code = runSim(t, "audit", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal audit --json', got %d. stderr: %s", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal audit --json': %v. stdout: %s", err, stdout)
	}
	if _, ok := res["score"]; !ok {
		t.Errorf("expected score key in JSON, got %v", res)
	}
	if _, ok := res["grade"]; !ok {
		t.Errorf("expected grade key in JSON, got %v", res)
	}
}

func TestSimulatedUser_ShareCommand(t *testing.T) {
	// 1. Terminal text mode
	stdout, stderr, code := runSim(t, "share", "--no-color")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal share', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Thermal") || !strings.Contains(stdout, "share · stateless streak card") {
		t.Errorf("'thermal share' missing header: %s", stdout)
	}
	if !strings.Contains(stdout, "https://jadmadi.net/project/thermal/share#v1.") {
		t.Errorf("'thermal share' missing valid share URL: %s", stdout)
	}

	// 2. JSON mode
	stdout, stderr, code = runSim(t, "share", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal share --json', got %d. stderr: %s", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal share --json': %v. stdout: %s", err, stdout)
	}
	if _, ok := res["url"]; !ok {
		t.Errorf("expected url key in JSON, got %v", res)
	}
	if _, ok := res["token"]; !ok {
		t.Errorf("expected token key in JSON, got %v", res)
	}
	if _, ok := res["snapshot"]; !ok {
		t.Errorf("expected snapshot key in JSON, got %v", res)
	}
}

func TestSimulatedUser_ChangelogCommand(t *testing.T) {
	// 1. Terminal text mode
	stdout, stderr, code := runSim(t, "changelog", "--no-color")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal changelog', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Thermal") || !strings.Contains(stdout, "changelog · release history & new features") {
		t.Errorf("'thermal changelog' missing header: %s", stdout)
	}
	if !strings.Contains(stdout, "Features:") {
		t.Errorf("'thermal changelog' missing Features: section: %s", stdout)
	}

	// 2. Limit with --top
	stdout, stderr, code = runSim(t, "changelog", "--top", "1", "--no-color")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal changelog --top 1', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Showing 1 of") {
		t.Errorf("'thermal changelog --top 1' missing Showing 1 of: %s", stdout)
	}

	// 3. JSON mode
	stdout, stderr, code = runSim(t, "changelog", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal changelog --json', got %d. stderr: %s", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal changelog --json': %v. stdout: %s", err, stdout)
	}
	releases, ok := res["releases"].([]any)
	if !ok || len(releases) == 0 {
		t.Fatalf("expected releases list in JSON, got %v", res)
	}
}

func TestSimulatedUser_LiveCommand(t *testing.T) {
	// 1. Non-TTY pipe execution
	stdout, stderr, code := runSim(t, "live")
	if code != 0 {
		t.Fatalf("expected exit code 0 for non-TTY 'thermal live', got %d. stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "thermal live is interactive and needs a terminal") {
		t.Errorf("'thermal live' non-TTY output missing hint: %s", stdout)
	}

	// 2. JSON snapshot
	jsonOut, stderr, code := runSim(t, "live", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal live --json', got %d. stderr: %s", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &res); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal live --json': %v. stdout: %s", err, jsonOut)
	}
	if _, ok := res["todayTokens"]; !ok {
		t.Errorf("expected todayTokens in JSON, got %v", res)
	}
	if _, ok := res["flameIntensity"]; !ok {
		t.Errorf("expected flameIntensity in JSON, got %v", res)
	}
	if _, ok := res["sessionTokens"]; !ok {
		t.Errorf("expected sessionTokens in JSON, got %v", res)
	}
	if session, ok := res["sessionTokens"].(float64); !ok || session != 0 {
		t.Errorf("expected launch-relative sessionTokens 0, got %v", res["sessionTokens"])
	}
	if _, ok := res["carriedInTokens"]; !ok {
		t.Errorf("expected carriedInTokens in JSON, got %v", res)
	}
	if _, ok := res["carriedInTurns"]; !ok {
		t.Errorf("expected carriedInTurns in JSON, got %v", res)
	}
	if _, ok := res["carriedInCost"]; !ok {
		t.Errorf("expected carriedInCost in JSON, got %v", res)
	}
	if _, ok := res["carriedInAt"]; !ok {
		t.Errorf("expected carriedInAt in JSON, got %v", res)
	}
	if _, ok := res["todayTurns"]; !ok {
		t.Errorf("expected todayTurns in JSON, got %v", res)
	}
	if _, ok := res["todayCacheHit"]; !ok {
		t.Errorf("expected todayCacheHit in JSON, got %v", res)
	}

	// 3. No-estimate option
	noEstOut, stderr, code := runSim(t, "live", "--json", "--no-estimate")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal live --json --no-estimate', got %d. stderr: %s", code, stderr)
	}
	var noEstRes map[string]any
	if err := json.Unmarshal([]byte(noEstOut), &noEstRes); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal live --json --no-estimate': %v", err)
	}

	// 4. Tool-filtered JSON snapshot
	jsonOut, stderr, code = runSim(t, "live", "codewhale", "--json")
	if code != 0 {
		t.Fatalf("expected exit code 0 for 'thermal live codewhale --json', got %d. stderr: %s", code, stderr)
	}
	if err := json.Unmarshal([]byte(jsonOut), &res); err != nil {
		t.Fatalf("failed to parse JSON from 'thermal live codewhale --json': %v. stdout: %s", err, jsonOut)
	}

	// 4. Inapplicable flags rejected
	for _, badFlag := range []string{"--against=claude-3-5-sonnet", "--chart", "--since=2026-09-01"} {
		_, _, code := runSim(t, "live", badFlag)
		if code == 0 {
			t.Errorf("expected non-zero exit code for 'thermal live %s', got 0", badFlag)
		}
	}
}

func TestSimulatedUser_FlameCeilingFlag(t *testing.T) {
	out, stderr, code := runSim(t, "live", "--json", "--flame-ceiling=9M")
	if code != 0 {
		t.Fatalf("expected --flame-ceiling=9M to succeed, got %d. stderr: %s", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("parse flame-ceiling JSON: %v. stdout: %s", err, out)
	}
	if ceiling, ok := res["flameCeiling"].(float64); !ok || ceiling != 9_000_000 {
		t.Errorf("expected flameCeiling 9M, got %v", res["flameCeiling"])
	}
	if _, ok := res["flameMetric"]; !ok {
		t.Errorf("expected flameMetric in JSON, got %v", res)
	}

	// The standard flag parser path before the positional verb also works.
	out, stderr, code = runSim(t, "--flame-ceiling=4.5M", "live", "--json")
	if code != 0 {
		t.Fatalf("expected pre-verb --flame-ceiling to succeed, got %d. stderr: %s", code, stderr)
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("parse pre-verb flame-ceiling JSON: %v", err)
	}
	if ceiling, ok := res["flameCeiling"].(float64); !ok || ceiling != 4_500_000 {
		t.Errorf("expected flameCeiling 4.5M, got %v", res["flameCeiling"])
	}

	for _, args := range [][]string{
		{"live", "--json", "--flame-ceiling=0"},
		{"live", "--json", "--flame-ceiling=abc"},
		{"live", "--json", "--flame-ceiling"},
		{"daily", "--flame-ceiling=9M"},
	} {
		_, stderr, code := runSim(t, args...)
		if code == 0 {
			t.Errorf("expected non-zero exit for %v", args)
		}
		if stderr == "" {
			t.Errorf("expected stderr diagnostic for %v", args)
		}
	}
}
