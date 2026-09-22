// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

// ProjectRow is one project's line in the Projects view. It is built from the
// same aggregation the static `thermal projects` command uses, so the two
// cannot disagree about a window.
type ProjectRow struct {
	Project    string
	Tools      []string
	ToolTokens map[string]int64
	Tokens     int64
	Input      int64
	Output     int64
	Cache      int64
	Cost       float64
	EstCost    float64
	StoredCost float64
	Turns      int
	Days       int
	FirstDay   string
	LastDay    string
	Models     map[string]thermal.ModelTokens
}

// ProjectsData is everything the Projects view prints for one window.
type ProjectsData struct {
	Range    Range
	Sort     ProjectSort
	Start    string
	End      string
	Rows     []ProjectRow
	Tokens   int64
	Cost     float64
	EstCost  float64
	Turns    int
	Days     int
	LastDay  string
	ToolName string // filter: empty means every tool
	Tools    []string
}

// ProjectSort is the Projects view ordering.
type ProjectSort string

const (
	ProjectSortTokens ProjectSort = "tokens"
	ProjectSortCost   ProjectSort = "cost"
	ProjectSortDays   ProjectSort = "days"
	ProjectSortRecent ProjectSort = "recent"
)

// ProjectSortOrder is the cycle order for the s key in this view.
var ProjectSortOrder = []ProjectSort{ProjectSortTokens, ProjectSortCost, ProjectSortDays, ProjectSortRecent}

func (s ProjectSort) next() ProjectSort {
	for i, v := range ProjectSortOrder {
		if v == s {
			return ProjectSortOrder[(i+1)%len(ProjectSortOrder)]
		}
	}
	return ProjectSortTokens
}

// projectDays flattens the adapter's project rows, honouring the tool filter.
func (a Adapter) projectDays(toolFilter string) []thermal.ProjectDay {
	var out []thermal.ProjectDay
	for _, t := range a.Tools {
		if toolFilter != "" && t.Name != toolFilter {
			continue
		}
		for _, p := range t.Projects {
			if p.Tool == "" {
				p.Tool = t.Name
			}
			out = append(out, p)
		}
	}
	return out
}

// BuildProjects aggregates the window through thermal.AggregateProjects, the
// same function behind `thermal projects`, and then adds the fields the table
// needs. Passing explicit Since and Until is what makes the parity exact: the
// dashboard's window ends at the newest recorded day, not at the wall clock.
func (a Adapter) BuildProjects(r Range, s ProjectSort, toolFilter string) ProjectsData {
	start, end := a.window(r)
	out := ProjectsData{Range: r, Sort: s, Start: start, End: end, ToolName: toolFilter}

	days := a.projectDays(toolFilter)
	if len(days) == 0 {
		return out
	}
	rep := thermal.AggregateProjects(days, thermal.ProjectOptions{
		Since: start,
		Until: end,
	}, a.Pricer)

	out.Tokens = rep.Totals.Tokens
	out.Cost = rep.Totals.Cost
	out.EstCost = rep.Totals.EstimatedCost
	out.Turns = rep.Totals.Turns
	out.Days = rep.Totals.ActiveDays
	out.LastDay = rep.Totals.LastDay

	for _, row := range rep.Rows {
		out.Rows = append(out.Rows, ProjectRow{
			Project:    row.Project,
			Tools:      row.Tools,
			ToolTokens: row.ToolTokens,
			Tokens:     row.Tokens,
			Input:      row.Input,
			Output:     row.Output,
			Cache:      row.Cache,
			Cost:       row.Cost,
			EstCost:    row.EstimatedCost,
			StoredCost: row.StoredCost,
			Turns:      row.Turns,
			Days:       row.ActiveDays,
			FirstDay:   row.FirstDay,
			LastDay:    row.LastDay,
			Models:     row.Models,
		})
	}

	sort.SliceStable(out.Rows, func(i, j int) bool {
		return projectLess(out.Rows[i], out.Rows[j], s)
	})
	return out
}

