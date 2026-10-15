// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

// Command benchmark provisions synthetic Thermal fixtures under a temporary
// directory, validates exact ingestion totals and delta behavior, and measures
// end-to-end CLI latency plus sustained live-monitor CPU behavior with Go
// monotonic timing. It is invoked by scripts/benchmark.sh; correctness failures
// stop the run with a nonzero exit before any performance number is published.
package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	baselineTotal      = int64(1_873_500)
	largeClaudeFiles   = 1000
	largeClaudeRecords = 30
	claudeRecordTokens = int64(9700) // 1000 in + 500 out + 200 cache write + 8000 cache read
)

// baselineByTool is the exact per-source ingestion total produced by the
// deterministic small fixture. These five sources are the only token sources
// the fixture creates; any other non-zero tool fails validation.
var baselineByTool = map[string]int64{
	"devin":     425_000,
	"codex":     70_000,
	"opencode":  87_500,
	"claude":    291_000,
	"codewhale": 1_000_000,
}

func main() {
	repo := flag.String("repo", ".", "path to the Thermal repository root")
	out := flag.String("out", "", "directory for the JSON and text reports (default: new temp dir)")
	coldSamples := flag.Int("cold", 30, "number of cold ingestion samples")
	warmSamples := flag.Int("warm", 100, "number of warm ingestion samples")
	largeWarmSamples := flag.Int("large-warm", 20, "number of warm samples against the large fixture")
	streamSeconds := flag.Int("stream-seconds", 60, "measured streaming window per interval")
	streamWarmup := flag.Int("stream-warmup", 5, "streaming warm-up seconds before the measured window")
	skipStream := flag.Bool("skip-stream", false, "skip sustained live-monitor measurements")
	flag.Parse()

	opts := runOptions{
		Repo:            *repo,
		Out:             *out,
		ColdSamples:     *coldSamples,
		WarmSamples:     *warmSamples,
		LargeWarm:       *largeWarmSamples,
		StreamSeconds:   *streamSeconds,
		StreamWarmupSec: *streamWarmup,
		SkipStream:      *skipStream,
	}
	if err := run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "\nthermal benchmark failed: %v\n", err)
		os.Exit(1)
	}
}

type runOptions struct {
	Repo            string
	Out             string
	ColdSamples     int
	WarmSamples     int
	LargeWarm       int
	StreamSeconds   int
	StreamWarmupSec int
	SkipStream      bool
}

