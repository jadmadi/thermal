// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"math"
	"strings"
	"testing"
	"time"
)

type mockReplayPricer struct {
	prices map[string][4]float64 // model -> [in, out, cacheRead, cacheWrite]
}

func (m *mockReplayPricer) PriceDay(_ DailyRow) (float64, []string) {
	return 0, nil
}

func (m *mockReplayPricer) LookupPrice(model string) (input, output, cacheRead, cacheWrite float64, ok bool) {
	if p, exists := m.prices[model]; exists {
		return p[0], p[1], p[2], p[3], true
	}
	return 0, 0, 0, 0, false
}

func TestReplayVolumeConservation(t *testing.T) {
	t.Run("partial model coverage", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 1000, Input: 800, Output: 200,
			Models: map[string]ModelTokens{"known": {Input: 80, Output: 20}}}
		snap, _, _ := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 1000 {
			t.Fatalf("TotalTokens = %d, want 1000", snap.TotalTokens)
		}
		if snap.PeakDailyTokens != 1000 {
			t.Fatalf("PeakDailyTokens = %d, want 1000", snap.PeakDailyTokens)
		}
		if snap.UncachedInput != 800 || snap.Output != 200 || snap.UnclassifiedTokens != 0 {
			t.Fatalf("categories = (%d in, %d out, %d unclassified); want (800, 200, 0)",
				snap.UncachedInput, snap.Output, snap.UnclassifiedTokens)
		}
	})

	t.Run("complete model coverage stays priced", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 1000, Input: 800, Output: 200,
			Models: map[string]ModelTokens{"m": {Input: 800, Output: 200}}}
		pricer := &mockReplayPricer{prices: map[string][4]float64{"m": {3.0, 15.0, 0.3, 3.75}}}
		snap, _, calendarDays := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 1000 || snap.UncachedInput != 800 || snap.Output != 200 || snap.UnclassifiedTokens != 0 {
			t.Fatalf("complete coverage categories mismatch: %+v", snap)
		}
		row := SimulateModel(snap, calendarDays, "m", pricer)
		if row.PricingStatus != "complete" {
			t.Fatalf("PricingStatus = %q, want complete", row.PricingStatus)
		}
		wantMonthly := ((800*3.0 + 200*15.0) / 1e6) * 30.0
		if math.Abs(row.MonthlyCost-wantMonthly) > 1e-9 {
			t.Fatalf("MonthlyCost = %f, want %f", row.MonthlyCost, wantMonthly)
		}
	})

	t.Run("unclassified model tokens retained", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 1000,
			Models: map[string]ModelTokens{"recorded-model": {Unclassified: 1000}}}
		snap, _, _ := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 1000 {
			t.Fatalf("TotalTokens = %d, want 1000", snap.TotalTokens)
		}
		if snap.UnclassifiedTokens != 1000 {
			t.Fatalf("UnclassifiedTokens = %d, want 1000", snap.UnclassifiedTokens)
		}
		if snap.UncachedInput != 0 || snap.Output != 0 || snap.CacheRead != 0 || snap.CacheWrite != 0 {
			t.Fatalf("unclassified shape leaked into typed categories: %+v", snap)
		}
	})

	t.Run("empty maps use canonical categories", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 500, Input: 300, Output: 200}
		snap, _, _ := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 500 || snap.UncachedInput != 300 || snap.Output != 200 || snap.UnclassifiedTokens != 0 {
			t.Fatalf("canonical categories mismatch: %+v", snap)
		}
	})

	t.Run("activity-only rows excluded", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 12, Turns: 12}
		snap, byDay, _ := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 0 || len(byDay) != 0 || snap.ActiveDays != 0 {
			t.Fatalf("activity-only row counted as tokens: %+v", snap)
		}
	})

	t.Run("date exclusion", func(t *testing.T) {
		days := []DailyRow{
			{Day: "2026-10-13", Tokens: 900, Input: 900},
			{Day: "2026-10-14", Tokens: 1000, Input: 1000},
		}
		snap, _, _ := BuildWorkloadSnapshot(days, ReplayOptions{Since: "2026-10-14", Until: "2026-10-14"}, nil)
		if snap.TotalTokens != 1000 || snap.PeakDailyTokens != 1000 {
			t.Fatalf("date window not applied: %+v", snap)
		}
	})

	t.Run("overclassified canonical input", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 100, Input: 150, Output: 75}
		snap, _, _ := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 100 {
			t.Fatalf("TotalTokens = %d, want canonical 100", snap.TotalTokens)
		}
		if snap.UnclassifiedTokens != 100 || snap.UncachedInput != 0 || snap.Output != 0 {
			t.Fatalf("inconsistent day was scaled instead of unclassified: %+v", snap)
		}
	})

	t.Run("overclassified model input", func(t *testing.T) {
		day := DailyRow{Day: "2026-10-14", Tokens: 100,
			Models: map[string]ModelTokens{"m": {Input: 80, Output: 50}}}
		snap, _, _ := BuildWorkloadSnapshot([]DailyRow{day}, ReplayOptions{}, nil)
		if snap.TotalTokens != 100 {
			t.Fatalf("TotalTokens = %d, want canonical 100", snap.TotalTokens)
		}
		if snap.UnclassifiedTokens != 100 || snap.UncachedInput != 0 || snap.Output != 0 {
			t.Fatalf("inconsistent model day was scaled instead of unclassified: %+v", snap)
		}
	})
}

