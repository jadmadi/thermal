// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/thermal"
)

// manyProjects builds more projects than any terminal can show, so the window
// has to move. Without this the cursor test would pass by accident.
func manyProjects(n int) Adapter {
	a := Adapter{Latest: "2026-09-18", Earliest: "2025-09-20"}
	for i := 0; i < n; i++ {
		a.Tools = append(a.Tools, ToolDay{
			Name: fmt.Sprintf("Tool%02d", i),
			Days: []thermal.DailyRow{{Day: "2026-09-18", Tokens: int64(1_000_000 - i*10_000), Turns: 1}},
			Projects: []thermal.ProjectDay{{
				Day: "2026-09-18", Project: fmt.Sprintf("project-%02d", i),
				Tool: fmt.Sprintf("Tool%02d", i), Tokens: int64(1_000_000 - i*10_000), Turns: 1,
			}},
		})
	}
	return a
}

// TestProjectsCursorStaysVisible is the rule this window exists for: wherever
// the cursor goes, that row is on screen. A cursor that walks off the frame is
// worse than no cursor, because the user cannot tell where it went.
func TestProjectsCursorStaysVisible(t *testing.T) {
	a := manyProjects(80)
	rows := len(a.BuildProjects(Range30d, ProjectSortTokens, "").Rows)
	if rows < 40 {
		t.Fatalf("fixture only produced %d rows; the test needs more than a screen", rows)
	}

	for _, size := range []struct{ w, h int }{{80, 24}, {120, 40}, {100, 12}} {
		m := New(a, false)
		m.tab = tabProjects

		for move := 0; move < rows; move++ {
			updated, _ := m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
			m = updated.(Model)
			// Walk one row at a time, checking each position.
			updated, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
			m = updated.(Model)

			view := ansi.Strip(m.View().Content)
			want := a.BuildProjects(m.rng, m.projSort, m.projFilter).Rows[m.projSel].Project
			label := labelFor(projectLabels(a.BuildProjects(m.rng, m.projSort, m.projFilter).Rows), want)
			if !strings.Contains(view, truncateRunes(label, pcolName)) {
				t.Fatalf("size %dx%d: cursor at row %d (%s) is not on screen:\n%s",
					size.w, size.h, m.projSel, label, view)
			}
		}
	}
}

// TestProjectsWindowReportsWhatItHides proves a windowed table says so. A short
// table that silently omits rows reads as the whole story.
func TestProjectsWindowReportsWhatItHides(t *testing.T) {
	a := manyProjects(80)
	m := New(a, false)
	m.tab = tabProjects
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(Model)

	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "below") {
		t.Errorf("windowed table does not say how many rows are hidden:\n%s", view)
	}

	// Walk to the bottom: the marker must flip to rows above.
	for i := 0; i < 79; i++ {
		updated, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		m = updated.(Model)
	}
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "above") {
		t.Errorf("at the last row the marker should count rows above:\n%s", view)
	}
}

// TestViewportWindow pins the arithmetic, including the ends, where an
// off-by-one would hide either the first or the last row forever.
func TestViewportWindow(t *testing.T) {
	for _, tc := range []struct {
		name      string
		total     int
		cursor    int
		available int
		wantFirst int
		wantLast  int
	}{
		{"everything fits", 5, 2, 10, 0, 5},
		{"cursor at the top", 50, 0, 10, 0, 10},
		{"cursor at the bottom", 50, 49, 10, 40, 50},
		{"cursor in the middle", 50, 25, 10, 20, 30},
		{"one row tall", 50, 10, 1, 10, 11},
		{"no room at all", 50, 10, 0, 10, 11},
		{"cursor past the end", 50, 999, 10, 40, 50},
		{"cursor before the start", 50, -5, 10, 0, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, last := viewportWindow(tc.total, tc.cursor, tc.available)
			if first != tc.wantFirst || last != tc.wantLast {
				t.Errorf("viewportWindow(%d, %d, %d) = %d, %d; want %d, %d",
					tc.total, tc.cursor, tc.available, first, last, tc.wantFirst, tc.wantLast)
			}
			// The invariant, checked on every case rather than only the table.
			cursor := tc.cursor
			if cursor < 0 {
				cursor = 0
			}
			if cursor >= tc.total {
				cursor = tc.total - 1
			}
			if cursor < first || cursor >= last {
				t.Errorf("cursor %d is outside the window %d..%d", cursor, first, last)
			}
		})
	}
}

// TestModelsCursorStaysVisible applies the same rule to the Models table, which
// has the same shape and the same risk.
func TestModelsCursorStaysVisible(t *testing.T) {
	a := manyModels(60)
	rows := len(a.BuildModels(Range30d, MetricTokens, ModelSortTokens).Rows)
	if rows < 30 {
		t.Fatalf("fixture only produced %d models", rows)
	}

	for _, size := range []struct{ w, h int }{{80, 24}, {120, 40}} {
		m := New(a, false)
		m.tab = tabModels
		for move := 0; move < rows; move++ {
			updated, _ := m.Update(tea.WindowSizeMsg{Width: size.w, Height: size.h})
			m = updated.(Model)
			m = press(t, m, "j")

			view := ansi.Strip(m.View().Content)
			name := m.adapter.BuildModels(m.rng, m.metric, m.modelsSort).Rows[m.modelsSel].Name
			if !strings.Contains(view, name) {
				t.Fatalf("size %dx%d: cursor at model %d (%s) is not on screen:\n%s",
					size.w, size.h, m.modelsSel, name, view)
			}
		}
	}
}

// manyModels gives one tool a long model list, which is what a real machine
// with several providers looks like.
func manyModels(n int) Adapter {
	a := Adapter{Latest: "2026-09-18", Earliest: "2025-09-20"}
	models := make(map[string]thermal.ModelTokens, n)
	for i := 0; i < n; i++ {
		models[fmt.Sprintf("model-%02d", i)] = thermal.ModelTokens{Input: int64(1_000_000 - i*10_000)}
	}
	a.Tools = append(a.Tools, ToolDay{
		Name: "Tool",
		Days: []thermal.DailyRow{{
			Day: "2026-09-18", Tokens: 1_000_000, Turns: 1, Models: models,
		}},
	})
	return a
}
