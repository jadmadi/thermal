package thermal

import (
	"testing"
)

func TestAggregateStatsDistribution(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-07", Tokens: 10},
		{Day: "2026-09-08", Tokens: 20},
		{Day: "2026-09-09", Tokens: 30},
		{Day: "2026-09-10", Tokens: 40},
		{Day: "2026-09-11", Tokens: 1000},
		{Day: "2026-09-12", Tokens: 0}, // quiet day, excluded from percentiles
	}
	rep := AggregateStats(days, StatsOptions{Metric: "tokens"}, nil)

	if rep.Days != 5 {
		t.Errorf("active days = %d, want 5", rep.Days)
	}
	if rep.Total != 1100 || rep.Mean != 220 {
		t.Errorf("total/mean = %v/%v, want 1100/220", rep.Total, rep.Mean)
	}
	if rep.Median != 30 {
		t.Errorf("median = %v, want 30", rep.Median)
	}
	if rep.P90 != 40 {
		t.Errorf("p90 = %v, want 40 by nearest rank", rep.P90)
	}
	if rep.Max != 1000 {
		t.Errorf("max = %v, want 1000", rep.Max)
	}
	// MAD of {10,20,30,40,1000} around 30 is 10, so the threshold is 50.
	if rep.OutlierThreshold != 50 {
		t.Errorf("threshold = %v, want 50", rep.OutlierThreshold)
	}
	if len(rep.Outliers) != 1 || rep.Outliers[0].Day != "2026-09-11" {
		t.Errorf("outliers = %+v, want only 2026-09-11", rep.Outliers)
	}
	if len(rep.TopDays) == 0 || rep.TopDays[0].Value != 1000 {
		t.Errorf("top days = %+v", rep.TopDays)
	}
	if len(rep.Weekday) != 7 {
		t.Fatalf("weekday profile length = %d", len(rep.Weekday))
	}
	weekdayDays := 0
	for _, w := range rep.Weekday {
		weekdayDays += w.Days
	}
	if weekdayDays != 5 {
		t.Errorf("weekday day count = %d, want 5", weekdayDays)
	}
	if len(rep.Histogram) != histogramBins {
		t.Fatalf("histogram bins = %d", len(rep.Histogram))
	}
	// Bin 0 covers 0 to 100, holding 10, 20, 30 and 40. The max lands in the
	// last bin.
	if rep.Histogram[0].Count != 4 {
		t.Errorf("first bin count = %d, want 4", rep.Histogram[0].Count)
	}
	if rep.Histogram[histogramBins-1].Count != 1 {
		t.Errorf("last bin count = %d, want 1", rep.Histogram[histogramBins-1].Count)
	}
}

func TestAggregateStatsSumsToolRowsPerDay(t *testing.T) {
	// Loader output carries one row per tool per day. The same day twice must
	// count once, with the values summed.
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 100},
		{Day: "2026-09-01", Tokens: 300},
		{Day: "2026-09-02", Tokens: 200},
	}
	rep := AggregateStats(days, StatsOptions{Metric: "tokens"}, nil)
	if rep.Days != 2 {
		t.Fatalf("active days = %d, want 2", rep.Days)
	}
	if rep.Total != 600 {
		t.Errorf("total = %v, want 600", rep.Total)
	}
	if rep.Median != 300 {
		t.Errorf("median = %v, want 300 from day sums 400 and 200", rep.Median)
	}
	if rep.Max != 400 {
		t.Errorf("max = %v, want 400", rep.Max)
	}
}

func TestAggregateStatsEmptyAndFlat(t *testing.T) {
	empty := AggregateStats(nil, StatsOptions{Metric: "tokens"}, nil)
	if empty.Days != 0 || empty.Total != 0 || empty.Median != 0 || len(empty.Histogram) != 0 {
		t.Errorf("empty report = %+v", empty)
	}

	flat := AggregateStats([]DailyRow{
		{Day: "2026-09-01", Tokens: 100},
		{Day: "2026-09-02", Tokens: 100},
		{Day: "2026-09-03", Tokens: 100},
	}, StatsOptions{Metric: "tokens"}, nil)
	if flat.Median != 100 || flat.Mean != 100 {
		t.Errorf("flat report = %+v", flat)
	}
	if flat.OutlierThreshold != 100 {
		t.Errorf("flat threshold = %v, want the median when MAD is 0", flat.OutlierThreshold)
	}
	if len(flat.Outliers) != 0 {
		t.Errorf("flat report should have no outliers, got %+v", flat.Outliers)
	}
}

func TestAggregateStatsCostFlagsEstimate(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 1000, Cost: 2.0},
		{Day: "2026-09-02", Tokens: 1_000_000, Models: map[string]ModelTokens{"m": {Input: 1_000_000}}},
	}
	pricer := ratePricer{rates: map[string]float64{"m": 1.5}}
	rep := AggregateStats(days, StatsOptions{Metric: "cost"}, pricer)

	if !rep.Estimated {
		t.Error("mixed stored and estimated cost must flag the estimate")
	}
	if rep.Total != 3.5 {
		t.Errorf("total = %v, want 3.5", rep.Total)
	}
	if rep.Median == 0 {
		t.Error("median should be non-zero")
	}
}

func TestAggregateStatsExcludesActivityOnly(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 1000, Turns: 1, Input: 800, Output: 200},
		{Day: "2026-09-02", Tokens: 417, Turns: 417}, // activity-only
	}
	rep := AggregateStats(days, StatsOptions{Metric: "tokens"}, nil)
	if rep.Days != 1 {
		t.Fatalf("expected 1 day, got %d", rep.Days)
	}
	if rep.Total != 1000 {
		t.Errorf("expected total 1000, got %v", rep.Total)
	}
}

func TestAggregateStatsTokenComposition(t *testing.T) {
	days := []DailyRow{
		{
			Day:    "2026-09-01",
			Tokens: 10000,
			Turns:  5,
			Input:  500,
			Output: 300,
			Cache:  9200,
			Models: map[string]ModelTokens{
				"claude-sonnet": {
					Input:      500,
					Output:     300,
					CacheRead:  9000,
					CacheWrite: 200,
				},
			},
		},
	}
	rep := AggregateStats(days, StatsOptions{Metric: "tokens"}, nil)
	if rep.Composition == nil {
		t.Fatal("expected Composition to be non-nil")
	}
	c := rep.Composition
	if c.UncachedInput != 500 {
		t.Errorf("expected uncached input 500, got %d", c.UncachedInput)
	}
	if c.Output != 300 {
		t.Errorf("expected output 300, got %d", c.Output)
	}
	if c.CacheRead != 9000 {
		t.Errorf("expected cache read 9000, got %d", c.CacheRead)
	}
	if c.CacheWrite != 200 {
		t.Errorf("expected cache write 200, got %d", c.CacheWrite)
	}
	if c.Total != 10000 {
		t.Errorf("expected total 10000, got %d", c.Total)
	}
	if diff := c.CacheHitRate - (9000.0 / 9700.0); diff > 0.001 || diff < -0.001 {
		t.Errorf("expected cache hit rate ~0.9278, got %f", c.CacheHitRate)
	}
}
