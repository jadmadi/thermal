// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EvidenceKind designates the category of factual verification observed in a session.
type EvidenceKind int

const (
	EvidenceUnknown EvidenceKind = iota
	EvidenceTestPass
	EvidenceTestFail
	EvidenceLinterPass
	EvidenceLinterFail
	EvidenceGitCommit
	EvidenceBuildPass
	EvidenceBuildFail
	EvidenceAgentClaim
)

// Evidence represents a sanitized quantitative verification event.
type Evidence struct {
	Kind  EvidenceKind
	Label string
}

// VerificationTier classifies session outcomes according to the Factual Evidence Hierarchy.
type VerificationTier string

const (
	Tier1Verified   VerificationTier = "Tier 1 (Verified)"
	Tier2Claimed    VerificationTier = "Tier 2 (Claimed)"
	Tier3Failed     VerificationTier = "Tier 3 (Failed)"
	Tier3Unverified VerificationTier = "Tier 3 (Unverified)"
)

// WorkReceipt links agent token spend to concrete verification outcomes.
type WorkReceipt struct {
	SessionID       string           `json:"sessionId"`
	Tool            string           `json:"tool"`
	Project         string           `json:"project,omitempty"`
	Day             string           `json:"day"`
	Tier            VerificationTier `json:"tier"`
	Status          string           `json:"status"` // VERIFIED, CLAIMED, FAILED, UNVERIFIED
	Tokens          int64            `json:"tokens"`
	Cost            float64          `json:"cost"`
	EstimatedCost   bool             `json:"estimatedCost,omitempty"`
	TestsPassed     int              `json:"testsPassed"`
	TestsFailed     int              `json:"testsFailed"`
	LintersPassed   int              `json:"lintersPassed"`
	LintersFailed   int              `json:"lintersFailed"`
	CommitsCreated  int              `json:"commitsCreated"`
	EvidenceSummary []string         `json:"evidenceSummary,omitempty"`
	DurationMs      int64            `json:"durationMs,omitempty"`
}

// SessionOutcome is the result of outcome parser evaluation for a session.
type SessionOutcome struct {
	Status          string
	Tier            VerificationTier
	TestsPassed     int
	TestsFailed     int
	LintersPassed   int
	LintersFailed   int
	CommitsCreated  int
	EvidenceSummary []string
}

// ReceiptSummary contains aggregate FinOps and verification KPIs across receipts.
type ReceiptSummary struct {
	TotalReceipts     int     `json:"totalReceipts"`
	VerifiedCount     int     `json:"verifiedCount"`
	ClaimedCount      int     `json:"claimedCount"`
	FailedCount       int     `json:"failedCount"`
	UnverifiedCount   int     `json:"unverifiedCount"`
	VerificationRate  float64 `json:"verificationRate"`  // percentage 0.0 - 100.0
	TotalTokens       int64   `json:"totalTokens"`
	TokensVerified    int64   `json:"tokensVerified"`
	TokensUnverified  int64   `json:"tokensUnverified"`
	TotalCost         float64 `json:"totalCost"`
	CostVerified      float64 `json:"costVerified"`
	CostUnverified    float64 `json:"costUnverified"`
	VerifiedTokenRate float64 `json:"verifiedTokenRate"` // percentage 0.0 - 100.0
	SpendEfficiency   string  `json:"spendEfficiency"`   // EXCELLENT, HEALTHY, SPECULATIVE, EXPLORATORY
}

// ReceiptReport is the top-level payload for the 'thermal receipt' command.
type ReceiptReport struct {
	Type     string         `json:"type"` // "receipt"
	Tool     string         `json:"tool,omitempty"`
	Receipts []WorkReceipt  `json:"receipts"`
	Summary  ReceiptSummary `json:"summary"`
}

// ReceiptOptions controls filtering, sorting, and windowing for work receipts.
type ReceiptOptions struct {
	Since string
	Until string
	Last  int
	Sort  string // "tokens", "verified", "cost", "rate"; default "tokens"
	Top   int
	Now   time.Time
}

