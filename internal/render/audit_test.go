// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/audit"
)

func TestRenderAudit_PathsScoreAndLongNames(t *testing.T) {
	rep := audit.AuditReport{
		Timestamp:        time.Now(),
		Score:            58,
		Grade:            "D",
		RulePenalty:      30,
		MCPPenalty:       12,
		ReadinessPenalty: 0,
		TotalRulesTax:    34800,
		TotalMCPTax:      10584,
		RulesFindings: []audit.RuleFinding{
			{Path: "/repo/CLAUDE.md", Scope: "repo root", Lines: 204, EstimatedTokens: 6836, Status: "WARN"},
			{Path: "/home/me/.claude/CLAUDE.md", Scope: "~/.claude", Lines: 1, EstimatedTokens: 2, Status: "PASS"},
		},
		MCPFindings: []audit.MCPFinding{
			{Client: "Claude Code", ServerName: "a-very-long-mcp-server-name-that-breaks-layout", EstimatedTax: 900, Status: "WARN"},
		},
		ReadinessFindings: []audit.ReadinessFinding{
			{Check: "Instructions", Status: "PASS", Message: "AGENTS.md verified"},
			{Check: "Tool Databases", Status: "PASS", Message: "3 local tool database files readable (filesystem readability only; SQLite integrity and mmap were not checked)"},
		},
		Recommendations: []string{"Warning: CLAUDE.md (repo root) is 204 lines (~6836 tok). Prune stale rules to save context budget."},
	}
	out := RenderAudit(rep, true)
	for _, want := range []string{
		"CLAUDE.md (repo root)",
		"CLAUDE.md (~/.claude)",
		"Score components: rules -30 · MCP -12 · readiness -0",
		"Biggest win: Warning:",
		"PASS  Tool Databases    ",
		"a-very-long-mcp-server-…",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("audit output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "a-very-long-mcp-server-name-that-breaks-layout") {
		t.Errorf("long MCP name was not truncated:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if width := ansi.StringWidth(line); width > 100 {
			t.Errorf("audit line exceeds card width (%d): %q", width, line)
		}
	}

	var decoded audit.AuditReport
	if err := json.Unmarshal([]byte(RenderAuditJSON(rep)), &decoded); err != nil {
		t.Fatalf("decode audit JSON: %v", err)
	}
	if decoded.RulePenalty != 30 || len(decoded.RulesFindings) != 2 || decoded.RulesFindings[1].Scope != "~/.claude" {
		t.Errorf("audit JSON lost penalty/scope parity: %+v", decoded)
	}
}

func TestRenderAudit_NoColor(t *testing.T) {
	rep := audit.AuditReport{
		Timestamp:     time.Now(),
		Score:         90,
		Grade:         "A",
		TotalRulesTax: 1200,
		TotalMCPTax:   850,
		RulesFindings: []audit.RuleFinding{
			{Path: "CLAUDE.md", Lines: 80, EstimatedTokens: 500, Status: "PASS"},
		},
		MCPFindings: []audit.MCPFinding{
			{Client: "Claude Code", ServerName: "test-server", EstimatedTax: 850, Status: "ACTIVE"},
		},
		ReadinessFindings: []audit.ReadinessFinding{
			{Check: "Repository Instructions", Status: "PASS", Message: "AGENTS.md verified (280 lines)"},
		},
		Recommendations: []string{"Everything looks great."},
	}

	out := RenderAudit(rep, true)
	if strings.Contains(out, "\033[") {
		t.Errorf("expected no ANSI escape sequences with noColor=true")
	}
	if !strings.Contains(out, "Score: [90/100 · Grade A]") {
		t.Errorf("expected score string, got %s", out)
	}
	if !strings.Contains(out, "CLAUDE.md") {
		t.Errorf("expected CLAUDE.md in output")
	}
	if !strings.Contains(out, "test-server") {
		t.Errorf("expected test-server in output")
	}
	if !strings.Contains(out, "Agent Readiness & Repository Hygiene") {
		t.Errorf("expected Agent Readiness & Repository Hygiene card in output")
	}
	if !strings.Contains(out, "AGENTS.md verified") {
		t.Errorf("expected AGENTS.md verified in output")
	}

	jsonOut := RenderAuditJSON(rep)
	var decoded audit.AuditReport
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("failed to parse RenderAuditJSON: %v", err)
	}
	if decoded.Score != 90 {
		t.Errorf("expected decoded score 90, got %d", decoded.Score)
	}
}
