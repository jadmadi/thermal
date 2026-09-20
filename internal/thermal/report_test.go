package thermal

import (
	"testing"
	"time"
)

func TestParseDay(t *testing.T) {
	for _, in := range []string{"2026-09-16", "20260916"} {
		got, ok := ParseDay(in)
		if !ok {
			t.Fatalf("ParseDay(%q) failed", in)
		}
		if got.Format("2006-01-02") != "2026-09-16" {
			t.Errorf("ParseDay(%q) = %s", in, got.Format("2006-01-02"))
		}
	}
	if _, ok := ParseDay("16/09/2026"); ok {
		t.Error("expected ParseDay to reject 16/09/2026")
	}
	if _, ok := ParseDay(""); ok {
		t.Error("expected ParseDay to reject empty input")
	}
}

func TestParseWeekday(t *testing.T) {
	if d, ok := ParseWeekday("Sunday"); !ok || d != time.Sunday {
		t.Errorf("ParseWeekday(Sunday) = %v, %v", d, ok)
	}
	if d, ok := ParseWeekday("monday"); !ok || d != time.Monday {
		t.Errorf("ParseWeekday(monday) = %v, %v", d, ok)
	}
	if _, ok := ParseWeekday("someday"); ok {
		t.Error("expected ParseWeekday to reject someday")
	}
}

func TestAggregateWeeklySundayAndMonday(t *testing.T) {
	// 2026-09-16 is a Wednesday. Sunday starts 2026-09-13, Monday 2026-09-14.
	days := []DailyRow{
		{Day: "2026-09-13", Tokens: 100, Input: 100, Turns: 1},
		{Day: "2026-09-16", Tokens: 200, Input: 200, Turns: 2},
		{Day: "2026-09-20", Tokens: 300, Input: 300, Turns: 1},
	}

	sunday := Aggregate(days, GrainWeek, AggregateOptions{StartOfWeek: time.Sunday, Order: "asc"}, nil)
	if len(sunday.Rows) != 2 {
		t.Fatalf("expected 2 sunday weeks, got %d", len(sunday.Rows))
	}
	if sunday.Rows[0].Period != "2026-09-13" {
		t.Errorf("sunday week start = %s, want 2026-09-13", sunday.Rows[0].Period)
	}

	monday := Aggregate(days, GrainWeek, AggregateOptions{StartOfWeek: time.Monday, Order: "asc"}, nil)
	if monday.Rows[0].Period != "2026-09-07" {
		t.Errorf("monday first week start = %s, want 2026-09-07", monday.Rows[0].Period)
	}
	if monday.Rows[1].Period != "2026-09-14" || monday.Rows[1].Tokens != 500 {
		t.Errorf("monday second week = %+v, want week of 2026-09-14 with 500 tokens", monday.Rows[1])
	}
}

func TestAggregateMonthlyAndTotals(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-08-31", Tokens: 10, Turns: 1, Input: 6, Output: 4, Cost: 1.5},
		{Day: "2026-09-01", Tokens: 20, Turns: 1, Input: 10, Output: 10, Cost: 0.5},
		{Day: "2026-09-02", Tokens: 30, Turns: 1, Input: 10, Output: 20, Cost: 2.0},
	}
	rep := Aggregate(days, GrainMonth, AggregateOptions{Order: "asc"}, nil)
	if len(rep.Rows) != 2 {
		t.Fatalf("expected 2 months, got %d", len(rep.Rows))
	}
	if rep.Rows[1].Period != "2026-09" || rep.Rows[1].Tokens != 50 {
		t.Errorf("september row = %+v", rep.Rows[1])
	}
	if rep.Totals.Tokens != 60 || rep.Totals.Input != 26 || rep.Totals.Output != 34 {
		t.Errorf("totals = %+v", rep.Totals)
	}
	if rep.Totals.Cost != 4.0 || rep.Totals.StoredCost != 4.0 {
		t.Errorf("totals cost = %+v", rep.Totals)
	}
	if rep.Totals.ActiveDays != 3 {
		t.Errorf("totals activeDays = %d, want 3", rep.Totals.ActiveDays)
	}
}

