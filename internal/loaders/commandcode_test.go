// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCommandCodeData_MockSession(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "p1")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed creating directories: %v", err)
	}

	sessionFile := filepath.Join(projDir, "sess.jsonl")
	sessionContent := `{"timestamp":"2026-07-18T10:00:00Z","role":"user","source":"cli"}` + "\n" +
		`{"timestamp":"2026-07-18T10:01:00Z","role":"assistant","source":"cli"}` + "\n"
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	metaFile := filepath.Join(projDir, "sess.meta.json")
	metaContent := `{"model":"DeepSeek-Coder"}`
	if err := os.WriteFile(metaFile, []byte(metaContent), 0644); err != nil {
		t.Fatalf("failed writing meta: %v", err)
	}

	sum, daily, _, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 2 {
		t.Errorf("expected 1 session and 2 tokens (messages), got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
	// Model names are lowercased so the same model merges across tools.
	if sum.ModelBreakdown["deepseek-coder"] != 1 {
		t.Errorf("expected deepseek-coder model in breakdown, got %v", sum.ModelBreakdown)
	}
}

func TestLoadCommandCodeData_LineOver256KB(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects", "p1")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("failed creating directories: %v", err)
	}

	sessionFile := filepath.Join(projDir, "sess.jsonl")
	msg1 := `{"timestamp":"2026-08-01T10:00:00Z","role":"user","source":"cli"}` + "\n"
	largePayload := strings.Repeat("A", 300*1024)
	msg2 := fmt.Sprintf(`{"timestamp":"2026-08-01T10:01:00Z","role":"assistant","content":"%s"}`, largePayload) + "\n"
	msg3 := `{"timestamp":"2026-08-02T10:00:00Z","role":"user","source":"cli"}` + "\n"

	if err := os.WriteFile(sessionFile, []byte(msg1+msg2+msg3), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	sum, daily, _, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}

	if sum.LifetimeTokens != 3 {
		t.Fatalf("expected 3 messages, got %d", sum.LifetimeTokens)
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

	sessionFile := filepath.Join(projDir, "sess.jsonl")
	msg1 := `{"timestamp":"2026-08-01T10:00:00Z","role":"user","source":"cli"}` + "\n"
	largePayload := strings.Repeat("A", 600)
	msg2 := fmt.Sprintf(`{"timestamp":"2026-08-01T10:01:00Z","role":"assistant","content":"%s"}`, largePayload) + "\n"
	msg3 := `{"timestamp":"2026-08-02T10:00:00Z","role":"user","source":"cli"}` + "\n"

	if err := os.WriteFile(sessionFile, []byte(msg1+msg2+msg3), 0644); err != nil {
		t.Fatalf("failed writing session: %v", err)
	}

	sum, daily, _, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}

	// Pre-error message is kept; subsequent messages skipped
	if sum.LifetimeTokens != 1 {
		t.Fatalf("expected 1 message kept, got %d", sum.LifetimeTokens)
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
