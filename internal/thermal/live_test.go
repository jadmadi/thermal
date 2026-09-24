// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestLiveTracker_InitialBaseline(t *testing.T) {
	lt := NewLiveTracker()
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	res := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 10_000_000,
				Sessions:       500,
				Cost:           12.50,
				ModelBreakdown: map[string]int64{
					"gemini-2.5-pro": 10_000_000,
				},
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 1_000_000, Turns: 50, Cost: 1.25, Input: 100_000, Cache: 900_000},
			},
		},
	}

	snap, events := lt.Poll(res, nil, now)

	// First poll should initialize baselines without creating delta events
	if len(events) != 0 {
		t.Fatalf("expected 0 events on first baseline poll, got %d", len(events))
	}
	if snap.TodayTokens != 1_000_000 {
		t.Errorf("expected todayTokens 1M, got %d", snap.TodayTokens)
	}
	if snap.SessionTokens != 0 {
		t.Errorf("expected sessionTokens 0, got %d", snap.SessionTokens)
	}
}

func TestLiveTracker_DeltaDetection(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	res1 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 10_000_000,
				Sessions:       500,
				Cost:           12.50,
				ModelBreakdown: map[string]int64{
					"gemini-2.5-pro": 10_000_000,
				},
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 1_000_000, Turns: 50, Cost: 1.25},
			},
		},
	}

	// 1. Initial poll
	lt.Poll(res1, nil, t1)

	// 2. Second poll: OpenCode burns 25,000 tokens
	t2 := t1.Add(2 * time.Second)
	res2 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 10_025_000,
				Sessions:       501,
				Cost:           12.55,
				ModelBreakdown: map[string]int64{
					"gemini-2.5-pro": 10_025_000,
				},
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 1_025_000, Turns: 51, Cost: 1.30},
			},
		},
	}

	snap2, events2 := lt.Poll(res2, nil, t2)

	if len(events2) != 1 {
		t.Fatalf("expected 1 delta event, got %d", len(events2))
	}
	ev := events2[0]
	if ev.Tool != "OpenCode" {
		t.Errorf("expected event tool OpenCode, got %s", ev.Tool)
	}
	if ev.Tokens != 25_000 {
		t.Errorf("expected delta tokens 25,000, got %d", ev.Tokens)
	}
	if ev.Model != "gemini-2.5-pro" {
		t.Errorf("expected model gemini-2.5-pro, got %s", ev.Model)
	}
	if snap2.SessionTokens != 25_000 {
		t.Errorf("expected sessionTokens 25,000, got %d", snap2.SessionTokens)
	}
	if snap2.SessionTurns != 1 {
		t.Errorf("expected sessionTurns 1, got %d", snap2.SessionTurns)
	}
	if snap2.FlameIntensity <= 0 {
		t.Errorf("expected flameIntensity > 0, got %f", snap2.FlameIntensity)
	}
	if snap2.BurnTokensPerMin <= 0 {
		t.Errorf("expected positive burnTokensPerMin, got %f", snap2.BurnTokensPerMin)
	}
}

func TestLiveTracker_FlameDecayAndReset(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	res := []ToolResult{
		{
			Tool:    ToolOpenCode,
			Name:    "OpenCode",
			Summary: Summary{LifetimeTokens: 1_000_000, Sessions: 100, Cost: 5.0},
		},
	}

	lt.Poll(res, nil, t1)

	// Burst
	res[0].Summary.LifetimeTokens += 50_000
	t2 := t1.Add(1 * time.Second)
	snap2, _ := lt.Poll(res, nil, t2)
	initialFlame := snap2.FlameIntensity

	if initialFlame == 0 {
		t.Fatalf("expected non-zero flame intensity on burst")
	}

	// 5 seconds idle
	t3 := t2.Add(5 * time.Second)
	snap3, _ := lt.Poll(res, nil, t3)
	if snap3.FlameIntensity >= initialFlame {
		t.Errorf("expected flame to decay, got %f >= %f", snap3.FlameIntensity, initialFlame)
	}

	// Reset Session
	lt.ResetSession()
	snap4, _ := lt.Poll(res, nil, t3.Add(1*time.Second))
	if snap4.SessionTokens != 0 {
		t.Errorf("expected sessionTokens 0 after reset, got %d", snap4.SessionTokens)
	}
}

