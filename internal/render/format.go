// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

// formatCost is the recorded-cost form used by report tests and internal
// renderers. Estimated values must use formatCostValue/formatEstimatedCost.
func formatCost(v float64) string {
	return thermal.FormatCost(v, false)
}

func formatCostValue(v float64, estimated bool) string {
	return thermal.FormatCost(v, estimated)
}

func formatEstimatedCost(v float64) string {
	return thermal.FormatCost(v, true)
}

func formatCostOrDash(v float64) string {
	return thermal.FormatCostOrDash(v, false)
}

func formatCostOrDashValue(v float64, estimated bool) string {
	return thermal.FormatCostOrDash(v, estimated)
}

func formatTotalCost(total, recorded, estimated float64) string {
	return thermal.FormatTotalCost(total, recorded, estimated)
}

func formatPercent(v float64, decimals int) string {
	return thermal.FormatPercent(v, decimals)
}

func formatCacheHit(v float64) string {
	return thermal.FormatCacheHit(v)
}

func joinTools(tools []string) string {
	return strings.Join(tools, " · ")
}

func formatCostCellKind(cost float64, estimated bool, width int, colors bool) string {
	costPlain := thermal.FormatCostOrDash(cost, estimated)
	if !colors {
		return padLeft(costPlain, width)
	}
	if costPlain == "—" {
		return padLeftStyled(costPlain, width, func(s string) string { return theme.TextMuted.Sprint(true, s) })
	}
	if strings.HasPrefix(costPlain, "~") {
		return padLeftStyled(costPlain, width, func(s string) string { return theme.Secondary.Sprint(true, s) })
	}
	if cost >= 1000.0 {
		return padLeftStyled(costPlain, width, func(s string) string { return theme.Secondary.Sprint(true, s) })
	}
	return padLeftStyled(costPlain, width, func(s string) string { return theme.Text.Sprint(true, s) })
}
