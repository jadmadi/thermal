// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestLoadHermesData_MissingFile(t *testing.T) {
	_, _, _, err := LoadHermesData("/nonexistent/path/state.db")
	if err == nil {
		t.Fatal("expected error for missing hermes database, got nil")
	}
}

func TestLoadHermesData_MissingTable(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE other_table (id TEXT);"); err != nil {
		t.Fatalf("failed to create dummy table: %v", err)
	}
	db.Close()

	_, _, _, err = LoadHermesData(dbPath)
	if err == nil {
		t.Fatal("expected error for missing sessions table, got nil")
	}
}

func TestLoadHermesData_BasicSessions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	createSQL := `
	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		source TEXT,
		model TEXT,
		cwd TEXT,
		started_at REAL,
		ended_at REAL,
		message_count INTEGER,
		input_tokens INTEGER,
		output_tokens INTEGER,
		cache_read_tokens INTEGER,
		cache_write_tokens INTEGER,
		reasoning_tokens INTEGER,
		estimated_cost_usd REAL
	);
	`
	if _, err := db.Exec(createSQL); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	now := time.Now().Truncate(24 * time.Hour)
	t1 := now.Add(10 * time.Hour).Unix()
	t1End := now.Add(10*time.Hour + 30*time.Minute).Unix()
	t2 := now.Add(14 * time.Hour).Unix()
	t2End := now.Add(14*time.Hour + 10*time.Minute).Unix()

	insertSQL := `
	INSERT INTO sessions (id, source, model, cwd, started_at, ended_at, message_count, input_tokens, output_tokens, cache_read_tokens, cache_write_tokens, reasoning_tokens, estimated_cost_usd)
	VALUES 
	('s1', 'cli', 'hermes-3-llama-3.1-405b', '/tmp/myproject', ?, ?, 4, 1000, 200, 5000, 100, 50, 0.05),
	('s2', 'cli', 'nous-hermes-2-mixtral', '/tmp/myproject', ?, ?, 2, 500, 100, 2000, 0, 0, 0.02);
	`
	if _, err := db.Exec(insertSQL, float64(t1), float64(t1End), float64(t2), float64(t2End)); err != nil {
		t.Fatalf("failed to insert sessions: %v", err)
	}
	db.Close()

	sum, daily, projects, err := LoadHermesData(dbPath)
	if err != nil {
		t.Fatalf("LoadHermesData failed: %v", err)
	}

	if sum.Sessions != 2 {
		t.Errorf("expected 2 sessions, got %d", sum.Sessions)
	}
	// Total tokens: s1 = 1000+200+5000+100+50 = 6350, s2 = 500+100+2000 = 2600. Sum = 8950
	if sum.LifetimeTokens != 8950 {
		t.Errorf("expected 8950 lifetime tokens, got %d", sum.LifetimeTokens)
	}
	if sum.InputTokens != 1500 {
		t.Errorf("expected 1500 input tokens, got %d", sum.InputTokens)
	}
	if sum.OutputTokens != 300 {
		t.Errorf("expected 300 output tokens, got %d", sum.OutputTokens)
	}
	if sum.CacheTokens != 7100 {
		t.Errorf("expected 7100 cache tokens, got %d", sum.CacheTokens)
	}
	if sum.ReasoningTokens != 50 {
		t.Errorf("expected 50 reasoning tokens, got %d", sum.ReasoningTokens)
	}
	if sum.Cost < 0.069 || sum.Cost > 0.071 {
		t.Errorf("expected cost ~0.07, got %f", sum.Cost)
	}
	if sum.LongestSessionMs != 30*60*1000 {
		t.Errorf("expected longest session 30m (%d ms), got %d", 30*60*1000, sum.LongestSessionMs)
	}
	if sum.ModelBreakdown["hermes-3-llama-3.1-405b"] != 1 {
		t.Errorf("expected hermes-3-llama-3.1-405b in breakdown, got %v", sum.ModelBreakdown)
	}
	if sum.ModelBreakdown["nous-hermes-2-mixtral"] != 1 {
		t.Errorf("expected nous-hermes-2-mixtral in breakdown, got %v", sum.ModelBreakdown)
	}

	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	expectedDay := thermal.LocalDay(now)
	if daily[0].Day != expectedDay {
		t.Errorf("expected day %s, got %s", expectedDay, daily[0].Day)
	}
	if daily[0].Turns != 6 {
		t.Errorf("expected 6 turns, got %d", daily[0].Turns)
	}
	if daily[0].Tokens != 8950 {
		t.Errorf("expected 8950 tokens in daily row, got %d", daily[0].Tokens)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project day row, got %d", len(projects))
	}
	if projects[0].Tokens != 8950 {
		t.Errorf("expected 8950 tokens in project day, got %d", projects[0].Tokens)
	}
}

