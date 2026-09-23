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
	yieldRankWidth    = 4
	yieldEntityWidth  = 30
	yieldTokensWidth  = 11
	yieldLinesWidth   = 9
	yieldNetWidth     = 10
	yieldRatioWidth   = 13
	yieldEffWidth     = 13
)

// RenderYield renders the token yield and code output delta table.
func RenderYield(rep thermal.YieldReport, top int, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	green := func(s string) string { return ColorCode(colors, "1;32", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }
	magenta := func(s string) string { return ColorCode(colors, "1;35", s) }
	cyan := func(s string) string { return ColorCode(colors, "36", s) }

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

	renderSection := func(title string, rows []thermal.YieldRow) {
		if len(rows) == 0 {
			return
		}
		if title != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", highlight(title)))
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

			sb.WriteString("  ")
			sb.WriteString(thermal.PadLeft(fmt.Sprintf("%d.", i+1), widths[0]))
			sb.WriteString("  ")
			sb.WriteString(thermal.PadRight(truncate(r.Name, widths[1]), widths[1]))
			sb.WriteString("  ")
			sb.WriteString(thermal.PadLeft(thermal.CompactNumber(r.Tokens), widths[2]))
			sb.WriteString("  ")
			sb.WriteString(thermal.PadLeft(addStr, widths[3]))
			sb.WriteString("  ")
			sb.WriteString(thermal.PadLeft(delStr, widths[4]))
			sb.WriteString("  ")
			sb.WriteString(thermal.PadLeft(netStr, widths[5]))
			sb.WriteString("  ")
			sb.WriteString(thermal.PadLeft(yieldStr, widths[6]))
			sb.WriteString("  ")
			sb.WriteString(effBadge(r.Efficiency, r.Status))
			sb.WriteString("\n")
		}

		if rest := len(rows) - shown; rest > 0 {
			sb.WriteString(fmt.Sprintf("  %s\n", dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest))))
		}
		sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n\n")
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

	return sb.String()
}
