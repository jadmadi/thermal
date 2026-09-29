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
	modelWidth = 30
	// modelToolsWidth leaves room for "OpenCode,MiMoCode".
	modelToolsWidth = 20
)

// RenderModels prints a model leaderboard ranked by tokens. Cost is always an
// estimate, because recorded cost attaches to a session or a day, never to a
// single model. That is called out below the table.
func RenderModels(rep thermal.ModelReport, top int, noColor bool) string {
	return renderModels(rep, top, noColor, false)
}

// RenderModelsBreakdown adds token types and cognitive intensity (reasoning effort)
// breakdown rows under each model.
func RenderModelsBreakdown(rep thermal.ModelReport, top int, noColor bool) string {
	return renderModels(rep, top, noColor, true)
}

func renderModels(rep thermal.ModelReport, top int, noColor bool, breakdown bool) string {
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
		highlight("models"),
	))

	if len(rep.Rows) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No model activity in the selected window.")))
		return sb.String()
	}

	paths := make([]string, 0, len(rep.Rows))
	for _, row := range rep.Rows {
		paths = append(paths, row.Model)
	}
	names := displayNames(paths)

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

	headers := []string{"#", "Model", "Tools", "Tokens", "Cost", "Days", "Last"}
	alignRight := []bool{true, false, false, true, true, true, false}
	widths := []int{rWidth, modelWidth, modelToolsWidth, numberWidth, numberWidth, daysWidth, lastWidth}

	var cardLines []string
	cardLines = append(cardLines, formatHeaderRow(headers, widths, alignRight, colors))
	cardLines = append(cardLines, " "+tableRule(widths))

	for i := 0; i < shown; i++ {
		row := rep.Rows[i]
		rankCell := formatRankCell(i+1, rWidth, colors)
		truncName := truncate(names[row.Model], modelWidth)
		var modelCell string
		if colors {
			if row.Tokens == 0 {
				modelCell = padRightStyled(truncName, modelWidth, func(s string) string { return theme.TextMuted.Sprint(true, s) })
			} else if i == 0 {
				modelCell = padRightStyled(truncName, modelWidth, func(s string) string { return theme.Text.SprintBold(true, s) })
			} else {
				modelCell = padRightStyled(truncName, modelWidth, func(s string) string { return theme.Text.Sprint(true, s) })
			}
		} else {
			modelCell = padRight(truncName, modelWidth)
		}
		toolsCellStr := formatToolsCell(row.Tools, modelToolsWidth, colors)
		tokCell := formatTokensCell(row.Tokens, numberWidth, colors)
		costCell := formatCostCell(row.Cost, numberWidth, colors)
		daysCell := formatDaysCell(row.Days, daysWidth, colors)
		lastCell := formatDateCell(row.LastDay, lastWidth, colors)

		cells := []string{rankCell, modelCell, toolsCellStr, tokCell, costCell, daysCell, lastCell}
		var rowSb strings.Builder
		rowSb.WriteString(" ")
		for j, c := range cells {
			rowSb.WriteString(c)
			if j < len(cells)-1 {
				rowSb.WriteString("  ")
			}
		}
		cardLines = append(cardLines, rowSb.String())
		if breakdown {
			for _, bl := range formatModelBreakdown(row, rWidth+4, colors, st) {
				cardLines = append(cardLines, bl)
			}
		}
	}
	if rest := len(rep.Rows) - shown; rest > 0 {
		cardLines = append(cardLines, " "+dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest)))
	}

	cardLines = append(cardLines, " "+tableRule(widths))
	totals := rep.Totals
	totalCells := []string{
		padRight("", rWidth),
		padRight("Total", modelWidth),
		padRight(toolsCell(totals.Tools, modelToolsWidth), modelToolsWidth),
		padLeft(thermal.CompactNumber(totals.Tokens), numberWidth),
		padLeft(formatCostOrDash(totals.Cost), numberWidth),
		padLeft(fmt.Sprintf("%d", totals.Days), daysWidth),
		padRight(totals.LastDay, lastWidth),
	}
	var totalSb strings.Builder
	totalSb.WriteString(" ")
	for i, c := range totalCells {
		totalSb.WriteString(gold(c))
		if i < len(totalCells)-1 {
			totalSb.WriteString("  ")
		}
	}
	cardLines = append(cardLines, totalSb.String())

	sb.WriteString(RenderCard(CardOptions{
		Title:       "Models",
		RightHeader: "tokens & spend",
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))

	footLimit := BoundedCardWidth(2)
	if totals.Cost > 0 {
		for _, wl := range WrapText("Cost is estimated from models.dev list prices; recorded session cost is not attributable to one model.", footLimit) {
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

func formatModelBreakdown(row thermal.ModelRow, indent int, colors bool, st Style) []string {
	dim := st.Dim
	muted := st.Muted
	pad := strings.Repeat(" ", indent)
	var lines []string

	// Line 1: Disjoint tokens breakdown with cognitive intensity share
	var parts []string
	parts = append(parts, fmt.Sprintf("Input: %s", thermal.CompactNumber(row.Input)))
	parts = append(parts, fmt.Sprintf("Output: %s", thermal.CompactNumber(row.Output)))
	if row.Reasoning > 0 {
		share := 0.0
		if row.Output+row.Reasoning > 0 {
			share = (float64(row.Reasoning) / float64(row.Output+row.Reasoning)) * 100
		}
		parts = append(parts, fmt.Sprintf("Reasoning: %s (%s)",
			thermal.CompactNumber(row.Reasoning),
			dim(fmt.Sprintf("%.1f%% cognitive share", share)),
		))
	}
	cacheTotal := row.CacheRead + row.CacheWrite
	if cacheTotal > 0 {
		parts = append(parts, fmt.Sprintf("Cache: %s", thermal.CompactNumber(cacheTotal)))
	}

	lines = append(lines, fmt.Sprintf("%s%s %s", pad, dim("└─"), strings.Join(parts, muted("  ·  "))))

	// Line 2: If reasoning effort distribution is recorded (e.g. Codex or reasoning models)
	if len(row.ReasoningEffort) > 0 {
		type effortCount struct {
			effort string
			n      int
		}
		var efforts []effortCount
		for e, n := range row.ReasoningEffort {
			efforts = append(efforts, effortCount{e, n})
		}
		sort.Slice(efforts, func(i, j int) bool { return efforts[i].n > efforts[j].n })
		var eparts []string
		for _, ec := range efforts {
			eparts = append(eparts, fmt.Sprintf("%s: %d", ec.effort, ec.n))
		}
		lines = append(lines, fmt.Sprintf("%s   %s %s", pad, dim("Effort:"), strings.Join(eparts, muted("  "))))
	}

	return lines
}
