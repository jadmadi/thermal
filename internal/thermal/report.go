package thermal

import (
	"sort"
	"strings"
	"time"
)

// ParseDay accepts YYYY-MM-DD or YYYYMMDD and returns local midnight. It
// returns false for empty or unparseable input.
func ParseDay(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02", "20060102"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// ParseWeekday maps a full weekday name to time.Weekday. It accepts any case.
func ParseWeekday(s string) (time.Weekday, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sunday":
		return time.Sunday, true
	case "monday":
		return time.Monday, true
	case "tuesday":
		return time.Tuesday, true
	case "wednesday":
		return time.Wednesday, true
	case "thursday":
		return time.Thursday, true
	case "friday":
		return time.Friday, true
	case "saturday":
		return time.Saturday, true
	}
	return time.Sunday, false
}

func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// WeekStart returns local midnight of the week start on or before t.
func WeekStart(t time.Time, start time.Weekday) time.Time {
	t = midnight(t)
	delta := (int(t.Weekday()) - int(start) + 7) % 7
	return t.AddDate(0, 0, -delta)
}

func periodStart(t time.Time, grain Grain, start time.Weekday) time.Time {
	switch grain {
	case GrainWeek:
		return WeekStart(t, start)
	case GrainMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return midnight(t)
	}
}

func periodLabel(t time.Time, grain Grain) string {
	if grain == GrainMonth {
		return t.Format("2006-01")
	}
	return t.Format("2006-01-02")
}

// TopModels returns model names sorted by total tokens descending. Ties fall
// back to name order so output is stable.
func TopModels(models map[string]ModelTokens) []string {
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		ti, tj := models[names[i]].Total(), models[names[j]].Total()
		if ti != tj {
			return ti > tj
		}
		return names[i] < names[j]
	})
	return names
}

// hasTokenData reports whether a day carries tokens or cost. A positive Tokens
// total counts on its own: a source that reports a session total without a
// type breakdown, such as codewhale, is a token source, and dropping it would
// hide real spend. Activity-only days, where Tokens holds a message or step
// count and no cost or model is recorded, return false so period reports never
// mix steps into token columns and totals. Streaks and the leaderboard keep
// that activity separately.
func hasTokenData(day DailyRow) bool {
	if day.Tokens > 0 {
		return true
	}
	if day.Input != 0 || day.Output != 0 || day.Reasoning != 0 || day.Cache != 0 {
		return true
	}
	if day.Cost != 0 {
		return true
	}
	for _, counts := range day.Models {
		if counts.Total() > 0 {
			return true
		}
	}
	return false
}

// Aggregate folds day rows into a report at the requested grain. Days outside
// the since, until, or last window are dropped, as are days with no token or
// cost telemetry (activity-only and empty session days). Rows come back newest
// first unless Order is "asc". Totals accumulate from the kept days directly,
// so every grain reports bit-identical totals for the same window. A nil
// pricer leaves EstimatedCost at zero.
func Aggregate(days []DailyRow, grain Grain, opts AggregateOptions, pricer Pricer) Report {
	if grain != GrainWeek && grain != GrainMonth {
		grain = GrainDay
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	var since, until time.Time
	if t, ok := ParseDay(opts.Since); ok {
		since = t
	}
	if t, ok := ParseDay(opts.Until); ok {
		until = t
	}

	// --last counts whole periods back from the current one, so it includes
	// today even when the current period has no activity yet.
	var lastStart time.Time
	if opts.Last > 0 {
		start := periodStart(now, grain, opts.StartOfWeek)
		switch grain {
		case GrainMonth:
			lastStart = start.AddDate(0, -(opts.Last - 1), 0)
		case GrainWeek:
			lastStart = start.AddDate(0, 0, -7*(opts.Last-1))
		default:
			lastStart = start.AddDate(0, 0, -(opts.Last - 1))
		}
	}

	byKey := make(map[string]*PeriodRow)
	activeSets := make(map[string]map[string]bool)

	var total PeriodRow
	totalModels := make(map[string]ModelTokens)
	totalActive := make(map[string]bool)

	for _, day := range days {
		t, ok := ParseDay(day.Day)
		if !ok {
			continue
		}
		if !since.IsZero() && t.Before(since) {
			continue
		}
		if !until.IsZero() && t.After(until) {
			continue
		}
		if !lastStart.IsZero() && t.Before(lastStart) {
			continue
		}
		if !hasTokenData(day) {
			continue
		}

		key := periodLabel(periodStart(t, grain, opts.StartOfWeek), grain)
		row := byKey[key]
		if row == nil {
			row = &PeriodRow{Period: key, Models: make(map[string]ModelTokens)}
			byKey[key] = row
		}

		row.Input += day.Input
		row.Output += day.Output
		row.Reasoning += day.Reasoning
		row.Cache += day.Cache
		row.Tokens += day.Tokens
		row.Turns += day.Turns
		row.StoredCost += day.Cost

		total.Input += day.Input
		total.Output += day.Output
		total.Reasoning += day.Reasoning
		total.Cache += day.Cache
		total.Tokens += day.Tokens
		total.Turns += day.Turns
		total.StoredCost += day.Cost

		if day.Turns > 0 {
			set := activeSets[key]
			if set == nil {
				set = make(map[string]bool)
				activeSets[key] = set
			}
			set[day.Day] = true
			totalActive[day.Day] = true
		}

		// A day whose source names no model can never be priced. Count those
		// tokens so the footer can state them instead of staying silent.
		if day.Cost == 0 && len(day.Models) == 0 && day.Tokens > 0 {
			row.UnattributedTokens += day.Tokens
			total.UnattributedTokens += day.Tokens
		}

		// Estimate only days with no stored cost. A day with recorded cost
		// gets no pricing pass, so the two sources never mix inside a day.
		if pricer != nil && day.Cost == 0 && day.Tokens > 0 {
			cost, missing := pricer.PriceDay(day)
			row.EstimatedCost += cost
			total.EstimatedCost += cost
			for _, m := range missing {
				if !containsString(row.MissingPricing, m) {
					row.MissingPricing = append(row.MissingPricing, m)
				}
				if !containsString(total.MissingPricing, m) {
					total.MissingPricing = append(total.MissingPricing, m)
				}
			}
		}

		// Zero-count model entries are loader artifacts (a named session with
		// no counted tokens). They carry no signal, so they never become rows
		// or inflate the "+N" model suffix.
		for model, counts := range day.Models {
			if counts.Total() == 0 {
				continue
			}
			row.Models[model] = row.Models[model].Add(counts)
			totalModels[model] = totalModels[model].Add(counts)
		}
	}

	rows := make([]PeriodRow, 0, len(byKey))
	for key, row := range byKey {
		row.ActiveDays = len(activeSets[key])
		row.Cost = row.StoredCost + row.EstimatedCost
		if len(row.Models) == 0 {
			row.Models = nil
		}
		sort.Strings(row.MissingPricing)
		if len(row.MissingPricing) == 0 {
			row.MissingPricing = nil
		}
		rows = append(rows, *row)
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Period < rows[j].Period })
	if !strings.EqualFold(opts.Order, "asc") {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}

	total.ActiveDays = len(totalActive)
	total.Cost = total.StoredCost + total.EstimatedCost
	sort.Strings(total.MissingPricing)
	if len(total.MissingPricing) == 0 {
		total.MissingPricing = nil
	}
	if len(totalModels) > 0 {
		total.Models = totalModels
	}

	return Report{Type: string(grain), Rows: rows, Totals: total}
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
