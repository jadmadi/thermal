// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"path/filepath"
	"testing"

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
