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
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 100_000,
				Sessions:       100,
			},
		},
	}
	proj1 := []ProjectDay{
		{Tool: "OpenCode", Project: "/workspace/projectA", Day: "2026-09-24", Tokens: 80_000, Turns: 80},
		{Tool: "OpenCode", Project: "/workspace/projectB", Day: "2026-09-24", Tokens: 20_000, Turns: 20},
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
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 100_010,
				Sessions:       101,
			},
		},
	}
	proj2 := []ProjectDay{
		{Tool: "OpenCode", Project: "/workspace/projectA", Day: "2026-09-24", Tokens: 80_000, Turns: 80}, // unchanged!
		{Tool: "OpenCode", Project: "/workspace/projectB", Day: "2026-09-24", Tokens: 20_010, Turns: 21}, // +10 tokens, +1 turn
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
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 100_015,
				Sessions:       102,
			},
		},
	}
	proj3 := []ProjectDay{
		{Tool: "OpenCode", Project: "/workspace/projectA", Day: "2026-09-24", Tokens: 80_005, Turns: 81}, // +5 tokens!
		{Tool: "OpenCode", Project: "/workspace/projectB", Day: "2026-09-24", Tokens: 20_010, Turns: 21}, // unchanged!
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

func TestLiveTracker_SeedTodayBaselineAndDelta(t *testing.T) {
	lt := NewLiveTracker()
	lt.SetSeedToday(true)
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

	// 1. First poll with seedToday should initialize session counters to today's totals
	snap, events := lt.Poll(res, nil, now)
	if len(events) != 0 {
		t.Fatalf("expected 0 new delta events on first poll, got %d", len(events))
	}
	if snap.TodayTokens != 1_000_000 {
		t.Errorf("expected todayTokens 1M, got %d", snap.TodayTokens)
	}
	if snap.SessionTokens != 1_000_000 {
		t.Errorf("expected sessionTokens seeded with 1M, got %d", snap.SessionTokens)
	}
	if snap.SessionTurns != 50 {
		t.Errorf("expected sessionTurns seeded with 50, got %d", snap.SessionTurns)
	}
	if snap.SessionCost != 1.25 {
		t.Errorf("expected sessionCost seeded with 1.25, got %f", snap.SessionCost)
	}
	if snap.FlameIntensity <= 0 {
		t.Errorf("expected warm flameIntensity > 0 on seedToday, got %f", snap.FlameIntensity)
	}
	if len(snap.RecentEvents) != 1 {
		t.Fatalf("expected 1 baseline event in RecentEvents, got %d", len(snap.RecentEvents))
	}
	if snap.RecentEvents[0].Tokens != 1_000_000 {
		t.Errorf("expected baseline event tokens 1M, got %d", snap.RecentEvents[0].Tokens)
	}

	// 2. Second poll: +25,000 tokens and +1 turn
	t2 := now.Add(2 * time.Second)
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
	if snap2.SessionTokens != 1_025_000 {
		t.Errorf("expected sessionTokens incremented to 1_025_000, got %d", snap2.SessionTokens)
	}
	if snap2.SessionTurns != 51 {
		t.Errorf("expected sessionTurns incremented to 51, got %d", snap2.SessionTurns)
	}
	if math.Abs(snap2.SessionCost-1.30) > 0.001 {
		t.Errorf("expected sessionCost 1.30, got %f", snap2.SessionCost)
	}

	// 3. ResetSession should clear session counters back to 0
	lt.ResetSession()
	snapReset, _ := lt.Poll(res2, nil, t2.Add(time.Second))
	if snapReset.SessionTokens != 0 {
		t.Errorf("expected sessionTokens 0 after reset, got %d", snapReset.SessionTokens)
	}
	if snapReset.SessionTurns != 0 {
		t.Errorf("expected sessionTurns 0 after reset, got %d", snapReset.SessionTurns)
	}
	if snapReset.SessionCost != 0 {
		t.Errorf("expected sessionCost 0 after reset, got %f", snapReset.SessionCost)
	}

	// 4. Activity-only tool tracking (e.g. Agy)
	ltAct := NewLiveTracker()
	ltAct.SetSeedToday(true)
	actRes1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 0,
				Sessions:       10,
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 0, Turns: 300},
			},
		},
	}
	actSnap1, _ := ltAct.Poll(actRes1, nil, now)
	if actSnap1.SessionTokens != 0 {
		t.Errorf("expected sessionTokens 0 for activity tool, got %d", actSnap1.SessionTokens)
	}
	if actSnap1.SessionTurns != 300 {
		t.Errorf("expected sessionTurns 300 for activity tool, got %d", actSnap1.SessionTurns)
	}
	if actSnap1.ActivityIntensity <= 0 {
		t.Errorf("expected warm activityIntensity > 0 for activity tool on seedToday")
	}
	if actSnap1.FlameIntensity != 0 {
		t.Errorf("expected flameIntensity 0 for activity tool, got %f", actSnap1.FlameIntensity)
	}

	// Delta turns for activity tool
	actRes2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 0,
				Sessions:       11,
			},
			Daily: []DailyRow{
				{Day: "2026-09-24", Tokens: 0, Turns: 310},
			},
		},
	}
	actSnap2, actEvents2 := ltAct.Poll(actRes2, nil, now.Add(2*time.Second))
	if len(actEvents2) != 1 {
		t.Fatalf("expected 1 activity delta event, got %d", len(actEvents2))
	}
	if actEvents2[0].Turns != 10 || actEvents2[0].Tokens != 0 {
		t.Errorf("expected delta event 10 turns and 0 tokens, got turns=%d tok=%d", actEvents2[0].Turns, actEvents2[0].Tokens)
	}
	if actSnap2.SessionTurns != 310 {
		t.Errorf("expected sessionTurns 310, got %d", actSnap2.SessionTurns)
	}
	if actSnap2.BurnTurnsPerMin <= 0 {
		t.Errorf("expected BurnTurnsPerMin > 0 for activity turns, got %f", actSnap2.BurnTurnsPerMin)
	}
	if actSnap2.BurnTokensPerMin != 0 {
		t.Errorf("expected BurnTokensPerMin 0 for activity turns, got %f", actSnap2.BurnTokensPerMin)
	}
	if actSnap2.ActivityIntensity <= 0 {
		t.Errorf("expected activityIntensity > 0 after activity turn delta, got %f", actSnap2.ActivityIntensity)
	}
	if actSnap2.FlameIntensity != 0 {
		t.Errorf("expected flameIntensity 0 after activity turn delta, got %f", actSnap2.FlameIntensity)
	}
}

