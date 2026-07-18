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
	out := RenderLeaderboard(results, 52, true)
	if !strings.Contains(out, "Token Warriors") {
		t.Errorf("expected Token Warriors section in output")
	}
	if !strings.Contains(out, "OpenCode") {
		t.Errorf("expected OpenCode in output")
	}
}
