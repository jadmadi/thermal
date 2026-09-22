// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	projectWidth = 28
	toolsWidth   = 20
	daysWidth    = 4
	lastWidth    = 10
	rankWidth    = 3
)

// RenderProjects prints a project leaderboard. top limits the printed rows;
// zero means print every row. The full project paths stay in --json output.
func RenderProjects(rep thermal.ProjectReport, top int, noColor bool) string {
	return renderProjects(rep, top, noColor, false)
}

// RenderProjectsBreakdown adds a tool split and the top models under every
// project row.
func RenderProjectsBreakdown(rep thermal.ProjectReport, top int, noColor bool) string {
	return renderProjects(rep, top, noColor, true)
}

func renderProjects(rep thermal.ProjectReport, top int, noColor bool, breakdown bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("projects"),
	))

	if len(rep.Rows) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No project activity in the selected window.")))
		return sb.String()
	}

	names := displayNames(projectPaths(rep.Rows))

	headers := []string{"#", "Project", "Tools", "Tokens", "Cost", "Days", "Last"}
	alignRight := []bool{false, false, false, true, true, true, false}
	widths := []int{rankWidth, projectWidth, toolsWidth, numberWidth, numberWidth, daysWidth, lastWidth}

	sb.WriteString("  ")
	for i, h := range headers {
		cell := thermal.PadRight(h, widths[i])
		if alignRight[i] {
			cell = thermal.PadLeft(h, widths[i])
		}
		sb.WriteString(dim(cell))
		if i < len(headers)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}
	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")

	shown := len(rep.Rows)
	if top > 0 && top < shown {
		shown = top
	}
	for i := 0; i < shown; i++ {
		row := rep.Rows[i]
		printProjectRow(&sb, i+1, row, names[row.Project], widths, nil)
		if breakdown {
			printProjectBreakdown(&sb, row, dim)
		}
	}
	if rest := len(rep.Rows) - shown; rest > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest))))
	}

	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")
	totals := rep.Totals
	sb.WriteString("  ")
	cells := []string{
		thermal.PadRight("", rankWidth),
		thermal.PadRight("Total", projectWidth),
		thermal.PadRight(toolsCell(totals.Tools, toolsWidth), toolsWidth),
		thermal.PadLeft(thermal.CompactNumber(totals.Tokens), numberWidth),
		thermal.PadLeft(formatCostOrDash(totals.Cost), numberWidth),
		thermal.PadLeft(fmt.Sprintf("%d", totals.ActiveDays), daysWidth),
		thermal.PadRight(totals.LastDay, lastWidth),
	}
	for i, c := range cells {
		sb.WriteString(gold(c))
		if i < len(cells)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	if totals.EstimatedCost > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim(fmt.Sprintf(
			"Total = %s recorded + ~%s estimated from pricing data.",
			formatCost(totals.StoredCost), formatCost(totals.EstimatedCost)))))
	}
	if note := unpriceableNote(totals.UnattributedTokens); note != "" {
		sb.WriteString("  " + dim(note) + "\n")
	}
	if len(totals.MissingPricing) > 0 {
		models := totals.MissingPricing
		if len(models) > 4 {
			models = append(append([]string{}, models[:4]...), fmt.Sprintf("+%d more", len(models)-4))
		}
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No pricing for: "+strings.Join(models, ", "))))
	}

	sb.WriteString("\n")
	return sb.String()
}

