package loaders

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCodexData_NestedTokenTypes(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state_5.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE threads (
			id TEXT, tokens_used INTEGER, model TEXT, source TEXT,
			reasoning_effort TEXT, agent_role TEXT, created_at INTEGER,
			updated_at INTEGER, rollout_path TEXT, archived INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}

	// OpenAI-style usage: cached reads sit inside input_tokens and reasoning
	// sits inside output_tokens. The loader must subtract them so the stored
	// types are disjoint and add up to the recorded total.
	rolloutPath := filepath.Join(dir, "rollout.jsonl")
	rolloutContent := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":60,"cached_input_tokens":20,"output_tokens":30,"reasoning_output_tokens":10,"total_tokens":90}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(rolloutContent), 0644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	// tokens_used matches the rollout total, so no scaling applies.
	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 90, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}
	db.Close()

	sum, daily, _, err := LoadCodexData(dir)
	if err != nil {
		t.Fatalf("LoadCodexData error: %v", err)
	}
	if sum.LifetimeTokens != 90 {
		t.Errorf("expected lifetime 90, got %d", sum.LifetimeTokens)
	}
	if sum.InputTokens != 40 || sum.OutputTokens != 20 || sum.ReasoningTokens != 10 || sum.CacheTokens != 20 {
		t.Errorf("expected disjoint summary 40/20/10/20, got %d/%d/%d/%d",
			sum.InputTokens, sum.OutputTokens, sum.ReasoningTokens, sum.CacheTokens)
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	d := daily[0]
	if d.Input != 40 || d.Output != 20 || d.Reasoning != 10 || d.Cache != 20 {
		t.Errorf("expected disjoint daily 40/20/10/20, got %d/%d/%d/%d", d.Input, d.Output, d.Reasoning, d.Cache)
	}
	if d.Input+d.Output+d.Reasoning+d.Cache != d.Tokens {
		t.Errorf("daily types %d must add up to tokens %d",
			d.Input+d.Output+d.Reasoning+d.Cache, d.Tokens)
	}
	if m := d.Models["gpt-4o"]; m.Input != 40 || m.CacheRead != 20 || m.Output != 20 || m.Reasoning != 10 {
		t.Errorf("expected disjoint per-model counts, got %+v", m)
	}
}

func TestLoadCodexData_ReconciliationAndRollouts(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state_5.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE threads (
			id TEXT, tokens_used INTEGER, model TEXT, source TEXT,
			reasoning_effort TEXT, agent_role TEXT, created_at INTEGER,
			updated_at INTEGER, rollout_path TEXT, archived INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}

	rolloutPath := filepath.Join(dir, "rollout.jsonl")
	rolloutContent := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":60,"output_tokens":30,"reasoning_output_tokens":0,"cached_input_tokens":0}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(rolloutContent), 0644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	// Insert row with 100 tokens used (while rollout has 90 sum -> requires exact scaling and remainder check)
	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 100, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}
	db.Close()

	sum, daily, _, err := LoadCodexData(dir)
	if err != nil {
		t.Fatalf("LoadCodexData error: %v", err)
	}
	if sum.LifetimeTokens != 100 {
		t.Errorf("expected lifetime 100, got %d", sum.LifetimeTokens)
	}
	totalParts := sum.InputTokens + sum.OutputTokens + sum.ReasoningTokens + sum.CacheTokens
	if totalParts != sum.LifetimeTokens {
		t.Errorf("breakdown sum %d != lifetime %d", totalParts, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}
