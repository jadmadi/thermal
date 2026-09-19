package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

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

	table := renderProjects(m.adapter.BuildProjects(m.rng, m.projSort, ""), 116, m.palette)
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
