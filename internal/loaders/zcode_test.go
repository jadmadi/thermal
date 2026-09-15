package loaders

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestLoadZCodeData_ModelUsage(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session (
			id TEXT PRIMARY KEY, time_created INTEGER, time_updated INTEGER,
			summary_additions INTEGER, summary_deletions INTEGER, summary_files INTEGER
		);
		CREATE TABLE model_usage (
			id TEXT PRIMARY KEY, session_id TEXT, status TEXT,
			started_at INTEGER, model_id TEXT, agent TEXT,
			input_tokens INTEGER, output_tokens INTEGER, reasoning_tokens INTEGER,
			cache_creation_input_tokens INTEGER, cache_read_input_tokens INTEGER,
			computed_total_tokens INTEGER
		);
		CREATE TABLE turn_usage (
			session_id TEXT, turn_id TEXT PRIMARY KEY, status TEXT, started_at INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	// Day 1: session s1, one completed turn with two model requests, plus
	// one cancelled request that must not count. Day 2: session s2.
	_, err = db.Exec(`
		INSERT INTO session VALUES ('s1', 1710504000000, 1710504060000, 10, 2, 1);
		INSERT INTO session VALUES ('s2', 1710590400000, 1710590460000, 0, 0, 0);
		INSERT INTO model_usage VALUES ('m1', 's1', 'completed', 1710504000000, 'GLM-5.3', 'build', 100, 50, 10, 5, 5, 170);
		INSERT INTO model_usage VALUES ('m2', 's1', 'completed', 1710504010000, 'GLM-5.3-Flash', 'build', 200, 100, 0, 0, 0, 300);
		INSERT INTO model_usage VALUES ('m3', 's1', 'cancelled', 1710504020000, 'GLM-5.3', 'build', 999, 999, 999, 999, 999, 999);
		INSERT INTO model_usage VALUES ('m4', 's2', 'completed', 1710590400000, 'deepseek-v4-flash', 'edit', 10, 5, 0, 0, 0, 15);
		INSERT INTO turn_usage VALUES ('s1', 't1', 'completed', 1710504000000);
		INSERT INTO turn_usage VALUES ('s1', 't2', 'cancelled', 1710504030000);
		INSERT INTO turn_usage VALUES ('s2', 't3', 'completed', 1710590400000);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, err := LoadZCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadZCodeData error: %v", err)
	}
	if sum.Sessions != 2 || sum.LifetimeTokens != 485 {
		t.Errorf("expected sessions=2, lifetime=485, got %d, %d", sum.Sessions, sum.LifetimeTokens)
	}
	if sum.InputTokens != 310 || sum.OutputTokens != 155 || sum.ReasoningTokens != 10 || sum.CacheTokens != 10 {
		t.Errorf("expected tokens 310/155/10/10, got %d/%d/%d/%d",
			sum.InputTokens, sum.OutputTokens, sum.ReasoningTokens, sum.CacheTokens)
	}
	if sum.LinesAdded != 10 || sum.LinesDeleted != 2 || sum.FilesTouched != 1 {
		t.Errorf("expected lines 10/2/1, got %d/%d/%d", sum.LinesAdded, sum.LinesDeleted, sum.FilesTouched)
	}
	if sum.ModelBreakdown["GLM-5.3"] != 1 || sum.ModelBreakdown["GLM-5.3-Flash"] != 1 || sum.ModelBreakdown["deepseek-v4-flash"] != 1 {
		t.Errorf("expected one completed request per model, got %v", sum.ModelBreakdown)
	}
	if sum.AgentBreakdown["build"] != 2 || sum.AgentBreakdown["edit"] != 1 {
		t.Errorf("expected agent breakdown build=2 edit=1, got %v", sum.AgentBreakdown)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 470 || daily[0].Turns != 1 {
		t.Errorf("expected first day tokens=470 turns=1, got %d/%d", daily[0].Tokens, daily[0].Turns)
	}
	if daily[1].Tokens != 15 || daily[1].Turns != 1 {
		t.Errorf("expected second day tokens=15 turns=1, got %d/%d", daily[1].Tokens, daily[1].Turns)
	}
}

func TestLoadZCodeData_MissingModelUsage(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE session (id TEXT PRIMARY KEY);`); err != nil {
		t.Fatalf("exec create error: %v", err)
	}
	if _, _, err := LoadZCodeData(dbPath); err == nil {
		t.Errorf("expected error for DB without model_usage table, got nil")
	}
}
