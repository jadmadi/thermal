// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestFrameFitsHeight proves the frame never grows past the terminal. A frame
// one line too tall makes the whole UI scroll, which pushes the tab bar off
// screen and breaks the window calculation with it.
func TestFrameFitsHeight(t *testing.T) {
	a := LoadTools(nil)
	if len(a.Tools) == 0 {
		t.Skip("no tool data on this machine")
	}
	for _, h := range []int{24, 30, 40, 60} {
		for _, tab := range []int{tabOverview, tabProjects, tabMix, tabModels, tabStats} {
			m := New(a, false)
			m.tab = tab
			updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: h})
			m = updated.(Model)
			// Walk to the bottom of any list, where the window is under the
			// most pressure.
			for i := 0; i < 60; i++ {
				m = press(t, m, "j")
			}
			lines := len(strings.Split(m.View().Content, "\n"))
			if lines > h {
				t.Errorf("height %d, tab %d: frame is %d lines", h, tab, lines)
			}
		}
	}
}

// TestRealDataProjects renders the Projects table and a drill-down over this
// machine's real data, which is where a sparse fixture cannot reach.
func TestRealDataProjects(t *testing.T) {
	a := LoadTools(nil)
	if len(a.Tools) == 0 {
		t.Skip("no tool data on this machine")
	}
	m := New(a, false)
	m.tab = tabProjects
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	table := renderProjects(m.adapter.BuildProjects(m.rng, m.projSort, ""), 116, 40, 0, m.palette)
	if strings.Count(table, "\n") < 6 {
		t.Fatalf("table too short:\n%s", table)
	}
	for i, line := range strings.Split(table, "\n") {
		if i > 10 {
			break
		}
		t.Logf("T %2d | %s", i, line)
	}

	data := m.adapter.BuildProjects(m.rng, m.projSort, "")
	if len(data.Rows) == 0 {
		t.Fatal("no project rows in the window")
	}
	detail := m.adapter.BuildProjectDetail(data.Rows[0].Project, m.rng)
	out := renderProjectDetail(detail, 116, m.palette)
	for i, line := range strings.Split(out, "\n") {
		if i > 18 {
			break
		}
		t.Logf("D %2d | %s", i, line)
	}
}
