// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// writePricingCatalog installs a deterministic two-model pricing catalog so
// every run can operate in --offline mode.
func writePricingCatalog(home string) error {
	cacheDir := filepath.Join(home, ".cache", "thermal")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	catalog := map[string]any{
		"version":   1,
		"fetchedAt": time.Now().UTC().Format(time.RFC3339),
		"source":    "https://models.dev/api.json",
		"models": map[string]any{
			"claude-3-5-sonnet": map[string]float64{"input": 3.0, "output": 15.0, "cacheRead": 0.3, "cacheWrite": 3.75},
			"gpt-4o":            map[string]float64{"input": 2.5, "output": 10.0, "cacheRead": 1.25},
		},
	}
	data, err := json.Marshal(catalog)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cacheDir, "pricing.json"), data, 0o644)
}

// createSmallFixture builds the deterministic five-source workload whose
// per-source totals are asserted exactly.
func createSmallFixture(workDir string) (fixtureDims, error) {
	home := filepath.Join(workDir, "home-small")
	if err := writePricingCatalog(home); err != nil {
		return fixtureDims{}, fmt.Errorf("pricing catalog: %w", err)
	}

	now := time.Now().UTC()
	base := now.Add(-4 * time.Hour).Unix()

	// Devin: 25 sessions × 20 message nodes × 850 tokens = 425,000.
	devinDB := filepath.Join(home, ".local", "share", "devin", "cli", "sessions.db")
	if err := os.MkdirAll(filepath.Dir(devinDB), 0o755); err != nil {
		return fixtureDims{}, err
	}
	db, err := sql.Open("sqlite", devinDB)
	if err != nil {
		return fixtureDims{}, err
	}
	if _, err := db.Exec(`
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER, working_directory TEXT, model TEXT);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, session_id TEXT, created_at INTEGER, chat_message TEXT);
		CREATE TABLE prompt_history (id INTEGER PRIMARY KEY AUTOINCREMENT, session_id TEXT, prompt TEXT, created_at INTEGER, updated_at INTEGER);
	`); err != nil {
		db.Close()
		return fixtureDims{}, err
	}
	msgJSON := `{"role":"assistant","metadata":{"metrics":{"input_tokens":500,"output_tokens":100,"cache_creation_tokens":50,"cache_read_tokens":200}}}`
	for i := 1; i <= 25; i++ {
		ts := base + int64(i)*3600
		if _, err := db.Exec(`INSERT INTO sessions VALUES (?, ?, ?, 0, ?, 'claude-3-5-sonnet')`,
			fmt.Sprintf("devin-sess-%d", i), ts, ts, filepath.Join(home, "projects", "thermal-repo")); err != nil {
			db.Close()
			return fixtureDims{}, err
		}
		for m := 1; m <= 20; m++ {
			if _, err := db.Exec(`INSERT INTO message_nodes (session_id, created_at, chat_message) VALUES (?, ?, ?)`,
				fmt.Sprintf("devin-sess-%d", i), ts+int64(m)*60, msgJSON); err != nil {
				db.Close()
				return fixtureDims{}, err
			}
		}
	}
	db.Close()

	// Codex: 20 threads × 3,500 recorded tokens = 70,000, with rollout logs.
	codexDir := filepath.Join(home, ".codex")
	codexDB := filepath.Join(codexDir, "state_5.sqlite")
	if err := os.MkdirAll(filepath.Join(codexDir, "sessions"), 0o755); err != nil {
		return fixtureDims{}, err
	}
	db, err = sql.Open("sqlite", codexDB)
	if err != nil {
		return fixtureDims{}, err
	}
	if _, err := db.Exec(`
		CREATE TABLE threads (id TEXT PRIMARY KEY, tokens_used INTEGER, model TEXT, source TEXT, reasoning_effort TEXT, agent_role TEXT, created_at INTEGER, updated_at INTEGER, rollout_path TEXT, archived INTEGER, cwd TEXT);
	`); err != nil {
		db.Close()
		return fixtureDims{}, err
	}
	rolloutLine := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":500,"cached_input_tokens":100,"output_tokens":200,"reasoning_output_tokens":50,"total_tokens":700}}}}` + "\n"
	for i := 1; i <= 20; i++ {
		rollout := filepath.Join(codexDir, "sessions", fmt.Sprintf("thread_%d.jsonl", i))
		content := ""
		for r := 0; r < 5; r++ {
			content += rolloutLine
		}
		if err := os.WriteFile(rollout, []byte(content), 0o644); err != nil {
			db.Close()
			return fixtureDims{}, err
		}
		ts := base + int64(i)*3600
		if _, err := db.Exec(`INSERT INTO threads VALUES (?, 3500, 'gpt-4o', 'cli', 'medium', 'developer', ?, ?, ?, 0, ?)`,
			fmt.Sprintf("codex-thread-%d", i), ts, ts, rollout, filepath.Join(home, "projects", "thermal-repo")); err != nil {
			db.Close()
			return fixtureDims{}, err
		}
	}
	db.Close()

	// OpenCode: 50 sessions × 1,750 tokens = 87,500.
	opencodeDB := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
	if err := os.MkdirAll(filepath.Dir(opencodeDB), 0o755); err != nil {
		return fixtureDims{}, err
	}
	db, err = sql.Open("sqlite", opencodeDB)
	if err != nil {
		return fixtureDims{}, err
	}
	if _, err := db.Exec(`
		CREATE TABLE session_v2 (id TEXT PRIMARY KEY, tokens_input INTEGER, tokens_output INTEGER, tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER, cost REAL, summary_additions INTEGER, summary_deletions INTEGER, summary_files INTEGER, agent TEXT, time_created INTEGER, time_updated INTEGER, model TEXT);
		CREATE TABLE session (id TEXT, time_created INTEGER, time_updated INTEGER, tokens_input INTEGER, tokens_output INTEGER, tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER, cost REAL, summary_additions INTEGER, summary_deletions INTEGER, summary_files INTEGER, agent TEXT);
	`); err != nil {
		db.Close()
		return fixtureDims{}, err
	}
	for i := 1; i <= 50; i++ {
		ts := (base + int64(i)*1800) * 1000
		if _, err := db.Exec(`INSERT INTO session_v2 VALUES (?, 1000, 200, 50, 400, 100, 0.05, 10, 2, 3, 'code', ?, ?, '{"id":"claude-3-5-sonnet"}')`,
			fmt.Sprintf("oc-sess-%d", i), ts, ts); err != nil {
			db.Close()
			return fixtureDims{}, err
		}
	}
	db.Close()

	// Claude: 30 files × 9,700 tokens = 291,000.
	claudeDir := filepath.Join(home, ".claude", "projects", "thermal")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fixtureDims{}, err
	}
	for i := 1; i <= 30; i++ {
		path := filepath.Join(claudeDir, fmt.Sprintf("sess_%d.jsonl", i))
		if err := os.WriteFile(path, []byte(claudeLine(fmt.Sprintf("claude-msg-%d", i), now)), 0o644); err != nil {
			return fixtureDims{}, err
		}
	}

	// CodeWhale: 20 files × 50,000 tokens = 1,000,000.
	codewhaleDir := filepath.Join(home, ".codewhale", "sessions")
	if err := os.MkdirAll(codewhaleDir, 0o755); err != nil {
		return fixtureDims{}, err
	}
	for i := 1; i <= 20; i++ {
		session := map[string]any{
			"session_id": fmt.Sprintf("whale-sess-%d", i),
			"metadata": map[string]any{
				"created_at":    now.Add(-time.Duration(20-i) * 30 * time.Minute).Format(time.RFC3339),
				"updated_at":    now.Add(-time.Duration(20-i) * 30 * time.Minute).Format(time.RFC3339),
				"message_count": 10,
				"total_tokens":  50000,
				"cost":          map[string]float64{"session_cost_usd": 0.15},
				"model":         "claude-3-5-sonnet",
				"mode":          "chat",
				"workspace":     filepath.Join(home, "projects", "thermal-repo"),
			},
		}
		data, err := json.Marshal(session)
		if err != nil {
			return fixtureDims{}, err
		}
		if err := os.WriteFile(filepath.Join(codewhaleDir, fmt.Sprintf("session_%d.json", i)), data, 0o644); err != nil {
			return fixtureDims{}, err
		}
	}

	return fixtureDims{
		Home:           home,
		Sources:        5,
		Files:          73,
		ExpectedTokens: baselineTotal,
	}, nil
}

