// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"sort"
	"strings"
	"time"
)

// ParseDiffStats parses a unified diff patch string and calculates the total
// added and deleted lines. It discards source code and file contents at the parser
// boundary, retaining only sanitized quantitative metrics. Zero heap allocations.
func ParseDiffStats(diff string) (added, deleted int64) {
	for len(diff) > 0 {
		var line string
		if idx := strings.IndexByte(diff, '\n'); idx >= 0 {
			line = diff[:idx]
			diff = diff[idx+1:]
		} else {
			line = diff
			diff = ""
		}
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "@@") {
			continue
		}
		if strings.HasPrefix(line, "+") {
			added++
		} else if strings.HasPrefix(line, "-") {
			deleted++
		}
	}
	return added, deleted
}

// CountLines counts the number of lines in a text block, handling trailing
// newlines properly. Zero heap allocations.
func CountLines(content string) int64 {
	if len(content) == 0 {
		return 0
	}
	var count int64 = 1
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			count++
		}
	}
	if content[len(content)-1] == '\n' {
		count--
	}
	return count
}

// YieldRow carries code output delta telemetry and token yield metrics for
// a single tool, model, project, or aggregate total.
type YieldRow struct {
	Name           string  `json:"name"`
	Path           string  `json:"path,omitempty"`
	Type           string  `json:"type,omitempty"` // "tool", "model", "project", "total"
	Tokens         int64   `json:"tokens"`
	LinesAdded     int64   `json:"linesAdded"`
	LinesDeleted   int64   `json:"linesDeleted"`
	NetLines       int64   `json:"netLines"`
	GrossLines     int64   `json:"grossLines"`
	FilesTouched   int64   `json:"filesTouched"`
	TokensPerNet   float64 `json:"tokensPerNetLine"`   // 0 if unmeasured
	TokensPerGross float64 `json:"tokensPerGrossLine"` // 0 if unmeasured
	Efficiency     string  `json:"efficiency"`         // HIGH, BALANCED, VERBOSE, EXPLORATORY
	Status         string  `json:"status"`             // MEASURED, UNMEASURED
}

// YieldReport is the top-level payload for the 'thermal yield' command.
type YieldReport struct {
	Type     string     `json:"type"` // "yield"
	Tool     string     `json:"tool,omitempty"`
	Tools    []YieldRow `json:"tools,omitempty"`
	Models   []YieldRow `json:"models,omitempty"`
	Projects []YieldRow `json:"projects,omitempty"`
	Totals   YieldRow   `json:"totals"`
}

// YieldOptions controls filtering, sorting, and row limits for yield reports.
type YieldOptions struct {
	Since string
	Until string
	Last  int
	Sort  string // "tokens", "lines", "yield"; default "tokens"
	Top   int
	Now   time.Time
}

// CalculateYield computes the yield ratio, efficiency rating, and status for a given
// combination of tokens and line deltas. Division by zero on non-positive net lines
// is handled defensively without error.
func CalculateYield(tokens, added, deleted, files int64) (net, gross int64, perNet, perGross float64, eff, status string) {
	net = added - deleted
	gross = added + deleted

	if gross > 0 {
		perGross = float64(tokens) / float64(gross)
	}

	if net > 0 {
		perNet = float64(tokens) / float64(net)
		status = "MEASURED"
		switch {
		case perNet <= 250:
			eff = "HIGH"
		case perNet <= 1000:
			eff = "BALANCED"
		default:
			eff = "VERBOSE"
		}
	} else {
		perNet = 0
		status = "UNMEASURED"
		eff = "EXPLORATORY"
	}

	return net, gross, perNet, perGross, eff, status
}

