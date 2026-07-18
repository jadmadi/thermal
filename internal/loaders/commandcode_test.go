package loaders

import (
	"os"
	"path/filepath"
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

	sum, daily, err := LoadCommandCodeData(dir)
	if err != nil {
		t.Fatalf("LoadCommandCodeData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 2 {
		t.Errorf("expected 1 session and 2 tokens (messages), got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
	if sum.ModelBreakdown["DeepSeek-Coder"] != 1 {
		t.Errorf("expected DeepSeek-Coder model in breakdown, got %v", sum.ModelBreakdown)
	}
}
