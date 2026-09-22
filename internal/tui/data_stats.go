// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"math"
	"sort"

	"github.com/jadmadi/thermal/internal/thermal"
)

// StatsView is the data behind the Stats tab. Every number comes from
// thermal.AggregateStats or thermal.AggregateTrend, the functions behind
// `thermal stats` and `thermal trend`, so a reader can reproduce any figure
// with the static command and the same window.
type StatsView struct {
	Range     Range
	Metric    Metric
	Start     string
	End       string
	Days      int
	Total     float64
	Mean      float64
	Median    float64
	P90       float64
	Max       float64
	Histogram []HistBin
	UseLog    bool
	LogReason string
	Weekday   []thermal.WeekdayStat
	TopDays   []thermal.DayValue
	Outliers  []thermal.DayValue
	Threshold float64
	Slope     float64
	MeanTrend float64
	FirstDay  string
	LastDay   string
	Project   *thermal.Projection
	Estimated bool
}

// HistBin is one histogram bar with the labels the view prints.
type HistBin struct {
	Label string
	From  float64
	To    float64
	Count int
}

// logRatio is the threshold at which the histogram switches to a log scale.
// A maximum more than ten times the median means the linear bars would put
// every ordinary day in the first bin and hide the shape.
const logRatio = 10

// BuildStats aggregates the daily distribution and the trend for the window.
func (a Adapter) BuildStats(r Range, m Metric) StatsView {
	start, end := a.window(r)
	view := StatsView{Range: r, Metric: m, Start: start, End: end}

	var days []thermal.DailyRow
	for _, t := range a.Tools {
		days = append(days, t.Days...)
	}

	stats := thermal.AggregateStats(days, thermal.StatsOptions{
		Since:  start,
		Until:  end,
		Metric: string(m),
	}, a.Pricer)
	trend := thermal.AggregateTrend(days, thermal.TrendOptions{
		Since:  start,
		Until:  end,
		Metric: string(m),
	}, a.Pricer)

	view.Days = stats.Days
	view.Total = stats.Total
	view.Mean = stats.Mean
	view.Median = stats.Median
	view.P90 = stats.P90
	view.Max = stats.Max
	view.Weekday = stats.Weekday
	view.TopDays = stats.TopDays
	view.Outliers = stats.Outliers
	view.Threshold = stats.OutlierThreshold
	view.Estimated = stats.Estimated
	view.Slope = trend.Slope
	view.MeanTrend = trend.Mean
	view.FirstDay = trend.FirstDay
	view.LastDay = trend.LastDay
	view.Project = trend.Projection

	// The scale decision is the view's own: the aggregator always returns
	// linear bins, and a long tail needs a log axis to be readable.
	view.UseLog = stats.Median > 0 && stats.Max > logRatio*stats.Median
	if view.UseLog {
		view.LogReason = fmt.Sprintf("log scale: the largest day is %.0f times the median", stats.Max/stats.Median)
	}
	view.Histogram = buildHistogram(days, start, end, m, a.Pricer, view.UseLog, stats.Histogram)
	return view
}

// buildHistogram bins the window's daily metric. Linear bins come from the
// aggregator when the scale is linear; the log case is derived here because a
// logarithmic axis is a presentation choice, not an aggregation.
func buildHistogram(days []thermal.DailyRow, start, end string, m Metric, pricer thermal.Pricer, useLog bool, linear []thermal.HistogramBin) []HistBin {
	const bins = 12

	values := make([]float64, 0, len(days))
	for _, d := range days {
		if !inWindow(d.Day, start, end) {
			continue
		}
		if m == MetricCost {
			cost, _ := dayMetricCost(d, pricer)
			if cost > 0 {
				values = append(values, cost)
			}
			continue
		}
		if d.Tokens > 0 {
			values = append(values, float64(d.Tokens))
		}
	}
	if len(values) == 0 {
		return nil
	}

	var out []HistBin
	if !useLog {
		for _, bin := range linear {
			out = append(out, HistBin{
				Label: fmt.Sprintf("%s-%s", shortMetric(bin.From, m), shortMetric(bin.To, m)),
				From:  bin.From,
				To:    bin.To,
				Count: bin.Count,
			})
		}
		return out
	}

	sort.Float64s(values)
	lo := math.Log10(math.Max(values[0], 1))
	hi := math.Log10(values[len(values)-1])
	if hi <= lo {
		hi = lo + 1
	}
	edges := make([]float64, bins+1)
	for i := range edges {
		edges[i] = math.Pow(10, lo+(hi-lo)*float64(i)/float64(bins))
	}
	counts := make([]int, bins)
	for _, v := range values {
		idx := sort.SearchFloat64s(edges, v)
		if idx >= bins {
			idx = bins - 1
		}
		counts[idx]++
	}
	for i := 0; i < bins; i++ {
		out = append(out, HistBin{
			Label: fmt.Sprintf("%s-%s", shortMetric(edges[i], m), shortMetric(edges[i+1], m)),
			From:  edges[i],
			To:    edges[i+1],
			Count: counts[i],
		})
	}
	return out
}

// dayMetricCost applies the same rule the reports use: recorded cost wins, and
// a day without one is priced only when the source named its models.
func dayMetricCost(day thermal.DailyRow, pricer thermal.Pricer) (float64, bool) {
	if day.Cost > 0 {
		return day.Cost, false
	}
	if pricer == nil || day.Tokens == 0 || len(day.Models) == 0 {
		return 0, false
	}
	cost, _ := pricer.PriceDay(day)
	return cost, cost > 0
}

// shortMetric prints a bin edge compactly. Token counts get the compact suffix
// and costs get two decimals, so a bin label stays readable at 80 columns.
func shortMetric(v float64, m Metric) string {
	if m == MetricCost {
		if v < 0.01 {
			return fmt.Sprintf("%.4f", v)
		}
		return fmt.Sprintf("%.2f", v)
	}
	return thermal.CompactNumber(int64(v))
}
