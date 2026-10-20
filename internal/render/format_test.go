// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import "testing"

func TestRenderCostContract(t *testing.T) {
	if got := formatCost(0); got != "$0.00" {
		t.Errorf("formatCost(0) = %q, want $0.00", got)
	}
	if got := formatCostValue(0.005, true); got != "~$0.0050" {
		t.Errorf("formatCostValue sub-cent = %q", got)
	}
	if got := formatTotalCost(10, 4, 6); got != "$10.00" {
		t.Errorf("mixed total = %q, want $10.00", got)
	}
	if got := formatTotalCost(10, 0, 10); got != "~$10.00" {
		t.Errorf("all-estimated total = %q", got)
	}
	if got := formatCostCellKind(0, false, 8, false); got != "       —" {
		t.Errorf("unpriced cost cell = %q", got)
	}
}

func TestRenderToolAndPercentContract(t *testing.T) {
	if got := joinTools([]string{"OpenCode", "Devin"}); got != "OpenCode · Devin" {
		t.Errorf("joinTools = %q", got)
	}
	if got := formatCacheHit(96.777); got != "96.8%" {
		t.Errorf("formatCacheHit = %q", got)
	}
	if got := formatPercent(50, 0); got != "50%" {
		t.Errorf("formatPercent = %q", got)
	}
}
