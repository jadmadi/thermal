// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
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
	VerificationRate  float64 `json:"verificationRate"` // percentage 0.0 - 100.0
	TotalTokens       int64   `json:"totalTokens"`
	TokensVerified    int64   `json:"tokensVerified"`
	TokensUnverified  int64   `json:"tokensUnverified"`
	TotalCost         float64 `json:"totalCost"`
	CostVerified      float64 `json:"costVerified"`
	CostUnverified    float64 `json:"costUnverified"`
	VerifiedTokenRate float64 `json:"verifiedTokenRate"` // percentage 0.0 - 100.0
	SpendEfficiency   string  `json:"spendEfficiency"`   // EXCELLENT, HEALTHY, SPECULATIVE, EXPLORATORY
	UnpricedTokens    int64   `json:"unpricedTokens,omitempty"`
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

// UnwrapCommand strips outer quotes and shell execution wrappers (e.g. /bin/bash -lc '...')
// to expose the inner executable command string for evidence parsing.
func UnwrapCommand(cmd string) string {
	c := strings.TrimSpace(cmd)
	if (strings.HasPrefix(c, "\"") && strings.HasSuffix(c, "\"")) ||
		(strings.HasPrefix(c, "'") && strings.HasSuffix(c, "'")) {
		if len(c) >= 2 {
			c = strings.TrimSpace(c[1 : len(c)-1])
		}
	}
	for {
		stripped := false
		for _, prefix := range []string{
			"/bin/bash -lc ", "bash -lc ",
			"/bin/bash -c ", "bash -c ",
			"/bin/sh -c ", "sh -c ",
			"/usr/bin/bash -lc ", "/usr/bin/bash -c ",
			"/usr/bin/sh -c ",
		} {
			if strings.HasPrefix(c, prefix) {
				rest := strings.TrimSpace(strings.TrimPrefix(c, prefix))
				if (strings.HasPrefix(rest, "'") && strings.HasSuffix(rest, "'")) ||
					(strings.HasPrefix(rest, "\"") && strings.HasSuffix(rest, "\"")) {
					if len(rest) >= 2 {
						c = strings.TrimSpace(rest[1 : len(rest)-1])
						stripped = true
						break
					}
				} else {
					c = rest
					stripped = true
					break
				}
			}
		}
		if !stripped {
			break
		}
	}
	return c
}

// hasShellComposition reports whether an unwrapped command string contains
// unquoted shell operators that compose or substitute multiple invocations
// (lists, pipelines, substitutions, background and grouping operators).
// A compound command's process exit status cannot be attributed to a single
// recognized invocation, so such commands stay unknown unless the source
// adapter supplies separate command/result pairs.
func hasShellComposition(cmd string) bool {
	var quote rune
	runes := []rune(cmd)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if quote != 0 {
			if quote == '"' && c == '\\' {
				i++
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
		case '\\':
			i++
		case '`', ';', '|', '(', ')', '\n':
			return true
		case '&':
			if i > 0 && (runes[i-1] == '>' || runes[i-1] == '<') {
				continue // redirection fd duplication: >&1, <&0
			}
			return true
		case '$':
			if i+1 < len(runes) && runes[i+1] == '(' {
				return true
			}
		}
	}
	return false
}