func TestLoadHermesData_TimestampVariations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	createSQL := `
	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		created_at TEXT,
		model_id TEXT,
		tokens_input INTEGER,
		tokens_output INTEGER
	);
	`
	if _, err := db.Exec(createSQL); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// 1. ISO string, 2. SQLite date string, 3. Milliseconds string
	insertSQL := `
	INSERT INTO sessions (id, created_at, model_id, tokens_input, tokens_output)
	VALUES
	('s1', '2026-09-10T12:00:00Z', 'Hermes-3-70B', 100, 50),
	('s2', '2026-09-11 15:30:00', 'hermes-3-70b', 200, 80),
	('s3', '1789041600000', 'Nous-Hermes-Llama', 300, 120);
	`
	if _, err := db.Exec(insertSQL); err != nil {
		t.Fatalf("failed to insert sessions: %v", err)
	}
	db.Close()

	sum, daily, _, err := LoadHermesData(dbPath)
	if err != nil {
		t.Fatalf("LoadHermesData failed: %v", err)
	}

	if sum.Sessions != 3 {
		t.Errorf("expected 3 sessions, got %d", sum.Sessions)
	}
	// Models should canonicalize to lowercase
	if sum.ModelBreakdown["hermes-3-70b"] != 2 {
		t.Errorf("expected hermes-3-70b to have 2 sessions, got %d", sum.ModelBreakdown["hermes-3-70b"])
	}
	if sum.ModelBreakdown["nous-hermes-llama"] != 1 {
		t.Errorf("expected nous-hermes-llama to have 1 session, got %d", sum.ModelBreakdown["nous-hermes-llama"])
	}
	if len(daily) < 2 {
		t.Errorf("expected at least 2 distinct days, got %d", len(daily))
	}
}

func TestLoadHermesData_SessionModelUsage(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	createSQL := `
	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		started_at REAL,
		cwd TEXT,
		model TEXT
	);
	CREATE TABLE session_model_usage (
		session_id TEXT,
		model TEXT,
		input_tokens INTEGER,
		output_tokens INTEGER,
		reasoning_tokens INTEGER,
		cache_read_tokens INTEGER,
		cache_write_tokens INTEGER
	);
	`
	if _, err := db.Exec(createSQL); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	now := time.Now().Truncate(24 * time.Hour).Add(12 * time.Hour).Unix()
	if _, err := db.Exec(`INSERT INTO sessions (id, started_at, cwd, model) VALUES ('sess_multi', ?, '/tmp/proj', 'hermes-3')`, float64(now)); err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO session_model_usage (session_id, model, input_tokens, output_tokens, reasoning_tokens, cache_read_tokens, cache_write_tokens)
		VALUES 
		('sess_multi', 'hermes-router', 200, 50, 0, 1000, 0),
		('sess_multi', 'hermes-coder', 800, 400, 100, 4000, 200);
	`); err != nil {
		t.Fatalf("failed to insert smu rows: %v", err)
	}
	db.Close()

	sum, daily, projects, err := LoadHermesData(dbPath)
	if err != nil {
		t.Fatalf("LoadHermesData failed: %v", err)
	}

	if sum.Sessions != 1 {
		t.Errorf("expected 1 session, got %d", sum.Sessions)
	}
	// router = 200+50+1000 = 1250, coder = 800+400+100+4000+200 = 5500. Total = 6750
	if sum.LifetimeTokens != 6750 {
		t.Errorf("expected 6750 tokens, got %d", sum.LifetimeTokens)
	}
	if sum.InputTokens != 1000 {
		t.Errorf("expected 1000 input tokens, got %d", sum.InputTokens)
	}
	if sum.OutputTokens != 450 {
		t.Errorf("expected 450 output tokens, got %d", sum.OutputTokens)
	}
	if sum.ReasoningTokens != 100 {
		t.Errorf("expected 100 reasoning tokens, got %d", sum.ReasoningTokens)
	}
	if sum.CacheTokens != 5200 {
		t.Errorf("expected 5200 cache tokens, got %d", sum.CacheTokens)
	}

	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	if len(daily[0].Models) != 2 {
		t.Errorf("expected 2 models in daily row, got %d", len(daily[0].Models))
	}
	if daily[0].Models["hermes-router"].Input != 200 {
		t.Errorf("expected router input 200, got %d", daily[0].Models["hermes-router"].Input)
	}
	if daily[0].Models["hermes-coder"].Output != 400 {
		t.Errorf("expected coder output 400, got %d", daily[0].Models["hermes-coder"].Output)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project day row, got %d", len(projects))
	}
	if len(projects[0].Models) != 2 {
		t.Errorf("expected 2 models in project row, got %d", len(projects[0].Models))
	}
}
