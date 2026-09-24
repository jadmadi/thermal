// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestLiveCollector_UnchangedSourcesReuseSnapshot(t *testing.T) {
	td := t.TempDir()
	logs := filepath.Join(td, ".claude", "projects", "test-proj")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatalf("failed creating test dir: %v", err)
	}

	sessionFile := filepath.Join(logs, "s-1.jsonl")
	records := []map[string]any{
		{
			"type":      "assistant",
			"timestamp": "2026-09-24T12:00:00Z",
			"message": map[string]any{
				"id":    "m-1",
				"model": "claude-3-7-sonnet",
				"usage": map[string]any{
					"input_tokens":  100,
					"output_tokens": 50,
				},
			},
		},
	}
	var content string
	for _, r := range records {
		b, _ := json.Marshal(r)
		content += string(b) + "\n"
	}
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing session file: %v", err)
	}

	t.Setenv("HOME", td)

	opts := thermal.Options{
		Tool:       "claude",
		Offline:    true,
		NoEstimate: true,
	}

	collector := newLiveCollector(opts, nil)

	// First collection: should load data
	res1, prj1, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected error on first collect: %v", err)
	}
	if len(res1) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res1))
	}
	if res1[0].Summary.LifetimeTokens != 150 {
		t.Fatalf("expected 150 lifetime tokens, got %d", res1[0].Summary.LifetimeTokens)
	}

	// Verify cache entry exists
	cached, ok := collector.cache[thermal.ToolClaude]
	if !ok || cached == nil {
		t.Fatalf("expected cached entry for claude")
	}

	// Second collection immediately: unchanged source should return cached snapshot
	res2, prj2, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected error on second collect: %v", err)
	}
	if len(res2) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res2))
	}
	if res2[0].Summary.LifetimeTokens != 150 {
		t.Fatalf("expected 150 lifetime tokens, got %d", res2[0].Summary.LifetimeTokens)
	}
	if len(prj1) != len(prj2) {
		t.Fatalf("project count mismatch: %d vs %d", len(prj1), len(prj2))
	}
}

func TestLiveCollector_DetectsFileModifications(t *testing.T) {
	td := t.TempDir()
	logs := filepath.Join(td, ".claude", "projects", "test-proj")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatalf("failed creating test dir: %v", err)
	}

	sessionFile := filepath.Join(logs, "s-1.jsonl")
	rec1 := map[string]any{
		"type":      "assistant",
		"timestamp": "2026-09-24T12:00:00Z",
		"message": map[string]any{
			"id":    "m-1",
			"model": "claude-3-7-sonnet",
			"usage": map[string]any{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		},
	}
	b1, _ := json.Marshal(rec1)
	if err := os.WriteFile(sessionFile, append(b1, '\n'), 0o644); err != nil {
		t.Fatalf("failed writing session file: %v", err)
	}

	t.Setenv("HOME", td)

	opts := thermal.Options{
		Tool:       "claude",
		Offline:    true,
		NoEstimate: true,
	}

	collector := newLiveCollector(opts, nil)
	res1, _, err := collector.Collect()
	if err != nil || len(res1) == 0 {
		t.Fatalf("initial collection failed: %v", err)
	}
	if res1[0].Summary.LifetimeTokens != 150 {
		t.Fatalf("expected 150 tokens, got %d", res1[0].Summary.LifetimeTokens)
	}

	// Advance time past the 200ms directory check throttling window
	if cached, ok := collector.cache[thermal.ToolClaude]; ok {
		cached.lastCheck = time.Now().Add(-500 * time.Millisecond)
	}

	// Append a second turn
	rec2 := map[string]any{
		"type":      "assistant",
		"timestamp": "2026-09-24T12:05:00Z",
		"message": map[string]any{
			"id":    "m-2",
			"model": "claude-3-7-sonnet",
			"usage": map[string]any{
				"input_tokens":  200,
				"output_tokens": 100,
			},
		},
	}
	b2, _ := json.Marshal(rec2)
	f, err := os.OpenFile(sessionFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("failed opening file for append: %v", err)
	}
	_, _ = f.Write(append(b2, '\n'))
	_ = f.Close()

	// Update mtime explicitly so it reflects change on coarse filesystem timestamps
	now := time.Now().Add(time.Second)
	_ = os.Chtimes(sessionFile, now, now)

	// Next collect should detect modification
	res2, _, err := collector.Collect()
	if err != nil || len(res2) == 0 {
		t.Fatalf("second collection failed: %v", err)
	}
	if res2[0].Summary.LifetimeTokens != 450 {
		t.Fatalf("expected 450 tokens after append, got %d", res2[0].Summary.LifetimeTokens)
	}
}

func TestLiveCollector_SQLiteSourceTracking(t *testing.T) {
	td := t.TempDir()
	dbPath := filepath.Join(td, "state.db")

	// Create dummy file for SQLite stat test
	if err := os.WriteFile(dbPath, []byte("sqlite header"), 0o644); err != nil {
		t.Fatalf("failed creating db file: %v", err)
	}

	info := loaders.ToolInfo{
		Name:   "Nous Hermes",
		DBPath: dbPath,
	}

	sig1 := getLiveToolSourceSig(thermal.ToolHermes, info, "")
	if sig1.missing {
		t.Fatalf("expected sig1 not missing")
	}
	if sig1.fileCount != 1 || sig1.totalSize == 0 {
		t.Fatalf("unexpected sig1: %+v", sig1)
	}

	// Modify file
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(dbPath, []byte("sqlite header updated data"), 0o644); err != nil {
		t.Fatalf("failed modifying db file: %v", err)
	}

	sig2 := getLiveToolSourceSig(thermal.ToolHermes, info, "")
	if sig2.totalSize == sig1.totalSize || sig2.maxNano == sig1.maxNano {
		t.Fatalf("expected sig2 to differ from sig1 after write: %+v vs %+v", sig1, sig2)
	}
}

func TestLiveCollector_MissingSourceHandledGracefully(t *testing.T) {
	opts := thermal.Options{
		Tool: "nonexistent-tool-xyz",
	}
	collector := newLiveCollector(opts, nil)
	_, _, err := collector.Collect()
	if err == nil {
		t.Fatalf("expected error for nonexistent tool")
	}
}
