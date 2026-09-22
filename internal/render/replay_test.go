// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
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
