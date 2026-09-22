// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/thermal"
)

func renderStatsFixture(t *testing.T, width, height int, r Range, m Metric, forceLog bool) string {
	t.Helper()
	m2 := New(projectsFixture(), false)
	m2.tab = tabStats
	m2.rng = r
	m2.metric = m
	m2.statsLog = forceLog
	updated, _ := m2.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return ansi.Strip(updated.(Model).View().Content)
}

// TestStatsParityAgainstStatic proves every headline number in the view comes
// from the same aggregation as `thermal stats`.
func TestStatsParityAgainstStatic(t *testing.T) {
	a := projectsFixture()
	sv := a.BuildStats(Range30d, MetricTokens)

	var days []thermal.DailyRow
	for _, tool := range a.Tools {
		days = append(days, tool.Days...)
	}
	stats := thermal.AggregateStats(days, thermal.StatsOptions{
		Since:  sv.Start,
		Until:  sv.End,
		Metric: string(sv.Metric),
	}, a.Pricer)

	if sv.Days != stats.Days {
		t.Errorf("active days = %d, static = %d", sv.Days, stats.Days)
	}
	if diff := sv.Total - stats.Total; diff > 0.001 || diff < -0.001 {
		t.Errorf("total = %.4f, static = %.4f", sv.Total, stats.Total)
	}
	if diff := sv.Median - stats.Median; diff > 0.001 || diff < -0.001 {
		t.Errorf("median = %.4f, static = %.4f", sv.Median, stats.Median)
	}
	if diff := sv.Max - stats.Max; diff > 0.001 || diff < -0.001 {
		t.Errorf("max = %.4f, static = %.4f", sv.Max, stats.Max)
	}
	if len(sv.TopDays) != len(stats.TopDays) {
		t.Errorf("top days = %d, static = %d", len(sv.TopDays), len(stats.TopDays))
	}
	if len(sv.Outliers) != len(stats.Outliers) {
		t.Errorf("outliers = %d, static = %d", len(sv.Outliers), len(stats.Outliers))
	}
	if diff := sv.Threshold - stats.OutlierThreshold; diff > 0.001 || diff < -0.001 {
		t.Errorf("outlier threshold = %.4f, static = %.4f", sv.Threshold, stats.OutlierThreshold)
	}
}

// TestStatsTrendParity covers the projection, which is the one number a reader
// cannot recompute by eye.
func TestStatsTrendParity(t *testing.T) {
	a := projectsFixture()
	sv := a.BuildStats(Range30d, MetricTokens)

	var days []thermal.DailyRow
	for _, tool := range a.Tools {
		days = append(days, tool.Days...)
	}
	trend := thermal.AggregateTrend(days, thermal.TrendOptions{
		Since:  sv.Start,
		Until:  sv.End,
		Metric: string(sv.Metric),
	}, a.Pricer)

	if diff := sv.Slope - trend.Slope; diff > 1e-6 || diff < -1e-6 {
		t.Errorf("slope = %.6f, static = %.6f", sv.Slope, trend.Slope)
	}
	if (sv.Project == nil) != (trend.Projection == nil) {
		t.Fatalf("projection presence differs: view %v, static %v", sv.Project != nil, trend.Projection != nil)
	}
	if sv.Project != nil {
		if diff := sv.Project.Expected - trend.Projection.Expected; diff > 0.001 || diff < -0.001 {
			t.Errorf("projected = %.4f, static = %.4f", sv.Project.Expected, trend.Projection.Expected)
		}
		if sv.Project.To != trend.Projection.To {
			t.Errorf("projection target = %s, static = %s", sv.Project.To, trend.Projection.To)
		}
		if sv.Project.Low > sv.Project.Expected || sv.Project.High < sv.Project.Expected {
			t.Errorf("expected %.0f sits outside its band %.0f to %.0f",
				sv.Project.Expected, sv.Project.Low, sv.Project.High)
		}
	}
}

// TestStatsLogScaleDecision pins when the histogram switches axes. A maximum
// ten times the median means linear bars would bury the shape in one column.
func TestStatsLogScaleDecision(t *testing.T) {
	a := Adapter{Latest: "2026-09-18", Earliest: "2025-09-20"}
	// Ten ordinary days and one enormous day: a long tail by construction.
	for i := 1; i <= 10; i++ {
		a.Tools = append(a.Tools, ToolDay{
			Name: "Flat",
			Days: []thermal.DailyRow{{Day: dayInWindow(i), Tokens: 100_000, Turns: 1}},
		})
	}
	a.Tools = append(a.Tools, ToolDay{
		Name: "Spike",
		Days: []thermal.DailyRow{{Day: "2026-09-10", Tokens: 100_000_000, Turns: 1}},
	})

	sv := a.BuildStats(Range30d, MetricTokens)
	if !sv.UseLog {
		t.Errorf("expected a log scale when max (%0.f) dwarfs the median (%.0f)", sv.Max, sv.Median)
	}
	if sv.LogReason == "" {
		t.Error("log scale chosen without saying why")
	}
	// A flat fixture must not switch scales.
	flat := Adapter{Latest: "2026-09-18", Earliest: "2025-09-20"}
	for i := 1; i <= 10; i++ {
		flat.Tools = append(flat.Tools, ToolDay{
			Name: "Flat",
			Days: []thermal.DailyRow{{Day: dayInWindow(i), Tokens: 100_000, Turns: 1}},
		})
	}
	if got := flat.BuildStats(Range30d, MetricTokens); got.UseLog {
		t.Error("flat data should not switch to a log scale")
	}
}

func dayInWindow(i int) string {
	// Window is 2026-08-20 to 2026-09-18; spread the days across it.
	day := 10 + i
	if day > 18 {
		day -= 28
	}
	return "2026-09-" + twoDigits(day)
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// TestStatsGolden pins both frames at two sizes, linear and log.
func TestStatsGolden(t *testing.T) {
	for _, tc := range []struct {
		name     string
		w, h     int
		r        Range
		m        Metric
		forceLog bool
		file     string
	}{
		{"80x24", 80, 24, Range30d, MetricTokens, false, "stats_80x24.golden"},
		{"120x40", 120, 40, Range30d, MetricTokens, false, "stats_120x40.golden"},
		{"120x40 cost", 120, 40, Range30d, MetricCost, false, "stats_cost_120x40.golden"},
		{"120x40 log", 120, 40, Range30d, MetricTokens, true, "stats_log_120x40.golden"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderStatsFixture(t, tc.w, tc.h, tc.r, tc.m, tc.forceLog)
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

// TestStatsKeyFlow covers the l key and proves the override survives a tab
// round trip.
func TestStatsKeyFlow(t *testing.T) {
	m := New(projectsFixture(), false)
	m.tab = tabStats
	m = press(t, m, "l")
	if !m.statsLog {
		t.Fatal("l did not turn the log scale on")
	}
	m = press(t, m, "1")
	m = press(t, m, "5")
	if !m.statsLog {
		t.Error("the scale override did not survive a tab round trip")
	}
	m = press(t, m, "l")
	if m.statsLog {
		t.Error("l did not turn the log scale back off")
	}
}

// TestStatsFitsWidth proves the frame never exceeds the terminal at the widths
// the goal names.
func TestStatsFitsWidth(t *testing.T) {
	for _, w := range []int{80, 100, 130, 160, 200} {
		got := renderStatsFixture(t, w, 40, Range30d, MetricTokens, false)
		for i, line := range splitLines(got) {
			if n := len([]rune(line)); n > w {
				t.Errorf("width %d: line %d is %d cells: %q", w, i, n, line)
			}
		}
	}
}

var _ = tea.KeyPressMsg{}
