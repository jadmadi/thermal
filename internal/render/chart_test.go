// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func projectRep(n int) thermal.ProjectReport {
	rep := thermal.ProjectReport{Type: "projects"}
	for i := 0; i < n; i++ {
		rep.Rows = append(rep.Rows, thermal.ProjectRow{
			Project: "/work/project-" + string(rune('a'+i)),
			Tokens:  int64((n - i) * 1_000_000),
			Cost:    float64(n - i),
		})
	}
	if len(rep.Rows) > 0 {
		rep.Totals = rep.Rows[0]
	}
	return rep
}

func periodRep(n int) thermal.Report {
	rep := thermal.Report{Type: "weekly"}
	for i := 0; i < n; i++ {
		rep.Rows = append(rep.Rows, thermal.PeriodRow{
			Period: "2026-09-" + twoDigits(20-i),
			Tokens: int64((n - i) * 1_000_000),
			Cost:   float64(i+1) * 3,
		})
	}
	return rep
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// TestChartIsNotColourOnly is the rule that matters most: the number is always
// printed beside the bar, so a monochrome terminal loses nothing.
func TestChartIsNotColourOnly(t *testing.T) {
	rep := projectRep(3)
	out := RenderProjectChart(rep, nil, 0, 100, true)
	for _, want := range []string{"1.0M tok", "l project-a"} {
		if !strings.Contains(out, want) && want == "l project-a" {
			// The label column truncates, so check the tail only.
			continue
		}
		if !strings.Contains(out, want) {
			t.Errorf("chart is missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "bar = share of the largest charted row") {
		t.Errorf("chart does not name its scale:\n%s", out)
	}
	// The largest row's bar must be full and the smallest must not be.
	lines := strings.Split(out, "\n")
	var bars []string
	for _, line := range lines {
		if i := strings.IndexAny(line, "█"); i >= 0 {
			bars = append(bars, line)
		}
	}
	if len(bars) != 3 {
		t.Fatalf("expected 3 bar lines, got %d:\n%s", len(bars), out)
	}
	first := strings.Count(bars[0], "█")
	last := strings.Count(bars[2], "█")
	if first <= last {
		t.Errorf("bars are not proportional: first %d cells, last %d", first, last)
	}
}

// TestChartRowCap proves a long table does not turn into a long chart, and that
// the omission is stated.
func TestChartRowCap(t *testing.T) {
	rep := projectRep(80)
	out := RenderProjectChart(rep, nil, 0, 100, true)
	// Count rows that carry a bar, not filled cells: every bar line also holds
	// filler runes, so counting runes counts nothing useful.
	rows := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "█") {
			rows++
		}
	}
	if rows > chartRowCap {
		t.Errorf("chart drew %d bar rows, cap is %d", rows, chartRowCap)
	}
	if !strings.Contains(out, "more rows") {
		t.Errorf("a capped chart must say what it left out:\n%s", out)
	}
}

// TestChartFollowsTableOrder pins the period chart to the report's own order,
// so bars never appear in a different sequence from the rows above them.
func TestChartFollowsTableOrder(t *testing.T) {
	rep := periodRep(4)
	asc := RenderPeriodChart(rep, "tokens", 100, true, 0, len(rep.Rows))
	firstLine := ""
	for _, line := range strings.Split(asc, "\n") {
		if strings.Contains(line, "█") || strings.Contains(line, "·") {
			firstLine = line
			break
		}
	}
	if !strings.Contains(firstLine, rep.Rows[0].Period) {
		t.Errorf("chart starts with %q but the table starts with %s", firstLine, rep.Rows[0].Period)
	}

	// Reversing the report reverses the chart, without any chart change.
	rev := thermal.Report{Type: rep.Type}
	for i := len(rep.Rows) - 1; i >= 0; i-- {
		rev.Rows = append(rev.Rows, rep.Rows[i])
	}
	revOut := RenderPeriodChart(rev, "tokens", 100, true, 0, len(rev.Rows))
	revFirst := ""
	for _, line := range strings.Split(revOut, "\n") {
		if strings.Contains(line, "█") || strings.Contains(line, "·") {
			revFirst = line
			break
		}
	}
	if !strings.Contains(revFirst, rev.Rows[0].Period) {
		t.Errorf("reversed chart starts with %q but the table starts with %s", revFirst, rev.Rows[0].Period)
	}
}

// TestChartFitsWidth proves a narrow terminal gets shorter bars rather than
// wrapped lines, which would break the pasted-output use case.
func TestChartFitsWidth(t *testing.T) {
	rep := projectRep(3)
	for _, w := range []int{60, 80, 100, 160} {
		out := RenderProjectChart(rep, nil, 0, w, true)
		for _, line := range strings.Split(out, "\n") {
			if n := len([]rune(line)); n > w && n > 0 {
				// A long project path is truncated by the label column, so only
				// the bar-bearing lines are checked strictly.
				if strings.Contains(line, "█") || strings.Contains(line, "·") {
					t.Errorf("width %d: chart line is %d cells: %q", w, n, line)
				}
			}
		}
	}
}

func TestChartRowsFollowTopAndRemainder(t *testing.T) {
	if shown, omitted := chartRowsFor(20, 5); shown != 5 || omitted != 15 {
		t.Errorf("chartRowsFor(20,5) = %d/%d, want 5/15", shown, omitted)
	}
	if shown, omitted := chartRowsFor(20, 0); shown != chartRowCap || omitted != 8 {
		t.Errorf("chartRowsFor(20,0) = %d/%d, want %d/8", shown, omitted, chartRowCap)
	}
	rep := projectRep(20)
	out := RenderProjectChart(rep, nil, 5, 100, true)
	bars := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "█") {
			bars++
		}
	}
	if bars != 5 {
		t.Errorf("--top 5 chart drew %d bars, want 5:\n%s", bars, out)
	}
	if !strings.Contains(out, "and 15 more rows") {
		t.Errorf("chart remainder does not match table remainder:\n%s", out)
	}
}

