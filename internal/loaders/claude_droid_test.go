package loaders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClaudeData_AssistantUsage(t *testing.T) {
	dir := t.TempDir()
	proj := filepath.Join(dir, "projects", "proj")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	// Day 1: session s1 with two assistant messages (second carries cache
	// tokens and a model), plus a user message that must not count.
	// Day 2: session s2 with one assistant message.
	s1 := `{"type":"user","timestamp":"2026-07-21T10:00:00Z","message":{"role":"user","content":"hi"}}
{"type":"assistant","timestamp":"2026-07-21T10:00:05Z","message":{"model":"claude-opus","usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
{"type":"assistant","timestamp":"2026-07-21T10:00:10Z","message":{"model":"claude-opus","usage":{"input_tokens":200,"output_tokens":100,"cache_read_input_tokens":1000,"cache_creation_input_tokens":100}}}
`
	s2 := `{"type":"assistant","timestamp":"2026-07-24T10:00:00Z","message":{"model":"claude-sonnet","usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
`
	if err := os.WriteFile(filepath.Join(proj, "s1.jsonl"), []byte(s1), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(proj, "s2.jsonl"), []byte(s2), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	sum, daily, err := LoadClaudeData(dir)
	if err != nil {
		t.Fatalf("LoadClaudeData error: %v", err)
	}
	if sum.Sessions != 2 {
		t.Errorf("expected sessions=2, got %d", sum.Sessions)
	}
	// Lifetime: (100+50) + (200+100+1100) + (10+5) = 1565.
	if sum.LifetimeTokens != 1565 {
		t.Errorf("expected lifetime=1565, got %d", sum.LifetimeTokens)
	}
	if sum.InputTokens != 310 || sum.OutputTokens != 155 || sum.CacheTokens != 1100 {
		t.Errorf("expected tokens 310/155/1100, got %d/%d/%d",
			sum.InputTokens, sum.OutputTokens, sum.CacheTokens)
	}
	if sum.ModelBreakdown["claude-opus"] != 2 || sum.ModelBreakdown["claude-sonnet"] != 1 {
		t.Errorf("expected model breakdown opus=2 sonnet=1, got %v", sum.ModelBreakdown)
	}
	if sum.LongestSessionMs != 5000 { // s1 spans 5 seconds
		t.Errorf("expected longest=5000ms, got %d", sum.LongestSessionMs)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 1550 || daily[0].Turns != 2 {
		t.Errorf("expected first day 1550/2, got %d/%d", daily[0].Tokens, daily[0].Turns)
	}
	if daily[0].Input != 300 || daily[0].Output != 150 || daily[0].Cache != 1100 {
		t.Errorf("expected first day 300/150/1100 typed tokens, got %d/%d/%d",
			daily[0].Input, daily[0].Output, daily[0].Cache)
	}
	if len(daily[0].Models) != 1 || daily[0].Models["claude-opus"].Input != 300 {
		t.Errorf("expected per-day model claude-opus with 300 input, got %v", daily[0].Models)
	}
	if daily[1].Tokens != 15 || daily[1].Turns != 1 {
		t.Errorf("expected second day 15/1, got %d/%d", daily[1].Tokens, daily[1].Turns)
	}
	if len(daily[1].Models) != 1 || daily[1].Models["claude-sonnet"].Output != 5 {
		t.Errorf("expected per-day model claude-sonnet with 5 output, got %v", daily[1].Models)
	}
}

func TestLoadDroidData_Messages(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, "sessions", "proj")
	if err := os.MkdirAll(sess, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	// Day 1: session s1 with two messages (one non-message record skipped).
	// Day 2: session s2 with one message.
	s1 := `{"type":"session_start","id":"s1"}
{"type":"message","id":"m1","timestamp":"2026-08-21T03:29:23Z","message":{"role":"user"}}
{"type":"message","id":"m2","timestamp":"2026-08-21T03:30:01Z","message":{"role":"user"}}
`
	s2 := `{"type":"message","id":"m3","timestamp":"2026-08-24T03:29:23Z","message":{"role":"user"}}
`
	if err := os.WriteFile(filepath.Join(sess, "s1.jsonl"), []byte(s1), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sess, "s2.jsonl"), []byte(s2), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	sum, daily, err := LoadDroidData(dir)
	if err != nil {
		t.Fatalf("LoadDroidData error: %v", err)
	}
	if sum.Sessions != 2 {
		t.Errorf("expected sessions=2, got %d", sum.Sessions)
	}
	if sum.LifetimeTokens != 3 {
		t.Errorf("expected lifetime=3 messages, got %d", sum.LifetimeTokens)
	}
	if sum.LongestSessionMs != 38000 { // s1 spans 38 seconds
		t.Errorf("expected longest=38000ms, got %d", sum.LongestSessionMs)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 2 || daily[0].Turns != 2 {
		t.Errorf("expected first day 2/2, got %d/%d", daily[0].Tokens, daily[0].Turns)
	}
	if daily[1].Tokens != 1 || daily[1].Turns != 1 {
		t.Errorf("expected second day 1/1, got %d/%d", daily[1].Tokens, daily[1].Turns)
	}
}
