// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// LoadCommandCodeData reads command-code session transcripts in parallel using a bounded worker pool.
// Modern transcripts record per-assistant usage, a per-message model, and the session working
// directory in the header. inputTokens already includes cache read and write tokens, so those are
// subtracted to keep token types disjoint. Legacy flat transcripts carry no usage, so their
// assistant messages contribute turns only.
func LoadCommandCodeData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	projectsDir := filepath.Join(dataDir, "projects")

	var summary thermal.Summary
	summary.Tool = "command-code"

	pattern := filepath.Join(projectsDir, "*", "*.jsonl")
	allFiles, err := filepath.Glob(pattern)
	if err != nil {
		return summary, nil, nil, err
	}

	var sessionFiles []string
	for _, p := range allFiles {
		base := filepath.Base(p)
		if strings.Contains(base, ".checkpoints.") || strings.HasPrefix(base, "hooks-audit-") {
			continue
		}
		sessionFiles = append(sessionFiles, p)
	}

	type msgAgg struct {
		day        string
		input      int64
		output     int64
		cacheRead  int64
		cacheWrite int64
		cost       float64
		model      string
	}
	type fileResult struct {
		msgs       []msgAgg
		project    string
		metaModel  string
		durationMs int64
		hasSession bool
		warning    string
	}

	results := make(chan fileResult, len(sessionFiles))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8) // max 8 concurrent workers

	for _, path := range sessionFiles {
		sem <- struct{}{}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			defer func() { <-sem }()

			var res fileResult

			f, err := os.Open(p)
			if err != nil {
				res.warning = formatScanWarning(p, err)
				results <- res
				return
			}

			var cwd string
			var firstTs, lastTs time.Time
			seenIDs := make(map[string]struct{})
			scanner := newJSONLScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var rec struct {
					Type      string `json:"type"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Cwd       string `json:"cwd"`
					Role      string `json:"role"`
					Model     string `json:"model"`
					Message   struct {
						Role string `json:"role"`
					} `json:"message"`
					Usage *struct {
						InputTokens      int64   `json:"inputTokens"`
						OutputTokens     int64   `json:"outputTokens"`
						CacheReadTokens  int64   `json:"cacheReadTokens"`
						CacheWriteTokens int64   `json:"cacheWriteTokens"`
						CostUSD          float64 `json:"costUsd"`
					} `json:"usage"`
				}
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					continue
				}
				if cwd == "" && rec.Cwd != "" {
					cwd = rec.Cwd
				}

				role := rec.Message.Role
				if role == "" {
					role = rec.Role
				}
				if role == "" {
					continue
				}
				if rec.ID != "" {
					if _, seen := seenIDs[rec.ID]; seen {
						continue
					}
					seenIDs[rec.ID] = struct{}{}
				}

				t, err := time.Parse(time.RFC3339, rec.Timestamp)
				if err != nil {
					continue
				}
				if firstTs.IsZero() || t.Before(firstTs) {
					firstTs = t
				}
				if t.After(lastTs) {
					lastTs = t
				}
				res.hasSession = true

				if role != "assistant" {
					continue
				}
				m := msgAgg{day: thermal.LocalDay(t.Local()), model: modelName(rec.Model)}
				if rec.Usage != nil {
					cache := rec.Usage.CacheReadTokens + rec.Usage.CacheWriteTokens
					input := rec.Usage.InputTokens - cache
					if input < 0 {
						input = 0
					}
					m.input = input
					m.output = rec.Usage.OutputTokens
					m.cacheRead = rec.Usage.CacheReadTokens
					m.cacheWrite = rec.Usage.CacheWriteTokens
					m.cost = rec.Usage.CostUSD
				}
				res.msgs = append(res.msgs, m)
			}
			if err := scanner.Err(); err != nil {
				res.warning = formatScanWarning(p, err)
			}
			f.Close()

			res.project = thermal.ProjectKey(cwd)
			if res.hasSession {
				res.durationMs = lastTs.Sub(firstTs).Milliseconds()
			}

			metaPath := strings.TrimSuffix(p, ".jsonl") + ".meta.json"
			if data, err := os.ReadFile(metaPath); err == nil {
				var meta struct {
					Model string `json:"model"`
				}
				if json.Unmarshal(data, &meta) == nil {
					res.metaModel = meta.Model
				}
			}
			results <- res
		}(path)
	}

	wg.Wait()
	close(results)

	type dayAgg struct {
		input, output         int64
		cacheRead, cacheWrite int64
		cost                  float64
		turns                 int
		models                map[string]thermal.ModelTokens
	}

	byDay := make(map[string]*dayAgg)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
	projectModels := make(map[projectDayKey]map[string]thermal.ModelTokens)
	modelCounts := make(map[string]int64)

	for res := range results {
		if res.warning != "" {
			summary.Warnings = append(summary.Warnings, res.warning)
		}
		if !res.hasSession {
			continue
		}
		summary.Sessions++
		if res.durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = res.durationMs
		}

		for _, m := range res.msgs {
			model := m.model
			if model == "" {
				model = modelName(res.metaModel)
			}
			if model != "" {
				modelCounts[model]++
			}

			tokens := m.input + m.output + m.cacheRead + m.cacheWrite
			summary.InputTokens += m.input
			summary.OutputTokens += m.output
			summary.CacheTokens += m.cacheRead + m.cacheWrite
			summary.LifetimeTokens += tokens
			summary.Cost += m.cost

			agg := byDay[m.day]
			if agg == nil {
				agg = &dayAgg{models: make(map[string]thermal.ModelTokens)}
				byDay[m.day] = agg
			}
			agg.input += m.input
			agg.output += m.output
			agg.cacheRead += m.cacheRead
			agg.cacheWrite += m.cacheWrite
			agg.cost += m.cost
			agg.turns++
			if model != "" && tokens > 0 {
				agg.models[model] = agg.models[model].Add(thermal.ModelTokens{
					Input:      m.input,
					Output:     m.output,
					CacheRead:  m.cacheRead,
					CacheWrite: m.cacheWrite,
				})
			}

			if res.project == "" {
				continue
			}
			key := projectDayKey{m.day, res.project}
			pd := byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{Project: res.project, Day: m.day}
				byProjectDay[key] = pd
			}
			pd.Input += m.input
			pd.Output += m.output
			pd.CacheRead += m.cacheRead
			pd.CacheWrite += m.cacheWrite
			pd.Tokens += tokens
			pd.Cost += m.cost
			pd.Turns++
			if model != "" && tokens > 0 {
				if projectModels[key] == nil {
					projectModels[key] = make(map[string]thermal.ModelTokens)
				}
				projectModels[key][model] = projectModels[key][model].Add(thermal.ModelTokens{
					Input:      m.input,
					Output:     m.output,
					CacheRead:  m.cacheRead,
					CacheWrite: m.cacheWrite,
				})
			}
		}
	}

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: agg.input + agg.output + agg.cacheRead + agg.cacheWrite,
			Input:  agg.input,
			Output: agg.output,
			Cache:  agg.cacheRead + agg.cacheWrite,
			Cost:   agg.cost,
			Turns:  agg.turns,
			Models: agg.models,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	projects := make([]thermal.ProjectDay, 0, len(byProjectDay))
	for key, pd := range byProjectDay {
		pd.Models = projectModels[key]
		projects = append(projects, *pd)
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Day != projects[j].Day {
			return projects[i].Day < projects[j].Day
		}
		return projects[i].Project < projects[j].Project
	})

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}

	if len(daily) == 0 {
		return loadJsonlData(dataDir, "t", true)
	}

	sort.Strings(summary.Warnings)
	return summary, daily, projects, nil
}