func TestAggregateOrderDefaultsToDesc(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 1, Input: 1, Turns: 1},
		{Day: "2026-09-03", Tokens: 1, Input: 1, Turns: 1},
	}
	asc := Aggregate(days, GrainDay, AggregateOptions{Order: "asc"}, nil)
	if asc.Rows[0].Period != "2026-09-01" {
		t.Errorf("asc first = %s", asc.Rows[0].Period)
	}
	def := Aggregate(days, GrainDay, AggregateOptions{}, nil)
	if def.Rows[0].Period != "2026-09-03" {
		t.Errorf("default first = %s, want newest first", def.Rows[0].Period)
	}
}

func TestAggregateFilters(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 1, Input: 1, Turns: 1},
		{Day: "2026-09-10", Tokens: 1, Input: 1, Turns: 1},
		{Day: "2026-09-16", Tokens: 1, Input: 1, Turns: 1},
	}
	window := Aggregate(days, GrainDay, AggregateOptions{Since: "20260905", Until: "2026-09-15"}, nil)
	if len(window.Rows) != 1 || window.Rows[0].Period != "2026-09-10" {
		t.Errorf("since/until window = %+v", window.Rows)
	}

	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.Local)
	last := Aggregate(days, GrainDay, AggregateOptions{Last: 3, Now: now}, nil)
	if len(last.Rows) != 1 || last.Rows[0].Period != "2026-09-16" {
		t.Errorf("--last 3 days = %+v, want only 2026-09-16", last.Rows)
	}

	lastWeek := Aggregate(days, GrainWeek, AggregateOptions{Last: 1, Now: now, StartOfWeek: time.Sunday}, nil)
	if len(lastWeek.Rows) != 1 || lastWeek.Rows[0].Period != "2026-09-13" {
		t.Errorf("--last 1 week = %+v", lastWeek.Rows)
	}
	if lastWeek.Rows[0].Tokens != 1 {
		t.Errorf("current week tokens = %d, want 1", lastWeek.Rows[0].Tokens)
	}

	lastMonth := Aggregate(days, GrainMonth, AggregateOptions{Last: 1, Now: now}, nil)
	if len(lastMonth.Rows) != 1 || lastMonth.Rows[0].Period != "2026-09" {
		t.Errorf("--last 1 month = %+v", lastMonth.Rows)
	}
}

func TestAggregateActiveDaysDeduplicates(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-16", Tokens: 10, Input: 10, Turns: 2},
		{Day: "2026-09-16", Tokens: 5, Input: 5, Turns: 1},
	}
	rep := Aggregate(days, GrainDay, AggregateOptions{}, nil)
	if rep.Rows[0].ActiveDays != 1 {
		t.Errorf("activeDays = %d, want 1", rep.Rows[0].ActiveDays)
	}
	if rep.Rows[0].Tokens != 15 {
		t.Errorf("tokens = %d, want 15", rep.Rows[0].Tokens)
	}
}

type stubPricer struct {
	cost    float64
	missing []string
	calls   int
}

func (s *stubPricer) PriceDay(day DailyRow) (float64, []string) {
	s.calls++
	return s.cost, s.missing
}

func TestAggregatePricingRule(t *testing.T) {
	pricer := &stubPricer{cost: 2.5, missing: []string{"unknown-model"}}
	days := []DailyRow{
		{Day: "2026-09-15", Tokens: 100, Input: 60, Output: 40, Turns: 1, Models: map[string]ModelTokens{"test-model": {Input: 60, Output: 40}}}, // estimated
		{Day: "2026-09-16", Tokens: 100, Input: 60, Output: 40, Turns: 1, Cost: 9.0},                                                             // stored, no pricing pass
		{Day: "2026-09-17", Tokens: 0, Turns: 0}, // no telemetry, skipped entirely
	}
	rep := Aggregate(days, GrainDay, AggregateOptions{Order: "asc"}, pricer)
	if len(rep.Rows) != 2 {
		t.Fatalf("expected 2 rows, empty day skipped, got %d", len(rep.Rows))
	}
	if pricer.calls != 1 {
		t.Errorf("pricer calls = %d, want 1", pricer.calls)
	}
	if rep.Rows[0].EstimatedCost != 2.5 || rep.Rows[0].Cost != 2.5 {
		t.Errorf("estimated row = %+v", rep.Rows[0])
	}
	if len(rep.Rows[0].MissingPricing) != 1 || rep.Rows[0].MissingPricing[0] != "unknown-model" {
		t.Errorf("missing pricing = %v", rep.Rows[0].MissingPricing)
	}
	if rep.Rows[1].EstimatedCost != 0 || rep.Rows[1].Cost != 9.0 {
		t.Errorf("stored row = %+v", rep.Rows[1])
	}
	if rep.Totals.Cost != 11.5 || rep.Totals.EstimatedCost != 2.5 || rep.Totals.StoredCost != 9.0 {
		t.Errorf("totals = %+v", rep.Totals)
	}
}

