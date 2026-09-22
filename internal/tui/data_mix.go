// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// MixSeriesCap is how many series the stacked chart draws. Everything past the
// cap folds into one bucket, so shares still sum to 100 percent and narrow
// terminals stay readable.
const MixSeriesCap = 4

// seriesRunes are the per-series glyphs. The palette has one accent, so runes
// carry the distinction and the legend carries the names: a reader with no
// colour still separates the series.
var seriesRunes = []string{"█", "▓", "▒", "░", "▚", "▪"}

// mixGrain picks a bucket size that keeps a chart inside a terminal. Daily
// buckets across a year would be noise, and monthly buckets across a month
// would be one bar.
func mixGrain(r Range) thermal.Grain {
	switch r {
	case Range30d:
		return thermal.GrainDay
	case Range90d:
		return thermal.GrainWeek
	default:
		return thermal.GrainMonth
	}
}

// MixView is the data behind the Mix tab.
type MixView struct {
	Range      Range
	Metric     Metric
	By         string // "tool" or "model"
	Grain      thermal.Grain
	Start      string
	End        string
	Series     []MixSerie
	Buckets    []MixBucketView
	Dominant   string
	Switches   int
	SwitchesMo map[string]int
	Conc       float64
	Active     int
	Total      float64
	Estimated  bool
}

// MixSerie is one stacked series after the cap is applied.
type MixSerie struct {
	Name   string
	Total  float64
	Share  float64
	Values []float64
}

// MixBucketView is one time bucket of the stacked chart.
type MixBucketView struct {
	Period string
	Total  float64
	Values []float64
}

// BuildMix aggregates the tool mix for the window. It reuses
// thermal.AggregateToolMix, the function behind `thermal mix`, so concentration
// and switch counts cannot drift from the static command.
func (a Adapter) BuildMix(r Range, m Metric) MixView {
	start, end := a.window(r)
	view := MixView{Range: r, Metric: m, By: "tool", Grain: mixGrain(r), Start: start, End: end}

	batches := make([]thermal.ToolDays, 0, len(a.Tools))
	for _, t := range a.Tools {
		batches = append(batches, thermal.ToolDays{Tool: t.Name, Days: t.Days})
	}
	rep := thermal.AggregateToolMix(batches, thermal.MixOptions{
		Since:  start,
		Until:  end,
		Grain:  view.Grain,
		By:     "tool",
		Metric: string(m),
	}, a.Pricer)

	view.Dominant = rep.Dominant
	view.Switches = rep.Switches
	view.SwitchesMo = rep.SwitchesByMonth
	view.Conc = rep.Concentration
	view.Active = rep.ActiveSeries
	view.Total = rep.Total
	view.Estimated = rep.Estimated

	// Fold everything past the cap into one bucket, then recompute shares so
	// the legend still accounts for the whole window.
	series := rep.Series
	var other MixSerie
	if len(series) > MixSeriesCap {
		other.Name = "other"
		for _, s := range series[MixSeriesCap:] {
			other.Total += s.Total
			if other.Values == nil {
				other.Values = make([]float64, len(s.Buckets))
			}
			for i, v := range s.Buckets {
				if i < len(other.Values) {
					other.Values[i] += v
				}
			}
		}
		series = series[:MixSeriesCap]
	}
	for _, s := range series {
		view.Series = append(view.Series, MixSerie{Name: s.Name, Total: s.Total, Share: s.Share, Values: s.Buckets})
	}
	if other.Total > 0 {
		view.Series = append(view.Series, other)
	}
	renormalizeMix(&view)

	for _, b := range rep.Buckets {
		vb := MixBucketView{Period: b.Period, Total: b.Total}
		vb.Values = append(vb.Values, b.Values[:minInt(len(b.Values), MixSeriesCap)]...)
		if len(b.Values) > MixSeriesCap {
			var rest float64
			for _, v := range b.Values[MixSeriesCap:] {
				rest += v
			}
			vb.Values = append(vb.Values, rest)
		}
		view.Buckets = append(view.Buckets, vb)
	}
	return view
}

