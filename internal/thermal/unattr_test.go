package thermal

import "testing"

type flatPricer struct{}

func (flatPricer) PriceDay(day DailyRow) (float64, []string) {
	if len(day.Models) == 0 {
		return 0, nil
	}
	return 1.5, nil
}

// TestUnattributedTokensCounted proves a day whose source names no model is
// counted, so a footer can state it. The pricer cannot price such a day, and
// without the count the estimate looks complete.
func TestUnattributedTokensCounted(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-15", Tokens: 1_000, Turns: 1, Models: map[string]ModelTokens{"m": {Input: 1_000}}},
		{Day: "2026-09-16", Tokens: 2_000, Turns: 1},            // no model
		{Day: "2026-09-17", Tokens: 4_000, Turns: 1, Cost: 3.0}, // recorded cost
		{Day: "2026-09-18", Tokens: 8_000, Turns: 1},            // no model
	}
	rep := Aggregate(days, GrainDay, AggregateOptions{Order: "asc"}, flatPricer{})
	if rep.Totals.UnattributedTokens != 10_000 {
		t.Errorf("unattributed = %d, want 10000 (the two model-less days)", rep.Totals.UnattributedTokens)
	}
	// A day with recorded cost is not unattributed: the money is known.
	for _, row := range rep.Rows {
		if row.Period == "2026-09-17" && row.UnattributedTokens != 0 {
			t.Errorf("a day with recorded cost counted as unattributed: %+v", row)
		}
	}
	if rep.Totals.EstimatedCost != 1.5 {
		t.Errorf("estimated = %.2f, want only the priced day", rep.Totals.EstimatedCost)
	}
}
