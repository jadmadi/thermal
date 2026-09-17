package thermal

import (
	"sort"
	"strings"
	"time"
)

// MixOptions controls the tool and model mix aggregation. By is "tool" or
// "model" and Metric is "tokens" or "cost". Grain buckets the timeline and
// defaults to weeks.
type MixOptions struct {
	Since       string
	Until       string
	Last        int
	Grain       Grain
	By          string
	Metric      string
	StartOfWeek time.Weekday
	Now         time.Time
}

// MixSeries is one tool or model across the whole window. Buckets aligns with
// MixReport.Buckets order.
type MixSeries struct {
	Name    string    `json:"name"`
	Total   float64   `json:"total"`
	Share   float64   `json:"share"`
	Buckets []float64 `json:"buckets"`
}

// MixBucket is one period with one value per series, in series order.
type MixBucket struct {
	Period string    `json:"period"`
	Values []float64 `json:"values"`
	Total  float64   `json:"total"`
}

// MixReport is the payload behind thermal mix. Switching fields come from
// day-level tool dominance and are empty for model mixes.
type MixReport struct {
	Type            string         `json:"type"`
	By              string         `json:"by"`
	Metric          string         `json:"metric"`
	Series          []MixSeries    `json:"series"`
	Buckets         []MixBucket    `json:"data"`
	Dominant        string         `json:"dominant,omitempty"`
	Switches        int            `json:"switches"`
	SwitchesByMonth map[string]int `json:"switchesByMonth,omitempty"`
	Concentration   float64        `json:"concentration"`
	ActiveSeries    int            `json:"activeSeries"`
	Total           float64        `json:"total"`
	Estimated       bool           `json:"estimatedCost,omitempty"`
}

// mixAccumulator folds day-level values into series and period buckets.
type mixAccumulator struct {
	metric      string
	startOfWeek time.Weekday
	grain       Grain

	seriesTotals map[string]float64
	bucketValues map[string]map[string]float64
	bucketOrder  []string

	// Day-level tool dominance, used for switching. Keyed by day.
	dayValues map[string]map[string]float64

	estimated bool
}

func newMixAccumulator(opts MixOptions) *mixAccumulator {
	grain := opts.Grain
	if grain != GrainDay && grain != GrainWeek {
		grain = GrainMonth
	}
	return &mixAccumulator{
		metric:       normalMixMetric(opts.Metric),
		startOfWeek:  opts.StartOfWeek,
		grain:        grain,
		seriesTotals: make(map[string]float64),
		bucketValues: make(map[string]map[string]float64),
		dayValues:    make(map[string]map[string]float64),
	}
}

func normalMixMetric(metric string) string {
	if strings.EqualFold(metric, "cost") {
		return "cost"
	}
	return "tokens"
}

func (a *mixAccumulator) add(day, series string, value float64) {
	if series == "" || value == 0 {
		return
	}
	t, ok := ParseDay(day)
	if !ok {
		return
	}
	period := periodLabel(periodStart(t, a.grain, a.startOfWeek), a.grain)

	a.seriesTotals[series] += value
	if a.bucketValues[period] == nil {
		a.bucketValues[period] = make(map[string]float64)
		a.bucketOrder = append(a.bucketOrder, period)
	}
	a.bucketValues[period][series] += value

	if a.dayValues[day] == nil {
		a.dayValues[day] = make(map[string]float64)
	}
	a.dayValues[day][series] += value
}

// finish assembles the report, ranking series by total and computing shares,
// concentration, and (for tools) switching statistics.
func (a *mixAccumulator) finish(opts MixOptions) MixReport {
	series := make([]string, 0, len(a.seriesTotals))
	var grand float64
	for name, total := range a.seriesTotals {
		series = append(series, name)
		grand += total
	}
	sort.Slice(series, func(i, j int) bool {
		if a.seriesTotals[series[i]] != a.seriesTotals[series[j]] {
			return a.seriesTotals[series[i]] > a.seriesTotals[series[j]]
		}
		return series[i] < series[j]
	})

	sort.Strings(a.bucketOrder)

	rep := MixReport{
		Type:         "mix",
		By:           normalMixBy(opts.By),
		Metric:       a.metric,
		Total:        grand,
		Estimated:    a.estimated,
		Buckets:      make([]MixBucket, 0, len(a.bucketOrder)),
		Series:       make([]MixSeries, 0, len(series)),
		ActiveSeries: len(series),
	}

	index := make(map[string]int, len(series))
	for i, name := range series {
		index[name] = i
	}

	for _, period := range a.bucketOrder {
		bucket := MixBucket{Period: period, Values: make([]float64, len(series))}
		for name, value := range a.bucketValues[period] {
			i, ok := index[name]
			if !ok {
				continue
			}
			bucket.Values[i] = value
			bucket.Total += value
		}
		rep.Buckets = append(rep.Buckets, bucket)
	}

	for _, name := range series {
		ms := MixSeries{Name: name, Total: a.seriesTotals[name], Buckets: make([]float64, len(rep.Buckets))}
		if grand > 0 {
			ms.Share = ms.Total / grand
		}
		for i, bucket := range rep.Buckets {
			ms.Buckets[i] = bucket.Values[index[name]]
		}
		rep.Series = append(rep.Series, ms)
		if grand > 0 {
			rep.Concentration += ms.Share * ms.Share
		}
	}

	if rep.By == "tool" {
		rep.Dominant, rep.Switches, rep.SwitchesByMonth = a.switching()
	}
	return rep
}

