package render

import (
	"strings"
	"testing"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderHeatmap_Dimensions(t *testing.T) {
	act := map[string]thermal.DayActivity{
		"2026-07-15": {Tokens: 1500, Turns: 5},
	}
	lines := RenderHeatmap(act, 52, false)
	// Must have 1 header line + 7 weekday lines + 1 legend line = 9 total lines
	if len(lines) != 9 {
		t.Fatalf("expected 9 lines in heatmap output, got %d", len(lines))
	}
	if !strings.Contains(lines[8], "Less") || !strings.Contains(lines[8], "More") {
		t.Errorf("legend line missing Less/More: %s", lines[8])
	}
}
