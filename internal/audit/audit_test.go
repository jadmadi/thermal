// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditRuleFiles(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()

	// 1. Create a clean PASS rule file (50 lines)
	passFile := filepath.Join(workDir, "CLAUDE.md")
	passContent := strings.Repeat("Rule line content\n", 50)
	if err := os.WriteFile(passFile, []byte(passContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Create a WARN rule file (180 lines)
	warnFile := filepath.Join(workDir, "AGENTS.md")
	warnContent := strings.Repeat("Agent instruction line\n", 180)
	if err := os.WriteFile(warnFile, []byte(warnContent), 0644); err != nil {
		t.Fatal(err)
	}

	rep := RunAuditWithOptions(workDir, homeDir)

	if len(rep.RulesFindings) < 2 {
		t.Fatalf("expected at least 2 rule findings, got %d", len(rep.RulesFindings))
	}

	foundPass := false
	foundWarn := false
	for _, rf := range rep.RulesFindings {
		if filepath.Base(rf.Path) == "CLAUDE.md" {
			foundPass = true
			if rf.Status != "PASS" {
				t.Errorf("CLAUDE.md expected PASS, got %s", rf.Status)
			}
			if rf.Lines != 50 {
				t.Errorf("CLAUDE.md expected 50 lines, got %d", rf.Lines)
			}
		}
		if filepath.Base(rf.Path) == "AGENTS.md" {
			foundWarn = true
			if rf.Status != "WARN" {
				t.Errorf("AGENTS.md expected WARN, got %s", rf.Status)
			}
			if rf.Lines != 180 {
				t.Errorf("AGENTS.md expected 180 lines, got %d", rf.Lines)
			}
		}
	}

	if !foundPass || !foundWarn {
		t.Errorf("missing expected findings: pass=%v, warn=%v", foundPass, foundWarn)
	}

	// Score should be 95 (100 - 5 for 1 WARN)
	if rep.Score != 95 {
		t.Errorf("expected score 95, got %d", rep.Score)
	}
	if rep.Grade != "A+" {
		t.Errorf("expected grade A+, got %s", rep.Grade)
	}
}

func TestAuditCriticalRuleFile(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()

	critFile := filepath.Join(workDir, "AGENTS.md")
	critContent := strings.Repeat("Long bloat rule line\n", 400)
	if err := os.WriteFile(critFile, []byte(critContent), 0644); err != nil {
		t.Fatal(err)
	}

	rep := RunAuditWithOptions(workDir, homeDir)

	var foundCrit bool
	for _, rf := range rep.RulesFindings {
		if filepath.Base(rf.Path) == "AGENTS.md" {
			foundCrit = true
			if rf.Status != "CRIT" {
				t.Errorf("expected CRIT status, got %s", rf.Status)
			}
			if rf.Lines != 400 {
				t.Errorf("expected 400 lines, got %d", rf.Lines)
			}
		}
	}
	if !foundCrit {
		t.Fatal("expected CRIT AGENTS.md finding")
	}

	if rep.Score != 85 {
		t.Errorf("expected score 85 (100 - 15), got %d", rep.Score)
	}
	if rep.Grade != "A" {
		t.Errorf("expected grade A, got %s", rep.Grade)
	}
}

func TestAuditMCPServerSchema(t *testing.T) {
	tmpDir := t.TempDir()
	mcpFile := filepath.Join(tmpDir, "test_mcp.json")
	mcpJSON := `{
		"mcpServers": {
			"server1": {"command": "node"},
			"server2": {"command": "python"}
		}
	}`
	if err := os.WriteFile(mcpFile, []byte(mcpJSON), 0644); err != nil {
		t.Fatal(err)
	}

	rep := AuditReport{Score: 100}
	seen := make(map[string]bool)
	scanMCPConfigFile(mcpFile, "TestClient", &rep, seen)

	if len(rep.MCPFindings) != 2 {
		t.Fatalf("expected 2 MCP findings, got %d", len(rep.MCPFindings))
	}
	if rep.TotalMCPTax != 1700 { // 850 * 2
		t.Errorf("expected 1700 total tax, got %d", rep.TotalMCPTax)
	}
}

func TestAuditAgentReadiness_MissingAgentsMD(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()

	rep := RunAuditWithOptions(workDir, homeDir)

	var foundMissing bool
	for _, rf := range rep.ReadinessFindings {
		if rf.Check == "Instructions" {
			foundMissing = true
			if rf.Status != "WARN" {
				t.Errorf("expected WARN for missing instructions, got %s", rf.Status)
			}
			if !strings.Contains(rf.Message, "No AGENTS.md") {
				t.Errorf("unexpected message: %s", rf.Message)
			}
		}
	}
	if !foundMissing {
		t.Fatal("expected Instructions finding")
	}
}

func TestAuditAgentReadiness_ValidAgentsMDAndAnchoredGitignore(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()

	agentsFile := filepath.Join(workDir, "AGENTS.md")
	if err := os.WriteFile(agentsFile, []byte("# Agents Standards\nRules and guidelines.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	gitignoreFile := filepath.Join(workDir, ".gitignore")
	gitignoreContent := "/thermal\n/dist/\n*.exe\n"
	if err := os.WriteFile(gitignoreFile, []byte(gitignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	rep := RunAuditWithOptions(workDir, homeDir)

	var foundAgents, foundGitignore bool
	for _, rf := range rep.ReadinessFindings {
		if rf.Check == "Instructions" {
			foundAgents = true
			if rf.Status != "PASS" {
				t.Errorf("expected PASS for valid AGENTS.md, got %s", rf.Status)
			}
		}
		if rf.Check == "Gitignore" {
			foundGitignore = true
			if rf.Status != "PASS" {
				t.Errorf("expected PASS for anchored .gitignore, got %s", rf.Status)
			}
		}
	}
	if !foundAgents || !foundGitignore {
		t.Errorf("missing readiness findings: agents=%v, gitignore=%v", foundAgents, foundGitignore)
	}
}

func TestAuditAgentReadiness_BareGitignoreRule(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workDir, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatal(err)
	}

	gitignoreFile := filepath.Join(workDir, ".gitignore")
	gitignoreContent := "thermal\n/dist/\n"
	if err := os.WriteFile(gitignoreFile, []byte(gitignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	rep := RunAuditWithOptions(workDir, homeDir)

	var foundBareRule bool
	for _, rf := range rep.ReadinessFindings {
		if rf.Check == "Gitignore" {
			foundBareRule = true
			if rf.Status != "WARN" {
				t.Errorf("expected WARN for bare binary rule, got %s", rf.Status)
			}
			if !strings.Contains(rf.Message, "Unanchored rule") {
				t.Errorf("unexpected message: %s", rf.Message)
			}
		}
	}
	if !foundBareRule {
		t.Fatal("expected Gitignore finding with bare rule")
	}
}

func TestAuditAgentReadiness_EmptyAgentsMD(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workDir, "AGENTS.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	rep := RunAuditWithOptions(workDir, homeDir)

	var foundEmpty bool
	for _, rf := range rep.ReadinessFindings {
		if rf.Check == "Instructions" {
			foundEmpty = true
			if rf.Status != "WARN" {
				t.Errorf("expected WARN for empty AGENTS.md, got %s", rf.Status)
			}
			if !strings.Contains(rf.Message, "empty") {
				t.Errorf("unexpected message: %s", rf.Message)
			}
		}
	}
	if !foundEmpty {
		t.Fatal("expected Repository Instructions finding for empty file")
	}
}
