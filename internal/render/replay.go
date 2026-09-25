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

	headers := []string{"Plan / Target Model", "Type", "Cost/Mo", "Delta vs Actual", "Capacity Verdict"}
	alignRight := []bool{false, false, true, true, false}
	widths := []int{planNameWidth, planTypeWidth, planCostWidth, planDeltaWidth, planVerdictWidth}

	rule := 2 * (len(widths) - 1)
	for _, width := range widths {
		rule += width
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

		var verdictStyled string
		switch p.CapacityVerdict {
		case "PASS":
			verdictStyled = green(fmt.Sprintf("PASS (%s)", p.VerdictDetail))
		case "DEGRADED":
			verdictStyled = yellow(fmt.Sprintf("DEGRADED (%s)", p.VerdictDetail))
		default:
			verdictStyled = red(fmt.Sprintf("FAIL (%s)", p.VerdictDetail))
		}

		nameLabel := p.Name
		if p.IsRecommended {
			nameLabel = "★ " + nameLabel
		}

		cells := []string{
			thermal.PadRight(truncate(nameLabel, widths[0]), widths[0]),
			thermal.PadRight(typeLabel, widths[1]),
			thermal.PadLeft(costStr, widths[2]),
			thermal.PadLeft(deltaStr, widths[3]),
			thermal.PadRight(verdictStyled, widths[4]),
		}

		var rowB strings.Builder
		rowB.WriteString(" ")
		for j, c := range cells {
			if p.IsRecommended && j < 4 {
				c = gold(c)
			}
			rowB.WriteString(c)
			if j < len(cells)-1 {
				rowB.WriteString("  ")
			}
		}
		cardLines = append(cardLines, rowB.String())
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
