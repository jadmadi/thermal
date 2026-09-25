// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
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

	headers := []string{"#", "Model", "Tools", "Tokens", "Cost", "Days", "Last"}
	alignRight := []bool{false, false, false, true, true, true, false}
	widths := []int{rankWidth, modelWidth, modelToolsWidth, numberWidth, numberWidth, daysWidth, lastWidth}

	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}

	var cardLines []string

	var hsb strings.Builder
	hsb.WriteString(" ")
	for i, h := range headers {
		cell := thermal.PadRight(h, widths[i])
		if alignRight[i] {
			cell = thermal.PadLeft(h, widths[i])
		}
		hsb.WriteString(dim(cell))
		if i < len(headers)-1 {
			hsb.WriteString("  ")
		}
	}
	cardLines = append(cardLines, hsb.String())
	cardLines = append(cardLines, " "+strings.Repeat("─", rule))

	shown := len(rep.Rows)
	if top > 0 && top < shown {
		shown = top
	}
	for i := 0; i < shown; i++ {
		row := rep.Rows[i]
		cells := []string{
			thermal.PadLeft(fmt.Sprintf("%d.", i+1), widths[0]),
			thermal.PadRight(truncate(names[row.Model], widths[1]), widths[1]),
			thermal.PadRight(toolsCell(row.Tools, widths[2]), widths[2]),
			thermal.PadLeft(thermal.CompactNumber(row.Tokens), widths[3]),
			thermal.PadLeft(formatCostOrDash(row.Cost), widths[4]),
			thermal.PadLeft(fmt.Sprintf("%d", row.Days), widths[5]),
			thermal.PadRight(row.LastDay, widths[6]),
		}
		var rowSb strings.Builder
		rowSb.WriteString(" ")
		for j, c := range cells {
			rowSb.WriteString(c)
			if j < len(cells)-1 {
				rowSb.WriteString("  ")
			}
		}
		cardLines = append(cardLines, rowSb.String())
	}
	if rest := len(rep.Rows) - shown; rest > 0 {
		cardLines = append(cardLines, " "+dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest)))
	}

	cardLines = append(cardLines, " "+strings.Repeat("─", rule))
	totals := rep.Totals
	totalCells := []string{
		thermal.PadRight("", rankWidth),
		thermal.PadRight("Total", modelWidth),
		thermal.PadRight(toolsCell(totals.Tools, modelToolsWidth), modelToolsWidth),
		thermal.PadLeft(thermal.CompactNumber(totals.Tokens), numberWidth),
		thermal.PadLeft(formatCostOrDash(totals.Cost), numberWidth),
		thermal.PadLeft(fmt.Sprintf("%d", totals.Days), daysWidth),
		thermal.PadRight(totals.LastDay, lastWidth),
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
