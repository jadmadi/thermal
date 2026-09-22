// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

// renderMix draws the Mix tab: a stacked chart over time, a legend that carries
// the numbers, and a switching panel.
func renderMix(mv MixView, width, height int, p Palette) string {
	right := fmt.Sprintf("by %s · %s · %s to %s", mv.By, mv.Grain, mv.Start, mv.End)
	title := spread(p.Emphasis.Render("Mix"), p.Muted.Render(right), width)

	if len(mv.Series) == 0 || len(mv.Buckets) == 0 {
		return title + "\n\n" + p.Muted.Render("No activity in this window.")
	}
	note := ""
	if mv.Estimated {
		note = p.Muted.Render("cost is estimated from pricing data for days with no recorded cost")
	}
	// Priority order: the chart, then the legend that explains it, then the
	// switching panel, then the footnote.
	return fitBlocks(height-4,
		title,
		renderStacked(mv, width, p),
		mixLegend(mv, width, p),
		switchingPanel(mv, width, p),
		note,
	)
}

// renderStacked draws one column per bucket. Each column is filled from the
// bottom up in series order, so the tallest series sits on the baseline. The
// rune identifies the series and the total above the chart carries the number,
// which is what keeps the picture readable without colour.
func renderStacked(mv MixView, width int, p Palette) string {
	const rows = 8
	cols := len(mv.Buckets)
	if cols > width-2 {
		cols = width - 2
	}
	if cols <= 0 {
		return ""
	}
	// Keep the newest buckets when the window does not fit.
	startAt := len(mv.Buckets) - cols

	var maxTotal float64
	for _, b := range mv.Buckets {
		if b.Total > maxTotal {
			maxTotal = b.Total
		}
	}
	if maxTotal <= 0 {
		return p.Muted.Render("no activity in this window")
	}

	// Build the grid bottom-up, then print top-down.
	grid := make([][]string, rows)
	for r := range grid {
		grid[r] = make([]string, cols)
		for c := range grid[r] {
			grid[r][c] = " "
		}
	}

	for c := 0; c < cols; c++ {
		bucket := mv.Buckets[startAt+c]
		height := int(bucket.Total / maxTotal * float64(rows))
		if height == 0 && bucket.Total > 0 {
			height = 1
		}
		if height > rows {
			height = rows
		}
		// Distribute the column's cells across series by their share of this
		// bucket, largest first, so the stacking order is stable.
		remaining := height
		filled := 0
		for si := range mv.Series {
			if remaining == 0 {
				break
			}
			if si >= len(bucket.Values) || bucket.Values[si] <= 0 {
				continue
			}
			cells := 1
			if bucket.Total > 0 {
				cells = int(bucket.Values[si] / bucket.Total * float64(height))
			}
			if cells > remaining {
				cells = remaining
			}
			glyph := "▪"
			if si < len(seriesRunes) {
				glyph = seriesRunes[si]
			}
			for i := 0; i < cells; i++ {
				grid[rows-1-filled][c] = glyph
				filled++
			}
			remaining -= cells
		}
		// Any cell left over is the largest series, so the column always
		// reaches its computed height.
		glyph := "▪"
		if len(mv.Series) > 0 && len(seriesRunes) > 0 {
			glyph = seriesRunes[0]
		}
		for remaining > 0 {
			grid[rows-1-filled][c] = glyph
			filled++
			remaining--
		}
	}

	var b strings.Builder
	for r := 0; r < rows; r++ {
		line := strings.Join(grid[r], "")
		if p.Colorful && r == rows-1 {
			line = p.Accent.Render(line)
		}
		b.WriteString(strings.TrimRight(line, " "))
		b.WriteString("\n")
	}
	// Axis labels: first and last bucket, plus the window total.
	first := mv.Buckets[startAt].Period
	last := mv.Buckets[len(mv.Buckets)-1].Period
	b.WriteString(p.Dim.Render(fmt.Sprintf("%-8s%s", first, last)))
	b.WriteString("\n")
	b.WriteString(p.Muted.Render(fmt.Sprintf("%s %s per bucket at peak",
		thermal.CompactNumber(int64(maxTotal)), metricUnit(mv.Metric))))
	return strings.TrimRight(b.String(), "\n")
}

func metricUnit(m Metric) string {
	if m == MetricCost {
		return "dollars"
	}
	return "tokens"
}