func TestLiveTracker_ResumedSessionProjectAndModel(t *testing.T) {
	lt := NewLiveTracker()
	// Mock time is 2026-09-25 (today)
	t1 := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

	// Session was created on 2026-09-18 (a week ago)
	res1 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 251_000_000,
				Sessions:       10,
				Cost:           15.00,
				ModelBreakdown: map[string]int64{
					"deepseek-flash": 10,
				},
			},
			Daily: []DailyRow{
				{
					Day:    "2026-09-18",
					Tokens: 251_000_000,
					Turns:  500,
					Models: map[string]ModelTokens{
						"deepseek-flash": {Input: 10_000_000, Output: 241_000_000},
					},
				},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/jad/.config/opencode",
			Day:     "2026-09-18",
			Tokens:  251_000_000,
			Turns:   500,
		},
	}

	// Poll 1: Initial baseline
	snap1, events1 := lt.Poll(res1, proj1, t1)
	if len(events1) != 0 {
		t.Fatalf("expected 0 baseline events, got %d", len(events1))
	}
	// ActiveProject and ActiveModel should fall back to the resumed session's project/model
	if snap1.ActiveProject != "opencode" {
		t.Errorf("expected activeProject 'opencode' on baseline, got %q", snap1.ActiveProject)
	}
	if snap1.ActiveModel != "deepseek-flash" {
		t.Errorf("expected activeModel 'deepseek-flash' on baseline, got %q", snap1.ActiveModel)
	}

	// Poll 2: The resumed session burns 379,500 tokens in the same project!
	t2 := t1.Add(time.Second)
	res2 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 251_379_500,
				Sessions:       10,
				Cost:           15.02,
				ModelBreakdown: map[string]int64{
					"deepseek-flash": 10,
				},
			},
			Daily: []DailyRow{
				{
					Day:    "2026-09-18",
					Tokens: 251_379_500,
					Turns:  501,
					Models: map[string]ModelTokens{
						"deepseek-flash": {Input: 10_000_000, Output: 241_379_500},
					},
				},
			},
		},
	}
	proj2 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/jad/.config/opencode",
			Day:     "2026-09-18",
			Tokens:  251_379_500,
			Turns:   501,
		},
	}

	snap2, events2 := lt.Poll(res2, proj2, t2)
	if len(events2) != 1 {
		t.Fatalf("expected 1 delta event for resumed session, got %d", len(events2))
	}
	ev := events2[0]
	if ev.Tool != "OpenCode" {
		t.Errorf("expected Tool OpenCode, got %q", ev.Tool)
	}
	if ev.Tokens != 379_500 {
		t.Errorf("expected Tokens 379,500, got %d", ev.Tokens)
	}
	if ev.Project != "opencode" {
		t.Errorf("expected Project 'opencode', got %q (must not be empty!)", ev.Project)
	}
	if ev.Model != "deepseek-flash" {
		t.Errorf("expected Model 'deepseek-flash', got %q (must not be empty!)", ev.Model)
	}
	if snap2.ActiveProject != "opencode" {
		t.Errorf("expected snap2.ActiveProject 'opencode', got %q", snap2.ActiveProject)
	}
	if snap2.ActiveModel != "deepseek-flash" {
		t.Errorf("expected snap2.ActiveModel 'deepseek-flash', got %q", snap2.ActiveModel)
	}
}

