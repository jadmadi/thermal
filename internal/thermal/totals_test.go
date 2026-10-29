// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"math"
	"testing"
	"time"
)

func TestProjectTotalsDistinctDays(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	days := []ProjectDay{
		{Project: "alpha", Tool: "OpenCode", Day: "2026-10-30", Turns: 1, Tokens: 100, Input: 100},
		{Project: "beta", Tool: "Codex", Day: "2026-10-30", Turns: 1, Tokens: 200, Output: 200},
	}
	rep := AggregateProjects(days, ProjectOptions{Now: now}, nil)
	if rep.Totals.ActiveDays != 1 {
		t.Fatalf("activeDays=%d, want 1 distinct calendar day", rep.Totals.ActiveDays)
	}
	if rep.Totals.ProjectDays != 2 {
		t.Fatalf("projectDays=%d, want 2", rep.Totals.ProjectDays)
	}
	if rep.Totals.ToolCount != 2 {
		t.Fatalf("toolCount=%d, want 2", rep.Totals.ToolCount)
	}
	if rep.Totals.Tokens != 300 {
		t.Fatalf("totalTokens=%d, want 300", rep.Totals.Tokens)
	}
}

func TestTotalRowAdditivity(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	projectDays := []ProjectDay{
		{Project: "alpha", Tool: "OpenCode", Day: "2026-10-29", Turns: 1, Tokens: 100, Cost: 1.25},
		{Project: "alpha", Tool: "OpenCode", Day: "2026-10-30", Turns: 1, Tokens: 200, Cost: 2.50},
		{Project: "beta", Tool: "Codex", Day: "2026-10-30", Turns: 1, Tokens: 300},
	}
	projects := AggregateProjects(projectDays, ProjectOptions{Now: now}, nil)
	var tokens int64
	var cost float64
	var projectDaysSum int
	for _, row := range projects.Rows {
		tokens += row.Tokens
		cost += row.Cost
		projectDaysSum += row.ActiveDays
	}
	if projects.Totals.Tokens != tokens || math.Abs(projects.Totals.Cost-cost) > 0.0001 {
		t.Fatalf("project totals not additive: %+v vs tokens=%d cost=%v", projects.Totals, tokens, cost)
	}
	if projects.Totals.ProjectDays != projectDaysSum {
		t.Fatalf("projectDays=%d, want %d", projects.Totals.ProjectDays, projectDaysSum)
	}

	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{{
			Day: "2026-10-30", Turns: 1, Tokens: 100,
			Models: map[string]ModelTokens{"model-a": {Output: 100}},
		}}},
		{Tool: "Codex", Days: []DailyRow{{
			Day: "2026-10-30", Turns: 1, Tokens: 200,
			Models: map[string]ModelTokens{"model-b": {Output: 200}},
		}}},
	}
	models := AggregateModels(batches, ModelOptions{Now: now}, nil)
	var modelTokens int64
	var modelDays int
	for _, row := range models.Rows {
		modelTokens += row.Tokens
		modelDays += row.Days
	}
	if models.Totals.Tokens != modelTokens {
		t.Fatalf("model totals not additive: %d vs %d", models.Totals.Tokens, modelTokens)
	}
	if models.Totals.ModelDays != modelDays {
		t.Fatalf("modelDays=%d, want %d", models.Totals.ModelDays, modelDays)
	}
}

func TestMixTotals(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{
			{Day: "2026-10-29", Turns: 1, Tokens: 100, Models: map[string]ModelTokens{"m1": {Output: 100}}},
			{Day: "2026-10-30", Turns: 1, Tokens: 300, Models: map[string]ModelTokens{"m1": {Output: 300}}},
		}},
		{Tool: "Codex", Days: []DailyRow{
			{Day: "2026-10-30", Turns: 1, Tokens: 200, Models: map[string]ModelTokens{"m2": {Output: 200}}},
		}},
	}
	rep := AggregateToolMix(batches, MixOptions{Now: now, Grain: GrainDay}, nil)
	var sum float64
	for _, bucket := range rep.Buckets {
		sum += bucket.Total
	}
	if math.Abs(sum-rep.Total) > 0.0001 {
		t.Fatalf("mix total=%v, bucket sum=%v", rep.Total, sum)
	}
	var share float64
	for _, series := range rep.Series {
		share += series.Share
	}
	if math.Abs(share-1) > 0.01 {
		t.Fatalf("shares sum to %v, want 1", share)
	}
}
