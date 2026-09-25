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
	receiptRankWidth   = 4
	receiptDayWidth    = 10
	receiptToolWidth   = 10
	receiptEntityWidth = 22
	receiptTokensWidth = 11
	receiptCostWidth   = 9
	receiptEvWidth     = 24
	receiptStatusWidth = 13
	receiptTierWidth   = 10
)

// RenderReceipt renders the verifiable work receipts table and KPI metrics.
func RenderReceipt(rep thermal.ReceiptReport, top int, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	green := st.Success
	gold := st.Gold
	red := st.Error
	cyan := st.Accent

	outcomeBadge := func(status string, tier thermal.VerificationTier) string {
		switch {
		case status == "VERIFIED" || tier == thermal.Tier1Verified:
			return green("[PASS] Verified")
		case status == "CLAIMED" || tier == thermal.Tier2Claimed:
			return gold("[CLAIM] Untested")
		case status == "FAILED" || tier == thermal.Tier3Failed:
			return red("[FAIL] Broken")
		default:
			return dim("[INFO] Unknown")
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

	// Resolve project slugs from paths
	var projPaths []string
	for _, r := range rep.Receipts {
		if r.Project != "" {
			projPaths = append(projPaths, r.Project)
		}
	}
	projDisplayNames := thermal.ProjectDisplayNames(projPaths)

	headers := []string{"#", "Day", "Tool", "Session / Project", "Tokens", "Cost", "Evidence / Tests", "Outcome"}
	alignRight := []bool{false, false, false, false, true, true, false, false}
	widths := []int{receiptRankWidth, receiptDayWidth, receiptToolWidth, receiptEntityWidth, receiptTokensWidth, receiptCostWidth, receiptEvWidth, 16}

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

	shown := len(rep.Receipts)
	if top > 0 && top < shown {
		shown = top
	}

	for i := 0; i < shown; i++ {
		r := rep.Receipts[i]

		entityName := r.SessionID
		if r.Project != "" {
			if slug, ok := projDisplayNames[r.Project]; ok && slug != "" {
				entityName = slug
			} else {
				entityName = thermal.ProjectSlug(r.Project)
			}
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

		var rowB strings.Builder
		rowB.WriteString(" ")
		rowB.WriteString(thermal.PadLeft(fmt.Sprintf("%d.", i+1), widths[0]))
		rowB.WriteString("  ")
		rowB.WriteString(thermal.PadRight(r.Day, widths[1]))
		rowB.WriteString("  ")
		rowB.WriteString(thermal.PadRight(truncate(r.Tool, widths[2]), widths[2]))
		rowB.WriteString("  ")
		rowB.WriteString(thermal.PadRight(truncate(entityName, widths[3]), widths[3]))
		rowB.WriteString("  ")
		rowB.WriteString(thermal.PadLeft(thermal.CompactNumber(r.Tokens), widths[4]))
		rowB.WriteString("  ")
		rowB.WriteString(thermal.PadLeft(cStr, widths[5]))
		rowB.WriteString("  ")
		rowB.WriteString(thermal.PadRight(evText, widths[6]))
		rowB.WriteString("  ")
		rowB.WriteString(outcomeBadge(r.Status, r.Tier))
		cardLines = append(cardLines, rowB.String())
	}

	if shown < len(rep.Receipts) {
		remainder := len(rep.Receipts) - shown
		cardLines = append(cardLines, " "+dim(fmt.Sprintf("... and %d more session receipts (use --top to expand)", remainder)))
	}

	sb.WriteString(RenderCard(CardOptions{
		Title:       "Work Receipts",
		RightHeader: fmt.Sprintf("%.1f%% verified · %d sessions", sum.VerificationRate, sum.TotalReceipts),
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	footLimit := BoundedCardWidth(2)
	sb.WriteString("\n")
	for _, wl := range WrapBullet("  • ", "Legend: [PASS] Verified (Tier 1: tests/linters passed with exit 0) · [CLAIM] Untested (Tier 2: session completed without tests) · [FAIL] Broken (Tier 3: tests/linters failed).", footLimit, dim) {
		sb.WriteString(wl + "\n")
	}
	for _, wl := range WrapBullet("  • ", "Evidence Hierarchy: Tier 1 = Observed zero-exit test/linter runs or git commits; Tier 2 = Claimed; Tier 3 = Unverified/Failed.", footLimit, dim) {
		sb.WriteString(wl + "\n")
	}
	for _, wl := range WrapBullet("  • ", "Privacy Preservation: Local evaluation only. Zero prompt or error trace details recorded.", footLimit, dim) {
		sb.WriteString(wl + "\n")
	}
	sb.WriteString("\n")

	return sb.String()
}
