// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"math"
	"time"
)

// TrendOptions controls the daily trend fit and projection. Metric is "tokens"
// or "cost".
type TrendOptions struct {
	Since  string
	Until  string
	Last   int
	Metric string
	Now    time.Time
}

// Projection is the month-end estimate from the fitted trend. Low and High
// form an approximate 95 percent band from the residual spread.
type Projection struct {
	To       string  `json:"to"`
	Days     int     `json:"days"`
	Expected float64 `json:"expected"`
	Low      float64 `json:"low"`
	High     float64 `json:"high"`
}

// TrendReport is the payload behind thermal trend. Points include quiet days
// inside the window as zeroes, so a gap lowers the slope instead of hiding.
type TrendReport struct {
	Type       string      `json:"type"`
	Metric     string      `json:"metric"`
	Points     []DayValue  `json:"data"`
	Slope      float64     `json:"slopePerDay"`
	Intercept  float64     `json:"intercept"`
	Mean       float64     `json:"mean"`
	FirstDay   string      `json:"firstDay"`
	LastDay    string      `json:"lastDay"`
	Projection *Projection `json:"projection,omitempty"`
	Estimated  bool        `json:"estimatedCost,omitempty"`
}

// AggregateTrend fits a least-squares line over the window's daily values and
// projects it to the end of the current calendar month.
func AggregateTrend(days []DailyRow, opts TrendOptions, pricer Pricer) TrendReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	metric := normalMixMetric(opts.Metric)

	byDay := make(map[string]float64)
	estimated := false
	for _, day := range days {
		t, ok := ParseDay(day.Day)
		if !ok || !inWindow(t, since, until, lastStart) {
			continue
		}
		if metric == "cost" {
			value, est := dayCost(day, pricer)
			if est {
				estimated = true
			}
			byDay[day.Day] += value
			continue
		}
		if isActivityOnly(day) {
			continue
		}
		byDay[day.Day] += float64(day.Tokens)
	}

	rep := TrendReport{Type: "trend", Metric: metric, Estimated: estimated}
	if len(byDay) == 0 {
		return rep
	}

	first, last := "", ""
	for day := range byDay {
		if first == "" || day < first {
			first = day
		}
		if day > last {
			last = day
		}
	}
	rep.FirstDay, rep.LastDay = first, last

	start, _ := ParseDay(first)
	end, _ := ParseDay(last)
	for t := start; !t.After(end); t = t.AddDate(0, 0, 1) {
		day := t.Format("2006-01-02")
		v := byDay[day]
		rep.Points = append(rep.Points, DayValue{Day: day, Value: v})
		rep.Mean += v
	}
	rep.Mean /= float64(len(rep.Points))

	n := float64(len(rep.Points))
	rep.Slope, rep.Intercept = leastSquares(rep.Points)

	if proj, ok := projectMonthEnd(rep, n, opts.Now); ok {
		rep.Projection = &proj
	}
	return rep
}

// leastSquares returns the slope per day and intercept for points evenly
// spaced by day index.
func leastSquares(points []DayValue) (float64, float64) {
	n := float64(len(points))
	if n < 2 {
		if n == 1 {
			return 0, points[0].Value
		}
		return 0, 0
	}
	var sumX, sumY, sumXY, sumXX float64
	for i, p := range points {
		x := float64(i)
		sumX += x
		sumY += p.Value
		sumXY += x * p.Value
		sumXX += x * x
	}
	denom := n*sumXX - sumX*sumX
	if denom == 0 {
		return 0, sumY / n
	}
	slope := (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n
	return slope, intercept
}

// residualStd returns the standard deviation of the fit residuals.
func residualStd(points []DayValue, slope, intercept float64) float64 {
	if len(points) < 3 {
		return 0
	}
	var sum float64
	for i, p := range points {
		d := p.Value - (intercept + slope*float64(i))
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(points)-2))
}

func projectMonthEnd(rep TrendReport, n float64, now time.Time) (Projection, bool) {
	if n < 2 {
		return Projection{}, false
	}
	if now.IsZero() {
		now = time.Now()
	}
	lastDay, ok := ParseDay(rep.LastDay)
	if !ok {
		return Projection{}, false
	}
	monthEnd := time.Date(lastDay.Year(), lastDay.Month()+1, 0, 0, 0, 0, 0, lastDay.Location())
	if !monthEnd.After(lastDay) {
		return Projection{}, false
	}
	days := int(monthEnd.Sub(lastDay).Hours() / 24)

	x := n - 1 + float64(days)
	expected := rep.Intercept + rep.Slope*x
	band := 1.96 * residualStd(rep.Points, rep.Slope, rep.Intercept)

	low, high := expected-band, expected+band
	if low < 0 {
		low = 0
	}
	return Projection{
		To:       monthEnd.Format("2006-01-02"),
		Days:     days,
		Expected: expected,
		Low:      low,
		High:     high,
	}, true
}
