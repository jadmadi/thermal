package thermal

import (
	"testing"
	"time"
)

func TestAggregateTrendFitsKnownLine(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 10},
		{Day: "2026-09-02", Tokens: 20},
		{Day: "2026-09-03", Tokens: 30},
		{Day: "2026-09-04", Tokens: 40},
	}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.Local)
	rep := AggregateTrend(days, TrendOptions{Metric: "tokens", Now: now}, nil)

	if rep.Slope != 10 {
		t.Errorf("slope = %v, want 10 per day", rep.Slope)
	}
	if rep.Intercept != 10 {
		t.Errorf("intercept = %v, want 10", rep.Intercept)
	}
	if rep.Mean != 25 {
		t.Errorf("mean = %v, want 25", rep.Mean)
	}
	if len(rep.Points) != 4 || rep.FirstDay != "2026-09-01" || rep.LastDay != "2026-09-04" {
		t.Errorf("points = %d, range %s..%s", len(rep.Points), rep.FirstDay, rep.LastDay)
	}
	if rep.Projection == nil {
		t.Fatal("expected a month-end projection")
	}
	if rep.Projection.To != "2026-09-30" || rep.Projection.Days != 26 {
		t.Errorf("projection = %+v, want to 2026-09-30 over 26 days", rep.Projection)
	}
	// Perfect fit: x = 3 + 26 = 29, y = 10 + 10*29.
	if rep.Projection.Expected != 300 {
		t.Errorf("projected = %v, want 300", rep.Projection.Expected)
	}
	if rep.Projection.Low != 300 || rep.Projection.High != 300 {
		t.Errorf("band = %v..%v, want zero width on a perfect fit", rep.Projection.Low, rep.Projection.High)
	}
}

func TestAggregateTrendFlatAndGaps(t *testing.T) {
	flat := AggregateTrend([]DailyRow{
		{Day: "2026-09-01", Tokens: 100},
		{Day: "2026-09-02", Tokens: 100},
		{Day: "2026-09-03", Tokens: 100},
	}, TrendOptions{Metric: "tokens", Now: time.Date(2026, 9, 4, 0, 0, 0, 0, time.Local)}, nil)
	if flat.Slope != 0 || flat.Intercept != 100 {
		t.Errorf("flat fit = slope %v intercept %v", flat.Slope, flat.Intercept)
	}
	if flat.Projection == nil || flat.Projection.Expected != 100 {
		t.Errorf("flat projection = %+v", flat.Projection)
	}

	// A quiet day inside the range appears as a zero point.
	gap := AggregateTrend([]DailyRow{
		{Day: "2026-09-01", Tokens: 10},
		{Day: "2026-09-03", Tokens: 30},
	}, TrendOptions{Metric: "tokens", Now: time.Date(2026, 9, 4, 0, 0, 0, 0, time.Local)}, nil)
	if len(gap.Points) != 3 {
		t.Fatalf("points = %d, want 3 with the quiet day filled", len(gap.Points))
	}
	if gap.Points[1].Day != "2026-09-02" || gap.Points[1].Value != 0 {
		t.Errorf("middle point = %+v, want a zero on 2026-09-02", gap.Points[1])
	}
	if gap.Slope != 10 {
		t.Errorf("slope with the gap = %v, want 10", gap.Slope)
	}
}

func TestAggregateTrendEmpty(t *testing.T) {
	rep := AggregateTrend(nil, TrendOptions{Metric: "tokens"}, nil)
	if len(rep.Points) != 0 || rep.Projection != nil {
		t.Errorf("empty trend = %+v", rep)
	}
}

func TestAggregateTrendExcludesActivityOnly(t *testing.T) {
	days := []DailyRow{
		{Day: "2026-09-01", Tokens: 1000, Turns: 1, Input: 800, Output: 200},
		{Day: "2026-09-02", Tokens: 417, Turns: 417}, // activity-only
	}
	rep := AggregateTrend(days, TrendOptions{Metric: "tokens"}, nil)
	if len(rep.Points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(rep.Points))
	}
	if rep.Mean != 1000 {
		t.Errorf("expected mean 1000, got %v", rep.Mean)
	}
}
