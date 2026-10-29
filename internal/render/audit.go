// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/x/ansi"
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

	scoreLine := fmt.Sprintf("Score: [%s/100 · Grade %s]  Context Tax: ≈%s tokens / turn (Rules: ≈%s tok, MCP: ≈%s tok)",
		highlight(fmt.Sprintf("%d", rep.Score)),
		gradeBadge(rep.Grade),
		highlight(thermal.CompactNumber(int64(rep.TotalRulesTax+rep.TotalMCPTax))),
		thermal.CompactNumber(int64(rep.TotalRulesTax)),
		thermal.CompactNumber(int64(rep.TotalMCPTax)),
	)
	for _, line := range WrapText(scoreLine, ContentWidth()-4) {
		sb.WriteString("  " + line + "\n")
	}

	cardWidth := ContentWidth()
	innerWidth := cardWidth - 4

	largest := "none"
	largestPenalty := 0
	for label, penalty := range map[string]int{
		"rule files":  rep.RulePenalty,
		"MCP density": rep.MCPPenalty,
		"readiness":   rep.ReadinessPenalty,
	} {
		if penalty > largestPenalty {
			largest = label
			largestPenalty = penalty
		}
	}
	explanation := fmt.Sprintf("Score components: rules -%d · MCP -%d · readiness -%d. Largest: %s.",
		rep.RulePenalty, rep.MCPPenalty, rep.ReadinessPenalty, largest)
	if len(rep.Recommendations) > 0 {
		explanation += " Biggest win: " + rep.Recommendations[0]
	}
	explanationWidth := cardWidth
	if cardWidth < 80 {
		explanationWidth = cardWidth - 4
	}
	for _, line := range WrapText(explanation, explanationWidth) {
		sb.WriteString(fmt.Sprintf("  %s\n", dim(line)))
	}
	sb.WriteString("\n")

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
			label := filepath.Base(rf.Path)
			if rf.Scope != "" {
				label += " (" + rf.Scope + ")"
			}
			label = truncateLabel(label, 30)
			status := padRightStyled(rf.Status, 4, func(s string) string { return statusBadge(s) })
			labelCell := padRightStyled(label, 30, highlight)
			cardLinesRules = append(cardLinesRules, fmt.Sprintf(" %s  %s  %4d lines  ≈%-9s",
				status,
				labelCell,
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
		Width:       CardWidthFor(cardWidth),
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
			status := padRightStyled(mf.Status, 6, func(s string) string { return statusBadge(s) })
			server := padRightStyled(truncateLabel(mf.ServerName, 24), 24, highlight)
			client := padRightStyled(truncateLabel(mf.Client, 12), 12, faint)
			cardLinesMCP = append(cardLinesMCP, fmt.Sprintf(" %s  %s  %s  ≈%-9s",
				status,
				server,
				client,
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
		Width:       CardWidthFor(cardWidth),
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Secondary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	// 3. Agent Readiness & Repository Hygiene
	var cardLinesReadiness []string
	if len(rep.ReadinessFindings) == 0 {
		cardLinesReadiness = append(cardLinesReadiness, " "+faint("No readiness checks recorded."))
	} else {
		checkWidth := 16
		for _, rf := range rep.ReadinessFindings {
			if len(rf.Check) > checkWidth {
				checkWidth = len(rf.Check)
			}
		}
		if checkWidth > 28 {
			checkWidth = 28
		}
		for _, rf := range rep.ReadinessFindings {
			badge := padRightStyled(rf.Status, 4, func(s string) string { return statusBadge(s) })
			chk := padRightStyled(truncateLabel(rf.Check, checkWidth), checkWidth, highlight)
			prefix := fmt.Sprintf(" %s  %s  ", badge, chk)
			prefixWidth := ansi.StringWidth(prefix)
			avail := innerWidth - prefixWidth
			if avail < 12 {
				avail = 12
			}
			msgLines := WrapText(rf.Message, avail)
			for i, line := range msgLines {
				if i == 0 {
					cardLinesReadiness = append(cardLinesReadiness, prefix+line)
				} else {
					cardLinesReadiness = append(cardLinesReadiness, strings.Repeat(" ", prefixWidth)+line)
				}
			}
			if rf.Advice != "" {
				advicePrefix := strings.Repeat(" ", prefixWidth) + faint("↳ ")
				adviceLines := WrapText(rf.Advice, innerWidth-prefixWidth-2)
				for i, line := range adviceLines {
					if i == 0 {
						cardLinesReadiness = append(cardLinesReadiness, advicePrefix+faint(line))
					} else {
						cardLinesReadiness = append(cardLinesReadiness, strings.Repeat(" ", prefixWidth+2)+faint(line))
					}
				}
			}
		}
	}
	sb.WriteString(RenderCard(CardOptions{
		Title:       "Agent Readiness & Repository Hygiene",
		RightHeader: "environment check",
		Lines:       cardLinesReadiness,
		Width:       CardWidthFor(cardWidth),
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	// 4. Recommendations
	if len(rep.Recommendations) > 0 {
		var cardLinesRec []string
		for _, rec := range rep.Recommendations {
			cardLinesRec = append(cardLinesRec, WrapBullet(" • ", rec, innerWidth)...)
		}
		sb.WriteString(RenderCard(CardOptions{
			Title:       "Actionable Advice",
			RightHeader: "optimizations",
			Lines:       cardLinesRec,
			Width:       CardWidthFor(cardWidth),
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
