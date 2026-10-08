// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
	"github.com/klauspost/compress/zstd"
)

func TestLoadZedData_MissingFile(t *testing.T) {
	_, _, _, err := LoadZedData("/nonexistent/path/threads.db")
	if err == nil {
		t.Fatal("expected error for missing zed database, got nil")
	}
}

func TestLoadZedData_MissingTable(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "threads.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE other_table (id TEXT);"); err != nil {
		t.Fatalf("failed to create dummy table: %v", err)
	}
	db.Close()

	_, _, _, err = LoadZedData(dbPath)
	if err == nil {
		t.Fatal("expected error for missing threads table, got nil")
	}
}

func TestLoadZedData_EmptyTable(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "threads.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	createSQL := `
	CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		summary TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		data_type TEXT NOT NULL,
		data BLOB NOT NULL,
		parent_id TEXT,
		worktree_branch TEXT,
		folder_paths TEXT,
		folder_paths_order TEXT,
		created_at TEXT
	);`
	if _, err := db.Exec(createSQL); err != nil {
		t.Fatalf("failed to create threads table: %v", err)
	}
	db.Close()

	summary, daily, projects, err := LoadZedData(dbPath)
	if err != nil {
		t.Fatalf("unexpected error for empty threads table: %v", err)
	}
	if summary.Sessions != 0 {
		t.Errorf("expected 0 sessions, got %d", summary.Sessions)
	}
	if summary.LifetimeTokens != 0 {
		t.Errorf("expected 0 tokens, got %d", summary.LifetimeTokens)
	}
	if len(daily) != 0 {
		t.Errorf("expected 0 daily rows, got %d", len(daily))
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 project days, got %d", len(projects))
	}
}

func compressZstdHelper(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	enc, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("failed to create zstd writer: %v", err)
	}
	if _, err := enc.Write(data); err != nil {
		t.Fatalf("failed to write zstd data: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("failed to close zstd writer: %v", err)
	}
	return buf.Bytes()
}

func TestLoadZedData_ValidSessions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "threads.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	createSQL := `
	CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		summary TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		data_type TEXT NOT NULL,
		data BLOB NOT NULL,
		parent_id TEXT,
		worktree_branch TEXT,
		folder_paths TEXT,
		folder_paths_order TEXT,
		created_at TEXT
	);`
	if _, err := db.Exec(createSQL); err != nil {
		t.Fatalf("failed to create threads table: %v", err)
	}

	// Payload 1: Thread with tokens and model
	p1 := zedThreadPayload{
		Title:   "Refactor parser",
		Version: "0.3.0",
		Model: zedModelPayload{
			Provider: "deepseek",
			Model:    "deepseek-v4-flash",
		},
		CumulativeTokenUsage: zedTokenUsagePayload{
			InputTokens:          1000,
			OutputTokens:         200,
			CacheReadInputTokens: 5000,
		},
		Messages: []json.RawMessage{
			json.RawMessage(`{"User":{"id":"u1"}}`),
			json.RawMessage(`{"Agent":{"content":[]}}`),
			json.RawMessage(`{"User":{"id":"u2"}}`),
			json.RawMessage(`{"Agent":{"content":[]}}`),
		},
	}
	p1Bytes, _ := json.Marshal(p1)
	p1Compressed := compressZstdHelper(t, p1Bytes)

	// Payload 2: Thread without tokens (e.g. Copilot Chat) on same day
	p2 := zedThreadPayload{
		Title:   "Fix bug",
		Version: "0.3.0",
		Model: zedModelPayload{
			Provider: "copilot_chat",
			Model:    "Grok Code Fast 1",
		},
		Messages: []json.RawMessage{
			json.RawMessage(`{"User":{"id":"u3"}}`),
			json.RawMessage(`{"Agent":{"content":[]}}`),
		},
	}
	p2Bytes, _ := json.Marshal(p2)
	p2Compressed := compressZstdHelper(t, p2Bytes)

	// Payload 3: Thread on different day with fallback project in snapshot
	p3 := zedThreadPayload{
		Title:   "Another day task",
		Version: "0.3.0",
		Model: zedModelPayload{
			Provider: "z.ai",
			Model:    "GLM-5.3-Flash",
		},
		CumulativeTokenUsage: zedTokenUsagePayload{
			InputTokens:          300,
			OutputTokens:         50,
			CacheReadInputTokens: 100,
		},
		InitialProjectSnapshot: zedInitialProjectSnapshot{
			Timestamp: "2026-10-02T10:00:00Z",
			WorktreeSnapshots: []struct {
				WorktreePath string `json:"worktree_path"`
			}{
				{WorktreePath: "/tmp/project-beta"},
			},
		},
		Messages: []json.RawMessage{
			json.RawMessage(`{"User":{"id":"u4"}}`),
			json.RawMessage(`{"Agent":{"content":[]}}`),
		},
	}
	p3Bytes, _ := json.Marshal(p3)
	p3Compressed := compressZstdHelper(t, p3Bytes)

	// Payload 4: Corrupted zstd payload to test warning handling
	corruptedBlob := []byte("not valid zstd data at all")

	insertStmt, err := db.Prepare(`
		INSERT INTO threads (id, summary, updated_at, data_type, data, parent_id, worktree_branch, folder_paths, folder_paths_order, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		t.Fatalf("failed to prepare insert: %v", err)
	}
	defer insertStmt.Close()

	projAlpha := filepath.Join(tmpDir, "project-alpha")
	projBeta := filepath.Join(tmpDir, "project-beta")
	if err := os.MkdirAll(filepath.Join(projAlpha, ".git"), 0755); err != nil {
		t.Fatalf("failed to create projAlpha .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projBeta, ".git"), 0755); err != nil {
		t.Fatalf("failed to create projBeta .git: %v", err)
	}

	// Insert 1: 2026-10-01, project projAlpha
	if _, err := insertStmt.Exec("t1", "Refactor parser", "2026-10-01T12:30:00Z", "zstd", p1Compressed, nil, "main", projAlpha, "0", "2026-10-01T12:00:00Z"); err != nil {
		t.Fatalf("failed to insert t1: %v", err)
	}

	// Insert 2: 2026-10-01, project projAlpha
	if _, err := insertStmt.Exec("t2", "Fix bug", "2026-10-01T14:15:00Z", "zstd", p2Compressed, nil, "main", projAlpha, "0", "2026-10-01T14:00:00Z"); err != nil {
		t.Fatalf("failed to insert t2: %v", err)
	}

	// Insert 3: 2026-10-02, no folder_paths, relies on snapshot worktree
	p3.InitialProjectSnapshot.WorktreeSnapshots[0].WorktreePath = projBeta
	p3Bytes, _ = json.Marshal(p3)
	p3Compressed = compressZstdHelper(t, p3Bytes)

	if _, err := insertStmt.Exec("t3", "Another day task", "2026-10-02T10:45:00Z", "zstd", p3Compressed, nil, "main", "", "", "2026-10-02T10:00:00Z"); err != nil {
		t.Fatalf("failed to insert t3: %v", err)
	}

	// Insert 4: Corrupted blob
	if _, err := insertStmt.Exec("t4", "Corrupt data", "2026-10-02T11:00:00Z", "zstd", corruptedBlob, nil, "main", projAlpha, "0", "2026-10-02T11:00:00Z"); err != nil {
		t.Fatalf("failed to insert t4: %v", err)
	}

	db.Close()

	summary, daily, projects, err := LoadZedData(dbPath)
	if err != nil {
		t.Fatalf("LoadZedData failed: %v", err)
	}

	// We expect 3 valid sessions, 1 corrupted warning
	if summary.Sessions != 3 {
		t.Errorf("expected 3 sessions, got %d", summary.Sessions)
	}
	if len(summary.Warnings) != 1 {
		t.Errorf("expected 1 warning for corrupted thread, got %d: %v", len(summary.Warnings), summary.Warnings)
	}

	// Total tokens:
	// t1: 1000 + 200 + 5000 = 6200
	// t2: 0
	// t3: 300 + 50 + 100 = 450
	// Total: 6650
	expectedTotalTokens := int64(6650)
	if summary.LifetimeTokens != expectedTotalTokens {
		t.Errorf("expected lifetime tokens %d, got %d", expectedTotalTokens, summary.LifetimeTokens)
	}
	if summary.InputTokens != 1300 {
		t.Errorf("expected input tokens 1300, got %d", summary.InputTokens)
	}
	if summary.OutputTokens != 250 {
		t.Errorf("expected output tokens 250, got %d", summary.OutputTokens)
	}
	if summary.CacheTokens != 5100 {
		t.Errorf("expected cache tokens 5100, got %d", summary.CacheTokens)
	}

	// Check model breakdown
	if summary.ModelBreakdown["deepseek-v4-flash"] != 1 {
		t.Errorf("expected 1 for deepseek-v4-flash, got %d", summary.ModelBreakdown["deepseek-v4-flash"])
	}
	if summary.ModelBreakdown["glm-5.3-flash"] != 1 {
		t.Errorf("expected 1 for glm-5.3-flash, got %d", summary.ModelBreakdown["glm-5.3-flash"])
	}

	// Check daily rows
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}

	// Day 1 (2026-10-01)
	d1 := daily[0]
	if d1.Day != "2026-10-01" {
		t.Errorf("expected day 2026-10-01, got %s", d1.Day)
	}
	if d1.Tokens != 6200 {
		t.Errorf("expected day 1 tokens 6200, got %d", d1.Tokens)
	}
	// Turns: 4 (t1) + 2 (t2) = 6
	if d1.Turns != 6 {
		t.Errorf("expected day 1 turns 6, got %d", d1.Turns)
	}

	// Day 2 (2026-10-02)
	d2 := daily[1]
	if d2.Day != "2026-10-02" {
		t.Errorf("expected day 2026-10-02, got %s", d2.Day)
	}
	if d2.Tokens != 450 {
		t.Errorf("expected day 2 tokens 450, got %d", d2.Tokens)
	}
	if d2.Turns != 2 {
		t.Errorf("expected day 2 turns 2, got %d", d2.Turns)
	}

	// Check project days
	if len(projects) != 2 {
		t.Fatalf("expected 2 project days, got %d", len(projects))
	}
	// p1 day 2026-10-01
	pDay1 := projects[0]
	if pDay1.Day != "2026-10-01" {
		t.Errorf("expected project day 2026-10-01, got %s", pDay1.Day)
	}
	if pDay1.Project != projAlpha {
		t.Errorf("expected project %s, got %s", projAlpha, pDay1.Project)
	}
	if pDay1.Tokens != 6200 {
		t.Errorf("expected project tokens 6200, got %d", pDay1.Tokens)
	}

	// p2 day 2026-10-02
	pDay2 := projects[1]
	if pDay2.Day != "2026-10-02" {
		t.Errorf("expected project day 2026-10-02, got %s", pDay2.Day)
	}
	if pDay2.Project != projBeta {
		t.Errorf("expected project %s, got %s", projBeta, pDay2.Project)
	}
	if pDay2.Tokens != 450 {
		t.Errorf("expected project tokens 450, got %d", pDay2.Tokens)
	}

	// Test cache hit: Second call should return cached results instantly
	summaryCached, dailyCached, projectsCached, err := LoadZedData(dbPath)
	if err != nil {
		t.Fatalf("cached LoadZedData failed: %v", err)
	}
	if summaryCached.LifetimeTokens != summary.LifetimeTokens {
		t.Errorf("cached tokens mismatch: %d vs %d", summaryCached.LifetimeTokens, summary.LifetimeTokens)
	}
	if len(dailyCached) != len(daily) {
		t.Errorf("cached daily rows length mismatch: %d vs %d", len(dailyCached), len(daily))
	}
	if len(projectsCached) != len(projects) {
		t.Errorf("cached project days length mismatch: %d vs %d", len(projectsCached), len(projects))
	}
}