func TestReplayMissingPriceNotRecommended(t *testing.T) {
	window := func(against string) ReplayOptions {
		return ReplayOptions{Since: "2026-10-14", Until: "2026-10-14", Against: against}
	}
	day := DailyRow{Day: "2026-10-14", Tokens: 1000, Input: 1000, Cost: 2}

	t.Run("nil pricer never recommends", func(t *testing.T) {
		rep := AggregateReplay([]DailyRow{day}, window("not-in-catalog"), nil)
		row := rep.Plans[0]
		if row.PricingStatus != "missing" {
			t.Fatalf("PricingStatus = %q, want missing", row.PricingStatus)
		}
		if row.IsRecommended {
			t.Fatal("unpriced target recommended as free")
		}
		if row.CostDelta != 0 || row.CostDeltaPercent != 0 {
			t.Fatalf("missing price row claims delta: %+v", row)
		}
		if strings.Contains(strings.ToLower(rep.Recommendation), "saving") {
			t.Fatalf("missing price row claimed savings: %q", rep.Recommendation)
		}
	})

	t.Run("unknown model price never recommends", func(t *testing.T) {
		pricer := &mockReplayPricer{prices: map[string][4]float64{"known": {1, 1, 1, 1}}}
		rep := AggregateReplay([]DailyRow{day}, window("unknown-model"), pricer)
		row := rep.Plans[0]
		if row.PricingStatus != "missing" || row.IsRecommended || row.CostDeltaPercent != 0 {
			t.Fatalf("unknown model price produced a claim: %+v", row)
		}
	})

	t.Run("partial unclassified volume is partial", func(t *testing.T) {
		pricer := &mockReplayPricer{prices: map[string][4]float64{"known": {3.0, 15.0, 0.3, 3.75}}}
		unclassifiedDay := DailyRow{Day: "2026-10-14", Tokens: 1000,
			Models: map[string]ModelTokens{"known": {Unclassified: 1000}}}
		rep := AggregateReplay([]DailyRow{unclassifiedDay}, window("known"), pricer)
		row := rep.Plans[0]
		if row.PricingStatus != "partial" {
			t.Fatalf("PricingStatus = %q, want partial", row.PricingStatus)
		}
		if row.IsRecommended {
			t.Fatal("partial pricing row recommended")
		}
		if row.CostDelta != 0 || row.CostDeltaPercent != 0 {
			t.Fatalf("partial price row claims delta: %+v", row)
		}
	})

	t.Run("complete priced control unchanged", func(t *testing.T) {
		pricer := &mockReplayPricer{prices: map[string][4]float64{"known": {3.0, 15.0, 0.3, 3.75}}}
		typedDay := DailyRow{Day: "2026-10-14", Tokens: 1000, Input: 800, Output: 200, Cost: 2}
		rep := AggregateReplay([]DailyRow{typedDay}, window("known"), pricer)
		row := rep.Plans[0]
		if row.PricingStatus != "complete" {
			t.Fatalf("PricingStatus = %q, want complete", row.PricingStatus)
		}
		wantMonthly := ((800*3.0 + 200*15.0) / 1e6) * 30.0
		if math.Abs(row.MonthlyCost-wantMonthly) > 1e-9 {
			t.Fatalf("MonthlyCost = %f, want %f", row.MonthlyCost, wantMonthly)
		}
		wantDelta := wantMonthly - 60.0
		if math.Abs(row.CostDelta-wantDelta) > 1e-9 {
			t.Fatalf("CostDelta = %f, want %f", row.CostDelta, wantDelta)
		}
		wantPercent := (wantDelta / 60.0) * 100
		if math.Abs(row.CostDeltaPercent-wantPercent) > 1e-9 {
			t.Fatalf("CostDeltaPercent = %f, want %f", row.CostDeltaPercent, wantPercent)
		}
		if !row.IsRecommended {
			t.Fatal("complete priced control should remain recommendable")
		}
	})
}

