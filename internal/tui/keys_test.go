package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// press sends a single printable key press through Update.
// press sends a printable key and returns the model Update produced. It takes
// the value by address so a handler that mutates a pointer receiver still
// updates the model the caller holds.
func press(t *testing.T, m Model, ch string) Model {
	t.Helper()
	r := []rune(ch)[0]
	updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: ch})
	switch v := updated.(type) {
	case Model:
		return v
	case *Model:
		return *v
	default:
		t.Fatalf("Update returned %T, want Model or *Model", updated)
		return m
	}
}

// TestKeyFlow covers the interactions the shell owns: tab switching, the three
// controls, help, and quit. It asserts state, not pixels, so a layout change
// never breaks it.
func TestKeyFlow(t *testing.T) {
	t.Run("number keys select tabs", func(t *testing.T) {
		m := New(fixture(), false)
		for i, key := range []string{"2", "3", "4", "5", "1"} {
			m = press(t, m, key)
			want := map[string]int{"2": 1, "3": 2, "4": 3, "5": 4, "1": 0}[key]
			if m.tab != want {
				t.Fatalf("after key %s tab = %d, want %d (case %d)", key, m.tab, want, i)
			}
		}
	})

	t.Run("tab and shift-tab cycle", func(t *testing.T) {
		m := New(fixture(), false)
		updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = updated.(Model)
		if m.tab != 1 {
			t.Fatalf("tab: tab = %d, want 1", m.tab)
		}
		updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
		m = updated.(Model)
		if m.tab != 0 {
			t.Fatalf("shift+tab: tab = %d, want 0", m.tab)
		}
		// Wrap backwards off the first tab.
		updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
		m = updated.(Model)
		if m.tab != len(Tabs)-1 {
			t.Fatalf("shift+tab wrap: tab = %d, want %d", m.tab, len(Tabs)-1)
		}
	})

	t.Run("r cycles the range and re-aggregates", func(t *testing.T) {
		m := New(fixture(), false)
		if m.overview().Range != Range30d {
			t.Fatalf("default range = %s, want 30d", m.overview().Range)
		}
		seen := map[Range]bool{}
		for i := 0; i < len(RangeOrder); i++ {
			m = press(t, m, "r")
			seen[m.overview().Range] = true
		}
		for _, r := range RangeOrder {
			if !seen[r] {
				t.Errorf("cycling never reached range %s", r)
			}
		}
		if m.overview().Range != Range30d {
			t.Errorf("cycle did not return to 30d, got %s", m.overview().Range)
		}
	})

	t.Run("t toggles the metric", func(t *testing.T) {
		m := New(fixture(), false)
		m = press(t, m, "t")
		if m.overview().Metric != MetricCost {
			t.Fatalf("metric = %s, want cost", m.overview().Metric)
		}
		m = press(t, m, "t")
		if m.overview().Metric != MetricTokens {
			t.Fatalf("metric = %s, want tokens", m.overview().Metric)
		}
	})

	t.Run("s cycles the sort and reorders rows", func(t *testing.T) {
		m := New(fixture(), false)
		byTokens := m.overview().Rows[0].Tool
		m = press(t, m, "s") // cost
		if m.overview().Sort != SortCost {
			t.Fatalf("sort = %s, want cost", m.overview().Sort)
		}
		m = press(t, m, "s") // name
		rows := m.overview().Rows
		if m.overview().Sort != SortName {
			t.Fatalf("sort = %s, want name", m.overview().Sort)
		}
		for i := 1; i < len(rows); i++ {
			if rows[i-1].Tool > rows[i].Tool {
				t.Fatalf("name sort is not ascending: %s before %s", rows[i-1].Tool, rows[i].Tool)
			}
		}
		m = press(t, m, "s") // back to tokens
		if m.overview().Rows[0].Tool != byTokens {
			t.Errorf("returning to token sort did not restore the leading tool")
		}
	})

	t.Run("? toggles help and esc closes it", func(t *testing.T) {
		m := New(fixture(), false)
		m = press(t, m, "?")
		if !m.showHe {
			t.Fatal("? did not open help")
		}
		updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		m = updated.(Model)
		if m.showHe {
			t.Fatal("esc did not close help")
		}
	})

	t.Run("q quits", func(t *testing.T) {
		m := New(fixture(), false)
		_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		if cmd == nil {
			t.Fatal("q returned no command")
		}
	})

	t.Run("reserved keys explain themselves", func(t *testing.T) {
		m := New(fixture(), false)
		m = press(t, m, "/")
		if !strings.Contains(m.status, "Projects view") {
			t.Errorf("filter status = %q, want a pointer to the later goal", m.status)
		}
		updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		m = updated.(Model)
		if !strings.Contains(m.status, "Projects view") {
			t.Errorf("enter status = %q, want a pointer to the later goal", m.status)
		}
	})
}

