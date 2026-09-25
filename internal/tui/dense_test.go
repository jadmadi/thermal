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
		ToolMix: []thermal.FinOpsToolRow{
			{Name: "OpenCode", Tokens: 20400000, Cost: 28.47, Share: 45.1, Streak: 12, Sessions: 4},
			{Name: "Devin", Tokens: 14500000, Cost: 18.20, Share: 32.1, Streak: 8, Sessions: 3},
			{Name: "Claude", Tokens: 7000000, Cost: 12.50, Share: 15.5, Streak: 3, Sessions: 2},
			{Name: "Grok", Tokens: 3300000, Cost: 4.80, Share: 7.3, Streak: 1, Sessions: 1},
		},
		ModelBreakdown: []thermal.FinOpsModelRow{
			{Name: "Fable", Cost: 41.18, CachePct: 97.6, Calls: 80, OneShot: "-", TokPerS: "-"},
			{Name: "Opus 5", Cost: 16.50, CachePct: 96.9, Calls: 176, OneShot: "-", TokPerS: "-"},
			{Name: "Sonnet", Cost: 3.96, CachePct: 97.6, Calls: 128, OneShot: "-", TokPerS: "-"},
		},
		Receipt: thermal.FinOpsReceiptSummary{
			VerifiedCount:   18,
			ClaimedCount:    4,
			FailedCount:     0,
			TokensVerified:  38000000,
			VerifiedRate:    81.8,
			SpendEfficiency: "EXCELLENT",
			TestTools:       "go test, golangci-lint",
		},
		Replay: thermal.FinOpsReplaySummary{
			CapacityVerdict: "PASS",
			Plans: []thermal.FinOpsReplayPlan{
				{Name: "Claude Pro ($20)", MonthlyCost: 20.0, CostDelta: -104.50, CapacityVerdict: "PASS", ThrottleRate: 0.0},
				{Name: "Claude Max ($200)", MonthlyCost: 200.0, CostDelta: 75.50, CapacityVerdict: "PASS", ThrottleRate: 0.0},
				{Name: "ChatGPT Plus ($20)", MonthlyCost: 20.0, CostDelta: -104.50, CapacityVerdict: "PASS", ThrottleRate: 0.0},
				{Name: "Cursor Pro ($20)", MonthlyCost: 20.0, CostDelta: -104.50, CapacityVerdict: "PASS", ThrottleRate: 0.0},
				{Name: "DeepSeek V3 (API)", MonthlyCost: 18.40, CostDelta: -106.10, CapacityVerdict: "PASS", ThrottleRate: 0.0},
			},
		},
		Composition: thermal.FinOpsCompositionSummary{
			InputTokens:      25000000,
			OutputTokens:     6300000,
			ReasoningTokens:  1500000,
			CacheReadTokens:  12400000,
			CacheWriteTokens: 1500000,
			ReasoningRatio:   19.2,
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

			// Required 9 authentic boxes
			requiredBoxes := []string{
				"Daily Activity",
				"By Project",
				"Assistant Mix",
				"By Model",
				"Prompt Cache & Savings",
				"Code Yield & Velocity",
				"Work Receipts & Tests",
				"Subscription Replay",
				"Token Composition",
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

func TestDenseModel_PeriodNavigation(t *testing.T) {
	payload := sampleFinOpsPayload()
	m := NewDense(payload, false)

	// Direct numbers 1-6
	for i, key := range []string{"1", "2", "3", "4", "5", "6"} {
		updated, _ := m.Update(tea.KeyPressMsg{Code: rune(key[0]), Text: key})
		m = updated.(DenseModel)
		if m.periodIdx != i {
			t.Fatalf("pressing %s: periodIdx = %d, want %d", key, m.periodIdx, i)
		}
	}

	// Mnemonic keys: 't' (Today) and 'w' (7 Days)
	updated, _ := m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	m = updated.(DenseModel)
	if m.periodIdx != 0 {
		t.Fatalf("pressing t: periodIdx = %d, want 0", m.periodIdx)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'w', Text: "w"})
	m = updated.(DenseModel)
	if m.periodIdx != 1 {
		t.Fatalf("pressing w: periodIdx = %d, want 1", m.periodIdx)
	}

	// Forward cycling: tab, right, l, ]
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyTab},
		{Code: tea.KeyRight},
		{Code: 'l', Text: "l"},
		{Code: ']', Text: "]"},
	} {
		prev := m.periodIdx
		updated, _ := m.Update(key)
		m = updated.(DenseModel)
		want := (prev + 1) % len(m.periods)
		if m.periodIdx != want {
			t.Fatalf("after %v: periodIdx = %d, want %d", key, m.periodIdx, want)
		}
	}

	// Backward cycling: shift+tab, left, h, [
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyTab, Mod: tea.ModShift},
		{Code: tea.KeyLeft},
		{Code: 'h', Text: "h"},
		{Code: '[', Text: "["},
	} {
		prev := m.periodIdx
		updated, _ := m.Update(key)
		m = updated.(DenseModel)
		want := (prev - 1 + len(m.periods)) % len(m.periods)
		if m.periodIdx != want {
			t.Fatalf("after %v: periodIdx = %d, want %d", key, m.periodIdx, want)
		}
	}
}

