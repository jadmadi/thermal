// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func commandCodeRepo(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatalf("failed creating repo marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed writing repo marker: %v", err)
	}
	return repo
}

func TestLoadCommandCodeData_ModernTranscript(t *testing.T) {
	dir := t.TempDir()
	repo := commandCodeRepo(t)
	projDir := filepath.Join(dir, "projects", "p1")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed creating directories: %v", err)
	}

	sessionFile := filepath.Join(projDir, "sess.jsonl")
	sessionContent := fmt.Sprintf(`{"type":"session","id":"s1","timestamp":"2026-07-18T10:00:00Z","cwd":%q}`+"\n", repo) +
		`{"type":"message","id":"m1","timestamp":"2026-07-18T10:00:01Z","message":{"role":"user"}}` + "\n" +
		`{"type":"message","id":"m2","timestamp":"2026-07-18T10:01:00Z","model":"Qwen/Qwen3.7-Flash","message":{"role":"assistant"},"usage":{"inputTokens":1000,"outputTokens":100,"cacheReadTokens":600,"cacheWriteTokens":100,"costUsd":0.01}}` + "\n" +
		`{"type":"message","id":"m3","timestamp":"2026-07-19T10:00:00Z","message":{"role":"user"}}` + "\n" +
		`{"type":"message","id":"m4","timestamp":"2026-07-19T10:02:00Z","model":"deepseek/deepseek-v4.1-flash","message":{"role":"assistant"},"usage":{"inputTokens":2000,"outputTokens":50,"cacheReadTokens":0,"cacheWriteTokens":0,"costUsd":0.02}}` + "\n"
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	sum, daily, projects, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}
	if sum.Sessions != 1 {
		t.Errorf("sessions = %d, want 1", sum.Sessions)
	}
	if sum.InputTokens != 2300 || sum.OutputTokens != 150 || sum.CacheTokens != 700 {
		t.Errorf("token split = input %d, output %d, cache %d; want 2300, 150, 700", sum.InputTokens, sum.OutputTokens, sum.CacheTokens)
	}
	if sum.LifetimeTokens != 3150 {
		t.Errorf("lifetime tokens = %d, want 3150", sum.LifetimeTokens)
	}
	if math.Abs(sum.Cost-0.03) > 1e-9 {
		t.Errorf("cost = %v, want 0.03", sum.Cost)
	}
	if len(daily) != 2 {
		t.Fatalf("daily rows = %d, want 2", len(daily))
	}

	first := daily[0]
	if first.Day != "2026-07-18" || first.Tokens != 1100 || first.Input != 300 || first.Output != 100 || first.Cache != 700 || first.Turns != 1 {
		t.Errorf("first day = %+v, want 2026-07-18 / 1100 tokens / 300 in / 100 out / 700 cache / 1 turn", first)
	}
	if math.Abs(first.Cost-0.01) > 1e-9 {
		t.Errorf("first day cost = %v, want 0.01", first.Cost)
	}
	qwen := modelName("Qwen/Qwen3.7-Flash")
	if got := first.Models[qwen]; got.Input != 300 || got.Output != 100 || got.CacheRead != 600 || got.CacheWrite != 100 {
		t.Errorf("first day model tokens = %+v, want 300 in / 100 out / 600 read / 100 write", got)
	}

	second := daily[1]
	if second.Day != "2026-07-19" || second.Tokens != 2050 || second.Input != 2000 || second.Output != 50 || second.Cache != 0 || second.Turns != 1 {
		t.Errorf("second day = %+v, want 2026-07-19 / 2050 tokens / 2000 in / 50 out / 0 cache / 1 turn", second)
	}

	if sum.ModelBreakdown[qwen] != 1 || sum.ModelBreakdown[modelName("deepseek/deepseek-v4.1-flash")] != 1 {
		t.Errorf("model breakdown = %v, want one turn per model", sum.ModelBreakdown)
	}

	wantProject := thermal.ProjectKey(repo)
	if len(projects) != 2 {
		t.Fatalf("project rows = %d, want 2", len(projects))
	}
	for _, p := range projects {
		if p.Project != wantProject {
			t.Errorf("project = %q, want %q", p.Project, wantProject)
		}
	}
	if projects[0].Day != "2026-07-18" || projects[0].Tokens != 1100 || projects[0].Turns != 1 {
		t.Errorf("first project day = %+v, want 2026-07-18 / 1100 tokens / 1 turn", projects[0])
	}
	if projects[1].Day != "2026-07-19" || projects[1].Tokens != 2050 || projects[1].Turns != 1 {
		t.Errorf("second project day = %+v, want 2026-07-19 / 2050 tokens / 1 turn", projects[1])
	}
}

