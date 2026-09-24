// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"sort"
	"strings"
	"time"
)

// TelemetryOptions configures telemetry aggregation over loaded tool results.
type TelemetryOptions struct {
	Tool        string    `json:"tool,omitempty"`
	Since       string    `json:"since,omitempty"`
	Until       string    `json:"until,omitempty"`
	Last        int       `json:"last,omitempty"`
	Now         time.Time `json:"-"`
	NoEstimate  bool      `json:"noEstimate,omitempty"`
	Offline     bool      `json:"offline,omitempty"`
	StartOfWeek string    `json:"startOfWeek,omitempty"`
}

// TelemetryData encapsulates the complete local telemetry snapshot for the web UI.
type TelemetryData struct {
	GeneratedAt    string                 `json:"generatedAt"`
	CurrentStreak  int                    `json:"currentStreak"`
	LongestStreak  int                    `json:"longestStreak"`
	ActiveDays     int                    `json:"activeDays"`
	TotalTokens    int64                  `json:"totalTokens"`
	InputTokens    int64                  `json:"inputTokens"`
	OutputTokens   int64                  `json:"outputTokens"`
	Reasoning      int64                  `json:"reasoningTokens"`
	CacheTokens    int64                  `json:"cacheTokens"`
	TotalCost      float64                `json:"totalCost"`
	RecordedCost   float64                `json:"recordedCost,omitempty"`
	EstimatedCost  float64                `json:"estimatedCost,omitempty"`
	UnpricedTokens int64                  `json:"unpricedTokens,omitempty"`
	DailyActivity  map[string]DayActivity `json:"dailyActivity"`
	Results        []ToolResult           `json:"results"`
	Projects       []ProjectRow           `json:"projects"`
	Models         map[string]int64       `json:"models"`
}