func TestDenseModel_ProjectNavigation(t *testing.T) {
	payload := sampleFinOpsPayload()
	projects := []thermal.ProjectDay{
		{Project: "alpha"},
		{Project: "beta"},
		{Project: "gamma"},
	}
	m := NewDenseWithData(payload, nil, nil, projects, thermal.YieldReport{}, nil, false)

	if m.activeProject != "" {
		t.Fatalf("initial active project want empty, got %q", m.activeProject)
	}

	// Cycle forward with 'p'
	expected := []string{"alpha", "beta", "gamma", ""}
	for _, exp := range expected {
		updated, _ := m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
		m = updated.(DenseModel)
		if m.activeProject != exp {
			t.Fatalf("pressing p: activeProject = %q, want %q", m.activeProject, exp)
		}
	}

	// Cycle backward with 'P'
	expectedRev := []string{"gamma", "beta", "alpha", ""}
	for _, exp := range expectedRev {
		updated, _ := m.Update(tea.KeyPressMsg{Code: 'P', Text: "P"})
		m = updated.(DenseModel)
		if m.activeProject != exp {
			t.Fatalf("pressing P: activeProject = %q, want %q", m.activeProject, exp)
		}
	}

	// Reset with 'x'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	m = updated.(DenseModel)
	if m.activeProject != "alpha" {
		t.Fatalf("pre-reset activeProject = %q, want 'alpha'", m.activeProject)
	}
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m = updated.(DenseModel)
	if m.activeProject != "" {
		t.Fatalf("after x: activeProject = %q, want empty", m.activeProject)
	}

	// Reset with '0'
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	m = updated.(DenseModel)
	updated, _ = m.Update(tea.KeyPressMsg{Code: '0', Text: "0"})
	m = updated.(DenseModel)
	if m.activeProject != "" {
		t.Fatalf("after 0: activeProject = %q, want empty", m.activeProject)
	}

	// Clear with 'esc'
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	m = updated.(DenseModel)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = updated.(DenseModel)
	if m.activeProject != "" {
		t.Fatalf("after esc: activeProject = %q, want empty", m.activeProject)
	}
	if m.quitting {
		t.Fatalf("esc with active project should clear filter, not quit")
	}
}

