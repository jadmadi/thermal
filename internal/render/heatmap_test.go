// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestMonthHeader_NoOverlap(t *testing.T) {
	cases := []struct {
		name        string
		firstSunday time.Time
		weeks       int
		want        string
	}{
		{"late october", time.Date(2025, 10, 26, 0, 0, 0, 0, time.UTC), 52, "Nov"},
		{"year boundary", time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC), 52, "Jan"},
		{"short window", time.Date(2025, 10, 26, 0, 0, 0, 0, time.UTC), 4, "Oct"},
		{"104 weeks", time.Date(2024, 10, 27, 0, 0, 0, 0, time.UTC), 104, "Nov"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			header := monthHeader(tc.firstSunday, tc.weeks)
			if len(header) != tc.weeks {
				t.Fatalf("header width = %d, want %d", len(header), tc.weeks)
			}
			if strings.Contains(header, "ONov") {
				t.Errorf("overlapping October/November labels: %q", header)
			}
			if !strings.Contains(header, tc.want) {
				t.Errorf("header %q missing %q", header, tc.want)
			}
			for _, field := range strings.Fields(header) {
				if len(field) > 3 {
					t.Errorf("month label %q is longer than 3 cells (overlap): %q", field, header)
				}
			}
		})
	}
}

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
