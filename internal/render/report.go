// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	periodWidth = 10
	modelsWidth = 19
	numberWidth = 9
)

// RenderReport prints a daily, weekly, or monthly report as a table. Columns
// for reasoning tokens appear only when some row carries that data. Cost shows
// stored plus estimated cost; estimated totals are called out below the table.
func RenderReport(rep thermal.Report, noColor bool) string {
	return renderReport(rep, noColor, false)
}

// RenderReportBreakdown renders the same table with an indented per-model row
// under every period that recorded a model.
func RenderReportBreakdown(rep thermal.Report, noColor bool) string {
	return renderReport(rep, noColor, true)
}

func renderReport(rep thermal.Report, noColor bool, breakdown bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	faint := st.Faint
	gold := st.Gold

	hasReasoning := false
	for _, row := range rep.Rows {
		if row.Reasoning > 0 {
			hasReasoning = true
			break
		}
	}

	toolLabel := rep.Tool
	if toolLabel == "" {
		toolLabel = "all tools"
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight(rep.Type+" usage"),
		dim("· "+toolLabel),
	))

	if len(rep.Rows) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No activity in the selected window.")))
		return sb.String()
	}

	unit := rep.Type
	switch rep.Type {
	case "daily":
		unit = "day"
	case "weekly":
		unit = "week"
	case "monthly":
		unit = "month"
	}
	periodLabel := fmt.Sprintf("%d %ss", len(rep.Rows), unit)
	if len(rep.Rows) == 1 {
		periodLabel = fmt.Sprintf("1 %s", unit)
	}
	hitRateStr := "—"
	if rep.Totals.Cache+rep.Totals.Input > 0 {
		hitRateStr = fmt.Sprintf("%.1f%%", float64(rep.Totals.Cache)/float64(rep.Totals.Cache+rep.Totals.Input)*100)
	}

	costDisplay := formatCost(rep.Totals.Cost)
	if rep.Totals.EstimatedCost > 0 {
		costDisplay = "~" + costDisplay
	}

	sb.WriteString(fmt.Sprintf("  Window: [%s · %s]  Total: [%s · %s]  Cache: [%s hit]\n\n",
		highlight(periodLabel),
		dim(toolLabel),
		highlight(thermal.CompactNumber(rep.Totals.Tokens)+" tok"),
		highlight(costDisplay),
		st.Success(hitRateStr),
	))

	headers := []string{"Period", "Models", "Input", "Output"}
	alignRight := []bool{false, false, true, true}
	if hasReasoning {
		headers = append(headers, "Reason")
		alignRight = append(alignRight, true)
	}
	headers = append(headers, "Cache", "Tokens", "Cost")
	alignRight = append(alignRight, true, true, true)

	widths := []int{periodWidth, modelsWidth}
	for range headers[2:] {
		widths = append(widths, numberWidth)
	}

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
	cardLines = append(cardLines, strings.Repeat("─", rule))

	formatRow := func(period, models string, row thermal.PeriodRow, style func(string) string) string {
		cells := []string{thermal.PadRight(truncate(period, periodWidth), periodWidth),
			thermal.PadRight(truncate(models, modelsWidth), modelsWidth)}
		nums := []int64{row.Input, row.Output}
		if hasReasoning {
			nums = append(nums, row.Reasoning)
		}
		nums = append(nums, row.Cache, row.Tokens)
		for _, n := range nums {
			cells = append(cells, thermal.PadLeft(thermal.CompactNumber(n), numberWidth))
		}
		if row.Cost > 0 {
			cells = append(cells, thermal.PadLeft(formatCost(row.Cost), numberWidth))
		} else {
			cells = append(cells, thermal.PadLeft("—", numberWidth))
		}
		var rsb strings.Builder
		rsb.WriteString(" ")
		for i, c := range cells {
			if style != nil {
				c = style(c)
			}
			rsb.WriteString(c)
			if i < len(cells)-1 {
				rsb.WriteString("  ")
			}
		}
		return rsb.String()
	}

	for _, row := range rep.Rows {
		cardLines = append(cardLines, formatRow(row.Period, modelCell(row.Models, modelsWidth), row, nil))
		if !breakdown {
			continue
		}
		for _, name := range thermal.TopModels(row.Models) {
			m := row.Models[name]
			cardLines = append(cardLines, formatRow("", "  └─ "+name, thermal.PeriodRow{
				Input:     m.Input,
				Output:    m.Output,
				Reasoning: m.Reasoning,
				Cache:     m.Cache(),
				Tokens:    m.Total(),
			}, faint))
		}
	}

	cardLines = append(cardLines, strings.Repeat("─", rule))
	cardLines = append(cardLines, formatRow("Total", "", rep.Totals, gold))

	cardWidth := BoundedCardWidth(2)

	title := strings.ToUpper(rep.Type[:1]) + rep.Type[1:] + " Usage"

	sb.WriteString(RenderCard(CardOptions{
		Title:       title,
		RightHeader: toolLabel,
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
		MaxWidth:    cardWidth,
	}))

	if rep.Totals.EstimatedCost > 0 {
		note := fmt.Sprintf("Total = %s recorded + ~%s estimated from pricing data.",
			formatCost(rep.Totals.StoredCost), formatCost(rep.Totals.EstimatedCost))
		for _, line := range WrapText(note, cardWidth) {
			sb.WriteString(fmt.Sprintf("\n  %s", dim(line)))
		}
		sb.WriteString("\n")
	}
	if note := unpriceableNote(rep.Totals.UnattributedTokens); note != "" {
		for _, line := range WrapText(note, cardWidth) {
			sb.WriteString("  " + dim(line) + "\n")
		}
	}
	if len(rep.Totals.MissingPricing) > 0 {
		models := rep.Totals.MissingPricing
		if len(models) > 4 {
			models = append(append([]string{}, models[:4]...), fmt.Sprintf("+%d more", len(models)-4))
		}
		note := "No pricing for: " + strings.Join(models, ", ")
		for _, line := range WrapText(note, cardWidth) {
			sb.WriteString(fmt.Sprintf("  %s\n", dim(line)))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// truncate shortens s to width runes, adding an ellipsis when it cuts. Runes
// rather than bytes so a multi-byte model name is never split.
func truncate(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

// runeLen counts runes for width comparisons in the table.
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func formatCost(v float64) string {
	if v < 0.01 {
		return fmt.Sprintf("$%.4f", v)
	}
	return fmt.Sprintf("$%.2f", v)
}

// modelCell renders the top models in a fixed-width cell, truncating with an
// ellipsis and a "+N" count when the list does not fit.
func modelCell(models map[string]thermal.ModelTokens, width int) string {
	if len(models) == 0 {
		return "—"
	}
	names := thermal.TopModels(models)
	shown := make([]string, 0, len(names))
	for i, n := range names {
		suffix := ""
		if i < len(names)-1 {
			suffix = fmt.Sprintf(" +%d", len(names)-i-1)
		}
		candidate := strings.Join(append(append([]string{}, shown...), n), ", ") + suffix
		if runeLen(candidate) > width {
			if len(shown) == 0 {
				avail := width - runeLen(suffix) - 1
				if avail < 4 {
					avail = 4
				}
				if runeLen(n) > avail {
					n = truncate(n, avail)
				}
				shown = append(shown, n)
			}
			break
		}
		shown = append(shown, n)
	}
	cell := strings.Join(shown, ", ")
	if used := len(shown); used < len(names) {
		cell += fmt.Sprintf(" +%d", len(names)-used)
	}
	return cell
}