func TestLiveTracker_IdlePollBaselinePreservation(t *testing.T) {
	lt := NewLiveTracker()

	t1, err := time.Parse("2006-01-02 15:04:05", "2026-10-19 10:00:00")
	if err != nil {
		t.Fatalf("failed to parse t1: %v", err)
	}

	// Historical activity: Day 2026-10-18 has 125,600 turns in project lab
	res1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 0,
				Sessions:       125_600,
			},
			Daily: []DailyRow{
				{
					Day:   "2026-10-18",
					Turns: 125_600,
				},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "Agy",
			Project: "/mnt/Jad/github/lab/thermal-streak",
			Day:     "2026-10-18",
			Turns:   125_600,
		},
	}

	// 1. Initial poll
	snap1, events1 := lt.Poll(res1, proj1, t1)
	if len(events1) != 0 {
		t.Fatalf("expected 0 events on initial poll, got %d", len(events1))
	}
	if snap1.SessionTurns != 0 {
		t.Errorf("expected 0 session turns initially, got %d", snap1.SessionTurns)
	}

	// 2. Sequential idle polls (zero delta)
	for i := 1; i <= 3; i++ {
		tIdle := t1.Add(time.Duration(i) * time.Second)
		snapIdle, eventsIdle := lt.Poll(res1, proj1, tIdle)
		if len(eventsIdle) != 0 {
			t.Fatalf("idle poll %d: expected 0 events, got %d", i, len(eventsIdle))
		}
		if snapIdle.SessionTurns != 0 {
			t.Errorf("idle poll %d: expected 0 session turns, got %d", i, snapIdle.SessionTurns)
		}
	}

	// 3. +1 turn update arrives on today's day (2026-10-19)
	t2 := t1.Add(10 * time.Second)
	res2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 0,
				Sessions:       125_601,
			},
			Daily: []DailyRow{
				{
					Day:   "2026-10-18",
					Turns: 125_600,
				},
				{
					Day:   "2026-10-19",
					Turns: 1,
				},
			},
		},
	}
	proj2 := []ProjectDay{
		{
			Tool:    "Agy",
			Project: "/mnt/Jad/github/lab/thermal-streak",
			Day:     "2026-10-18",
			Turns:   125_600,
		},
		{
			Tool:    "Agy",
			Project: "/mnt/Jad/github/lab/thermal-streak",
			Day:     "2026-10-19",
			Turns:   1,
		},
	}

	snap2, events2 := lt.Poll(res2, proj2, t2)
	if len(events2) != 1 {
		t.Fatalf("expected exactly 1 event on +1 turn update, got %d", len(events2))
	}
	if events2[0].Turns != 1 {
		t.Errorf("expected event Turns: 1, got %d (phantom burst!)", events2[0].Turns)
	}
	if events2[0].Project != "thermal-streak" {
		t.Errorf("expected event Project 'thermal-streak', got %q", events2[0].Project)
	}

	var sumTurns int
	for _, ev := range snap2.RecentEvents {
		sumTurns += ev.Turns
	}
	if sumTurns != snap2.SessionTurns {
		t.Errorf("invariant violation: sum(RecentEvents.Turns)=%d != SessionTurns=%d", sumTurns, snap2.SessionTurns)
	}
	if snap2.SessionTurns != 1 {
		t.Errorf("expected SessionTurns 1, got %d", snap2.SessionTurns)
	}
}

func TestLiveTracker_IdlePollNoDuplicateBursts(t *testing.T) {
	lt := NewLiveTracker()

	t1, err := time.Parse("2006-01-02 15:04:05", "2026-10-19 12:00:00")
	if err != nil {
		t.Fatalf("failed to parse t1: %v", err)
	}

	res := []ToolResult{
		{
			Tool: ToolClaude,
			Name: "Claude",
			Summary: Summary{
				LifetimeTokens: 50_000,
				Sessions:       10,
			},
			Daily: []DailyRow{
				{
					Day:    "2026-10-18",
					Tokens: 40_000,
					Turns:  8,
				},
				{
					Day:    "2026-10-19",
					Tokens: 10_000,
					Turns:  2,
				},
			},
		},
	}
	proj := []ProjectDay{
		{
			Tool:    "Claude",
			Project: "/home/user/code/proj",
			Day:     "2026-10-18",
			Tokens:  40_000,
			Turns:   8,
		},
		{
			Tool:    "Claude",
			Project: "/home/user/code/proj",
			Day:     "2026-10-19",
			Tokens:  10_000,
			Turns:   2,
		},
	}

	_, eventsInit := lt.Poll(res, proj, t1)
	if len(eventsInit) != 0 {
		t.Fatalf("expected 0 events on init, got %d", len(eventsInit))
	}

	// Repeated identical polls
	for i := 1; i <= 5; i++ {
		tStep := t1.Add(time.Duration(i) * time.Second)
		snap, events := lt.Poll(res, proj, tStep)
		if len(events) != 0 {
			t.Fatalf("poll %d with zero delta produced %d events (expected 0)", i, len(events))
		}
		if snap.SessionTokens != 0 || snap.SessionTurns != 0 {
			t.Fatalf("poll %d mutated session counters without delta: tokens=%d, turns=%d", i, snap.SessionTokens, snap.SessionTurns)
		}
	}
}

