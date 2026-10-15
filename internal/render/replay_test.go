// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderReplay_NoColor(t *testing.T) {
	rep := thermal.ReplayReport{
		Type: "replay",
		Workload: thermal.WorkloadSnapshot{
			ActiveDays:        20,
			TotalDays:         30,
			TotalTokens:       500_000_000,
			CacheHitRate:      0.952,
			MedianDailyTokens: 25_000_000,
			P90DailyTokens:    45_000_000,
			PeakDailyTokens:   80_000_000,
			ActualSpend:       75.0,
		},
		Plans: []thermal.PlanReplayRow{
			{
				ID:               "claude-pro",
				Name:             "Claude Pro ($20)",
				Type:             "sub",
				MonthlyCost:      20.0,
				CostDelta:        -55.0,
				CostDeltaPercent: -73.3,
				CapacityVerdict:  "FAIL",
				VerdictDetail:    "throttled 15/20 days (75%)",
			},
			{
				ID:               "deepseek-api",
				Name:             "DeepSeek V3 (API)",
				Type:             "payg",
				MonthlyCost:      8.50,
				CostDelta:        -66.5,
				CostDeltaPercent: -88.7,
				CapacityVerdict:  "PASS",
				VerdictDetail:    "0% throttled (unmetered API)",
				IsRecommended:    true,
			},
		},
		Recommendation: "Your median daily volume (25.0M tok) exceeds flat rate limits.\nSwitching to DeepSeek V3 (API) saves $66.50/mo.",
	}

	out := RenderReplay(rep, true)

	if !strings.Contains(out, "Thermal · replay · 30-day workload simulation") {
		t.Errorf("missing banner in output: %s", out)
	}
	if !strings.Contains(out, "Real Workload") {
		t.Errorf("missing Real Workload header: %s", out)
	}
	if !strings.Contains(out, "Claude Pro ($20)") {
		t.Errorf("missing Claude Pro plan: %s", out)
	}
	if !strings.Contains(out, "★ DeepSeek V3 (API)") {
		t.Errorf("missing recommended star on DeepSeek: %s", out)
	}
	if !strings.Contains(out, "FAIL (throttled 15/20 days (75%))") {
		t.Errorf("missing FAIL verdict: %s", out)
	}
	if !strings.Contains(out, "PASS (0% throttled (unmetered API))") {
		t.Errorf("missing PASS verdict: %s", out)
	}
	if !strings.Contains(out, "Recommendation:") {
		t.Errorf("missing Recommendation block: %s", out)
	}
	// Verify no ANSI escape codes exist when noColor is true
	if strings.Contains(out, "\x1b[") {
		t.Errorf("found ANSI color escape codes when noColor=true: %s", out)
	}
}

func TestRenderReplay_Empty(t *testing.T) {
	rep := thermal.ReplayReport{
		Type:     "replay",
		Workload: thermal.WorkloadSnapshot{ActiveDays: 0},
	}
	out := RenderReplay(rep, true)
	if !strings.Contains(out, "No active coding sessions in the selected window.") {
		t.Errorf("expected empty window notice, got: %s", out)
	}
}