func TestAggregateReplay_Basic(t *testing.T) {
	// 5 days of activity:
	// Day 1: 5M tokens
	// Day 2: 25M tokens (exceeds 15M pro limit)
	// Day 3: 50M tokens (exceeds 15M pro limit)
	// Day 4: 10M tokens
	// Day 5: 100M tokens (exceeds 15M pro limit and 75M max limit)
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 5_000_000, Input: 500_000, Output: 100_000, Cache: 4_400_000},
		{Day: "2026-09-02", Tokens: 25_000_000, Input: 2_000_000, Output: 200_000, Cache: 22_800_000},
		{Day: "2026-09-03", Tokens: 50_000_000, Input: 4_000_000, Output: 400_000, Cache: 45_600_000},
		{Day: "2026-09-04", Tokens: 10_000_000, Input: 1_000_000, Output: 100_000, Cache: 8_900_000},
		{Day: "2026-09-05", Tokens: 100_000_000, Input: 8_000_000, Output: 800_000, Cache: 91_200_000},
	}

	pricer := &mockReplayPricer{
		prices: map[string][4]float64{
			"claude-3-5-sonnet": {3.0, 15.0, 0.30, 3.75},
			"deepseek-chat":     {0.14, 0.28, 0.014, 0.14},
		},
	}

	plans := []SubscriptionPlan{
		{
			ID:              "claude-pro",
			Name:            "Claude Pro ($20)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      20.0,
			DailyTokenLimit: 15_000_000,
			DefaultModel:    "claude-3-5-sonnet",
		},
		{
			ID:              "claude-max",
			Name:            "Claude Max ($200)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      200.0,
			DailyTokenLimit: 75_000_000,
			DefaultModel:    "claude-3-5-sonnet",
		},
		{
			ID:              "deepseek-api",
			Name:            "DeepSeek V3 (API)",
			Type:            PlanTypePayAsYouGo,
			MonthlyFee:      0.0,
			DailyTokenLimit: 0,
			DefaultModel:    "deepseek-chat",
		},
	}

	opts := ReplayOptions{
		Since:   "2026-09-01",
		Until:   "2026-09-05",
		Compare: plans,
		Now:     time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC),
	}

	rep := AggregateReplay(days, opts, pricer)

	if rep.Workload.ActiveDays != 5 {
		t.Errorf("expected 5 active days, got %d", rep.Workload.ActiveDays)
	}
	if rep.Workload.TotalTokens != 190_000_000 {
		t.Errorf("expected 190M total tokens, got %d", rep.Workload.TotalTokens)
	}
	if rep.Workload.PeakDailyTokens != 100_000_000 {
		t.Errorf("expected 100M peak tokens, got %d", rep.Workload.PeakDailyTokens)
	}
	if rep.Workload.CacheHitRate <= 0.90 {
		t.Errorf("expected cache hit rate > 90%%, got %f", rep.Workload.CacheHitRate)
	}

	if len(rep.Plans) != 3 {
		t.Fatalf("expected 3 simulated plans, got %d", len(rep.Plans))
	}

	// Claude Pro: days 2, 3, 5 exceed 15M -> 3/5 = 60% throttled -> FAIL
	pro := rep.Plans[0]
	if pro.ID != "claude-pro" {
		t.Errorf("expected first plan claude-pro, got %s", pro.ID)
	}
	if pro.CapacityVerdict != "FAIL" {
		t.Errorf("expected Claude Pro verdict FAIL, got %s", pro.CapacityVerdict)
	}
	if pro.ThrottledDays != 3 {
		t.Errorf("expected 3 throttled days for Claude Pro, got %d", pro.ThrottledDays)
	}

	// Claude Max: day 5 exceeds 75M -> 1/5 = 20% throttled -> DEGRADED
	max := rep.Plans[1]
	if max.ID != "claude-max" {
		t.Errorf("expected second plan claude-max, got %s", max.ID)
	}
	if max.CapacityVerdict != "DEGRADED" {
		t.Errorf("expected Claude Max verdict DEGRADED, got %s", max.CapacityVerdict)
	}
	if max.ThrottledDays != 1 {
		t.Errorf("expected 1 throttled day for Claude Max, got %d", max.ThrottledDays)
	}

	// DeepSeek API: unmetered -> PASS
	ds := rep.Plans[2]
	if ds.ID != "deepseek-api" {
		t.Errorf("expected third plan deepseek-api, got %s", ds.ID)
	}
	if ds.CapacityVerdict != "PASS" {
		t.Errorf("expected DeepSeek API verdict PASS, got %s", ds.CapacityVerdict)
	}
	if !ds.IsRecommended {
		t.Errorf("expected DeepSeek API to be marked recommended as lowest passing cost")
	}
	if ds.MonthlyCost <= 0 {
		t.Errorf("expected positive monthly cost for DeepSeek API, got %f", ds.MonthlyCost)
	}
}