func printProjectRow(sb *strings.Builder, rank int, row thermal.ProjectRow, name string, widths []int, style func(string) string) {
	cells := []string{
		thermal.PadLeft(fmt.Sprintf("%d.", rank), widths[0]),
		thermal.PadRight(truncate(name, widths[1]), widths[1]),
		thermal.PadRight(toolsCell(row.Tools, widths[2]), widths[2]),
		thermal.PadLeft(thermal.CompactNumber(row.Tokens), widths[3]),
		thermal.PadLeft(formatCostOrDash(row.Cost), widths[4]),
		thermal.PadLeft(fmt.Sprintf("%d", row.ActiveDays), widths[5]),
		thermal.PadRight(row.LastDay, widths[6]),
	}
	sb.WriteString("  ")
	for i, c := range cells {
		if style != nil {
			c = style(c)
		}
		sb.WriteString(c)
		if i < len(cells)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")
}

func formatCostOrDash(v float64) string {
	if v <= 0 {
		return "—"
	}
	return formatCost(v)
}

// displayNames maps each path to a short label. A lone repository shows as its
// own directory name, mahak-bench. When two paths share a name, the
// distinguishing parent segment is added in parentheses, mahak-bench (Jad).
// The full path stays in the JSON output.
func displayNames(paths []string) map[string]string {
	return thermal.ProjectDisplayNames(paths)
}

func projectPaths(rows []thermal.ProjectRow) []string {
	paths := make([]string, len(rows))
	for i, row := range rows {
		paths[i] = row.Project
	}
	return paths
}

// printProjectBreakdown writes the tool split and top models under one project
// row. Both lines show token totals rather than shares, because some tools
// record no model attribution, which would make percentages misleading.
func printProjectBreakdown(sb *strings.Builder, row thermal.ProjectRow, dim func(string) string) {
	if len(row.Tools) > 1 {
		sb.WriteString("        " + dim("tools   "+weightLine(row.Tools, row.ToolTokens, 3)) + "\n")
	}
	if len(row.Models) > 1 {
		names := make([]string, 0, len(row.Models))
		weights := make(map[string]int64, len(row.Models))
		for name, counts := range row.Models {
			names = append(names, name)
			weights[name] = counts.Total()
		}
		short := displayNames(names)
		labels := make([]string, len(names))
		for i, name := range names {
			labels[i] = short[name]
		}
		// Re-key the weights under the short labels so the line can print them.
		shortWeights := make(map[string]int64, len(names))
		for i, name := range names {
			shortWeights[labels[i]] = weights[name]
		}
		sb.WriteString("        " + dim("models  "+weightLine(labels, shortWeights, 3)) + "\n")
	}
}

// weightLine renders "name tokens" entries, largest first, capped at limit.
func weightLine(names []string, weights map[string]int64, limit int) string {
	sorted := append([]string{}, names...)
	sort.Slice(sorted, func(i, j int) bool {
		if weights[sorted[i]] != weights[sorted[j]] {
			return weights[sorted[i]] > weights[sorted[j]]
		}
		return sorted[i] < sorted[j]
	})

	shown := sorted
	if limit > 0 && len(shown) > limit {
		shown = shown[:limit]
	}
	parts := make([]string, 0, len(shown))
	for _, name := range shown {
		parts = append(parts, fmt.Sprintf("%s %s", name, thermal.CompactNumber(weights[name])))
	}
	line := strings.Join(parts, " · ")
	if rest := len(sorted) - len(shown); rest > 0 {
		line += fmt.Sprintf(" · +%d", rest)
	}
	return line
}

// toolsCell renders contributing tool names in a fixed-width cell, dropping
// the tail and adding a count when the list does not fit.
func toolsCell(tools []string, width int) string {
	if len(tools) == 0 {
		return "—"
	}
	shown := make([]string, 0, len(tools))
	for i, t := range tools {
		suffix := ""
		if i < len(tools)-1 {
			suffix = fmt.Sprintf(" +%d", len(tools)-i-1)
		}
		candidate := strings.Join(append(append([]string{}, shown...), t), ",") + suffix
		if runeLen(candidate) > width {
			if len(shown) == 0 {
				shown = append(shown, truncate(t, width))
			}
			break
		}
		shown = append(shown, t)
	}
	cell := strings.Join(shown, ",")
	if used := len(shown); used < len(tools) {
		cell += fmt.Sprintf(" +%d", len(tools)-used)
	}
	return cell
}
