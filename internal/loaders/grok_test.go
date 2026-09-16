package loaders

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func writeGrokSession(t *testing.T, root, cwd, id, summary string, updates []string) {
	t.Helper()
	dir := filepath.Join(root, "sessions", cwd, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(joinLines(updates)), 0o644); err != nil {
		t.Fatalf("write updates error: %v", err)
	}
	if summary != "" {
		if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(summary), 0o644); err != nil {
			t.Fatalf("write summary error: %v", err)
		}
	}
}

func joinLines(lines []string) string {
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

func TestLoadGrokData_TurnCompleted(t *testing.T) {
	dir := t.TempDir()

	// Day 1: session s1 with one completed turn (440742 in / 2127 out /
	// 327040 cached / 1624 reasoning / 3382780000 ticks), plus one
	// non-turn update and one empty-usage turn that must not count.
	// Day 2: session s2 with one completed turn.
	writeGrokSession(t, dir, "proj", "s1",
		`{"agent_name":"grok-build-plan","created_at":"2026-07-21T21:02:01Z","updated_at":"2026-07-21T21:10:22Z"}`,
		[]string{
			`{"timestamp":1784667722,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"hook_execution"}}}`,
			`{"timestamp":1784667819,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"turn_completed","usage":{"inputTokens":440742,"outputTokens":2127,"totalTokens":442869,"cachedReadTokens":327040,"reasoningTokens":1624,"modelCalls":8,"costUsdTicks":3382780000,"modelUsage":{"grok-4.5-build":{"modelCalls":8}}}}}}`,
			`{"timestamp":1784667820,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"turn_completed","usage":{"inputTokens":0,"outputTokens":0,"totalTokens":0}}}}`,
		})
	writeGrokSession(t, dir, "proj", "s2",
		`{"agent_name":"cursor","created_at":"2026-07-25T10:00:00Z","updated_at":"2026-07-25T10:05:00Z"}`,
		[]string{
			`{"timestamp":1785012000,"method":"_x.ai/session/update","params":{"sessionId":"s2","update":{"sessionUpdate":"turn_completed","usage":{"inputTokens":248171,"outputTokens":1650,"totalTokens":249821,"cachedReadTokens":241024,"reasoningTokens":1212,"modelCalls":4,"costUsdTicks":1065012000,"modelUsage":{"grok-4.5-build":{"modelCalls":4}}}}}}`,
		})

	sum, daily, err := LoadGrokData(dir)
	if err != nil {
		t.Fatalf("LoadGrokData error: %v", err)
	}
	if sum.Sessions != 2 {
		t.Errorf("expected sessions=2, got %d", sum.Sessions)
	}
	if sum.LifetimeTokens != 692690 { // 442869 + 249821, recorded totals
		t.Errorf("expected lifetime=692690, got %d", sum.LifetimeTokens)
	}
	// Input is uncached: 440742-327040 + 248171-241024 = 120849. Output
	// excludes reasoning, which is a nested subset of it:
	// (2127-1624) + (1650-1212) = 941.
	if sum.InputTokens != 120849 || sum.OutputTokens != 941 || sum.ReasoningTokens != 2836 {
		t.Errorf("expected tokens 120849/941/2836, got %d/%d/%d",
			sum.InputTokens, sum.OutputTokens, sum.ReasoningTokens)
	}
	if sum.CacheTokens != 568064 { // 327040 + 241024
		t.Errorf("expected cache=568064, got %d", sum.CacheTokens)
	}
	if sum.InputTokens+sum.OutputTokens+sum.ReasoningTokens+sum.CacheTokens != sum.LifetimeTokens {
		t.Errorf("token types must add up to lifetime %d, got %d", sum.LifetimeTokens,
			sum.InputTokens+sum.OutputTokens+sum.ReasoningTokens+sum.CacheTokens)
	}
	// Ticks to USD at 1e-10: 0.338278 + 0.1065012 = 0.4447792.
	if math.Abs(sum.Cost-0.4447792) > 1e-9 {
		t.Errorf("expected cost=0.4447792, got %f", sum.Cost)
	}
	if sum.ModelBreakdown["grok-4.5-build"] != 12 {
		t.Errorf("expected model calls=12, got %v", sum.ModelBreakdown)
	}
	if sum.AgentBreakdown["grok-build-plan"] != 1 || sum.AgentBreakdown["cursor"] != 1 {
		t.Errorf("expected one session per agent, got %v", sum.AgentBreakdown)
	}
	if sum.LongestSessionMs != 501000 { // s1: 21:02:01 -> 21:10:22
		t.Errorf("expected longest=501000ms, got %d", sum.LongestSessionMs)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(daily))
	}
	if daily[0].Tokens != 442869 || daily[0].Turns != 1 {
		t.Errorf("expected first day 442869/1, got %d/%d", daily[0].Tokens, daily[0].Turns)
	}
	if daily[1].Tokens != 249821 || daily[1].Turns != 1 {
		t.Errorf("expected second day 249821/1, got %d/%d", daily[1].Tokens, daily[1].Turns)
	}
}

func TestLoadGrokData_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	sum, daily, err := LoadGrokData(dir)
	if err != nil {
		t.Fatalf("LoadGrokData error: %v", err)
	}
	if sum.Sessions != 0 || len(daily) != 0 {
		t.Errorf("expected empty result, got sessions=%d rows=%d", sum.Sessions, len(daily))
	}
}
