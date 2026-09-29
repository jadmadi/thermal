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
	alignRight := []bool{true, false, false, false, true, true, false, false}

	shown := len(rep.Receipts)
	if top > 0 && top < shown {
		shown = top
	}

	rWidth := receiptRankWidth
	if shown >= 100 {
		rWidth = 5
	}
	widths := []int{rWidth, receiptDayWidth, receiptToolWidth, receiptEntityWidth, receiptTokensWidth, receiptCostWidth, receiptEvWidth, 16}

	var cardLines []string
	cardLines = append(cardLines, formatHeaderRow(headers, widths, alignRight, colors))
	cardLines = append(cardLines, " "+tableRule(widths))

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
			cStr = "—"
		}
		var costCell string
		if colors {
			if cStr == "—" {
				costCell = padLeftStyled(cStr, widths[5], dim)
			} else {
				costCell = padLeft(cStr, widths[5])
			}
		} else {
			costCell = padLeft(cStr, widths[5])
		}

		var evCell string
		if len(r.EvidenceSummary) > 0 {
			plainEv := truncate(strings.Join(r.EvidenceSummary, ", "), widths[6])
			if colors {
				evCell = padRightStyled(plainEv, widths[6], cyan)
			} else {
				evCell = padRight(plainEv, widths[6])
			}
		} else if r.Status == "CLAIMED" {
			if colors {
				evCell = padRightStyled("agent-claimed", widths[6], dim)
			} else {
				evCell = padRight("agent-claimed", widths[6])
			}
		} else {
			if colors {
				evCell = padRightStyled("none", widths[6], dim)
			} else {
				evCell = padRight("none", widths[6])
			}
		}

		rowStr := fmt.Sprintf(" %s  %s  %s  %s  %s  %s  %s  %s",
			formatRankCell(i+1, widths[0], colors),
			padRight(r.Day, widths[1]),
			padRight(truncate(r.Tool, widths[2]), widths[2]),
			padRight(truncate(entityName, widths[3]), widths[3]),
			padLeft(thermal.CompactNumber(r.Tokens), widths[4]),
			costCell,
			evCell,
			outcomeBadge(r.Status, r.Tier),
		)
		cardLines = append(cardLines, rowStr)
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

// RenderReceiptMarkdown formats verifiable work receipts into a GitHub PR-ready Markdown block.
func RenderReceiptMarkdown(rep thermal.ReceiptReport, top int) string {
	var sb strings.Builder

	sb.WriteString("### 🧾 Verifiable Work Receipts\n\n")

	if len(rep.Receipts) == 0 {
		sb.WriteString("> *No session receipts found in the selected window.*\n\n")
		return sb.String()
	}

	sum := rep.Summary
	spendEff := sum.SpendEfficiency
	if spendEff == "" {
		spendEff = "EXPLORATORY"
	}

	sb.WriteString(fmt.Sprintf("> **Verification Rate**: **%.1f%%** (%s / %s tok) · **%d Sessions Analyzed** (%d verified, %d claimed, %d unverified/failed)  \n",
		sum.VerificationRate,
		thermal.CompactNumber(sum.TokensVerified),
		thermal.CompactNumber(sum.TotalTokens),
		sum.TotalReceipts,
		sum.VerifiedCount,
		sum.ClaimedCount,
		sum.FailedCount+sum.UnverifiedCount,
	))
	sb.WriteString(fmt.Sprintf("> **Spend Efficiency**: `%s` · **Cost Attribution**: $%.2f ($%.2f verified)\n\n",
		spendEff,
		sum.TotalCost,
		sum.CostVerified,
	))

	// Resolve project display names
	var projPaths []string
	for _, r := range rep.Receipts {
		if r.Project != "" {
			projPaths = append(projPaths, r.Project)
		}
	}
	projDisplayNames := thermal.ProjectDisplayNames(projPaths)

	sb.WriteString("| # | Day | Tool | Session / Project | Tokens | Cost | Evidence / Tests | Outcome |\n")
	sb.WriteString("|---:|:---|:---|:---|---:|---:|:---|:---|\n")

	shown := len(rep.Receipts)
	if top > 0 && top < shown {
		shown = top
	}

	for i := 0; i < shown; i++ {
		r := rep.Receipts[i]

		entityName := r.SessionID
		if r.Project != "" {
			if dn, ok := projDisplayNames[r.Project]; ok && dn != "" {
				entityName = dn
			} else {
				entityName = thermal.ProjectSlug(r.Project)
			}
		}

		costStr := "—"
		if r.Cost > 0 {
			if r.EstimatedCost {
				costStr = fmt.Sprintf("~$%.2f", r.Cost)
			} else {
				costStr = fmt.Sprintf("$%.2f", r.Cost)
			}
		}

		evStr := "none"
		if len(r.EvidenceSummary) > 0 {
			var evItems []string
			for _, ev := range r.EvidenceSummary {
				evItems = append(evItems, "`"+ev+"`")
			}
			evStr = strings.Join(evItems, ", ")
		} else if r.Status == "CLAIMED" {
			evStr = "agent-claimed"
		}

		var outcomeStr string
		switch {
		case r.Status == "VERIFIED" || r.Tier == thermal.Tier1Verified:
			outcomeStr = "✅ Verified"
		case r.Status == "CLAIMED" || r.Tier == thermal.Tier2Claimed:
			outcomeStr = "⚠️ Claimed"
		case r.Status == "FAILED" || r.Tier == thermal.Tier3Failed:
			outcomeStr = "❌ Broken"
		default:
			outcomeStr = "ℹ️ Unverified"
		}

		sb.WriteString(fmt.Sprintf("| %d | %s | %s | `%s` | %s | %s | %s | %s |\n",
			i+1,
			r.Day,
			r.Tool,
			entityName,
			thermal.CompactNumber(r.Tokens),
			costStr,
			evStr,
			outcomeStr,
		))
	}

	if shown < len(rep.Receipts) {
		remainder := len(rep.Receipts) - shown
		sb.WriteString(fmt.Sprintf("\n*... and %d more session receipts*\n", remainder))
	}

	sb.WriteString("\n<details>\n<summary>ℹ️ Verification Hierarchy & Privacy Preservation</summary>\n\n")
	sb.WriteString("- **Tier 1 (Verified)**: Observed zero-exit test runner (`go test`, `pytest`, `cargo test`, `npm test`) or linter (`golangci-lint`, `eslint`, `ruff`) executions.\n")
	sb.WriteString("- **Tier 2 (Claimed)**: Completed agent sessions without verifiable test executions.\n")
	sb.WriteString("- **Tier 3 (Broken/Unverified)**: Non-zero test runner or linter exit codes, or unverified sessions.\n")
	sb.WriteString("- **Privacy Preservation**: Local evaluation only. Zero prompts, source files, error traces, or credentials recorded.\n")
	sb.WriteString("</details>\n")

	return sb.String()
}
