package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestRealDataFrame renders the frame over this machine's real data and fails
// if the frame is empty. The fixture goldens prove the layout; this proves the
// loader path produces something to lay out.
func TestRealDataFrame(t *testing.T) {
	a := LoadTools(nil)
	t.Logf("tools=%d latest=%s earliest=%s skipped=%v", len(a.Tools), a.Latest, a.Earliest, a.Skipped)
	if len(a.Tools) == 0 {
		t.Skip("no tool data on this machine")
	}
	m := New(a, false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	view := updated.(Model).View()
	content := view.Content
	if strings.Count(content, "\n") < 5 {
		t.Fatalf("frame has %d lines:\n%s", strings.Count(content, "\n"), content)
	}
	lines := strings.Split(content, "\n")
	for i := 0; i < 12 && i < len(lines); i++ {
		t.Logf("%2d | %s", i, lines[i])
	}
}