type environment struct {
	Revision   string `json:"revision"`
	Dirty      bool   `json:"dirty"`
	DiffSHA256 string `json:"diffSha256"`
	GoVersion  string `json:"goVersion"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	CPU        string `json:"cpu"`
	Cores      int    `json:"cores"`
}

type fixtureDims struct {
	Home           string `json:"-"`
	Sources        int    `json:"sources"`
	Files          int    `json:"files"`
	ExpectedTokens int64  `json:"expectedTokens"`
}

type latencyStats struct {
	Samples int     `json:"samples"`
	MinMs   float64 `json:"minMs"`
	MeanMs  float64 `json:"meanMs"`
	P50Ms   float64 `json:"p50Ms"`
	P95Ms   float64 `json:"p95Ms"`
	MaxMs   float64 `json:"maxMs"`
}

type deltaCase struct {
	Name     string  `json:"name"`
	Expected int64   `json:"expectedTokens"`
	Observed int64   `json:"observedTokens"`
	Status   string  `json:"status"`
	Seconds  float64 `json:"durationSeconds"`
}

type verdict struct {
	Name       string `json:"name"`
	Status     string `json:"status"` // PASS, MISSED, OBSERVED, UNAVAILABLE
	Detail     string `json:"detail"`
	SampleSize int    `json:"sampleSize"`
}

type report struct {
	Schema      string           `json:"schema"`
	Generated   time.Time        `json:"generatedAt"`
	Command     string           `json:"command"`
	Env         environment      `json:"environment"`
	Small       fixtureDims      `json:"smallFixture"`
	Large       fixtureDims      `json:"largeFixture"`
	Baseline    map[string]int64 `json:"baselineTokens"`
	Deltas      []deltaCase      `json:"deltaCases"`
	Cold        latencyStats     `json:"cold"`
	Warm        latencyStats     `json:"warm"`
	LargeWarm   latencyStats     `json:"largeWarm"`
	ColdSample  []float64        `json:"coldSamplesMs"`
	WarmSample  []float64        `json:"warmSamplesMs"`
	LargeSample []float64        `json:"largeWarmSamplesMs"`
	Streams     []streamResult   `json:"streams,omitempty"`
	Verdicts    []verdict        `json:"verdicts"`
}

func run(opts runOptions) error {
	repoRoot, err := filepath.Abs(opts.Repo)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root %s does not look like the Thermal module: %w", repoRoot, err)
	}

	outDir := opts.Out
	if outDir == "" {
		outDir, err = os.MkdirTemp("", "thermal-bench-report-*")
		if err != nil {
			return fmt.Errorf("create report dir: %w", err)
		}
	} else if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	workDir, err := os.MkdirTemp("", "thermal-bench-work-*")
	if err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	env := describeEnvironment(repoRoot)
	rep := report{
		Schema:    "thermal/benchmark/v1",
		Generated: time.Now().UTC(),
		Command:   strings.Join(os.Args, " "),
		Env:       env,
		Baseline:  baselineByTool,
	}

	fmt.Printf("THERMAL · Ingestion Performance & Latency Baseline\n")
	fmt.Printf("  Source:      %s (dirty=%v, diff sha256 %s)\n", env.Revision, env.Dirty, env.DiffSHA256)
	fmt.Printf("  Environment: %s %s/%s · CPU %s (%d cores)\n", env.GoVersion, env.OS, env.Arch, env.CPU, env.Cores)
	fmt.Printf("  Report dir:  %s\n\n", outDir)

	// 1. Build the CLI binary into the temporary work dir: the repository root
	// and the user's real home stay untouched.
	binPath := filepath.Join(workDir, "thermal")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	fmt.Printf("1. Building thermal into %s ...\n", binPath)
	build := exec.Command("go", "build", "-o", binPath, "./cmd/thermal")
	build.Dir = repoRoot
	build.Stdout = io.Discard
	var buildErr bytes.Buffer
	build.Stderr = &buildErr
	if err := build.Run(); err != nil {
		return fmt.Errorf("go build: %w\n%s", err, buildErr.String())
	}

	// 2. Small deterministic fixture with exact expected totals.
	fmt.Printf("2. Provisioning small fixture (exact %d tokens) ...\n", baselineTotal)
	small, err := createSmallFixture(workDir)
	if err != nil {
		return err
	}
	rep.Small = small
	smallEnv := isolatedEnv(small.Home)
	smallCache := filepath.Join(small.Home, ".cache", "thermal")

	observed, err := runThermalTotals(binPath, smallEnv)
	if err != nil {
		return fmt.Errorf("baseline run: %w", err)
	}
	if err := validateTotals(observed, baselineByTool, baselineTotal); err != nil {
		return fmt.Errorf("baseline fixture validation: %w", err)
	}
	fmt.Printf("   ✔ baseline totals exact: %d tokens across %d sources\n", baselineTotal, len(baselineByTool))

	// 3. Delta behavior: valid Devin/OpenCode appends plus rewrite, truncate,
	// and delete cases against the same tracked expectation.
	fmt.Printf("3. Validating delta behavior ...\n")
	expected := cloneTotals(baselineByTool)
	deltas, err := validateDeltas(binPath, smallEnv, small.Home, expected)
	if err != nil {
		return err
	}
	rep.Deltas = deltas
	for _, d := range deltas {
		fmt.Printf("   ✔ %-24s expected %+d observed %+d (%s)\n", d.Name, d.Expected, d.Observed, d.Status)
	}

	// 4. Cold and warm end-to-end latency on the small fixture.
	fmt.Printf("4. Measuring cold (%d samples) and warm (%d samples) latency ...\n", opts.ColdSamples, opts.WarmSamples)
	if err := resetThermalSnapshots(smallCache); err != nil {
		return err
	}
	cold, err := measureLatency(binPath, smallEnv, opts.ColdSamples, true, smallCache)
	if err != nil {
		return fmt.Errorf("cold sampling: %w", err)
	}
	rep.Cold = summarize(cold)
	rep.ColdSample = millis(cold)
	warm, err := measureLatency(binPath, smallEnv, opts.WarmSamples, false, smallCache)
	if err != nil {
		return fmt.Errorf("warm sampling: %w", err)
	}
	rep.Warm = summarize(warm)
	rep.WarmSample = millis(warm)
	if artifacts, err := countSnapshots(smallCache); err != nil {
		return err
	} else if artifacts == 0 {
		return errors.New("warm cache state never materialized: zero snapshot artifacts after sampling")
	}
	fmt.Printf("   ✔ cold p50 %.1fms p95 %.1fms · warm p50 %.1fms p95 %.1fms\n",
		rep.Cold.P50Ms, rep.Cold.P95Ms, rep.Warm.P50Ms, rep.Warm.P95Ms)

	// 5. Large fixture: 1000 Claude files with 30 records each.
	fmt.Printf("5. Provisioning large fixture (%d Claude files × %d records) ...\n", largeClaudeFiles, largeClaudeRecords)
	large, err := createLargeFixture(workDir)
	if err != nil {
		return err
	}
	rep.Large = large
	largeEnv := isolatedEnv(large.Home)
	largeObserved, err := runThermalTotals(binPath, largeEnv)
	if err != nil {
		return fmt.Errorf("large fixture run: %w", err)
	}
	largeWant := map[string]int64{"claude": large.ExpectedTokens}
	if err := validateTotals(largeObserved, largeWant, large.ExpectedTokens); err != nil {
		return fmt.Errorf("large fixture validation: %w", err)
	}
	fmt.Printf("   ✔ large fixture totals exact: %d tokens\n", large.ExpectedTokens)

	// Warm the large cache, then sample warm latency at that scale.
	if _, err := runThermalTotals(binPath, largeEnv); err != nil {
		return fmt.Errorf("large warm-up: %w", err)
	}
	largeWarm, err := measureLatency(binPath, largeEnv, opts.LargeWarm, false, "")
	if err != nil {
		return fmt.Errorf("large warm sampling: %w", err)
	}
	rep.LargeWarm = summarize(largeWarm)
	rep.LargeSample = millis(largeWarm)
	fmt.Printf("   ✔ large warm p50 %.1fms p95 %.1fms\n", rep.LargeWarm.P50Ms, rep.LargeWarm.P95Ms)

	// 6. Sustained live-monitor measurements: idle plus one source change per
	// interval, at least streamSeconds after warm-up.
	if !opts.SkipStream {
		for _, interval := range []time.Duration{100 * time.Millisecond, time.Second} {
			fmt.Printf("6. Sustained live monitor at %s (%ds window after %ds warm-up) ...\n",
				interval, opts.StreamSeconds, opts.StreamWarmupSec)
			res, err := measureStream(binPath, largeEnv, interval,
				time.Duration(opts.StreamWarmupSec)*time.Second,
				time.Duration(opts.StreamSeconds)*time.Second,
				func() (int64, error) { return appendLargeClaudeFile(large.Home) })
			if err != nil {
				return fmt.Errorf("stream at %s: %w", interval, err)
			}
			rep.Streams = append(rep.Streams, res)
			fmt.Printf("   ✔ %s: lines %d · events %d · CPU %.2f%% · freshness %.0fms · cancel %.0fms\n",
				res.Interval, res.Lines, res.Events, res.CPUPercent, res.FreshnessMs, res.CancellationMs)
		}
	}

	rep.Verdicts = buildVerdicts(rep)
	if err := writeReports(outDir, rep); err != nil {
		return err
	}
	printReport(rep, outDir)
	return nil
}

// describeEnvironment records the source revision, working-tree diff hash, and
// host details so every number binds to an exact candidate.
func describeEnvironment(repoRoot string) environment {
	env := environment{
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Cores:     runtime.NumCPU(),
		CPU:       cpuModel(),
	}
	env.Revision = gitOutput(repoRoot, "rev-parse", "HEAD")
	if env.Revision == "" {
		env.Revision = "unknown"
	}
	status := gitOutput(repoRoot, "status", "--porcelain")
	env.Dirty = strings.TrimSpace(status) != ""
	diff := ""
	if env.Dirty {
		diff = gitOutput(repoRoot, "diff", "HEAD")
	}
	sum := sha256.Sum256([]byte(diff))
	env.DiffSHA256 = hex.EncodeToString(sum[:])
	return env
}

func gitOutput(repoRoot string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func cpuModel() string {
	if runtime.GOOS != "linux" {
		return "unknown"
	}
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "model name") {
			if idx := strings.Index(line, ":"); idx >= 0 {
				return strings.TrimSpace(line[idx+1:])
			}
		}
	}
	return "unknown"
}

// isolatedEnv builds a fully hermetic environment: home directory plus every
// tool-root override and XDG path the loaders consult.
func isolatedEnv(home string) []string {
	return []string{
		"HOME=" + home,
		"USERPROFILE=" + home,
		"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
		"XDG_DATA_HOME=" + filepath.Join(home, ".local", "share"),
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
		"GROK_HOME=" + filepath.Join(home, ".grok"),
		"DSH_HOME=" + filepath.Join(home, ".dsh"),
		"HERMES_HOME=" + filepath.Join(home, ".hermes"),
		"CODEX_HOME=" + filepath.Join(home, ".codex"),
		"NO_COLOR=1",
		"CLICOLOR_FORCE=0",
		"THERMAL_NO_UPDATE_CHECK=1",
	}
}

// runThermalTotals runs the JSON leaderboard offline and returns non-zero
// per-tool lifetime totals.
func runThermalTotals(bin string, env []string) (map[string]int64, error) {
	cmd := exec.Command(bin, "--json", "--offline")
	cmd.Env = env
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("thermal --json --offline: %w\n%s", err, errOut.String())
	}
	return parseToolTotals(out.Bytes())
}

func parseToolTotals(data []byte) (map[string]int64, error) {
	var payload struct {
		Results []struct {
			Tool    string `json:"tool"`
			Summary struct {
				LifetimeTokens int64 `json:"lifetimeTokens"`
			} `json:"summary"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse thermal JSON: %w", err)
	}
	totals := make(map[string]int64)
	for _, r := range payload.Results {
		if r.Summary.LifetimeTokens > 0 {
			totals[r.Tool] = r.Summary.LifetimeTokens
		}
	}
	return totals, nil
}