func projectLess(a, b ProjectRow, key ProjectSort) bool {
	switch key {
	case ProjectSortCost:
		if a.Cost != b.Cost {
			return a.Cost > b.Cost
		}
	case ProjectSortDays:
		if a.Days != b.Days {
			return a.Days > b.Days
		}
	case ProjectSortRecent:
		if a.LastDay != b.LastDay {
			return a.LastDay > b.LastDay
		}
	default:
		if a.Tokens != b.Tokens {
			return a.Tokens > b.Tokens
		}
	}
	return a.Project < b.Project
}

// ProjectDetail is the drill-down for one project.
type ProjectDetail struct {
	Project   string // the aggregation key
	Label     string // the human label, from thermal.ProjectDisplayNames
	Start     string
	End       string
	Tokens    int64
	Cost      float64
	EstCost   float64
	Turns     int
	Days      int
	FirstDay  string
	LastDay   string
	Tools     []ProjectShare
	Models    []ModelShare
	Weekly    []int64 // last twelve weeks of tokens, oldest first
	Weeks     []string
	Heatmap   []HeatDay
	ModelNote string
}

// ProjectShare is one tool's contribution inside a project.
type ProjectShare struct {
	Name  string
	Value int64
	Pct   float64
}

// ModelShare is one model's contribution inside a project.
type ModelShare struct {
	Name  string
	Value int64
	Pct   float64
}

// HeatDay is one cell of the project heatmap: a day and its level from 0 to 4.
type HeatDay struct {
	Day   string
	Level int
}

// BuildProjectDetail aggregates one project for the window. Model attribution
// is partial across tools, so the note states the coverage rather than showing
// a share that looks complete.
func (a Adapter) BuildProjectDetail(project string, r Range) ProjectDetail {
	start, end := a.window(r)
	d := ProjectDetail{Project: project, Start: start, End: end}
	if names := thermal.ProjectDisplayNames([]string{project}); names[project] != "" {
		d.Label = names[project]
	} else {
		d.Label = project
	}

	var days []thermal.ProjectDay
	for _, p := range a.projectDays("") {
		if p.Project == project {
			days = append(days, p)
		}
	}
	if len(days) == 0 {
		return d
	}

	rep := thermal.AggregateProjects(days, thermal.ProjectOptions{Since: start, Until: end}, a.Pricer)
	if len(rep.Rows) == 0 {
		return d
	}
	row := rep.Rows[0]
	d.Tokens = row.Tokens
	d.Cost = row.Cost
	d.EstCost = row.EstimatedCost
	d.Turns = row.Turns
	d.Days = row.ActiveDays
	d.FirstDay = row.FirstDay
	d.LastDay = row.LastDay

	for _, name := range thermal.TopModels(row.Models) {
		m := row.Models[name]
		d.Models = append(d.Models, ModelShare{Name: name, Value: m.Total()})
	}
	var modelTotal int64
	for _, m := range d.Models {
		modelTotal += m.Value
	}
	for i := range d.Models {
		if modelTotal > 0 {
			d.Models[i].Pct = float64(d.Models[i].Value) / float64(modelTotal) * 100
		}
	}
	if row.Tokens > 0 && modelTotal < row.Tokens {
		d.ModelNote = fmt.Sprintf("models cover %s of %s recorded in this project",
			thermal.CompactNumber(modelTotal), thermal.CompactNumber(row.Tokens))
	}

	var toolTotal int64
	toolNames := make([]string, 0, len(row.ToolTokens))
	for name := range row.ToolTokens {
		toolNames = append(toolNames, name)
	}
	sort.Slice(toolNames, func(i, j int) bool {
		if row.ToolTokens[toolNames[i]] != row.ToolTokens[toolNames[j]] {
			return row.ToolTokens[toolNames[i]] > row.ToolTokens[toolNames[j]]
		}
		return toolNames[i] < toolNames[j]
	})
	for _, name := range toolNames {
		v := row.ToolTokens[name]
		toolTotal += v
		d.Tools = append(d.Tools, ProjectShare{Name: name, Value: v})
	}
	for i := range d.Tools {
		if toolTotal > 0 {
			d.Tools[i].Pct = float64(d.Tools[i].Value) / float64(toolTotal) * 100
		}
	}

	d.Weekly, d.Weeks = weeklySeries(days, start, end)
	d.Heatmap = heatmap(days, start, end)
	return d
}

