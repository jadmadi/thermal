// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

func TestLoadOpenCodeData_PreAggregatedSchema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session (
			id TEXT, time_created INTEGER, time_updated INTEGER,
			tokens_input INTEGER, tokens_output INTEGER, tokens_reasoning INTEGER,
			tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO session VALUES (
			's1', 1710504000000, 1710504060000, 100, 50, 10, 5, 5,
			0.05, 20, 5, 2, 'code'
		);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, _, err := LoadOpenCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadOpenCodeData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 170 {
		t.Errorf("expected sessions=1, lifetime=170, got %d, %d", sum.Sessions, sum.LifetimeTokens)
	}
	if sum.Cost != 0.05 {
		t.Errorf("expected cost=0.05, got %f", sum.Cost)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}

func TestLoadOpenCodeData_ModelColumnShapes(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	// session_v2 only, so the legacy union arm is skipped. The model column
	// holds a JSON object on current schemas, a plain string on some forks,
	// and NULL on older rows. None of those may fail the loader.
	_, err = db.Exec(`
		CREATE TABLE session_v2 (
			id TEXT PRIMARY KEY, tokens_input INTEGER, tokens_output INTEGER,
			tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT,
			time_created INTEGER, time_updated INTEGER, model TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO session_v2 VALUES ('s1', 100, 50, 0, 0, 0, 0.01, 0, 0, 0, 'code', 1710504000000, 1710504060000, '{"id":"json-model","providerID":"p"}');
		INSERT INTO session_v2 VALUES ('s2', 10, 5, 0, 0, 0, 0.01, 0, 0, 0, 'code', 1710504000000, 1710504060000, 'plain-model');
		INSERT INTO session_v2 VALUES ('s3', 1, 1, 0, 0, 0, 0.01, 0, 0, 0, 'code', 1710504000000, 1710504060000, NULL);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, _, err := LoadOpenCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadOpenCodeData error: %v", err)
	}
	if sum.Sessions != 3 {
		t.Errorf("expected sessions=3, got %d", sum.Sessions)
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	models := daily[0].Models
	if _, ok := models["json-model"]; !ok {
		t.Errorf("expected the JSON model id to be extracted, got %v", models)
	}
	if _, ok := models["plain-model"]; !ok {
		t.Errorf("expected a plain string model to pass through, got %v", models)
	}
	if _, ok := models[""]; ok {
		t.Errorf("a NULL model must not create an empty model key, got %v", models)
	}
}

func TestLoadOpenCodeData_ProjectAttribution(t *testing.T) {
	// A real repository marker so two subdirectories collapse into one
	// project root.
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub1 := filepath.Join(repo, "apps", "web")
	sub2 := filepath.Join(repo, "packages", "api")

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session_v2 (
			id TEXT PRIMARY KEY, tokens_input INTEGER, tokens_output INTEGER,
			tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT,
			time_created INTEGER, time_updated INTEGER, model TEXT, directory TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO session_v2 VALUES ('s1', 100, 50, 0, 0, 0, 0.10, 0, 0, 0, 'code', 1710504000000, 1710504060000, '{"id":"m1"}', ?);
		INSERT INTO session_v2 VALUES ('s2', 200, 60, 0, 0, 0, 0.20, 0, 0, 0, 'code', 1710504000000, 1710504060000, '{"id":"m2"}', ?);
	`, sub1, sub2)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, projects, err := LoadOpenCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadOpenCodeData error: %v", err)
	}
	if sum.Sessions != 2 || sum.LifetimeTokens != 410 {
		t.Errorf("expected sessions=2 lifetime=410, got %d/%d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(projects) != 1 {
		t.Fatalf("expected both subdirectories to fold into one project, got %d: %+v", len(projects), projects)
	}
	p := projects[0]
	if p.Project != repo {
		t.Errorf("project = %q, want the repository root %q", p.Project, repo)
	}
	if p.Tokens != 410 || p.Input != 300 || p.Output != 110 {
		t.Errorf("project totals = %+v", p)
	}
	if diff := p.Cost - 0.30; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("project cost = %v, want 0.30 (stored)", p.Cost)
	}
	if p.Turns != 2 {
		t.Errorf("project turns = %d, want 2", p.Turns)
	}
	if sum.ModelBreakdown["m1"] != 1 || sum.ModelBreakdown["m2"] != 1 {
		t.Errorf("expected ModelBreakdown m1=1 m2=1, got %v", sum.ModelBreakdown)
	}
	// Daily rows stay project-free.
	if len(daily) != 1 || daily[0].Tokens != 410 {
		t.Errorf("daily rows = %+v", daily)
	}
	if daily[0].Input+daily[0].Output+daily[0].Cache != daily[0].Tokens {
		t.Errorf("daily types must add up to the total: %+v", daily[0])
	}
}

