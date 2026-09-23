// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"os"
	"sort"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/thermal"
)

// toolOrder mirrors the leaderboard's stable display order.
var toolOrder = []thermal.Tool{
	thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolDevin,
	thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode,
	thermal.ToolGrok, thermal.ToolMuse, thermal.ToolClaude, thermal.ToolDroid,
	thermal.ToolDsh, thermal.ToolHermes,
}

// Range is the window a view aggregates over.
type Range string

const (
	Range30d Range = "30d"
	Range90d Range = "90d"
	Range1y  Range = "1y"
	RangeAll Range = "all"
)

// RangeOrder is the cycle order for the r key.
var RangeOrder = []Range{Range30d, Range90d, Range1y, RangeAll}

// days returns how far back the window reaches. "all" is a window wide enough
// to include every recorded day rather than a special case in every query.
func (r Range) days() int {
	switch r {
	case Range30d:
		return 30
	case Range90d:
		return 90
	case Range1y:
		return 365
	default:
		return 36500
	}
}

func (r Range) next() Range {
	for i, v := range RangeOrder {
		if v == r {
			return RangeOrder[(i+1)%len(RangeOrder)]
		}
	}
	return Range30d
}

// Metric selects which quantity the bars and the sort compare.
type Metric string

const (
	MetricTokens Metric = "tokens"
	MetricCost   Metric = "cost"
)

func (m Metric) next() Metric {
	if m == MetricTokens {
		return MetricCost
	}
	return MetricTokens
}

// SortKey is the row ordering within a view.
type SortKey string

const (
	SortTokens SortKey = "tokens"
	SortCost   SortKey = "cost"
	SortName   SortKey = "name"
)

func (s SortKey) next() SortKey {
	switch s {
	case SortTokens:
		return SortCost
	case SortCost:
		return SortName
	default:
		return SortTokens
	}
}

// ToolDay is one tool's loaded data. Rows are kept as loaded so every view can
// aggregate them differently without re-reading a database.
type ToolDay struct {
	Name     string
	Days     []thermal.DailyRow
	Projects []thermal.ProjectDay
}

// Adapter holds everything the dashboard reads. Loading happens once; each
// view aggregates from the same rows, which is what keeps a view total equal
// to the matching static command for the same window.
type Adapter struct {
	Tools    []ToolDay
	Pricer   thermal.Pricer
	Earliest string
	Latest   string
	Skipped  []string
}

// LoadTools reads every tool that has data. Tools that fail are recorded in
// Skipped rather than failing the dashboard, matching how the all-tools static
// commands behave.
func LoadTools(pricer thermal.Pricer) Adapter {
	var a Adapter
	a.Pricer = pricer

	infos := loaders.AllTools()
	for _, t := range toolOrder {
		info := infos[t]
		if !hasData(info) {
			continue
		}
		data, err := loaders.LoadToolData(t, info, "")
		if err != nil {
			a.Skipped = append(a.Skipped, info.Name)
			continue
		}
		for i := range data.Projects {
			data.Projects[i].Tool = info.Name
		}
		a.Tools = append(a.Tools, ToolDay{Name: info.Name, Days: data.Daily, Projects: data.Projects})
		for _, d := range data.Daily {
			if a.Earliest == "" || d.Day < a.Earliest {
				a.Earliest = d.Day
			}
			if d.Day > a.Latest {
				a.Latest = d.Day
			}
		}
	}
	sort.SliceStable(a.Tools, func(i, j int) bool { return a.Tools[i].Name < a.Tools[j].Name })
	return a
}

func hasData(info loaders.ToolInfo) bool {
	if info.DBPath != "" {
		if _, err := os.Stat(info.DBPath); err == nil {
			return true
		}
	}
	if info.DataDir != "" {
		if _, err := os.Stat(info.DataDir); err == nil {
			return true
		}
	}
	return false
}

// OverviewRow is one tool's line in the Overview table.
type OverviewRow struct {
	Tool       string
	Tokens     int64
	Input      int64
	Output     int64
	Cache      int64
	Cost       float64
	EstCost    float64
	StoredCost float64
	Turns      int
	ActiveDays int
	FirstDay   string
	LastDay    string
	Current    int
	Longest    int
	Spark      []int64
	SharePct   float64
}

// Overview is everything the Overview view prints for one window.
type Overview struct {
	Range    Range
	Metric   Metric
	Sort     SortKey
	Start    string
	End      string
	Rows     []OverviewRow
	Tokens   int64
	Cost     float64
	EstCost  float64
	Input    int64
	Output   int64
	Cache    int64
	Turns    int
	Tools    int
	Projects int
	Current  int
	Longest  int
	MaxSpark int64
	Shares   []Share
}

// Share is one tool's slice of the window, used by the share strip.
type Share struct {
	Tool  string
	Value int64
	Pct   float64
}

// window resolves the range into inclusive day strings against the newest day
// present in the data, not the wall clock. A machine whose sources are idle
// still shows its last window instead of an empty dashboard.
func (a Adapter) window(r Range) (start, end string) {
	end = a.Latest
	if end == "" {
		return "", ""
	}
	endT, ok := thermal.ParseDay(end)
	if !ok {
		return "", ""
	}
	startT := endT.AddDate(0, 0, -(r.days() - 1))
	return thermal.LocalDay(startT), end
}

