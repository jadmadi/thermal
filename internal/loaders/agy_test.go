// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgyData_MockSession(t *testing.T) {
	dir := t.TempDir()
	brainDir := filepath.Join(dir, "brain", "sess-1", ".system_generated", "logs")
	if err := os.MkdirAll(brainDir, 0755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	overview := `{"created_at":"2026-07-18T10:00:00Z","source":"SYSTEM","type":"STEP","content":"test step"}` + "\n"
	if err := os.WriteFile(filepath.Join(brainDir, "overview.txt"), []byte(overview), 0644); err != nil {
		t.Fatalf("write overview error: %v", err)
	}
	transcript := `{"type":"SYSTEM","content":"Model Selection changed from None to Gemini 3.1 Pro (High)."}` + "\n"
	if err := os.WriteFile(filepath.Join(brainDir, "transcript.jsonl"), []byte(transcript), 0644); err != nil {
		t.Fatalf("write transcript error: %v", err)
	}

	sum, daily, _, err := LoadAgyData(dir)
	if err != nil {
		t.Fatalf("LoadAgyData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 1 {
		t.Errorf("expected 1 session and 1 token (step), got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
	// Model names pass through modelName() for canonical representation.
	if sum.ModelBreakdown["gemini-3.1-pro"] != 1 {
		t.Errorf("expected model breakdown for gemini-3.1-pro, got %v", sum.ModelBreakdown)
	}
}

func TestLoadAgyData_TranscriptOnly(t *testing.T) {
	dir := t.TempDir()
	brainDir := filepath.Join(dir, "brain", "sess-1", ".system_generated", "logs")
	if err := os.MkdirAll(brainDir, 0755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	transcript := `{"created_at":"2026-09-10T10:00:00Z","source":"USER_EXPLICIT","type":"USER_INPUT","content":"hi"}` + "\n" +
		`{"created_at":"2026-09-11T10:00:00Z","source":"MODEL","type":"PLANNER_RESPONSE","content":"hello"}` + "\n" +
		`{"type":"SYSTEM","content":"Model Selection changed from None to Gemini 3.7 Flash (High)."}` + "\n"
	if err := os.WriteFile(filepath.Join(brainDir, "transcript.jsonl"), []byte(transcript), 0644); err != nil {
		t.Fatalf("write transcript error: %v", err)
	}

	sum, daily, _, err := LoadAgyData(dir)
	if err != nil {
		t.Fatalf("LoadAgyData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 2 {
		t.Errorf("expected 1 session and 2 steps, got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 2 {
		t.Errorf("expected 2 daily rows, got %d", len(daily))
	}
	if sum.ModelBreakdown["gemini-3.7-flash"] != 2 {
		t.Errorf("expected model breakdown for gemini-3.7-flash = 2, got %v", sum.ModelBreakdown)
	}
}

func TestLoadAgyData_LegacyRootFallback(t *testing.T) {
	dir := t.TempDir()
	cliBrain := filepath.Join(dir, "antigravity-cli", "brain", "sess-1", ".system_generated", "logs")
	if err := os.MkdirAll(cliBrain, 0755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	line := `{"created_at":"2026-09-10T10:00:00Z","source":"SYSTEM","type":"STEP","content":"x"}` + "\n"
	if err := os.WriteFile(filepath.Join(cliBrain, "transcript.jsonl"), []byte(line), 0644); err != nil {
		t.Fatalf("write transcript error: %v", err)
	}

	sum, _, _, err := LoadAgyData(filepath.Join(dir, "antigravity"))
	if err != nil {
		t.Fatalf("LoadAgyData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 1 {
		t.Errorf("expected fallback to find 1 session and 1 step, got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
}

func TestLoadAgyData_ProjectAttribution(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "cool-project")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
		t.Fatalf("mkdir repo error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("write HEAD error: %v", err)
	}

	cliRoot := filepath.Join(dir, "antigravity-cli")
	brainLogs := filepath.Join(cliRoot, "brain", "conv-123", ".system_generated", "logs")
	if err := os.MkdirAll(brainLogs, 0755); err != nil {
		t.Fatalf("mkdir brainLogs error: %v", err)
	}

	// Create conversation_summaries.db
	dbPath := filepath.Join(cliRoot, "conversation_summaries.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite error: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE conversation_summaries (conversation_id text, workspace_uris text);`)
	if err != nil {
		t.Fatalf("create table error: %v", err)
	}
	rawURIs := fmt.Sprintf(`["file://%s"]`, repoDir)
	_, err = db.Exec(`INSERT INTO conversation_summaries VALUES ('conv-123', ?);`, rawURIs)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	db.Close()

	line := `{"created_at":"2026-09-24T12:00:00Z","source":"SYSTEM","type":"STEP","content":"done"}` + "\n"
	if err := os.WriteFile(filepath.Join(brainLogs, "transcript.jsonl"), []byte(line), 0644); err != nil {
		t.Fatalf("write transcript error: %v", err)
	}

	sum, daily, projDays, err := LoadAgyData(cliRoot)
	if err != nil {
		t.Fatalf("LoadAgyData error: %v", err)
	}
	if sum.Sessions != 1 {
		t.Fatalf("expected 1 session, got %d", sum.Sessions)
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	if len(projDays) != 1 {
		t.Fatalf("expected 1 project day row, got %d", len(projDays))
	}
	if projDays[0].Project != repoDir {
		t.Errorf("project = %q, want %q", projDays[0].Project, repoDir)
	}
	if projDays[0].Day != "2026-09-24" {
		t.Errorf("day = %q, want 2026-09-24", projDays[0].Day)
	}
	if projDays[0].Turns != 1 {
		t.Errorf("turns = %d, want 1", projDays[0].Turns)
	}
}

func TestResolveAgyBrainDir_PrimaryHubPrecedence(t *testing.T) {
	dir := t.TempDir()
	cliBrain := filepath.Join(dir, "antigravity-cli", "brain")
	agyBrain := filepath.Join(dir, "antigravity", "brain")
	if err := os.MkdirAll(cliBrain, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(agyBrain, 0755); err != nil {
		t.Fatal(err)
	}

	// Case 1: Caller passes legacy antigravity-cli, but primary hub antigravity exists -> prefers primary hub
	got := ResolveAgyBrainDir(filepath.Join(dir, "antigravity-cli"))
	if got != agyBrain {
		t.Errorf("expected primary hub %s, got %s", agyBrain, got)
	}

	// Case 2: Caller passes primary hub antigravity -> returns primary hub
	got2 := ResolveAgyBrainDir(filepath.Join(dir, "antigravity"))
	if got2 != agyBrain {
		t.Errorf("expected primary hub %s, got %s", agyBrain, got2)
	}

	// Case 3: Primary hub does not exist -> falls back to legacy archive
	_ = os.RemoveAll(agyBrain)
	got3 := ResolveAgyBrainDir(filepath.Join(dir, "antigravity"))
	if got3 != cliBrain {
		t.Errorf("expected fallback to legacy %s, got %s", cliBrain, got3)
	}
}
