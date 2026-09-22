// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/thermal"
)

// projectsFixture mirrors the project rows the adapter holds on a real machine:
// several projects, more than one tool per project, and one project with no
// model attribution so the coverage note has something to say.
func projectsFixture() Adapter {
	return Adapter{
		Latest:   "2026-09-18",
		Earliest: "2025-09-20",
		Tools: []ToolDay{
			{
				Name: "OpenCode",
				Days: []thermal.DailyRow{
					{
						Day: "2026-09-18", Tokens: 2_193_493, Input: 101_725, Output: 6_712, Cache: 2_064_256, Turns: 3, Cost: 12.5,
						// A tool day carries its own model split: the Models
						// view reads days, not project rows.
						Models: map[string]thermal.ModelTokens{
							"deepseek-v4.1-flash": {Input: 101_725, Output: 6_712, CacheRead: 2_064_256},
						},
					},
				},
				Projects: []thermal.ProjectDay{
					{
						Day: "2026-09-18", Project: "atlas-gateway", Tool: "OpenCode",
						Tokens: 2_193_493, Input: 101_725, Output: 6_712, CacheRead: 2_064_256, Turns: 3, Cost: 12.5,
						Models: map[string]thermal.ModelTokens{
							"deepseek-v4.1-flash": {Input: 101_725, Output: 6_712, CacheRead: 2_064_256},
						},
					},
					{
						Day: "2026-09-17", Project: "beacon-api", Tool: "OpenCode",
						Tokens: 1_800_000, Input: 90_000, Output: 5_000, CacheRead: 1_700_000, Turns: 2, Cost: 8.25,
						Models: map[string]thermal.ModelTokens{
							"deepseek-v4.1-flash": {Input: 90_000, Output: 5_000, CacheRead: 1_700_000},
						},
					},
				},
			},
			{
				Name: "Devin",
				Days: []thermal.DailyRow{
					{
						Day: "2026-09-18", Tokens: 900_000_000, Turns: 8,
						Models: map[string]thermal.ModelTokens{
							"omen-alpha": {Input: 700_000_000, Output: 200_000_000},
						},
					},
				},
				Projects: []thermal.ProjectDay{
					// Devin records no models: this row is what makes the note appear.
					{Day: "2026-09-18", Project: "atlas-gateway", Tool: "Devin", Tokens: 900_000_000, Turns: 8, Cost: 41.0},
				},
			},
			{
				Name: "ZCode",
				Days: []thermal.DailyRow{
					{
						Day: "2026-09-16", Tokens: 42_000_000, Turns: 4,
						Models: map[string]thermal.ModelTokens{
							"glm-5.3-flash": {Input: 40_000_000, Output: 2_000_000},
						},
					},
				},
				Projects: []thermal.ProjectDay{
					{
						Day: "2026-09-16", Project: "northwind-cli", Tool: "ZCode",
						Tokens: 42_000_000, Turns: 4,
						Models: map[string]thermal.ModelTokens{
							"glm-5.3-flash": {Input: 40_000_000, Output: 2_000_000},
						},
					},
				},
			},
		},
	}
}

func renderProjectsFixture(t *testing.T, width, height int, drill string) string {
	t.Helper()
	m := New(projectsFixture(), false)
	// The Projects tab is index 1.
	m.tab = 1
	if drill != "" {
		m.drill = drill
		m.detail = m.adapter.BuildProjectDetail(drill, m.rng)
	}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	view := updated.(Model).View()
	return ansi.Strip(view.Content)
}

// TestProjectsParity proves the view totals equal the static aggregation for
// the same window. The dashboard must never disagree with `thermal projects`.
func TestProjectsParity(t *testing.T) {
	a := projectsFixture()
	got := a.BuildProjects(Range30d, ProjectSortTokens, "")

	rep := thermal.AggregateProjects(a.projectDays(""), thermal.ProjectOptions{
		Since: got.Start,
		Until: got.End,
	}, a.Pricer)

	if got.Tokens != rep.Totals.Tokens {
		t.Errorf("view tokens = %d, static = %d", got.Tokens, rep.Totals.Tokens)
	}
	if diff := got.Cost - rep.Totals.Cost; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("view cost = %.4f, static = %.4f", got.Cost, rep.Totals.Cost)
	}
	if got.Days != rep.Totals.ActiveDays {
		t.Errorf("view active days = %d, static = %d", got.Days, rep.Totals.ActiveDays)
	}
	if len(got.Rows) != len(rep.Rows) {
		t.Fatalf("view rows = %d, static rows = %d", len(got.Rows), len(rep.Rows))
	}

	// Every row must also agree, not just the total.
	byProject := make(map[string]thermal.ProjectRow, len(rep.Rows))
	for _, r := range rep.Rows {
		byProject[r.Project] = r
	}
	for _, row := range got.Rows {
		want, ok := byProject[row.Project]
		if !ok {
			t.Errorf("view has %s, static does not", row.Project)
			continue
		}
		if row.Tokens != want.Tokens {
			t.Errorf("%s tokens: view %d, static %d", row.Project, row.Tokens, want.Tokens)
		}
		if row.Days != want.ActiveDays {
			t.Errorf("%s days: view %d, static %d", row.Project, row.Days, want.ActiveDays)
		}
	}
}