func TestFlameBar(t *testing.T) {
	bar0, pct0 := FlameBar(0, 20)
	if pct0 != 0 {
		t.Errorf("expected 0%%, got %d", pct0)
	}
	if strings.Contains(bar0, "█") {
		t.Errorf("expected 0 filled cells in bar0: %s", bar0)
	}

	bar100, pct100 := FlameBar(1.0, 20)
	if pct100 != 100 {
		t.Errorf("expected 100%%, got %d", pct100)
	}
	if strings.Contains(bar100, "░") {
		t.Errorf("expected no empty cells in bar100: %s", bar100)
	}

	bar50, pct50 := FlameBar(0.5, 20)
	if pct50 != 50 {
		t.Errorf("expected 50%%, got %d", pct50)
	}
	if len([]rune(bar50)) != 20 {
		t.Errorf("expected width 20, got %d", len([]rune(bar50)))
	}
}

func TestSparkline60(t *testing.T) {
	var buckets [60]int64
	buckets[58] = 50_000
	buckets[59] = 100_000

	spark := Sparkline60(buckets, 59)
	runes := []rune(spark)
	if len(runes) != 60 {
		t.Fatalf("expected 60 runes, got %d", len(runes))
	}

	// The newest element (at index 59) is 100K (the max), should be '█'
	if runes[59] != '█' {
		t.Errorf("expected last rune to be '█', got %c", runes[59])
	}
}

func TestLiveTracker_ModelAttributionToday(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	// Historical data has 100 turns of gemini-3.7-flash, but today has 10 turns of gemini-3.8-flash
	res1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 110,
				Sessions:       110,
				ModelBreakdown: map[string]int64{
					"gemini-3.7-flash": 100,
					"gemini-3.8-flash": 10,
				},
			},
			Daily: []DailyRow{
				{
					Day:    "2026-09-24",
					Tokens: 10,
					Turns:  10,
					Models: map[string]ModelTokens{
						"gemini-3.8-flash": {Unclassified: 10},
					},
				},
			},
		},
	}

	snap1, _ := lt.Poll(res1, nil, t1)
	if snap1.ActiveModel != "gemini-3.8-flash" {
		t.Errorf("expected initial activeModel = gemini-3.8-flash, got %s", snap1.ActiveModel)
	}

	// Turn increment without ModelBreakdown change
	t2 := t1.Add(time.Second)
	res2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 111,
				Sessions:       111,
				ModelBreakdown: map[string]int64{
					"gemini-3.7-flash": 100,
					"gemini-3.8-flash": 10, // not changed in breakdown
				},
			},
			Daily: []DailyRow{
				{
					Day:    "2026-09-24",
					Tokens: 11,
					Turns:  11,
					Models: map[string]ModelTokens{
						"gemini-3.8-flash": {Unclassified: 11},
					},
				},
			},
		},
	}

	snap2, events := lt.Poll(res2, nil, t2)
	if len(events) != 1 {
		t.Fatalf("expected 1 delta event, got %d", len(events))
	}
	if events[0].Model != "gemini-3.8-flash" {
		t.Errorf("expected delta event model = gemini-3.8-flash, got %s", events[0].Model)
	}
	if snap2.ActiveModel != "gemini-3.8-flash" {
		t.Errorf("expected activeModel = gemini-3.8-flash, got %s", snap2.ActiveModel)
	}
}

