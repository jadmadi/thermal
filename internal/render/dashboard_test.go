// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"github.com/jadmadi/thermal/internal/thermal"
	"strings"
	"testing"
)

func TestRenderDashboard_Output(t *testing.T) {
	sum := thermal.Summary{
		Sessions:        5,
		LifetimeTokens:  250000,
		InputTokens:     150000,
		OutputTokens:    100000,
		ReasoningTokens: 10000,
		CacheTokens:     5000,
		Cost:            1.45,
		ModelBreakdown:  map[string]int64{"gpt-4o": 5},
	}
	daily := []thermal.DailyRow{
		{Day: "2026-07-15", Tokens: 1000, Turns: 4},
	}
	out := RenderDashboard("MiMoCode", sum, daily, "/path/to/db", 52, true)
	if !strings.Contains(out, "MiMoCode") {
		t.Errorf("expected MiMoCode activity in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1.0K tokens") {
		t.Errorf("expected 1.0K tokens metric, got:\n%s", out)
	}
	if !strings.Contains(out, "$1.45 spent") {
		t.Errorf("expected $1.45 spent, got:\n%s", out)
	}
}

func TestRenderDashboard_CommandCodeUsesTokenLabel(t *testing.T) {
	sum := thermal.Summary{
		Sessions:       1,
		LifetimeTokens: 2500,
		InputTokens:    2000,
		OutputTokens:   500,
	}
	daily := []thermal.DailyRow{
		{Day: "2026-07-15", Tokens: 2500, Input: 2000, Output: 500, Turns: 2},
	}
	out := RenderDashboard("CMD", sum, daily, "/path/to/projects", 52, true)
	if !strings.Contains(out, "2.5K tokens") {
		t.Errorf("expected CMD to report tokens, got:\n%s", out)
	}
	if strings.Contains(out, "2.5K messages") {
		t.Errorf("CMD must not be labelled as message activity, got:\n%s", out)
	}
}

func TestRenderDashboard_EstimatedCost(t *testing.T) {
	sum := thermal.Summary{
		Sessions:       2,
		LifetimeTokens: 100000,
		Cost:           0, // No recorded cost
	}
	daily := []thermal.DailyRow{
		{Day: "2026-07-15", Tokens: 1000, Turns: 2},
	}
	out := RenderDashboard("DeepSeek (DSH)", sum, daily, "/path/to/db", 52, true, 0.45)
	if !strings.Contains(out, "~$0.45 spent (est)") {
		t.Errorf("expected ~$0.45 spent (est) in output, got:\n%s", out)
	}
}

func TestRenderDashboard_MixedCostAndDisclosures(t *testing.T) {
	sum := thermal.Summary{
		Sessions:       5,
		LifetimeTokens: 200000,
		Cost:           10.50,
	}
	daily := []thermal.DailyRow{
		{Day: "2026-07-15", Tokens: 2000, Turns: 2},
	}
	out := RenderDashboard("OpenCode", sum, daily, "/path/to/db", 52, true, 2.25, []string{"mystery-model"}, int64(500))
	if !strings.Contains(out, "~$12.75 spent ($10.50 recorded + ~$2.25 est)") {
		t.Errorf("expected combined spent in output, got:\n%s", out)
	}
	if !strings.Contains(out, "500 tokens have no model recorded") {
		t.Errorf("expected unattributed tokens note in output, got:\n%s", out)
	}
	if !strings.Contains(out, "No pricing for: mystery-model") {
		t.Errorf("expected missing pricing note in output, got:\n%s", out)
	}
}