// mixLegend prints the series, their glyph, and their share of the window. The
// numbers are the source of truth; the glyph just ties a colour to a name.
func mixLegend(mv MixView, width int, p Palette) string {
	var b strings.Builder
	b.WriteString(p.Muted.Render(fmt.Sprintf("Series, %s of the window", string(mv.Metric))))
	b.WriteString("\n")

	nameW := 20
	if width < 80 {
		nameW = 14
	}
	barW := width - nameW - colShare - 8
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 40 {
		barW = 40
	}
	for i, s := range mv.Series {
		glyph := "▪"
		if i < len(seriesRunes) {
			glyph = seriesRunes[i]
		}
		if p.Colorful {
			glyph = p.Accent.Render(glyph)
		}
		b.WriteString("  " + glyph + " " +
			padTo(truncateRunes(s.Name, nameW), nameW) + "  " +
			p.bar(int64(s.Total), int64(mv.Series[0].Total), barW) + "  " +
			padLeft(fmt.Sprintf("%.1f%%", s.Share), colShare) + "\n")
	}
	b.WriteString("  " + p.Dim.Render(fmt.Sprintf("%d active series", mv.Active)))
	return strings.TrimRight(b.String(), "\n")
}

// switchingPanel answers who leads and how often that changed, which is the
// question the mix command exists for.
func switchingPanel(mv MixView, width int, p Palette) string {
	var b strings.Builder
	b.WriteString(p.Muted.Render("Switching"))
	b.WriteString("\n")

	dominant := mv.Dominant
	if dominant == "" {
		dominant = "—"
	}
	b.WriteString("  " + padTo("dominant", 14) + p.Emphasis.Render(dominant) + "\n")
	b.WriteString("  " + padTo("concentration", 14) +
		p.Emphasis.Render(fmt.Sprintf("%.3f", mv.Conc)) + "  " +
		p.Muted.Render(concNote(mv.Conc)) + "\n")
	b.WriteString("  " + padTo("switches", 14) + p.Emphasis.Render(fmt.Sprintf("%d", mv.Switches)) + "\n")

	if len(mv.SwitchesMo) > 0 {
		months := make([]string, 0, len(mv.SwitchesMo))
		for month := range mv.SwitchesMo {
			months = append(months, month)
		}
		sortStrings(months)
		var parts []string
		for _, month := range months {
			parts = append(parts, fmt.Sprintf("%s %d", month, mv.SwitchesMo[month]))
		}
		b.WriteString("  " + padTo("by month", 14) + p.Muted.Render(strings.Join(parts, " · ")) + "\n")
	}

	if width > 100 && len(mv.Buckets) > 0 {
		// A compact per-bucket total line, so a reader can see the shape of the
		// window without reading the chart row by row.
		values := make([]int64, 0, len(mv.Buckets))
		for _, bucket := range mv.Buckets {
			values = append(values, int64(bucket.Total))
		}
		b.WriteString("\n  " + p.Muted.Render("per bucket") + "  " + p.spark(values))
	}
	return strings.TrimRight(b.String(), "\n")
}

func concNote(conc float64) string {
	switch {
	case conc >= 0.85:
		return "one series dominates"
	case conc >= 0.5:
		return "a few series share the work"
	default:
		return "spread evenly across series"
	}
}