// TestOverviewParityAgainstStatic totals proves the dashboard and the static
// reports agree on the same rows. The adapter is fed the same DailyRow values a
// loader would emit, and the expected number is summed by hand.
func TestOverviewParityAgainstStatic(t *testing.T) {
	a := fixture()
	ov := a.BuildOverview(Range30d, MetricTokens, SortTokens)

	// Sum only the rows the window includes, which is what a static report for
	// the same window would print.
	var wantTokens, wantTurns int64
	var wantCost float64
	start, end := a.window(Range30d)
	for _, tool := range a.Tools {
		for _, d := range tool.Days {
			if !inWindow(d.Day, start, end) {
				continue
			}
			wantTokens += d.Tokens
			wantTurns += int64(d.Turns)
			wantCost += d.Cost
		}
	}
	if ov.Tokens != wantTokens {
		t.Errorf("dashboard tokens = %d, static sum = %d", ov.Tokens, wantTokens)
	}
	if int64(ov.Turns) != wantTurns {
		t.Errorf("dashboard turns = %d, static sum = %d", ov.Turns, wantTurns)
	}
	if diff := ov.Cost - wantCost; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("dashboard cost = %.4f, static sum = %.4f", ov.Cost, wantCost)
	}

	// The same window for one tool must equal that tool's own row.
	codex := ov.Rows[0]
	if codex.Tool != "Codex" {
		t.Fatalf("leading tool = %s, want Codex", codex.Tool)
	}
	if codex.Tokens != 938_700_000+300_000_000 {
		t.Errorf("Codex row tokens = %d, want the sum of its two days", codex.Tokens)
	}
}

// TestOverviewExcludesOutsideRange proves the window is respected, since a
// dashboard that quietly counts everything would still look plausible.
func TestOverviewExcludesOutsideRange(t *testing.T) {
	a := fixture()
	a.Latest = "2026-09-18"
	short := a.BuildOverview(Range30d, MetricTokens, SortTokens)
	all := a.BuildOverview(RangeAll, MetricTokens, SortTokens)
	if all.Tokens <= short.Tokens {
		t.Errorf("all-time tokens (%d) should exceed the 30 day window (%d)", all.Tokens, short.Tokens)
	}
	if short.Start != "2026-08-20" {
		t.Errorf("30d window start = %s, want 2026-08-20", short.Start)
	}
}

// TestProgramStartsAndExits proves the real binary honours its guards and
// leaves cleanly. A pseudo terminal from script(1) does not answer the
// capability queries bubbletea v2 sends at startup, so the frame itself is
// asserted by TestProbeRealData and the golden files instead.
func TestProgramStartsAndExits(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns the binary")
	}
	bin := buildBinary(t)

	t.Run("piped stdout points at the static commands", func(t *testing.T) {
		cmd := exec.Command(bin, "dashboard")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("exit: %v\n%s", err, out)
		}
		got := string(out)
		for _, want := range []string{"needs a terminal", "thermal weekly", "thermal projects", "--json"} {
			if !strings.Contains(got, want) {
				t.Errorf("guard output missing %q:\n%s", want, got)
			}
		}
	})

	t.Run("--json refuses to open a terminal UI", func(t *testing.T) {
		cmd := exec.Command(bin, "dashboard", "--json")
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("dashboard --json should exit non-zero")
		}
		if !strings.Contains(string(out), "does not apply") {
			t.Errorf("--json guard said: %s", out)
		}
	})
}

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "thermal-test")
	cmd := exec.Command("go", "build", "-o", bin, "../../cmd/thermal")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}
	return bin
}