// renormalizeMix recomputes shares from the values on screen, so the legend
// always adds up to 100 percent including the other bucket.
func renormalizeMix(v *MixView) {
	var total float64
	for _, s := range v.Series {
		total += s.Total
	}
	if total <= 0 {
		return
	}
	for i := range v.Series {
		v.Series[i].Share = v.Series[i].Total / total * 100
	}
}

// BuildModelMix aggregates the model mix. It reuses thermal.AggregateModelMix,
// which is the function behind `thermal mix --by model`.
func (a Adapter) BuildModelMix(r Range, m Metric) MixView {
	start, end := a.window(r)
	view := MixView{Range: r, Metric: m, By: "model", Grain: mixGrain(r), Start: start, End: end}

	var days []thermal.DailyRow
	for _, t := range a.Tools {
		days = append(days, t.Days...)
	}
	rep := thermal.AggregateModelMix(days, thermal.MixOptions{
		Since:  start,
		Until:  end,
		Grain:  view.Grain,
		By:     "model",
		Metric: string(m),
	}, a.Pricer)

	view.Dominant = rep.Dominant
	view.Switches = rep.Switches
	view.SwitchesMo = rep.SwitchesByMonth
	view.Conc = rep.Concentration
	view.Active = rep.ActiveSeries
	view.Total = rep.Total
	view.Estimated = rep.Estimated

	series := rep.Series
	var other MixSerie
	if len(series) > MixSeriesCap {
		other.Name = "other"
		for _, s := range series[MixSeriesCap:] {
			other.Total += s.Total
			if other.Values == nil {
				other.Values = make([]float64, len(s.Buckets))
			}
			for i, v := range s.Buckets {
				if i < len(other.Values) {
					other.Values[i] += v
				}
			}
		}
		series = series[:MixSeriesCap]
	}
	for _, s := range series {
		view.Series = append(view.Series, MixSerie{Name: s.Name, Total: s.Total, Share: s.Share, Values: s.Buckets})
	}
	if other.Total > 0 {
		view.Series = append(view.Series, other)
	}
	renormalizeMix(&view)
	return view
}

// ModelsView is the data behind the Models tab.
type ModelsView struct {
	Range   Range
	Metric  Metric
	Sort    ModelSort
	Start   string
	End     string
	Rows    []ModelRowView
	Tokens  int64
	Cost    float64
	EstCost float64
	Tools   []ToolShare
	Shares  []Share
	CovNote string
}

// ModelRowView is one model's line.
type ModelRowView struct {
	Name   string
	Tools  []string
	Tokens int64
	Input  int64
	Output int64
	Cache  int64
	Cost   float64
	Est    float64
	Days   int
	Last   string
	Share  float64
}

// ToolShare is one tool's contribution to the window.
type ToolShare struct {
	Name  string
	Value int64
	Pct   float64
}

// ModelSort orders the Models table.
type ModelSort string

const (
	ModelSortTokens ModelSort = "tokens"
	ModelSortCost   ModelSort = "cost"
	ModelSortName   ModelSort = "name"
)

// ModelSortOrder is the cycle for the s key in the Models view.
var ModelSortOrder = []ModelSort{ModelSortTokens, ModelSortCost, ModelSortName}

func (s ModelSort) next() ModelSort {
	for i, v := range ModelSortOrder {
		if v == s {
			return ModelSortOrder[(i+1)%len(ModelSortOrder)]
		}
	}
	return ModelSortTokens
}

