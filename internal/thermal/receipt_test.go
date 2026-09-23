// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseCommandEvidence(t *testing.T) {
	tests := []struct {
		cmd        string
		exitCode   int
		output     string
		wantKind   EvidenceKind
		wantLabel  string
		wantMatch  bool
	}{
		{
			cmd:       "go test ./...",
			exitCode:  0,
			wantKind:  EvidenceTestPass,
			wantLabel: "go test (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "go test -v -race ./cmd/thermal/...",
			exitCode:  1,
			wantKind:  EvidenceTestFail,
			wantLabel: "go test (exit 1)",
			wantMatch: true,
		},
		{
			cmd:       "pytest tests/test_calc.py",
			exitCode:  0,
			wantKind:  EvidenceTestPass,
			wantLabel: "pytest (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "cargo test --workspace",
			exitCode:  0,
			wantKind:  EvidenceTestPass,
			wantLabel: "cargo test (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "npm test",
			exitCode:  0,
			wantKind:  EvidenceTestPass,
			wantLabel: "npm test (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "golangci-lint run",
			exitCode:  0,
			wantKind:  EvidenceLinterPass,
			wantLabel: "golangci-lint (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "eslint src/",
			exitCode:  1,
			wantKind:  EvidenceLinterFail,
			wantLabel: "eslint (exit 1)",
			wantMatch: true,
		},
		{
			cmd:       "git commit -m 'feat: add thermal receipt'",
			exitCode:  0,
			wantKind:  EvidenceGitCommit,
			wantLabel: "git commit (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "go build -o /tmp/thermal ./cmd/thermal",
			exitCode:  0,
			wantKind:  EvidenceBuildPass,
			wantLabel: "build (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "cat README.md",
			exitCode:  0,
			wantKind:  EvidenceUnknown,
			wantLabel: "",
			wantMatch: false,
		},
		{
			cmd:       "go test ./...",
			exitCode:  -1,
			output:    "ok\tgithub.com/jadmadi/thermal\t0.5s\n",
			wantKind:  EvidenceTestPass,
			wantLabel: "go test (exit 0)",
			wantMatch: true,
		},
		{
			cmd:       "go test ./...",
			exitCode:  -1,
			output:    "--- FAIL: TestSomething (0.01s)\nFAIL\n",
			wantKind:  EvidenceTestFail,
			wantLabel: "go test (exit 1)",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			kind, label, ok := ParseCommandEvidence(tt.cmd, tt.exitCode, tt.output)
			if ok != tt.wantMatch {
				t.Fatalf("ParseCommandEvidence(%q) ok = %v; want %v", tt.cmd, ok, tt.wantMatch)
			}
			if kind != tt.wantKind {
				t.Errorf("ParseCommandEvidence(%q) kind = %v; want %v", tt.cmd, kind, tt.wantKind)
			}
			if label != tt.wantLabel {
				t.Errorf("ParseCommandEvidence(%q) label = %q; want %q", tt.cmd, label, tt.wantLabel)
			}
		})
	}
}

