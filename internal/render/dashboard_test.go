package render

import (
	"strings"
	"testing"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderDashboard_Output(t *testing.T) {
	sum := thermal.Summary{
		Sessions:       5,
		LifetimeTokens: 250000,
		InputTokens:    150000,
		OutputTokens:   100000,
		ReasoningTokens: 10000,
		CacheTokens:    5000,
		Cost:           1.45,
		ModelBreakdown: map[string]int64{"gpt-4o": 5},
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
}
