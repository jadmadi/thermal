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
	yieldRankWidth   = 4
	yieldEntityWidth = 30
	yieldTokensWidth = 11
	yieldLinesWidth  = 9
	yieldNetWidth    = 10
	yieldRatioWidth  = 13
	yieldEffWidth    = 13
)

// RenderYield renders the token yield and code output delta table.
func RenderYield(rep thermal.YieldReport, top int, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	green := st.Success
	gold := st.Gold
	magenta := st.Warning
	cyan := st.Accent

	effBadge := func(eff, status string) string {
		if status == "UNMEASURED" {
			return dim("[EXPLORATORY]")
		}
		switch eff {
		case "HIGH":
			return green("[HIGH]")
		case "BALANCED":
			return gold("[BALANCED]")
		case "VERBOSE":
			return magenta("[VERBOSE]")
		default:
			return cyan("[" + eff + "]")
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("yield (code output delta & efficiency)"),
	))

	if len(rep.Models) == 0 && len(rep.Tools) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No token activity in the selected window.")))
		return sb.String()
	}

	headers := []string{"#", "Entity / Model", "Tokens", "+Lines", "-Lines", "Net Lines", "Yield", "Efficiency"}
	alignRight := []bool{true, false, true, true, true, true, true, false}

	// Resolve project slugs from paths
	var projPaths []string
	for _, p := range rep.Projects {
		if p.Path != "" {
			projPaths = append(projPaths, p.Path)
		} else {
			projPaths = append(projPaths, p.Name)
		}
	}
	projDisplayNames := thermal.ProjectDisplayNames(projPaths)

	renderSection := func(title string, rows []thermal.YieldRow) {
		if len(rows) == 0 {
			return
		}
		shown := len(rows)
		if top > 0 && top < shown {
			shown = top
		}

		rWidth := yieldRankWidth
		if shown >= 100 {
			rWidth = 5
		}
		widths := []int{rWidth, yieldEntityWidth, yieldTokensWidth, yieldLinesWidth, yieldLinesWidth, yieldNetWidth, yieldRatioWidth, yieldEffWidth}

		var cardLines []string
		cardLines = append(cardLines, formatHeaderRow(headers, widths, alignRight, colors))
		cardLines = append(cardLines, " "+tableRule(widths))

		for i := 0; i < shown; i++ {
			r := rows[i]

			addStr := fmt.Sprintf("+%s", thermal.CompactNumber(r.LinesAdded))
			delStr := fmt.Sprintf("-%s", thermal.CompactNumber(r.LinesDeleted))
			netStr := fmt.Sprintf("+%s", thermal.CompactNumber(r.NetLines))
			if r.NetLines < 0 {
				netStr = fmt.Sprintf("-%s", thermal.CompactNumber(-r.NetLines))
			} else if r.NetLines == 0 {
				netStr = "0"
			}

			yieldStr := "unmeasured"
			if r.Status == "MEASURED" {
				yieldStr = fmt.Sprintf("%s tok/ln", thermal.CompactNumber(int64(r.TokensPerNet)))
			}

			displayName := r.Name
			if r.Type == "project" {
				if r.Path != "" {
					if dn, ok := projDisplayNames[r.Path]; ok && dn != "" {
						displayName = dn
					} else {
						displayName = thermal.ProjectSlug(r.Path)
					}
				} else if dn, ok := projDisplayNames[r.Name]; ok && dn != "" {
					displayName = dn
				} else {
					displayName = thermal.ProjectSlug(r.Name)
				}
			}

			var rowB strings.Builder
			rowB.WriteString(" ")
			rowB.WriteString(formatRankCell(i+1, widths[0], colors))
			rowB.WriteString("  ")
			rowB.WriteString(padRight(truncate(displayName, widths[1]), widths[1]))
			rowB.WriteString("  ")
			rowB.WriteString(padLeft(thermal.CompactNumber(r.Tokens), widths[2]))
			rowB.WriteString("  ")
			rowB.WriteString(padLeft(addStr, widths[3]))
			rowB.WriteString("  ")
			rowB.WriteString(padLeft(delStr, widths[4]))
			rowB.WriteString("  ")
			rowB.WriteString(padLeft(netStr, widths[5]))
			rowB.WriteString("  ")
			if r.Status == "MEASURED" {
				rowB.WriteString(padLeft(yieldStr, widths[6]))
			} else {
				if colors {
					rowB.WriteString(padLeftStyled(yieldStr, widths[6], dim))
				} else {
					rowB.WriteString(padLeft(yieldStr, widths[6]))
				}
			}
			rowB.WriteString("  ")
			rowB.WriteString(effBadge(r.Efficiency, r.Status))
			cardLines = append(cardLines, rowB.String())
		}

		if rest := len(rows) - shown; rest > 0 {
			cardLines = append(cardLines, " "+dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest)))
		}

		sb.WriteString(RenderCard(CardOptions{
			Title:       title,
			RightHeader: "code delta & efficiency",
			Lines:       cardLines,
			Indent:      2,
			Colors:      colors,
			TitleColor:  theme.Primary,
			BorderColor: theme.Border,
		}))
		sb.WriteString("\n")
	}

	// Render Models if present
	if len(rep.Models) > 0 {
		headers[1] = "Model"
		renderSection("Model Yield Telemetry", rep.Models)
	}

	// Render Tools if present
	if len(rep.Tools) > 0 {
		headers[1] = "Tool"
		renderSection("Tool Breakdown", rep.Tools)
	}

	// Render Projects if present
	if len(rep.Projects) > 0 {
		headers[1] = "Project"
		renderSection("Project Breakdown", rep.Projects)
	}

	// Render Totals Summary
	tot := rep.Totals
	totAddStr := fmt.Sprintf("+%s", thermal.CompactNumber(tot.LinesAdded))
	totDelStr := fmt.Sprintf("-%s", thermal.CompactNumber(tot.LinesDeleted))
	totNetStr := fmt.Sprintf("+%s", thermal.CompactNumber(tot.NetLines))
	if tot.NetLines < 0 {
		totNetStr = fmt.Sprintf("-%s", thermal.CompactNumber(-tot.NetLines))
	} else if tot.NetLines == 0 {
		totNetStr = "0"
	}

	totYieldStr := "unmeasured"
	if tot.Status == "MEASURED" {
		totYieldStr = fmt.Sprintf("%s tok/ln", thermal.CompactNumber(int64(tot.TokensPerNet)))
	}

	sb.WriteString(fmt.Sprintf("  %s %s tok  %s  %s / %s lines (net %s)  %s  %s  %s\n\n",
		highlight("Totals:"),
		highlight(thermal.CompactNumber(tot.Tokens)),
		dim("·"),
		green(totAddStr),
		magenta(totDelStr),
		highlight(totNetStr),
		dim("·"),
		highlight(totYieldStr),
		effBadge(tot.Efficiency, tot.Status),
	))

	if rep.Lineage != nil && (rep.Lineage.ForkTokens > 0 || rep.Lineage.ForkSessions > 0) {
		mainlineStr := thermal.CompactNumber(rep.Lineage.MainlineTokens)
		forkStr := thermal.CompactNumber(rep.Lineage.ForkTokens)
		mYieldStr := "unmeasured"
		if rep.Lineage.MainlineYield > 0 {
			mYieldStr = fmt.Sprintf("%s tok/ln", thermal.CompactNumber(int64(rep.Lineage.MainlineYield)))
		}
		sb.WriteString(fmt.Sprintf("  %s\n", highlight("Session Lineage & Branch Churn:")))
		sb.WriteString(fmt.Sprintf("    • Mainline:    %s tok (%d sessions) · %.1f%% · yield: %s\n",
			highlight(mainlineStr),
			rep.Lineage.RootSessions,
			100.0-rep.Lineage.ForkRate,
			green(mYieldStr),
		))
		sb.WriteString(fmt.Sprintf("    • Fork/Branch: %s tok (%d sessions) · %.1f%% exploratory churn\n\n",
			gold(forkStr),
			rep.Lineage.ForkSessions,
			rep.Lineage.ForkRate,
		))
	}

	effScale := fmt.Sprintf("%s %s · %s · %s · %s",
		highlight("Efficiency Scale:"),
		green("[HIGH] <=250 tok/ln"),
		gold("[BALANCED] <=1K tok/ln"),
		magenta("[VERBOSE] >1K tok/ln"),
		dim("[EXPLORATORY] unmeasured"),
	)
	for _, wl := range WrapBullet("  • ", effScale, BoundedCardWidth(2)) {
		sb.WriteString(wl + "\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

// RenderYieldMarkdown formats token yield and code delta metrics into a GitHub PR-ready Markdown block.
func RenderYieldMarkdown(rep thermal.YieldReport, top int) string {
	var sb strings.Builder

	sb.WriteString("### ⚡ Token Yield & Code Delta\n\n")

	if len(rep.Models) == 0 && len(rep.Tools) == 0 && len(rep.Projects) == 0 {
		sb.WriteString("> *No token activity or code deltas found in the selected window.*\n\n")
		return sb.String()
	}

	renderSection := func(title string, rows []thermal.YieldRow) {
		if len(rows) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("#### %s\n\n", title))
		sb.WriteString("| # | Entity / Model | Tokens | +Lines | -Lines | Net Lines | Yield | Efficiency |\n")
		sb.WriteString("|---:|:---|---:|---:|---:|---:|---:|:---|\n")

		shown := len(rows)
		if top > 0 && top < shown {
			shown = top
		}

		for i := 0; i < shown; i++ {
			r := rows[i]

			addStr := fmt.Sprintf("+%s", thermal.CompactNumber(r.LinesAdded))
			delStr := fmt.Sprintf("-%s", thermal.CompactNumber(r.LinesDeleted))
			netStr := fmt.Sprintf("+%s", thermal.CompactNumber(r.NetLines))
			if r.NetLines < 0 {
				netStr = fmt.Sprintf("-%s", thermal.CompactNumber(-r.NetLines))
			} else if r.NetLines == 0 {
				netStr = "0"
			}

			yieldStr := "unmeasured"
			if r.Status == "MEASURED" {
				yieldStr = fmt.Sprintf("%s tok/ln", thermal.CompactNumber(int64(r.TokensPerNet)))
			}

			effStr := "`[EXPLORATORY]`"
			if r.Status == "MEASURED" {
				effStr = fmt.Sprintf("`[%s]`", r.Efficiency)
			}

			sb.WriteString(fmt.Sprintf("| %d | `%s` | %s | %s | %s | %s | %s | %s |\n",
				i+1,
				r.Name,
				thermal.CompactNumber(r.Tokens),
				addStr,
				delStr,
				netStr,
				yieldStr,
				effStr,
			))
		}

		if shown < len(rows) {
			remainder := len(rows) - shown
			sb.WriteString(fmt.Sprintf("\n*... and %d more rows*\n", remainder))
		}
		sb.WriteString("\n")
	}

	if len(rep.Models) > 0 {
		renderSection("Models", rep.Models)
	}
	if len(rep.Projects) > 0 {
		renderSection("Projects", rep.Projects)
	}
	if len(rep.Tools) > 0 {
		renderSection("Tools", rep.Tools)
	}

	if rep.Lineage != nil && (rep.Lineage.ForkTokens > 0 || rep.Lineage.ForkSessions > 0) {
		mYieldStr := "unmeasured"
		if rep.Lineage.MainlineYield > 0 {
			mYieldStr = fmt.Sprintf("%s tok/ln", thermal.CompactNumber(int64(rep.Lineage.MainlineYield)))
		}
		sb.WriteString("#### 🌿 Session Branching & Exploratory Lineage\n\n")
		sb.WriteString(fmt.Sprintf("- **Mainline Work**: `%s tok` across %d sessions (%.1f%%) · Mainline Yield: `%s`\n",
			thermal.CompactNumber(rep.Lineage.MainlineTokens),
			rep.Lineage.RootSessions,
			100.0-rep.Lineage.ForkRate,
			mYieldStr,
		))
		sb.WriteString(fmt.Sprintf("- **Fork / Prototyping**: `%s tok` across %d sessions (%.1f%% exploratory churn)\n\n",
			thermal.CompactNumber(rep.Lineage.ForkTokens),
			rep.Lineage.ForkSessions,
			rep.Lineage.ForkRate,
		))
	}

	sb.WriteString("<details>\n<summary>ℹ️ Token Yield Legend & Efficiency Scale</summary>\n\n")
	sb.WriteString("- **Yield**: Tokens burned per net line of code added. Lower is leaner and more concise.\n")
	sb.WriteString("- `[HIGH]`: <=250 tok/net line (efficient, direct code generation).\n")
	sb.WriteString("- `[BALANCED]`: <=1,000 tok/net line (balanced code and iteration).\n")
	sb.WriteString("- `[VERBOSE]`: >1,000 tok/net line (high conversational or reasoning token volume).\n")
	sb.WriteString("- `[EXPLORATORY]`: Sessions without file changes or zero net delta.\n")
	sb.WriteString("- **Session Lineage**: Distinguishes mainline code generation from exploratory sub-agent forks and review threads, revealing true mainline yield vs prototyping churn.\n")
	sb.WriteString("</details>\n")

	return sb.String()
}
