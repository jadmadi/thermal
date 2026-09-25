// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderReceipt(t *testing.T) {
	rep := thermal.ReceiptReport{
		Type: "receipt",
		Receipts: []thermal.WorkReceipt{
			{
				SessionID:       "sess-12345",
				Tool:            "Claude",
				Project:         "thermal-streak",
				Day:             "2026-09-22",
				Tier:            thermal.Tier1Verified,
				Status:          "VERIFIED",
				Tokens:          150000,
				Cost:            1.25,
				TestsPassed:     2,
				EvidenceSummary: []string{"go test (exit 0)", "git commit (exit 0)"},
			},
			{
				SessionID: "sess-67890",
				Tool:      "codewhale",
				Day:       "2026-09-21",
				Tier:      thermal.Tier2Claimed,
				Status:    "CLAIMED",
				Tokens:    50000,
				Cost:      0.40,
			},
		},
		Summary: thermal.ReceiptSummary{
			TotalReceipts:     2,
			VerifiedCount:     1,
			ClaimedCount:      1,
			VerificationRate:  50.0,
			TotalTokens:       200000,
			TokensVerified:    150000,
			TokensUnverified:  50000,
			TotalCost:         1.65,
			CostVerified:      1.25,
			CostUnverified:    0.40,
			VerifiedTokenRate: 75.0,
			SpendEfficiency:   "HEALTHY",
		},
	}

	out := RenderReceipt(rep, 0, true)

	if !strings.Contains(out, "Thermal · receipt (verifiable work & session outcomes)") {
		t.Errorf("missing header in receipt render:\n%s", out)
	}
	if !strings.Contains(out, "Verification Rate: 50.0%") {
		t.Errorf("missing verification rate in receipt render:\n%s", out)
	}
	if !strings.Contains(out, "[PASS] Verified") || !strings.Contains(out, "[CLAIM] Untested") {
		t.Errorf("missing outcome badges in receipt render:\n%s", out)
	}
	if !strings.Contains(out, "Tier 1") || !strings.Contains(out, "Tier 2") {
		t.Errorf("missing tier references in receipt render:\n%s", out)
	}
	if !strings.Contains(out, "go test (exit 0)") {
		t.Errorf("missing evidence summary in receipt render:\n%s", out)
	}

	// Test empty receipts
	emptyOut := RenderReceipt(thermal.ReceiptReport{}, 0, true)
	if !strings.Contains(emptyOut, "No session receipts found") {
		t.Errorf("expected empty notice, got:\n%s", emptyOut)
	}
}