func TestLiveTracker_MidnightTransitionBaseline(t *testing.T) {
	lt := NewLiveTracker()

	// 23:59:50 on Day 1
	tPre, err := time.Parse("2006-01-02 15:04:05", "2026-10-19 23:59:50")
	if err != nil {
		t.Fatalf("failed to parse tPre: %v", err)
	}

	resPre := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 1_000_000,
				Sessions:       100,
			},
			Daily: []DailyRow{
				{
					Day:    "2026-10-19",
					Tokens: 1_000_000,
					Turns:  100,
				},
			},
		},
	}
	projPre := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-19",
			Tokens:  1_000_000,
			Turns:   100,
		},
	}

	_, eventsPre := lt.Poll(resPre, projPre, tPre)
	if len(eventsPre) != 0 {
		t.Fatalf("expected 0 events on init, got %d", len(eventsPre))
	}

	// Midnight crossing at 00:00:01 on Day 2 (no new activity yet)
	tMid, _ := time.Parse("2006-01-02 15:04:05", "2026-10-20 00:00:01")
	snapMid, eventsMid := lt.Poll(resPre, projPre, tMid)
	if len(eventsMid) != 0 {
		t.Fatalf("expected 0 events on midnight crossing without activity, got %d", len(eventsMid))
	}
	if snapMid.TodayTokens != 0 || snapMid.TodayTurns != 0 {
		t.Errorf("expected todayTokens and todayTurns 0 on new day, got tokens=%d turns=%d", snapMid.TodayTokens, snapMid.TodayTurns)
	}

	// +500 tokens and +2 turns on Day 2 at 00:00:10
	tPost, _ := time.Parse("2006-01-02 15:04:05", "2026-10-20 00:00:10")
	resPost := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 1_000_500,
				Sessions:       102,
			},
			Daily: []DailyRow{
				{
					Day:    "2026-10-19",
					Tokens: 1_000_000,
					Turns:  100,
				},
				{
					Day:    "2026-10-20",
					Tokens: 500,
					Turns:  2,
				},
			},
		},
	}
	projPost := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-19",
			Tokens:  1_000_000,
			Turns:   100,
		},
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-20",
			Tokens:  500,
			Turns:   2,
		},
	}

	snapPost, eventsPost := lt.Poll(resPost, projPost, tPost)
	if len(eventsPost) != 1 {
		t.Fatalf("expected 1 event for post-midnight increment, got %d", len(eventsPost))
	}
	ev := eventsPost[0]
	if ev.Tokens != 500 {
		t.Errorf("expected 500 tokens in post-midnight event, got %d (phantom burst!)", ev.Tokens)
	}
	if ev.Turns != 2 {
		t.Errorf("expected 2 turns in post-midnight event, got %d", ev.Turns)
	}
	if ev.Project != "app" {
		t.Errorf("expected Project 'app', got %q", ev.Project)
	}

	var sumTurns int
	for _, e := range snapPost.RecentEvents {
		sumTurns += e.Turns
	}
	if sumTurns != snapPost.SessionTurns {
		t.Errorf("invariant violation: sum(RecentEvents.Turns)=%d != SessionTurns=%d", sumTurns, snapPost.SessionTurns)
	}
}

func TestLiveTracker_MidnightTransitionNoPhantomReplay(t *testing.T) {
	lt := NewLiveTracker()

	t1, _ := time.Parse("2006-01-02 15:04:05", "2026-10-19 23:59:55")
	t2, _ := time.Parse("2006-01-02 15:04:05", "2026-10-20 00:00:05")

	res := []ToolResult{
		{
			Tool: ToolCodex,
			Name: "Codex",
			Summary: Summary{
				LifetimeTokens: 75_000,
				Sessions:       5,
			},
			Daily: []DailyRow{
				{
					Day:    "2026-10-19",
					Tokens: 75_000,
					Turns:  5,
				},
			},
		},
	}
	proj := []ProjectDay{
		{
			Tool:    "Codex",
			Project: "/mnt/Jad/github/lab/thermal-streak",
			Day:     "2026-10-19",
			Tokens:  75_000,
			Turns:   5,
		},
	}

	lt.Poll(res, proj, t1)
	snap2, events2 := lt.Poll(res, proj, t2)
	if len(events2) != 0 {
		t.Fatalf("expected 0 events across midnight when data is unchanged, got %d", len(events2))
	}
	if snap2.SessionTurns != 0 {
		t.Errorf("expected session turns 0, got %d", snap2.SessionTurns)
	}
}