func inWindow(day, start, end string) bool {
	if start == "" || end == "" {
		return false
	}
	return day >= start && day <= end
}

// BuildOverview aggregates the window for every tool. It is pure: no clock and
// no I/O, so parity tests call it directly.
func (a Adapter) BuildOverview(r Range, m Metric, s SortKey) Overview {
	start, end := a.window(r)
	ov := Overview{Range: r, Metric: m, Sort: s, Start: start, End: end}

	projects := make(map[string]bool)
	allActive := make(map[string]bool)

	for _, t := range a.Tools {
		row := OverviewRow{Tool: t.Name}
		active := make(map[string]bool)

		for _, d := range t.Days {
			if !inWindow(d.Day, start, end) {
				continue
			}
			row.Input += d.Input
			row.Output += d.Output
			row.Cache += d.Cache
			row.Tokens += d.Tokens
			row.Turns += d.Turns
			row.StoredCost += d.Cost
			if d.Turns > 0 {
				active[d.Day] = true
				allActive[d.Day] = true
			}
			if row.FirstDay == "" || d.Day < row.FirstDay {
				row.FirstDay = d.Day
			}
			if d.Day > row.LastDay {
				row.LastDay = d.Day
			}
			// Estimate only days with no stored cost, exactly as the static
			// reports do, so the two never mix inside one day.
			if a.Pricer != nil && d.Cost == 0 && d.Tokens > 0 {
				cost, _ := a.Pricer.PriceDay(d)
				row.EstCost += cost
			}
		}

		for _, p := range t.Projects {
			if inWindow(p.Day, start, end) {
				projects[p.Project] = true
			}
		}

		if row.Tokens == 0 && row.StoredCost == 0 && len(active) == 0 {
			continue
		}

		row.ActiveDays = len(active)
		row.Cost = row.StoredCost + row.EstCost
		row.Current, row.Longest = thermal.ComputeStreaks(active)
		row.Spark = a.toolSpark(t, start, end)

		ov.Input += row.Input
		ov.Output += row.Output
		ov.Cache += row.Cache
		ov.Tokens += row.Tokens
		ov.Cost += row.Cost
		ov.EstCost += row.EstCost
		ov.Turns += row.Turns
		ov.Rows = append(ov.Rows, row)
	}

	ov.Current, ov.Longest = thermal.ComputeStreaks(allActive)
	ov.Projects = len(projects)
	ov.Tools = len(ov.Rows)
	ov.MaxSpark = maxSpark(ov.Rows)

	sort.SliceStable(ov.Rows, func(i, j int) bool {
		return sortLess(ov.Rows[i], ov.Rows[j], s)
	})
	var total int64
	for _, row := range ov.Rows {
		total += row.Tokens
	}
	if total > 0 {
		for i := range ov.Rows {
			ov.Rows[i].SharePct = float64(ov.Rows[i].Tokens) / float64(total) * 100
		}
	}
	for _, row := range ov.Rows {
		ov.Shares = append(ov.Shares, Share{Tool: row.Tool, Value: row.Tokens, Pct: row.SharePct})
	}
	return ov
}

// toolSpark buckets the tool's last twelve weeks of tokens into twelve cells,
// oldest week first.
func (a Adapter) toolSpark(t ToolDay, start, end string) []int64 {
	endT, ok := thermal.ParseDay(end)
	if !ok {
		return nil
	}
	spark := make([]int64, 12)
	for _, d := range t.Days {
		if !inWindow(d.Day, start, end) || d.Tokens == 0 {
			continue
		}
		dayT, ok := thermal.ParseDay(d.Day)
		if !ok {
			continue
		}
		weeksBack := int(endT.Sub(dayT).Hours() / 24 / 7)
		idx := len(spark) - 1 - weeksBack
		if idx >= 0 && idx < len(spark) {
			spark[idx] += d.Tokens
		}
	}
	return spark
}

func maxSpark(rows []OverviewRow) int64 {
	var maxVal int64
	for _, r := range rows {
		for _, v := range r.Spark {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	return maxVal
}

func sortLess(a, b OverviewRow, key SortKey) bool {
	switch key {
	case SortCost:
		if a.Cost != b.Cost {
			return a.Cost > b.Cost
		}
	case SortName:
		return a.Tool < b.Tool
	default:
		if a.Tokens != b.Tokens {
			return a.Tokens > b.Tokens
		}
	}
	return a.Tool < b.Tool
}

// Money formats a cost the way the static reports do.
func Money(v float64) string {
	if v == 0 {
		return "—"
	}
	if v < 0.01 {
		return fmt.Sprintf("$%.4f", v)
	}
	return fmt.Sprintf("$%.2f", v)
}

// BuildFinOpsGrid compiles the 9-box FinOps grid payload across loaded tools.
func (a Adapter) BuildFinOpsGrid() thermal.FinOpsGridPayload {
	var allDays []thermal.DailyRow
	var allProjects []thermal.ProjectDay
	var results []thermal.ToolResult
	for _, t := range a.Tools {
		allDays = append(allDays, t.Days...)
		allProjects = append(allProjects, t.Projects...)
		results = append(results, thermal.ToolResult{
			Name:  t.Name,
			Daily: t.Days,
		})
	}
	yieldRep := thermal.AggregateYield(results, allProjects, thermal.YieldOptions{})
	return thermal.ComputeFinOpsGrid(allDays, results, allProjects, yieldRep, a.Pricer)
}
