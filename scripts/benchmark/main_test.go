// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMain doubles as a controllable stub thermal binary: when BENCH_STUB_MODE
// is set the test binary prints a canned result or exits nonzero, letting the
// subprocess-handling paths be tested without building extra binaries.
func TestMain(m *testing.M) {
	if mode := os.Getenv("BENCH_STUB_MODE"); mode != "" {
		switch mode {
		case "ok":
			fmt.Print(`{"results":[` +
				`{"tool":"devin","summary":{"lifetimeTokens":425000}},` +
				`{"tool":"codex","summary":{"lifetimeTokens":70000}},` +
				`{"tool":"opencode","summary":{"lifetimeTokens":87500}},` +
				`{"tool":"claude","summary":{"lifetimeTokens":291000}},` +
				`{"tool":"codewhale","summary":{"lifetimeTokens":1000000}}]}`)
			os.Exit(0)
		case "garbage":
			fmt.Print("<html>definitely not json</html>")
			os.Exit(0)
		case "exit1":
			fmt.Fprint(os.Stderr, "stub failure")
			os.Exit(3)
		}
		os.Exit(9)
	}
	os.Exit(m.Run())
}

func TestValidateTotalsExactAndPerturbed(t *testing.T) {
	exact := cloneTotals(baselineByTool)
	if err := validateTotals(exact, baselineByTool, baselineTotal); err != nil {
		t.Fatalf("exact baseline rejected: %v", err)
	}

	offByOne := cloneTotals(baselineByTool)
	offByOne["claude"]--
	if err := validateTotals(offByOne, baselineByTool, baselineTotal-1); err == nil {
		t.Fatal("per-source perturbation accepted")
	}

	extra := cloneTotals(baselineByTool)
	extra["agy"] = 7
	if err := validateTotals(extra, baselineByTool, baselineTotal+7); err == nil {
		t.Fatal("unexpected extra source accepted")
	}

	missing := cloneTotals(baselineByTool)
	delete(missing, "codex")
	if err := validateTotals(missing, baselineByTool, baselineTotal-70000); err == nil {
		t.Fatal("missing source accepted")
	}

	if err := validateTotals(cloneTotals(baselineByTool), baselineByTool, baselineTotal+1); err == nil {
		t.Fatal("grand-total mismatch accepted despite equal per-source values")
	}
}

func TestRunThermalTotalsSubprocessHandling(t *testing.T) {
	bin := os.Args[0]

	env := append(os.Environ(), "BENCH_STUB_MODE=ok")
	totals, err := runThermalTotals(bin, env)
	if err != nil {
		t.Fatalf("stub ok mode rejected: %v", err)
	}
	if err := validateTotals(totals, baselineByTool, baselineTotal); err != nil {
		t.Fatalf("stub ok totals did not match exact baseline: %v", err)
	}

	if _, err := runThermalTotals(bin, append(os.Environ(), "BENCH_STUB_MODE=exit1")); err == nil {
		t.Fatal("nonzero subprocess exit accepted")
	}
	if _, err := runThermalTotals(bin, append(os.Environ(), "BENCH_STUB_MODE=garbage")); err == nil {
		t.Fatal("malformed subprocess output accepted")
	}
}

func TestResetThermalSnapshotsKeepsPricingCatalog(t *testing.T) {
	cacheDir := t.TempDir()
	mustWrite := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(filepath.Join(cacheDir, "pricing.json"), "{}")
	mustWrite(filepath.Join(cacheDir, "devin", "snapshot.json"), "{}")
	mustWrite(filepath.Join(cacheDir, "codex.json"), "{}")
	mustWrite(filepath.Join(cacheDir, "update.json"), "{}")

	if err := resetThermalSnapshots(cacheDir); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "pricing.json")); err != nil {
		t.Fatalf("pricing catalog was removed: %v", err)
	}
	for _, gone := range []string{"devin", "codex.json", "update.json"} {
		if _, err := os.Stat(filepath.Join(cacheDir, gone)); err == nil {
			t.Fatalf("snapshot artifact %s survived the cold reset", gone)
		}
	}

	// Cold reset then warm artifacts: the counter must distinguish states.
	if count, err := countSnapshots(cacheDir); err != nil || count != 0 {
		t.Fatalf("cold snapshot count = %d (err=%v), want 0", count, err)
	}
	mustWrite(filepath.Join(cacheDir, "claude", "state.json"), "{}")
	if count, err := countSnapshots(cacheDir); err != nil || count != 1 {
		t.Fatalf("warm snapshot count = %d (err=%v), want 1", count, err)
	}
}

func TestSummarizePercentiles(t *testing.T) {
	var durations []time.Duration
	for i := 1; i <= 100; i++ {
		durations = append(durations, time.Duration(i)*time.Millisecond)
	}
	stats := summarize(durations)
	if stats.Samples != 100 || stats.MinMs != 1 || stats.MaxMs != 100 {
		t.Fatalf("unexpected summarize bounds: %+v", stats)
	}
	if stats.P50Ms != 51 || stats.P95Ms != 96 {
		t.Fatalf("unexpected percentiles: p50=%v p95=%v", stats.P50Ms, stats.P95Ms)
	}
	if stats.MeanMs != 50.5 {
		t.Fatalf("unexpected mean: %v", stats.MeanMs)
	}
}

func TestParseStreamLineValidation(t *testing.T) {
	line, valid := parseStreamLine([]byte(`{"type":"tick","snapshot":{"todayTokens":42}}`))
	if !valid || !line.hasSnapshot || line.tokens != 42 {
		t.Fatalf("snapshot line misparsed: %+v valid=%v", line, valid)
	}
	line, valid = parseStreamLine([]byte(`{"type":"event","event":{"tool":"Claude","tokens":9700}}`))
	if !valid || !line.hasEvent || line.eventTool != "Claude" || line.eventTokens != 9700 {
		t.Fatalf("event line misclassified: %+v valid=%v", line, valid)
	}
	if _, valid := parseStreamLine([]byte(`{oops`)); valid {
		t.Fatal("malformed line accepted")
	}
}

func TestCreateSmallFixtureDims(t *testing.T) {
	workDir := t.TempDir()
	dims, err := createSmallFixture(workDir)
	if err != nil {
		t.Fatalf("create small fixture: %v", err)
	}
	if dims.Sources != 5 || dims.ExpectedTokens != baselineTotal {
		t.Fatalf("unexpected fixture dims: %+v", dims)
	}
	var files int
	_ = filepath.Walk(dims.Home, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			files++
		}
		return nil
	})
	// 73 workload files plus the pricing catalog.
	if files != dims.Files+1 {
		t.Fatalf("fixture file count = %d, want %d workload files plus catalog", files, dims.Files)
	}
}