func TestLiveTracker_ProjectDeltaClamping(t *testing.T) {
	lt := NewLiveTracker()

	t1, _ := time.Parse("2006-01-02 15:04:05", "2026-10-19 14:00:00")
	t2, _ := time.Parse("2006-01-02 15:04:05", "2026-10-19 14:00:05")

	res1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 0,
				Sessions:       10,
			},
			Daily: []DailyRow{
				{
					Day:   "2026-10-19",
					Turns: 10,
				},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "Agy",
			Project: "/mnt/Jad/github/lab",
			Day:     "2026-10-19",
			Turns:   10,
		},
	}

	lt.Poll(res1, proj1, t1)

	// Tool delta is +2 turns. But anomalous project telemetry reports +10,000 turns!
	res2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 0,
				Sessions:       12, // +2 turns
			},
			Daily: []DailyRow{
				{
					Day:   "2026-10-19",
					Turns: 12,
				},
			},
		},
	}
	proj2 := []ProjectDay{
		{
			Tool:    "Agy",
			Project: "/mnt/Jad/github/lab",
			Day:     "2026-10-19",
			Turns:   10_010, // anomalous +10,000 turns!
		},
	}

	snap2, events2 := lt.Poll(res2, proj2, t2)
	if len(events2) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events2))
	}
	if events2[0].Turns != 2 {
		t.Errorf("expected clamped Turns: 2, got %d", events2[0].Turns)
	}
	if snap2.SessionTurns != 2 {
		t.Errorf("expected SessionTurns 2, got %d", snap2.SessionTurns)
	}

	var sumTurns int
	for _, ev := range snap2.RecentEvents {
		sumTurns += ev.Turns
	}
	if sumTurns != snap2.SessionTurns {
		t.Errorf("invariant violation: sum(RecentEvents.Turns)=%d != SessionTurns=%d", sumTurns, snap2.SessionTurns)
	}
}

func TestLiveTracker_ProjectDeltaExceedsToolRejected(t *testing.T) {
	lt := NewLiveTracker()

	t1, _ := time.Parse("2006-01-02 15:04:05", "2026-10-19 15:00:00")
	t2, _ := time.Parse("2006-01-02 15:04:05", "2026-10-19 15:00:05")

	res := []ToolResult{
		{
			Tool: ToolDevin,
			Name: "Devin",
			Summary: Summary{
				LifetimeTokens: 200_000,
				Sessions:       20,
			},
			Daily: []DailyRow{
				{
					Day:    "2026-10-19",
					Tokens: 200_000,
					Turns:  20,
				},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "Devin",
			Project: "/home/user/devin-proj",
			Day:     "2026-10-19",
			Tokens:  200_000,
			Turns:   20,
		},
	}
	proj2 := []ProjectDay{
		{
			Tool:    "Devin",
			Project: "/home/user/devin-proj",
			Day:     "2026-10-19",
			Tokens:  300_000, // anomalous change without tool delta
			Turns:   30,
		},
	}

	lt.Poll(res, proj1, t1)
	snap2, events2 := lt.Poll(res, proj2, t2)
	if len(events2) != 0 {
		t.Fatalf("expected 0 events when tool delta is 0, got %d", len(events2))
	}
	if snap2.SessionTokens != 0 || snap2.SessionTurns != 0 {
		t.Errorf("expected 0 session counters, got tokens=%d turns=%d", snap2.SessionTokens, snap2.SessionTurns)
	}
}

func TestLiveTracker_ActivityToolDifferentLifetimeTokens(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Second)

	// Agy has LifetimeTokens: 100 (lifetime steps from legacy), but Daily turns: 10
	res1 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 100,
				Sessions:       100,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 10, Turns: 10},
			},
		},
	}

	snap1, events1 := lt.Poll(res1, nil, t1)
	if len(events1) != 0 {
		t.Fatalf("expected 0 events on init, got %d", len(events1))
	}
	if snap1.SessionTokens != 0 {
		t.Fatalf("expected 0 session tokens on init, got %d", snap1.SessionTokens)
	}

	// +1 step on Agy (LifetimeTokens becomes 101, Daily turns becomes 11)
	res2 := []ToolResult{
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				LifetimeTokens: 101,
				Sessions:       101,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 11, Turns: 11},
			},
		},
	}

	snap2, events2 := lt.Poll(res2, nil, t2)
	if len(events2) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events2))
	}
	if events2[0].Tokens != 0 {
		t.Errorf("expected 0 tokens for activity tool event, got %d", events2[0].Tokens)
	}
	if events2[0].Turns != 1 {
		t.Errorf("expected 1 turn for activity tool event, got %d", events2[0].Turns)
	}
	if snap2.SessionTokens != 0 {
		t.Errorf("expected SessionTokens to remain 0, got %d", snap2.SessionTokens)
	}
	if snap2.SessionTurns != 1 {
		t.Errorf("expected SessionTurns 1, got %d", snap2.SessionTurns)
	}
}

func TestLiveTracker_MultiProjectNoPhantomTokenDuplication(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Second)

	res1 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 1_000,
				Sessions:       10,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 1_000, Turns: 10},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/projA",
			Day:     "2026-10-19",
			Tokens:  1_000,
			Turns:   5,
		},
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/projB",
			Day:     "2026-10-19",
			Tokens:  0,
			Turns:   5,
		},
	}

	lt.Poll(res1, proj1, t1)

	// Tool delta: +100 tokens, +2 turns
	// projA gets +100 tokens, 0 turns
	// projB gets 0 tokens, +2 turns
	res2 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 1_100,
				Sessions:       12,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 1_100, Turns: 12},
			},
		},
	}
	proj2 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/projA",
			Day:     "2026-10-19",
			Tokens:  1_100, // +100 tokens, 0 turns
			Turns:   5,
		},
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/projB",
			Day:     "2026-10-19",
			Tokens:  0, // 0 tokens, +2 turns
			Turns:   7,
		},
	}

	snap2, events2 := lt.Poll(res2, proj2, t2)
	var totalEvTokens int64
	var totalEvTurns int
	for _, ev := range events2 {
		totalEvTokens += ev.Tokens
		totalEvTurns += ev.Turns
		if ev.Project == "projB" && ev.Tokens != 0 {
			t.Errorf("projB should have 0 tokens, but got %d (phantom token inflation!)", ev.Tokens)
		}
	}

	if totalEvTokens != 100 {
		t.Errorf("expected total emitted tokens strictly 100, got %d (phantom burst!)", totalEvTokens)
	}
	if totalEvTurns != 2 {
		t.Errorf("expected total emitted turns strictly 2, got %d", totalEvTurns)
	}
	if snap2.SessionTokens != 100 {
		t.Errorf("expected SessionTokens 100, got %d", snap2.SessionTokens)
	}
	if snap2.SessionTurns != 2 {
		t.Errorf("expected SessionTurns 2, got %d", snap2.SessionTurns)
	}
}

