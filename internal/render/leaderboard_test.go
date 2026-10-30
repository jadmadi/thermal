// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderLeaderboard_Output(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: 10, LongestStreak: 15,
			ActiveDays: 30, TotalActivity: 1500000, Summary: thermal.Summary{Cost: 12.50},
		},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(out, "Token Warriors") {
		t.Errorf("expected Token Warriors section in output")
	}
	if !strings.Contains(out, "OpenCode") {
		t.Errorf("expected OpenCode in output")
	}
}

func TestRenderLeaderboard_CommandCodeIsTokenWarrior(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolCommandCode, Name: "CMD", CurrentStreak: 1, LongestStreak: 2,
			ActiveDays: 3, TotalActivity: 2500, Summary: thermal.Summary{Cost: 0.5},
		},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(out, "Token Warriors") {
		t.Errorf("expected CMD under Token Warriors, got:\n%s", out)
	}
	if strings.Contains(out, "Activity Hunters") {
		t.Errorf("expected no Activity Hunters section for CMD alone, got:\n%s", out)
	}
}

func TestLeaderboardTopCapsEachCategory(t *testing.T) {
	results := []thermal.ToolResult{
		{Tool: thermal.ToolOpenCode, Name: "oc-1", CurrentStreak: 5, LongestStreak: 5, ActiveDays: 5, TotalActivity: 5000},
		{Tool: thermal.ToolCodex, Name: "codex-1", CurrentStreak: 4, LongestStreak: 4, ActiveDays: 4, TotalActivity: 4000},
		{Tool: thermal.ToolClaude, Name: "claude-1", CurrentStreak: 3, LongestStreak: 3, ActiveDays: 3, TotalActivity: 3000},
		{Tool: thermal.ToolAgy, Name: "agy-1", CurrentStreak: 5, LongestStreak: 5, ActiveDays: 5, TotalActivity: 500},
		{Tool: thermal.ToolDroid, Name: "droid-1", CurrentStreak: 4, LongestStreak: 4, ActiveDays: 4, TotalActivity: 400},
		{Tool: thermal.ToolMuse, Name: "muse-1", CurrentStreak: 3, LongestStreak: 3, ActiveDays: 3, TotalActivity: 300},
	}
	out := RenderLeaderboard(results, 52, true, "", true, 2)
	for _, want := range []string{"oc-1", "codex-1", "agy-1", "droid-1"} {
		if !strings.Contains(out, want) {
			t.Errorf("top-2 leaderboard missing %q:\n%s", want, out)
		}
	}
	for _, notWant := range []string{"claude-1", "muse-1"} {
		if strings.Contains(out, notWant) {
			t.Errorf("top-2 leaderboard should omit %q:\n%s", notWant, out)
		}
	}
}

func TestRenderLeaderboardSortKeys(t *testing.T) {
	// steady leads on streak, whale leads on tokens, pricey leads on cost.
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "steady", CurrentStreak: 30, LongestStreak: 40,
			TotalActivity: 1000, Summary: thermal.Summary{Cost: 1},
		},
		{
			Tool: thermal.ToolCodex, Name: "whale", CurrentStreak: 2, LongestStreak: 5,
			TotalActivity: 900000, Summary: thermal.Summary{Cost: 1},
		},
		{
			Tool: thermal.ToolDevin, Name: "pricey", CurrentStreak: 1, LongestStreak: 2,
			TotalActivity: 100, Summary: thermal.Summary{Cost: 500},
		},
	}

	firstPlace := func(out string) string {
		// The first ranked row carries the gold styling only when colors are
		// on, so read the plain order of names instead.
		idx := len(out)
		name := ""
		for _, candidate := range []string{"steady", "whale", "pricey"} {
			if i := strings.Index(out, candidate); i >= 0 && i < idx {
				idx, name = i, candidate
			}
		}
		return name
	}

	if got := firstPlace(RenderLeaderboard(results, 52, true, "", true)); got != "steady" {
		t.Errorf("default order leader = %s, want steady", got)
	}
	if got := firstPlace(RenderLeaderboard(results, 52, true, "tokens", true)); got != "whale" {
		t.Errorf("token order leader = %s, want whale", got)
	}
	if got := firstPlace(RenderLeaderboard(results, 52, true, "cost", true)); got != "pricey" {
		t.Errorf("cost order leader = %s, want pricey", got)
	}
}

