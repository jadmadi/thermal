// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

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

func TestCodexRolloutCache_UnchangedHit(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "state_5.sqlite")
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

	rolloutPath := filepath.Join(dataDir, "rollout.jsonl")
	rolloutContent := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":60,"cached_input_tokens":20,"output_tokens":30,"reasoning_output_tokens":10,"total_tokens":90}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(rolloutContent), 0o644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 90, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}
	db.Close()

	// First load: cold scan, writes cache
	sum1, daily1, _, err := LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData 1 error: %v", err)
	}

	// Verify cache file created
	cachePath, err := codexRolloutCachePath(dbPath)
	if err != nil {
		t.Fatalf("codexRolloutCachePath error: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file at %s: %v", cachePath, err)
	}

	// Second load: warm scan using cache
	sum2, daily2, _, err := LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData 2 error: %v", err)
	}

	if sum1.LifetimeTokens != sum2.LifetimeTokens || sum1.InputTokens != sum2.InputTokens || sum1.OutputTokens != sum2.OutputTokens {
		t.Fatalf("mismatch between cold (%+v) and warm (%+v)", sum1, sum2)
	}
	if len(daily1) != len(daily2) || daily1[0].Tokens != daily2[0].Tokens {
		t.Fatalf("daily mismatch between cold (%+v) and warm (%+v)", daily1, daily2)
	}
}

func TestCodexRolloutCache_AppendDelta(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "state_5.sqlite")
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

	rolloutPath := filepath.Join(dataDir, "rollout.jsonl")
	ev1 := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":50,"cached_input_tokens":10,"output_tokens":20,"reasoning_output_tokens":5,"total_tokens":70}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(ev1), 0o644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 70, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}

	// 1. Initial cold load
	_, _, _, err = LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData 1: %v", err)
	}

	// 2. Append new event to rollout and update DB thread tokens
	ev2 := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":40,"reasoning_output_tokens":10,"total_tokens":140}}}}` + "\n"
	f, err := os.OpenFile(rolloutPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open append: %v", err)
	}
	if _, err := f.WriteString(ev2); err != nil {
		t.Fatalf("append: %v", err)
	}
	f.Close()

	_, err = db.Exec(`UPDATE threads SET tokens_used = 140 WHERE id = 't1'`)
	if err != nil {
		t.Fatalf("update db: %v", err)
	}
	db.Close()

	// 3. Load with append delta
	sumDelta, dailyDelta, _, err := LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData delta: %v", err)
	}

	// 4. Invalidate cache and load cold to compare exact parity
	cachePath, _ := codexRolloutCachePath(dbPath)
	_ = os.Remove(cachePath)

	sumCold, dailyCold, _, err := LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData cold: %v", err)
	}

	if sumDelta.LifetimeTokens != sumCold.LifetimeTokens ||
		sumDelta.InputTokens != sumCold.InputTokens ||
		sumDelta.OutputTokens != sumCold.OutputTokens ||
		sumDelta.ReasoningTokens != sumCold.ReasoningTokens ||
		sumDelta.CacheTokens != sumCold.CacheTokens {
		t.Fatalf("delta vs cold mismatch:\ndelta: %+v\ncold:  %+v", sumDelta, sumCold)
	}
	if len(dailyDelta) != len(dailyCold) || dailyDelta[0].Tokens != dailyCold[0].Tokens {
		t.Fatalf("daily delta vs cold mismatch")
	}
}

func TestCodexRolloutCache_ShrinkFallback(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "state_5.sqlite")
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

	rolloutPath := filepath.Join(dataDir, "rollout.jsonl")
	ev1 := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":40,"reasoning_output_tokens":10,"total_tokens":140}}}}` + "\n"
	ev2 := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":200,"cached_input_tokens":40,"output_tokens":80,"reasoning_output_tokens":20,"total_tokens":280}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(ev1+ev2), 0o644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 280, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}

	// Warm cache
	_, _, _, err = LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData 1: %v", err)
	}

	// Truncate rollout file to just ev1 (shrinking file size)
	if err := os.WriteFile(rolloutPath, []byte(ev1), 0o644); err != nil {
		t.Fatalf("shrink rollout: %v", err)
	}
	_, err = db.Exec(`UPDATE threads SET tokens_used = 140 WHERE id = 't1'`)
	if err != nil {
		t.Fatalf("update db: %v", err)
	}
	db.Close()

	// Scan after shrink: should detect shrink and fall back to full scan
	sumShrink, _, _, err := LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData after shrink: %v", err)
	}
	if sumShrink.LifetimeTokens != 140 {
		t.Fatalf("expected lifetime 140 after shrink fallback, got %d", sumShrink.LifetimeTokens)
	}
}

func TestCodexRolloutCache_CorruptFallback(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "state_5.sqlite")
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

	rolloutPath := filepath.Join(dataDir, "rollout.jsonl")
	ev := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":50,"cached_input_tokens":10,"output_tokens":20,"reasoning_output_tokens":5,"total_tokens":70}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(ev), 0o644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 70, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}
	db.Close()

	// Pre-create corrupt cache file
	cachePath, err := codexRolloutCachePath(dbPath)
	if err != nil {
		t.Fatalf("cache path: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte(`{not valid json!`), 0o644); err != nil {
		t.Fatalf("write corrupt cache: %v", err)
	}

	// Should not panic, and should fall back gracefully
	sum, _, _, err := LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData corrupt cache: %v", err)
	}
	if sum.LifetimeTokens != 70 {
		t.Fatalf("expected 70 tokens, got %d", sum.LifetimeTokens)
	}
}

func TestCodexRolloutCache_Pruning(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "state_5.sqlite")
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

	r1 := filepath.Join(dataDir, "r1.jsonl")
	r2 := filepath.Join(dataDir, "r2.jsonl")
	ev := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":50,"cached_input_tokens":10,"output_tokens":20,"reasoning_output_tokens":5,"total_tokens":70}}}}` + "\n"
	_ = os.WriteFile(r1, []byte(ev), 0o644)
	_ = os.WriteFile(r2, []byte(ev), 0o644)

	_, _ = db.Exec(`INSERT INTO threads VALUES ('t1', 70, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, r1)
	_, _ = db.Exec(`INSERT INTO threads VALUES ('t2', 70, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, r2)

	// First load: caches both r1 and r2
	_, _, _, err = LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData 1: %v", err)
	}

	c := loadCodexRolloutCache(dbPath)
	if len(c.Entries) != 2 {
		t.Fatalf("expected 2 cache entries, got %d", len(c.Entries))
	}

	// Remove thread t2
	_, _ = db.Exec(`DELETE FROM threads WHERE id = 't2'`)
	db.Close()

	// Second load: prunes r2 from cache
	_, _, _, err = LoadCodexData(dataDir)
	if err != nil {
		t.Fatalf("LoadCodexData 2: %v", err)
	}

	c2 := loadCodexRolloutCache(dbPath)
	if len(c2.Entries) != 1 {
		t.Fatalf("expected 1 cache entry after pruning, got %d", len(c2.Entries))
	}
	if _, ok := c2.Entries[r1]; !ok {
		t.Fatalf("expected r1 to remain in cache")
	}
}