func TestLiveTracker_LateToolDiscoveryBaselineNotDoubled(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Second)
	t3 := t2.Add(time.Second)

	// Poll 1: Tool Devin is NOT in results, but its historical project is in projDays
	proj := []ProjectDay{
		{
			Tool:    "Devin",
			Project: "/home/user/devin-app",
			Day:     "2026-10-19",
			Tokens:  50_000,
			Turns:   10,
		},
	}
	lt.Poll(nil, proj, t1)

	// Poll 2: Devin appears in results for the first time
	res2 := []ToolResult{
		{
			Tool: ToolDevin,
			Name: "Devin",
			Summary: Summary{
				LifetimeTokens: 50_000,
				Sessions:       10,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 50_000, Turns: 10},
			},
		},
	}
	snap2, events2 := lt.Poll(res2, proj, t2)
	if len(events2) != 0 {
		t.Fatalf("expected 0 events on first discovery of Devin, got %d", len(events2))
	}
	if snap2.SessionTokens != 0 {
		t.Errorf("expected SessionTokens 0, got %d", snap2.SessionTokens)
	}

	// Poll 3: Devin burns +100 tokens and +1 turn
	res3 := []ToolResult{
		{
			Tool: ToolDevin,
			Name: "Devin",
			Summary: Summary{
				LifetimeTokens: 50_100,
				Sessions:       11,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 50_100, Turns: 11},
			},
		},
	}
	proj3 := []ProjectDay{
		{
			Tool:    "Devin",
			Project: "/home/user/devin-app",
			Day:     "2026-10-19",
			Tokens:  50_100,
			Turns:   11,
		},
	}
	snap3, events3 := lt.Poll(res3, proj3, t3)
	if len(events3) != 1 {
		t.Fatalf("expected 1 event on poll 3, got %d", len(events3))
	}
	if events3[0].Tokens != 100 {
		t.Errorf("expected 100 tokens in event, got %d (baseline was likely doubled and delta swallowed!)", events3[0].Tokens)
	}
	if events3[0].Turns != 1 {
		t.Errorf("expected 1 turn in event, got %d", events3[0].Turns)
	}
	if snap3.SessionTokens != 100 {
		t.Errorf("expected SessionTokens 100, got %d", snap3.SessionTokens)
	}
}

func TestLiveTracker_MultiProjectTurnSortingClamping(t *testing.T) {
	lt := NewLiveTracker()
	t1 := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Second)

	res1 := []ToolResult{
		{
			Tool:    ToolAgy,
			Name:    "Agy",
			Summary: Summary{LifetimeTokens: 0, Sessions: 20},
			Daily:   []DailyRow{{Day: "2026-10-19", Turns: 20}},
		},
	}
	proj1 := []ProjectDay{
		{Tool: "Agy", Project: "/repo/aaa", Day: "2026-10-19", Turns: 10},
		{Tool: "Agy", Project: "/repo/zzz", Day: "2026-10-19", Turns: 10},
	}
	lt.Poll(res1, proj1, t1)

	// Tool delta is +5 turns
	// aaa reports +1 turn (Turns: 11)
	// zzz reports +10 turns (Turns: 20)
	// Clamping budget is 5 turns total. Higher delta (zzz: 10) must be prioritized over lower delta (aaa: 1).
	res2 := []ToolResult{
		{
			Tool:    ToolAgy,
			Name:    "Agy",
			Summary: Summary{LifetimeTokens: 0, Sessions: 25},
			Daily:   []DailyRow{{Day: "2026-10-19", Turns: 25}},
		},
	}
	proj2 := []ProjectDay{
		{Tool: "Agy", Project: "/repo/aaa", Day: "2026-10-19", Turns: 11},
		{Tool: "Agy", Project: "/repo/zzz", Day: "2026-10-19", Turns: 20},
	}

	snap2, events2 := lt.Poll(res2, proj2, t2)
	var sumTurns int
	for _, ev := range events2 {
		sumTurns += ev.Turns
	}
	if sumTurns != 5 {
		t.Errorf("expected sum of event turns strictly 5, got %d", sumTurns)
	}
	if snap2.SessionTurns != 5 {
		t.Errorf("expected SessionTurns 5, got %d", snap2.SessionTurns)
	}
	// zzz should get 5 turns (clamped from 10) because it was prioritized
	if len(events2) > 0 && events2[0].Project == "zzz" && events2[0].Turns != 5 {
		t.Errorf("expected prioritized zzz to get 5 turns, got %d", events2[0].Turns)
	}
}