func TestLoadCommandCodeData_LegacyTranscript(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "p1")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed creating directories: %v", err)
	}

	sessionFile := filepath.Join(projDir, "sess.jsonl")
	sessionContent := `{"id":"a","timestamp":"2026-07-18T10:00:00Z","role":"user"}` + "\n" +
		`{"id":"b","timestamp":"2026-07-18T10:01:00Z","role":"assistant"}` + "\n" +
		`{"id":"c","timestamp":"2026-07-19T10:01:00Z","role":"assistant"}` + "\n"
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	metaFile := filepath.Join(projDir, "sess.meta.json")
	if err := os.WriteFile(metaFile, []byte(`{"model":"DeepSeek-Coder"}`), 0644); err != nil {
		t.Fatalf("failed writing meta: %v", err)
	}

	sum, daily, projects, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}
	if sum.Sessions != 1 {
		t.Errorf("sessions = %d, want 1", sum.Sessions)
	}
	if sum.LifetimeTokens != 0 || sum.InputTokens != 0 || sum.OutputTokens != 0 || sum.Cost != 0 {
		t.Errorf("legacy transcript should report turns only, got %+v", sum)
	}
	if len(daily) != 2 {
		t.Fatalf("daily rows = %d, want 2", len(daily))
	}
	for _, row := range daily {
		if row.Tokens != 0 || row.Turns != 1 {
			t.Errorf("legacy day = %+v, want 0 tokens / 1 assistant turn", row)
		}
	}
	if len(projects) != 0 {
		t.Errorf("legacy transcript has no cwd, got %d project rows", len(projects))
	}
	if sum.ModelBreakdown[modelName("DeepSeek-Coder")] != 2 {
		t.Errorf("model breakdown = %v, want two turns of deepseek-coder", sum.ModelBreakdown)
	}
}

func TestLoadCommandCodeData_LineOver256KB(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "p1")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed creating directories: %v", err)
	}

	largePayload := strings.Repeat("A", 300*1024)
	sessionFile := filepath.Join(projDir, "sess.jsonl")
	sessionContent := `{"type":"message","id":"m1","timestamp":"2026-08-01T10:00:00Z","model":"m","message":{"role":"assistant"},"usage":{"inputTokens":10,"outputTokens":5}}` + "\n" +
		fmt.Sprintf(`{"type":"message","id":"m2","timestamp":"2026-08-02T10:01:00Z","model":"m","message":{"role":"assistant","content":"%s"},"usage":{"inputTokens":20,"outputTokens":5}}`, largePayload) + "\n"
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	sum, daily, _, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}
	if sum.LifetimeTokens != 40 {
		t.Fatalf("expected 40 tokens, got %d", sum.LifetimeTokens)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 days, got %d", len(daily))
	}
	if len(sum.Warnings) != 0 {
		t.Fatalf("expected 0 warnings on clean data, got %v", sum.Warnings)
	}
}

func TestLoadCommandCodeData_CeilingWarning(t *testing.T) {
	origCeiling := jsonlMaxLineBytes
	jsonlMaxLineBytes = 512
	defer func() { jsonlMaxLineBytes = origCeiling }()

	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "p1")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed creating directories: %v", err)
	}

	largePayload := strings.Repeat("A", 600)
	sessionFile := filepath.Join(projDir, "sess.jsonl")
	sessionContent := `{"type":"message","id":"m1","timestamp":"2026-08-01T10:00:00Z","model":"m","message":{"role":"assistant"},"usage":{"inputTokens":10,"outputTokens":5}}` + "\n" +
		fmt.Sprintf(`{"type":"message","id":"m2","timestamp":"2026-08-02T10:01:00Z","message":{"role":"user","content":"%s"}}`, largePayload) + "\n" +
		`{"type":"message","id":"m3","timestamp":"2026-08-03T10:01:00Z","model":"m","message":{"role":"assistant"},"usage":{"inputTokens":20,"outputTokens":5}}` + "\n"
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	sum, daily, _, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}

	if sum.LifetimeTokens != 15 {
		t.Fatalf("expected 15 tokens kept, got %d", sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 day, got %d", len(daily))
	}
	if len(sum.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(sum.Warnings), sum.Warnings)
	}
	w := sum.Warnings[0]
	if !strings.Contains(w, sessionFile) || !strings.Contains(w, "512 byte ceiling") || !strings.Contains(w, "rest of file skipped") {
		t.Errorf("warning text didn't match expected pattern, got: %s", w)
	}
}