func TestLoadZedData_RealDatabase(t *testing.T) {
	realPath := filepath.Join(thermal.HomeDir(), ".local", "share", "zed", "threads", "threads.db")
	if _, err := os.Stat(realPath); err != nil {
		t.Skip("skipping real database test: threads.db not present")
	}

	start := time.Now()
	summary, daily, projects, err := LoadZedData(realPath)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("LoadZedData on real database failed: %v", err)
	}
	if summary.Sessions == 0 {
		t.Errorf("expected sessions > 0 on real database, got 0")
	}
	if summary.LifetimeTokens == 0 {
		t.Errorf("expected lifetime tokens > 0 on real database, got 0")
	}
	if len(daily) == 0 {
		t.Errorf("expected daily rows > 0 on real database, got 0")
	}
	if len(projects) == 0 {
		t.Errorf("expected project days > 0 on real database, got 0")
	}

	t.Logf("Real Zed database loaded in %v: %d sessions, %d tokens, %d daily rows, %d project days",
		duration, summary.Sessions, summary.LifetimeTokens, len(daily), len(projects))

	// Second invocation should hit cache in <10ms
	startCached := time.Now()
	cachedSummary, _, _, err := LoadZedData(realPath)
	cachedDuration := time.Since(startCached)

	if err != nil {
		t.Fatalf("cached LoadZedData failed: %v", err)
	}
	if cachedSummary.LifetimeTokens != summary.LifetimeTokens {
		t.Errorf("cached token mismatch: %d vs %d", cachedSummary.LifetimeTokens, summary.LifetimeTokens)
	}
	t.Logf("Cached Zed database loaded in %v (target: <10ms)", cachedDuration)
}