// ParseCommandEvidence inspects a command string, its exit code, and output snippet
// to extract sanitized verification evidence without preserving sensitive paths or prompts.
func ParseCommandEvidence(cmd string, exitCode int, output string) (EvidenceKind, string, bool) {
	cmdClean := strings.TrimSpace(cmd)
	if strings.HasPrefix(cmdClean, "\"") && strings.HasSuffix(cmdClean, "\"") && len(cmdClean) >= 2 {
		cmdClean = strings.TrimSpace(cmdClean[1 : len(cmdClean)-1])
	}
	cmdLower := strings.ToLower(cmdClean)
	outLower := strings.ToLower(output)

	// Resolve effective exit code if unspecified (-1)
	effCode := exitCode
	if effCode < 0 {
		if strings.Contains(outLower, "the command exited with code 0") ||
			strings.Contains(outLower, "exit code 0") ||
			strings.Contains(outLower, "ok\t") ||
			strings.Contains(outLower, "pass") ||
			strings.Contains(outLower, "passed") ||
			strings.Contains(outLower, "0 failed") {
			effCode = 0
		} else if strings.Contains(outLower, "the command exited with code") ||
			strings.Contains(outLower, "exit code 1") ||
			strings.Contains(outLower, "fail") ||
			strings.Contains(outLower, "failed") ||
			strings.Contains(outLower, "error:") {
			effCode = 1
		} else {
			effCode = 0 // Default to zero if no error string is detected
		}
	}

	// 1. Test runner patterns
	var testName string
	switch {
	case strings.HasPrefix(cmdLower, "go test") || strings.Contains(cmdLower, " go test"):
		testName = "go test"
	case strings.HasPrefix(cmdLower, "cargo test") || strings.Contains(cmdLower, " cargo test"):
		testName = "cargo test"
	case strings.HasPrefix(cmdLower, "pytest") || strings.Contains(cmdLower, " pytest") ||
		strings.Contains(cmdLower, "python -m pytest") || strings.Contains(cmdLower, "python3 -m pytest") ||
		strings.Contains(cmdLower, "python -m unittest"):
		testName = "pytest"
	case strings.HasPrefix(cmdLower, "npm test") || strings.HasPrefix(cmdLower, "npm t ") ||
		strings.Contains(cmdLower, "pnpm test") || strings.Contains(cmdLower, "yarn test") ||
		strings.Contains(cmdLower, "bun test") || strings.Contains(cmdLower, "jest") ||
		strings.Contains(cmdLower, "vitest") || strings.Contains(cmdLower, "mocha"):
		testName = "npm test"
	case strings.HasPrefix(cmdLower, "make test") || strings.HasPrefix(cmdLower, "ctest"):
		testName = "make test"
	case strings.Contains(cmdLower, "rspec"):
		testName = "rspec"
	case strings.Contains(cmdLower, "mvn test") || strings.Contains(cmdLower, "gradle test"):
		testName = "mvn test"
	}

	if testName != "" {
		if effCode == 0 {
			return EvidenceTestPass, fmt.Sprintf("%s (exit 0)", testName), true
		}
		return EvidenceTestFail, fmt.Sprintf("%s (exit %d)", testName, effCode), true
	}

	// 2. Linters & Checkers
	var linterName string
	switch {
	case strings.Contains(cmdLower, "golangci-lint") || strings.Contains(cmdLower, "go vet") ||
		strings.Contains(cmdLower, "staticcheck") || strings.Contains(cmdLower, "revive"):
		linterName = "golangci-lint"
	case strings.Contains(cmdLower, "eslint") || strings.Contains(cmdLower, "biome") ||
		strings.Contains(cmdLower, "prettier --check"):
		linterName = "eslint"
	case strings.Contains(cmdLower, "ruff") || strings.Contains(cmdLower, "flake8") ||
		strings.Contains(cmdLower, "mypy") || strings.Contains(cmdLower, "pyright"):
		linterName = "ruff"
	case strings.Contains(cmdLower, "cargo clippy") || strings.Contains(cmdLower, "cargo check"):
		linterName = "cargo clippy"
	case strings.Contains(cmdLower, "tsc") || strings.Contains(cmdLower, "npm run lint") ||
		strings.Contains(cmdLower, "pnpm lint") || strings.Contains(cmdLower, "yarn lint"):
		linterName = "tsc"
	}

	if linterName != "" {
		if effCode == 0 {
			return EvidenceLinterPass, fmt.Sprintf("%s (exit 0)", linterName), true
		}
		return EvidenceLinterFail, fmt.Sprintf("%s (exit %d)", linterName, effCode), true
	}

	// 3. Git commit & merge creation
	if strings.Contains(cmdLower, "git commit") || strings.Contains(cmdLower, "git merge") || strings.Contains(cmdLower, "git tag") {
		if effCode == 0 {
			return EvidenceGitCommit, "git commit (exit 0)", true
		}
	}

	// 4. Build commands
	if strings.Contains(cmdLower, "go build") || strings.Contains(cmdLower, "cargo build") || strings.Contains(cmdLower, "npm run build") {
		if effCode == 0 {
			return EvidenceBuildPass, "build (exit 0)", true
		}
		return EvidenceBuildFail, fmt.Sprintf("build (exit %d)", effCode), true
	}

	return EvidenceUnknown, "", false
}