func TestLiveTracker_MultiProjectDelta(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	// Baseline poll: Project A has 80,000 tokens, Project B has 20,000 tokens
	res1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 100_000,
				Sessions:       100,
			},
		},
	}
	proj1 := []ProjectDay{
		{Tool: "Agy", Project: "/workspace/projectA", Day: "2026-09-24", Tokens: 80_000, Turns: 80},
		{Tool: "Agy", Project: "/workspace/projectB", Day: "2026-09-24", Tokens: 20_000, Turns: 20},
	}

	snap1, events1 := lt.Poll(res1, proj1, t1)
	if len(events1) != 0 {
		t.Fatalf("expected 0 events on initial poll, got %d", len(events1))
	}
	if snap1.ActiveProject != "projectA" {
		t.Errorf("expected initial activeProject = projectA (largest), got %s", snap1.ActiveProject)
	}

	// Poll 2: Delta happens exclusively in Project B (+10 tokens)
	t2 := t1.Add(time.Second)
	res2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 100_010,
				Sessions:       101,
			},
		},
	}
	proj2 := []ProjectDay{
		{Tool: "Agy", Project: "/workspace/projectA", Day: "2026-09-24", Tokens: 80_000, Turns: 80}, // unchanged!
		{Tool: "Agy", Project: "/workspace/projectB", Day: "2026-09-24", Tokens: 20_010, Turns: 21}, // +10 tokens, +1 turn
	}

	snap2, events2 := lt.Poll(res2, proj2, t2)
	if len(events2) != 1 {
		t.Fatalf("expected 1 event on projectB delta, got %d", len(events2))
	}
	if events2[0].Project != "projectB" {
		t.Errorf("expected event project to be projectB, got %s (must not pick projectA just because it has higher total tokens!)", events2[0].Project)
	}
	if events2[0].Tokens != 10 {
		t.Errorf("expected 10 tokens in event, got %d", events2[0].Tokens)
	}
	if snap2.ActiveProject != "projectB" {
		t.Errorf("expected activeProject to switch to projectB, got %s", snap2.ActiveProject)
	}

	// Poll 3: Delta happens in Project A (+5 tokens)
	t3 := t2.Add(time.Second)
	res3 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 100_015,
				Sessions:       102,
			},
		},
	}
	proj3 := []ProjectDay{
		{Tool: "Agy", Project: "/workspace/projectA", Day: "2026-09-24", Tokens: 80_005, Turns: 81}, // +5 tokens!
		{Tool: "Agy", Project: "/workspace/projectB", Day: "2026-09-24", Tokens: 20_010, Turns: 21}, // unchanged!
	}

	snap3, events3 := lt.Poll(res3, proj3, t3)
	if len(events3) != 1 {
		t.Fatalf("expected 1 event on projectA delta, got %d", len(events3))
	}
	if events3[0].Project != "projectA" {
		t.Errorf("expected event project to be projectA, got %s", events3[0].Project)
	}
	if snap3.ActiveProject != "projectA" {
		t.Errorf("expected activeProject to switch back to projectA, got %s", snap3.ActiveProject)
	}
}

func TestLiveTracker_ActivityOnlyNoTokenBurn(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	// Agy has 7 activity steps recorded today
	res1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 7,
				Sessions:       1,
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 7, Turns: 7}, // Activity only (Tokens == Turns, no token types)
			},
		},
	}

	snap1, events1 := lt.Poll(res1, nil, t1)
	if len(events1) != 0 {
		t.Fatalf("expected 0 events on baseline poll, got %d", len(events1))
	}
	if snap1.TodayTokens != 0 {
		t.Errorf("expected TodayTokens 0 for activity-only tool, got %d", snap1.TodayTokens)
	}
	if snap1.TodayTurns != 7 {
		t.Errorf("expected TodayTurns 7, got %d", snap1.TodayTurns)
	}
	if snap1.SessionTokens != 0 {
		t.Errorf("expected SessionTokens 0, got %d", snap1.SessionTokens)
	}

	// Canonical report agreement
	dailyReport := Aggregate(res1[0].Daily, GrainDay, AggregateOptions{Now: t1}, nil)
	var canonicalTodayTokens int64
	for _, row := range dailyReport.Rows {
		if row.Period == "2026-09-24" {
			canonicalTodayTokens = row.Tokens
		}
	}
	if snap1.TodayTokens != canonicalTodayTokens {
		t.Errorf("expected live todayTokens %d to agree with canonical daily report %d", snap1.TodayTokens, canonicalTodayTokens)
	}

	// Delta: 3 more activity steps
	t2 := t1.Add(time.Second)
	res2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 10,
				Sessions:       1,
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 10, Turns: 10},
			},
		},
	}

	snap2, events2 := lt.Poll(res2, nil, t2)
	if snap2.TodayTokens != 0 {
		t.Errorf("expected TodayTokens to remain 0, got %d", snap2.TodayTokens)
	}
	if snap2.TodayTurns != 10 {
		t.Errorf("expected TodayTurns 10, got %d", snap2.TodayTurns)
	}
	if snap2.SessionTokens != 0 {
		t.Errorf("expected SessionTokens to remain 0 (no token burn), got %d", snap2.SessionTokens)
	}
	if snap2.SessionTurns != 3 {
		t.Errorf("expected SessionTurns 3, got %d", snap2.SessionTurns)
	}
	if snap2.BurnTokensPerMin != 0 {
		t.Errorf("expected BurnTokensPerMin 0 for activity-only tool, got %f", snap2.BurnTokensPerMin)
	}
	if snap2.FlameIntensity != 0 {
		t.Errorf("expected FlameIntensity 0, got %f", snap2.FlameIntensity)
	}
	// An activity event may be generated for turn tracking, but with 0 tokens
	if len(events2) > 0 && events2[0].Tokens != 0 {
		t.Errorf("expected event tokens to be 0 for activity tool, got %d", events2[0].Tokens)
	}
}

