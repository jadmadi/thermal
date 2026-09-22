// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

// TestChartScalesWithWidth proves bars grow with the terminal up to the cap and
// never exceed the frame, which is what keeps pasted output readable.
func TestChartScalesWithWidth(t *testing.T) {
	rep := thermal.ProjectReport{Type: "projects"}
	for i := 0; i < 3; i++ {
		rep.Rows = append(rep.Rows, thermal.ProjectRow{
			Project: "project-name", Tokens: int64((3 - i) * 1_000_000),
		})
	}
	var prev int
	for _, w := range []int{80, 100, 120, 160, 200} {
		out := RenderProjectChart(rep, nil, w, true)
		var widest int
		for _, line := range strings.Split(out, "\n") {
			if n := len([]rune(line)); n > widest {
				widest = n
			}
		}
		if widest > w {
			t.Errorf("width %d: chart drew %d cells", w, widest)
		}
		// Measure the run of filled cells that follows the longest run of
		// filler: that pair is the bar, in order, on every line.
		filled := 0
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "█") && strings.Contains(line, "·") {
				filled = strings.Count(line[strings.Index(line, "█"):strings.LastIndex(line, "·")+1], "█")
				break
			}
		}
		t.Logf("width %d -> widest line %d, first bar %d cells", w, widest, filled)
		if filled < prev && w >= 100 {
			t.Errorf("the bar did not grow from width %d to %d (%d -> %d)", w-20, w, prev, filled)
		}
		prev = filled
	}
}