// TestLeaderboardStatesCostProvenance pins the line that stops the leaderboard
// and the reports from looking like they disagree about money, which they do
// by design: the leaderboard prints recorded cost, reports add estimates.
func TestLeaderboardStatesCostProvenance(t *testing.T) {
	results := []thermal.ToolResult{
		{Tool: "one", Name: "one", Summary: thermal.Summary{Cost: 10.5}, ActiveDays: 1},
		{Tool: "two", Name: "two", Summary: thermal.Summary{Cost: 2.25}, ActiveDays: 1},
	}

	withEstimate := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(withEstimate, "$12.75") {
		t.Errorf("recorded total missing:\n%s", withEstimate)
	}
	if !strings.Contains(withEstimate, "recorded") || !strings.Contains(withEstimate, "shown") {
		t.Errorf("cost arithmetic disclosure missing:\n%s", withEstimate)
	}

	plain := RenderLeaderboard(results, 52, true, "", false)
	if !strings.Contains(plain, "$12.75") {
		t.Errorf("recorded total missing with estimation off:\n%s", plain)
	}
	if !strings.Contains(plain, "estimation disabled") {
		t.Errorf("the off state is not stated:\n%s", plain)
	}

	// No recorded cost still states the arithmetic.
	dash := RenderLeaderboard(results[:0], 52, true, "", true)
	if !strings.Contains(dash, "$0.00 recorded + $0.00 estimated = $0.00 shown") {
		t.Errorf("zero-cost arithmetic missing:\n%s", dash)
	}
}

func TestLeaderboardEstimatedCost(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolDevin, Name: "Devin", CurrentStreak: 5, LongestStreak: 10,
			ActiveDays: 20, TotalActivity: 5000000, Summary: thermal.Summary{Cost: 0},
			EstimatedCost: 45.50,
		},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(out, "~$45.50") {
		t.Errorf("expected ~$45.50 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "+ ~$45.50 estimated = ~$45.50 shown") {
		t.Errorf("expected estimated summary in footer, got:\n%s", out)
	}

	// With estimation disabled
	outNoEst := RenderLeaderboard(results, 52, true, "", false)
	if strings.Contains(outNoEst, "~$45.50") {
		t.Errorf("did not expect ~$45.50 when estimate is false, got:\n%s", outNoEst)
	}
	if !strings.Contains(outNoEst, "—") {
		t.Errorf("expected dash when estimate is false, got:\n%s", outNoEst)
	}
}

func TestRenderLeaderboardSortKeys_EstimatedCost(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "recorded", CurrentStreak: 1,
			TotalActivity: 100, Summary: thermal.Summary{Cost: 10},
		},
		{
			Tool: thermal.ToolDevin, Name: "estimated", CurrentStreak: 1,
			TotalActivity: 100, Summary: thermal.Summary{Cost: 0},
			EstimatedCost: 50,
		},
	}
	out := RenderLeaderboard(results, 52, true, "cost", true)
	idxEst := strings.Index(out, "estimated")
	idxRec := strings.Index(out, "recorded")
	if idxEst == -1 || idxRec == -1 || idxEst > idxRec {
		t.Errorf("expected estimated before recorded under sort=cost, got:\n%s", out)
	}
}

func TestRenderLeaderboard_MixedCostAndDisclosures(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: 5, LongestStreak: 10,
			ActiveDays: 10, TotalActivity: 100000,
			Summary:            thermal.Summary{Cost: 64.50},
			EstimatedCost:      0.16,
			MissingPricing:     []string{"mystery-model"},
			UnattributedTokens: 1500,
		},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(out, "~$64.66") {
		t.Errorf("expected ~$64.66 combined cost on row, got:\n%s", out)
	}
	if !strings.Contains(out, "$64.50 recorded") {
		t.Errorf("expected recorded cost breakdown, got:\n%s", out)
	}
	if !strings.Contains(out, "~$0.16 estimated") {
		t.Errorf("expected estimated cost breakdown, got:\n%s", out)
	}
	if !strings.Contains(out, "1.5K tokens have no model recorded") {
		t.Errorf("expected unattributed tokens disclosure, got:\n%s", out)
	}
	if !strings.Contains(out, "Unpriced models: mystery-model") {
		t.Errorf("expected missing pricing disclosure, got:\n%s", out)
	}
}