func TestAggregateSkipsDaysWithoutTokenData(t *testing.T) {
	days := []DailyRow{
		// A step count: Tokens equals Turns and nothing else is recorded, so it
		// is activity, not tokens, and it is skipped.
		{Day: "2026-09-10", Tokens: 21, Turns: 21},
		{Day: "2026-09-11", Tokens: 0, Turns: 2},                                                  // empty session, turns but no tokens
		{Day: "2026-09-12", Tokens: 0, Turns: 0},                                                  // fully empty
		{Day: "2026-09-13", Tokens: 0, Turns: 0, Models: map[string]ModelTokens{"ghost": {}}},     // zero-count model only
		{Day: "2026-09-14", Tokens: 10, Input: 10, Turns: 1},                                      // kept: classified tokens
		{Day: "2026-09-15", Tokens: 0, Turns: 1, Cost: 1.5},                                       // kept: recorded cost
		{Day: "2026-09-16", Tokens: 0, Turns: 0, Models: map[string]ModelTokens{"m": {Input: 5}}}, // kept: positive model
	}
	rep := Aggregate(days, GrainDay, AggregateOptions{Order: "asc"}, nil)
	if len(rep.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %d: %+v", len(rep.Rows), rep.Rows)
	}
	for i, want := range []string{"2026-09-14", "2026-09-15", "2026-09-16"} {
		if rep.Rows[i].Period != want {
			t.Errorf("row %d = %s, want %s", i, rep.Rows[i].Period, want)
		}
	}
	if rep.Totals.Tokens != 10 || rep.Totals.Turns != 2 || rep.Totals.ActiveDays != 2 {
		t.Errorf("totals = %+v, want 10 tokens / 2 turns / 2 active days", rep.Totals)
	}
	// The day with no model and no recorded cost is what no price can cover.
	if rep.Totals.UnattributedTokens != 10 {
		t.Errorf("unattributed = %d, want 10 (the model-less token day)", rep.Totals.UnattributedTokens)
	}
	if rep.Totals.Cost != 1.5 || rep.Totals.StoredCost != 1.5 {
		t.Errorf("totals cost = %+v, want 1.5 stored", rep.Totals)
	}
}

func TestAggregateSkipsZeroCountModels(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-14", Tokens: 10, Input: 10, Turns: 1, Models: map[string]ModelTokens{
			"real":  {Input: 10},
			"ghost": {},
		}},
	}
	rep := Aggregate(days, GrainDay, AggregateOptions{}, nil)
	if len(rep.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rep.Rows))
	}
	if _, ok := rep.Rows[0].Models["ghost"]; ok {
		t.Errorf("zero-count model leaked into row: %v", rep.Rows[0].Models)
	}
	if _, ok := rep.Totals.Models["ghost"]; ok {
		t.Errorf("zero-count model leaked into totals: %v", rep.Totals.Models)
	}
	if got := rep.Totals.Models["real"].Total(); got != 10 {
		t.Errorf("real model total = %d, want 10", got)
	}
}