// isEnvAssignment reports whether a token is a leading shell variable
// assignment (e.g. CGO_ENABLED=0) rather than the command itself.
func isEnvAssignment(tok string) bool {
	eq := strings.IndexByte(tok, '=')
	if eq <= 0 {
		return false
	}
	for i, r := range tok[:eq] {
		switch {
		case r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
		case r >= '0' && r <= '9':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// evidenceRunner identifies a single recognized verification invocation.
type evidenceRunner struct {
	name string
	kind string // "test", "linter", "commit", "build"
}

// matchEvidenceRunner recognizes a single simple invocation by token boundary,
// never by arbitrary substring presence. Commands are matched on their leading
// executable token(s) only; "go tester" is not "go test".
func matchEvidenceRunner(cmd string) (evidenceRunner, bool) {
	tokens := strings.Fields(cmd)
	i := 0
	for i < len(tokens) && isEnvAssignment(tokens[i]) {
		i++
	}
	if i >= len(tokens) {
		return evidenceRunner{}, false
	}
	tokens = tokens[i:]

	base := func(tok string) string { return strings.ToLower(filepath.Base(tok)) }
	arg := func(idx int) string {
		if idx < len(tokens) {
			return strings.ToLower(tokens[idx])
		}
		return ""
	}

	switch first := base(tokens[0]); first {
	case "go":
		switch arg(1) {
		case "test":
			return evidenceRunner{"go test", "test"}, true
		case "vet":
			return evidenceRunner{"golangci-lint", "linter"}, true
		case "build":
			return evidenceRunner{"build", "build"}, true
		}
	case "cargo":
		switch arg(1) {
		case "test":
			return evidenceRunner{"cargo test", "test"}, true
		case "clippy", "check":
			return evidenceRunner{"cargo clippy", "linter"}, true
		case "build":
			return evidenceRunner{"build", "build"}, true
		}
	case "pytest":
		return evidenceRunner{"pytest", "test"}, true
	case "python", "python3":
		if arg(1) == "-m" && (arg(2) == "pytest" || arg(2) == "unittest") {
			return evidenceRunner{"pytest", "test"}, true
		}
	case "npm", "pnpm", "yarn", "bun":
		switch arg(1) {
		case "test":
			return evidenceRunner{"npm test", "test"}, true
		case "t":
			if first == "npm" {
				return evidenceRunner{"npm test", "test"}, true
			}
		case "run":
			switch arg(2) {
			case "test":
				return evidenceRunner{"npm test", "test"}, true
			case "build":
				return evidenceRunner{"build", "build"}, true
			case "lint":
				return evidenceRunner{"tsc", "linter"}, true
			}
		case "lint":
			if first == "pnpm" || first == "yarn" {
				return evidenceRunner{"tsc", "linter"}, true
			}
		}
	case "jest", "vitest", "mocha":
		return evidenceRunner{"npm test", "test"}, true
	case "npx", "pnpx", "bunx":
		switch base(arg(1)) {
		case "jest", "vitest", "mocha":
			return evidenceRunner{"npm test", "test"}, true
		case "eslint":
			return evidenceRunner{"eslint", "linter"}, true
		case "tsc":
			return evidenceRunner{"tsc", "linter"}, true
		case "prettier":
			if arg(2) == "--check" {
				return evidenceRunner{"eslint", "linter"}, true
			}
		}
	case "make":
		if arg(1) == "test" {
			return evidenceRunner{"make test", "test"}, true
		}
	case "ctest":
		return evidenceRunner{"make test", "test"}, true
	case "rspec":
		return evidenceRunner{"rspec", "test"}, true
	case "mvn", "mvnw", "gradle", "gradlew":
		if arg(1) == "test" {
			return evidenceRunner{"mvn test", "test"}, true
		}
	case "golangci-lint", "staticcheck", "revive":
		return evidenceRunner{"golangci-lint", "linter"}, true
	case "eslint", "biome":
		return evidenceRunner{"eslint", "linter"}, true
	case "prettier":
		if arg(1) == "--check" {
			return evidenceRunner{"eslint", "linter"}, true
		}
	case "ruff", "flake8", "mypy", "pyright":
		return evidenceRunner{"ruff", "linter"}, true
	case "tsc":
		return evidenceRunner{"tsc", "linter"}, true
	case "git":
		switch arg(1) {
		case "commit", "merge":
			return evidenceRunner{"git commit", "commit"}, true
		}
	}
	return evidenceRunner{}, false
}

var (
	reReceiptExitLine   = regexp.MustCompile(`the command exited with code\s+([0-9]+)`)
	reReceiptStatusLine = regexp.MustCompile(`exit status\s+([0-9]+)`)
	reReceiptPassed     = regexp.MustCompile(`([0-9]+)\s+passed`)
	reReceiptFailed     = regexp.MustCompile(`([0-9]+)\s+(?:failed|failing)`)
	reReceiptErrors     = regexp.MustCompile(`([0-9]+)\s+errors?`)
	reReceiptFailMarker = regexp.MustCompile(`(?m)^--- fail:`)
	reReceiptFailLine   = regexp.MustCompile(`(?m)^fail(\s|$)`)
	reReceiptPassLine   = regexp.MustCompile(`(?m)^pass\s*$`)
	reReceiptOKLine     = regexp.MustCompile(`(?m)^ok[\t ]`)
	reReceiptCargoOK    = regexp.MustCompile(`test result: ok\.`)
	reReceiptVerboseOK  = regexp.MustCompile(`(?m)^--- pass:`)
)

// maxMatchedCount returns the largest captured count for a summary regex,
// or zero when the output contains no matching count at all.
func maxMatchedCount(re *regexp.Regexp, s string) int {
	max := 0
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil && n > max {
			max = n
		}
	}
	return max
}

// hasContradictingFailureText reports whether a completed zero-exit run's
// output also carries explicit structural failure text.
func hasContradictingFailureText(outLower string) bool {
	return strings.Contains(outLower, "--- fail:") ||
		strings.Contains(outLower, "\nfail\t") ||
		strings.Contains(outLower, "failures:")
}

// resolveUnobservedExit infers an exit status from a complete recognized
// result line when the process exit status was not recorded, or -1 when the
// output does not establish success or failure. Success lines are matched
// before generic failure words; a bare "0 failed" proves neither outcome.
func resolveUnobservedExit(outLower string) int {
	if m := reReceiptExitLine.FindStringSubmatch(outLower); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	if m := reReceiptStatusLine.FindStringSubmatch(outLower); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}

	failed := maxMatchedCount(reReceiptFailed, outLower)
	errors := maxMatchedCount(reReceiptErrors, outLower)
	passed := maxMatchedCount(reReceiptPassed, outLower)
	if failed > 0 || errors > 0 {
		return 1
	}

	if reReceiptFailMarker.MatchString(outLower) || reReceiptFailLine.MatchString(outLower) ||
		hasContradictingFailureText(outLower) || strings.Contains(outLower, "panic:") {
		return 1
	}
	if passed > 0 {
		return 0
	}
	if reReceiptOKLine.MatchString(outLower) || reReceiptPassLine.MatchString(outLower) ||
		reReceiptCargoOK.MatchString(outLower) || reReceiptVerboseOK.MatchString(outLower) {
		return 0
	}
	return -1
}

// ParseCommandEvidence inspects a command string, its exit code, and output snippet
// to extract sanitized verification evidence without preserving sensitive paths or prompts.
// Evidence requires a single recognized invocation observed to completion: compound shell
// commands stay unknown, explicit nonzero exits win, and a missing exit is only resolved
// by a complete recognized result line in the output.
func ParseCommandEvidence(cmd string, exitCode int, output string) (EvidenceKind, string, bool) {
	cmdClean := UnwrapCommand(cmd)
	if cmdClean == "" || hasShellComposition(cmdClean) {
		return EvidenceUnknown, "", false
	}

	runner, ok := matchEvidenceRunner(cmdClean)
	if !ok {
		return EvidenceUnknown, "", false
	}

	outLower := strings.ToLower(output)
	effCode := exitCode
	if effCode < 0 {
		effCode = resolveUnobservedExit(outLower)
	}
	if effCode < 0 {
		return EvidenceUnknown, "", false
	}
	if effCode == 0 && hasContradictingFailureText(outLower) {
		effCode = 1
	}

	switch runner.kind {
	case "test":
		if effCode == 0 {
			return EvidenceTestPass, fmt.Sprintf("%s (exit 0)", runner.name), true
		}
		return EvidenceTestFail, fmt.Sprintf("%s (exit %d)", runner.name, effCode), true
	case "linter":
		if effCode == 0 {
			return EvidenceLinterPass, fmt.Sprintf("%s (exit 0)", runner.name), true
		}
		return EvidenceLinterFail, fmt.Sprintf("%s (exit %d)", runner.name, effCode), true
	case "commit":
		if effCode == 0 {
			return EvidenceGitCommit, "git commit (exit 0)", true
		}
		return EvidenceUnknown, "", false
	case "build":
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
		unpricedTokens   int64
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
		if r.Tokens > 0 && r.Cost == 0 {
			unpricedTokens += r.Tokens
		}

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
			UnpricedTokens:    unpricedTokens,
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
	filter := strings.ToLower(strings.TrimSpace(toolFilter))
	switch filter {
	case "claude", "claude-code", "cc":
		filter = "claude"
	case "agy", "antigravity":
		filter = "agy"
	case "codewhale", "whale", "cw":
		filter = "codewhale"
	case "command-code", "commandcode", "ccode":
		filter = "command-code"
	case "codex", "codex-cli":
		filter = "codex"
	}
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
		brainDir := filepath.Join(homeDir, ".gemini", "antigravity", "brain")
		cliBrain := filepath.Join(homeDir, ".gemini", "antigravity-cli", "brain")
		stAgy, errAgy := os.Stat(brainDir)
		stCli, errCli := os.Stat(cliBrain)
		if errAgy == nil && errCli == nil {
			if stCli.ModTime().After(stAgy.ModTime()) {
				brainDir = cliBrain
			}
		} else if errCli == nil {
			brainDir = cliBrain
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

	// 4. Scan Codex transcripts and state (~/.codex/state_5.sqlite, ~/.codex/sessions/**/*.jsonl)
	if includeAll || filter == "codex" {
		codexDir := os.Getenv("CODEX_HOME")
		if codexDir == "" {
			codexDir = filepath.Join(homeDir, ".codex")
		}
		receipts = append(receipts, ScanCodexReceipts(codexDir, pricer)...)
	}

	return receipts
}

func extractContentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var parts []string
		for _, b := range blocks {
			if b.Text != "" {
				parts = append(parts, b.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return string(raw)
}

func parseClaudeReceipt(path string, pricer Pricer) (WorkReceipt, bool) {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return WorkReceipt{}, false
	}

	var (
		tokens       int64
		firstTs      time.Time
		lastTs       time.Time
		cwd          string
		modelName    string
		evList       []Evidence
		agentClaimed bool
	)

	type contentBlock struct {
		Type      string          `json:"type"`
		ID        string          `json:"id"`
		Name      string          `json:"name"`
		ToolUseID string          `json:"tool_use_id"`
		Content   json.RawMessage `json:"content"`
		IsError   bool            `json:"is_error"`
		ExitCode  *int            `json:"exit_code"`
		Input     struct {
			Command string `json:"command"`
		} `json:"input"`
	}

	pendingMap := make(map[string]string)
	var pendingList []string
	seenMsgIDs := make(map[string]bool)
	modelsMap := make(map[string]ModelTokens)

	handleToolResult := func(toolUseID string, rawContent json.RawMessage, isError bool, exitCodePtr *int) {
		var cmd string
		if toolUseID != "" {
			if c, ok := pendingMap[toolUseID]; ok {
				cmd = c
				delete(pendingMap, toolUseID)
				for i, pid := range pendingList {
					if pid == toolUseID {
						pendingList = append(pendingList[:i], pendingList[i+1:]...)
						break
					}
				}
			} else {
				// ID specified but not in pending commands map: unrelated or already resolved tool use
				return
			}
		} else if len(pendingList) > 0 {
			oldest := pendingList[0]
			cmd = pendingMap[oldest]
			delete(pendingMap, oldest)
			pendingList = pendingList[1:]
		}
		if cmd == "" {
			return
		}

		exitCode := -1
		if exitCodePtr != nil {
			exitCode = *exitCodePtr
		}
		if isError && exitCode <= 0 {
			exitCode = 1
		}

		contentStr := extractContentText(rawContent)
		if kind, label, ok := ParseCommandEvidence(cmd, exitCode, contentStr); ok {
			evList = append(evList, Evidence{Kind: kind, Label: label})
		}
	}

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
				ID    string `json:"id"`
				Model string `json:"model"`
				Usage struct {
					InputTokens              int64 `json:"input_tokens"`
					OutputTokens             int64 `json:"output_tokens"`
					CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
					CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
				} `json:"usage"`
				Content []contentBlock `json:"content"`
			} `json:"message"`
			ToolUse struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Input struct {
					Command string `json:"command"`
				} `json:"input"`
			} `json:"tool_use"`
			ToolUseID  string `json:"tool_use_id"`
			ToolResult struct {
				ExitCode *int            `json:"exit_code"`
				Content  json.RawMessage `json:"content"`
				IsError  bool            `json:"is_error"`
			} `json:"tool_result"`
			ExitCode *int            `json:"exit_code"`
			IsError  bool            `json:"is_error"`
			Content  json.RawMessage `json:"content"`
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

		// Token usage accounting with message deduplication and disjoint categories
		u := rec.Message.Usage
		msgTok := u.InputTokens + u.OutputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
		if msgTok > 0 {
			msgID := rec.Message.ID
			isDupe := false
			if msgID != "" {
				if seenMsgIDs[msgID] {
					isDupe = true
				} else {
					seenMsgIDs[msgID] = true
				}
			}
			if !isDupe {
				tokens += msgTok
				m := strings.ToLower(strings.TrimSpace(rec.Message.Model))
				if m != "" {
					mt := modelsMap[m]
					mt.Input += u.InputTokens
					mt.Output += u.OutputTokens
					mt.CacheRead += u.CacheReadInputTokens
					mt.CacheWrite += u.CacheCreationInputTokens
					modelsMap[m] = mt
				}
			}
		}

		// 1. Process tool_use in message.content
		for _, c := range rec.Message.Content {
			if c.Type == "tool_use" && c.Input.Command != "" {
				id := c.ID
				if id == "" {
					id = fmt.Sprintf("anon_%d", len(pendingList))
				}
				pendingMap[id] = c.Input.Command
				pendingList = append(pendingList, id)
			}
		}
		// Also check top-level tool_use
		if rec.ToolUse.Input.Command != "" {
			id := rec.ToolUse.ID
			if id == "" {
				id = fmt.Sprintf("anon_%d", len(pendingList))
			}
			pendingMap[id] = rec.ToolUse.Input.Command
			pendingList = append(pendingList, id)
		}

		// 2. Process tool_result in message.content
		for _, c := range rec.Message.Content {
			if c.Type == "tool_result" {
				handleToolResult(c.ToolUseID, c.Content, c.IsError, c.ExitCode)
			}
		}

		// 3. Process top-level tool_result
		if rec.Type == "tool_result" || len(rec.ToolResult.Content) > 0 || rec.ToolResult.ExitCode != nil || rec.ExitCode != nil {
			tID := rec.ToolUseID
			contentRaw := rec.ToolResult.Content
			if len(contentRaw) == 0 && rec.Type == "tool_result" {
				contentRaw = rec.Content
			}
			exitCodePtr := rec.ToolResult.ExitCode
			if exitCodePtr == nil {
				exitCodePtr = rec.ExitCode
			}
			isError := rec.ToolResult.IsError || rec.IsError
			handleToolResult(tID, contentRaw, isError, exitCodePtr)
		}

		// Check agent claim text
		contentStr := extractContentText(rec.Content)
		lower := strings.ToLower(contentStr)
		if strings.Contains(lower, "task complete") || strings.Contains(lower, "all tests pass") ||
			strings.Contains(lower, "verification complete") || strings.Contains(lower, "verified") {
			agentClaimed = true
		}
	}

	if firstTs.IsZero() {
		firstTs = time.Now()
	}
	day := LocalDay(firstTs)
	sessID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	outcome := EvaluateSessionOutcome(evList, agentClaimed)

	// Fallback to modelName if modelsMap was empty but a model name was seen
	if len(modelsMap) == 0 && modelName != "" && tokens > 0 {
		m := strings.ToLower(strings.TrimSpace(modelName))
		modelsMap[m] = ModelTokens{Unclassified: tokens}
	}

	var cost float64
	var estCost bool
	if pricer != nil && len(modelsMap) > 0 && tokens > 0 {
		c, _ := pricer.PriceDay(DailyRow{
			Day:    day,
			Tokens: tokens,
			Models: modelsMap,
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
		project       string
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
					Cwd         string `json:"Cwd"`
				} `json:"args"`
			} `json:"tool_calls"`
			Content string `json:"content"`
		}

		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}

		if project == "" {
			for _, tc := range rec.ToolCalls {
				cwd := strings.Trim(strings.TrimSpace(tc.Args.Cwd), "\"")
				if cwd != "" && filepath.IsAbs(cwd) {
					project = ProjectKey(cwd)
					break
				}
			}
			if project == "" && strings.Contains(rec.Content, " -> ") {
				idx := strings.Index(rec.Content, " -> ")
				start := strings.LastIndex(rec.Content[:idx], "\n")
				cand := strings.TrimSpace(rec.Content[start+1 : idx])
				cand = strings.TrimPrefix(cand, "file://")
				if filepath.IsAbs(cand) {
					project = ProjectKey(cand)
				}
			}
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

		hasToolCall := false
		for _, tc := range rec.ToolCalls {
			if tc.Args.CommandLine != "" {
				activeCommand = tc.Args.CommandLine
				hasToolCall = true
			}
		}

		if !hasToolCall && activeCommand != "" && rec.Content != "" {
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

	if firstTs.IsZero() {
		firstTs = time.Now()
	}
	day := LocalDay(firstTs)
	outcome := EvaluateSessionOutcome(evList, agentClaimed)

	duration := int64(0)
	if !firstTs.IsZero() && !lastTs.IsZero() && lastTs.After(firstTs) {
		duration = lastTs.Sub(firstTs).Milliseconds()
	}

	// Agy records step activity rather than token usage. Activity rows carry 0 tokens.
	return WorkReceipt{
		SessionID:       sessionID,
		Tool:            "Agy",
		Project:         project,
		Day:             day,
		Tier:            outcome.Tier,
		Status:          outcome.Status,
		Tokens:          0,
		Cost:            0,
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

// IsActivityOnly reports whether a row is activity telemetry rather than token telemetry.
func IsActivityOnly(day DailyRow) bool {
	return isActivityOnly(day)
}

func newReceiptJSONLScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 32*1024*1024)
	return scanner
}

func extractCodexCommand(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return UnwrapCommand(s)
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		if len(arr) == 0 {
			return ""
		}
		if len(arr) >= 3 && (strings.HasSuffix(arr[0], "bash") || strings.HasSuffix(arr[0], "sh")) && (arr[1] == "-lc" || arr[1] == "-c") {
			return UnwrapCommand(arr[2])
		}
		return UnwrapCommand(strings.Join(arr, " "))
	}
	return ""
}

func hasColumnReceipt(db *sql.DB, table, col string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
			if strings.EqualFold(name, col) {
				return true
			}
		}
	}
	return rows.Err() == nil
}