func TestLoadOpenCodeData_SessionV2(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session_v2 (
			id TEXT PRIMARY KEY, tokens_input INTEGER, tokens_output INTEGER,
			tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT,
			time_created INTEGER, time_updated INTEGER
		);
		CREATE TABLE session (
			id TEXT PRIMARY KEY, tokens_input INTEGER, tokens_output INTEGER,
			tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT,
			time_created INTEGER, time_updated INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	// Two v2 sessions plus a legacy-only session; the legacy table also
	// holds a duplicate of v2 row s1 with inflated numbers that must NOT be
	// counted twice.
	_, err = db.Exec(`
		INSERT INTO session_v2 VALUES ('s1', 100, 50, 10, 5, 5, 0.05, 20, 5, 2, 'code', 1710504000000, 1710504060000);
		INSERT INTO session_v2 VALUES ('s2', 200, 100, 0, 0, 0, 0.10, 0, 0, 0, 'plan', 1710590400000, 1710590400000);
		INSERT INTO session    VALUES ('s1', 999999, 0, 0, 0, 0, 9.99, 0, 0, 0, 'code', 1710504000000, 1710504060000);
		INSERT INTO session    VALUES ('s3', 10, 5, 0, 0, 0, 0.01, 1, 1, 1, 'code', 1710504000000, 1710504060000);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, _, err := LoadOpenCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadOpenCodeData error: %v", err)
	}
	// Sessions: s1 + s2 from v2, s3 from legacy. 3 * 24h apart day buckets
	// for s1/s3 and s2 one day later.
	if sum.Sessions != 3 {
		t.Errorf("expected sessions=3, got %d", sum.Sessions)
	}
	if sum.LifetimeTokens != 485 { // 170 + 300 + 15
		t.Errorf("expected lifetime=485, got %d", sum.LifetimeTokens)
	}
	if sum.Cost < 0.159 || sum.Cost > 0.161 {
		t.Errorf("expected cost=0.16, got %f", sum.Cost)
	}
	if sum.LinesAdded != 21 || sum.LinesDeleted != 6 || sum.FilesTouched != 3 {
		t.Errorf("expected lines 21/6/3, got %d/%d/%d", sum.LinesAdded, sum.LinesDeleted, sum.FilesTouched)
	}
	if sum.AgentBreakdown["code"] != 2 || sum.AgentBreakdown["plan"] != 1 {
		t.Errorf("expected agent breakdown code=2 plan=1, got %v", sum.AgentBreakdown)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 185 || daily[0].Turns != 2 { // s1 + s3
		t.Errorf("expected first day tokens=185 turns=2, got %d/%d", daily[0].Tokens, daily[0].Turns)
	}
	if daily[1].Tokens != 300 || daily[1].Turns != 1 {
		t.Errorf("expected second day tokens=300 turns=1, got %d/%d", daily[1].Tokens, daily[1].Turns)
	}
}

func TestLoadOpenCodeData_SessionV2Only(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	// Newer layouts may drop the legacy session table entirely.
	_, err = db.Exec(`
		CREATE TABLE session_v2 (
			id TEXT PRIMARY KEY, tokens_input INTEGER, tokens_output INTEGER,
			tokens_reasoning INTEGER, tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT,
			time_created INTEGER, time_updated INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO session_v2 VALUES ('s1', 100, 50, 10, 5, 5, 0.05, 0, 0, 0, 'build', 1710504000000, 1710504060000);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, _, err := LoadOpenCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadOpenCodeData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 170 {
		t.Errorf("expected sessions=1, lifetime=170, got %d, %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}

func TestLoadMiMoCodeData_MessageLevel(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mimocode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE session (
			id TEXT PRIMARY KEY,
			time_created INTEGER,
			time_updated INTEGER,
			summary_additions INTEGER,
			summary_deletions INTEGER,
			summary_files INTEGER
		);
		CREATE TABLE message (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			time_created INTEGER,
			data TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO session VALUES ('s1', 1710504000000, 1710504060000, 10, 2, 1);
		INSERT INTO message VALUES (
			'm1', 's1', 1710504000000,
			'{"role":"assistant","tokens":{"input":50,"output":20,"reasoning":0,"cache":{"read":0,"write":0}},"cost":0.01,"agent":"coder"}'
		);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	sum, daily, _, err := LoadMiMoCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadMiMoCodeData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 70 {
		t.Errorf("expected sessions=1, lifetime=70, got %d, %d", sum.Sessions, sum.LifetimeTokens)
	}
	if sum.Cost != 0.01 {
		t.Errorf("expected cost=0.01, got %f", sum.Cost)
	}
	if sum.LinesAdded != 10 || sum.LinesDeleted != 2 {
		t.Errorf("expected lines added=10, deleted=2, got %d, %d", sum.LinesAdded, sum.LinesDeleted)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}

func TestLoadDevinData_Schema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER,
			last_activity_at INTEGER,
			hidden INTEGER
		);
		CREATE TABLE message_nodes (
			row_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at INTEGER,
			chat_message TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sessions VALUES ('sess1', 1710504000, 1710504060, 0);
		INSERT INTO message_nodes (created_at, chat_message) VALUES (
			1710504000,
			'{"role":"assistant","metadata":{"metrics":{"input_tokens":100,"output_tokens":50,"cache_read_tokens":10,"cache_creation_tokens":10}}}'
		);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	// Disable global cache directory setting in unit test to force scan
	t.Setenv("HOME", dir)

	sum, daily, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("LoadDevinData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 170 {
		t.Errorf("expected sessions=1, lifetime=170, got %d, %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}

func TestLoadDevinData_DeltaScan(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin_delta.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER,
			last_activity_at INTEGER,
			hidden INTEGER
		);
		CREATE TABLE message_nodes (
			row_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at INTEGER,
			chat_message TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sessions VALUES ('sess1', 1710504000, 1710504060, 0);
		INSERT INTO message_nodes (created_at, chat_message) VALUES (
			1710504000,
			'{"role":"assistant","metadata":{"metrics":{"input_tokens":100,"output_tokens":50,"cache_read_tokens":10,"cache_creation_tokens":10}}}'
		);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}

	t.Setenv("HOME", dir)

	// Initial scan (populates cache with MaxRowID = 1)
	sum1, daily1, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("initial LoadDevinData error: %v", err)
	}
	if sum1.LifetimeTokens != 170 || daily1[0].Turns != 1 {
		t.Fatalf("expected lifetime=170, turns=1, got %d, %d", sum1.LifetimeTokens, daily1[0].Turns)
	}

	// Insert second message (row_id = 2) with 300 tokens
	_, err = db.Exec(`
		INSERT INTO message_nodes (created_at, chat_message) VALUES (
			1710504000,
			'{"role":"assistant","metadata":{"metrics":{"input_tokens":200,"output_tokens":100,"cache_read_tokens":0,"cache_creation_tokens":0}}}'
		);
	`)
	if err != nil {
		t.Fatalf("exec insert second error: %v", err)
	}

	// Second scan (should trigger fast delta scan via row_id > 1)
	sum2, daily2, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("delta LoadDevinData error: %v", err)
	}
	if sum2.LifetimeTokens != 470 {
		t.Errorf("expected delta lifetime=470, got %d", sum2.LifetimeTokens)
	}
	if len(daily2) != 1 || daily2[0].Turns != 2 || daily2[0].Tokens != 470 {
		t.Errorf("expected daily turns=2, tokens=470, got turns=%d, tokens=%d", daily2[0].Turns, daily2[0].Tokens)
	}

	// Verify exact cache hit after delta scan
	sum3, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("cached LoadDevinData error: %v", err)
	}
	if sum3.LifetimeTokens != 470 {
		t.Errorf("expected cached lifetime=470, got %d", sum3.LifetimeTokens)
	}
}

// TestLoadDevinData_ModelAttribution pins the join that lets the estimator
// price Devin's tokens. Without a model on the day, Catalog.PriceDay returns
// zero and the cost silently vanishes from every report, which is what used to
// happen to 19.2B tokens.
func TestLoadDevinData_ModelAttribution(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin_models.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER,
			last_activity_at INTEGER,
			hidden INTEGER,
			working_directory TEXT,
			model TEXT
		);
		CREATE TABLE message_nodes (
			row_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at INTEGER,
			session_id TEXT,
			chat_message TEXT
		);
	`)
	if err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	msg := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":100,"output_tokens":50,"cache_read_tokens":10,"cache_creation_tokens":20}}}'`
	_, err = db.Exec(`
		INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0, '/work/atlas', 'glm-5-2');
		INSERT INTO sessions VALUES ('s2', 1710504000, 1710504060, 0, '/work/atlas', 'GLM-5-2');
		INSERT INTO sessions VALUES ('s3', 1710504000, 1710504060, 0, '/work/atlas', '');
		INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg + `);
		INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's2', ` + msg + `);
		INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's3', ` + msg + `);
	`)
	if err != nil {
		t.Fatalf("exec insert error: %v", err)
	}
	t.Setenv("HOME", dir)

	_, daily, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("LoadDevinData: %v", err)
	}
	if len(daily) != 1 {
		t.Fatalf("expected one day, got %d", len(daily))
	}
	day := daily[0]

	// Two sessions name the same model in different cases: one model, one key.
	if len(day.Models) != 1 {
		t.Fatalf("expected 1 model after normalisation, got %d: %v", len(day.Models), day.Models)
	}
	counts, ok := day.Models["glm-5-2"]
	if !ok {
		t.Fatalf("model key is not normalised; got %v", day.Models)
	}

	// The invariant that makes the estimate trustworthy: attributed tokens add
	// up to the day total, never more and never less. The third session names no
	// model, so it must not appear here.
	wantInput := int64(200)
	wantOutput := int64(100)
	wantCache := int64(20 + 40)
	if counts.Input != wantInput || counts.Output != wantOutput || counts.CacheRead != 20 || counts.CacheWrite != 40 {
		t.Errorf("attributed counts = %+v, want input %d output %d cache read 20 write 40",
			counts, wantInput, wantOutput)
	}

	// The day totals stay whole, including the unattributed session.
	wantDayTokens := int64(3 * (100 + 50 + 10 + 20))
	if day.Tokens != wantDayTokens {
		t.Errorf("day tokens = %d, want %d", day.Tokens, wantDayTokens)
	}
	attributed := counts.Input + counts.Output + counts.CacheRead + counts.CacheWrite
	if attributed >= day.Tokens {
		t.Errorf("attributed %d covers the whole day %d, but one session had no model", attributed, day.Tokens)
	}
	_ = wantCache
}

// TestLoadDevinData_ProjectModelAttribution proves project rows carry the same
// attribution as daily rows. The projects report prices from project rows, so a
// missing model there sends a tool's spend to the unpriceable line even though
// its daily rows name a model.
func TestLoadDevinData_ProjectModelAttribution(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin_projects.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER,
			last_activity_at INTEGER,
			hidden INTEGER,
			working_directory TEXT,
			model TEXT
		);
		CREATE TABLE message_nodes (
			row_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at INTEGER,
			session_id TEXT,
			chat_message TEXT
		);
	`); err != nil {
		t.Fatalf("exec create error: %v", err)
	}

	msg := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":100,"output_tokens":50,"cache_read_tokens":10,"cache_creation_tokens":20}}}'`
	if _, err := db.Exec(`
		INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0, '/work/atlas', 'glm-5-2');
		INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg + `);
	`); err != nil {
		t.Fatalf("exec insert error: %v", err)
	}
	t.Setenv("HOME", dir)

	_, daily, projects, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("LoadDevinData: %v", err)
	}
	if len(daily) != 1 || len(projects) == 0 {
		t.Fatalf("expected 1 day and at least 1 project, got %d and %d", len(daily), len(projects))
	}

	var found bool
	for _, p := range projects {
		if len(p.Models) == 0 {
			t.Errorf("project %s day %s has no model attribution", p.Project, p.Day)
			continue
		}
		found = true
		counts, ok := p.Models["glm-5-2"]
		if !ok {
			t.Errorf("project models = %v, want glm-5-2", p.Models)
			continue
		}
		if counts.Input != 100 || counts.Output != 50 || counts.CacheRead != 10 || counts.CacheWrite != 20 {
			t.Errorf("project model counts = %+v, want the daily counts", counts)
		}
	}
	if !found {
		t.Fatal("no project row carried a model")
	}
}

func TestLoadDevinData_SourceIsolation(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dirA := t.TempDir()
	dbPathA := filepath.Join(dirA, "devin.db")
	dbA, err := sql.Open("sqlite", dbPathA)
	if err != nil {
		t.Fatal(err)
	}
	defer dbA.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER, working_directory TEXT, model TEXT);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := dbA.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := dbA.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0, '/work/a', 'modelA');`); err != nil {
		t.Fatal(err)
	}
	if _, err := dbA.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);`); err != nil {
		t.Fatal(err)
	}

	dirB := t.TempDir()
	dbPathB := filepath.Join(dirB, "devin.db")
	dbB, err := sql.Open("sqlite", dbPathB)
	if err != nil {
		t.Fatal(err)
	}
	defer dbB.Close()
	if _, err := dbB.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg900 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":600,"output_tokens":300,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := dbB.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0, '/work/b', 'modelB');`); err != nil {
		t.Fatal(err)
	}
	if _, err := dbB.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg900 + `);`); err != nil {
		t.Fatal(err)
	}

	sumA, _, _, err := LoadDevinData(dbPathA)
	if err != nil {
		t.Fatal(err)
	}
	if sumA.LifetimeTokens != 100 {
		t.Fatalf("expected sumA tokens 100, got %d", sumA.LifetimeTokens)
	}

	// Sequential load of database B: probe is identical (1 session, max row_id=1),
	// but because source databases are isolated, sumB must be 900, not cached 100.
	sumB, _, _, err := LoadDevinData(dbPathB)
	if err != nil {
		t.Fatal(err)
	}
	if sumB.LifetimeTokens != 900 {
		t.Fatalf("expected sumB tokens 900, got %d (cross-database cache collision!)", sumB.LifetimeTokens)
	}
}

