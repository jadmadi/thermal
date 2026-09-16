package loaders

import (
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
	// Model names are lowercased so the same model merges across tools.
	if sum.ModelBreakdown["gemini 3.1 pro (high)"] != 1 {
		t.Errorf("expected model breakdown for Gemini 3.1 Pro (High), got %v", sum.ModelBreakdown)
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