// validateTotals requires every expected source to match exactly, forbids
// unexpected non-zero sources, and checks the grand total.
func validateTotals(observed, want map[string]int64, wantTotal int64) error {
	for tool, wantTokens := range want {
		got, ok := observed[tool]
		if !ok {
			return fmt.Errorf("expected source %q missing from results", tool)
		}
		if got != wantTokens {
			return fmt.Errorf("source %q reported %d tokens, want exactly %d", tool, got, wantTokens)
		}
	}
	for tool, got := range observed {
		if _, ok := want[tool]; !ok {
			return fmt.Errorf("unexpected source %q with %d tokens", tool, got)
		}
	}
	if total := sumTotals(observed); total != wantTotal {
		return fmt.Errorf("total tokens %d, want exactly %d", total, wantTotal)
	}
	return nil
}

func sumTotals(totals map[string]int64) int64 {
	var total int64
	for _, v := range totals {
		total += v
	}
	return total
}

func cloneTotals(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// validateDeltas mutates the small fixture one source at a time and asserts the
// observed delta matches the tracked expectation. The tracked map is updated in
// place so later cases build on earlier ones.
func validateDeltas(bin string, env []string, home string, expected map[string]int64) ([]deltaCase, error) {
	devinDB := filepath.Join(home, ".local", "share", "devin", "cli", "sessions.db")
	opencodeDB := filepath.Join(home, ".local", "share", "opencode", "opencode.db")

	type mutation struct {
		name   string
		apply  func() error
		adjust func(map[string]int64)
	}
	mutations := []mutation{
		{
			name: "devin append",
			apply: func() error {
				return execSQL(devinDB,
					`INSERT INTO message_nodes (session_id, created_at, chat_message) VALUES ('devin-sess-1', ?, '{"role":"assistant","metadata":{"metrics":{"input_tokens":50,"output_tokens":20,"cache_creation_tokens":0,"cache_read_tokens":10}}}')`,
					time.Now().Unix())
			},
			adjust: func(m map[string]int64) { m["devin"] += 80 },
		},
		{
			name: "opencode append",
			apply: func() error {
				return execSQL(opencodeDB,
					`INSERT INTO session_v2 VALUES ('oc-delta-1', 1000, 200, 50, 400, 100, 0.05, 0, 0, 0, 'code', ?, ?, '{"id":"claude-3-5-sonnet"}')`,
					time.Now().UnixMilli(), time.Now().UnixMilli())
			},
			adjust: func(m map[string]int64) { m["opencode"] += 1750 },
		},
		{
			name: "devin rewrite",
			apply: func() error {
				return execSQL(devinDB,
					`UPDATE message_nodes SET chat_message = '{"role":"assistant","metadata":{"metrics":{"input_tokens":100,"output_tokens":0,"cache_creation_tokens":0,"cache_read_tokens":0}}}' WHERE row_id = (SELECT MIN(row_id) FROM message_nodes WHERE session_id = 'devin-sess-1')`)
			},
			adjust: func(m map[string]int64) { m["devin"] -= 750 },
		},
		{
			name: "devin truncate",
			apply: func() error {
				return execSQL(devinDB,
					`DELETE FROM message_nodes WHERE row_id = (SELECT MAX(row_id) FROM message_nodes WHERE session_id = 'devin-sess-2')`)
			},
			adjust: func(m map[string]int64) { m["devin"] -= 850 },
		},
		{
			name: "codewhale delete",
			apply: func() error {
				return os.Remove(filepath.Join(home, ".codewhale", "sessions", "session_1.json"))
			},
			adjust: func(m map[string]int64) { m["codewhale"] -= 50_000 },
		},
	}

	var cases []deltaCase
	for _, mu := range mutations {
		before := cloneTotals(expected)
		if err := mu.apply(); err != nil {
			return nil, fmt.Errorf("mutation %s: %w", mu.name, err)
		}
		mu.adjust(expected)
		start := time.Now()
		observed, err := runThermalTotals(bin, env)
		if err != nil {
			return nil, fmt.Errorf("mutation %s run: %w", mu.name, err)
		}
		elapsed := time.Since(start)
		if err := validateTotals(observed, expected, sumTotals(expected)); err != nil {
			return nil, fmt.Errorf("delta %s rejected: %w", mu.name, err)
		}
		cases = append(cases, deltaCase{
			Name:     mu.name,
			Expected: sumTotals(expected) - sumTotals(before),
			Observed: sumTotals(observed) - sumTotals(before),
			Status:   "PASS",
			Seconds:  elapsed.Seconds(),
		})
	}
	return cases, nil
}

func execSQL(dbPath, query string, args ...any) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(query, args...)
	return err
}

