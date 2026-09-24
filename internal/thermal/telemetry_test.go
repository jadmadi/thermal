// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"testing"
	"time"
)

type mockTelemetryPricer struct{}

func (m *mockTelemetryPricer) PriceDay(day DailyRow) (float64, []string) {
	for model, mt := range day.Models {
		if model == "claude-3-5-sonnet" {
			return float64(mt.Total()) * 0.015, nil
		}
	}
	return 0, nil
}

func TestAggregateTelemetry_MixedActivityAndCostParity(t *testing.T) {
	// Claude fixture:
	// Day 1: 100 tokens, 1 turn, recorded cost 1.50
	// Day 2: 50 tokens, 1 turn, zero recorded cost, estimated via mockTelemetryPricer at 50 * 0.015 = 0.75
	claudeDaily := []DailyRow{
		{
			Day:    "2026-09-20",
			Tokens: 100,
			Turns:  1,
			Input:  60,
			Output: 40,
			Cost:   1.50,
			Models: map[string]ModelTokens{
				"claude-3-5-sonnet": {Input: 60, Output: 40},
			},
		},
		{
			Day:    "2026-09-21",
			Tokens: 50,
			Turns:  1,
			Input:  30,
			Output: 20,
			Cost:   0,
			Models: map[string]ModelTokens{
				"claude-3-5-sonnet": {Input: 30, Output: 20},
			},
		},
	}

	// Agy fixture:
	// Day 1: 1 action step (activity-only row: Tokens == Turns, Input/Output/Cache == 0)
	agyDaily := []DailyRow{
		{
			Day:    "2026-09-20",
			Tokens: 1,
			Turns:  1,
			Input:  0,
			Output: 0,
			Cost:   0,
		},
	}

	results := []ToolResult{
		{
			Tool:    ToolClaude,
			Name:    "Claude",
			Daily:   claudeDaily,
			Summary: Summary{LifetimeTokens: 150, Cost: 1.50},
		},
		{
			Tool:    ToolAgy,
			Name:    "Agy",
			Daily:   agyDaily,
			Summary: Summary{LifetimeTokens: 1},
		},
	}

	projects := []ProjectDay{
		{
			Project: "repo-alpha",
			Day:     "2026-09-20",
			Tool:    "Claude",
			Tokens:  100,
			Cost:    1.50,
		},
		{
			Project: "repo-alpha",
			Day:     "2026-09-21",
			Tool:    "Claude",
			Tokens:  50,
			Cost:    0,
			Models: map[string]ModelTokens{
				"claude-3-5-sonnet": {Input: 30, Output: 20},
			},
		},
	}

	pricer := &mockTelemetryPricer{}
	now, _ := time.Parse("2006-01-02", "2026-09-22")

	// 1. Compute canonical static totals for comparison
	allDays := append([]DailyRow{}, claudeDaily...)
	allDays = append(allDays, agyDaily...)

	canonicalStats := AggregateStats(allDays, StatsOptions{Now: now, Metric: "tokens"}, pricer)
	canonicalCostStats := AggregateStats(allDays, StatsOptions{Now: now, Metric: "cost"}, pricer)
	canonicalProjects := AggregateProjects(projects, ProjectOptions{Now: now}, pricer)

	// 2. Compute telemetry snapshot
	tele := AggregateTelemetry(results, projects, TelemetryOptions{Now: now}, pricer)

	// Verify token total: must match AggregateStats exactly and exclude Agy step
	if tele.TotalTokens != int64(canonicalStats.Total) {
		t.Fatalf("TotalTokens = %d, want AggregateStats.Total = %d", tele.TotalTokens, int64(canonicalStats.Total))
	}
	if tele.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150 (Agy 1-step excluded)", tele.TotalTokens)
	}

	// Verify cost total: must match AggregateStats cost exactly ($1.50 recorded + $0.75 estimated = $2.25)
	if tele.TotalCost != canonicalCostStats.Total {
		t.Fatalf("TotalCost = %v, want canonicalCostStats.Total = %v", tele.TotalCost, canonicalCostStats.Total)
	}
	if tele.RecordedCost != 1.50 {
		t.Errorf("RecordedCost = %v, want 1.50", tele.RecordedCost)
	}
	if tele.EstimatedCost != 0.75 {
		t.Errorf("EstimatedCost = %v, want 0.75", tele.EstimatedCost)
	}

	// Verify active days and streaks: both days 2026-09-20 and 2026-09-21 are active
	if tele.ActiveDays != 2 {
		t.Errorf("ActiveDays = %d, want 2", tele.ActiveDays)
	}
	if tele.CurrentStreak != 2 {
		t.Errorf("CurrentStreak = %d, want 2", tele.CurrentStreak)
	}

	// Verify projects match AggregateProjects
	if len(tele.Projects) != len(canonicalProjects.Rows) {
		t.Fatalf("Projects length = %d, want %d", len(tele.Projects), len(canonicalProjects.Rows))
	}
	if tele.Projects[0].Project != canonicalProjects.Rows[0].Project {
		t.Errorf("Project name = %s, want %s", tele.Projects[0].Project, canonicalProjects.Rows[0].Project)
	}

	// Verify --no-estimate turns off estimated pricing
	teleNoEst := AggregateTelemetry(results, projects, TelemetryOptions{Now: now, NoEstimate: true}, pricer)
	if teleNoEst.EstimatedCost != 0 {
		t.Errorf("NoEstimate EstimatedCost = %v, want 0", teleNoEst.EstimatedCost)
	}
	if teleNoEst.TotalCost != 1.50 {
		t.Errorf("NoEstimate TotalCost = %v, want 1.50", teleNoEst.TotalCost)
	}

	// Verify tool filter
	teleClaude := AggregateTelemetry(results, projects, TelemetryOptions{Now: now, Tool: "claude"}, pricer)
	if len(teleClaude.Results) != 1 || teleClaude.Results[0].Tool != ToolClaude {
		t.Errorf("Tool filter results = %+v, expected only Claude", teleClaude.Results)
	}

	// Verify window filter
	teleWindow := AggregateTelemetry(results, projects, TelemetryOptions{Now: now, Since: "2026-09-21", Until: "2026-09-21"}, pricer)
	if teleWindow.TotalTokens != 50 {
		t.Errorf("Window TotalTokens = %d, want 50", teleWindow.TotalTokens)
	}
	if teleWindow.ActiveDays != 1 {
		t.Errorf("Window ActiveDays = %d, want 1", teleWindow.ActiveDays)
	}
}