func TestLiveTracker_ClockJumpBackwardsAcrossMidnight(t *testing.T) {
	lt := NewLiveTracker()

	// Step 1: Poll on Day 2 (2026-10-20 00:01:00)
	t1 := time.Date(2026, 10, 20, 0, 1, 0, 0, time.UTC)
	res1 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 1_000_000,
				Sessions:       100,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 900_000, Turns: 90},
				{Day: "2026-10-20", Tokens: 100_000, Turns: 10},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-19",
			Tokens:  900_000,
			Turns:   90,
		},
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-20",
			Tokens:  100_000,
			Turns:   10,
		},
	}

	snap1, ev1 := lt.Poll(res1, proj1, t1)
	if len(ev1) != 0 {
		t.Fatalf("expected 0 events on init, got %d", len(ev1))
	}
	if snap1.TodayTokens != 100_000 {
		t.Errorf("expected TodayTokens 100,000, got %d", snap1.TodayTokens)
	}

	// Step 2: System clock jumps backwards across midnight to 2026-10-19 23:59:50
	t2 := time.Date(2026, 10, 19, 23, 59, 50, 0, time.UTC)
	snap2, ev2 := lt.Poll(res1, proj1, t2)
	if len(ev2) != 0 {
		t.Fatalf("expected 0 events on backward clock jump without delta, got %d", len(ev2))
	}
	if snap2.TodayTokens != 900_000 {
		t.Errorf("expected TodayTokens 900,000 for Day 1, got %d", snap2.TodayTokens)
	}
	if snap2.SessionTokens != 0 {
		t.Errorf("expected SessionTokens 0, got %d", snap2.SessionTokens)
	}

	// Step 3: Clock advances forward again across midnight to 2026-10-20 00:00:05 with +50 tokens and +1 turn
	t3 := time.Date(2026, 10, 20, 0, 0, 5, 0, time.UTC)
	res3 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 1_000_050,
				Sessions:       101,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 900_000, Turns: 90},
				{Day: "2026-10-20", Tokens: 100_050, Turns: 11},
			},
		},
	}
	proj3 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-19",
			Tokens:  900_000,
			Turns:   90,
		},
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-20",
			Tokens:  100_050,
			Turns:   11,
		},
	}

	snap3, ev3 := lt.Poll(res3, proj3, t3)
	if len(ev3) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ev3))
	}
	if ev3[0].Tokens != 50 {
		t.Errorf("expected 50 tokens, got %d", ev3[0].Tokens)
	}
	if ev3[0].Turns != 1 {
		t.Errorf("expected 1 turn, got %d", ev3[0].Turns)
	}
	if snap3.SessionTokens != 50 {
		t.Errorf("expected SessionTokens 50, got %d", snap3.SessionTokens)
	}
	if snap3.TodayTokens != 100_050 {
		t.Errorf("expected TodayTokens 100,050, got %d", snap3.TodayTokens)
	}
}

func TestLiveTracker_MultiDaySleepWakeup(t *testing.T) {
	lt := NewLiveTracker()

	// Monday initial poll
	t1 := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	res1 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 10_000,
				Sessions:       10,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 10_000, Turns: 10},
			},
		},
	}
	proj1 := []ProjectDay{
		{
			Tool:    "OpenCode",
			Project: "/home/user/code/app",
			Day:     "2026-10-19",
			Tokens:  10_000,
			Turns:   10,
		},
	}

	snap1, ev1 := lt.Poll(res1, proj1, t1)
	if len(ev1) != 0 {
		t.Fatalf("expected 0 events on init, got %d", len(ev1))
	}
	if snap1.TodayTokens != 10_000 {
		t.Errorf("expected TodayTokens 10,000, got %d", snap1.TodayTokens)
	}

	// Thursday wakeup (3 days later) with activity across intervening days
	t2 := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	res2 := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Summary: Summary{
				LifetimeTokens: 13_500,
				Sessions:       13,
			},
			Daily: []DailyRow{
				{Day: "2026-10-19", Tokens: 10_000, Turns: 10},
				{Day: "2026-10-20", Tokens: 2_000, Turns: 1},
				{Day: "2026-10-21", Tokens: 1_000, Turns: 1},
				{Day: "2026-10-22", Tokens: 500, Turns: 1},
			},
		},
	}
	proj2 := []ProjectDay{
		{Tool: "OpenCode", Project: "/home/user/code/app", Day: "2026-10-19", Tokens: 10_000, Turns: 10},
		{Tool: "OpenCode", Project: "/home/user/code/app", Day: "2026-10-20", Tokens: 2_000, Turns: 1},
		{Tool: "OpenCode", Project: "/home/user/code/app", Day: "2026-10-21", Tokens: 1_000, Turns: 1},
		{Tool: "OpenCode", Project: "/home/user/code/app", Day: "2026-10-22", Tokens: 500, Turns: 1},
	}

	snap2, ev2 := lt.Poll(res2, proj2, t2)
	if snap2.TodayTokens != 500 {
		t.Errorf("expected TodayTokens 500 on Thursday, got %d", snap2.TodayTokens)
	}
	if snap2.SessionTokens != 3_500 {
		t.Errorf("expected SessionTokens 3,500, got %d", snap2.SessionTokens)
	}
	if snap2.SessionTurns != 3 {
		t.Errorf("expected SessionTurns 3, got %d", snap2.SessionTurns)
	}

	var sumEvTok int64
	var sumEvTurns int
	for _, e := range ev2 {
		sumEvTok += e.Tokens
		sumEvTurns += e.Turns
	}
	if sumEvTok != 3_500 {
		t.Errorf("expected sum of event tokens 3,500, got %d", sumEvTok)
	}
	if sumEvTurns != 3 {
		t.Errorf("expected sum of event turns 3, got %d", sumEvTurns)
	}

	var sumRecentTurns int
	for _, e := range snap2.RecentEvents {
		sumRecentTurns += e.Turns
	}
	if sumRecentTurns != snap2.SessionTurns {
		t.Errorf("invariant violated: sum(RecentEvents.Turns)=%d != SessionTurns=%d", sumRecentTurns, snap2.SessionTurns)
	}

	// Sequential idle poll on Thursday
	t3 := t2.Add(time.Second)
	snap3, ev3 := lt.Poll(res2, proj2, t3)
	if len(ev3) != 0 {
		t.Fatalf("expected 0 events on idle poll, got %d", len(ev3))
	}
	if snap3.SessionTokens != 3_500 {
		t.Errorf("expected SessionTokens preserved at 3,500, got %d", snap3.SessionTokens)
	}
}