// parseCodexRollout parses a Codex rollout JSONL file to extract commands, exit codes,
// tokens, and factual verification evidence.
func parseCodexRollout(path string, threadID string, tokensOverride int64, modelOverride, cwdOverride string, createdOverride, updatedOverride int64, pricer Pricer) (WorkReceipt, bool) {
	f, err := os.Open(path)
	if err != nil {
		return WorkReceipt{}, false
	}
	defer f.Close()

	var (
		tokens       int64
		firstTs      time.Time
		lastTs       time.Time
		cwd          = cwdOverride
		model        = modelOverride
		evList       []Evidence
		agentClaimed bool
		sessID       = threadID
		seenEvLabels = make(map[string]bool)
	)

	if sessID == "" {
		base := filepath.Base(path)
		sessID = strings.TrimSuffix(base, ".jsonl")
		if strings.HasPrefix(sessID, "rollout-") {
			parts := strings.Split(sessID, "-")
			if len(parts) >= 8 {
				sessID = strings.Join(parts[len(parts)-5:], "-")
			}
		}
	}

	scanner := newReceiptJSONLScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var rec struct {
			Timestamp string          `json:"timestamp"`
			Type      string          `json:"type"`
			Payload   json.RawMessage `json:"payload"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}

		if rec.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, rec.Timestamp); err == nil {
				if firstTs.IsZero() || t.Before(firstTs) {
					firstTs = t
				}
				if t.After(lastTs) {
					lastTs = t
				}
			} else if t, err := time.Parse(time.RFC3339, rec.Timestamp); err == nil {
				if firstTs.IsZero() || t.Before(firstTs) {
					firstTs = t
				}
				if t.After(lastTs) {
					lastTs = t
				}
			}
		}

		// 1. turn_context: cwd, model
		if rec.Type == "turn_context" {
			var tc struct {
				Cwd   string `json:"cwd"`
				Model string `json:"model"`
			}
			if json.Unmarshal(rec.Payload, &tc) == nil {
				if cwd == "" && tc.Cwd != "" {
					cwd = strings.TrimPrefix(tc.Cwd, "file://")
				}
				if model == "" && tc.Model != "" {
					model = tc.Model
				}
			}
			continue
		}

		// 2. event_msg
		if rec.Type == "event_msg" {
			var ev struct {
				Type     string          `json:"type"`
				ThreadID string          `json:"thread_id"`
				Item     json.RawMessage `json:"item"`
				Info     *struct {
					TotalTokenUsage *struct {
						InputTokens           int64 `json:"input_tokens"`
						CachedInputTokens     int64 `json:"cached_input_tokens"`
						OutputTokens          int64 `json:"output_tokens"`
						ReasoningOutputTokens int64 `json:"reasoning_output_tokens"`
						TotalTokens           int64 `json:"total_tokens"`
					} `json:"total_token_usage"`
				} `json:"info"`
				LastAgentMessage *string `json:"last_agent_message"`
			}
			if json.Unmarshal(rec.Payload, &ev) != nil {
				continue
			}

			if sessID == "" && ev.ThreadID != "" {
				sessID = ev.ThreadID
			}

			// Token count
			if ev.Type == "token_count" && ev.Info != nil && ev.Info.TotalTokenUsage != nil {
				tu := ev.Info.TotalTokenUsage
				if tu.TotalTokens > 0 {
					tokens = tu.TotalTokens
				} else {
					tokens = tu.InputTokens + tu.OutputTokens
				}
			}

			// Task complete / agent message claims
			if ev.Type == "task_complete" {
				agentClaimed = true
				if ev.LastAgentMessage != nil {
					lower := strings.ToLower(*ev.LastAgentMessage)
					if strings.Contains(lower, "all tests pass") || strings.Contains(lower, "verified") || strings.Contains(lower, "task complete") {
						agentClaimed = true
					}
				}
			}

			// Item handling (CommandExecution, AgentMessage)
			if len(ev.Item) > 0 {
				var itHeader struct {
					Type string `json:"type"`
				}
				if json.Unmarshal(ev.Item, &itHeader) == nil {
					if itHeader.Type == "CommandExecution" {
						var itCmd struct {
							Command    json.RawMessage `json:"command"`
							ExitCode   *int            `json:"exit_code"`
							Status     string          `json:"status"`
							Stdout     string          `json:"stdout"`
							Stderr     string          `json:"stderr"`
							Aggregated string          `json:"aggregated_output"`
							Cwd        string          `json:"cwd"`
						}
						if json.Unmarshal(ev.Item, &itCmd) == nil {
							if cwd == "" && itCmd.Cwd != "" {
								cwd = strings.TrimPrefix(itCmd.Cwd, "file://")
							}
							rawCmd := extractCodexCommand(itCmd.Command)
							if rawCmd != "" {
								exitCode := -1
								if itCmd.ExitCode != nil {
									exitCode = *itCmd.ExitCode
								}
								if exitCode == 0 && (itCmd.Status == "failed" || itCmd.Status == "error") {
									exitCode = 1
								}
								output := itCmd.Aggregated
								if output == "" {
									output = itCmd.Stdout + "\n" + itCmd.Stderr
								}

								if kind, label, ok := ParseCommandEvidence(rawCmd, exitCode, output); ok {
									if !seenEvLabels[label] {
										evList = append(evList, Evidence{Kind: kind, Label: label})
										seenEvLabels[label] = true
									}
								}
							}
						}
					} else if itHeader.Type == "AgentMessage" {
						var itMsg struct {
							Content []struct {
								Type string `json:"type"`
								Text string `json:"text"`
							} `json:"content"`
						}
						if json.Unmarshal(ev.Item, &itMsg) == nil {
							for _, c := range itMsg.Content {
								lower := strings.ToLower(c.Text)
								if strings.Contains(lower, "all tests pass") || strings.Contains(lower, "task complete") || strings.Contains(lower, "verified") {
									agentClaimed = true
								}
							}
						}
					}
				}
			}
		}

		// 3. direct item.completed format
		if rec.Type == "item.completed" {
			var itCmd struct {
				Type     string          `json:"type"`
				Command  json.RawMessage `json:"command"`
				ExitCode *int            `json:"exit_code"`
				Status   string          `json:"status"`
			}
			if json.Unmarshal(rec.Payload, &itCmd) == nil && (itCmd.Type == "command_execution" || itCmd.Type == "CommandExecution" || itCmd.Type == "bash") {
				rawCmd := extractCodexCommand(itCmd.Command)
				if rawCmd != "" {
					exitCode := -1
					if itCmd.ExitCode != nil {
						exitCode = *itCmd.ExitCode
					}
					if kind, label, ok := ParseCommandEvidence(rawCmd, exitCode, ""); ok {
						if !seenEvLabels[label] {
							evList = append(evList, Evidence{Kind: kind, Label: label})
							seenEvLabels[label] = true
						}
					}
				}
			}
		}
	}

	if tokensOverride > 0 {
		tokens = tokensOverride
	}
	if modelOverride != "" {
		model = modelOverride
	}

	var day string
	var durationMs int64
	if createdOverride > 0 {
		day = LocalDay(time.Unix(createdOverride, 0).Local())
		if updatedOverride > createdOverride {
			durationMs = (updatedOverride - createdOverride) * 1000
		}
	} else if !firstTs.IsZero() {
		day = LocalDay(firstTs)
		if !lastTs.IsZero() && lastTs.After(firstTs) {
			durationMs = lastTs.Sub(firstTs).Milliseconds()
		}
	} else {
		day = LocalDay(time.Now())
	}

	outcome := EvaluateSessionOutcome(evList, agentClaimed)

	var cost float64
	var estCost bool
	if pricer != nil && model != "" && tokens > 0 {
		c, _ := pricer.PriceDay(DailyRow{
			Day:    day,
			Tokens: tokens,
			Models: map[string]ModelTokens{
				model: {Unclassified: tokens},
			},
		})
		if c > 0 {
			cost = c
			estCost = true
		}
	}

	return WorkReceipt{
		SessionID:       sessID,
		Tool:            "Codex",
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
		DurationMs:      durationMs,
	}, true
}

// ScanCodexReceipts scans Codex SQLite database and rollout logs in dataDir
// to extract factual work receipts with local test and linter evidence.
func ScanCodexReceipts(codexDir string, pricer Pricer) []WorkReceipt {
	var receipts []WorkReceipt

	stateDB := filepath.Join(codexDir, "state_5.sqlite")
	if fi, err := os.Stat(stateDB); err == nil && !fi.IsDir() {
		db, err := sql.Open("sqlite", "file:"+stateDB+"?mode=ro")
		if err == nil {
			defer db.Close()
			_, _ = db.Exec("PRAGMA mmap_size=268435456")

			cwdExpr := "''"
			if hasColumnReceipt(db, "threads", "cwd") {
				cwdExpr = "cwd"
			}

			rows, err := db.Query(`
				SELECT id, tokens_used, model, created_at, updated_at, rollout_path, ` + cwdExpr + `
				FROM threads
				WHERE archived = 0
				ORDER BY created_at
			`)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var id, model, rolloutPath, cwd sql.NullString
					var tokensUsed, createdAt, updatedAt int64
					if err := rows.Scan(&id, &tokensUsed, &model, &createdAt, &updatedAt, &rolloutPath, &cwd); err != nil {
						continue
					}

					var r WorkReceipt
					var ok bool
					if rolloutPath.Valid && rolloutPath.String != "" {
						r, ok = parseCodexRollout(rolloutPath.String, id.String, tokensUsed, model.String, cwd.String, createdAt, updatedAt, pricer)
					}
					if !ok {
						day := LocalDay(time.Unix(createdAt, 0).Local())
						tier := Tier3Unverified
						status := "UNVERIFIED"
						if tokensUsed > 0 {
							tier = Tier2Claimed
							status = "CLAIMED"
						}
						var cost float64
						var estCost bool
						if pricer != nil && model.Valid && model.String != "" && tokensUsed > 0 {
							c, _ := pricer.PriceDay(DailyRow{
								Day:    day,
								Tokens: tokensUsed,
								Models: map[string]ModelTokens{
									model.String: {Unclassified: tokensUsed},
								},
							})
							if c > 0 {
								cost = c
								estCost = true
							}
						}
						duration := int64(0)
						if updatedAt > createdAt {
							duration = (updatedAt - createdAt) * 1000
						}
						r = WorkReceipt{
							SessionID:     id.String,
							Tool:          "Codex",
							Project:       ProjectKey(cwd.String),
							Day:           day,
							Tier:          tier,
							Status:        status,
							Tokens:        tokensUsed,
							Cost:          cost,
							EstimatedCost: estCost,
							DurationMs:    duration,
						}
					}
					receipts = append(receipts, r)
				}
				if err := rows.Err(); err == nil && len(receipts) > 0 {
					return receipts
				}
			}
		}
	}

	// Fallback when state_5.sqlite is missing or empty: scan session rollout JSONL files directly
	var files []string
	for _, pattern := range []string{
		filepath.Join(codexDir, "sessions", "*", "*", "*", "*.jsonl"),
		filepath.Join(codexDir, "sessions", "*", "*.jsonl"),
		filepath.Join(codexDir, "sessions", "*.jsonl"),
		filepath.Join(codexDir, "*.jsonl"),
	} {
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			files = append(files, matches...)
		}
	}

	for _, file := range files {
		if r, ok := parseCodexRollout(file, "", 0, "", "", 0, 0, pricer); ok {
			receipts = append(receipts, r)
		}
	}

	return receipts
}
