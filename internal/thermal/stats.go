package thermal

import (
	"sort"
	"time"
)

// StatsOptions controls the daily distribution summary. Metric is "tokens" or
// "cost".
type StatsOptions struct {
	Since  string
	Until  string
	Last   int
	Metric string
	Now    time.Time
}

// DayValue pairs a day with a metric value.
type DayValue struct {
	Day   string  `json:"day"`
	Value float64 `json:"value"`
}

// WeekdayStat is the mean value for one weekday across active days.
type WeekdayStat struct {
	Weekday string  `json:"weekday"`
	Mean    float64 `json:"mean"`
	Days    int     `json:"days"`
}

// HistogramBin counts active days within a value range.
type HistogramBin struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Count int     `json:"count"`
}

// StatsReport is the payload behind thermal stats. Percentiles and outliers
// use active days only, because zero days would drag every percentile to zero.
type StatsReport struct {
	Type             string         `json:"type"`
	Metric           string         `json:"metric"`
	Days             int            `json:"activeDays"`
	Total            float64        `json:"total"`
	Mean             float64        `json:"mean"`
	Median           float64        `json:"median"`
	P90              float64        `json:"p90"`
	Max              float64        `json:"max"`
	Weekday          []WeekdayStat  `json:"weekday"`
	TopDays          []DayValue     `json:"topDays"`
	Outliers         []DayValue     `json:"outliers"`
	OutlierThreshold float64        `json:"outlierThreshold"`
	Histogram        []HistogramBin `json:"histogram"`
	Estimated        bool           `json:"estimatedCost,omitempty"`
}

const histogramBins = 10

// AggregateStats summarises the daily distribution of tokens or cost over the
// requested window.
func AggregateStats(days []DailyRow, opts StatsOptions, pricer Pricer) StatsReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	metric := normalMixMetric(opts.Metric)

	var values []DayValue
	estimated := false
	byDay := make(map[string]float64)
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
	// Rows arrive per tool per day, so sum them before describing a day.
	for day, value := range byDay {
		if value <= 0 {
			continue
		}
		values = append(values, DayValue{Day: day, Value: value})
	}

	rep := StatsReport{
		Type:      "stats",
		Metric:    metric,
		Estimated: estimated,
		Days:      len(values),
	}
	if len(values) == 0 {
		return rep
	}

	sort.Slice(values, func(i, j int) bool { return values[i].Day < values[j].Day })

	nums := make([]float64, len(values))
	for i, v := range values {
		nums[i] = v.Value
		rep.Total += v.Value
	}
	rep.Mean = rep.Total / float64(len(nums))

	sorted := append([]float64{}, nums...)
	sort.Float64s(sorted)
	rep.Median = median(sorted)
	rep.P90 = percentile(sorted, 0.9)
	rep.Max = sorted[len(sorted)-1]

	rep.Weekday = weekdayProfile(values)
	rep.TopDays = topDays(values, 10)

	medianValue := rep.Median
	mad := medianAbsoluteDeviation(sorted, medianValue)
	rep.OutlierThreshold = medianValue + 2*mad
	if mad > 0 {
		for _, v := range values {
			if v.Value > rep.OutlierThreshold {
				rep.Outliers = append(rep.Outliers, v)
			}
		}
	}
	sort.Slice(rep.Outliers, func(i, j int) bool {
		if rep.Outliers[i].Value != rep.Outliers[j].Value {
			return rep.Outliers[i].Value > rep.Outliers[j].Value
		}
		return rep.Outliers[i].Day < rep.Outliers[j].Day
	})

	rep.Histogram = histogram(nums, rep.Max)
	return rep
}

// median returns the middle value, averaging the two middles for an even
// count. The slice must be sorted.
func median(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// percentile expects a sorted slice. It uses nearest rank.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p * float64(len(sorted)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func medianAbsoluteDeviation(sorted []float64, med float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	devs := make([]float64, len(sorted))
	for i, v := range sorted {
		devs[i] = v - med
		if devs[i] < 0 {
			devs[i] = -devs[i]
		}
	}
	sort.Float64s(devs)
	return median(devs)
}

func weekdayProfile(values []DayValue) []WeekdayStat {
	type acc struct {
		total float64
		days  int
	}
	slots := make([]acc, 7)
	for _, v := range values {
		t, err := time.ParseInLocation("2006-01-02", v.Day, time.Local)
		if err != nil {
			continue
		}
		i := int(t.Weekday())
		slots[i].total += v.Value
		slots[i].days++
	}
	names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	out := make([]WeekdayStat, 0, 7)
	for i, slot := range slots {
		stat := WeekdayStat{Weekday: names[i], Days: slot.days}
		if slot.days > 0 {
			stat.Mean = slot.total / float64(slot.days)
		}
		out = append(out, stat)
	}
	return out
}

func topDays(values []DayValue, limit int) []DayValue {
	sorted := append([]DayValue{}, values...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Value != sorted[j].Value {
			return sorted[i].Value > sorted[j].Value
		}
		return sorted[i].Day < sorted[j].Day
	})
	if limit > 0 && len(sorted) > limit {
		sorted = sorted[:limit]
	}
	return sorted
}

func histogram(values []float64, maxVal float64) []HistogramBin {
	if maxVal <= 0 || len(values) == 0 {
		return nil
	}
	bins := make([]HistogramBin, histogramBins)
	width := maxVal / float64(histogramBins)
	for i := range bins {
		bins[i] = HistogramBin{From: float64(i) * width, To: float64(i+1) * width}
	}
	for _, v := range values {
		idx := int(v / width)
		if idx >= histogramBins {
			idx = histogramBins - 1
		}
		if idx < 0 {
			idx = 0
		}
		bins[idx].Count++
	}
	return bins
}
