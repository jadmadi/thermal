// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

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

	sum, daily, _, err := LoadGrokData(dir)
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
	sum, daily, _, err := LoadGrokData(dir)
	if err != nil {
		t.Fatalf("LoadGrokData error: %v", err)
	}
	if sum.Sessions != 0 || len(daily) != 0 {
		t.Errorf("expected empty result, got sessions=%d rows=%d", sum.Sessions, len(daily))
	}
}

func TestLoadGrokData_Deduplication(t *testing.T) {
	dir := t.TempDir()

	writeGrokSession(t, dir, "proj", "s1",
		`{"agent_name":"grok-build","created_at":"2026-07-21T21:00:00Z","updated_at":"2026-07-21T21:10:00Z"}`,
		[]string{
			// First turn p1
			`{"timestamp":1784667819,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"turn_completed","prompt_id":"p1","usage":{"inputTokens":1000,"outputTokens":100,"totalTokens":1100,"costUsdTicks":1000000000,"modelUsage":{"grok-4.5-build":{"modelCalls":1}}}}}}`,
			// Duplicate turn p1 (re-streamed or duplicate event)
			`{"timestamp":1784667820,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"turn_completed","prompt_id":"p1","usage":{"inputTokens":1000,"outputTokens":100,"totalTokens":1100,"costUsdTicks":1000000000,"modelUsage":{"grok-4.5-build":{"modelCalls":1}}}}}}`,
			// Second distinct turn p2
			`{"timestamp":1784667825,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"turn_completed","prompt_id":"p2","usage":{"inputTokens":500,"outputTokens":50,"totalTokens":550,"costUsdTicks":500000000,"modelUsage":{"grok-4.5-build":{"modelCalls":1}}}}}}`,
			// Duplicate turn p2 via eventId fallback when prompt_id matches
			`{"timestamp":1784667826,"method":"_x.ai/session/update","params":{"sessionId":"s1","update":{"sessionUpdate":"turn_completed","prompt_id":"p2","usage":{"inputTokens":500,"outputTokens":50,"totalTokens":550,"costUsdTicks":500000000,"modelUsage":{"grok-4.5-build":{"modelCalls":1}}}}}}`,
		})

	sum, daily, _, err := LoadGrokData(dir)
	if err != nil {
		t.Fatalf("LoadGrokData error: %v", err)
	}

	// Should only count 1100 + 550 = 1650 lifetime tokens (not 1100*2 + 550*2 = 3300)
	if sum.LifetimeTokens != 1650 {
		t.Errorf("expected lifetime tokens=1650, got %d", sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	if daily[0].Turns != 2 {
		t.Errorf("expected 2 turns, got %d", daily[0].Turns)
	}
	if len(sum.Warnings) != 1 || sum.Warnings[0] != "grok: removed 2 duplicate turn(s)" {
		t.Errorf("expected duplicate warning, got %v", sum.Warnings)
	}

	// Test fallback identifiers: turn_id and _meta.eventId
	dir2 := t.TempDir()
	writeGrokSession(t, dir2, "proj", "s2",
		`{"agent_name":"grok-build"}`,
		[]string{
			`{"timestamp":1784667819,"method":"_x.ai/session/update","params":{"sessionId":"s2","update":{"sessionUpdate":"turn_completed","turn_id":"t1","usage":{"inputTokens":100,"outputTokens":10,"totalTokens":110}}}}`,
			`{"timestamp":1784667820,"method":"_x.ai/session/update","params":{"sessionId":"s2","update":{"sessionUpdate":"turn_completed","turn_id":"t1","usage":{"inputTokens":100,"outputTokens":10,"totalTokens":110}}}}`,
			`{"timestamp":1784667821,"method":"_x.ai/session/update","_meta":{"eventId":"ev-1"},"params":{"sessionId":"s2","update":{"sessionUpdate":"turn_completed","usage":{"inputTokens":200,"outputTokens":20,"totalTokens":220}}}}`,
			`{"timestamp":1784667822,"method":"_x.ai/session/update","_meta":{"eventId":"ev-1"},"params":{"sessionId":"s2","update":{"sessionUpdate":"turn_completed","usage":{"inputTokens":200,"outputTokens":20,"totalTokens":220}}}}`,
		})

	sum2, daily2, _, err2 := LoadGrokData(dir2)
	if err2 != nil {
		t.Fatalf("LoadGrokData error: %v", err2)
	}
	if sum2.LifetimeTokens != 330 {
		t.Errorf("expected lifetime tokens=330, got %d", sum2.LifetimeTokens)
	}
	if len(daily2) != 1 || daily2[0].Turns != 2 {
		t.Errorf("expected 2 turns, got %v", daily2)
	}
	if len(sum2.Warnings) != 1 || sum2.Warnings[0] != "grok: removed 2 duplicate turn(s)" {
		t.Errorf("expected 2 removed turns warning, got %v", sum2.Warnings)
	}
}