func TestLiveTracker_CacheHitRate(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	// Prompt 1: Cache-write only (no cache read, no input)
	res1 := []ToolResult{
		{
			Tool:    ToolClaude,
			Name:    "Claude",
			Summary: Summary{LifetimeTokens: 1000},
			Daily: []DailyRow{
				{
					Day:    "2026-09-24",
					Tokens: 1000,
					Turns:  1,
					Models: map[string]ModelTokens{
						"claude-3-5-sonnet": {CacheWrite: 1000},
					},
				},
			},
		},
	}

	snap1, _ := lt.Poll(res1, nil, t1)
	if snap1.TodayCacheHit != 0.0 {
		t.Errorf("expected cache-write-only prompt to yield 0%% hit rate, got %f%%", snap1.TodayCacheHit)
	}

	// Prompt 2: With cache read
	// Input: 1000, CacheRead: 500, CacheWrite: 500 -> Total prompt = 2000. Hit rate = 500 / 2000 = 25%
	res2 := []ToolResult{
		{
			Tool:    ToolClaude,
			Name:    "Claude",
			Summary: Summary{LifetimeTokens: 3000},
			Daily: []DailyRow{
				{
					Day:    "2026-09-24",
					Tokens: 3000,
					Turns:  2,
					Models: map[string]ModelTokens{
						"claude-3-5-sonnet": {Input: 1000, CacheRead: 500, CacheWrite: 500, Output: 1000},
					},
				},
			},
		},
	}

	snap2, _ := lt.Poll(res2, nil, t1.Add(time.Second))
	if math.Abs(snap2.TodayCacheHit-25.0) > 0.01 {
		t.Errorf("expected 25%% cache hit rate, got %f%%", snap2.TodayCacheHit)
	}
}

func TestLiveTracker_TurnDeltaExistingSession(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	// Session count is 1, with 5 turns completed so far
	res1 := []ToolResult{
		{
			Tool: ToolDevin,
			Name: "Devin",
			Summary: Summary{
				LifetimeTokens: 50_000,
				Sessions:       1,
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 50_000, Turns: 5},
			},
		},
	}

	lt.Poll(res1, nil, t1)

	// New turn executed inside the same session: Sessions stays 1, Turns increments to 6
	res2 := []ToolResult{
		{
			Tool: ToolDevin,
			Name: "Devin",
			Summary: Summary{
				LifetimeTokens: 60_000,
				Sessions:       1, // session count unchanged!
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 60_000, Turns: 6},
			},
		},
	}

	snap2, events2 := lt.Poll(res2, nil, t1.Add(time.Second))
	if len(events2) != 1 {
		t.Fatalf("expected 1 event on additional turn, got %d", len(events2))
	}
	if snap2.SessionTurns != 1 {
		t.Errorf("expected SessionTurns 1, got %d", snap2.SessionTurns)
	}
	if snap2.TodayTurns != 6 {
		t.Errorf("expected TodayTurns 6, got %d", snap2.TodayTurns)
	}
}

func TestLiveTracker_FirstAppearanceAndReset(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	// Baseline poll: OpenCode only
	res1 := []ToolResult{
		{
			Tool:    ToolOpenCode,
			Name:    "OpenCode",
			Summary: Summary{LifetimeTokens: 100_000},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 10_000, Turns: 5},
			},
		},
	}
	lt.Poll(res1, nil, t1)

	// Poll 2: Claude first appearance with 5M historical tokens
	t2 := t1.Add(time.Second)
	res2 := append([]ToolResult{}, res1...)
	res2 = append(res2, ToolResult{
		Tool:    ToolClaude,
		Name:    "Claude",
		Summary: Summary{LifetimeTokens: 5_000_000},
		Daily: []DailyRow{
			{Day: "2026-09-24", Tokens: 50_000, Turns: 20},
		},
	})

	snap2, events2 := lt.Poll(res2, nil, t2)
	// Claude first appearance must establish baseline without emitting a 5M live burst
	if len(events2) != 0 {
		t.Fatalf("expected 0 events on first appearance of Claude, got %d", len(events2))
	}
	if snap2.SessionTokens != 0 {
		t.Errorf("expected SessionTokens to remain 0 on first appearance, got %d", snap2.SessionTokens)
	}

	// Poll 3: Claude burns 250 tokens
	t3 := t2.Add(time.Second)
	res3 := []ToolResult{
		res1[0],
		{
			Tool:    ToolClaude,
			Name:    "Claude",
			Summary: Summary{LifetimeTokens: 5_000_250},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 50_250, Turns: 21},
			},
		},
	}
	snap3, events3 := lt.Poll(res3, nil, t3)
	if len(events3) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events3))
	}
	if snap3.SessionTokens != 250 {
		t.Errorf("expected SessionTokens 250, got %d", snap3.SessionTokens)
	}

	// Poll 4: Log truncation / counter reset (lifetime tokens resets to 100)
	t4 := t3.Add(time.Second)
	res4 := []ToolResult{
		res1[0],
		{
			Tool:    ToolClaude,
			Name:    "Claude",
			Summary: Summary{LifetimeTokens: 100},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 100, Turns: 1},
			},
		},
	}
	snap4, events4 := lt.Poll(res4, nil, t4)
	if len(events4) != 0 {
		t.Fatalf("expected 0 events on counter reset, got %d", len(events4))
	}
	if snap4.SessionTokens != 250 {
		t.Errorf("expected SessionTokens to preserve accumulated 250 on counter reset, got %d", snap4.SessionTokens)
	}
}

