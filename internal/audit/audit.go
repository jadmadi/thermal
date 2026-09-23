// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

const maxFileSize = 1024 * 1024 // 1MB scan ceiling

// RuleFinding captures context overhead metrics for an AI instructions or rules file.
type RuleFinding struct {
	Path            string `json:"path"`
	Lines           int    `json:"lines"`
	EstimatedTokens int    `json:"estimatedTokens"`
	Status          string `json:"status"` // "PASS", "WARN", "CRIT"
	Advice          string `json:"advice,omitempty"`
}

// MCPFinding captures overhead metrics for an installed or configured MCP server.
type MCPFinding struct {
	Client       string `json:"client"` // e.g. "Claude Code", "Claude Desktop", "Antigravity", "Cursor"
	ServerName   string `json:"serverName"`
	Command      string `json:"command,omitempty"`
	EstimatedTax int    `json:"estimatedTaxTokens"`
	Status       string `json:"status"` // "ACTIVE", "WARN"
	Advice       string `json:"advice,omitempty"`
}

// AuditReport aggregates all local setup and context hygiene diagnostics.
type AuditReport struct {
	Timestamp       time.Time     `json:"timestamp"`
	Score           int           `json:"score"`
	Grade           string        `json:"grade"`
	TotalRulesTax   int           `json:"totalRulesTaxTokens"`
	TotalMCPTax     int           `json:"totalMCPTaxTokens"`
	RulesFindings   []RuleFinding `json:"rules"`
	MCPFindings     []MCPFinding  `json:"mcp"`
	Recommendations []string      `json:"recommendations"`
}

// RunAudit performs a non-destructive, read-only diagnostic of local agent configurations,
// instructions files, and MCP registrations.
func RunAudit(workDir string) AuditReport {
	return RunAuditWithOptions(workDir, "")
}

// RunAuditWithOptions allows overriding workDir and homeDir for deterministic testing.
func RunAuditWithOptions(workDir, homeDir string) AuditReport {
	if workDir == "" {
		if wd, err := os.Getwd(); err == nil {
			workDir = wd
		}
	}
	if homeDir == "" {
		homeDir = thermal.HomeDir()
	}

	report := AuditReport{
		Timestamp:       time.Now(),
		Score:           100,
		RulesFindings:   []RuleFinding{},
		MCPFindings:     []MCPFinding{},
		Recommendations: []string{},
	}

	// 1. Audit Rule Files
	auditRuleFiles(workDir, homeDir, &report)

	// 2. Audit MCP Servers
	auditMCPServers(homeDir, &report)

	// 3. Compute Grade & Recommendations
	computeGradeAndRecommendations(&report)

	return report
}

// auditRuleFiles searches current workspace and home directory for AI rule files.
func auditRuleFiles(workDir, homeDir string, rep *AuditReport) {
	ruleNames := []string{
		"CLAUDE.md",
		"AGENTS.md",
		".cursorrules",
		"GEMINI.md",
		".windsurfrules",
		".copilot-instructions.md",
	}

	seenPaths := make(map[string]bool)

	// Check project work directory
	for _, name := range ruleNames {
		p := filepath.Join(workDir, name)
		if checkRuleFile(p, rep, seenPaths) {
			seenPaths[p] = true
		}
	}

	// Check .cursor/rules directory if present
	cursorRulesDir := filepath.Join(workDir, ".cursor", "rules")
	if entries, err := os.ReadDir(cursorRulesDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".mdc") {
				p := filepath.Join(cursorRulesDir, e.Name())
				if checkRuleFile(p, rep, seenPaths) {
					seenPaths[p] = true
				}
			}
		}
	}

	// Check user home config directories
	if homeDir != "" && homeDir != workDir {
		homeRules := []string{
			filepath.Join(homeDir, ".claude", "CLAUDE.md"),
			filepath.Join(homeDir, ".gemini", "GEMINI.md"),
			filepath.Join(homeDir, ".agents", "AGENTS.md"),
		}
		for _, p := range homeRules {
			if checkRuleFile(p, rep, seenPaths) {
				seenPaths[p] = true
			}
		}
	}
}

func checkRuleFile(path string, rep *AuditReport, seen map[string]bool) bool {
	if seen[path] {
		return false
	}
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() || fi.Size() == 0 {
		return false
	}
	if fi.Size() > maxFileSize {
		rep.RulesFindings = append(rep.RulesFindings, RuleFinding{
			Path:            path,
			Lines:           9999,
			EstimatedTokens: int(fi.Size() / 4),
			Status:          "CRIT",
			Advice:          "File exceeds 1MB ceiling. Extreme context penalty.",
		})
		rep.Score -= 25
		return true
	}

	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	lineCount := 0
	byteCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineCount++
		byteCount += len(scanner.Bytes()) + 1
	}

	tokens := byteCount / 4
	if tokens == 0 && byteCount > 0 {
		tokens = 1
	}

	finding := RuleFinding{
		Path:            path,
		Lines:           lineCount,
		EstimatedTokens: tokens,
		Status:          "PASS",
	}

	if lineCount > 300 {
		finding.Status = "CRIT"
		finding.Advice = fmt.Sprintf("Rule file exceeds 300 lines (%d lines, ~%d tok). High initial context tax.", lineCount, tokens)
		rep.Score -= 15
	} else if lineCount > 150 {
		finding.Status = "WARN"
		finding.Advice = fmt.Sprintf("Rule file is moderately large (%d lines, ~%d tok). Consider keeping under 150 lines.", lineCount, tokens)
		rep.Score -= 5
	}

	rep.TotalRulesTax += tokens
	rep.RulesFindings = append(rep.RulesFindings, finding)
	return true
}