func TestDenseModel_Scrolling(t *testing.T) {
	payload := sampleFinOpsPayload()
	m := NewDense(payload, false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m = updated.(DenseModel)

	// Down arrow and 'j'
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(DenseModel)
	if m.scrollOffset != 2 {
		t.Fatalf("after down arrow: scrollOffset = %d, want 2", m.scrollOffset)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = updated.(DenseModel)
	if m.scrollOffset != 4 {
		t.Fatalf("after j: scrollOffset = %d, want 4", m.scrollOffset)
	}

	// Up arrow and 'k'
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = updated.(DenseModel)
	if m.scrollOffset != 2 {
		t.Fatalf("after up arrow: scrollOffset = %d, want 2", m.scrollOffset)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	m = updated.(DenseModel)
	if m.scrollOffset != 0 {
		t.Fatalf("after k: scrollOffset = %d, want 0", m.scrollOffset)
	}

	// Section jump 'J' and 'K'
	updated, _ = m.Update(tea.KeyPressMsg{Code: 'J', Text: "J"})
	m = updated.(DenseModel)
	if m.scrollOffset != 12 {
		t.Fatalf("after J: scrollOffset = %d, want 12", m.scrollOffset)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'K', Text: "K"})
	m = updated.(DenseModel)
	if m.scrollOffset != 0 {
		t.Fatalf("after K: scrollOffset = %d, want 0", m.scrollOffset)
	}

	// Page down and Page up
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	m = updated.(DenseModel)
	if m.scrollOffset != 15 { // height / 2 = 15
		t.Fatalf("after PgDown: scrollOffset = %d, want 15", m.scrollOffset)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	m = updated.(DenseModel)
	if m.scrollOffset != 0 {
		t.Fatalf("after PgUp: scrollOffset = %d, want 0", m.scrollOffset)
	}

	// Home / End
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	m = updated.(DenseModel)
	if m.scrollOffset < 50 {
		t.Fatalf("after End: scrollOffset = %d, want >= 50", m.scrollOffset)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	m = updated.(DenseModel)
	if m.scrollOffset != 0 {
		t.Fatalf("after Home: scrollOffset = %d, want 0", m.scrollOffset)
	}
}

func TestDenseModel_Help(t *testing.T) {
	payload := sampleFinOpsPayload()
	m := NewDense(payload, false)

	// Press '?' to open help
	updated, _ := m.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	m = updated.(DenseModel)
	if !m.showHelp {
		t.Fatalf("expected showHelp=true after pressing ?")
	}

	view := m.View().Content
	if !strings.Contains(view, "Thermal FinOps 9-Box Grid — Navigation Keys") {
		t.Fatalf("expected help title in view output:\n%s", view)
	}
	if !strings.Contains(view, "Period Selection") {
		t.Fatalf("expected 'Period Selection' section in help output:\n%s", view)
	}

	// Press '?' again to close help
	updated, _ = m.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	m = updated.(DenseModel)
	if m.showHelp {
		t.Fatalf("expected showHelp=false after closing with ?")
	}

	// Press '?' then 'esc' to close
	updated, _ = m.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	m = updated.(DenseModel)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = updated.(DenseModel)
	if m.showHelp {
		t.Fatalf("expected showHelp=false after closing with esc")
	}
	if m.quitting {
		t.Fatalf("esc inside help should close help, not quit")
	}
}

func TestDenseModel_Reset(t *testing.T) {
	payload := sampleFinOpsPayload()
	projects := []thermal.ProjectDay{
		{Project: "alpha"},
	}
	m := NewDenseWithData(payload, nil, nil, projects, thermal.YieldReport{}, nil, false)

	// Mutate state: period 3, project alpha, scrollOffset 10
	m.periodIdx = 3
	m.activeProject = "alpha"
	m.scrollOffset = 10

	// Press 'R' to reset
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
	m = updated.(DenseModel)

	if m.periodIdx != 0 {
		t.Fatalf("after R: periodIdx = %d, want 0", m.periodIdx)
	}
	if m.activeProject != "" {
		t.Fatalf("after R: activeProject = %q, want empty", m.activeProject)
	}
	if m.scrollOffset != 0 {
		t.Fatalf("after R: scrollOffset = %d, want 0", m.scrollOffset)
	}
}
