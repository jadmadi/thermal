// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

// padLeft returns s left-padded with spaces to width visual display cells.
func padLeft(s string, width int) string {
	w := ansi.StringWidth(s)
	if n := width - w; n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

// padRight returns s right-padded with spaces to width visual display cells.
func padRight(s string, width int) string {
	w := ansi.StringWidth(s)
	if n := width - w; n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

// padLeftStyled pads s to visual width, applying style to s without padding escape codes.
func padLeftStyled(s string, width int, style func(string) string) string {
	w := ansi.StringWidth(s)
	styled := s
	if style != nil {
		styled = style(s)
	}
	if n := width - w; n > 0 {
		return strings.Repeat(" ", n) + styled
	}
	return styled
}

// padRightStyled pads s to visual width, applying style to s without padding escape codes.
func padRightStyled(s string, width int, style func(string) string) string {
	w := ansi.StringWidth(s)
	styled := s
	if style != nil {
		styled = style(s)
	}
	if n := width - w; n > 0 {
		return styled + strings.Repeat(" ", n)
	}
	return styled
}

// padRightStyledPrecolored takes a pre-styled string whose plain visual width is known,
// and right-pads it to width display cells.
func padRightStyledPrecolored(plain string, styled string, width int) string {
	w := ansi.StringWidth(plain)
	if n := width - w; n > 0 {
		return styled + strings.Repeat(" ", n)
	}
	return styled
}

// formatHeaderRow formats table column headers in Blaze Amber Orange (theme.Primary.SprintBold)
// with consistent 2-space separators and column alignment.
func formatHeaderRow(headers []string, widths []int, alignRight []bool, colors bool) string {
	var hsb strings.Builder
	hsb.WriteString(" ")
	for i, h := range headers {
		var cell string
		isRight := i < len(alignRight) && alignRight[i]
		if isRight {
			cell = padLeftStyled(h, widths[i], func(s string) string { return theme.Primary.SprintBold(colors, s) })
		} else {
			cell = padRightStyled(h, widths[i], func(s string) string { return theme.Primary.SprintBold(colors, s) })
		}
		hsb.WriteString(cell)
		if i < len(headers)-1 {
			hsb.WriteString("  ")
		}
	}
	return hsb.String()
}

// tableRule returns a horizontal divider line spanning all columns and separators.
func tableRule(widths []int) string {
	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}
	return strings.Repeat("─", rule)
}

// formatCostOrDash formats cost as $X.XX or em dash if zero or negative.
func formatCostOrDash(v float64) string {
	if v <= 0 {
		return "—"
	}
	if v < 0.01 {
		return fmt.Sprintf("$%.4f", v)
	}
	return fmt.Sprintf("$%.2f", v)
}

// formatRankCell formats a rank number (e.g. "1.") with podium medal styling.
func formatRankCell(rank int, rWidth int, colors bool) string {
	rankPlain := fmt.Sprintf("%d.", rank)
	if !colors {
		return padLeft(rankPlain, rWidth)
	}
	switch rank {
	case 1:
		return padLeftStyled(rankPlain, rWidth, func(s string) string { return theme.Secondary.SprintBold(true, s) })
	case 2:
		return padLeftStyled(rankPlain, rWidth, func(s string) string { return theme.Text.Sprint(true, s) })
	case 3:
		return padLeftStyled(rankPlain, rWidth, func(s string) string { return theme.Warning.Sprint(true, s) })
	default:
		return padLeftStyled(rankPlain, rWidth, func(s string) string { return theme.TextMuted.Sprint(true, s) })
	}
}

// formatTokensCell formats token count with unified thermal heatmap color tiers.
func formatTokensCell(tokens int64, width int, colors bool) string {
	tokPlain := thermal.CompactNumber(tokens)
	if !colors {
		return padLeft(tokPlain, width)
	}
	if tokens == 0 {
		return padLeftStyled(tokPlain, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
	}
	if tokens >= 1_000_000_000 {
		return padLeftStyled(tokPlain, width, func(s string) string { return theme.Primary.SprintBold(true, s) })
	}
	if tokens >= 100_000_000 {
		return padLeftStyled(tokPlain, width, func(s string) string { return theme.Secondary.Sprint(true, s) })
	}
	if tokens >= 1_000_000 {
		return padLeftStyled(tokPlain, width, func(s string) string { return theme.Text.Sprint(true, s) })
	}
	return padLeftStyled(tokPlain, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
}

// formatCostCell formats dollar cost or em dash with unified styling.
func formatCostCell(cost float64, width int, colors bool) string {
	costPlain := formatCostOrDash(cost)
	if !colors {
		return padLeft(costPlain, width)
	}
	if costPlain == "—" {
		return padLeftStyled(costPlain, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
	}
	if cost >= 1000.0 {
		return padLeftStyled(costPlain, width, func(s string) string { return theme.Secondary.Sprint(true, s) })
	}
	return padLeftStyled(costPlain, width, func(s string) string { return theme.Text.Sprint(true, s) })
}

// formatDaysCell formats active days with high-activity badges.
func formatDaysCell(days int, width int, colors bool) string {
	daysPlain := fmt.Sprintf("%d", days)
	if !colors {
		return padLeft(daysPlain, width)
	}
	if days >= 30 {
		return padLeftStyled(daysPlain, width, func(s string) string { return theme.Success.Sprint(true, s) })
	}
	if days >= 10 {
		return padLeftStyled(daysPlain, width, func(s string) string { return theme.Text.Sprint(true, s) })
	}
	if days <= 1 {
		return padLeftStyled(daysPlain, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
	}
	return padLeftStyled(daysPlain, width, func(s string) string { return theme.Text.Sprint(true, s) })
}

// formatDateCell formats date string with muted styling.
func formatDateCell(date string, width int, colors bool) string {
	if !colors {
		return padRight(date, width)
	}
	return padRightStyled(date, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
}

// formatToolsCell formats contributing tools with dim overflow indicators.
func formatToolsCell(tools []string, width int, colors bool) string {
	tCell := toolsCell(tools, width)
	if !colors {
		return padRight(tCell, width)
	}
	if tCell == "—" {
		return padRightStyled(tCell, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
	}
	if idx := strings.Index(tCell, " +"); idx != -1 {
		baseTools := tCell[:idx]
		overflow := tCell[idx:]
		styled := theme.Text.Sprint(true, baseTools) + theme.TextMuted.Sprint(true, overflow)
		return padRightStyledPrecolored(tCell, styled, width)
	}
	return padRightStyled(tCell, width, func(s string) string { return theme.Text.Sprint(true, s) })
}