func TestAggregateReplay_AgainstModel(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 1_000_000, Input: 100_000, Output: 50_000, Cache: 850_000},
	}

	pricer := &mockReplayPricer{
		prices: map[string][4]float64{
			"custom-model": {2.0, 10.0, 0.20, 2.0},
		},
	}

	opts := ReplayOptions{
		Since:   "2026-09-01",
		Until:   "2026-09-01",
		Against: "custom-model",
		Now:     time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	}

	rep := AggregateReplay(days, opts, pricer)
	if len(rep.Plans) != 1 {
		t.Fatalf("expected 1 plan for --against, got %d", len(rep.Plans))
	}
	if rep.Plans[0].ID != "custom-model" {
		t.Errorf("expected custom-model, got %s", rep.Plans[0].ID)
	}
	if rep.Plans[0].CapacityVerdict != "PASS" {
		t.Errorf("expected PASS verdict for API model replay, got %s", rep.Plans[0].CapacityVerdict)
	}
}

func TestAggregateReplay_ChatGPTRollingCapacityUnknown(t *testing.T) {
	// One day of 4M tokens with 2M reasoning. Daily totals cannot prove a
	// rolling 5-hour peak, so neither plan can be throttled or cleared.
	days := []DailyRow{
		{
			Day:       "2026-09-01",
			Tokens:    4_000_000,
			Input:     1_000_000,
			Output:    1_000_000,
			Reasoning: 2_000_000,
		},
	}

	pricer := &mockReplayPricer{
		prices: map[string][4]float64{
			"gpt-4o": {2.5, 10.0, 1.25, 0},
		},
	}

	plans := []SubscriptionPlan{
		{
			ID:                 "chatgpt-plus",
			Name:               "ChatGPT Plus ($20)",
			Type:               PlanTypeSubscription,
			MonthlyFee:         20.0,
			DailyTokenLimit:    20_000_000,
			RollingWindowLimit: 5_000_000,
			BurstMultiplier:    3.0,
			DefaultModel:       "gpt-4o",
		},
		{
			ID:                 "chatgpt-pro",
			Name:               "ChatGPT Pro ($200)",
			Type:               PlanTypeSubscription,
			MonthlyFee:         200.0,
			DailyTokenLimit:    100_000_000,
			RollingWindowLimit: 40_000_000,
			BurstMultiplier:    3.0,
			DefaultModel:       "gpt-4o",
		},
	}

	opts := ReplayOptions{
		Since:   "2026-09-01",
		Until:   "2026-09-01",
		Compare: plans,
		Now:     time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	}

	rep := AggregateReplay(days, opts, pricer)
	if len(rep.Plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(rep.Plans))
	}

	for _, row := range rep.Plans {
		if row.CapacityVerdict != "UNKNOWN" {
			t.Errorf("%s: expected UNKNOWN without intraday evidence, got %s", row.ID, row.CapacityVerdict)
		}
		if row.BurstThrottles != 0 {
			t.Errorf("%s: expected 0 burst throttles for unavailable evidence, got %d", row.ID, row.BurstThrottles)
		}
		if !strings.Contains(row.VerdictDetail, "intraday evidence") {
			t.Errorf("%s: expected verdict detail to explain missing intraday evidence, got: %s", row.ID, row.VerdictDetail)
		}
		if row.IsRecommended {
			t.Errorf("%s: UNKNOWN capacity must never be recommended", row.ID)
		}
	}
}