// measureLatency samples a full CLI run. When cold is true every Thermal
// snapshot artifact is removed before each sample; pricing stays cached.
func measureLatency(bin string, env []string, samples int, cold bool, cacheDir string) ([]time.Duration, error) {
	durations := make([]time.Duration, 0, samples)
	for i := 0; i < samples; i++ {
		if cold {
			if err := resetThermalSnapshots(cacheDir); err != nil {
				return nil, err
			}
		}
		start := time.Now()
		cmd := exec.Command(bin, "--json", "--offline")
		cmd.Env = env
		cmd.Stdout = io.Discard
		var errOut bytes.Buffer
		cmd.Stderr = &errOut
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("sample %d: %w\n%s", i, err, errOut.String())
		}
		durations = append(durations, time.Since(start))
	}
	return durations, nil
}

func summarize(durations []time.Duration) latencyStats {
	if len(durations) == 0 {
		return latencyStats{}
	}
	ms := millis(durations)
	sort.Float64s(ms)
	stats := latencyStats{
		Samples: len(ms),
		MinMs:   ms[0],
		MaxMs:   ms[len(ms)-1],
	}
	var sum float64
	for _, v := range ms {
		sum += v
	}
	stats.MeanMs = sum / float64(len(ms))
	stats.P50Ms = percentile(ms, 50)
	stats.P95Ms = percentile(ms, 95)
	return stats
}

