// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"testing"
	"time"
)

func basisFixture() []DailyRow {
	rows := make([]DailyRow, 0, 10)
	for i := 1; i <= 10; i++ {
		row := DailyRow{Day: time.Date(2026, 1, i, 0, 0, 0, 0, time.UTC).Format("2006-01-02")}
		if i <= 4 {
			row.Tokens = 25
			row.Input = 25
			row.Turns = 1
		}
		rows = append(rows, row)
	}
	return rows
}

func TestStatsMeanBasis(t *testing.T) {
	rep := AggregateStats(basisFixture(), StatsOptions{}, nil)
	if rep.MeanBasis != "active-day" {
		t.Fatalf("meanBasis=%q", rep.MeanBasis)
	}
	if rep.Days != 4 || rep.Mean != 25 {
		t.Fatalf("days=%d mean=%v, want 4/25", rep.Days, rep.Mean)
	}
}

func TestTrendMeanBasis(t *testing.T) {
	rep := AggregateTrend(basisFixture(), TrendOptions{Now: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)}, nil)
	if rep.MeanBasis != "calendar-day" {
		t.Fatalf("meanBasis=%q", rep.MeanBasis)
	}
	if len(rep.Points) != 10 || rep.Mean != 10 || rep.ActiveDays != 4 {
		t.Fatalf("points=%d mean=%v active=%d, want 10/10/4", len(rep.Points), rep.Mean, rep.ActiveDays)
	}
	if rep.Projection == nil || rep.Projection.BandMethod != "residual-sd-1.96" {
		t.Fatalf("projection=%+v", rep.Projection)
	}
}
