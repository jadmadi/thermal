// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"testing"
	"time"
)

func TestAggregateToolMixBucketsSharesAndSwitching(t *testing.T) {
	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{
			{Day: "2026-09-07", Tokens: 100, Turns: 1},
			{Day: "2026-09-09", Tokens: 200, Turns: 2},
		}},
		{Tool: "Codex", Days: []DailyRow{
			{Day: "2026-09-08", Tokens: 300, Turns: 1},
		}},
	}
	rep := AggregateToolMix(batches, MixOptions{Grain: GrainWeek, StartOfWeek: time.Sunday, Metric: "tokens"}, nil)

	if rep.By != "tool" || rep.Metric != "tokens" {
		t.Errorf("report labels = %s/%s", rep.By, rep.Metric)
	}
	if len(rep.Buckets) != 1 || rep.Buckets[0].Period != "2026-09-06" {
		t.Fatalf("expected one week bucket, got %+v", rep.Buckets)
	}
	if rep.Buckets[0].Total != 600 {
		t.Errorf("bucket total = %v, want 600", rep.Buckets[0].Total)
	}
	if len(rep.Series) != 2 {
		t.Fatalf("expected 2 series, got %+v", rep.Series)
	}
	// Equal totals, so names break the tie: Codex before OpenCode.
	if rep.Series[0].Name != "Codex" || rep.Series[1].Name != "OpenCode" {
		t.Errorf("series order = %s, %s", rep.Series[0].Name, rep.Series[1].Name)
	}
	if rep.Series[0].Total != 300 || rep.Series[1].Total != 300 {
		t.Errorf("series totals = %v, %v", rep.Series[0].Total, rep.Series[1].Total)
	}
	if rep.Series[0].Share != 0.5 {
		t.Errorf("share = %v, want 0.5", rep.Series[0].Share)
	}
	if rep.Concentration != 0.5 {
		t.Errorf("concentration = %v, want 0.5", rep.Concentration)
	}
	if rep.Switches != 2 || rep.SwitchesByMonth["2026-09"] != 2 {
		t.Errorf("switches = %d %v, want 2 in 2026-09", rep.Switches, rep.SwitchesByMonth)
	}
	if rep.Dominant != "Codex" {
		t.Errorf("dominant = %q, want Codex on the name tie-break", rep.Dominant)
	}
	if rep.ActiveSeries != 2 {
		t.Errorf("active series = %d", rep.ActiveSeries)
	}
}

func TestAggregateModelMixCostUsesPricer(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-10", Tokens: 3_000_000, Models: map[string]ModelTokens{
			"cheap":  {Input: 1_000_000},
			"pricey": {Input: 2_000_000},
		}},
	}
	pricer := ratePricer{rates: map[string]float64{"cheap": 1.0, "pricey": 2.0}}
	rep := AggregateModelMix(days, MixOptions{Grain: GrainMonth, By: "model", Metric: "cost"}, pricer)

	if rep.By != "model" || rep.Metric != "cost" {
		t.Errorf("report labels = %s/%s", rep.By, rep.Metric)
	}
	if !rep.Estimated {
		t.Error("model cost must be flagged as estimated")
	}
	if len(rep.Series) != 2 || rep.Series[0].Name != "pricey" {
		t.Fatalf("series = %+v, want pricey first", rep.Series)
	}
	if rep.Series[0].Total != 4.0 || rep.Series[1].Total != 1.0 {
		t.Errorf("costs = %v, %v want 4 and 1", rep.Series[0].Total, rep.Series[1].Total)
	}
	if rep.Total != 5.0 {
		t.Errorf("total = %v, want 5", rep.Total)
	}
	if rep.Switches != 0 || rep.Dominant != "" {
		t.Errorf("model mix must not report tool switching: %d %q", rep.Switches, rep.Dominant)
	}
}

func TestAggregateToolMixCostPrefersStored(t *testing.T) {
	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{
			{Day: "2026-09-11", Tokens: 1000, Cost: 3.0},
			{Day: "2026-09-12", Tokens: 1_000_000, Models: map[string]ModelTokens{"m": {Input: 1_000_000}}},
		}},
	}
	pricer := ratePricer{rates: map[string]float64{"m": 2.0}}
	rep := AggregateToolMix(batches, MixOptions{Grain: GrainDay, Metric: "cost"}, pricer)

	if rep.Total != 5.0 {
		t.Errorf("total = %v, want 3 stored plus 2 estimated", rep.Total)
	}
	if !rep.Estimated {
		t.Error("a day without stored cost must flag the estimate")
	}
}

func TestAggregateMixEmptyWindow(t *testing.T) {
	rep := AggregateModelMix([]DailyRow{{Day: "2026-01-01", Tokens: 5, Models: map[string]ModelTokens{"m": {Input: 5}}}},
		MixOptions{Since: "2026-09-01", Grain: GrainWeek, By: "model"}, nil)
	if len(rep.Series) != 0 || len(rep.Buckets) != 0 || rep.Total != 0 {
		t.Errorf("expected an empty report, got %+v", rep)
	}
}

func TestAggregateToolMixExcludesActivityOnly(t *testing.T) {
	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{
			{Day: "2026-09-07", Tokens: 1000, Turns: 1, Input: 800, Output: 200},
		}},
		{Tool: "Agy", Days: []DailyRow{
			// Activity-only: Tokens == Turns, no input/output/cache/reasoning/cost/models
			{Day: "2026-09-07", Tokens: 417, Turns: 417},
		}},
	}
	rep := AggregateToolMix(batches, MixOptions{Grain: GrainDay, Metric: "tokens"}, nil)
	if len(rep.Series) != 1 || rep.Series[0].Name != "OpenCode" {
		t.Fatalf("expected only OpenCode series, got %+v", rep.Series)
	}
	if rep.Total != 1000 {
		t.Errorf("expected total 1000, got %v", rep.Total)
	}
}
