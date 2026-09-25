// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"testing"
)

func TestClassifyCommand(t *testing.T) {
	tests := []struct {
		cmd          string
		wantCategory ActivityCategory
		wantUtil     string
	}{
		{"go test ./...", ActivityTesting, "go"},
		{"cargo test", ActivityTesting, "cargo"},
		{"pytest tests/", ActivityTesting, "pytest"},
		{"golangci-lint run", ActivityTesting, "golangci-lint"},
		{"git commit -s -m 'feat: ...'", ActivityCoding, "git"},
		{"go build ./cmd/thermal", ActivityCoding, "go"},
		{"git diff main", ActivityExploration, "git"},
		{"cat README.md", ActivityExploration, "cat"},
		{"grep -r 'pattern' .", ActivityExploration, "grep"},
		{"dmesg -T", ActivityDebugging, "dmesg"},
		{"journalctl -xe", ActivityDebugging, "journalctl"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			cat, util := ClassifyCommand(tt.cmd)
			if cat != tt.wantCategory {
				t.Errorf("ClassifyCommand(%q) cat = %v; want %v", tt.cmd, cat, tt.wantCategory)
			}
			if util != tt.wantUtil {
				t.Errorf("ClassifyCommand(%q) util = %v; want %v", tt.cmd, util, tt.wantUtil)
			}
		})
	}
}

func TestComputeFinOpsGrid(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-20", Tokens: 100000, Input: 80000, Cache: 20000, Cost: 0.50, Turns: 10},
		{Day: "2026-09-21", Tokens: 200000, Input: 160000, Cache: 40000, Cost: 1.00, Turns: 20},
		{Day: "2026-09-22", Tokens: 300000, Input: 240000, Cache: 60000, Cost: 1.50, Turns: 30},
	}

	yieldRep := YieldReport{
		Totals: YieldRow{
			Tokens:       600000,
			LinesAdded:   1500,
			LinesDeleted: 500,
			NetLines:     1000,
			Efficiency:   "HIGH",
			Status:       "MEASURED",
		},
	}

	grid := ComputeFinOpsGrid(days, nil, nil, yieldRep, nil)

	if grid.TotalTokens != 600000 {
		t.Errorf("expected 600000 total tokens, got %d", grid.TotalTokens)
	}
	if grid.ActiveDays != 3 {
		t.Errorf("expected 3 active days, got %d", grid.ActiveDays)
	}
	if grid.AvgDailyTokens != 200000 {
		t.Errorf("expected 200000 avg daily tokens, got %d", grid.AvgDailyTokens)
	}
	if grid.TotalCost != 3.00 {
		t.Errorf("expected $3.00 total cost, got %.2f", grid.TotalCost)
	}
	if grid.SpendEfficiency != "HIGH" {
		t.Errorf("expected HIGH spend efficiency, got %s", grid.SpendEfficiency)
	}

	// Verify Cache hit rate: cacheRead (120000) / (input 480000 + cacheRead 120000) = 20.0%
	if grid.Cache.HitRate != 20.0 {
		t.Errorf("expected 20.0%% cache hit rate, got %.1f", grid.Cache.HitRate)
	}

	// Verify Taxonomy shares (5 disjoint token categories)
	if len(grid.Taxonomy) != 5 {
		t.Fatalf("expected 5 taxonomy categories, got %d", len(grid.Taxonomy))
	}
	if grid.Composition.CacheReadTokens != 120000 {
		t.Errorf("expected 120000 cache read tokens, got %d", grid.Composition.CacheReadTokens)
	}
}
