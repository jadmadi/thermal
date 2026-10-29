// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

// Charts print concrete bar rows under a table. They exist for two reasons: a
// reader can see relative size without comparing columns of digits, and the
// output survives being pasted into a chat or an issue, where a terminal
// heatmap or a TUI does not.
//
// Three rules hold for every chart here:
//   - the number is always printed beside the bar, so colour is never the only
//     encoding and a single-colour terminal loses nothing
//   - bars are scaled against the largest row on screen, and the scale is named
//     in the caption
//   - rows are dropped or shortened when the terminal is narrow, never wrapped

const chartBarMax = 48

// chartRowCap is how many bars a chart prints. A bar chart with eighty rows is
// a second copy of the table, not a chart, so the tail is summarised instead.
const chartRowCap = 12

// chartRowsFor returns how many rows to draw and how many were left out. A
// positive top follows the table's --top limit first, then the chart cap.
func chartRowsFor(n, top int) (int, int) {
	shown := n
	if top > 0 {
		if top < shown {
			shown = top
		}
	} else if shown > chartRowCap {
		shown = chartRowCap
	}
	if shown < 0 {
		shown = 0
	}
	return shown, n - shown
}

// chartRows returns how many rows to draw and how many were left out.
func chartRows(n int) (int, int) {
	return chartRowsFor(n, 0)
}

func chartMetricLabel(v float64, metric string) string {
	if metric == "cost" {
		return formatCost(v)
	}
	return thermal.CompactNumber(int64(v)) + " tok"
}

// chartFooter prints the tail summary and a valid remediation. The suggested
// --top value always refers to the full row count, so raising it widens both
// the table and the chart.
func chartFooter(shown, total, top int, colors bool) string {
	if total <= 0 {
		return ""
	}
	omitted := total - shown
	if omitted <= 0 {
		return ""
	}
	hint := fmt.Sprintf("chart shows %d of %d; use --top %d for all", shown, total, total)
	return fmt.Sprintf("  %s\n", theme.Border.Sprint(colors,
		fmt.Sprintf("… and %d more rows (%s)", omitted, hint)))
}

// chartWidth is the space a bar may occupy. It subtracts the indent, the label
// column, the two gaps, and the value column, because a chart that overshoots
// the terminal wraps and stops being pasteable.
func chartWidth(total, labelW int) int {
	const indent, gaps, value = 2, 4, 12
	w := total - indent - labelW - gaps - value
	if w > chartBarMax {
		w = chartBarMax
	}
	if w < 8 {
		w = 8
	}
	return w
}

// chartBar renders a proportional bar with its scale implied by maxVal.
func chartBar(value, maxVal float64, width int, colors bool) string {
	if maxVal <= 0 || width <= 0 {
		return ""
	}
	filled := int(value / maxVal * float64(width))
	if filled > width {
		filled = width
	}
	if filled == 0 && value > 0 {
		filled = 1
	}
	full := strings.Repeat("█", filled)
	rest := strings.Repeat("·", width-filled)
	if !colors {
		return full + rest
	}
	return theme.Primary.Sprint(true, full) + theme.Border.Sprint(true, rest)
}

// titleCase capitalises the first letter of an ASCII word. strings.Title is
// deprecated and does more than this needs.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// unpriceableNote states the tokens the estimate does not cover, because the
// source names no model for them. Without it an estimated total reads as
// complete when it is not.
func unpriceableNote(tokens int64) string {
	if tokens <= 0 {
		return ""
	}
	return fmt.Sprintf("%s tokens have no model recorded, so no price is applied to them.",
		thermal.CompactNumber(tokens))
}

