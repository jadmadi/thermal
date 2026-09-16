package loaders

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

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