func percentile(sorted []float64, p int) float64 {
	idx := (len(sorted) * p) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func millis(durations []time.Duration) []float64 {
	out := make([]float64, len(durations))
	for i, d := range durations {
		out[i] = float64(d.Microseconds()) / 1000.0
	}
	return out
}

// resetThermalSnapshots removes every Thermal snapshot artifact in the given
// cache directory while preserving the cached pricing catalog. It is confined
// to the fixture home: user caches are never touched.
func resetThermalSnapshots(cacheDir string) error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(cacheDir, 0o755)
		}
		return err
	}
	for _, e := range entries {
		if e.Name() == "pricing.json" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(cacheDir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func countSnapshots(cacheDir string) (int, error) {
	count := 0
	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "pricing.json" {
			return nil
		}
		count++
		return nil
	})
	return count, err
}

func buildVerdicts(rep report) []verdict {
	var verdicts []verdict
	warm := verdict{
		Name:       "warm latency (small fixture)",
		SampleSize: rep.Warm.Samples,
	}
	switch {
	case rep.Warm.Samples == 0:
		warm.Status = "UNAVAILABLE"
		warm.Detail = "no warm samples recorded"
	case rep.Warm.P50Ms <= 10.0:
		warm.Status = "PASS"
		warm.Detail = fmt.Sprintf("p50 %.1fms within the 10ms governance budget (p95 %.1fms) on this machine", rep.Warm.P50Ms, rep.Warm.P95Ms)
	default:
		warm.Status = "MISSED"
		warm.Detail = fmt.Sprintf("p50 %.1fms exceeds the 10ms governance budget on this machine (p95 %.1fms); disclosed gap, not a universal failure", rep.Warm.P50Ms, rep.Warm.P95Ms)
	}
	verdicts = append(verdicts, warm)

	large := verdict{
		Name:       "warm latency (large fixture)",
		SampleSize: rep.LargeWarm.Samples,
	}
	switch {
	case rep.LargeWarm.Samples == 0:
		large.Status = "UNAVAILABLE"
		large.Detail = "no large-fixture samples recorded"
	case rep.LargeWarm.P50Ms <= 10.0:
		large.Status = "PASS"
		large.Detail = fmt.Sprintf("p50 %.1fms within the 10ms governance budget at %d Claude files (p95 %.1fms)",
			rep.LargeWarm.P50Ms, largeClaudeFiles, rep.LargeWarm.P95Ms)
	default:
		large.Status = "MISSED"
		large.Detail = fmt.Sprintf("p50 %.1fms exceeds the 10ms governance budget at %d Claude files (p95 %.1fms); disclosed scale gap, not a universal failure",
			rep.LargeWarm.P50Ms, largeClaudeFiles, rep.LargeWarm.P95Ms)
	}
	verdicts = append(verdicts, large)

	for _, s := range rep.Streams {
		// The previously published CPU budgets were measured on the small
		// fixture over two-second windows, so they are reported as context,
		// not as a pass/fail gate at this fixture scale.
		v := verdict{
			Name:       "live CPU (" + s.Interval + ")",
			SampleSize: int(s.WallSeconds - s.WarmupSeconds),
		}
		if s.CPUSource != "rusage" {
			v.Status = "UNAVAILABLE"
			v.Detail = "child CPU accounting unavailable on this platform"
		} else {
			v.Status = "OBSERVED"
			v.Detail = fmt.Sprintf("%.2f%% CPU over a %.0fs window at %d Claude files (%.3fs CPU, %.1fs process wall including warm-up); previously published budgets (<5%% at 100ms, <2%% at 1s) were measured on a 30-file fixture and are not directly comparable",
				s.CPUPercent, s.WallSeconds-s.WarmupSeconds, largeClaudeFiles, s.CPUSeconds, s.WallSeconds)
		}
		verdicts = append(verdicts, v)

		if s.FreshnessMs > 0 {
			verdicts = append(verdicts, verdict{
				Name:       "source freshness (" + s.Interval + ")",
				Status:     "OBSERVED",
				SampleSize: 1,
				Detail:     fmt.Sprintf("source append reflected after %.0fms (%d tokens, %d change event(s)); no published budget", s.FreshnessMs, s.FreshnessToks, s.Events),
			})
		}
	}
	return verdicts
}