func TestLoadDevinData_SymlinkConsistency(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg + `);`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	sum1, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum1.LifetimeTokens != 100 {
		t.Fatalf("expected 100, got %d", sum1.LifetimeTokens)
	}

	linkPath := filepath.Join(dir, "devin_symlink.db")
	if err := os.Symlink(dbPath, linkPath); err != nil {
		t.Fatal(err)
	}

	sum2, _, _, err := LoadDevinData(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum2.LifetimeTokens != 100 {
		t.Fatalf("expected 100 from symlink, got %d", sum2.LifetimeTokens)
	}
}

func TestLoadDevinData_FileReplacement(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	sum1, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum1.LifetimeTokens != 100 {
		t.Fatalf("expected 100, got %d", sum1.LifetimeTokens)
	}

	if err := os.Remove(dbPath); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	db2, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db2.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg900 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":600,"output_tokens":300,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db2.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db2.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg900 + `);`); err != nil {
		t.Fatal(err)
	}
	db2.Close()

	sum2, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum2.LifetimeTokens != 900 {
		t.Fatalf("expected replaced database to return 900, got %d", sum2.LifetimeTokens)
	}
}

func TestLoadDevinData_ExistingRowUpdate(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);`); err != nil {
		t.Fatal(err)
	}

	sum1, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum1.LifetimeTokens != 100 {
		t.Fatalf("expected 100, got %d", sum1.LifetimeTokens)
	}

	msg500 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":350,"output_tokens":150,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`UPDATE message_nodes SET chat_message = ` + msg500 + ` WHERE row_id = 1;`); err != nil {
		t.Fatal(err)
	}

	sum2, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum2.LifetimeTokens != 500 {
		t.Fatalf("expected 500 after updating existing row, got %d", sum2.LifetimeTokens)
	}
}

func TestLoadDevinData_HiddenSessionSwap(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	msg200 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":140,"output_tokens":60,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`
		INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);
		INSERT INTO sessions VALUES ('s2', 1710504000, 1710504060, 1);
		INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);
		INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's2', ` + msg200 + `);
	`); err != nil {
		t.Fatal(err)
	}

	sum1, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum1.Sessions != 1 {
		t.Fatalf("expected 1 session, got %d", sum1.Sessions)
	}

	if _, err := db.Exec(`UPDATE sessions SET hidden = CASE WHEN id = 's1' THEN 1 ELSE 0 END;`); err != nil {
		t.Fatal(err)
	}

	sum2, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum2.Sessions != 1 {
		t.Fatalf("expected 1 session, got %d", sum2.Sessions)
	}
}

func TestLoadDevinData_ModelChange(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER, working_directory TEXT, model TEXT);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0, '/work', 'gpt-4o');`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg + `);`); err != nil {
		t.Fatal(err)
	}

	_, daily1, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := daily1[0].Models["gpt-4o"]; !ok {
		t.Fatalf("expected gpt-4o, got %v", daily1[0].Models)
	}

	if _, err := db.Exec(`UPDATE sessions SET model = 'claude-3-5-sonnet' WHERE id = 's1';`); err != nil {
		t.Fatal(err)
	}

	_, daily2, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := daily2[0].Models["claude-3.5-sonnet"]; !ok {
		t.Fatalf("expected claude-3.5-sonnet after update, got %v", daily2[0].Models)
	}
}