func BenchmarkRenderLeaderboard(b *testing.B) {
	results := make([]thermal.ToolResult, 0, 12)
	for i := 0; i < 6; i++ {
		results = append(results, thermal.ToolResult{
			Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: i, TotalActivity: int64(i * 1000),
		})
		results = append(results, thermal.ToolResult{
			Tool: thermal.ToolAgy, Name: "Agy", CurrentStreak: i, TotalActivity: int64(i * 10),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderLeaderboard(results, 52, true, "streak", true)
	}
}

func TestLeaderboardCaption(t *testing.T) {
	results := []thermal.ToolResult{
		{Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: 3, TotalActivity: 1000},
	}
	cases := map[string]string{
		"":       "Ranked by current streak · all time",
		"tokens": "Ranked by tokens · all time",
		"cost":   "Ranked by cost · all time",
	}
	for sortKey, want := range cases {
		if out := RenderLeaderboard(results, 52, true, sortKey, true); !strings.Contains(out, want) {
			t.Errorf("sort %q missing caption %q:\n%s", sortKey, want, out)
		}
	}
}

func TestLeaderboardNextHint(t *testing.T) {
	out := RenderLeaderboard([]thermal.ToolResult{{
		Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: 3,
		TotalActivity: 1000, Summary: thermal.Summary{Cost: 1},
	}}, 52, true, "", true)
	if !strings.Contains(out, "Next: thermal weekly --chart") {
		t.Fatalf("missing next hint:\n%s", out)
	}
}

func TestLeaderboardLegend(t *testing.T) {
	results := []thermal.ToolResult{
		{Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: 3, LongestStreak: 5, ActiveDays: 4, TotalActivity: 1000, Summary: thermal.Summary{Cost: 1.25}},
		{Tool: thermal.ToolAgy, Name: "Agy", CurrentStreak: 2, TotalActivity: 12},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	for _, want := range []string{"Strk = current streak", "Best = longest streak", "Days = active days", "CostLegend", "~$", "—", "≥$", "$0.00", "not comparable"} {
		if want == "CostLegend" {
			continue
		}
		if !strings.Contains(out, want) {
			t.Errorf("legend missing %q:\n%s", want, out)
		}
	}
}

func TestCostFooterArithmetic(t *testing.T) {
	results := []thermal.ToolResult{
		{Tool: thermal.ToolOpenCode, Name: "OpenCode", Summary: thermal.Summary{Cost: 10.50}, EstimatedCost: 2.25, TotalActivity: 1000},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(out, "$10.50 recorded + ~$2.25 estimated = ~$12.75 shown") {
		t.Fatalf("arithmetic footer missing:\n%s", out)
	}
	noEst := RenderLeaderboard(results, 52, true, "", false)
	if !strings.Contains(noEst, "estimation disabled") || !strings.Contains(noEst, "$10.50 recorded + estimation disabled = $10.50 shown") {
		t.Fatalf("no-estimate footer missing:\n%s", noEst)
	}
}

func TestUnpricedModels(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "OpenCode", TotalActivity: 1000,
			Summary:        thermal.Summary{Cost: 1, ModelBreakdown: map[string]int64{"mystery-model": 2500}},
			MissingPricing: []string{"mystery-model"},
		},
	}
	out := RenderLeaderboard(results, 52, true, "", true)
	if !strings.Contains(out, "Unpriced models: mystery-model (2.5K tok)") {
		t.Fatalf("unpriced model line missing:\n%s", out)
	}
	if strings.Contains(out, "No pricing for:") {
		t.Fatalf("old unpriced wording remains:\n%s", out)
	}
}

func TestCostMarkersShared(t *testing.T) {
	legend := CostLegendText()
	for _, marker := range []string{CostMarkerRecorded, CostMarkerEstimated, CostMarkerPartial, CostMarkerAbsent, CostMarkerApprox} {
		if !strings.Contains(legend, marker) {
			t.Errorf("shared legend missing marker %q: %s", marker, legend)
		}
	}
}
