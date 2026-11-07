// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"strings"
	"testing"
	"time"
)

func TestResolvePeriodPreset(t *testing.T) {
	// Wednesday, 2026-04-15
	now := time.Date(2026, 4, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		token     string
		startDay  time.Weekday
		wantSince string
		wantUntil string
		wantLabel string
		wantOK    bool
	}{
		{
			token:     "today",
			startDay:  time.Sunday,
			wantSince: "2026-04-15",
			wantUntil: "2026-04-15",
			wantLabel: "today (2026-04-15)",
			wantOK:    true,
		},
		{
			token:     "yesterday",
			startDay:  time.Sunday,
			wantSince: "2026-04-14",
			wantUntil: "2026-04-14",
			wantLabel: "yesterday (2026-04-14)",
			wantOK:    true,
		},
		{
			token:     "lastweek",
			startDay:  time.Sunday,
			wantSince: "2026-04-09",
			wantUntil: "2026-04-15",
			wantLabel: "last week (2026-04-09 to 2026-04-15)",
			wantOK:    true,
		},
		{
			token:     "prevweek",
			startDay:  time.Sunday,
			wantSince: "2026-04-05",
			wantUntil: "2026-04-11",
			wantLabel: "previous week (2026-04-05 to 2026-04-11)",
			wantOK:    true,
		},
		{
			token:     "prevweek",
			startDay:  time.Monday,
			wantSince: "2026-04-06",
			wantUntil: "2026-04-12",
			wantLabel: "previous week (2026-04-06 to 2026-04-12)",
			wantOK:    true,
		},
		{
			token:     "thisweek",
			startDay:  time.Sunday,
			wantSince: "2026-04-12",
			wantUntil: "2026-04-15",
			wantLabel: "this week (2026-04-12 to 2026-04-15)",
			wantOK:    true,
		},
		{
			token:     "lastmonth",
			startDay:  time.Sunday,
			wantSince: "2026-03-17",
			wantUntil: "2026-04-15",
			wantLabel: "last month (2026-03-17 to 2026-04-15)",
			wantOK:    true,
		},
		{
			token:     "prevmonth",
			startDay:  time.Sunday,
			wantSince: "2026-03-01",
			wantUntil: "2026-03-31",
			wantLabel: "previous month (2026-03)",
			wantOK:    true,
		},
		{
			token:     "7d",
			startDay:  time.Sunday,
			wantSince: "2026-04-09",
			wantUntil: "2026-04-15",
			wantLabel: "last 7 days (2026-04-09 to 2026-04-15)",
			wantOK:    true,
		},
		{
			token:     "thismonth",
			startDay:  time.Sunday,
			wantSince: "2026-04-01",
			wantUntil: "2026-04-15",
			wantLabel: "this month (2026-04)",
			wantOK:    true,
		},
		{
			token:     "lastbilling",
			startDay:  time.Sunday,
			wantSince: "2026-03-01",
			wantUntil: "2026-03-31",
			wantLabel: "last billing (2026-03-01 to 2026-03-31)",
			wantOK:    true,
		},
		{
			token:     "billing",
			startDay:  time.Sunday,
			wantSince: "2026-04-01",
			wantUntil: "2026-04-30",
			wantLabel: "billing (2026-04-01 to 2026-04-30)",
			wantOK:    true,
		},
		{
			token:     "annual",
			startDay:  time.Sunday,
			wantSince: "2026-01-01",
			wantUntil: "2026-04-15",
			wantLabel: "annual (2026 YTD)",
			wantOK:    true,
		},
		{
			token:  "notapreset",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		s, u, lbl, ok := ResolvePeriodPreset(tc.token, now, tc.startDay)
		if ok != tc.wantOK {
			t.Fatalf("ResolvePeriodPreset(%q) ok = %v, want %v", tc.token, ok, tc.wantOK)
		}
		if !tc.wantOK {
			continue
		}
		if s != tc.wantSince || u != tc.wantUntil || lbl != tc.wantLabel {
			t.Errorf("ResolvePeriodPreset(%q, %v) = (%q, %q, %q), want (%q, %q, %q)",
				tc.token, tc.startDay, s, u, lbl, tc.wantSince, tc.wantUntil, tc.wantLabel)
		}
	}
}

func TestResolveProjectScope(t *testing.T) {
	projects := []ProjectDay{
		{Project: "/work/sila", Day: "2026-04-10", Tokens: 5000, Tool: "Codex"},
		{Project: "/work/thermal-streak", Day: "2026-04-11", Tokens: 2000, Tool: "OpenCode"},
		{Project: "/work/client/api", Day: "2026-04-11", Tokens: 1000, Tool: "Codex"},
		{Project: "/personal/api", Day: "2026-04-12", Tokens: 500, Tool: "Claude"},
	}

	keys, disp, sugg, ok := ResolveProjectScope("sila", projects)
	if !ok || len(keys) != 1 || keys[0] != "/work/sila" || disp != "sila" || len(sugg) != 0 {
		t.Fatalf("exact match sila: got keys=%v disp=%q sugg=%v ok=%v", keys, disp, sugg, ok)
	}

	// Disambiguated display name match
	keys, disp, _, ok = ResolveProjectScope("client/api", projects)
	if !ok || len(keys) != 1 || keys[0] != "/work/client/api" || disp != "api (client)" {
		t.Fatalf("display match client/api: got keys=%v disp=%q ok=%v", keys, disp, ok)
	}

	// Typo suggestion
	_, _, sugg, ok = ResolveProjectScope("silx", projects)
	if ok || len(sugg) == 0 || sugg[0] != "sila" {
		t.Fatalf("typo silx: want ok=false sugg[0]=sila, got ok=%v sugg=%v", ok, sugg)
	}
}

type stubLeaderboardPricer struct {
	price float64
}

func (s stubLeaderboardPricer) PriceDay(d DailyRow) (float64, []string) {
	var missing []string
	var total float64
	for m, tok := range d.Models {
		if m == "unknown-model" {
			missing = append(missing, m)
			continue
		}
		total += float64(tok.Total()) * s.price
	}
	return total, missing
}

func TestBuildLeaderboardResultsWindowAndProject(t *testing.T) {
	loaded := []ToolResult{
		{
			Tool: ToolCodex,
			Name: "Codex",
			Summary: Summary{
				Tool:           "Codex",
				Sessions:       10,
				LifetimeTokens: 3000,
				Cost:           3.0,
			},
			Daily: []DailyRow{
				{Day: "2026-03-10", Tokens: 1000, Turns: 3, Cost: 1.0},
				{Day: "2026-04-14", Tokens: 1000, Turns: 3, Cost: 1.0},
				{Day: "2026-04-15", Tokens: 1000, Turns: 4, Cost: 1.0},
			},
		},
		{
			Tool: ToolAgy,
			Name: "Agy",
			Summary: Summary{
				Tool:           "Agy",
				Sessions:       50,
				LifetimeTokens: 50,
			},
			Daily: []DailyRow{
				{Day: "2026-03-08", Tokens: 10, Turns: 10},
				{Day: "2026-03-09", Tokens: 10, Turns: 10},
				{Day: "2026-03-10", Tokens: 20, Turns: 20},
				{Day: "2026-04-15", Tokens: 30, Turns: 30},
			},
		},
	}

	projects := []ProjectDay{
		{
			Project: "/work/sila",
			Day:     "2026-04-14",
			Tool:    "Codex",
			Tokens:  400,
			Turns:   2,
			Models:  map[string]ModelTokens{"gpt-5": {Input: 300, Output: 100}},
		},
		{
			Project: "/work/sila",
			Day:     "2026-04-15",
			Tool:    "Codex",
			Tokens:  600,
			Turns:   2,
			Cost:    0.75,
			Models:  map[string]ModelTokens{"gpt-5": {Input: 400, Output: 200}},
		},
		{
			Project: "/work/sila",
			Day:     "2026-04-15",
			Tool:    "Agy",
			Tokens:  12,
			Turns:   12,
		},
	}

	// 1. Today window (2026-04-15): preserves active streak up through today
	// (Codex has a 2-day streak from 04-14..04-15; Agy has 1d current and 3d longest from March)
	// while keeping ActiveDays = 1 and scoping tokens/cost to today.
	todayRes := BuildLeaderboardResults(loaded, projects, LeaderboardOptions{
		Since:    "2026-04-15",
		Until:    "2026-04-15",
		Estimate: true,
	}, stubLeaderboardPricer{price: 0.001})

	if len(todayRes) != 2 {
		t.Fatalf("today results count = %d, want 2", len(todayRes))
	}
	if todayRes[0].TotalActivity != 1000 || todayRes[0].CurrentStreak != 2 || todayRes[0].LongestStreak != 2 || todayRes[0].ActiveDays != 1 || todayRes[0].Summary.Cost != 1.0 {
		t.Errorf("today Codex = %+v", todayRes[0])
	}
	if todayRes[1].TotalActivity != 30 || todayRes[1].CurrentStreak != 1 || todayRes[1].LongestStreak != 3 || todayRes[1].ActiveDays != 1 {
		t.Errorf("today Agy = %+v", todayRes[1])
	}

	// 2. Project filter (/work/sila)
	keys, _, _, ok := ResolveProjectScope("sila", projects)
	if !ok {
		t.Fatalf("ResolveProjectScope(sila) failed")
	}
	projRes := BuildLeaderboardResults(loaded, projects, LeaderboardOptions{
		Project:  strings.Join(keys, "\x00"),
		Estimate: true,
	}, stubLeaderboardPricer{price: 0.001})

	if len(projRes) != 2 {
		t.Fatalf("project sila results count = %d, want 2", len(projRes))
	}
	// On 2026-04-14 Codex recorded Cost=1.0 in Daily, so RecordedCostDays marks 2026-04-14 as recorded;
	// on 2026-04-15 project row has Cost=0.75. Total tokens = 1000, streak = 2d.
	if projRes[0].TotalActivity != 1000 || projRes[0].CurrentStreak != 2 || projRes[0].LongestStreak != 2 || projRes[0].ActiveDays != 2 {
		t.Errorf("project sila Codex = %+v", projRes[0])
	}
	if projRes[1].Name != "Agy" || projRes[1].TotalActivity != 12 || projRes[1].CurrentStreak != 1 {
		t.Errorf("project sila Agy = %+v", projRes[1])
	}
}
