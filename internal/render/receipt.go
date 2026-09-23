// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	receiptRankWidth    = 4
	receiptDayWidth     = 10
	receiptToolWidth    = 10
	receiptEntityWidth  = 22
	receiptTokensWidth  = 11
	receiptCostWidth    = 9
	receiptEvWidth      = 24
	receiptStatusWidth  = 13
	receiptTierWidth    = 10
)

// RenderReceipt renders the verifiable work receipts table and KPI metrics.
func RenderReceipt(rep thermal.ReceiptReport, top int, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	green := func(s string) string { return ColorCode(colors, "1;32", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }
	red := func(s string) string { return ColorCode(colors, "1;31", s) }
	cyan := func(s string) string { return ColorCode(colors, "36", s) }

	statusBadge := func(status string) string {
		switch status {
		case "VERIFIED":
			return green("[VERIFIED]")
		case "CLAIMED":
			return gold("[CLAIMED]")
		case "FAILED":
			return red("[FAILED]")
		default:
			return dim("[UNVERIFIED]")
		}
	}

	tierBadge := func(tier thermal.VerificationTier) string {
		switch tier {
		case thermal.Tier1Verified:
			return green("Tier 1")
		case thermal.Tier2Claimed:
			return gold("Tier 2")
		case thermal.Tier3Failed:
			return red("Tier 3")
		default:
			return dim("Tier 3")
		}
	}

	effBadge := func(eff string) string {
		switch eff {
		case "EXCELLENT":
			return green("[EXCELLENT]")
		case "HEALTHY":
			return gold("[HEALTHY]")
		case "SPECULATIVE":
			return red("[SPECULATIVE]")
		default:
			return dim("[EXPLORATORY]")
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("receipt (verifiable work & session outcomes)"),
	))

	if len(rep.Receipts) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n\n", dim("No session receipts found in the selected window.")))
		return sb.String()
	}

	// KPI Summary Banner
	sum := rep.Summary
	sb.WriteString(fmt.Sprintf("  • %s %s (%s, %s, %s)\n",
		highlight("Sessions Analyzed:"),
		highlight(fmt.Sprintf("%d", sum.TotalReceipts)),
		green(fmt.Sprintf("%d verified", sum.VerifiedCount)),
		gold(fmt.Sprintf("%d claimed", sum.ClaimedCount)),
		dim(fmt.Sprintf("%d unverified/failed", sum.FailedCount+sum.UnverifiedCount)),
	))

	sb.WriteString(fmt.Sprintf("  • %s %s  %s  %s %s / %s (%s)\n",
		highlight("Verification Rate:"),
		green(fmt.Sprintf("%.1f%%", sum.VerificationRate)),
		dim("·"),
		highlight("Verified Volume:"),
		green(thermal.CompactNumber(sum.TokensVerified)),
		thermal.CompactNumber(sum.TotalTokens),
		fmt.Sprintf("%.1f%%", sum.VerifiedTokenRate),
	))

	costStr := fmt.Sprintf("$%.2f", sum.TotalCost)
	costVerStr := fmt.Sprintf("$%.2f verified", sum.CostVerified)
	sb.WriteString(fmt.Sprintf("  • %s %s (%s)  %s  %s %s\n\n",
		highlight("Cost Attribution:"),
		costStr,
		costVerStr,
		dim("·"),
		highlight("Spend Efficiency:"),
		effBadge(sum.SpendEfficiency),
	))

	headers := []string{"#", "Day", "Tool", "Session / Project", "Tokens", "Cost", "Evidence / Tests", "Status", "Tier"}
	alignRight := []bool{false, false, false, false, true, true, false, false, false}
	widths := []int{receiptRankWidth, receiptDayWidth, receiptToolWidth, receiptEntityWidth, receiptTokensWidth, receiptCostWidth, receiptEvWidth, receiptStatusWidth, receiptTierWidth}

	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}

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
	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")

	shown := len(rep.Receipts)
	if top > 0 && top < shown {
		shown = top
	}

	for i := 0; i < shown; i++ {
		r := rep.Receipts[i]

		entityName := r.SessionID
		if r.Project != "" {
			entityName = r.Project
		}

		cStr := fmt.Sprintf("$%.2f", r.Cost)
		if r.EstimatedCost {
			cStr = "~" + cStr
		}
		if r.Cost == 0 {
			cStr = "-"
		}

		evText := dim("none")
		if len(r.EvidenceSummary) > 0 {
			evText = cyan(truncate(strings.Join(r.EvidenceSummary, ", "), widths[6]))
		} else if r.Status == "CLAIMED" {
			evText = dim("agent-claimed")
		}

		sb.WriteString("  ")
		sb.WriteString(thermal.PadLeft(fmt.Sprintf("%d.", i+1), widths[0]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadRight(r.Day, widths[1]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadRight(truncate(r.Tool, widths[2]), widths[2]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadRight(truncate(entityName, widths[3]), widths[3]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadLeft(thermal.CompactNumber(r.Tokens), widths[4]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadLeft(cStr, widths[5]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadRight(evText, widths[6]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadRight(statusBadge(r.Status), widths[7]))
		sb.WriteString("  ")
		sb.WriteString(thermal.PadRight(tierBadge(r.Tier), widths[8]))
		sb.WriteString("\n")
	}

	if shown < len(rep.Receipts) {
		remainder := len(rep.Receipts) - shown
		sb.WriteString(fmt.Sprintf("  %s\n", dim(fmt.Sprintf("... and %d more session receipts (use --top to expand)", remainder))))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		dim("•"),
		dim("Evidence Hierarchy: Tier 1 = Observed zero-exit test/linter runs or git commits; Tier 2 = Claimed; Tier 3 = Unverified."),
	))
	sb.WriteString(fmt.Sprintf("  %s %s\n\n",
		dim("•"),
		dim("Privacy Preservation: Local evaluation only. Zero prompt or error trace details recorded."),
	))

	return sb.String()
}