func TestLiveTracker_ActivityMidnightRollover(t *testing.T) {
	lt := NewLiveTracker()

	// Initial poll on Day 1 (2026-10-19 23:59:00) with 100 turns in project lab
	t1 := time.Date(2026, 10, 19, 23, 59, 0, 0, time.UTC)
	res1 := []ToolResult{
		{
			Tool:    ToolAgy,
			Name:    "Agy",
			Summary: Summary{LifetimeTokens: 0, Sessions: 100},
			Daily:   []DailyRow{{Day: "2026-10-19", Turns: 100}},
		},
	}
	proj1 := []ProjectDay{
		{Tool: "Agy", Project: "/repo/lab", Day: "2026-10-19", Turns: 100},
	}

	snap1, ev1 := lt.Poll(res1, proj1, t1)
	if len(ev1) != 0 {
		t.Fatalf("expected 0 events on init, got %d", len(ev1))
	}
	if snap1.TodayTurns != 100 {
		t.Errorf("expected TodayTurns 100 on Day 1, got %d", snap1.TodayTurns)
	}

	// Day 2 (2026-10-20 00:01:00) with +1 turn in project lab
	t2 := time.Date(2026, 10, 20, 0, 1, 0, 0, time.UTC)
	res2 := []ToolResult{
		{
			Tool:    ToolAgy,
			Name:    "Agy",
			Summary: Summary{LifetimeTokens: 0, Sessions: 101},
			Daily: []DailyRow{
				{Day: "2026-10-19", Turns: 100},
				{Day: "2026-10-20", Turns: 1},
			},
		},
	}
	proj2 := []ProjectDay{
		{Tool: "Agy", Project: "/repo/lab", Day: "2026-10-19", Turns: 100},
		{Tool: "Agy", Project: "/repo/lab", Day: "2026-10-20", Turns: 1},
	}

	snap2, ev2 := lt.Poll(res2, proj2, t2)
	if len(ev2) != 1 {
		t.Fatalf("expected exactly 1 event on midnight +1 turn update, got %d", len(ev2))
	}
	if ev2[0].Turns != 1 {
		t.Errorf("expected event Turns: 1, got %d", ev2[0].Turns)
	}
	if ev2[0].Tokens != 0 {
		t.Errorf("expected event Tokens: 0 for activity tool, got %d", ev2[0].Tokens)
	}
	if ev2[0].Project != "lab" {
		t.Errorf("expected event Project 'lab', got %q", ev2[0].Project)
	}
	if snap2.TodayTurns != 1 {
		t.Errorf("expected TodayTurns 1 on Day 2, got %d", snap2.TodayTurns)
	}
	if snap2.SessionTurns != 1 {
		t.Errorf("expected SessionTurns 1, got %d", snap2.SessionTurns)
	}

	var sumTurns int
	for _, e := range snap2.RecentEvents {
		sumTurns += e.Turns
	}
	if sumTurns != snap2.SessionTurns {
		t.Errorf("invariant violated: sum(RecentEvents.Turns)=%d != SessionTurns=%d", sumTurns, snap2.SessionTurns)
	}

	// Idle poll on Day 2
	t3 := t2.Add(time.Second)
	snap3, ev3 := lt.Poll(res2, proj2, t3)
	if len(ev3) != 0 {
		t.Fatalf("expected 0 events on idle poll, got %d", len(ev3))
	}
	if snap3.SessionTurns != 1 {
		t.Errorf("expected SessionTurns preserved at 1, got %d", snap3.SessionTurns)
	}
}