func TestLiveTracker_MidnightTransition(t *testing.T) {
	lt := NewLiveTracker()
	// Day 1: 23:59:58
	t1 := time.Date(2026, 9, 24, 23, 59, 58, 0, time.UTC)

	res1 := []ToolResult{
		{
			Tool:    ToolOpenCode,
			Name:    "OpenCode",
			Summary: Summary{LifetimeTokens: 100_000},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 10_000, Turns: 5},
			},
		},
	}
	lt.Poll(res1, nil, t1)

	// Day 2 (across midnight): 00:00:05
	t2 := time.Date(2026, 9, 25, 0, 0, 5, 0, time.UTC)
	res2 := []ToolResult{
		{
			Tool:    ToolOpenCode,
			Name:    "OpenCode",
			Summary: Summary{LifetimeTokens: 100_050}, // +50 tokens burned post-midnight
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 10_000, Turns: 5},
				{Day: "2026-09-25", Tokens: 50, Turns: 1}, // Today's new row
			},
		},
	}

	snap2, events2 := lt.Poll(res2, nil, t2)
	// TodayTokens should now strictly report the new day's 50 tokens
	if snap2.TodayTokens != 50 {
		t.Errorf("expected TodayTokens 50 after midnight transition, got %d", snap2.TodayTokens)
	}
	if snap2.TodayTurns != 1 {
		t.Errorf("expected TodayTurns 1 after midnight transition, got %d", snap2.TodayTurns)
	}
	if len(events2) != 1 {
		t.Fatalf("expected 1 event for post-midnight burn, got %d", len(events2))
	}
	if events2[0].Tokens != 50 {
		t.Errorf("expected event tokens 50, got %d", events2[0].Tokens)
	}
}

type testLivePricer struct{}

func (p *testLivePricer) PriceDay(day DailyRow) (float64, []string) {
	var cost float64
	for _, m := range day.Models {
		cost += float64(m.Total()) * 0.001
	}
	return cost, nil
}

func TestLiveTracker_PricingAndNoEstimate(t *testing.T) {
	t1 := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	pricer := &testLivePricer{}

	res := []ToolResult{
		{
			Tool:    ToolOpenCode,
			Name:    "OpenCode",
			Summary: Summary{LifetimeTokens: 1000},
			Daily: []DailyRow{
				{
					Day:    "2026-09-24",
					Tokens: 1000,
					Turns:  1,
					Models: map[string]ModelTokens{
						"mock-model": {Output: 1000},
					},
				},
			},
		},
	}

	// 1. With estimation enabled
	ltEst := NewLiveTrackerWithPricer(pricer, false)
	snapEst, _ := ltEst.Poll(res, nil, t1)
	if snapEst.TodayCost != 1.0 {
		t.Errorf("expected estimated todayCost 1.0, got %f", snapEst.TodayCost)
	}

	// 2. With --no-estimate
	ltNoEst := NewLiveTrackerWithPricer(pricer, true)
	snapNoEst, _ := ltNoEst.Poll(res, nil, t1)
	if snapNoEst.TodayCost != 0.0 {
		t.Errorf("expected todayCost 0.0 under no-estimate, got %f", snapNoEst.TodayCost)
	}
}
