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
	if !strings.Contains(withEstimate, "Reports estimate the rest") {
		t.Errorf("estimate disclosure missing:\n%s", withEstimate)
	}

	plain := RenderLeaderboard(results, 52, true, "", false)
	if !strings.Contains(plain, "$12.75") {
		t.Errorf("recorded total missing with estimation off:\n%s", plain)
	}
	if !strings.Contains(plain, "Estimates are off") {
		t.Errorf("the off state is not stated:\n%s", plain)
	}

	// No recorded cost at all: the line would be noise, so it must not print.
	dash := RenderLeaderboard(results[:0], 52, true, "", true)
	if strings.Contains(dash, "Recorded cost") {
		t.Errorf("cost line printed with no recorded cost:\n%s", dash)
	}
}
