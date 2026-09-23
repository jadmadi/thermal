// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"strings"
	"testing"

	"charm.land/bubbletea/v2"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sampleFinOpsPayload() thermal.FinOpsGridPayload {
	return thermal.FinOpsGridPayload{
		Today: thermal.TodaySummary{
			Cost:         67.02,
			Calls:        776,
			Sessions:     12,
			CacheHitRate: 97.0,
			InputTokens:  17900,
			OutputTokens: 258200,
			CachedTokens: 86400000,
			WriteTokens:  2700000,
		},
		DailyHistory: []thermal.FinOpsDayRow{
			{Date: "2026-09-22", Cost: 67.02, Calls: 776},
			{Date: "2026-09-21", Cost: 145.30, Calls: 860},
			{Date: "2026-09-20", Cost: 525.49, Calls: 4095},
		},
		TotalDaysScanned: 122,
		ProjectBreakdown: []thermal.FinOpsProjectRow{
			{Name: "codeburn", Cost: 46.04, AvgCost: 23.02, Sessions: 2, Overhead: 18000},
			{Name: "thermal", Cost: 20.63, AvgCost: 2.58, Sessions: 8, Overhead: 18000},
		},
		Taxonomy: []thermal.ActivityShare{
			{Category: thermal.ActivityCoding, Tokens: 20340000, Cost: 28.47, Turns: 18, Percent: 42.5, OneShot: "0%"},
			{Category: thermal.ActivityConversation, Tokens: 9790000, Cost: 13.67, Turns: 208, Percent: 20.4, OneShot: "-"},
			{Category: thermal.ActivityExploration, Tokens: 7660000, Cost: 10.70, Turns: 97, Percent: 16.0, OneShot: "-"},
			{Category: thermal.ActivityDelegation, Tokens: 7570000, Cost: 10.58, Turns: 7, Percent: 15.8, OneShot: "0%"},
			{Category: thermal.ActivityFeatureDev, Tokens: 1270000, Cost: 1.78, Turns: 49, Percent: 2.7, OneShot: "-"},
			{Category: thermal.ActivityTesting, Tokens: 640000, Cost: 0.893, Turns: 1, Percent: 1.3, OneShot: "-"},
			{Category: thermal.ActivityBuildDeploy, Tokens: 340000, Cost: 0.474, Turns: 1, Percent: 0.7, OneShot: "-"},
			{Category: thermal.ActivityBrainstorming, Tokens: 160000, Cost: 0.229, Turns: 15, Percent: 0.3, OneShot: "-"},
			{Category: thermal.ActivityDebugging, Tokens: 168000, Cost: 0.223, Turns: 17, Percent: 0.3, OneShot: "-"},
		},
		ModelBreakdown: []thermal.FinOpsModelRow{
			{Name: "Fable", Cost: 41.18, CachePct: 97.6, Calls: 80, OneShot: "-", TokPerS: "-"},
			{Name: "Opus 5", Cost: 16.50, CachePct: 96.9, Calls: 176, OneShot: "-", TokPerS: "-"},
			{Name: "Sonnet", Cost: 3.96, CachePct: 97.6, Calls: 128, OneShot: "-", TokPerS: "-"},
		},
		CoreTools: []thermal.CoreToolCall{
			{Name: "cursor:read", Calls: 393},
			{Name: "Bash", Calls: 303},
			{Name: "cursor:grep", Calls: 178},
		},
		SubTools: []thermal.SubToolCall{
			{Name: "grep", Calls: 237, Share: 28.0},
			{Name: "echo", Calls: 178, Share: 21.0},
			{Name: "head", Calls: 145, Share: 17.0},
		},
		SkillsAgents: []thermal.SkillAgentCall{
			{Name: "claude-code-guide", Uses: 1, Cost: 0.187},
			{Name: "Explore", Uses: 1, Cost: 0.135},
		},
		Workflow: thermal.WorkflowMetrics{
			Corrections: "0% (1)",
			FirstEdit:   "-",
			Rework:      "-",
			Coverage:    "100%",
		},
		MCP: thermal.MCPMetrics{
			ServersActive:  0,
			ServerCalls:    0,
			OverheadTokens: 0,
		},
		TotalTokens:     45200000,
		TotalCost:       124.50,
		ActiveDays:      30,
		AvgDailyTokens:  1506666,
		AvgDailyCost:    4.15,
		SpendEfficiency: "HIGH",
		Cache: thermal.CacheMetrics{
			CacheReadTokens:  12400000,
			CacheWriteTokens: 1500000,
			InputTokens:      25000000,
			HitRate:          33.2,
			EstimatedSavings: 27.28,
		},
		TopModels: []thermal.YieldRow{
			{Name: "claude-3-5-sonnet", Tokens: 25400000},
			{Name: "gpt-4o", Tokens: 12100000},
			{Name: "deepseek-v3", Tokens: 7700000},
		},
		TopProjects: []thermal.YieldRow{
			{Name: "thermal-streak", Tokens: 30200000},
			{Name: "lib-shared", Tokens: 10500000},
			{Name: "infra-deploy", Tokens: 4500000},
		},
		Yield: thermal.YieldRow{
			Tokens:         45200000,
			LinesAdded:     1240,
			LinesDeleted:   380,
			NetLines:       860,
			TokensPerNet:   525,
			TokensPerGross: 279,
			Efficiency:     "BALANCED",
			Status:         "MEASURED",
		},
		MonthEndRunRateTokens: 45200000,
		MonthEndRunRateCost:   124.50,
		CapacityVerdict:       "PASS",
	}
}

