// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"math"
	"testing"
	"time"
)

type reconcilePricer struct{}

func (reconcilePricer) PriceDay(day DailyRow) (float64, []string) {
	var classified int64
	for _, m := range day.Models {
		classified += m.Total()
	}
	if classified == 0 {
		return 0, nil
	}
	return float64(classified) / 100.0, nil
}

func TestCrossSurfaceTotalsReconcile(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	days := []DailyRow{
		{Day: "2026-01-01", Tokens: 100, Input: 100, Turns: 1, Cost: 5, Models: map[string]ModelTokens{"m1": {Input: 100}}},
		{Day: "2026-01-01", Tokens: 200, Input: 200, Turns: 1, Models: map[string]ModelTokens{"m1": {Input: 200}}},
		{Day: "2026-01-02", Tokens: 50, Turns: 1},
	}
	projects := []ProjectDay{
		{Project: "/p1", Tool: "A", Day: "2026-01-01", Tokens: 100, Input: 100, Turns: 1, Cost: 5, Models: map[string]ModelTokens{"m1": {Input: 100}}},
		{Project: "/p1", Tool: "B", Day: "2026-01-01", Tokens: 200, Input: 200, Turns: 1, Models: map[string]ModelTokens{"m1": {Input: 200}}},
		{Project: "/p2", Tool: "B", Day: "2026-01-02", Tokens: 50, Turns: 1},
	}
	batches := []ToolDays{
		{Tool: "A", Days: days[:1]},
		{Tool: "B", Days: days[1:]},
	}
	pricer := reconcilePricer{}

	daily := Aggregate(days, GrainDay, AggregateOptions{Now: now}, pricer)
	if daily.Totals.Tokens != 350 || daily.Totals.StoredCost != 5 || daily.Totals.EstimatedCost != 2 || daily.Totals.Cost != 7 {
		t.Fatalf("daily totals = %+v, want 350 tok / 5 recorded / 2 estimated / 7 total", daily.Totals)
	}
	if daily.Totals.UnattributedTokens != 50 {
		t.Errorf("daily unclassified = %d, want 50", daily.Totals.UnattributedTokens)
	}
	for _, grain := range []Grain{GrainWeek, GrainMonth} {
		rep := Aggregate(days, grain, AggregateOptions{Now: now}, pricer)
		if rep.Totals.Tokens != daily.Totals.Tokens || math.Abs(rep.Totals.Cost-daily.Totals.Cost) > 1e-9 {
			t.Errorf("%s totals %d/%v != daily %d/%v", grain, rep.Totals.Tokens, rep.Totals.Cost, daily.Totals.Tokens, daily.Totals.Cost)
		}
		if rep.CostBasis != CostBasisTotal {
			t.Errorf("%s cost basis = %q, want total", grain, rep.CostBasis)
		}
	}

	projectOpts := ProjectOptions{
		Now:              now,
		RecordedCostDays: RecordedCostDaysForResults([]ToolResult{{Tool: ToolOpenCode, Name: "A", Daily: days[:1]}}),
	}
	projectsRep := AggregateProjects(projects, projectOpts, pricer)
	if projectsRep.Totals.Tokens != daily.Totals.Tokens || projectsRep.Totals.Cost != daily.Totals.Cost {
		t.Errorf("projects total %d/%v does not reconcile to daily %d/%v",
			projectsRep.Totals.Tokens, projectsRep.Totals.Cost, daily.Totals.Tokens, daily.Totals.Cost)
	}

	telemetry := AggregateTelemetry(
		[]ToolResult{
			{Tool: ToolOpenCode, Name: "A", Daily: days[:1]},
			{Tool: ToolCommandCode, Name: "B", Daily: days[1:]},
		},
		projects,
		TelemetryOptions{Now: now},
		pricer,
	)
	if telemetry.TotalTokens != daily.Totals.Tokens || math.Abs(telemetry.TotalCost-daily.Totals.Cost) > 1e-9 {
		t.Errorf("leaderboard telemetry %d/%v does not reconcile to daily %d/%v",
			telemetry.TotalTokens, telemetry.TotalCost, daily.Totals.Tokens, daily.Totals.Cost)
	}

	models := AggregateModels(batches, ModelOptions{Now: now}, pricer)
	if models.CostBasis != CostBasisEstimated {
		t.Errorf("models cost basis = %q, want estimated", models.CostBasis)
	}
	if models.Totals.Tokens+models.Totals.UnattributedTokens != daily.Totals.Tokens {
		t.Errorf("models classified %d + unclassified %d != daily %d",
			models.Totals.Tokens, models.Totals.UnattributedTokens, daily.Totals.Tokens)
	}
	if models.Totals.UnattributedTokens != 50 {
		t.Errorf("models unclassified = %d, want 50", models.Totals.UnattributedTokens)
	}

	tokenStats := AggregateStats(days, StatsOptions{Now: now, Metric: "tokens"}, pricer)
	if tokenStats.Days != 2 || tokenStats.Total != float64(daily.Totals.Tokens) {
		t.Errorf("token stats days/total = %d/%v, want 2/%d", tokenStats.Days, tokenStats.Total, daily.Totals.Tokens)
	}
	if tokenStats.Composition == nil || tokenStats.Composition.Unclassified != 50 {
		t.Errorf("token stats composition missing 50 unclassified: %+v", tokenStats.Composition)
	}
	costStats := AggregateStats(days, StatsOptions{Now: now, Metric: "cost"}, pricer)
	if costStats.Days != 2 || costStats.CostedDays != 1 {
		t.Errorf("cost stats active/costed days = %d/%d, want 2/1", costStats.Days, costStats.CostedDays)
	}
	if math.Abs(costStats.Total-daily.Totals.Cost) > 1e-9 {
		t.Errorf("cost stats total %v != daily %v", costStats.Total, daily.Totals.Cost)
	}

	toolMix := AggregateToolMix(batches, MixOptions{Now: now, Metric: "cost"}, pricer)
	if math.Abs(toolMix.Total-daily.Totals.Cost) > 1e-9 {
		t.Errorf("tool mix cost total %v != daily %v", toolMix.Total, daily.Totals.Cost)
	}
	if toolMix.CostBasis != CostBasisTotal {
		t.Errorf("tool mix cost basis = %q, want total", toolMix.CostBasis)
	}
	modelMix := AggregateModelMix(days, MixOptions{Now: now, Metric: "cost"}, pricer)
	if modelMix.CostBasis != CostBasisEstimated {
		t.Errorf("model mix cost basis = %q, want estimated", modelMix.CostBasis)
	}
}
