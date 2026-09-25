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
	alignRight := []bool{false, false, true, true, true, true, true, false}
	widths := []int{yieldRankWidth, yieldEntityWidth, yieldTokensWidth, yieldLinesWidth, yieldLinesWidth, yieldNetWidth, yieldRatioWidth, yieldEffWidth}

	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}

	cardWidth := BoundedCardWidth(2)

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

			yieldStr := dim("unmeasured")
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
			rowB.WriteString(thermal.PadLeft(fmt.Sprintf("%d.", i+1), widths[0]))
			rowB.WriteString("  ")
			rowB.WriteString(thermal.PadRight(truncate(displayName, widths[1]), widths[1]))
			rowB.WriteString("  ")
			rowB.WriteString(thermal.PadLeft(thermal.CompactNumber(r.Tokens), widths[2]))
			rowB.WriteString("  ")
			rowB.WriteString(thermal.PadLeft(addStr, widths[3]))
			rowB.WriteString("  ")
			rowB.WriteString(thermal.PadLeft(delStr, widths[4]))
			rowB.WriteString("  ")
			rowB.WriteString(thermal.PadLeft(netStr, widths[5]))
			rowB.WriteString("  ")
			rowB.WriteString(thermal.PadLeft(yieldStr, widths[6]))
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
			MaxWidth:    cardWidth,
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