// switching walks active days in order and counts where the dominant tool
// changes, both overall and per calendar month.
func (a *mixAccumulator) switching() (string, int, map[string]int) {
	days := make([]string, 0, len(a.dayValues))
	for day := range a.dayValues {
		days = append(days, day)
	}
	sort.Strings(days)

	byMonth := make(map[string]int)
	var dominant string
	switches := 0
	totals := make(map[string]float64)

	for _, day := range days {
		dayDominant := ""
		var best float64
		for series, value := range a.dayValues[day] {
			totals[series] += value
			if value > best || (value == best && series < dayDominant) {
				best = value
				dayDominant = series
			}
		}
		if dayDominant == "" {
			continue
		}
		if dominant != "" && dayDominant != dominant {
			switches++
			byMonth[day[:7]]++
		}
		dominant = dayDominant
	}

	overall := ""
	var bestTotal float64
	for series, total := range totals {
		if total > bestTotal || (total == bestTotal && series < overall) {
			bestTotal = total
			overall = series
		}
	}
	return overall, switches, byMonth
}

func normalMixBy(by string) string {
	if strings.EqualFold(by, "model") {
		return "model"
	}
	return "tool"
}

// AggregateToolMix buckets per-day tool usage into periods. Cost uses the
// recorded value when a day has one, otherwise the pricer, and reports whether
// any estimate entered the total.
func AggregateToolMix(batches []ToolDays, opts MixOptions, pricer Pricer) MixReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	a := newMixAccumulator(opts)

	for _, batch := range batches {
		for _, day := range batch.Days {
			t, ok := ParseDay(day.Day)
			if !ok || !inWindow(t, since, until, lastStart) {
				continue
			}
			if a.metric == "cost" {
				cost, estimated := dayCost(day, pricer)
				if cost > 0 {
					a.add(day.Day, batch.Tool, cost)
				}
				a.estimated = a.estimated || estimated
				continue
			}
			a.add(day.Day, batch.Tool, float64(day.Tokens))
		}
	}
	return a.finish(opts)
}

// AggregateModelMix buckets per-day model usage into periods. Model cost is
// always estimated, because recorded cost is not attributable to one model.
func AggregateModelMix(days []DailyRow, opts MixOptions, pricer Pricer) MixReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	a := newMixAccumulator(opts)

	for _, day := range days {
		t, ok := ParseDay(day.Day)
		if !ok || !inWindow(t, since, until, lastStart) {
			continue
		}
		for model, counts := range day.Models {
			if counts.Total() == 0 {
				continue
			}
			if a.metric == "cost" {
				if pricer == nil {
					continue
				}
				cost, _ := pricer.PriceDay(DailyRow{
					Tokens: counts.Total(),
					Input:  counts.Input,
					Output: counts.Output,
					Cache:  counts.Cache(),
					Models: map[string]ModelTokens{model: counts},
				})
				if cost > 0 {
					a.add(day.Day, model, cost)
				}
				a.estimated = true
				continue
			}
			a.add(day.Day, model, float64(counts.Total()))
		}
	}
	return a.finish(opts)
}

// dayCost returns a day's cost, preferring the recorded value and falling back
// to the pricer. The boolean reports whether an estimate was used.
func dayCost(day DailyRow, pricer Pricer) (float64, bool) {
	if day.Cost > 0 {
		return day.Cost, false
	}
	if pricer == nil || day.Tokens == 0 || len(day.Models) == 0 {
		return 0, false
	}
	cost, _ := pricer.PriceDay(day)
	return cost, cost > 0
}