// TestProjectsSortOrders proves each sort key actually reorders, so the cycle
// is not decorative.
func TestProjectsSortOrders(t *testing.T) {
	a := projectsFixture()
	byTokens := a.BuildProjects(Range30d, ProjectSortTokens, "")
	if byTokens.Rows[0].Project != "atlas-gateway" {
		t.Fatalf("leading project by tokens = %s, want atlas-gateway", byTokens.Rows[0].Project)
	}

	byCost := a.BuildProjects(Range30d, ProjectSortCost, "")
	if byCost.Rows[0].Project != "atlas-gateway" {
		t.Errorf("leading by cost = %s, want atlas-gateway", byCost.Rows[0].Project)
	}

	byDays := a.BuildProjects(Range30d, ProjectSortDays, "")
	for i := 1; i < len(byDays.Rows); i++ {
		if byDays.Rows[i-1].Days < byDays.Rows[i].Days {
			t.Errorf("days sort is not descending: %d before %d", byDays.Rows[i-1].Days, byDays.Rows[i].Days)
		}
	}

	byRecent := a.BuildProjects(Range30d, ProjectSortRecent, "")
	for i := 1; i < len(byRecent.Rows); i++ {
		if byRecent.Rows[i-1].LastDay < byRecent.Rows[i].LastDay {
			t.Errorf("recent sort is not descending: %s before %s", byRecent.Rows[i-1].LastDay, byRecent.Rows[i].LastDay)
		}
	}
}

// TestProjectsToolFilter proves the filter narrows the data and that clearing
// it restores the full set.
func TestProjectsToolFilter(t *testing.T) {
	a := projectsFixture()
	all := a.BuildProjects(Range30d, ProjectSortTokens, "")
	filtered := a.BuildProjects(Range30d, ProjectSortTokens, "ZCode")

	if len(filtered.Rows) >= len(all.Rows) {
		t.Errorf("filtered rows = %d, all rows = %d; filter did nothing", len(filtered.Rows), len(all.Rows))
	}
	if filtered.ToolName != "ZCode" {
		t.Errorf("filter label = %q, want ZCode", filtered.ToolName)
	}
	for _, row := range filtered.Rows {
		found := false
		for _, tool := range row.Tools {
			if tool == "ZCode" {
				found = true
			}
		}
		if !found {
			t.Errorf("row %s has no ZCode in its tools: %v", row.Project, row.Tools)
		}
	}
	if filtered.Tokens >= all.Tokens {
		t.Errorf("filtered tokens = %d, all = %d", filtered.Tokens, all.Tokens)
	}
}

// TestProjectsDetail proves the drill-down reports the project it was asked
// for and states its model coverage, since attribution is partial.
func TestProjectsDetail(t *testing.T) {
	a := projectsFixture()
	d := a.BuildProjectDetail("atlas-gateway", Range30d)

	if d.Project != "atlas-gateway" {
		t.Fatalf("detail project = %s", d.Project)
	}
	if d.Tokens != 2_193_493+900_000_000 {
		t.Errorf("detail tokens = %d, want the sum of both tools", d.Tokens)
	}
	if len(d.Tools) != 2 {
		t.Errorf("detail tools = %d, want 2 (OpenCode and Devin)", len(d.Tools))
	}
	if d.Tools[0].Name != "Devin" {
		t.Errorf("largest tool = %s, want Devin", d.Tools[0].Name)
	}
	// Model coverage is partial, so the note must exist and must not claim full
	// attribution.
	if d.ModelNote == "" {
		t.Error("expected a model coverage note when a tool records no models")
	}
	if strings.Contains(d.ModelNote, "100") {
		t.Errorf("coverage note overstates: %s", d.ModelNote)
	}
	if len(d.Weekly) != 12 {
		t.Errorf("weekly buckets = %d, want 12", len(d.Weekly))
	}
	if len(d.Heatmap) == 0 {
		t.Error("heatmap is empty")
	}
}