// weeklySeries buckets the project's tokens into twelve weeks ending at the
// window end, oldest first, and records the label for each bucket.
func weeklySeries(days []thermal.ProjectDay, start, end string) ([]int64, []string) {
	endT, ok := thermal.ParseDay(end)
	if !ok {
		return nil, nil
	}
	values := make([]int64, 12)
	labels := make([]string, 12)
	for i := 0; i < 12; i++ {
		weekEnd := endT.AddDate(0, 0, -7*(11-i))
		labels[i] = weekEnd.Format("01-02")
	}
	for _, p := range days {
		if !inWindow(p.Day, start, end) || p.Tokens == 0 {
			continue
		}
		dayT, ok := thermal.ParseDay(p.Day)
		if !ok {
			continue
		}
		weeksBack := int(endT.Sub(dayT).Hours() / 24 / 7)
		idx := len(values) - 1 - weeksBack
		if idx >= 0 && idx < len(values) {
			values[idx] += p.Tokens
		}
	}
	return values, labels
}

// heatmap builds daily levels for the window, using the same threshold rule as
// the terminal heatmap: levels come from the spread of the activity, not from
// fixed cutoffs. The span follows the selected range, so a 30 day window shows
// five weeks instead of a year of empty cells.
func heatmap(days []thermal.ProjectDay, start, end string) []HeatDay {
	startT, ok := thermal.ParseDay(start)
	if !ok {
		return nil
	}
	endT, ok := thermal.ParseDay(end)
	if !ok {
		return nil
	}
	activity := make(map[string]thermal.DayActivity)
	for _, p := range days {
		if !inWindow(p.Day, start, end) {
			continue
		}
		a := activity[p.Day]
		a.Tokens += p.Tokens
		a.Turns += p.Turns
		activity[p.Day] = a
	}
	thresholds := thermal.ActivityThresholds(activity)

	var out []HeatDay
	for d := startT; !d.After(endT); d = d.AddDate(0, 0, 1) {
		day := thermal.LocalDay(d)
		out = append(out, HeatDay{Day: day, Level: thermal.ActivityLevel(activity[day], thresholds)})
	}
	return out
}

// toolFilterOptions lists the tools that contributed inside the window, for the
// filter cycle. Only tools with project rows can appear.
func (a Adapter) toolFilterOptions(r Range) []string {
	start, end := a.window(r)
	seen := make(map[string]bool)
	var out []string
	for _, t := range a.Tools {
		for _, p := range t.Projects {
			if inWindow(p.Day, start, end) {
				seen[t.Name] = true
				break
			}
		}
	}
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// ModelNoteForTools explains the model coverage of a window in one line.
func ModelNoteForTools(rows []ProjectRow, tokens int64) string {
	var covered int64
	for _, row := range rows {
		for _, m := range row.Models {
			covered += m.Total()
		}
	}
	if covered >= tokens || tokens == 0 {
		return ""
	}
	return fmt.Sprintf("model rows cover %s of %s tokens; the rest comes from tools that record no model",
		thermal.CompactNumber(covered), thermal.CompactNumber(tokens))
}

func joinTools(names []string, limit int) string {
	if len(names) == 0 {
		return "—"
	}
	if len(names) <= limit {
		return strings.Join(names, ",")
	}
	return strings.Join(names[:limit], ",") + fmt.Sprintf(" +%d", len(names)-limit)
}
