// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestProbeRealViewport prints the frame at a small size after moving the
// cursor deep into the list, which is where the window has to travel.
func TestProbeRealViewport(t *testing.T) {
	a := LoadTools(nil)
	if len(a.Tools) == 0 {
		t.Skip("no data")
	}
	m := New(a, false)
	m.tab = tabProjects
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(Model)
	rows := len(m.adapter.BuildProjects(m.rng, m.projSort, "").Rows)
	t.Logf("projects in window: %d", rows)

	for i := 0; i < 40 && i < rows; i++ {
		updated, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		m = updated.(Model)
	}
	view := ansi.Strip(m.View().Content)
	for i, line := range strings.Split(view, "\n") {
		t.Logf("P %2d | %s", i, line)
	}

	// The same walk for Models, which has its own chrome height.
	m2 := New(a, false)
	m2.tab = tabModels
	updated, _ = m2.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m2 = updated.(Model)
	models := len(m2.adapter.BuildModels(m2.rng, m2.metric, m2.modelsSort).Rows)
	t.Logf("models in window: %d", models)
	for i := 0; i < 15 && i < models; i++ {
		m2 = press(t, m2, "j")
	}
	out := ansi.Strip(m2.View().Content)
	for i, line := range strings.Split(out, "\n") {
		t.Logf("M %2d | %s", i, line)
	}
}