func writeReports(outDir string, rep report) error {
	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "bench-report.json"), raw, 0o644); err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("# Thermal benchmark report (thermal/benchmark/v1)\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n", rep.Generated.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Command: %s\n", rep.Command))
	sb.WriteString(fmt.Sprintf("Revision: %s (dirty=%v) diff sha256 %s\n", rep.Env.Revision, rep.Env.Dirty, rep.Env.DiffSHA256))
	sb.WriteString(fmt.Sprintf("Environment: %s %s/%s · CPU %s (%d cores)\n\n", rep.Env.GoVersion, rep.Env.OS, rep.Env.Arch, rep.Env.CPU, rep.Env.Cores))
	sb.WriteString(fmt.Sprintf("Small fixture: %d sources, %d tokens\n", rep.Small.Sources, rep.Small.ExpectedTokens))
	sb.WriteString(fmt.Sprintf("Large fixture: %d files, %d tokens\n\n", rep.Large.Files, rep.Large.ExpectedTokens))
	sb.WriteString(fmt.Sprintf("Cold: %+v\nWarm: %+v\nLarge warm: %+v\n\n", rep.Cold, rep.Warm, rep.LargeWarm))
	for _, s := range rep.Streams {
		sb.WriteString(fmt.Sprintf("Stream %s: %+v\n", s.Interval, s))
	}
	sb.WriteString("\nVerdicts:\n")
	for _, v := range rep.Verdicts {
		sb.WriteString(fmt.Sprintf("  [%s] %s (n=%d): %s\n", v.Status, v.Name, v.SampleSize, v.Detail))
	}
	return os.WriteFile(filepath.Join(outDir, "bench-report.txt"), []byte(sb.String()), 0o644)
}

