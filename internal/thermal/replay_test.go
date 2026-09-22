// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
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
