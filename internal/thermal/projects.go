package thermal

import (
	"sort"
	"strings"
	"time"
)

// AggregateProjects folds project-day usage into one row per project, summing
// across tools and days inside the requested window. Rows come back largest
// first by token total unless Order is "asc". A nil pricer leaves
// EstimatedCost at zero, and stored cost always wins for a given day.
func AggregateProjects(days []ProjectDay, opts ProjectOptions, pricer Pricer) ProjectReport {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	var since, until, lastStart time.Time
	if t, ok := ParseDay(opts.Since); ok {
		since = t
	}
	if t, ok := ParseDay(opts.Until); ok {
		until = t
	}
	if opts.Last > 0 {
		lastStart = midnight(now).AddDate(0, 0, -(opts.Last - 1))
	}

	byKey := make(map[string]*ProjectRow)
	activeSets := make(map[string]map[string]bool)
	toolSets := make(map[string]map[string]bool)

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
		if day.Project == "" {
			continue
		}

		row := byKey[day.Project]
		if row == nil {
			row = &ProjectRow{Project: day.Project, Models: make(map[string]ModelTokens)}
			byKey[day.Project] = row
		}

		row.Input += day.Input
		row.Output += day.Output
		row.Reasoning += day.Reasoning
		row.Cache += day.CacheRead + day.CacheWrite
		row.Tokens += day.Tokens
		row.Turns += day.Turns
		row.StoredCost += day.Cost
		if row.FirstDay == "" || day.Day < row.FirstDay {
			row.FirstDay = day.Day
		}
		if day.Day > row.LastDay {
			row.LastDay = day.Day
		}
		if day.Turns > 0 {
			if activeSets[day.Project] == nil {
				activeSets[day.Project] = make(map[string]bool)
			}
			activeSets[day.Project][day.Day] = true
		}
		if day.Tool != "" {
			if toolSets[day.Project] == nil {
				toolSets[day.Project] = make(map[string]bool)
			}
			toolSets[day.Project][day.Tool] = true
		}

		if pricer != nil && day.Cost == 0 && day.Tokens > 0 {
			cost, missing := pricer.PriceDay(DailyRow{
				Day:       day.Day,
				Tokens:    day.Tokens,
				Input:     day.Input,
				Output:    day.Output,
				Reasoning: day.Reasoning,
				Cache:     day.CacheRead + day.CacheWrite,
				Models:    day.Models,
			})
			row.EstimatedCost += cost
			for _, m := range missing {
				if !containsString(row.MissingPricing, m) {
					row.MissingPricing = append(row.MissingPricing, m)
				}
			}
		}

		for model, counts := range day.Models {
			row.Models[model] = row.Models[model].Add(counts)
		}
	}

	rows := make([]ProjectRow, 0, len(byKey))
	for project, row := range byKey {
		row.ActiveDays = len(activeSets[project])
		for tool := range toolSets[project] {
			row.Tools = append(row.Tools, tool)
		}
		sort.Strings(row.Tools)
		sort.Strings(row.MissingPricing)
		row.Cost = row.StoredCost + row.EstimatedCost
		if len(row.MissingPricing) == 0 {
			row.MissingPricing = nil
		}
		if len(row.Models) == 0 {
			row.Models = nil
		}
		rows = append(rows, *row)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Tokens != rows[j].Tokens {
			return rows[i].Tokens > rows[j].Tokens
		}
		return rows[i].Project < rows[j].Project
	})
	if strings.EqualFold(opts.Order, "asc") {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}

	return ProjectReport{Type: "projects", Rows: rows, Totals: sumProjects(rows)}
}

func sumProjects(rows []ProjectRow) ProjectRow {
	var total ProjectRow
	for _, row := range rows {
		total.Input += row.Input
		total.Output += row.Output
		total.Reasoning += row.Reasoning
		total.Cache += row.Cache
		total.Tokens += row.Tokens
		total.Turns += row.Turns
		total.ActiveDays += row.ActiveDays
		total.StoredCost += row.StoredCost
		total.EstimatedCost += row.EstimatedCost
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
		for model, counts := range row.Models {
			if total.Models == nil {
				total.Models = make(map[string]ModelTokens)
			}
			total.Models[model] = total.Models[model].Add(counts)
		}
	}
	sort.Strings(total.Tools)
	sort.Strings(total.MissingPricing)
	total.Cost = total.StoredCost + total.EstimatedCost
	if len(total.MissingPricing) == 0 {
		total.MissingPricing = nil
	}
	if len(total.Models) == 0 {
		total.Models = nil
	}
	return total
}
