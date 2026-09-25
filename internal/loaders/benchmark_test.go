// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func createDevinBenchFixture(b *testing.B, dir string, sessionCount, msgCount int) string {
	b.Helper()
	dbPath := filepath.Join(dir, "sessions.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		b.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER,
			last_activity_at INTEGER,
			hidden INTEGER,
			working_directory TEXT,
			model TEXT
		);
		CREATE TABLE IF NOT EXISTS message_nodes (
			row_id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			created_at INTEGER,
			chat_message TEXT
		);
		CREATE TABLE IF NOT EXISTS prompt_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			prompt TEXT,
			created_at INTEGER,
			updated_at INTEGER
		);
	`)
	if err != nil {
		b.Fatalf("create devin schema: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		b.Fatalf("begin tx: %v", err)
	}
	sessStmt, err := tx.Prepare("INSERT INTO sessions VALUES (?, ?, ?, 0, ?, ?)")
	if err != nil {
		b.Fatalf("prep sess: %v", err)
	}
	defer sessStmt.Close()

	msgStmt, err := tx.Prepare("INSERT INTO message_nodes (session_id, created_at, chat_message) VALUES (?, ?, ?)")
	if err != nil {
		b.Fatalf("prep msg: %v", err)
	}
	defer msgStmt.Close()

	baseTime := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC).Unix()
	for i := 0; i < sessionCount; i++ {
		sessID := fmt.Sprintf("bench-sess-%d", i)
		ts := baseTime + int64(i*3600)
		if _, err := sessStmt.Exec(sessID, ts, ts, "/tmp/repo", "claude-3-5-sonnet"); err != nil {
			b.Fatalf("insert sess: %v", err)
		}

		for m := 0; m < msgCount; m++ {
			msgTs := ts + int64(m*60)
			msgJSON := `{"role":"assistant","metadata":{"metrics":{"input_tokens":500,"output_tokens":100,"cache_creation_tokens":50,"cache_read_tokens":200}}}`
			if _, err := msgStmt.Exec(sessID, msgTs, msgJSON); err != nil {
				b.Fatalf("insert msg: %v", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		b.Fatalf("commit devin: %v", err)
	}

	return dbPath
}

func BenchmarkDevin_Cold(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := createDevinBenchFixture(b, tmpDir, 25, 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Invalidate cache for cold measurement
		cPath, _ := devinCachePath(CanonicalDatabasePath(dbPath))
		_ = os.Remove(cPath)
		b.StartTimer()

		_, _, _, err := LoadDevinData(dbPath)
		if err != nil {
			b.Fatalf("LoadDevinData cold: %v", err)
		}
	}
}

func BenchmarkDevin_Warm(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := createDevinBenchFixture(b, tmpDir, 25, 20)

	// Warm the cache once
	if _, _, _, err := LoadDevinData(dbPath); err != nil {
		b.Fatalf("warmup LoadDevinData: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum, _, _, err := LoadDevinData(dbPath)
		if err != nil {
			b.Fatalf("LoadDevinData warm: %v", err)
		}
		if sum.LifetimeTokens == 0 {
			b.Fatalf("expected non-zero tokens")
		}
	}
}

func BenchmarkDevin_AppendDelta(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := createDevinBenchFixture(b, tmpDir, 20, 10)

	// Warm the cache
	if _, _, _, err := LoadDevinData(dbPath); err != nil {
		b.Fatalf("warmup LoadDevinData: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		b.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Append 1 new message node
		msgTs := time.Now().Unix()
		msgJSON := `{"role":"assistant","metadata":{"metrics":{"input_tokens":50,"output_tokens":20,"cache_creation_tokens":0,"cache_read_tokens":10}}}`
		_, err := db.Exec(`INSERT INTO message_nodes (session_id, created_at, chat_message) VALUES ('bench-sess-0', ?, ?)`, msgTs, msgJSON)
		if err != nil {
			b.Fatalf("append message: %v", err)
		}
		b.StartTimer()

		_, _, _, err = LoadDevinData(dbPath)
		if err != nil {
			b.Fatalf("LoadDevinData delta: %v", err)
		}
	}
}

func createCodexBenchFixture(b *testing.B, dir string, threadCount, rolloutsPerThread int) string {
	b.Helper()
	dbPath := filepath.Join(dir, "state_5.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		b.Fatalf("open codex db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS threads (
			id TEXT PRIMARY KEY,
			tokens_used INTEGER,
			model TEXT,
			source TEXT,
			reasoning_effort TEXT,
			agent_role TEXT,
			created_at INTEGER,
			updated_at INTEGER,
			rollout_path TEXT,
			archived INTEGER,
			cwd TEXT
		);
	`)
	if err != nil {
		b.Fatalf("create codex schema: %v", err)
	}

	rolloutDir := filepath.Join(dir, "sessions")
	_ = os.MkdirAll(rolloutDir, 0o755)

	baseTime := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC).Unix()
	tx, err := db.Begin()
	if err != nil {
		b.Fatalf("begin tx: %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO threads VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)")
	if err != nil {
		b.Fatalf("prep stmt: %v", err)
	}
	defer stmt.Close()

	for i := 0; i < threadCount; i++ {
		tID := fmt.Sprintf("thread-%d", i)
		rPath := filepath.Join(rolloutDir, fmt.Sprintf("%s.jsonl", tID))

		// Write rollout JSONL
		rContent := ""
		for r := 0; r < rolloutsPerThread; r++ {
			rContent += fmt.Sprintf(`{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":500,"cached_input_tokens":100,"output_tokens":200,"reasoning_output_tokens":50,"total_tokens":700}}}}` + "\n")
		}
		if err := os.WriteFile(rPath, []byte(rContent), 0o644); err != nil {
			b.Fatalf("write rollout: %v", err)
		}

		totalTok := int64(700 * rolloutsPerThread)
		ts := baseTime + int64(i*3600)
		if _, err := stmt.Exec(tID, totalTok, "gpt-4o", "cli", "medium", "developer", ts, ts, rPath, "/tmp/codex-repo"); err != nil {
			b.Fatalf("insert thread: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		b.Fatalf("commit codex: %v", err)
	}

	return dir
}

func BenchmarkCodex_Cold(b *testing.B) {
	tmpDir := b.TempDir()
	dataDir := createCodexBenchFixture(b, tmpDir, 20, 5)
	stateDB := filepath.Join(dataDir, "state_5.sqlite")
	cPath, _ := codexRolloutCachePath(stateDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if cPath != "" {
			_ = os.Remove(cPath)
		}
		b.StartTimer()
		_, _, _, err := LoadCodexData(dataDir)
		if err != nil {
			b.Fatalf("LoadCodexData cold: %v", err)
		}
	}
}

func BenchmarkCodex_Warm(b *testing.B) {
	tmpDir := b.TempDir()
	dataDir := createCodexBenchFixture(b, tmpDir, 20, 5)

	// Pre-warm cache
	_, _, _, err := LoadCodexData(dataDir)
	if err != nil {
		b.Fatalf("LoadCodexData warmup: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := LoadCodexData(dataDir)
		if err != nil {
			b.Fatalf("LoadCodexData warm: %v", err)
		}
	}
}

func createOpenCodeBenchFixture(b *testing.B, dir string, sessionCount int) string {
	b.Helper()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		b.Fatalf("open opencode: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS session_v2 (
			id TEXT PRIMARY KEY,
			tokens_input INTEGER,
			tokens_output INTEGER,
			tokens_reasoning INTEGER,
			tokens_cache_read INTEGER,
			tokens_cache_write INTEGER,
			cost REAL,
			summary_additions INTEGER,
			summary_deletions INTEGER,
			summary_files INTEGER,
			agent TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			model TEXT
		);
		CREATE TABLE IF NOT EXISTS session (
			id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			tokens_input INTEGER,
			tokens_output INTEGER,
			tokens_reasoning INTEGER,
			tokens_cache_read INTEGER,
			tokens_cache_write INTEGER,
			cost REAL,
			summary_additions INTEGER,
			summary_deletions INTEGER,
			summary_files INTEGER,
			agent TEXT
		);
	`)
	if err != nil {
		b.Fatalf("create opencode schema: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		b.Fatalf("begin tx: %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO session_v2 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'code', ?, ?, '{\"id\":\"claude-3-5-sonnet\"}')")
	if err != nil {
		b.Fatalf("prep stmt: %v", err)
	}
	defer stmt.Close()

	baseTime := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC).UnixMilli()
	for i := 0; i < sessionCount; i++ {
		sID := fmt.Sprintf("oc-sess-%d", i)
		ts := baseTime + int64(i*3600000)
		if _, err := stmt.Exec(sID, 1000, 200, 50, 400, 100, 0.05, 10, 2, 3, ts, ts); err != nil {
			b.Fatalf("insert oc: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		b.Fatalf("commit oc: %v", err)
	}
	return dbPath
}

func BenchmarkOpenCode_Scan(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := createOpenCodeBenchFixture(b, tmpDir, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := LoadOpenCodeData(dbPath)
		if err != nil {
			b.Fatalf("LoadOpenCodeData: %v", err)
		}
	}
}

func createClaudeBenchFixture(b *testing.B, dir string, projectCount, sessionsPerProject int) string {
	b.Helper()
	projectsDir := filepath.Join(dir, "projects")
	_ = os.MkdirAll(projectsDir, 0o755)

	ts := "2026-09-24T12:00:00Z"
	for p := 0; p < projectCount; p++ {
		pDir := filepath.Join(projectsDir, fmt.Sprintf("proj-%d", p))
		_ = os.MkdirAll(pDir, 0o755)
		for s := 0; s < sessionsPerProject; s++ {
			sFile := filepath.Join(pDir, fmt.Sprintf("sess-%d.jsonl", s))
			content := fmt.Sprintf(`{"type":"assistant","timestamp":"%s","cwd":"/tmp/claude-repo","message":{"id":"msg-1","model":"claude-3-5-sonnet","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000}}}`+"\n", ts)
			_ = os.WriteFile(sFile, []byte(content), 0o644)
		}
	}
	return dir
}

func BenchmarkClaude_Scan(b *testing.B) {
	tmpDir := b.TempDir()
	dataDir := createClaudeBenchFixture(b, tmpDir, 5, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := LoadClaudeData(dataDir)
		if err != nil {
			b.Fatalf("LoadClaudeData: %v", err)
		}
	}
}