// renderModels draws the ranked model table with a share strip and the tool
// split. Like the projects table it is windowed, because a machine with sixty
// models would otherwise push the cursor off screen.
func renderModels(mv ModelsView, width, height, cursor int, p Palette) string {
	var b strings.Builder

	right := fmt.Sprintf("estimated cost · %s to %s", mv.Start, mv.End)
	b.WriteString(spread(p.Emphasis.Render("Models"), p.Muted.Render(right), width))
	b.WriteString("\n\n")

	if len(mv.Rows) == 0 {
		b.WriteString(p.Muted.Render("No model attribution in this window."))
		return b.String()
	}

	type optional struct {
		render func(int) string
		drop   int
	}
	var opts []optional
	opts = append(opts, optional{render: func(int) string { return padLeft("Days", pcolDays) }, drop: pcolDays})
	opts = append(opts, optional{render: func(int) string { return padTo("Tools", 16) }, drop: 16})
	opts = append(opts, optional{render: func(barW int) string { return padLeft("Tokens", barW) }, drop: maxBarW})

	base := []string{padTo("#", pcolRank), padTo("Model", 30), padLeft("Tokens", pcolTokens), padLeft("Cost", pcolCost), padLeft("Share", colShare)}
	layout := func(include, barW int) string {
		cells := append([]string{}, base...)
		for i := 0; i < include; i++ {
			cells = append(cells, opts[i].render(barW))
		}
		return strings.Join(cells, tableGap)
	}

	include := len(opts)
	barW := maxBarW
	header := layout(include, barW)
	for lipglossWidth(header) > width && include > 0 {
		include--
		barW += opts[include].drop
		if include == 0 {
			barW = 0
		}
		header = layout(include, barW)
	}
	if slack := width - lipglossWidth(header); include >= 2 && slack > 0 {
		barW += slack
		if barW > maxBarW {
			barW = maxBarW
		}
		header = layout(include, barW)
	}

	b.WriteString(p.Muted.Render(header))
	b.WriteString("\n")
	b.WriteString(p.Dim.Render(strings.Repeat("─", minInt(width, lipglossWidth(header)))))
	b.WriteString("\n")

	// Chrome: tab bar and blank, title and blank, header and rule, the rows,
	// then a rule, the total, the coverage note, the tool block and the hint.
	// The tool block is a fixed three lines below, never one line per tool,
	// because a long tool list would otherwise squeeze the table to a single
	// row and make the cursor useless.
	const minRows = 6
	chrome := 16
	if mv.CovNote != "" {
		chrome++
	}
	if len(mv.Tools) > 0 {
		chrome += 3
	}
	available := height - chrome
	showExtra := true
	if available < minRows {
		// Shed the optional blocks before the table: a table with one row is
		// worse than a missing footnote. showExtra keeps the renderer below in
		// step with the budget, so a dropped block is also not printed.
		showExtra = false
		available = height - 16
		if available < minRows {
			available = minRows
		}
	}
	first, last := viewportWindow(len(mv.Rows), cursor, available)

	for i := first; i < last; i++ {
		row := mv.Rows[i]
		cells := []string{
			padTo(fmt.Sprintf("%d.", i+1), pcolRank),
			padTo(truncateRunes(row.Name, 30), 30),
			padLeft(thermal.CompactNumber(row.Tokens), pcolTokens),
			padLeft(Money(row.Cost), pcolCost),
			padLeft(fmt.Sprintf("%.1f%%", row.Share), colShare),
		}
		if include >= 1 {
			cells = append(cells, padLeft(fmt.Sprintf("%d", row.Days), pcolDays))
		}
		if include >= 2 {
			cells = append(cells, padTo(truncateRunes(joinTools(row.Tools, 2), 16), 16))
		}
		if include >= 3 {
			cells = append(cells, p.bar(row.Tokens, mv.Rows[0].Tokens, barW))
		}
		b.WriteString(strings.Join(cells, tableGap))
		b.WriteString("\n")
	}

	b.WriteString(p.Dim.Render(strings.Repeat("─", minInt(width, lipglossWidth(header)))))
	b.WriteString("\n")
	if hidden := len(mv.Rows) - (last - first); hidden > 0 {
		var parts []string
		if first > 0 {
			parts = append(parts, fmt.Sprintf("%d above", first))
		}
		if below := len(mv.Rows) - last; below > 0 {
			parts = append(parts, fmt.Sprintf("%d below", below))
		}
		b.WriteString(p.Dim.Render(fmt.Sprintf("  rows %d-%d of %d (%s)",
			first+1, last, len(mv.Rows), strings.Join(parts, ", "))))
		b.WriteString("\n")
	}
	b.WriteString(p.Emphasis.Render(padTo("Total", pcolRank+30)) +
		padLeft(thermal.CompactNumber(mv.Tokens), pcolTokens) + tableGap +
		padLeft(Money(mv.Cost), pcolCost))
	b.WriteString("\n")

	if showExtra && mv.CovNote != "" {
		b.WriteString(p.Muted.Render(mv.CovNote))
		b.WriteString("\n")
	}
	if showExtra && len(mv.Tools) > 0 {
		// One line, top four tools, remainder counted: three lines total,
		// whatever the machine has installed.
		b.WriteString("\n")
		b.WriteString(p.Muted.Render("By tool") + "  ")
		const shown = 4
		var parts []string
		for i, t := range mv.Tools {
			if i == shown {
				parts = append(parts, fmt.Sprintf("+%d more", len(mv.Tools)-shown))
				break
			}
			parts = append(parts, fmt.Sprintf("%s %.1f%%", t.Name, t.Pct))
		}
		b.WriteString(p.Muted.Render(strings.Join(parts, " · ")))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func sortStrings(list []string) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j] < list[j-1]; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}