func TestAggregateTotalsMatchAcrossGrains(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-08-31", Tokens: 10, Input: 6, Output: 4, Turns: 1, Cost: 1.5},
		{Day: "2026-09-01", Tokens: 20, Input: 10, Output: 10, Turns: 1, Models: map[string]ModelTokens{"m": {Input: 10, Output: 10}}},
		{Day: "2026-09-02", Tokens: 30, Input: 10, Output: 20, Turns: 1, Cost: 2.0},
		// A session total with no model breakdown is a token day: it is counted,
		// it cannot be priced, and the footer is told about it. Its Tokens and
		// Turns differ, which is what separates it from an activity count.
		{Day: "2026-09-03", Tokens: 7, Turns: 1},
	}
	pricer := &stubPricer{cost: 2.5, missing: []string{"unknown-model"}}
	daily := Aggregate(days, GrainDay, AggregateOptions{}, pricer)
	if pricer.calls != 2 {
		t.Fatalf("daily pricer calls = %d, want 2 (the two days with no recorded cost)", pricer.calls)
	}
	if daily.Totals.UnattributedTokens != 7 {
		t.Errorf("unattributed = %d, want 7 (the model-less day)", daily.Totals.UnattributedTokens)
	}
	pricer.calls = 0
	weekly := Aggregate(days, GrainWeek, AggregateOptions{StartOfWeek: time.Sunday}, pricer)
	pricer.calls = 0
	monthly := Aggregate(days, GrainMonth, AggregateOptions{}, pricer)
	for _, pair := range [][2]PeriodRow{{daily.Totals, monthly.Totals}, {weekly.Totals, monthly.Totals}} {
		a, b := pair[0], pair[1]
		if a.Tokens != b.Tokens || a.Input != b.Input || a.Output != b.Output ||
			a.Turns != b.Turns || a.ActiveDays != b.ActiveDays ||
			a.StoredCost != b.StoredCost || a.EstimatedCost != b.EstimatedCost || a.Cost != b.Cost {
			t.Errorf("grain totals differ:\n a=%+v\n b=%+v", a, b)
		}
		if len(a.Models) != len(b.Models) || a.Models["m"].Total() != b.Models["m"].Total() {
			t.Errorf("grain model totals differ: %v vs %v", a.Models, b.Models)
		}
	}
	if monthly.Totals.Tokens != 67 || monthly.Totals.StoredCost != 3.5 || monthly.Totals.EstimatedCost != 5.0 {
		t.Errorf("totals = %+v, want 67 tokens / 3.5 stored / 5.0 estimated", monthly.Totals)
	}
}

func TestTopModels(t *testing.T) {
	models := map[string]ModelTokens{
		"small": {Input: 10},
		"big":   {Input: 100, Output: 100},
		"huge":  {Input: 300},
		"aab":   {Input: 50},
		"aaa":   {Input: 50},
	}
	got := TopModels(models)
	want := []string{"huge", "big", "aaa", "aab", "small"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("TopModels = %v, want %v", got, want)
		}
	}
}

// TestIsActivityOnly pins the rule that keeps step counts out of token totals.
// Every activity loader writes a call count into Tokens and the same number
// into Turns, so the shape is distinguishable from a token total.
func TestIsActivityOnly(t *testing.T) {
	cases := []struct {
		name     string
		day      DailyRow
		activity bool // the count is calls, not tokens
		data     bool // the row carries tokens or cost
	}{
		{"step count", DailyRow{Day: "2026-09-15", Tokens: 500, Turns: 500}, true, false},
		{"message count", DailyRow{Day: "2026-09-15", Tokens: 14, Turns: 14}, true, false},
		{"session total", DailyRow{Day: "2026-09-15", Tokens: 5000, Turns: 3}, false, true},
		{"classified tokens", DailyRow{Day: "2026-09-15", Tokens: 500, Input: 500, Turns: 500}, false, true},
		{"recorded cost", DailyRow{Day: "2026-09-15", Tokens: 500, Turns: 500, Cost: 1.5}, false, true},
		{"model present", DailyRow{Day: "2026-09-15", Tokens: 500, Turns: 500, Models: map[string]ModelTokens{"m": {Input: 500}}}, false, true},
		{"turns without tokens", DailyRow{Day: "2026-09-15", Tokens: 0, Turns: 3}, false, false},
		{"empty", DailyRow{Day: "2026-09-15"}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isActivityOnly(tc.day); got != tc.activity {
				t.Errorf("isActivityOnly(%+v) = %v, want %v", tc.day, got, tc.activity)
			}
			if got := hasTokenData(tc.day); got != tc.data {
				t.Errorf("hasTokenData(%+v) = %v, want %v", tc.day, got, tc.data)
			}
		})
	}
}