// EvaluateSessionOutcome computes the factual evidence tier and verification status
// for a session given its observed command evidence and agent claims.
func EvaluateSessionOutcome(evList []Evidence, agentClaimed bool) SessionOutcome {
	var out SessionOutcome
	seenLabels := make(map[string]bool)

	for _, ev := range evList {
		switch ev.Kind {
		case EvidenceTestPass:
			out.TestsPassed++
			if !seenLabels[ev.Label] {
				out.EvidenceSummary = append(out.EvidenceSummary, ev.Label)
				seenLabels[ev.Label] = true
			}
		case EvidenceTestFail:
			out.TestsFailed++
			if !seenLabels[ev.Label] {
				out.EvidenceSummary = append(out.EvidenceSummary, ev.Label)
				seenLabels[ev.Label] = true
			}
		case EvidenceLinterPass:
			out.LintersPassed++
			if !seenLabels[ev.Label] {
				out.EvidenceSummary = append(out.EvidenceSummary, ev.Label)
				seenLabels[ev.Label] = true
			}
		case EvidenceLinterFail:
			out.LintersFailed++
			if !seenLabels[ev.Label] {
				out.EvidenceSummary = append(out.EvidenceSummary, ev.Label)
				seenLabels[ev.Label] = true
			}
		case EvidenceGitCommit:
			out.CommitsCreated++
			if !seenLabels[ev.Label] {
				out.EvidenceSummary = append(out.EvidenceSummary, ev.Label)
				seenLabels[ev.Label] = true
			}
		}
	}

	// Factual Evidence Hierarchy:
	// Tier 1 (Verified): Observed zero-exit test/linter runs or git commits with no failures.
	// Tier 2 (Claimed): Agent completion claims or command executions without tests.
	// Tier 3 (Failed/Unverified): Non-zero test/linter runs, or interrupted sessions.
	if out.TestsPassed > 0 || out.LintersPassed > 0 || out.CommitsCreated > 0 {
		if out.TestsFailed == 0 && out.LintersFailed == 0 {
			out.Status = "VERIFIED"
			out.Tier = Tier1Verified
		} else {
			out.Status = "FAILED"
			out.Tier = Tier3Failed
		}
	} else if out.TestsFailed > 0 || out.LintersFailed > 0 {
		out.Status = "FAILED"
		out.Tier = Tier3Failed
	} else if agentClaimed || len(evList) > 0 {
		out.Status = "CLAIMED"
		out.Tier = Tier2Claimed
	} else {
		out.Status = "UNVERIFIED"
		out.Tier = Tier3Unverified
	}

	return out
}