func printReport(rep report, outDir string) {
	fmt.Printf("\nSummary\n")
	fmt.Printf("  Cold:  p50 %.1fms  p95 %.1fms  (n=%d)\n", rep.Cold.P50Ms, rep.Cold.P95Ms, rep.Cold.Samples)
	fmt.Printf("  Warm:  p50 %.1fms  p95 %.1fms  (n=%d)\n", rep.Warm.P50Ms, rep.Warm.P95Ms, rep.Warm.Samples)
	fmt.Printf("  Large: p50 %.1fms  p95 %.1fms  (n=%d)\n", rep.LargeWarm.P50Ms, rep.LargeWarm.P95Ms, rep.LargeWarm.Samples)
	for _, s := range rep.Streams {
		fmt.Printf("  Stream %s: lines %d · events %d · CPU %.2f%% · freshness %.0fms · cancel %.0fms\n",
			s.Interval, s.Lines, s.Events, s.CPUPercent, s.FreshnessMs, s.CancellationMs)
	}
	fmt.Printf("\nVerdicts\n")
	for _, v := range rep.Verdicts {
		fmt.Printf("  [%s] %s (n=%d): %s\n", v.Status, v.Name, v.SampleSize, v.Detail)
	}
	fmt.Printf("\nRaw artifacts: %s\n", outDir)
}
