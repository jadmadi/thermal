// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/audit"
)

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

	jsonOut := RenderAuditJSON(rep)
	var decoded audit.AuditReport
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("failed to parse RenderAuditJSON: %v", err)
	}
	if decoded.Score != 90 {
		t.Errorf("expected decoded score 90, got %d", decoded.Score)
	}
}