// AggregateYield aggregates token spend and code delta statistics across tools,
// models, and projects over the requested time window.
func AggregateYield(results []ToolResult, projects []ProjectDay, opts YieldOptions) YieldReport {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	hasWindow := opts.Since != "" || opts.Until != "" || opts.Last > 0

	type modelAgg struct {
		tokens  int64
		added   int64
		deleted int64
		files   int64
	}
	modelsMap := make(map[string]*modelAgg)

	type toolAgg struct {
		name    string
		tokens  int64
		added   int64
		deleted int64
		files   int64
	}
	var toolRows []YieldRow

	var (
		totalTokens  int64
		totalAdded   int64
		totalDeleted int64
		totalFiles   int64
	)

	for _, res := range results {
		var toolTok, toolAdd, toolDel, toolFil int64
		daysWithLines := 0

		for _, day := range res.Daily {
			if isActivityOnly(day) {
				continue
			}
			t, ok := ParseDay(day.Day)
			if !ok || !inWindow(t, since, until, lastStart) {
				continue
			}

			toolTok += day.Tokens
			toolAdd += day.LinesAdded
			toolDel += day.LinesDeleted
			toolFil += day.FilesTouched
			if day.LinesAdded > 0 || day.LinesDeleted > 0 || day.FilesTouched > 0 {
				daysWithLines++
			}

			// Aggregate per-model tokens and lines
			if len(day.Models) > 0 {
				for model, mt := range day.Models {
					mName := model
					m := modelsMap[mName]
					if m == nil {
						m = &modelAgg{}
						modelsMap[mName] = m
					}
					m.tokens += mt.Total()

					// Attribute line changes if tracked per model
					if ld, hasLd := day.ModelLines[model]; hasLd {
						m.added += ld.Added
						m.deleted += ld.Deleted
						m.files += ld.Files
					} else if len(day.Models) == 1 {
						// Single-model day: attribute all day's lines to this model
						m.added += day.LinesAdded
						m.deleted += day.LinesDeleted
						m.files += day.FilesTouched
					}
				}
			}
		}

		// Fallback: If no date window is active and daily rows didn't track lines,
		// use summary level lines from the tool if available.
		if !hasWindow && daysWithLines == 0 && (res.Summary.LinesAdded > 0 || res.Summary.LinesDeleted > 0 || res.Summary.FilesTouched > 0) {
			toolAdd = res.Summary.LinesAdded
			toolDel = res.Summary.LinesDeleted
			toolFil = res.Summary.FilesTouched
			if toolTok == 0 {
				toolTok = res.Summary.LifetimeTokens
			}
		}

		if toolTok == 0 && toolAdd == 0 && toolDel == 0 {
			continue
		}

		net, gross, perNet, perGross, eff, status := CalculateYield(toolTok, toolAdd, toolDel, toolFil)
		toolRows = append(toolRows, YieldRow{
			Name:           res.Name,
			Type:           "tool",
			Tokens:         toolTok,
			LinesAdded:     toolAdd,
			LinesDeleted:   toolDel,
			NetLines:       net,
			GrossLines:     gross,
			FilesTouched:   toolFil,
			TokensPerNet:   perNet,
			TokensPerGross: perGross,
			Efficiency:     eff,
			Status:         status,
		})

		totalTokens += toolTok
		totalAdded += toolAdd
		totalDeleted += toolDel
		totalFiles += toolFil
	}

	// Build Model rows
	var modelRows []YieldRow
	for mName, m := range modelsMap {
		if m.tokens == 0 && m.added == 0 && m.deleted == 0 {
			continue
		}
		net, gross, perNet, perGross, eff, status := CalculateYield(m.tokens, m.added, m.deleted, m.files)
		modelRows = append(modelRows, YieldRow{
			Name:           mName,
			Type:           "model",
			Tokens:         m.tokens,
			LinesAdded:     m.added,
			LinesDeleted:   m.deleted,
			NetLines:       net,
			GrossLines:     gross,
			FilesTouched:   m.files,
			TokensPerNet:   perNet,
			TokensPerGross: perGross,
			Efficiency:     eff,
			Status:         status,
		})
	}

	// Build Project rows if projects are supplied
	type projAgg struct {
		name    string
		tokens  int64
		added   int64
		deleted int64
		files   int64
	}
	projMap := make(map[string]*projAgg)
	for _, p := range projects {
		t, ok := ParseDay(p.Day)
		if !ok || !inWindow(t, since, until, lastStart) {
			continue
		}
		if p.Project == "" {
			continue
		}
		pa := projMap[p.Project]
		if pa == nil {
			pa = &projAgg{name: p.Project}
			projMap[p.Project] = pa
		}
		pa.tokens += p.Tokens
		pa.added += p.LinesAdded
		pa.deleted += p.LinesDeleted
		pa.files += p.FilesTouched
	}
	var projectPaths []string
	for p := range projMap {
		projectPaths = append(projectPaths, p)
	}
	projDisplayNames := ProjectDisplayNames(projectPaths)

	var projectRows []YieldRow
	for _, pa := range projMap {
		net, gross, perNet, perGross, eff, status := CalculateYield(pa.tokens, pa.added, pa.deleted, pa.files)
		slug := projDisplayNames[pa.name]
		if slug == "" {
			slug = ProjectSlug(pa.name)
		}
		projectRows = append(projectRows, YieldRow{
			Name:           slug,
			Path:           pa.name,
			Type:           "project",
			Tokens:         pa.tokens,
			LinesAdded:     pa.added,
			LinesDeleted:   pa.deleted,
			NetLines:       net,
			GrossLines:     gross,
			FilesTouched:   pa.files,
			TokensPerNet:   perNet,
			TokensPerGross: perGross,
			Efficiency:     eff,
			Status:         status,
		})
	}

	// Sort helper
	sortRows := func(rows []YieldRow) {
		switch strings.ToLower(opts.Sort) {
		case "lines":
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].NetLines != rows[j].NetLines {
					return rows[i].NetLines > rows[j].NetLines
				}
				return rows[i].Tokens > rows[j].Tokens
			})
		case "yield":
			sort.Slice(rows, func(i, j int) bool {
				// Measured (positive perNet) before Unmeasured
				iMeasured := rows[i].Status == "MEASURED"
				jMeasured := rows[j].Status == "MEASURED"
				if iMeasured != jMeasured {
					return iMeasured
				}
				if iMeasured && jMeasured {
					if rows[i].TokensPerNet != rows[j].TokensPerNet {
						return rows[i].TokensPerNet < rows[j].TokensPerNet // Most efficient (lowest tok/line) first
					}
				}
				return rows[i].Tokens > rows[j].Tokens
			})
		default: // "tokens"
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].Tokens != rows[j].Tokens {
					return rows[i].Tokens > rows[j].Tokens
				}
				return rows[i].NetLines > rows[j].NetLines
			})
		}
	}

	sortRows(toolRows)
	sortRows(modelRows)
	sortRows(projectRows)

	if opts.Top > 0 {
		if len(toolRows) > opts.Top {
			toolRows = toolRows[:opts.Top]
		}
		if len(modelRows) > opts.Top {
			modelRows = modelRows[:opts.Top]
		}
		if len(projectRows) > opts.Top {
			projectRows = projectRows[:opts.Top]
		}
	}

	// Total Row
	totNet, totGross, totPerNet, totPerGross, totEff, totStatus := CalculateYield(totalTokens, totalAdded, totalDeleted, totalFiles)
	totals := YieldRow{
		Name:           "Totals",
		Type:           "total",
		Tokens:         totalTokens,
		LinesAdded:     totalAdded,
		LinesDeleted:   totalDeleted,
		NetLines:       totNet,
		GrossLines:     totGross,
		FilesTouched:   totalFiles,
		TokensPerNet:   totPerNet,
		TokensPerGross: totPerGross,
		Efficiency:     totEff,
		Status:         totStatus,
	}

	return YieldReport{
		Type:     "yield",
		Tools:    toolRows,
		Models:   modelRows,
		Projects: projectRows,
		Totals:   totals,
	}
}
