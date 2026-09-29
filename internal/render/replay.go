// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	planNameWidth    = 26
	planTypeWidth    = 6
	planCostWidth    = 12
	planDeltaWidth   = 18
	planVerdictWidth = 32
)

// RenderReplay formats the simulation outcome into an executive comparison table.
func RenderReplay(rep thermal.ReplayReport, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	gold := st.Gold
	green := st.Success
	yellow := st.Warning
	red := st.Error

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("replay"),
		dim("·"),
		dim(fmt.Sprintf("%d-day workload simulation", rep.Workload.TotalDays)),
	))

	w := rep.Workload
	if w.ActiveDays == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n\n", dim("No active coding sessions in the selected window.")))
		return sb.String()
	}

	// Workload summary banner
	sb.WriteString(fmt.Sprintf("  %s\n", highlight("Real Workload")))
	hitRatePercent := fmt.Sprintf("%.1f%%", w.CacheHitRate*100)
	spendPrefix := ""
	if w.IsEstimatedSpend {
		spendPrefix = "~"
	}
	sb.WriteString(fmt.Sprintf("  %s %d %s  %s %s  %s %s  %s %s$%.2f\n",
		dim("Active days:"), w.ActiveDays, dim("·"),
		dim("Total:"), thermal.CompactNumber(w.TotalTokens),
		dim("· Cache hit rate:"), hitRatePercent,
		dim("· Spend/mo:"), spendPrefix, w.ActualSpend,
	))
	sb.WriteString(fmt.Sprintf("  %s %s  %s %s  %s %s\n\n",
		dim("Daily volume: Median"), thermal.CompactNumber(w.MedianDailyTokens),
		dim("· p90 burst"), thermal.CompactNumber(w.P90DailyTokens),
		dim("· Peak"), thermal.CompactNumber(w.PeakDailyTokens),
	))

	verdictColWidth := planVerdictWidth
	for _, p := range rep.Plans {
		plainVerdict := fmt.Sprintf("%s (%s)", p.CapacityVerdict, p.VerdictDetail)
		if len(plainVerdict) > verdictColWidth {
			verdictColWidth = len(plainVerdict)
		}
	}

	headers := []string{"Plan / Target Model", "Type", "Cost/Mo", "Delta vs Actual", "Capacity Verdict"}
	alignRight := []bool{false, false, true, true, false}
	widths := []int{planNameWidth, planTypeWidth, planCostWidth, planDeltaWidth, verdictColWidth}

	rule := 2 * (len(widths) - 1)
	for _, width := range widths {
		rule += width
	}

	var cardLines []string
	cardLines = append(cardLines, formatHeaderRow(headers, widths, alignRight, colors))
	cardLines = append(cardLines, " "+tableRule(widths))

	for _, p := range rep.Plans {
		typeLabel := "Sub"
		if p.Type == "payg" {
			typeLabel = "Payg"
		}

		costStr := fmt.Sprintf("$%.2f", p.MonthlyCost)
		deltaSign := "+"
		if p.CostDelta < 0 {
			deltaSign = "-"
		}
		deltaStr := fmt.Sprintf("%s$%.2f (%+.0f%%)", deltaSign, math.Abs(p.CostDelta), p.CostDeltaPercent)

		plainVerdict := fmt.Sprintf("%s (%s)", p.CapacityVerdict, p.VerdictDetail)
		var verdictCell string
		if colors {
			var verdictStyled string
			switch p.CapacityVerdict {
			case "PASS":
				verdictStyled = green(plainVerdict)
			case "DEGRADED":
				verdictStyled = yellow(plainVerdict)
			default:
				verdictStyled = red(plainVerdict)
			}
			verdictCell = padRightStyledPrecolored(plainVerdict, verdictStyled, widths[4])
		} else {
			verdictCell = padRight(plainVerdict, widths[4])
		}

		nameLabel := p.Name
		if p.IsRecommended {
			nameLabel = "★ " + nameLabel
		}
		var nameCell string
		if p.IsRecommended && colors {
			nameCell = padRightStyled(truncate(nameLabel, widths[0]), widths[0], gold)
		} else {
			nameCell = padRight(truncate(nameLabel, widths[0]), widths[0])
		}

		var typeCell string
		if p.IsRecommended && colors {
			typeCell = padRightStyled(typeLabel, widths[1], gold)
		} else {
			typeCell = padRight(typeLabel, widths[1])
		}

		var costCell string
		if p.IsRecommended && colors {
			costCell = padLeftStyled(costStr, widths[2], gold)
		} else {
			costCell = padLeft(costStr, widths[2])
		}

		var deltaCell string
		if p.IsRecommended && colors {
			deltaCell = padLeftStyled(deltaStr, widths[3], gold)
		} else {
			deltaCell = padLeft(deltaStr, widths[3])
		}

		cells := []string{nameCell, typeCell, costCell, deltaCell, verdictCell}
		cardLines = append(cardLines, " "+strings.Join(cells, "  "))
	}

	sb.WriteString(RenderCard(CardOptions{
		Title:       "Simulation Results",
		RightHeader: "subscription comparison",
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	if rep.Recommendation != "" {
		sb.WriteString(fmt.Sprintf("\n  %s\n", highlight("Recommendation:")))
		recLimit := BoundedCardWidth(2)
		for _, line := range strings.Split(rep.Recommendation, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				for _, wl := range WrapBullet("  • ", line, recLimit) {
					sb.WriteString(wl + "\n")
				}
			}
		}
	}

	sb.WriteString("\n")
	return sb.String()
}
