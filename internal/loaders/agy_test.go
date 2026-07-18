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

	sum, daily, err := LoadAgyData(dir)
	if err != nil {
		t.Fatalf("LoadAgyData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 1 {
		t.Errorf("expected 1 session and 1 token (step), got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
	if sum.ModelBreakdown["Gemini 3.1 Pro (High)"] != 1 {
		t.Errorf("expected model breakdown for Gemini 3.1 Pro (High), got %v", sum.ModelBreakdown)
	}
}
