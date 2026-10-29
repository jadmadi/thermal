// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

func TestLoadMuseData_SessionIndex(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "session-index.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sessions (
			session_id TEXT PRIMARY KEY, model_id TEXT, prompt_count INTEGER,
			created_at_us INTEGER, updated_at_us INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	// Day 1: two sessions (one with zero prompts, contributing no active
	// day). Day 2: one session. Timestamps are UTC midday so the local-day
	// bucketing holds in any timezone.
	_, err = db.Exec(`
		INSERT INTO sessions VALUES ('s1', 'muse-spark', 5, 1784570400000000, 1784570700000000);
		INSERT INTO sessions VALUES ('s2', 'muse-spark', 0, 1784570400000000, 1784570400000000);
		INSERT INTO sessions VALUES ('s3', 'muse-spark-2', 3, 1784834400000000, 1784834700000000);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, _, err := LoadMuseData(dbPath)
	if err != nil {
		t.Fatalf("LoadMuseData error: %v", err)
	}
	if sum.Sessions != 3 {
		t.Errorf("expected sessions=3, got %d", sum.Sessions)
	}
	if sum.LifetimeTokens != 8 {
		t.Errorf("expected lifetime=8 prompts, got %d", sum.LifetimeTokens)
	}
	if sum.LongestSessionMs != 300000 { // s1 and s3 tie at 5 minutes
		t.Errorf("expected longest=300000ms, got %d", sum.LongestSessionMs)
	}
	if sum.ModelBreakdown["muse-spark"] != 2 || sum.ModelBreakdown["muse-spark-2"] != 1 {
		t.Errorf("expected model breakdown 2/1, got %v", sum.ModelBreakdown)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 5 || daily[0].Turns != 5 {
		t.Errorf("expected first day 5/5, got %d/%d", daily[0].Tokens, daily[0].Turns)
	}
	if daily[1].Tokens != 3 || daily[1].Turns != 3 {
		t.Errorf("expected second day 3/3, got %d/%d", daily[1].Tokens, daily[1].Turns)
	}
}

func TestLoadMuseData_SessionJSONL(t *testing.T) {
	dataDir := t.TempDir()
	sessDir := filepath.Join(dataDir, "sessions", "2026", "10", "29", "sess-alpha")
	subDir := filepath.Join(sessDir, "subagent", "sub-1")
	revDir := filepath.Join(sessDir, "approval-review")
	catDir := filepath.Join(dataDir, "model-catalog")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(revDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(catDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Local model-catalog file with official Muse USD rates per 1M tokens
	catJSON := `{"schema_version":2,"provider_id":"meta","rows":[` +
		`{"model_id":"muse-spark-1.3-contributor","cost":{"input":"0.10","output":"0.20","cached":"0.002","currency":"USD"}},` +
		`{"model_id":"muse-spark-1.3","cost":{"input":"1.25","output":"4.25","cached":"0.15","currency":"USD"}}` +
		`]}`
	if err := os.WriteFile(filepath.Join(catDir, "meta.json"), []byte(catJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Main session on 2026-10-29 (1793275200000000 is 2026-10-29 12:00:00 UTC)
	mainLog := filepath.Join(sessDir, "session.jsonl")
	mainContent := `{"recorded_at": 1793275200000000, "payload_type": "runtime.session.metadata", "payload": {"record": {"workspace_root": "/workspace/my-project"}}}` + "\n" +
		`{"recorded_at": 1793275210000000, "payload_type": "runtime.session", "payload": {"event": {"kind": "model_completed", "model": "muse-spark-1.3-contributor", "duration_ms": 2500, "usage": {"input_tokens": 1000, "output_tokens": 300, "reasoning_tokens": 100, "cache_read_tokens": 800, "cache_write_tokens": 0}}}}` + "\n"
	if err := os.WriteFile(mainLog, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Subagent log: inherits workspace_root from sess-alpha
	subLog := filepath.Join(subDir, "session.jsonl")
	subContent := `{"recorded_at": 1793275220000000, "payload_type": "runtime.session", "payload": {"event": {"kind": "model_completed", "model": "muse-spark-1.3-contributor", "duration_ms": 1500, "usage": {"input_tokens": 500, "output_tokens": 100, "reasoning_tokens": 20, "cache_read_tokens": 400, "cache_write_tokens": 0}}}}` + "\n"
	if err := os.WriteFile(subLog, []byte(subContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Approval review log: uses synthetic sandbox timestamp 1780531500000001 (2026-06-04)
	// and declares model via status event facets before model_completed omits model.
	revLog := filepath.Join(revDir, "rev-1.jsonl")
	revContent := `{"recorded_at": 1780531500000001, "stream": {"kind": "approval.review.runtime"}, "payload_type": "runtime.session", "payload": {"event": {"kind": "status", "details": {"facets": [{"detail": {"model": "muse-spark-1.3"}}]}}}}` + "\n" +
		`{"recorded_at": 1780531500000002, "stream": {"kind": "approval.review.runtime"}, "payload_type": "runtime.session", "payload": {"event": {"kind": "model_completed", "duration_ms": 1000, "usage": {"input_tokens": 200, "output_tokens": 50, "reasoning_tokens": 10, "cache_read_tokens": 150, "cache_write_tokens": 0}}}}` + "\n"
	if err := os.WriteFile(revLog, []byte(revContent), 0644); err != nil {
		t.Fatal(err)
	}

	sum, daily, projects, err := LoadMuseData(dataDir)
	if err != nil {
		t.Fatalf("LoadMuseData failed: %v", err)
	}

	if sum.Sessions != 1 {
		t.Errorf("expected 1 session, got %d", sum.Sessions)
	}
	// Synthetic timestamp in approval-review must not inflate session duration:
	// (1793275220000000 - 1793275200000000) / 1000 = 20000ms
	if sum.LongestSessionMs != 20000 {
		t.Errorf("expected longest session 20000ms, got %d", sum.LongestSessionMs)
	}

	// Main: netIn = 1000-800 = 200, netOut = 300-100 = 200, reas = 100, cache = 800 -> 1300 tokens
	// Sub: netIn = 500-400 = 100, netOut = 100-20 = 80, reas = 20, cache = 400 -> 600 tokens
	// Rev: netIn = 200-150 = 50, netOut = 50-10 = 40, reas = 10, cache = 150 -> 250 tokens
	// Total: in = 350, out = 320, reas = 130, cache = 1350 -> total = 2150 tokens
	expectedTotal := int64(2150)
	expectedIn := int64(350)
	expectedOut := int64(320)
	expectedReas := int64(130)
	expectedCache := int64(1350)

	if sum.LifetimeTokens != expectedTotal {
		t.Errorf("expected total tokens %d, got %d", expectedTotal, sum.LifetimeTokens)
	}
	if sum.InputTokens != expectedIn {
		t.Errorf("expected input tokens %d, got %d", expectedIn, sum.InputTokens)
	}
	if sum.OutputTokens != expectedOut {
		t.Errorf("expected output tokens %d, got %d", expectedOut, sum.OutputTokens)
	}
	if sum.ReasoningTokens != expectedReas {
		t.Errorf("expected reasoning tokens %d, got %d", expectedReas, sum.ReasoningTokens)
	}
	if sum.CacheTokens != expectedCache {
		t.Errorf("expected cache tokens %d, got %d", expectedCache, sum.CacheTokens)
	}
	// Strict cost provenance: Muse records token counts, not billed invoice receipts.
	// Rate-card pricing belongs in internal/pricing (EstimatedCost), so StoredCost stays 0.
	if sum.Cost != 0 {
		t.Errorf("expected 0 recorded cost, got %f", sum.Cost)
	}

	// Model breakdown: 2 turns of muse-spark-1.3-contributor, 1 turn of muse-spark-1.3
	if sum.ModelBreakdown["muse-spark-1.3-contributor"] != 2 || sum.ModelBreakdown["muse-spark-1.3"] != 1 {
		t.Errorf("expected 2/1 model breakdown, got %v", sum.ModelBreakdown)
	}

	// Daily rows: all 3 turns bucket into 2026-10-29 (no synthetic 2026-06-04 leak)
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d (%+v)", len(daily), daily)
	}
	d := daily[0]
	if d.Tokens != expectedTotal || d.Input != expectedIn || d.Output != expectedOut || d.Reasoning != expectedReas || d.Cache != expectedCache || d.Turns != 3 || d.Cost != 0 {
		t.Errorf("daily row mismatch: %+v", d)
	}

	// Project attribution
	if len(projects) != 1 {
		t.Fatalf("expected 1 project day, got %d", len(projects))
	}
	p := projects[0]
	if p.Project != "/workspace/my-project" {
		t.Errorf("expected project '/workspace/my-project', got %q", p.Project)
	}
	if p.Tokens != expectedTotal || p.Turns != 3 || p.Cost != 0 {
		t.Errorf("project day mismatch: %+v", p)
	}
}

func TestLoadMuseData_Hybrid(t *testing.T) {
	dataDir := t.TempDir()
	sessDir := filepath.Join(dataDir, "sessions", "2026", "10", "29", "sess-live")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}

	// sess-live in JSONL
	mainLog := filepath.Join(sessDir, "session.jsonl")
	mainContent := `{"recorded_at": 1780531500000000, "payload_type": "runtime.session", "payload": {"event": {"kind": "model_completed", "model": "muse-spark-1.3-contributor", "duration_ms": 2000, "usage": {"input_tokens": 100, "output_tokens": 50, "reasoning_tokens": 0, "cache_read_tokens": 0, "cache_write_tokens": 0}}}}` + "\n"
	if err := os.WriteFile(mainLog, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	// DB containing sess-live (already in JSONL) AND sess-legacy (only in DB)
	dbPath := filepath.Join(dataDir, "session-index.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sessions (
			session_id TEXT PRIMARY KEY, model_id TEXT, workspace_root TEXT, prompt_count INTEGER,
			created_at_us INTEGER, updated_at_us INTEGER
		);
		INSERT INTO sessions VALUES ('sess-live', 'muse-spark-1.3-contributor', '/workspace/live', 1, 1780531500000000, 1780531502000000);
		INSERT INTO sessions VALUES ('sess-legacy', 'muse-spark-legacy', '/workspace/legacy', 4, 1780531600000000, 1780531605000000);
	`)
	if err != nil {
		t.Fatal(err)
	}

	sum, _, projects, err := LoadMuseData(dataDir)
	if err != nil {
		t.Fatalf("LoadMuseData failed: %v", err)
	}

	// sess-live (from JSONL, 150 tokens) + sess-legacy (from DB, 4 prompt tokens) -> 2 sessions, 154 tokens
	if sum.Sessions != 2 {
		t.Errorf("expected 2 sessions, got %d", sum.Sessions)
	}
	if sum.LifetimeTokens != 154 {
		t.Errorf("expected lifetime tokens 154, got %d", sum.LifetimeTokens)
	}
	if sum.ModelBreakdown["muse-spark-1.3-contributor"] != 1 {
		t.Errorf("expected 1 turn of muse-spark-1.3-contributor, got %v", sum.ModelBreakdown)
	}
	if sum.ModelBreakdown["muse-spark-legacy"] != 1 {
		t.Errorf("expected 1 turn of muse-spark-legacy, got %v", sum.ModelBreakdown)
	}

	// Check projects
	foundLive := false
	foundLegacy := false
	for _, p := range projects {
		if p.Project == "/workspace/live" {
			foundLive = true
		}
		if p.Project == "/workspace/legacy" {
			foundLegacy = true
		}
	}
	if !foundLive {
		t.Error("expected '/workspace/live' project in projects")
	}
	if !foundLegacy {
		t.Error("expected '/workspace/legacy' project in projects")
	}
}

func TestLoadMuseData_SourceSig(t *testing.T) {
	dataDir := t.TempDir()
	sessDir := filepath.Join(dataDir, "sessions", "2026", "10", "29", "sess-sig")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(sessDir, "session.jsonl")
	if err := os.WriteFile(logPath, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	info := ToolInfo{
		Name:    "Muse",
		DataDir: dataDir,
	}

	sig1 := GetToolSourceSig(thermal.ToolMuse, info, "")
	if sig1.Missing {
		t.Fatal("expected sig1 not to be missing")
	}
	if sig1.FileCount != 1 {
		t.Errorf("expected 1 file in sig1, got %d", sig1.FileCount)
	}

	// Append data to simulate active running session
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("line2\n")
	_ = f.Close()

	sig2 := GetToolSourceSig(thermal.ToolMuse, info, "")
	if sig2.TotalSize <= sig1.TotalSize {
		t.Errorf("expected sig2.TotalSize (%d) > sig1.TotalSize (%d)", sig2.TotalSize, sig1.TotalSize)
	}
}
