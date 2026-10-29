// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"testing"
	"time"
)

func TestModelTotalsDistinctDays(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	batches := []ToolDays{
		{Tool: "OpenCode", Days: []DailyRow{{
			Day: "2026-10-30", Turns: 1, Tokens: 100,
			Models: map[string]ModelTokens{"model-a": {Output: 100}},
		}}},
		{Tool: "Codex", Days: []DailyRow{{
			Day: "2026-10-30", Turns: 1, Tokens: 200,
			Models: map[string]ModelTokens{"model-b": {Output: 200}},
		}}},
	}
	rep := AggregateModels(batches, ModelOptions{Now: now}, nil)
	if rep.Totals.Days != 1 {
		t.Fatalf("activeDays=%d, want 1 distinct calendar day", rep.Totals.Days)
	}
	if rep.Totals.ModelDays != 2 {
		t.Fatalf("modelDays=%d, want 2", rep.Totals.ModelDays)
	}
	if rep.Totals.ToolCount != 2 {
		t.Fatalf("toolCount=%d, want 2", rep.Totals.ToolCount)
	}
	if rep.Totals.Tokens != 300 {
		t.Fatalf("totalTokens=%d, want 300", rep.Totals.Tokens)
	}
}
