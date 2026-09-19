package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestRealDataMixAndModels renders both new tabs over this machine's data,
// which is denser than any fixture and exposes layout and share problems.
func TestRealDataMixAndModels(t *testing.T) {
	a := LoadTools(nil)
	if len(a.Tools) == 0 {
		t.Skip("no tool data on this machine")
	}
	m := New(a, false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	for _, tab := range []int{tabMix, tabModels} {
		m.tab = tab
		out := ""
		if tab == tabMix {
			out = renderMix(m.buildMixView(), 116, 40, m.palette)
		} else {
			out = renderModels(m.buildModelsView(), 116, 40, 0, m.palette)
		}
		lines := strings.Split(out, "\n")
		if len(lines) < 8 {
			t.Fatalf("tab %d rendered %d lines", tab, len(lines))
		}
		for _, line := range lines {
			if len([]rune(line)) > 116 {
				t.Errorf("tab %d line exceeds the frame: %q", tab, line)
			}
		}
		t.Logf("--- tab %d ---", tab)
		for i, line := range lines {
			if i > 14 {
				break
			}
			t.Logf("%2d | %s", i, line)
		}
	}
}
