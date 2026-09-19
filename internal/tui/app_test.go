package tui

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/thermal"
)

// update regenerates golden files: go test ./internal/tui -update
var update = flag.Bool("update", false, "rewrite golden files")

// fixture builds an adapter with fixed rows and no pricer, so every assertion
// in this package is independent of the machine it runs on.
func fixture() Adapter {
	days := func(rows ...thermal.DailyRow) []thermal.DailyRow { return rows }
	return Adapter{
		Latest:   "2026-09-18",
		Earliest: "2025-09-20",
		Tools: []ToolDay{
			{
				Name: "OpenCode",
				Days: days(
					thermal.DailyRow{Day: "2026-09-18", Tokens: 2_193_493, Input: 101_725, Output: 6_712, Cache: 2_064_256, Turns: 3, Cost: 12.5},
					thermal.DailyRow{Day: "2026-09-17", Tokens: 1_800_000, Input: 90_000, Output: 5_000, Cache: 1_700_000, Turns: 2, Cost: 8.25},
					thermal.DailyRow{Day: "2026-09-10", Tokens: 900_000, Input: 40_000, Output: 3_000, Cache: 850_000, Turns: 1, Cost: 4.0},
					// Outside the 30 day window: RangeAll must exceed Range30d.
					thermal.DailyRow{Day: "2026-06-01", Tokens: 500_000, Input: 20_000, Output: 2_000, Cache: 470_000, Turns: 1, Cost: 2.5},
				),
				Projects: []thermal.ProjectDay{
					{Day: "2026-09-18", Project: "atlas-gateway", Tokens: 2_193_493, Turns: 3},
					{Day: "2026-09-17", Project: "beacon-api", Tokens: 1_800_000, Turns: 2},
				},
			},
			{
				Name: "Codex",
				Days: days(
					thermal.DailyRow{Day: "2026-09-18", Tokens: 938_700_000, Input: 800_000_000, Output: 138_700_000, Turns: 12, Cost: 42.0},
					thermal.DailyRow{Day: "2026-09-16", Tokens: 300_000_000, Input: 260_000_000, Output: 40_000_000, Turns: 5},
				),
			},
			{
				Name: "Agy",
				Days: days(
					thermal.DailyRow{Day: "2026-09-18", Tokens: 60_100, Turns: 36},
					thermal.DailyRow{Day: "2026-09-15", Tokens: 41_000, Turns: 20},
				),
			},
		},
	}
}

func renderFixture(t *testing.T, width, height int) string {
	t.Helper()
	m := New(fixture(), false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	view := updated.(Model).View()
	return ansi.Strip(view.Content)
}

// TestOverviewGolden pins the whole frame at the two sizes the goal names.
// Colour is off, so a diff is a real layout change rather than a palette one.
func TestOverviewGolden(t *testing.T) {
	cases := []struct {
		name string
		w, h int
		file string
	}{
		{"80x24", 80, 24, "overview_80x24.golden"},
		{"120x40", 120, 40, "overview_120x40.golden"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderFixture(t, tc.w, tc.h)
			path := filepath.Join("testdata", tc.file)
			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("golden missing: %v (run: go test ./internal/tui -update)", err)
			}
			if got != string(want) {
				t.Errorf("frame changed at %s\n--- got ---\n%s\n--- want ---\n%s", tc.name, got, want)
			}
		})
	}
}

// TestViewFitsWidth proves the frame never exceeds the terminal, which is what
// keeps the dashboard readable from 80 to 200 columns.
func TestViewFitsWidth(t *testing.T) {
	for _, w := range []int{80, 100, 130, 160, 200} {
		got := renderFixture(t, w, 40)
		for i, line := range splitLines(got) {
			if n := len([]rune(line)); n > w {
				t.Errorf("width %d: line %d is %d cells: %q", w, i, n, line)
			}
		}
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