// CalculateSpendEfficiency computes the spend efficiency rating based on verified token volume.
func CalculateSpendEfficiency(verifiedRate float64, totalTokens int64) string {
	if totalTokens == 0 {
		return "EXPLORATORY"
	}
	switch {
	case verifiedRate >= 80.0:
		return "EXCELLENT"
	case verifiedRate >= 50.0:
		return "HEALTHY"
	default:
		return "SPECULATIVE"
	}
}

// AggregateReceipts filters, sorts, and compiles work receipts into a full ReceiptReport.
func AggregateReceipts(receipts []WorkReceipt, opts ReceiptOptions) ReceiptReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	hasWindow := opts.Since != "" || opts.Until != "" || opts.Last > 0

	var filtered []WorkReceipt
	var (
		totalTokens      int64
		tokensVerified   int64
		tokensUnverified int64
		totalCost        float64
		costVerified     float64
		costUnverified   float64
		verifiedCount    int
		claimedCount     int
		failedCount      int
		unverifiedCount  int
	)

	for _, r := range receipts {
		if hasWindow {
			t, ok := ParseDay(r.Day)
			if !ok || !inWindow(t, since, until, lastStart) {
				continue
			}
		}

		filtered = append(filtered, r)

		totalTokens += r.Tokens
		totalCost += r.Cost

		switch r.Status {
		case "VERIFIED":
			verifiedCount++
			tokensVerified += r.Tokens
			costVerified += r.Cost
		case "CLAIMED":
			claimedCount++
			tokensUnverified += r.Tokens
			costUnverified += r.Cost
		case "FAILED":
			failedCount++
			tokensUnverified += r.Tokens
			costUnverified += r.Cost
		default:
			unverifiedCount++
			tokensUnverified += r.Tokens
			costUnverified += r.Cost
		}
	}

	// Sort receipts according to requested sort key
	sortKey := strings.ToLower(opts.Sort)
	if sortKey == "" {
		sortKey = "tokens"
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		switch sortKey {
		case "verified":
			tierRank := func(t VerificationTier) int {
				switch t {
				case Tier1Verified:
					return 1
				case Tier2Claimed:
					return 2
				case Tier3Failed:
					return 3
				default:
					return 4
				}
			}
			if tierRank(filtered[i].Tier) != tierRank(filtered[j].Tier) {
				return tierRank(filtered[i].Tier) < tierRank(filtered[j].Tier)
			}
			return filtered[i].Tokens > filtered[j].Tokens
		case "cost":
			if filtered[i].Cost != filtered[j].Cost {
				return filtered[i].Cost > filtered[j].Cost
			}
			return filtered[i].Tokens > filtered[j].Tokens
		case "rate":
			if filtered[i].TestsPassed != filtered[j].TestsPassed {
				return filtered[i].TestsPassed > filtered[j].TestsPassed
			}
			return filtered[i].Tokens > filtered[j].Tokens
		default: // "tokens"
			if filtered[i].Tokens != filtered[j].Tokens {
				return filtered[i].Tokens > filtered[j].Tokens
			}
			return filtered[i].Day > filtered[j].Day
		}
	})

	totalReceipts := len(filtered)
	var verificationRate float64
	if totalReceipts > 0 {
		verificationRate = (float64(verifiedCount) / float64(totalReceipts)) * 100.0
	}

	var verifiedTokenRate float64
	if totalTokens > 0 {
		verifiedTokenRate = (float64(tokensVerified) / float64(totalTokens)) * 100.0
	}

	spendEfficiency := CalculateSpendEfficiency(verifiedTokenRate, totalTokens)

	if opts.Top > 0 && opts.Top < len(filtered) {
		filtered = filtered[:opts.Top]
	}

	return ReceiptReport{
		Type:     "receipt",
		Receipts: filtered,
		Summary: ReceiptSummary{
			TotalReceipts:     totalReceipts,
			VerifiedCount:     verifiedCount,
			ClaimedCount:      claimedCount,
			FailedCount:       failedCount,
			UnverifiedCount:   unverifiedCount,
			VerificationRate:  verificationRate,
			TotalTokens:       totalTokens,
			TokensVerified:    tokensVerified,
			TokensUnverified:  tokensUnverified,
			TotalCost:         totalCost,
			CostVerified:      costVerified,
			CostUnverified:    costUnverified,
			VerifiedTokenRate: verifiedTokenRate,
			SpendEfficiency:   spendEfficiency,
		},
	}
}