func TestEvaluateSessionOutcome(t *testing.T) {
	// Case 1: Verified (passing tests + git commit)
	ev1 := []Evidence{
		{Kind: EvidenceTestPass, Label: "go test (exit 0)"},
		{Kind: EvidenceGitCommit, Label: "git commit (exit 0)"},
	}
	out1 := EvaluateSessionOutcome(ev1, false)
	if out1.Status != "VERIFIED" || out1.Tier != Tier1Verified {
		t.Fatalf("out1 = %v; want VERIFIED / Tier1Verified", out1)
	}
	if out1.TestsPassed != 1 || out1.CommitsCreated != 1 {
		t.Fatalf("out1 counts mismatch: %+v", out1)
	}

	// Case 2: Failed (passing test, but unresolved linter failure)
	ev2 := []Evidence{
		{Kind: EvidenceTestPass, Label: "go test (exit 0)"},
		{Kind: EvidenceLinterFail, Label: "golangci-lint (exit 1)"},
	}
	out2 := EvaluateSessionOutcome(ev2, false)
	if out2.Status != "FAILED" || out2.Tier != Tier3Failed {
		t.Fatalf("out2 = %v; want FAILED / Tier3Failed", out2)
	}

	// Case 3: Claimed (agent claimed completion, but zero verifiable command evidence)
	out3 := EvaluateSessionOutcome(nil, true)
	if out3.Status != "CLAIMED" || out3.Tier != Tier2Claimed {
		t.Fatalf("out3 = %v; want CLAIMED / Tier2Claimed", out3)
	}

	// Case 4: Unverified (zero commands, no agent claim)
	out4 := EvaluateSessionOutcome(nil, false)
	if out4.Status != "UNVERIFIED" || out4.Tier != Tier3Unverified {
		t.Fatalf("out4 = %v; want UNVERIFIED / Tier3Unverified", out4)
	}
}

func TestCalculateSpendEfficiency(t *testing.T) {
	if eff := CalculateSpendEfficiency(85.0, 10000); eff != "EXCELLENT" {
		t.Errorf("expected EXCELLENT, got %s", eff)
	}
	if eff := CalculateSpendEfficiency(65.0, 10000); eff != "HEALTHY" {
		t.Errorf("expected HEALTHY, got %s", eff)
	}
	if eff := CalculateSpendEfficiency(35.0, 10000); eff != "SPECULATIVE" {
		t.Errorf("expected SPECULATIVE, got %s", eff)
	}
	if eff := CalculateSpendEfficiency(90.0, 0); eff != "EXPLORATORY" {
		t.Errorf("expected EXPLORATORY, got %s", eff)
	}
}

func TestAggregateReceipts(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	receipts := []WorkReceipt{
		{
			SessionID:   "s-1",
			Tool:        "Claude",
			Day:         "2026-09-22",
			Tier:        Tier1Verified,
			Status:      "VERIFIED",
			Tokens:      100000,
			Cost:        1.50,
			TestsPassed: 2,
		},
		{
			SessionID: "s-2",
			Tool:      "Claude",
			Day:       "2026-09-22",
			Tier:      Tier2Claimed,
			Status:    "CLAIMED",
			Tokens:    50000,
			Cost:      0.75,
		},
		{
			SessionID:   "s-3",
			Tool:        "Agy",
			Day:         "2026-09-21",
			Tier:        Tier3Failed,
			Status:      "FAILED",
			Tokens:      25000,
			Cost:        0.25,
			TestsFailed: 1,
		},
		{
			SessionID: "s-4",
			Tool:      "codewhale",
			Day:       "2026-09-10", // Outside 7-day window
			Tier:      Tier1Verified,
			Status:    "VERIFIED",
			Tokens:    30000,
			Cost:      0.40,
		},
	}

	opts := ReceiptOptions{
		Last: 7,
		Now:  now,
		Sort: "tokens",
	}

	report := AggregateReceipts(receipts, opts)

	if len(report.Receipts) != 3 {
		t.Fatalf("expected 3 receipts in window, got %d", len(report.Receipts))
	}
	if report.Summary.TotalReceipts != 3 {
		t.Errorf("expected 3 total receipts, got %d", report.Summary.TotalReceipts)
	}
	if report.Summary.VerifiedCount != 1 {
		t.Errorf("expected 1 verified count, got %d", report.Summary.VerifiedCount)
	}
	if report.Summary.TotalTokens != 175000 {
		t.Errorf("expected 175000 total tokens, got %d", report.Summary.TotalTokens)
	}
	if report.Summary.TokensVerified != 100000 {
		t.Errorf("expected 100000 tokens verified, got %d", report.Summary.TokensVerified)
	}

	expectedRate := (1.0 / 3.0) * 100.0
	if fmt.Sprintf("%.1f", report.Summary.VerificationRate) != fmt.Sprintf("%.1f", expectedRate) {
		t.Errorf("expected verification rate %.1f, got %.1f", expectedRate, report.Summary.VerificationRate)
	}

	expectedTokenRate := (100000.0 / 175000.0) * 100.0
	if fmt.Sprintf("%.1f", report.Summary.VerifiedTokenRate) != fmt.Sprintf("%.1f", expectedTokenRate) {
		t.Errorf("expected token rate %.1f, got %.1f", expectedTokenRate, report.Summary.VerifiedTokenRate)
	}

	if report.Summary.SpendEfficiency != "HEALTHY" {
		t.Errorf("expected HEALTHY spend efficiency, got %s", report.Summary.SpendEfficiency)
	}
}

