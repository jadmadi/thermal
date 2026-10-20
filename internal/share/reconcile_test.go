// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package share

import (
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestBuildShareSnapshot_TokenScope(t *testing.T) {
	days := []thermal.DailyRow{
		{Day: "2026-01-01", Tokens: 100, Turns: 1},
		{Day: "2026-01-02", Tokens: 10, Turns: 10}, // activity-only row
		{Day: "2026-01-03", Tokens: 50, Turns: 1},
	}
	snap := BuildShareSnapshot("all", days, 0, 7, "all time (2 active days)")
	if snap.ActiveDays != 2 {
		t.Errorf("activeDays = %d, want 2 token-active days", snap.ActiveDays)
	}
	if snap.TotalTokens != 150 {
		t.Errorf("totalTokens = %d, want 150", snap.TotalTokens)
	}
	if snap.CurrentStreak != 1 || snap.LongestStreak != 1 {
		t.Errorf("streak = %d/%d, want token-scoped 1/1", snap.CurrentStreak, snap.LongestStreak)
	}
	if snap.EstimatedCost != 7 || snap.TotalCost != 7 || snap.Coverage != "estimated" {
		t.Errorf("cost = total %v estimated %v coverage %q, want 7/7/estimated", snap.TotalCost, snap.EstimatedCost, snap.Coverage)
	}
	if snap.Scope != "all" || snap.Window != "all time (2 active days)" {
		t.Errorf("scope/window = %q/%q", snap.Scope, snap.Window)
	}
}
