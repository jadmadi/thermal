// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"testing"
	"time"
)

func TestParseDiffStats(t *testing.T) {
	patch := `--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main
 
-func oldFunction() {
+func newFunction() {
+	// added comment
 }
-func deprecated() {}
`
	added, deleted := ParseDiffStats(patch)
	if added != 2 {
		t.Errorf("expected 2 added lines, got %d", added)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted lines, got %d", deleted)
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int64
	}{
		{"empty", "", 0},
		{"single line no nl", "hello", 1},
		{"single line with nl", "hello\n", 1},
		{"three lines", "one\ntwo\nthree\n", 3},
		{"three lines no final nl", "one\ntwo\nthree", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CountLines(tc.content)
			if got != tc.expected {
				t.Errorf("expected %d lines, got %d", tc.expected, got)
			}
		})
	}
}

func TestCalculateYield_DefensiveGuards(t *testing.T) {
	// 1. Measured: positive net lines
	net, gross, perNet, perGross, eff, status := CalculateYield(1000, 10, 5, 2)
	if net != 5 || gross != 15 {
		t.Errorf("net=%d, gross=%d; expected 5 and 15", net, gross)
	}
	if perNet != 200.0 { // 1000 / 5 = 200
		t.Errorf("perNet=%.2f, expected 200.0", perNet)
	}
	if perGross != 1000.0/15.0 {
		t.Errorf("perGross=%.2f, expected %.2f", perGross, 1000.0/15.0)
	}
	if status != "MEASURED" || eff != "HIGH" {
		t.Errorf("expected MEASURED/HIGH, got %s/%s", status, eff)
	}

	// 2. Defensive guard: zero net lines (exploratory / reading)
	net0, _, perNet0, _, eff0, status0 := CalculateYield(5000, 0, 0, 0)
	if net0 != 0 || perNet0 != 0 {
		t.Errorf("expected 0 net/perNet on zero delta, got %d/%.2f", net0, perNet0)
	}
	if status0 != "UNMEASURED" || eff0 != "EXPLORATORY" {
		t.Errorf("expected UNMEASURED/EXPLORATORY, got %s/%s", status0, eff0)
	}

	// 3. Defensive guard: negative net lines (refactor deleting code)
	netNeg, _, perNetNeg, _, effNeg, statusNeg := CalculateYield(3000, 2, 20, 1)
	if netNeg != -18 || perNetNeg != 0 {
		t.Errorf("expected -18 net and 0 perNet on negative delta, got %d/%.2f", netNeg, perNetNeg)
	}
	if statusNeg != "UNMEASURED" || effNeg != "EXPLORATORY" {
		t.Errorf("expected UNMEASURED/EXPLORATORY, got %s/%s", statusNeg, effNeg)
	}
}

func TestAggregateYield(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	results := []ToolResult{
		{
			Tool: ToolOpenCode,
			Name: "OpenCode",
			Daily: []DailyRow{
				{
					Day:          "2026-09-22",
					Tokens:       10000,
					LinesAdded:   100,
					LinesDeleted: 20,
					FilesTouched: 3,
					Models: map[string]ModelTokens{
						"claude-3-7-sonnet": {Input: 7000, Output: 3000},
					},
					ModelLines: map[string]LineDelta{
						"claude-3-7-sonnet": {Added: 100, Deleted: 20, Files: 3},
					},
				},
				{
					Day:          "2026-09-23",
					Tokens:       20000,
					LinesAdded:   50,
					LinesDeleted: 10,
					FilesTouched: 2,
					Models: map[string]ModelTokens{
						"gpt-5-mini": {Input: 15000, Output: 5000},
					},
					ModelLines: map[string]LineDelta{
						"gpt-5-mini": {Added: 50, Deleted: 10, Files: 2},
					},
				},
			},
		},
	}

	report := AggregateYield(results, nil, YieldOptions{Now: now, Sort: "yield"})

	if len(report.Tools) != 1 {
		t.Fatalf("expected 1 tool row, got %d", len(report.Tools))
	}
	tool := report.Tools[0]
	if tool.Tokens != 30000 || tool.LinesAdded != 150 || tool.LinesDeleted != 30 || tool.NetLines != 120 {
		t.Errorf("unexpected tool metrics: %+v", tool)
	}
	// 30000 / 120 = 250 tok/line
	if tool.TokensPerNet != 250.0 || tool.Efficiency != "HIGH" {
		t.Errorf("expected 250.0 / HIGH, got %.2f / %s", tool.TokensPerNet, tool.Efficiency)
	}

	if len(report.Models) != 2 {
		t.Fatalf("expected 2 model rows, got %d", len(report.Models))
	}
	// Sorted by yield asc (lowest tok/line first):
	// claude: 10000 / 80 = 125 tok/line (HIGH)
	// gpt: 20000 / 40 = 500 tok/line (BALANCED)
	if report.Models[0].Name != "claude-3-7-sonnet" {
		t.Errorf("expected claude-3-7-sonnet first, got %s", report.Models[0].Name)
	}
	if report.Models[0].TokensPerNet != 125.0 {
		t.Errorf("expected 125.0 tok/line for claude, got %.2f", report.Models[0].TokensPerNet)
	}
	if report.Models[1].Name != "gpt-5-mini" {
		t.Errorf("expected gpt-5-mini second, got %s", report.Models[1].Name)
	}
	if report.Models[1].TokensPerNet != 500.0 || report.Models[1].Efficiency != "BALANCED" {
		t.Errorf("expected 500.0 / BALANCED for gpt, got %.2f / %s", report.Models[1].TokensPerNet, report.Models[1].Efficiency)
	}

	// Totals
	if report.Totals.Tokens != 30000 || report.Totals.NetLines != 120 || report.Totals.TokensPerNet != 250.0 {
		t.Errorf("unexpected totals: %+v", report.Totals)
	}
}