func TestScanSessionReceipts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "thermal-receipt-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock Claude session
	claudeDir := filepath.Join(tmpDir, ".claude", "projects", "repo-x")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	claudeContent := `{"type":"assistant","timestamp":"2026-09-22T10:00:00Z","cwd":"/repo-x","message":{"id":"m1","model":"claude-3-5-sonnet","usage":{"input_tokens":500,"output_tokens":250,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"tool_use","input":{"command":"go test ./..."}}]}}` + "\n" +
		`{"type":"tool_result","content":"PASS\nok\tgithub.com/repo-x\t0.12s\n","exit_code":0}` + "\n" +
		`{"type":"assistant","timestamp":"2026-09-22T10:01:00Z","content":"Task completed successfully."}` + "\n"
	if err := os.WriteFile(filepath.Join(claudeDir, "sess-claude.jsonl"), []byte(claudeContent), 0644); err != nil {
		t.Fatalf("failed to write claude session: %v", err)
	}

	// Create mock CodeWhale session
	cwDir := filepath.Join(tmpDir, ".codewhale", "sessions")
	if err := os.MkdirAll(cwDir, 0755); err != nil {
		t.Fatalf("failed to create codewhale dir: %v", err)
	}
	cwContent := `{
		"session_id": "cw-1",
		"metadata": {
			"created_at": "2026-09-22T11:00:00Z",
			"total_tokens": 15000,
			"mode": "chat",
			"message_count": 5,
			"cost": { "session_cost_usd": 0.05 }
		}
	}`
	if err := os.WriteFile(filepath.Join(cwDir, "cw-1.json"), []byte(cwContent), 0644); err != nil {
		t.Fatalf("failed to write codewhale session: %v", err)
	}

	receipts := ScanSessionReceipts(tmpDir, "all", nil)
	if len(receipts) != 2 {
		t.Fatalf("expected 2 receipts scanned, got %d", len(receipts))
	}

	var foundClaude, foundCw bool
	for _, r := range receipts {
		if r.Tool == "Claude" {
			foundClaude = true
			if r.Status != "VERIFIED" || r.Tier != Tier1Verified {
				t.Errorf("expected Claude receipt to be VERIFIED, got %s / %s", r.Status, r.Tier)
			}
			if r.TestsPassed != 1 {
				t.Errorf("expected Claude TestsPassed == 1, got %d", r.TestsPassed)
			}
			if r.Tokens != 750 {
				t.Errorf("expected Claude tokens == 750, got %d", r.Tokens)
			}
		}
		if r.Tool == "codewhale" {
			foundCw = true
			if r.Status != "CLAIMED" || r.Tier != Tier2Claimed {
				t.Errorf("expected CodeWhale receipt to be CLAIMED, got %s / %s", r.Status, r.Tier)
			}
			if r.Tokens != 15000 {
				t.Errorf("expected CodeWhale tokens == 15000, got %d", r.Tokens)
			}
		}
	}

	if !foundClaude || !foundCw {
		t.Fatalf("missing expected scanned tools (claude: %v, codewhale: %v)", foundClaude, foundCw)
	}
}
