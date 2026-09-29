// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
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
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	gold := st.Gold

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

	shown := len(rep.Rows)
	if top > 0 && top < shown {
		shown = top
	}

	rWidth := rankWidth
	if shown >= 100 {
		rWidth = 4
	}
	if shown >= 1000 {
		rWidth = 5
	}

	headers := []string{"#", "Project", "Tools", "Tokens", "Cost", "Days", "Last"}
	alignRight := []bool{true, false, false, true, true, true, false}
	widths := []int{rWidth, projectWidth, toolsWidth, numberWidth, numberWidth, daysWidth, lastWidth}

	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}

	var cardLines []string

	cardLines = append(cardLines, formatHeaderRow(headers, widths, alignRight, colors))
	cardLines = append(cardLines, " "+tableRule(widths))

	breakdownIndent := 1 + rWidth + 2 + 1
	for i := 0; i < shown; i++ {
		row := rep.Rows[i]
		cardLines = append(cardLines, formatProjectRow(i+1, row, names[row.Project], widths, colors, st))
		if breakdown {
			for _, bl := range formatProjectBreakdown(row, breakdownIndent, colors, st) {
				cardLines = append(cardLines, bl)
			}
		}
	}
	if rest := len(rep.Rows) - shown; rest > 0 {
		cardLines = append(cardLines, " "+dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest)))
	}

	cardLines = append(cardLines, " "+strings.Repeat("─", rule))
	totals := rep.Totals
	cells := []string{
		padRight("", rWidth),
		padRight("Total", projectWidth),
		padRight(toolsCell(totals.Tools, toolsWidth), toolsWidth),
		padLeft(thermal.CompactNumber(totals.Tokens), numberWidth),
		padLeft(formatCostOrDash(totals.Cost), numberWidth),
		padLeft(fmt.Sprintf("%d", totals.ActiveDays), daysWidth),
		padRight(totals.LastDay, lastWidth),
	}
	var totalSb strings.Builder
	totalSb.WriteString(" ")
	for i, c := range cells {
		totalSb.WriteString(gold(c))
		if i < len(cells)-1 {
			totalSb.WriteString("  ")
		}
	}
	cardLines = append(cardLines, totalSb.String())

	sb.WriteString(RenderCard(CardOptions{
		Title:       "Projects",
		RightHeader: "attribution & spend",
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))

	footLimit := BoundedCardWidth(2)
	if totals.EstimatedCost > 0 {
		note := fmt.Sprintf("Total = %s recorded + ~%s estimated from pricing data.", formatCost(totals.StoredCost), formatCost(totals.EstimatedCost))
		for _, wl := range WrapText(note, footLimit) {
			sb.WriteString(fmt.Sprintf("  %s\n", dim(wl)))
		}
	}
	if note := unpriceableNote(totals.UnattributedTokens); note != "" {
		for _, wl := range WrapText(note, footLimit) {
			sb.WriteString(fmt.Sprintf("  %s\n", dim(wl)))
		}
	}
	if len(totals.MissingPricing) > 0 {
		models := totals.MissingPricing
		if len(models) > 4 {
			models = append(append([]string{}, models[:4]...), fmt.Sprintf("+%d more", len(models)-4))
		}
		for _, wl := range WrapText("No pricing for: "+strings.Join(models, ", "), footLimit) {
			sb.WriteString(fmt.Sprintf("  %s\n", dim(wl)))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

func formatProjectRow(rank int, row thermal.ProjectRow, name string, widths []int, colors bool, st Style) string {
	rWidth := widths[0]
	pWidth := widths[1]
	tWidth := widths[2]
	tokWidth := widths[3]
	costWidth := widths[4]
	daysWidth := widths[5]
	lastWidth := widths[6]

	rankCell := formatRankCell(rank, rWidth, colors)

	// 2. Project Name
	truncName := truncate(name, pWidth)
	var projCell string
	if colors {
		if row.Tokens == 0 {
			projCell = padRightStyled(truncName, pWidth, func(s string) string { return theme.TextMuted.Sprint(true, s) })
		} else if rank == 1 {
			projCell = padRightStyled(truncName, pWidth, func(s string) string { return theme.Text.SprintBold(true, s) })
		} else {
			// Subtle dimming for parent disambiguation hint like " (aqaba-dev)"
			if idx := strings.Index(truncName, " ("); idx != -1 && strings.HasSuffix(truncName, ")") {
				base := truncName[:idx]
				hint := truncName[idx:]
				styled := theme.Text.Sprint(true, base) + theme.TextMuted.Sprint(true, hint)
				projCell = padRightStyledPrecolored(truncName, styled, pWidth)
			} else {
				projCell = padRightStyled(truncName, pWidth, func(s string) string { return theme.Text.Sprint(true, s) })
			}
		}
	} else {
		projCell = padRight(truncName, pWidth)
	}

	toolsCellStr := formatToolsCell(row.Tools, tWidth, colors)
	tokCell := formatTokensCell(row.Tokens, tokWidth, colors)
	costCell := formatCostCell(row.Cost, costWidth, colors)
	daysCell := formatDaysCell(row.ActiveDays, daysWidth, colors)
	lastCell := formatDateCell(row.LastDay, lastWidth, colors)

	cells := []string{rankCell, projCell, toolsCellStr, tokCell, costCell, daysCell, lastCell}
	var b strings.Builder
	b.WriteString(" ")
	for i, c := range cells {
		b.WriteString(c)
		if i < len(cells)-1 {
			b.WriteString("  ")
		}
	}
	return b.String()
}

// displayNames maps each path to a short label. A lone repository shows as its
// own directory name, atlas-gateway. When two paths share a name, the
// distinguishing parent segment is added in parentheses, atlas-gateway (core).
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

// formatProjectBreakdown formats the tool split and top models under one project
// row. Both lines show token totals rather than shares, because some tools
// record no model attribution, which would make percentages misleading.
func formatProjectBreakdown(row thermal.ProjectRow, indent int, colors bool, st Style) []string {
	var lines []string
	prefix := strings.Repeat(" ", indent)
	if len(row.Tools) > 1 {
		line := weightLine(row.Tools, row.ToolTokens, 3, colors, st)
		if colors {
			lines = append(lines, prefix+st.Dim("tools   ")+line)
		} else {
			lines = append(lines, prefix+"tools   "+line)
		}
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
		line := weightLine(labels, shortWeights, 3, colors, st)
		if colors {
			lines = append(lines, prefix+st.Dim("models  ")+line)
		} else {
			lines = append(lines, prefix+"models  "+line)
		}
	}
	return lines
}

// weightLine renders "name tokens" entries, largest first, capped at limit.
func weightLine(names []string, weights map[string]int64, limit int, colors bool, st Style) string {
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
		tokStr := thermal.CompactNumber(weights[name])
		if colors {
			parts = append(parts, fmt.Sprintf("%s %s", theme.Text.Sprint(true, name), theme.Secondary.Sprint(true, tokStr)))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", name, tokStr))
		}
	}
	sep := " · "
	if colors {
		sep = st.Dim(" · ")
	}
	line := strings.Join(parts, sep)
	if rest := len(sorted) - len(shown); rest > 0 {
		if colors {
			line += st.Dim(fmt.Sprintf(" · +%d", rest))
		} else {
			line += fmt.Sprintf(" · +%d", rest)
		}
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