type mcpConfigSchema struct {
	MCPServers map[string]struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	} `json:"mcpServers"`
}

// auditMCPServers inspects Claude, Gemini, and Cursor MCP registrations.
func auditMCPServers(homeDir string, rep *AuditReport) {
	if homeDir == "" {
		return
	}

	seenServers := make(map[string]bool)

	// 1. Claude CLI / Desktop (~/.claude.json or ~/.config/Claude/claude_desktop_config.json)
	claudePaths := []string{
		filepath.Join(homeDir, ".claude.json"),
		filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json"),
		filepath.Join(homeDir, ".claude", "claude_desktop_config.json"),
	}
	for _, cp := range claudePaths {
		scanMCPConfigFile(cp, "Claude", rep, seenServers)
	}

	// 2. Cursor MCP (~/.cursor/mcp.json)
	cursorPaths := []string{
		filepath.Join(homeDir, ".cursor", "mcp.json"),
	}
	for _, cp := range cursorPaths {
		scanMCPConfigFile(cp, "Cursor", rep, seenServers)
	}

	// 3. Antigravity CLI MCP (~/.gemini/antigravity-cli/mcp/)
	agyMcpDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "mcp")
	if entries, err := os.ReadDir(agyMcpDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				sName := e.Name()
				if !seenServers[sName] {
					seenServers[sName] = true
					const tax = 900 // schema tokens
					rep.TotalMCPTax += tax
					rep.MCPFindings = append(rep.MCPFindings, MCPFinding{
						Client:       "Antigravity",
						ServerName:   sName,
						EstimatedTax: tax,
						Status:       "ACTIVE",
					})
				}
			}
		}
	}

	// Penalize excessive MCP overhead
	mcpCount := len(rep.MCPFindings)
	if mcpCount > 8 {
		penalty := (mcpCount - 8) * 3
		rep.Score -= penalty
		for i := range rep.MCPFindings {
			if i >= 8 {
				rep.MCPFindings[i].Status = "WARN"
				rep.MCPFindings[i].Advice = "High server density. System prompt consumes schema overhead on every turn."
			}
		}
	}
}

func scanMCPConfigFile(path, client string, rep *AuditReport, seen map[string]bool) {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() || fi.Size() == 0 || fi.Size() > maxFileSize {
		return
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var conf mcpConfigSchema
	if err := json.Unmarshal(b, &conf); err != nil {
		return
	}

	for sName, srv := range conf.MCPServers {
		key := fmt.Sprintf("%s:%s", client, sName)
		if seen[key] {
			continue
		}
		seen[key] = true

		const tax = 850
		rep.TotalMCPTax += tax
		rep.MCPFindings = append(rep.MCPFindings, MCPFinding{
			Client:       client,
			ServerName:   sName,
			Command:      srv.Command,
			EstimatedTax: tax,
			Status:       "ACTIVE",
		})
	}
}

func computeGradeAndRecommendations(rep *AuditReport) {
	if rep.Score < 0 {
		rep.Score = 0
	}
	if rep.Score > 100 {
		rep.Score = 100
	}

	switch {
	case rep.Score >= 95:
		rep.Grade = "A+"
	case rep.Score >= 85:
		rep.Grade = "A"
	case rep.Score >= 75:
		rep.Grade = "B"
	case rep.Score >= 60:
		rep.Grade = "C"
	case rep.Score >= 45:
		rep.Grade = "D"
	default:
		rep.Grade = "F"
	}

	// Build recommendations
	var recs []string
	for _, rf := range rep.RulesFindings {
		if rf.Status == "CRIT" {
			recs = append(recs, fmt.Sprintf("Critical: %s is %d lines (~%d tok). Split into specialized skills or trim to <150 lines.", filepath.Base(rf.Path), rf.Lines, rf.EstimatedTokens))
		} else if rf.Status == "WARN" {
			recs = append(recs, fmt.Sprintf("Warning: %s is %d lines (~%d tok). Prune stale rules to save context budget.", filepath.Base(rf.Path), rf.Lines, rf.EstimatedTokens))
		}
	}

	if len(rep.MCPFindings) > 8 {
		recs = append(recs, fmt.Sprintf("You have %d MCP servers configured (~%d tok total context tax). Unload unused servers to accelerate turn latency.", len(rep.MCPFindings), rep.TotalMCPTax))
	}

	if len(recs) == 0 {
		recs = append(recs, "Excellent setup! All instruction rules and MCP servers are lean and well within recommended token budgets.")
	}

	rep.Recommendations = recs
}