// BuildModels aggregates one row per canonical model id over the window. It
// reuses thermal.AggregateModels, the function behind `thermal models`, so row
// cost and the unpriced-model list match the static command exactly.
func (a Adapter) BuildModels(r Range, m Metric, s ModelSort) ModelsView {
	start, end := a.window(r)
	out := ModelsView{Range: r, Metric: m, Sort: s, Start: start, End: end}

	batches := make([]thermal.ToolDays, 0, len(a.Tools))
	for _, t := range a.Tools {
		batches = append(batches, thermal.ToolDays{Tool: t.Name, Days: t.Days})
	}
	rep := thermal.AggregateModels(batches, thermal.ModelOptions{Since: start, Until: end}, a.Pricer)

	var covered int64
	for _, batch := range batches {
		for _, day := range batch.Days {
			if inWindow(day.Day, start, end) {
				covered += day.Tokens
			}
		}
	}

	for _, row := range rep.Rows {
		out.Rows = append(out.Rows, ModelRowView{
			Name:   row.Model,
			Tools:  row.Tools,
			Tokens: row.Tokens,
			Input:  row.Input,
			Output: row.Output,
			Cache:  row.CacheRead + row.CacheWrite,
			Cost:   row.Cost,
			Est:    row.Cost,
			Days:   row.Days,
			Last:   row.LastDay,
		})
	}
	// Token and cost order come from the static aggregator. Name order is the
	// view's own, because the static command has no name sort to reuse.
	if s == ModelSortName {
		sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Name < out.Rows[j].Name })
	}

	out.Tokens = rep.Totals.Tokens
	out.Cost = rep.Totals.Cost
	out.EstCost = rep.Totals.Cost // model cost is always an estimate
	if out.Tokens > 0 {
		for i := range out.Rows {
			out.Rows[i].Share = float64(out.Rows[i].Tokens) / float64(out.Tokens) * 100
		}
	}
	for _, row := range out.Rows {
		out.Shares = append(out.Shares, Share{Tool: row.Name, Value: row.Tokens, Pct: row.Share})
	}
	// The note is only worth printing when something is actually uncovered. A
	// rounding difference between the two sums is not a coverage gap.
	note := ""
	if gap := covered - out.Tokens; gap > covered/100 && gap > 0 {
		note = fmt.Sprintf("models cover %s of %s tokens in this window; the rest comes from tools that record no model",
			thermal.CompactNumber(out.Tokens), thermal.CompactNumber(covered))
	}
	if len(rep.Totals.MissingPricing) > 0 {
		if note != "" {
			note += " · "
		}
		note += "no pricing for " + joinSeriesNames(rep.Totals.MissingPricing, 3)
	}
	out.CovNote = note
	return out
}

func modelLess(a, b ModelRowView, key ModelSort) bool {
	switch key {
	case ModelSortCost:
		if a.Cost != b.Cost {
			return a.Cost > b.Cost
		}
	case ModelSortName:
		return a.Name < b.Name
	default:
		if a.Tokens != b.Tokens {
			return a.Tokens > b.Tokens
		}
	}
	return a.Name < b.Name
}

// toolSharesFor sums tokens per tool over the window, for the Mix panel.
func (a Adapter) toolSharesFor(r Range) []ToolShare {
	start, end := a.window(r)
	var out []ToolShare
	var total int64
	for _, t := range a.Tools {
		var sum int64
		for _, d := range t.Days {
			if inWindow(d.Day, start, end) {
				sum += d.Tokens
			}
		}
		if sum > 0 {
			out = append(out, ToolShare{Name: t.Name, Value: sum})
			total += sum
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Value != out[j].Value {
			return out[i].Value > out[j].Value
		}
		return out[i].Name < out[j].Name
	})
	for i := range out {
		if total > 0 {
			out[i].Pct = float64(out[i].Value) / float64(total) * 100
		}
	}
	return out
}

// StartOfWeek returns local midnight on the Sunday on or before t. Charts need
// a fixed week origin so the same data always buckets the same way.
func StartOfWeek(t time.Time) time.Time {
	t = t.AddDate(0, 0, -int(t.Weekday()))
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BumpWeek returns the week bucket label for a day, used by the stacked chart
// legend when the grain is weekly.
func BumpWeek(day string) string {
	t, ok := thermal.ParseDay(day)
	if !ok {
		return day
	}
	return StartOfWeek(t).Format("01-02")
}

// joinSeriesNames renders series names for a legend line.
func joinSeriesNames(names []string, limit int) string {
	if len(names) <= limit {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:limit], ", ") + fmt.Sprintf(" +%d", len(names)-limit)
}
