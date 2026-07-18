package loaders

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	_ "modernc.org/sqlite"
)

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

	sum, daily, err := LoadCodexData(dir)
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