// ScanSessionReceipts scans local agent directories (Claude, Agy, CodeWhale, Codex, OpenCode)
// to extract factual work receipts from session transcripts and records.
func ScanSessionReceipts(homeDir string, toolFilter string, pricer Pricer) []WorkReceipt {
	var receipts []WorkReceipt
	filter := strings.ToLower(toolFilter)
	includeAll := filter == "" || filter == "all" || filter == "auto"

	// 1. Scan Claude transcripts (~/.claude/projects/*/*.jsonl)
	if includeAll || filter == "claude" {
		claudeDir := filepath.Join(homeDir, ".claude", "projects")
		if files, err := filepath.Glob(filepath.Join(claudeDir, "*", "*.jsonl")); err == nil {
			for _, file := range files {
				if r, ok := parseClaudeReceipt(file, pricer); ok {
					receipts = append(receipts, r)
				}
			}
		}
	}

	// 2. Scan Agy transcripts (~/.gemini/antigravity-cli/brain/*/.system_generated/logs/transcript.jsonl)
	if includeAll || filter == "agy" {
		brainDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "brain")
		if _, err := os.Stat(brainDir); os.IsNotExist(err) {
			brainDir = filepath.Join(homeDir, ".gemini", "antigravity", "brain")
		}
		if entries, err := os.ReadDir(brainDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				transPath := filepath.Join(brainDir, entry.Name(), ".system_generated", "logs", "transcript.jsonl")
				if r, ok := parseAgyReceipt(entry.Name(), transPath); ok {
					receipts = append(receipts, r)
				}
			}
		}
	}

	// 3. Scan CodeWhale sessions (~/.codewhale/sessions/*.json)
	if includeAll || filter == "codewhale" {
		cwDir := filepath.Join(homeDir, ".codewhale", "sessions")
		if entries, err := os.ReadDir(cwDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
					continue
				}
				if r, ok := parseCodewhaleReceipt(filepath.Join(cwDir, entry.Name())); ok {
					receipts = append(receipts, r)
				}
			}
		}
	}

	return receipts
}