// TestProjectsGolden pins the table and the drill-down at two sizes.
func TestProjectsGolden(t *testing.T) {
	for _, tc := range []struct {
		name  string
		w, h  int
		drill string
		file  string
	}{
		{"table 80x24", 80, 24, "", "projects_80x24.golden"},
		{"table 120x40", 120, 40, "", "projects_120x40.golden"},
		{"detail 80x24", 80, 24, "atlas-gateway", "project_detail_80x24.golden"},
		{"detail 120x40", 120, 40, "atlas-gateway", "project_detail_120x40.golden"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderProjectsFixture(t, tc.w, tc.h, tc.drill)
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

// TestProjectsKeyFlow covers the keys this view owns, including drill-down and
// back, which is where a shell and a view most often disagree about state.
func TestProjectsKeyFlow(t *testing.T) {
	m := New(projectsFixture(), false)
	m.tab = 1

	t.Run("s cycles the project sort", func(t *testing.T) {
		seen := map[ProjectSort]bool{m.projSort: true}
		for i := 0; i < len(ProjectSortOrder); i++ {
			m = press(t, m, "s")
			seen[m.projSort] = true
		}
		for _, want := range ProjectSortOrder {
			if !seen[want] {
				t.Errorf("cycling never reached %s", want)
			}
		}
	})

	t.Run("f cycles the tool filter and clears", func(t *testing.T) {
		m = press(t, m, "f")
		if m.projFilter == "" {
			t.Fatal("f did not set a filter")
		}
		first := m.projFilter
		// Cycling through every option must return to no filter.
		for i := 0; i < 20 && m.projFilter != ""; i++ {
			m = press(t, m, "f")
		}
		if m.projFilter != "" {
			t.Errorf("filter never cleared, stopped at %q", m.projFilter)
		}
		m = press(t, m, "f")
		if m.projFilter != first {
			t.Errorf("filter cycle restarted at %q, want %q", m.projFilter, first)
		}
		m = press(t, m, "f")
		for m.projFilter != "" {
			m = press(t, m, "f")
		}
	})

	t.Run("enter opens the detail and esc returns", func(t *testing.T) {
		m.projSel = 0
		updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		m = updated.(Model)
		if m.drill == "" {
			t.Fatal("enter did not open a detail")
		}
		if m.detail.Project != m.drill {
			t.Errorf("detail is for %s but drill says %s", m.detail.Project, m.drill)
		}
		updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		m = updated.(Model)
		if m.drill != "" {
			t.Errorf("esc did not close the detail, still on %s", m.drill)
		}
	})

	t.Run("j and k move the selection and wrap", func(t *testing.T) {
		m.drill = ""
		m.projSel = 0
		m = press(t, m, "j")
		if m.projSel != 1 {
			t.Errorf("j: selection = %d, want 1", m.projSel)
		}
		m = press(t, m, "k")
		if m.projSel != 0 {
			t.Errorf("k: selection = %d, want 0", m.projSel)
		}
		m = press(t, m, "k")
		if m.projSel != len(m.adapter.BuildProjects(m.rng, m.projSort, m.projFilter).Rows)-1 {
			t.Errorf("k from the top did not wrap to the last row, selection = %d", m.projSel)
		}
	})
}

// TestHeatGlyphRamp proves the level mapping, which a sparse fixture cannot:
// with one active day the shared threshold function collapses every day to
// level 1, so the ramp is asserted by calling the renderer directly.
func TestHeatGlyphRamp(t *testing.T) {
	days := []HeatDay{
		{Day: "2026-09-14", Level: 0},
		{Day: "2026-09-15", Level: 1},
		{Day: "2026-09-16", Level: 2},
		{Day: "2026-09-17", Level: 3},
		{Day: "2026-09-18", Level: 4},
	}
	out := renderHeat(days, 20, newPalette(false))
	for _, want := range []string{"░", "▒", "▓", "█"} {
		if !strings.Contains(out, want) {
			t.Errorf("ramp is missing glyph %q in:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "·") {
		t.Errorf("quiet day did not render as a dot:\n%s", out)
	}
}
