package loaders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCodewhaleData_SessionMetadata(t *testing.T) {
	dir := t.TempDir()
	sessions := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	s1 := `{"metadata":{"created_at":"2026-08-10T09:00:00Z","updated_at":"2026-08-10T09:30:00Z","message_count":12,"total_tokens":5000,"model":"claude-opus-4-6","mode":"build","cumulative_turn_secs":1800,"cost":{"session_cost_usd":1.25}}}`
	s2 := `{"metadata":{"created_at":"2026-08-11T09:00:00Z","updated_at":"2026-08-11T09:30:00Z","message_count":3,"total_tokens":1000,"model":"gpt-5.4","mode":"plan","cumulative_turn_secs":600,"cost":{"session_cost_usd":0.5}}}`
	if err := os.WriteFile(filepath.Join(sessions, "a.json"), []byte(s1), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessions, "b.json"), []byte(s2), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	sum, daily, _, err := LoadCodewhaleData(dir)
	if err != nil {
		t.Fatalf("LoadCodewhaleData error: %v", err)
	}
	if sum.Sessions != 2 || sum.LifetimeTokens != 6000 {
		t.Errorf("expected sessions=2 lifetime=6000, got %d/%d", sum.Sessions, sum.LifetimeTokens)
	}
	if sum.Cost != 1.75 {
		t.Errorf("expected cost=1.75, got %f", sum.Cost)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 5000 || daily[0].Cost != 1.25 || daily[0].Turns != 12 {
		t.Errorf("expected first day 5000 tokens, $1.25, 12 turns, got %d/$%f/%d",
			daily[0].Tokens, daily[0].Cost, daily[0].Turns)
	}
	// codewhale has no token type split, so the day's tokens sit in the
	// unclassified bucket of the per-model map.
	if got := daily[0].Models["claude-opus-4-6"].Unclassified; got != 5000 {
		t.Errorf("expected 5000 unclassified tokens for claude-opus-4-6, got %d", got)
	}
	if _, ok := daily[1].Models["gpt-5.4"]; !ok {
		t.Errorf("expected gpt-5.4 in second day models, got %v", daily[1].Models)
	}
}