func parseClaudeReceipt(path string, pricer Pricer) (WorkReceipt, bool) {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return WorkReceipt{}, false
	}

	var (
		tokens        int64
		firstTs       time.Time
		lastTs        time.Time
		cwd           string
		modelName     string
		evList        []Evidence
		agentClaimed  bool
		activeCommand string
	)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var rec struct {
			Type      string `json:"type"`
			Timestamp string `json:"timestamp"`
			Cwd       string `json:"cwd"`
			Message   struct {
				Model string `json:"model"`
				Usage struct {
					InputTokens              int64 `json:"input_tokens"`
					OutputTokens             int64 `json:"output_tokens"`
					CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
					CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
				} `json:"usage"`
				Content []struct {
					Type  string `json:"type"`
					Name  string `json:"name"`
					Input struct {
						Command string `json:"command"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
			ToolUse struct {
				Name  string `json:"name"`
				Input struct {
					Command string `json:"command"`
				} `json:"input"`
			} `json:"tool_use"`
			ToolResult struct {
				ExitCode int    `json:"exit_code"`
				Content  string `json:"content"`
				IsError  bool   `json:"is_error"`
			} `json:"tool_result"`
			Content string `json:"content"`
		}

		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}

		if rec.Cwd != "" {
			cwd = rec.Cwd
		}
		if rec.Message.Model != "" {
			modelName = rec.Message.Model
		}
		if rec.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, rec.Timestamp); err == nil {
				if firstTs.IsZero() || t.Before(firstTs) {
					firstTs = t
				}
				if t.After(lastTs) {
					lastTs = t
				}
			}
		}

		u := rec.Message.Usage
		tokens += u.InputTokens + u.OutputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens

		// Check tool call command
		for _, c := range rec.Message.Content {
			if c.Type == "tool_use" && c.Input.Command != "" {
				activeCommand = c.Input.Command
			}
		}
		if rec.ToolUse.Input.Command != "" {
			activeCommand = rec.ToolUse.Input.Command
		}

		// Check tool result
		if rec.Type == "tool_result" || rec.ToolResult.Content != "" || rec.ToolResult.ExitCode != 0 {
			exitCode := rec.ToolResult.ExitCode
			if rec.ToolResult.IsError && exitCode == 0 {
				exitCode = 1
			}
			if activeCommand != "" {
				if kind, label, ok := ParseCommandEvidence(activeCommand, exitCode, rec.ToolResult.Content); ok {
					evList = append(evList, Evidence{Kind: kind, Label: label})
				}
				activeCommand = ""
			}
		}

		// Check agent claim text
		lower := strings.ToLower(rec.Content)
		if strings.Contains(lower, "task complete") || strings.Contains(lower, "all tests pass") ||
			strings.Contains(lower, "verification complete") || strings.Contains(lower, "verified") {
			agentClaimed = true
		}
	}

	if activeCommand != "" {
		if kind, label, ok := ParseCommandEvidence(activeCommand, 0, ""); ok {
			evList = append(evList, Evidence{Kind: kind, Label: label})
		}
	}

	if firstTs.IsZero() {
		firstTs = time.Now()
	}
	day := LocalDay(firstTs)
	sessID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	outcome := EvaluateSessionOutcome(evList, agentClaimed)

	var cost float64
	var estCost bool
	if pricer != nil && modelName != "" && tokens > 0 {
		c, _ := pricer.PriceDay(DailyRow{
			Day:    day,
			Tokens: tokens,
			Models: map[string]ModelTokens{
				modelName: {Unclassified: tokens},
			},
		})
		if c > 0 {
			cost = c
			estCost = true
		}
	}

	duration := int64(0)
	if !firstTs.IsZero() && !lastTs.IsZero() && lastTs.After(firstTs) {
		duration = lastTs.Sub(firstTs).Milliseconds()
	}

	return WorkReceipt{
		SessionID:       sessID,
		Tool:            "Claude",
		Project:         ProjectKey(cwd),
		Day:             day,
		Tier:            outcome.Tier,
		Status:          outcome.Status,
		Tokens:          tokens,
		Cost:            cost,
		EstimatedCost:   estCost,
		TestsPassed:     outcome.TestsPassed,
		TestsFailed:     outcome.TestsFailed,
		LintersPassed:   outcome.LintersPassed,
		LintersFailed:   outcome.LintersFailed,
		CommitsCreated:  outcome.CommitsCreated,
		EvidenceSummary: outcome.EvidenceSummary,
		DurationMs:      duration,
	}, true
}

func parseAgyReceipt(sessionID, path string) (WorkReceipt, bool) {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return WorkReceipt{}, false
	}

	var (
		firstTs       time.Time
		lastTs        time.Time
		evList        []Evidence
		agentClaimed  bool
		activeCommand string
		steps         int
	)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var rec struct {
			CreatedAt string `json:"created_at"`
			Type      string `json:"type"`
			ToolCalls []struct {
				Name string `json:"name"`
				Args struct {
					CommandLine string `json:"CommandLine"`
				} `json:"args"`
			} `json:"tool_calls"`
			Content string `json:"content"`
		}

		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}

		if rec.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, rec.CreatedAt); err == nil {
				if firstTs.IsZero() || t.Before(firstTs) {
					firstTs = t
				}
				if t.After(lastTs) {
					lastTs = t
				}
			}
		}

		steps++

		for _, tc := range rec.ToolCalls {
			if tc.Args.CommandLine != "" {
				activeCommand = tc.Args.CommandLine
			}
		}

		if activeCommand != "" && rec.Content != "" {
			if kind, label, ok := ParseCommandEvidence(activeCommand, -1, rec.Content); ok {
				evList = append(evList, Evidence{Kind: kind, Label: label})
			}
			activeCommand = ""
		}

		lower := strings.ToLower(rec.Content)
		if strings.Contains(lower, "task completed") || strings.Contains(lower, "goal achieved") ||
			strings.Contains(lower, "all tests pass") || strings.Contains(lower, "verified") {
			agentClaimed = true
		}
	}

	if activeCommand != "" {
		if kind, label, ok := ParseCommandEvidence(activeCommand, 0, ""); ok {
			evList = append(evList, Evidence{Kind: kind, Label: label})
		}
	}

	if firstTs.IsZero() {
		firstTs = time.Now()
	}
	day := LocalDay(firstTs)
	outcome := EvaluateSessionOutcome(evList, agentClaimed)

	duration := int64(0)
	if !firstTs.IsZero() && !lastTs.IsZero() && lastTs.After(firstTs) {
		duration = lastTs.Sub(firstTs).Milliseconds()
	}

	// Agy records steps rather than tokens; each step represents roughly 500 equivalent tokens
	approxTokens := int64(steps) * 500

	return WorkReceipt{
		SessionID:       sessionID,
		Tool:            "Agy",
		Day:             day,
		Tier:            outcome.Tier,
		Status:          outcome.Status,
		Tokens:          approxTokens,
		TestsPassed:     outcome.TestsPassed,
		TestsFailed:     outcome.TestsFailed,
		LintersPassed:   outcome.LintersPassed,
		LintersFailed:   outcome.LintersFailed,
		CommitsCreated:  outcome.CommitsCreated,
		EvidenceSummary: outcome.EvidenceSummary,
		DurationMs:      duration,
	}, true
}

func parseCodewhaleReceipt(path string) (WorkReceipt, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WorkReceipt{}, false
	}

	var session struct {
		SessionID string `json:"session_id"`
		Metadata  struct {
			CreatedAt    string `json:"created_at"`
			TotalTokens  int64  `json:"total_tokens"`
			Mode         string `json:"mode"`
			Workspace    string `json:"workspace"`
			MessageCount int    `json:"message_count"`
			Cost         struct {
				SessionCostUSD float64 `json:"session_cost_usd"`
			} `json:"cost"`
		} `json:"metadata"`
	}

	if json.Unmarshal(data, &session) != nil {
		return WorkReceipt{}, false
	}

	md := session.Metadata
	if md.CreatedAt == "" {
		return WorkReceipt{}, false
	}
	t, err := time.Parse(time.RFC3339, md.CreatedAt)
	if err != nil {
		return WorkReceipt{}, false
	}

	day := LocalDay(t)
	sessID := session.SessionID
	if sessID == "" {
		sessID = strings.TrimSuffix(filepath.Base(path), ".json")
	}

	// CodeWhale sessions without transcript tool calls fall into Claimed (if completed) or Unverified
	tier := Tier3Unverified
	status := "UNVERIFIED"
	if md.Mode == "chat" || md.MessageCount > 0 {
		tier = Tier2Claimed
		status = "CLAIMED"
	}

	return WorkReceipt{
		SessionID:     sessID,
		Tool:          "codewhale",
		Project:       ProjectKey(md.Workspace),
		Day:           day,
		Tier:          tier,
		Status:        status,
		Tokens:        md.TotalTokens,
		Cost:          md.Cost.SessionCostUSD,
		EstimatedCost: false,
	}, true
}
