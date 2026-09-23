// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jadmadi/thermal/internal/audit"
	"github.com/jadmadi/thermal/internal/thermal"
)

// RenderAudit formats the audit report for terminal display.
func RenderAudit(rep audit.AuditReport, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	faint := func(s string) string { return ColorCode(colors, "38;5;245", s) }
	green := func(s string) string { return ColorCode(colors, "1;32", s) }
	yellow := func(s string) string { return ColorCode(colors, "1;33", s) }
	red := func(s string) string { return ColorCode(colors, "1;31", s) }

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

	// 1. Rule Files
	sb.WriteString(fmt.Sprintf("  %s\n", highlight("Instruction & Rule Files:")))
	if len(rep.RulesFindings) == 0 {
		sb.WriteString(fmt.Sprintf("    %s\n", faint("No local rule files detected.")))
	} else {
		for _, rf := range rep.RulesFindings {
			rel := filepath.Base(rf.Path)
			advice := ""
			if rf.Advice != "" {
				advice = fmt.Sprintf(" (%s)", rf.Advice)
			}
			sb.WriteString(fmt.Sprintf("    %-4s  %-24s  %4d lines  ~%-9s%s\n",
				statusBadge(rf.Status),
				highlight(rel),
				rf.Lines,
				thermal.CompactNumber(int64(rf.EstimatedTokens))+" tok",
				faint(advice),
			))
		}
	}
	sb.WriteString("\n")

	// 2. MCP Servers
	sb.WriteString(fmt.Sprintf("  %s\n", highlight("MCP Server Registrations:")))
	if len(rep.MCPFindings) == 0 {
		sb.WriteString(fmt.Sprintf("    %s\n", faint("No MCP servers configured.")))
	} else {
		for _, mf := range rep.MCPFindings {
			advice := ""
			if mf.Advice != "" {
				advice = fmt.Sprintf(" (%s)", mf.Advice)
			}
			sb.WriteString(fmt.Sprintf("    %-6s  %-24s  %-12s  ~%-9s%s\n",
				statusBadge(mf.Status),
				highlight(mf.ServerName),
				faint(mf.Client),
				thermal.CompactNumber(int64(mf.EstimatedTax))+" tok",
				faint(advice),
			))
		}
	}
	sb.WriteString("\n")

	// 3. Recommendations
	sb.WriteString(fmt.Sprintf("  %s\n", highlight("Actionable Advice:")))
	for _, rec := range rep.Recommendations {
		sb.WriteString(fmt.Sprintf("    • %s\n", rec))
	}
	sb.WriteString("\n")

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