func TestRenderReplayIncompletePricing(t *testing.T) {
	rep := thermal.ReplayReport{
		Type: "replay",
		Workload: thermal.WorkloadSnapshot{
			ActiveDays:         5,
			TotalDays:          30,
			TotalTokens:        1000,
			UncachedInput:      100,
			UnclassifiedTokens: 900,
			CacheHitRate:       0.5,
			MedianDailyTokens:  200,
			P90DailyTokens:     200,
			PeakDailyTokens:    200,
			ActualSpend:        2.0,
		},
		Plans: []thermal.PlanReplayRow{
			{
				ID:              "partial-target",
				Name:            "Partial Target (API)",
				Type:            "payg",
				MonthlyCost:     1.23,
				PricingStatus:   "partial",
				CapacityVerdict: "PASS",
				VerdictDetail:   "0% throttled (unmetered API)",
			},
			{
				ID:              "missing-target",
				Name:            "Missing Target (API)",
				Type:            "payg",
				PricingStatus:   "missing",
				CapacityVerdict: "PASS",
				VerdictDetail:   "0% throttled (unmetered API)",
			},
		},
	}

	out := RenderReplay(rep, true)
	if !strings.Contains(out, "Unpriced") {
		t.Errorf("missing Unpriced label for missing rate: %s", out)
	}
	if !strings.Contains(out, "Partial estimate") {
		t.Errorf("missing partial estimate label: %s", out)
	}
	if !strings.Contains(out, "~$1.23") {
		t.Errorf("missing partial subtotal: %s", out)
	}
	if !strings.Contains(out, "900 tok (90.0% of total)") {
		t.Errorf("missing unclassified token coverage: %s", out)
	}
	if strings.Contains(out, "$0.00") {
		t.Errorf("incomplete rows read as zero-cost complete plans: %s", out)
	}
	if strings.Contains(out, "★") {
		t.Errorf("incomplete rows must never be recommended: %s", out)
	}
	if strings.Contains(out, "(+") || strings.Contains(out, "(-") {
		t.Errorf("incomplete rows show numeric savings deltas: %s", out)
	}
	if !strings.Contains(out, "scenario comparison") {
		t.Errorf("missing scenario comparison clarification: %s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("found ANSI color escape codes under noColor: %s", out)
	}

	// The color path keeps the same labels and does not panic.
	t.Setenv("CLICOLOR_FORCE", "1")
	outColor := RenderReplay(rep, false)
	if !strings.Contains(outColor, "Unpriced") || !strings.Contains(outColor, "Partial estimate") {
		t.Errorf("color output lost incomplete pricing labels: %s", outColor)
	}
	if !strings.Contains(outColor, "\x1b[") {
		t.Errorf("expected ANSI styling in color output")
	}
	t.Setenv("CLICOLOR_FORCE", "")

	// JSON fixture expectations: pricing status and unclassified coverage
	// round-trip, and incomplete rows carry no recommendation flag.
	raw, err := json.Marshal(rep)
	if err != nil {
		t.Fatalf("marshal replay report: %v", err)
	}
	for _, want := range []string{
		`"pricingStatus":"partial"`,
		`"pricingStatus":"missing"`,
		`"unclassifiedTokens":900`,
	} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("JSON missing %s: %s", want, raw)
		}
	}
	if strings.Contains(string(raw), `"isRecommended"`) {
		t.Errorf("incomplete rows must not recommend in JSON: %s", raw)
	}
}

func TestRenderReplay_BurstSpikes(t *testing.T) {
	rep := thermal.ReplayReport{
		Type: "replay",
		Workload: thermal.WorkloadSnapshot{
			ActiveDays:        10,
			TotalDays:         14,
			TotalTokens:       100_000_000,
			CacheHitRate:      0.80,
			MedianDailyTokens: 10_000_000,
			P90DailyTokens:    18_000_000,
			PeakDailyTokens:   25_000_000,
			ActualSpend:       30.0,
		},
		Plans: []thermal.PlanReplayRow{
			{
				ID:               "chatgpt-plus",
				Name:             "ChatGPT Plus ($20)",
				Type:             "sub",
				MonthlyCost:      20.0,
				CostDelta:        -10.0,
				CostDeltaPercent: -33.3,
				CapacityVerdict:  "DEGRADED",
				VerdictDetail:    "throttled 2/10 days (20%) · 2 burst limit spikes",
				BurstThrottles:   2,
			},
		},
	}

	out := RenderReplay(rep, true)
	if !strings.Contains(out, "DEGRADED (throttled 2/10 days (20%) · 2 burst limit spikes)") {
		t.Errorf("missing formatted burst verdict detail, got: %s", out)
	}
}
