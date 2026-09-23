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
		TotalTokens:     45200000,
		TotalCost:       124.50,
		ActiveDays:      30,
		AvgDailyTokens:  1506666,
		AvgDailyCost:    4.15,
		SpendEfficiency: "HIGH",
		Taxonomy: []thermal.ActivityShare{
			{Category: thermal.ActivityCoding, Tokens: 20340000, Cost: 56.02, Percent: 45.0},
			{Category: thermal.ActivityDebugging, Tokens: 9040000, Cost: 24.90, Percent: 20.0},
			{Category: thermal.ActivityTesting, Tokens: 9040000, Cost: 24.90, Percent: 20.0},
			{Category: thermal.ActivityExploration, Tokens: 6780000, Cost: 18.68, Percent: 15.0},
		},
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
		SubTools: []thermal.SubToolCall{
			{Name: "git", Calls: 142, Share: 32.5},
			{Name: "go", Calls: 118, Share: 27.0},
			{Name: "grep / rg", Calls: 76, Share: 17.4},
			{Name: "bash", Calls: 54, Share: 12.4},
		},
		MCP: thermal.MCPMetrics{
			ServersActive:  3,
			ServerCalls:    28,
			OverheadTokens: 2034000,
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

			// Required 9 boxes present in output
			requiredBoxes := []string{
				"Executive KPIs",
				"Activity Taxonomy",
				"Cache & FinOps Savings",
				"Top Models Spend",
				"Sub-Tool Shell Calls",
				"MCP Protocol Overhead",
				"Project Allocation",
				"Velocity & Code Yield",
				"Month-End Forecast",
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
	if !strings.Contains(view, "FinOps 9-Box Grid") {
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
