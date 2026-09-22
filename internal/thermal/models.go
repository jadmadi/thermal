// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"sort"
	"strings"
	"time"
)

// ToolDays is one tool's day rows. The model ranking needs the tool label that
// day rows themselves do not carry.
type ToolDays struct {
	Tool string
	Days []DailyRow
}

// AggregateModels ranks models across tools and time by token volume or
// estimated cost. Model cost is always estimated, because recorded cost is
// attached to a session or a day, never to a single model. A nil pricer leaves
// Cost at zero and skips the missing-price list.
func AggregateModels(batches []ToolDays, opts ModelOptions, pricer Pricer) ModelReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)

	byModel := make(map[string]*ModelRow)
	activeDays := make(map[string]map[string]bool)
	toolSets := make(map[string]map[string]bool)
	var order []string

	for _, batch := range batches {
		for _, day := range batch.Days {
			t, ok := ParseDay(day.Day)
			if !ok || !inWindow(t, since, until, lastStart) {
				continue
			}
			for model, counts := range day.Models {
				if model == "" || counts.Total() == 0 {
					continue
				}
				row := byModel[model]
				if row == nil {
					row = &ModelRow{Model: model}
					byModel[model] = row
					order = append(order, model)
				}
				row.Input += counts.Input
				row.Output += counts.Output
				row.Reasoning += counts.Reasoning
				row.CacheRead += counts.CacheRead
				row.CacheWrite += counts.CacheWrite
				row.Tokens += counts.Total()

				if activeDays[model] == nil {
					activeDays[model] = make(map[string]bool)
				}
				activeDays[model][day.Day] = true
				if batch.Tool != "" {
					if toolSets[model] == nil {
						toolSets[model] = make(map[string]bool)
					}
					toolSets[model][batch.Tool] = true
				}

				if row.FirstDay == "" || day.Day < row.FirstDay {
					row.FirstDay = day.Day
				}
				if day.Day > row.LastDay {
					row.LastDay = day.Day
				}
			}
		}
	}

	rows := make([]ModelRow, 0, len(order))
	for _, model := range order {
		row := byModel[model]
		row.Days = len(activeDays[model])
		for tool := range toolSets[model] {
			row.Tools = append(row.Tools, tool)
		}
		sort.Strings(row.Tools)

		if pricer != nil {
			counts := ModelTokens{
				Input:      row.Input,
				Output:     row.Output,
				Reasoning:  row.Reasoning,
				CacheRead:  row.CacheRead,
				CacheWrite: row.CacheWrite,
			}
			cost, missing := pricer.PriceDay(DailyRow{
				Tokens:    row.Tokens,
				Input:     row.Input,
				Output:    row.Output,
				Reasoning: row.Reasoning,
				Cache:     row.CacheRead + row.CacheWrite,
				Models:    map[string]ModelTokens{model: counts},
			})
			row.Cost = cost
			for _, m := range missing {
				if !containsString(row.MissingPricing, m) {
					row.MissingPricing = append(row.MissingPricing, m)
				}
			}
			sort.Strings(row.MissingPricing)
			if len(row.MissingPricing) == 0 {
				row.MissingPricing = nil
			}
		}
		rows = append(rows, *row)
	}

	sortModelRows(rows, opts.Sort)
	if strings.EqualFold(opts.Order, "asc") {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}

	return ModelReport{Type: "models", Rows: rows, Totals: sumModels(rows)}
}

// sortModelRows ranks models by the requested key, largest first. Ties fall
// back to token total and then to the model name so the order is stable.
func sortModelRows(rows []ModelRow, key string) {
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		var primary int
		switch strings.ToLower(key) {
		case "cost":
			switch {
			case a.Cost > b.Cost:
				primary = 1
			case a.Cost < b.Cost:
				primary = -1
			}
		default: // tokens
			switch {
			case a.Tokens > b.Tokens:
				primary = 1
			case a.Tokens < b.Tokens:
				primary = -1
			}
		}
		if primary != 0 {
			return primary > 0
		}
		if a.Tokens != b.Tokens {
			return a.Tokens > b.Tokens
		}
		return a.Model < b.Model
	})
}

func sumModels(rows []ModelRow) ModelRow {
	var total ModelRow
	for _, row := range rows {
		total.Input += row.Input
		total.Output += row.Output
		total.Reasoning += row.Reasoning
		total.CacheRead += row.CacheRead
		total.CacheWrite += row.CacheWrite
		total.Tokens += row.Tokens
		total.Days += row.Days
		total.Cost += row.Cost
		if total.FirstDay == "" || (row.FirstDay != "" && row.FirstDay < total.FirstDay) {
			total.FirstDay = row.FirstDay
		}
		if row.LastDay > total.LastDay {
			total.LastDay = row.LastDay
		}
		for _, tool := range row.Tools {
			if !containsString(total.Tools, tool) {
				total.Tools = append(total.Tools, tool)
			}
		}
		for _, m := range row.MissingPricing {
			if !containsString(total.MissingPricing, m) {
				total.MissingPricing = append(total.MissingPricing, m)
			}
		}
	}
	sort.Strings(total.Tools)
	sort.Strings(total.MissingPricing)
	if len(total.MissingPricing) == 0 {
		total.MissingPricing = nil
	}
	return total
}

// windowBounds resolves the report window. Last counts calendar days back from
// now, matching the report commands.
func windowBounds(since, until string, last int, now time.Time) (time.Time, time.Time, time.Time) {
	if now.IsZero() {
		now = time.Now()
	}
	var sinceT, untilT, lastStart time.Time
	if t, ok := ParseDay(since); ok {
		sinceT = t
	}
	if t, ok := ParseDay(until); ok {
		untilT = t
	}
	if last > 0 {
		lastStart = midnight(now).AddDate(0, 0, -(last - 1))
	}
	return sinceT, untilT, lastStart
}

func inWindow(t, since, until, lastStart time.Time) bool {
	if !since.IsZero() && t.Before(since) {
		return false
	}
	if !until.IsZero() && t.After(until) {
		return false
	}
	if !lastStart.IsZero() && t.Before(lastStart) {
		return false
	}
	return true
}
