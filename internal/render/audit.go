// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jadmadi/thermal/internal/audit"
	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

// RenderAudit formats the audit report for terminal display.
func RenderAudit(rep audit.AuditReport, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	faint := st.Faint
	green := st.Success
	yellow := st.Warning
	red := st.Error

	statusBadge := func(status string) string {
		switch status {
		case "PASS":
			return green("PASS")
		case "WARN":
			return yellow("WARN")
		case "CRIT":
			return red("CRIT")
		default:
			return faint(status)
		}
	}

	gradeBadge := func(grade string) string {
		switch grade {
		case "A+", "A":
			return green(grade)
		case "B":
			return yellow(grade)
		default:
			return red(grade)
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("audit · local setup & context health"),
	))

	sb.WriteString(fmt.Sprintf("  Score: [%s/100 · Grade %s]  Context Tax: ~%s tokens / turn (Rules: ~%s tok, MCP: ~%s tok)\n\n",
		highlight(fmt.Sprintf("%d", rep.Score)),
		gradeBadge(rep.Grade),
		highlight(thermal.CompactNumber(int64(rep.TotalRulesTax+rep.TotalMCPTax))),
		thermal.CompactNumber(int64(rep.TotalRulesTax)),
		thermal.CompactNumber(int64(rep.TotalMCPTax)),
	))

	cardWidth := BoundedCardWidth(2)
	innerWidth := cardWidth - 4

	// 1. Rule Files
	var cardLinesRules []string
	if len(rep.RulesFindings) == 0 {
		cardLinesRules = append(cardLinesRules, " "+faint("No local rule files detected."))
	} else {
		hasWarn := false
		hasCrit := false
		for _, rf := range rep.RulesFindings {
			if rf.Status == "WARN" {
				hasWarn = true
			} else if rf.Status == "CRIT" {
				hasCrit = true
			}
			rel := filepath.Base(rf.Path)
			cardLinesRules = append(cardLinesRules, fmt.Sprintf(" %-4s  %-24s  %4d lines  ~%-9s",
				statusBadge(rf.Status),
				highlight(rel),
				rf.Lines,
				thermal.CompactNumber(int64(rf.EstimatedTokens))+" tok",
			))
		}
		if hasWarn || hasCrit {
			var legend string
			if hasWarn && hasCrit {
				legend = "Legend: [WARN] >150 lines (moderate tax) · [CRIT] >300 lines (high tax)"
			} else if hasCrit {
				legend = "Legend: [CRIT] >300 lines or >1MB (high initial context tax)"
			} else {
				legend = "Legend: [WARN] >150 lines (moderate context tax)"
			}
			cardLinesRules = append(cardLinesRules, " "+dim(strings.Repeat("─", innerWidth-2)))
			cardLinesRules = append(cardLinesRules, " "+faint(legend))
		}
	}
	sb.WriteString(RenderCard(CardOptions{
		Title:       "Instruction & Rule Files",
		RightHeader: "rule overhead",
		Lines:       cardLinesRules,
		Width:       cardWidth,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	// 2. MCP Servers
	var cardLinesMCP []string
	if len(rep.MCPFindings) == 0 {
		cardLinesMCP = append(cardLinesMCP, " "+faint("No MCP servers configured."))
	} else {
		hasWarn := false
		for _, mf := range rep.MCPFindings {
			if mf.Status == "WARN" {
				hasWarn = true
			}
			cardLinesMCP = append(cardLinesMCP, fmt.Sprintf(" %-6s  %-24s  %-12s  ~%-9s",
				statusBadge(mf.Status),
				highlight(mf.ServerName),
				faint(mf.Client),
				thermal.CompactNumber(int64(mf.EstimatedTax))+" tok",
			))
		}
		if hasWarn {
			legend := "Legend: [WARN] High server density (>8 servers) adds schema overhead to every turn."
			cardLinesMCP = append(cardLinesMCP, " "+dim(strings.Repeat("─", innerWidth-2)))
			cardLinesMCP = append(cardLinesMCP, " "+faint(legend))
		}
	}
	sb.WriteString(RenderCard(CardOptions{
		Title:       "MCP Server Registrations",
		RightHeader: "schema tax",
		Lines:       cardLinesMCP,
		Width:       cardWidth,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Secondary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	// 3. Recommendations
	if len(rep.Recommendations) > 0 {
		var cardLinesRec []string
		for _, rec := range rep.Recommendations {
			cardLinesRec = append(cardLinesRec, WrapBullet(" • ", rec, innerWidth)...)
		}
		sb.WriteString(RenderCard(CardOptions{
			Title:       "Actionable Advice",
			RightHeader: "optimizations",
			Lines:       cardLinesRec,
			Width:       cardWidth,
			Indent:      2,
			Colors:      colors,
			TitleColor:  theme.Accent,
			BorderColor: theme.Border,
		}))
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderAuditJSON returns the audit report as indented JSON.
func RenderAuditJSON(rep audit.AuditReport) string {
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