// truncateLabel shortens a name to fit the label column.
func truncateLabel(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

func chartColors(noColor bool) bool {
	return !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""
}

// RenderProjectChart prints ranked project bars under the projects table. It
// takes the same display names the table printed, so a chart label and a row
// label are never two different strings for one project.
func RenderProjectChart(rep thermal.ProjectReport, names map[string]string, top, width int, noColor bool) string {
	if len(rep.Rows) == 0 {
		return ""
	}
	colors := chartColors(noColor)
	labelW := 24
	barW := chartWidth(width, labelW)

	if names == nil {
		var paths []string
		for _, r := range rep.Rows {
			paths = append(paths, r.Project)
		}
		names = thermal.ProjectDisplayNames(paths)
	}

	shown, _ := chartRowsFor(len(rep.Rows), top)
	if shown == 0 {
		return ""
	}
	maxVal := float64(rep.Rows[0].Tokens)
	for _, row := range rep.Rows[1:shown] {
		if float64(row.Tokens) > maxVal {
			maxVal = float64(row.Tokens)
		}
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  %s\n",
		theme.Text.SprintBold(colors, "Projects by tokens"),
		theme.TextMuted.Sprint(colors, fmt.Sprintf("bar = share of the largest charted row, %s tok", thermal.CompactNumber(int64(maxVal))))))
	b.WriteString("\n")

	for _, row := range rep.Rows[:shown] {
		label := row.Project
		if display, ok := names[label]; ok && display != "" {
			label = display
		}
		name := truncateLabel(label, labelW)
		b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
			thermal.PadRight(name, labelW),
			chartBar(float64(row.Tokens), maxVal, barW, colors),
			thermal.PadLeft(thermal.CompactNumber(row.Tokens)+" tok", 10)))
	}
	b.WriteString(chartFooter(shown, len(rep.Rows), top, colors))
	return b.String()
}

// RenderModelChart prints ranked model bars under the models table.
func RenderModelChart(rep thermal.ModelReport, top, width int, noColor bool) string {
	if len(rep.Rows) == 0 {
		return ""
	}
	colors := chartColors(noColor)
	labelW := 28
	barW := chartWidth(width, labelW)

	shown, _ := chartRowsFor(len(rep.Rows), top)
	if shown == 0 {
		return ""
	}
	maxVal := float64(rep.Rows[0].Tokens)
	for _, row := range rep.Rows[1:shown] {
		if float64(row.Tokens) > maxVal {
			maxVal = float64(row.Tokens)
		}
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  %s\n",
		theme.Text.SprintBold(colors, "Models by tokens"),
		theme.TextMuted.Sprint(colors, fmt.Sprintf("bar = share of the largest charted row, %s tok; ~ = estimated cost", thermal.CompactNumber(int64(maxVal))))))
	b.WriteString("\n")

	for _, row := range rep.Rows[:shown] {
		cost := formatCostOrDashValue(row.Cost, true)
		b.WriteString(fmt.Sprintf("  %s  %s  %s %s\n",
			thermal.PadRight(truncateLabel(row.Model, labelW), labelW),
			chartBar(float64(row.Tokens), maxVal, barW, colors),
			thermal.PadLeft(thermal.CompactNumber(row.Tokens), 8),
			thermal.PadLeft(cost, 9)))
	}
	b.WriteString(chartFooter(shown, len(rep.Rows), top, colors))
	return b.String()
}

// RenderPeriodChart prints bars for token or cost totals under a period report.
// totalRows is the full window row count before the table's --top cap, so the
// footer can suggest a valid widening value.
func RenderPeriodChart(rep thermal.Report, metric string, width int, noColor bool, top, totalRows int) string {
	if len(rep.Rows) == 0 {
		return ""
	}
	colors := chartColors(noColor)

	values := make([]float64, len(rep.Rows))
	for i, row := range rep.Rows {
		if metric == "cost" {
			values[i] = row.Cost
		} else {
			values[i] = float64(row.Tokens)
		}
	}

	// The chart follows the table's own order, so a reader never has to work
	// out whether the bars are in a different sequence from the rows above.
	shown, _ := chartRowsFor(len(rep.Rows), top)
	if shown == 0 {
		return ""
	}
	var maxVal float64
	for _, value := range values[:shown] {
		if value > maxVal {
			maxVal = value
		}
	}
	if maxVal == 0 {
		return ""
	}

	labelW := 12
	barW := chartWidth(width, labelW)
	title := fmt.Sprintf("%s by period", titleCase(metric))
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s  %s\n",
		theme.Text.SprintBold(colors, title),
		theme.TextMuted.Sprint(colors, fmt.Sprintf("bar = share of the largest charted period, %s", chartMetricLabel(maxVal, metric)))))
	b.WriteString("\n")

	for i := 0; i < shown; i++ {
		row := rep.Rows[i]
		label := row.Period
		if label == "" {
			continue
		}
		value := thermal.CompactNumber(row.Tokens) + " tok"
		if metric == "cost" {
			value = formatCostOrDashValue(row.Cost, row.EstimatedCost > 0)
		}
		b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
			thermal.PadRight(truncateLabel(label, labelW), labelW),
			chartBar(values[i], maxVal, barW, colors),
			thermal.PadLeft(value, 10)))
	}
	if totalRows <= 0 {
		totalRows = len(rep.Rows)
	}
	b.WriteString(chartFooter(shown, totalRows, top, colors))
	return b.String()
}
