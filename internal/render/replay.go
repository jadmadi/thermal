package render

import (
	"fmt"
	"math"
	"os"
	"strings"

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
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }
	green := func(s string) string { return ColorCode(colors, "1;32", s) }
	yellow := func(s string) string { return ColorCode(colors, "1;33", s) }
	red := func(s string) string { return ColorCode(colors, "1;31", s) }

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
	for _, width := range widths {
		rule += width
	}
	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")

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

		sb.WriteString("  ")
		for j, c := range cells {
			if p.IsRecommended && j < 4 {
				c = gold(c)
			}
			sb.WriteString(c)
			if j < len(cells)-1 {
				sb.WriteString("  ")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")

	if rep.Recommendation != "" {
		sb.WriteString(fmt.Sprintf("\n  %s\n", highlight("Recommendation:")))
		for _, line := range strings.Split(rep.Recommendation, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				sb.WriteString(fmt.Sprintf("  • %s\n", line))
			}
		}
	}

	sb.WriteString("\n")
	return sb.String()
}