func TestLoadDevinData_PromptHistoryActivity(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
		CREATE TABLE prompt_history (id INTEGER PRIMARY KEY, created_at INTEGER, updated_at INTEGER, prompt TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg + `);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO prompt_history VALUES (1, 1710590400, 1710590400, 'test prompt');`); err != nil {
		t.Fatal(err)
	}

	_, daily1, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(daily1) != 2 {
		t.Fatalf("expected 2 days (message day + prompt-only day), got %d: %+v", len(daily1), daily1)
	}

	if _, err := db.Exec(`UPDATE prompt_history SET updated_at = 1710676800 WHERE id = 1;`); err != nil {
		t.Fatal(err)
	}

	_, daily2, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(daily2) != 2 {
		t.Fatalf("expected 2 days after prompt update, got %d: %+v", len(daily2), daily2)
	}
	wantDay := thermal.UnixDay(1710676800)
	if daily2[1].Day != wantDay {
		t.Fatalf("expected day %s, got %s", wantDay, daily2[1].Day)
	}
}

func TestLoadDevinData_AppendOnlyViolationFallback(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);`); err != nil {
		t.Fatal(err)
	}

	sum1, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum1.LifetimeTokens != 100 {
		t.Fatalf("expected 100, got %d", sum1.LifetimeTokens)
	}

	msg500 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":350,"output_tokens":150,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	msg200 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":140,"output_tokens":60,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`UPDATE message_nodes SET chat_message = ` + msg500 + ` WHERE row_id = 1;`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg200 + `);`); err != nil {
		t.Fatal(err)
	}

	sum2, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if sum2.LifetimeTokens != 700 {
		t.Fatalf("expected 700 (500 + 200) after base row modification during append, got %d", sum2.LifetimeTokens)
	}
}

func TestLoadDevinData_CorruptCacheFallback(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);`); err != nil {
		t.Fatal(err)
	}

	_, _, _, err = LoadDevinData(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	canonicalPath := CanonicalDatabasePath(dbPath)
	cacheP, err := devinCachePath(canonicalPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cacheP, []byte("NOT_VALID_JSON{{{"), 0o600); err != nil {
		t.Fatal(err)
	}

	sum2, _, _, err := LoadDevinData(dbPath)
	if err != nil {
		t.Fatalf("expected graceful recovery from corrupt cache, got error: %v", err)
	}
	if sum2.LifetimeTokens != 100 {
		t.Fatalf("expected 100, got %d", sum2.LifetimeTokens)
	}
}

func TestLoadDevinData_ConcurrentWriters(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devin.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createSchema := `
		CREATE TABLE sessions (id TEXT PRIMARY KEY, created_at INTEGER, last_activity_at INTEGER, hidden INTEGER);
		CREATE TABLE message_nodes (row_id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER, session_id TEXT, chat_message TEXT);
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatal(err)
	}
	msg100 := `'{"role":"assistant","metadata":{"metrics":{"input_tokens":70,"output_tokens":30,"cache_read_tokens":0,"cache_creation_tokens":0}}}'`
	if _, err := db.Exec(`INSERT INTO sessions VALUES ('s1', 1710504000, 1710504060, 0);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO message_nodes (created_at, session_id, chat_message) VALUES (1710504000, 's1', ` + msg100 + `);`); err != nil {
		t.Fatal(err)
	}

	const workers = 8
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum, _, _, err := LoadDevinData(dbPath)
			if err != nil {
				errCh <- err
				return
			}
			if sum.LifetimeTokens != 100 {
				errCh <- fmt.Errorf("expected 100, got %d", sum.LifetimeTokens)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
}