func TestRenderDenseFinOps_Dimensions(t *testing.T) {
	payload := sampleFinOpsPayload()

	sizes := []struct {
		name  string
		width int
	}{
		{"Compact (80 cols)", 80},
		{"Medium (110 cols)", 110},
		{"Full 9-Box Grid (150 cols)", 150},
	}

	for _, sz := range sizes {
		t.Run(sz.name, func(t *testing.T) {
			out := RenderDenseFinOps(payload, sz.width, false)

			// Required 9 boxes matching codeburn reference
			requiredBoxes := []string{
				"Daily Activity",
				"By Project",
				"By Activity",
				"By Model",
				"MCP Servers",
				"Core Tools",
				"Shell Commands",
				"Skills & Agents",
				"Workflow",
			}

			for _, b := range requiredBoxes {
				if !strings.Contains(out, b) {
					t.Errorf("size %s missing box: %q", sz.name, b)
				}
			}

			// Check for ANSI stripping when colorful=false
			if strings.Contains(out, "\x1b[") {
				t.Errorf("size %s contains ANSI escape sequences with colorful=false", sz.name)
			}
		})
	}
}

func TestDenseModel_KeyFlow(t *testing.T) {
	payload := sampleFinOpsPayload()
	m := NewDense(payload, false)

	// Test window resize
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	dm := updated.(DenseModel)
	if dm.width != 150 || dm.height != 40 {
		t.Fatalf("expected width 150, height 40, got %d, %d", dm.width, dm.height)
	}

	// Test metric toggle 'm'
	if dm.metric != MetricTokens {
		t.Fatalf("expected initial metric Tokens, got %v", dm.metric)
	}
	updated, _ = dm.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
	dm = updated.(DenseModel)
	if dm.metric != MetricCost {
		t.Fatalf("expected toggled metric Cost, got %v", dm.metric)
	}

	// Test range cycle 'r'
	if dm.rng != Range30d {
		t.Fatalf("expected initial range 30d, got %v", dm.rng)
	}
	updated, _ = dm.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	dm = updated.(DenseModel)
	if dm.rng != Range90d {
		t.Fatalf("expected cycled range 90d, got %v", dm.rng)
	}

	// Test view rendering non-empty
	view := dm.View().Content
	if !strings.Contains(view, "Daily Activity") {
		t.Fatalf("unexpected view output:\n%s", view)
	}

	// Test quit 'q'
	updated, cmd := dm.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	dm = updated.(DenseModel)
	if !dm.quitting {
		t.Fatalf("expected quitting to be true after pressing q")
	}
	if cmd == nil {
		t.Fatalf("expected non-nil quit command after pressing q")
	}
}