// createLargeFixture builds 1000 Claude session files with 30 records each
// (291,000,000 tokens) to exercise directory-scale scans and the live monitor.
func createLargeFixture(workDir string) (fixtureDims, error) {
	home := filepath.Join(workDir, "home-large")
	if err := writePricingCatalog(home); err != nil {
		return fixtureDims{}, fmt.Errorf("pricing catalog: %w", err)
	}
	claudeDir := filepath.Join(home, ".claude", "projects", "large")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fixtureDims{}, err
	}
	now := time.Now().UTC()
	var content string
	for r := 0; r < largeClaudeRecords; r++ {
		content += claudeLine(fmt.Sprintf("large-msg-%d", r), now.Add(-time.Duration(r)*time.Minute))
	}
	for i := 0; i < largeClaudeFiles; i++ {
		path := filepath.Join(claudeDir, fmt.Sprintf("sess_%d.jsonl", i))
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fixtureDims{}, err
		}
	}
	return fixtureDims{
		Home:           home,
		Sources:        1,
		Files:          largeClaudeFiles,
		ExpectedTokens: int64(largeClaudeFiles) * largeClaudeRecords * claudeRecordTokens,
	}, nil
}

// appendLargeClaudeFile adds one fresh session file to the large fixture and
// returns the tokens it contributes.
func appendLargeClaudeFile(home string) (int64, error) {
	claudeDir := filepath.Join(home, ".claude", "projects", "large")
	path := filepath.Join(claudeDir, fmt.Sprintf("append_%d.jsonl", time.Now().UnixNano()))
	if err := os.WriteFile(path, []byte(claudeLine("append-msg", time.Now().UTC())), 0o644); err != nil {
		return 0, err
	}
	return claudeRecordTokens, nil
}

// claudeLine renders one assistant usage record for Claude transcripts.
func claudeLine(messageID string, ts time.Time) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"cwd":"/tmp/thermal-bench-repo","message":{"id":%q,"model":"claude-3-5-sonnet","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":200,"cache_read_input_tokens":8000}}}`+"\n",
		ts.Format(time.RFC3339), messageID)
}