// AggregateTelemetry derives a canonical telemetry snapshot from loaded tool results
// and projects, ensuring web totals strictly match AggregateStats and AggregateProjects.
func AggregateTelemetry(results []ToolResult, projects []ProjectDay, opts TelemetryOptions, pricer Pricer) *TelemetryData {
	if opts.NoEstimate {
		pricer = nil
	}

	targetTool := strings.ToLower(strings.TrimSpace(opts.Tool))
	hasToolFilter := targetTool != "" && targetTool != "all" && targetTool != "auto"

	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)

	var filteredResults []ToolResult
	var filteredProjects []ProjectDay

	for _, r := range results {
		if hasToolFilter {
			rTool := strings.ToLower(strings.TrimSpace(string(r.Tool)))
			rName := strings.ToLower(strings.TrimSpace(r.Name))
			if rTool != targetTool && rName != targetTool {
				continue
			}
		}
		filteredResults = append(filteredResults, r)
	}

	for _, p := range projects {
		if hasToolFilter {
			pTool := strings.ToLower(strings.TrimSpace(p.Tool))
			if pTool != targetTool {
				continue
			}
		}
		filteredProjects = append(filteredProjects, p)
	}

	activeDaysSet := make(map[string]bool)
	dailyActivity := make(map[string]DayActivity)
	allModels := make(map[string]int64)
	var allDailyRows []DailyRow

	var updatedResults []ToolResult
	for _, r := range filteredResults {
		toolActiveDays := make(map[string]bool)
		var toolDailyInWindow []DailyRow
		var toolTokens int64
		var toolTurns int
		var toolRecordedCost float64
		var toolEstimatedCost float64

		for _, d := range r.Daily {
			t, ok := ParseDay(d.Day)
			if !ok || !inWindow(t, since, until, lastStart) {
				continue
			}

			toolDailyInWindow = append(toolDailyInWindow, d)
			allDailyRows = append(allDailyRows, d)

			if d.Turns > 0 {
				toolActiveDays[d.Day] = true
				activeDaysSet[d.Day] = true

				curAct := dailyActivity[d.Day]
				if !isActivityOnly(d) {
					curAct.Tokens += d.Tokens
				}
				curAct.Turns += d.Turns
				dailyActivity[d.Day] = curAct
			}

			c, isEst := dayCost(d, pricer)
			if isEst {
				toolEstimatedCost += c
			} else {
				toolRecordedCost += c
			}

			if isActivityOnly(d) {
				toolTurns += d.Turns
			} else {
				toolTokens += d.Tokens
			}

			for m, mt := range d.Models {
				allModels[m] += mt.Total()
			}
		}

		curStreak, longStreak := ComputeStreaks(toolActiveDays)
		toolSummary := r.Summary
		toolSummary.Cost = toolRecordedCost + toolEstimatedCost

		// When window bounds are specified, recalculate summary activity to match window
		if opts.Since != "" || opts.Until != "" || opts.Last > 0 {
			if isActivityTool(r.Tool) {
				toolSummary.LifetimeTokens = int64(toolTurns)
			} else {
				toolSummary.LifetimeTokens = toolTokens
			}
		}

		totalActivity := toolSummary.LifetimeTokens
		if isActivityTool(r.Tool) && totalActivity == 0 && toolTurns > 0 {
			totalActivity = int64(toolTurns)
		}

		updatedResults = append(updatedResults, ToolResult{
			Tool:          r.Tool,
			Name:          r.Name,
			Summary:       toolSummary,
			Daily:         toolDailyInWindow,
			CurrentStreak: curStreak,
			LongestStreak: longStreak,
			ActiveDays:    len(toolActiveDays),
			TotalActivity: totalActivity,
			DataPath:      r.DataPath,
			EstimatedCost: toolEstimatedCost,
		})
	}

	// Sort results by TotalActivity desc
	sort.Slice(updatedResults, func(i, j int) bool {
		return updatedResults[i].TotalActivity > updatedResults[j].TotalActivity
	})

	curStreak, longStreak := ComputeStreaks(activeDaysSet)

	// Canonical token aggregation via AggregateStats
	statsReport := AggregateStats(allDailyRows, StatsOptions{
		Since:  opts.Since,
		Until:  opts.Until,
		Last:   opts.Last,
		Now:    opts.Now,
		Metric: "tokens",
	}, pricer)

	var inTokens, outTokens, reasTokens, cacheTokens int64
	if statsReport.Composition != nil {
		inTokens = statsReport.Composition.UncachedInput
		outTokens = statsReport.Composition.Output
		reasTokens = statsReport.Composition.Reasoning
		cacheTokens = statsReport.Composition.CacheRead + statsReport.Composition.CacheWrite
	}

	var totRecordedCost float64
	var totEstimatedCost float64
	var totUnpricedTokens int64

	for _, d := range allDailyRows {
		c, isEst := dayCost(d, pricer)
		if isEst {
			totEstimatedCost += c
		} else {
			totRecordedCost += c
		}
		if !isActivityOnly(d) && d.Tokens > 0 && c == 0 && d.Cost == 0 {
			totUnpricedTokens += d.Tokens
		}
	}

	totCost := totRecordedCost + totEstimatedCost

	// Canonical project aggregation via AggregateProjects
	projReport := AggregateProjects(filteredProjects, ProjectOptions{
		Since: opts.Since,
		Until: opts.Until,
		Last:  opts.Last,
		Now:   opts.Now,
	}, pricer)

	genTime := opts.Now
	if genTime.IsZero() {
		genTime = time.Now().UTC()
	}

	return &TelemetryData{
		GeneratedAt:    genTime.Format(time.RFC3339),
		CurrentStreak:  curStreak,
		LongestStreak:  longStreak,
		ActiveDays:     len(activeDaysSet),
		TotalTokens:    int64(statsReport.Total),
		InputTokens:    inTokens,
		OutputTokens:   outTokens,
		Reasoning:      reasTokens,
		CacheTokens:    cacheTokens,
		TotalCost:      totCost,
		RecordedCost:   totRecordedCost,
		EstimatedCost:  totEstimatedCost,
		UnpricedTokens: totUnpricedTokens,
		DailyActivity:  dailyActivity,
		Results:        updatedResults,
		Projects:       projReport.Rows,
		Models:         allModels,
	}
}

func isActivityTool(t Tool) bool {
	return t == ToolAgy || t == ToolCommandCode
}