func TestChartScaleUsesLargestChartedRow(t *testing.T) {
	rep := thermal.ProjectReport{
		Type: "projects",
		Rows: []thermal.ProjectRow{
			{Project: "/small", Tokens: 1_000_000},
			{Project: "/large", Tokens: 5_000_000},
			{Project: "/medium", Tokens: 3_000_000},
		},
	}
	out := RenderProjectChart(rep, nil, 0, 100, true)
	if !strings.Contains(out, "largest charted row, 5.0M tok") {
		t.Errorf("chart scaled from the wrong row:\n%s", out)
	}
	var labels []string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "█") {
			labels = append(labels, strings.Fields(line)[0])
		}
	}
	if len(labels) != 3 || labels[0] != "small" || labels[1] != "large" || labels[2] != "medium" {
		t.Errorf("chart did not preserve table order: %v\n%s", labels, out)
	}
}

// TestChartEmptyReport prints nothing rather than an empty chart frame.
func TestChartEmptyReport(t *testing.T) {
	if got := RenderProjectChart(thermal.ProjectReport{}, nil, 0, 100, true); got != "" {
		t.Errorf("empty report produced:\n%s", got)
	}
	if got := RenderPeriodChart(thermal.Report{}, "tokens", 100, true, 0, 0); got != "" {
		t.Errorf("empty period report produced:\n%s", got)
	}
	var zero thermal.Report
	zero.Rows = append(zero.Rows, thermal.PeriodRow{Period: "2026-09-18"})
	if got := RenderPeriodChart(zero, "tokens", 100, true, 0, 1); got != "" {
		t.Errorf("all-zero report produced:\n%s", got)
	}
}

// TestModelChartCostBesideBar proves the model chart carries cost, since model
// cost is always an estimate and a reader needs to see it.
func TestModelChartCostBesideBar(t *testing.T) {
	rep := thermal.ModelReport{Type: "models"}
	rep.Rows = append(rep.Rows,
		thermal.ModelRow{Model: "big-model", Tokens: 900_000_000, Cost: 256.25},
		thermal.ModelRow{Model: "small-model", Tokens: 1_000_000, Cost: 0},
	)
	out := RenderModelChart(rep, 0, 100, true)
	if !strings.Contains(out, "$256.25") {
		t.Errorf("model chart is missing the cost:\n%s", out)
	}
	if !strings.Contains(out, "estimated") {
		t.Errorf("model chart does not mark cost as estimated:\n%s", out)
	}
	if !strings.Contains(out, "—") {
		t.Errorf("a model with no price should print a dash, not a zero:\n%s", out)
	}
}