func TestReplayDailyDataCannotProveRollingCapacity(t *testing.T) {
	pricer := &mockReplayPricer{prices: map[string][4]float64{"m": {1, 1, 1, 1}}}
	rollingPlan := SubscriptionPlan{
		ID:                 "rolling-5m",
		Name:               "Rolling 5M ($20)",
		Type:               PlanTypeSubscription,
		MonthlyFee:         20.0,
		DailyTokenLimit:    100_000_000,
		RollingWindowLimit: 5_000_000,
		BurstMultiplier:    3.0,
		DefaultModel:       "m",
	}
	simulate := func(days []DailyRow) PlanReplayRow {
		t.Helper()
		rep := AggregateReplay(days, ReplayOptions{
			Since:   "2026-09-01",
			Until:   "2026-09-01",
			Compare: []SubscriptionPlan{rollingPlan},
			Now:     time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
		}, pricer)
		return rep.Plans[0]
	}

	// The same daily total can be one concentrated burst or a workload spread
	// across the day. Daily rows cannot distinguish the two, so neither may be
	// converted into an invented burst count.
	concentrated := []DailyRow{{Day: "2026-09-01", Tokens: 4_000_000, Input: 2_000_000, Reasoning: 2_000_000}}
	spread := []DailyRow{{Day: "2026-09-01", Tokens: 4_000_000, Input: 4_000_000}}
	for name, days := range map[string][]DailyRow{"concentrated": concentrated, "spread": spread} {
		row := simulate(days)
		if row.CapacityVerdict != "UNKNOWN" {
			t.Errorf("%s: expected UNKNOWN, got %s (%s)", name, row.CapacityVerdict, row.VerdictDetail)
		}
		if row.BurstThrottles != 0 {
			t.Errorf("%s: expected no invented burst count, got %d", name, row.BurstThrottles)
		}
	}

	// A daily total below the rolling limit still cannot exclude a
	// cross-midnight rolling peak.
	below := simulate([]DailyRow{{Day: "2026-09-01", Tokens: 1_000_000, Input: 1_000_000}})
	if below.CapacityVerdict != "UNKNOWN" {
		t.Errorf("below-window daily total: expected UNKNOWN, got %s", below.CapacityVerdict)
	}

	// A daily-ceiling exceedance remains a daily failure, with rolling spikes
	// left unasserted.
	dailyPlan := rollingPlan
	dailyPlan.DailyTokenLimit = 5_000_000
	rep := AggregateReplay([]DailyRow{{Day: "2026-09-01", Tokens: 8_000_000, Input: 8_000_000}}, ReplayOptions{
		Since:   "2026-09-01",
		Until:   "2026-09-01",
		Compare: []SubscriptionPlan{dailyPlan},
		Now:     time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	}, pricer)
	row := rep.Plans[0]
	if row.CapacityVerdict != "FAIL" {
		t.Errorf("daily ceiling exceeded: expected FAIL, got %s", row.CapacityVerdict)
	}
	if row.BurstThrottles != 0 {
		t.Errorf("daily ceiling exceeded: expected no rolling burst assertion, got %d", row.BurstThrottles)
	}
	if !strings.Contains(row.VerdictDetail, "rolling spikes unasserted") {
		t.Errorf("daily ceiling exceeded: expected unasserted rolling spikes note, got: %s", row.VerdictDetail)
	}
}
