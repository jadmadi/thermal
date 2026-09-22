// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// dshSessionWrapper handles individual session files which usually enclose
// the identity and rows inside a "record" object. Top-level fallbacks exist
// in case a variant stores them unnested.
type dshSessionWrapper struct {
	Version  int         `json:"version"`
	Record   dshRecord   `json:"record"`
	Identity dshIdentity `json:"identity"`
	Rows     dshRows     `json:"rows"`
}

type dshRecord struct {
	Identity dshIdentity `json:"identity"`
	Rows     dshRows     `json:"rows"`
}

type dshIdentity struct {
	CreatedAt any    `json:"createdAt"`
	Cwd       string `json:"cwd"`
}

type dshRows struct {
	SessionStats   *dshSessionStatsRow   `json:"sessionStats"`
	ModelSelection *dshModelSelectionRow `json:"modelSelection"`
	TokenUsage     *dshTokenUsageRow     `json:"tokenUsage"`
}

type dshSessionStatsRow struct {
	Val *dshSessionStatsVal `json:"val"`
}

type dshSessionStatsVal struct {
	Turns  int   `json:"turns"`
	Steps  int   `json:"steps"`
	LLMMs  int64 `json:"llmMs"`
	ToolMs int64 `json:"toolMs"`
}

type dshModelSelectionRow struct {
	Val *dshModelSelectionVal `json:"val"`
}

type dshModelSelectionVal struct {
	LastUsed *dshLastUsed `json:"lastUsed"`
}

type dshLastUsed struct {
	Model string `json:"model"`
}

type dshTokenUsageRow struct {
	Val *dshTokenUsageVal `json:"val"`
}

type dshTokenUsageVal struct {
	Totals *dshTokenTotals `json:"totals"`
}

type dshTokenTotals struct {
	UncachedInputTokens int64 `json:"uncachedInputTokens"`
	InputTokens         int64 `json:"inputTokens"`
	OutputTokens        int64 `json:"outputTokens"`
	ReasoningTokens     int64 `json:"reasoningTokens"`
	CacheReadTokens     int64 `json:"cacheReadTokens"`
	CacheWriteTokens    int64 `json:"cacheWriteTokens"`
}

type dshMonolith struct {
	Tables struct {
		Sessions map[string]dshRecord `json:"sessions"`
	} `json:"tables"`
}

// parseDshTimestamp robustly parses Unix timestamps (seconds, milliseconds,
// microseconds) or RFC3339 strings from dsh identity records.
func parseDshTimestamp(v any) (time.Time, bool) {
	switch val := v.(type) {
	case float64:
		return parseDshEpoch(int64(val))
	case int64:
		return parseDshEpoch(val)
	case int:
		return parseDshEpoch(int64(val))
	case string:
		val = strings.TrimSpace(val)
		if val == "" {
			return time.Time{}, false
		}
		if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
			return t, true
		}
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t, true
		}
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return parseDshEpoch(n)
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return parseDshEpoch(int64(f))
		}
	}
	return time.Time{}, false
}

func parseDshEpoch(n int64) (time.Time, bool) {
	if n <= 0 {
		return time.Time{}, false
	}
	if n > 1e14 { // microseconds
		return time.UnixMicro(n), true
	}
	if n > 1e11 { // milliseconds
		return time.UnixMilli(n), true
	}
	// seconds
	return time.Unix(n, 0), true
}

type dshSessionResult struct {
	id         string
	day        string
	project    string
	model      string
	input      int64
	output     int64
	cacheRead  int64
	cacheWrite int64
	reasoning  int64
	total      int64
	turns      int
	durationMs int64
	active     bool
	warning    string
}

func extractDshSession(id string, rec dshRecord) dshSessionResult {
	t, ok := parseDshTimestamp(rec.Identity.CreatedAt)
	if !ok || t.IsZero() {
		return dshSessionResult{id: id}
	}
	day := thermal.LocalDay(t.Local())

	var turns int
	var durationMs int64
	if rec.Rows.SessionStats != nil && rec.Rows.SessionStats.Val != nil {
		turns = rec.Rows.SessionStats.Val.Turns
		durationMs = rec.Rows.SessionStats.Val.LLMMs + rec.Rows.SessionStats.Val.ToolMs
	}

	var model string
	if rec.Rows.ModelSelection != nil && rec.Rows.ModelSelection.Val != nil && rec.Rows.ModelSelection.Val.LastUsed != nil {
		model = modelName(rec.Rows.ModelSelection.Val.LastUsed.Model)
	}

	var input, output, cacheRead, cacheWrite, reasoning int64
	if rec.Rows.TokenUsage != nil && rec.Rows.TokenUsage.Val != nil && rec.Rows.TokenUsage.Val.Totals != nil {
		tot := rec.Rows.TokenUsage.Val.Totals
		cacheRead = tot.CacheReadTokens
		cacheWrite = tot.CacheWriteTokens
		reasoning = tot.ReasoningTokens
		output = tot.OutputTokens

		if tot.UncachedInputTokens > 0 {
			input = tot.UncachedInputTokens
		} else if tot.InputTokens > 0 {
			if tot.InputTokens >= cacheRead {
				input = tot.InputTokens - cacheRead
			} else {
				input = tot.InputTokens
			}
		}

		if reasoning > 0 && output > reasoning {
			output -= reasoning
		}
	}

	total := input + output + cacheRead + cacheWrite + reasoning
	if turns <= 0 && total > 0 {
		turns = 1
	}

	if turns == 0 && total == 0 {
		return dshSessionResult{id: id}
	}

	var project string
	if rec.Identity.Cwd != "" {
		project = thermal.ProjectKey(rec.Identity.Cwd)
	}

	return dshSessionResult{
		id:         id,
		day:        day,
		project:    project,
		model:      model,
		input:      input,
		output:     output,
		cacheRead:  cacheRead,
		cacheWrite: cacheWrite,
		reasoning:  reasoning,
		total:      total,
		turns:      turns,
		durationMs: durationMs,
		active:     true,
	}
}

func loadDshSessionFile(path string) dshSessionResult {
	id := strings.TrimSuffix(filepath.Base(path), ".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return dshSessionResult{id: id, warning: formatScanWarning(path, err)}
	}

	var wrapper dshSessionWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return dshSessionResult{id: id, warning: formatScanWarning(path, err)}
	}

	rec := wrapper.Record
	if rec.Identity.CreatedAt == nil && wrapper.Identity.CreatedAt != nil {
		rec = dshRecord{
			Identity: wrapper.Identity,
			Rows:     wrapper.Rows,
		}
	}

	return extractDshSession(id, rec)
}

func loadDshMonolith(path string, seenIDs map[string]bool) ([]dshSessionResult, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var mono dshMonolith
	if err := json.Unmarshal(data, &mono); err != nil {
		return nil, nil, err
	}

	var results []dshSessionResult
	var warnings []string
	for id, rec := range mono.Tables.Sessions {
		if seenIDs[id] {
			continue
		}
		res := extractDshSession(id, rec)
		if res.warning != "" {
			warnings = append(warnings, res.warning)
		}
		if res.active {
			seenIDs[id] = true
			results = append(results, res)
		}
	}
	return results, warnings, nil
}

func loadDshSingleFile(path string) ([]dshSessionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try monolith first
	var mono dshMonolith
	if err := json.Unmarshal(data, &mono); err == nil && len(mono.Tables.Sessions) > 0 {
		seen := make(map[string]bool)
		res, _, err := loadDshMonolith(path, seen)
		return res, err
	}

	// Otherwise try individual session
	res := loadDshSessionFile(path)
	if res.warning != "" {
		return nil, fmt.Errorf("%s", res.warning)
	}
	if res.active {
		return []dshSessionResult{res}, nil
	}
	return nil, nil
}

// LoadDshData scans DeepSeek harness (dsh) session cache records.
// It scans ~/.dsh/storages/session_projcache/sessions/*.json in parallel,
// supplements with ~/.dsh/storages/session_projcache.json when present,
// and attributes projects and models.
func LoadDshData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	var summary thermal.Summary
	summary.Tool = "DeepSeek (DSH)"

	info, err := os.Stat(dataDir)
	if err != nil {
		return summary, nil, nil, err
	}

	var sessionResults []dshSessionResult
	seenIDs := make(map[string]bool)

	if !info.IsDir() {
		results, err := loadDshSingleFile(dataDir)
		if err != nil {
			return summary, nil, nil, err
		}
		sessionResults = append(sessionResults, results...)
	} else {
		candidatePatterns := []string{
			filepath.Join(dataDir, "storages", "session_projcache", "sessions", "*.json"),
			filepath.Join(dataDir, "session_projcache", "sessions", "*.json"),
			filepath.Join(dataDir, "sessions", "*.json"),
		}

		var sessionFiles []string
		for _, pat := range candidatePatterns {
			matches, err := filepath.Glob(pat)
			if err == nil && len(matches) > 0 {
				sessionFiles = matches
				break
			}
		}

		if len(sessionFiles) == 0 {
			matches, err := filepath.Glob(filepath.Join(dataDir, "*.json"))
			if err == nil && len(matches) > 0 {
				for _, m := range matches {
					base := filepath.Base(m)
					if base != "session_projcache.json" && base != "workspace.json" {
						sessionFiles = append(sessionFiles, m)
					}
				}
			}
		}

		if len(sessionFiles) > 0 {
			resultsChan := make(chan dshSessionResult, len(sessionFiles))
			var wg sync.WaitGroup
			sem := make(chan struct{}, 16)

			for _, path := range sessionFiles {
				sem <- struct{}{}
				wg.Add(1)
				go func(p string) {
					defer wg.Done()
					defer func() { <-sem }()

					res := loadDshSessionFile(p)
					resultsChan <- res
				}(path)
			}

			wg.Wait()
			close(resultsChan)

			for r := range resultsChan {
				if r.warning != "" {
					summary.Warnings = append(summary.Warnings, r.warning)
				}
				if r.active {
					seenIDs[r.id] = true
					sessionResults = append(sessionResults, r)
				}
			}
		}

		monoPaths := []string{
			filepath.Join(dataDir, "storages", "session_projcache.json"),
			filepath.Join(dataDir, "session_projcache.json"),
		}
		for _, mp := range monoPaths {
			if _, err := os.Stat(mp); err == nil {
				monoRes, warnings, err := loadDshMonolith(mp, seenIDs)
				if err != nil {
					summary.Warnings = append(summary.Warnings, formatScanWarning(mp, err))
				} else {
					summary.Warnings = append(summary.Warnings, warnings...)
					sessionResults = append(sessionResults, monoRes...)
				}
				break
			}
		}
	}

	if len(sessionResults) == 0 {
		sort.Strings(summary.Warnings)
		return summary, nil, nil, nil
	}

	type dayAgg struct {
		input, output, cacheRead, cacheWrite, reasoning int64
		turns                                           int
		models                                          map[string]thermal.ModelTokens
	}
	byDay := make(map[string]*dayAgg)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
	projectModels := make(map[projectDayKey]map[string]thermal.ModelTokens)
	modelCounts := make(map[string]int64)

	for _, res := range sessionResults {
		summary.Sessions++
		if res.durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = res.durationMs
		}
		summary.InputTokens += res.input
		summary.OutputTokens += res.output
		summary.CacheTokens += res.cacheRead + res.cacheWrite
		summary.ReasoningTokens += res.reasoning
		summary.LifetimeTokens += res.total

		if res.model != "" {
			modelCounts[res.model]++
		}

		agg := byDay[res.day]
		if agg == nil {
			agg = &dayAgg{models: make(map[string]thermal.ModelTokens)}
			byDay[res.day] = agg
		}
		agg.input += res.input
		agg.output += res.output
		agg.cacheRead += res.cacheRead
		agg.cacheWrite += res.cacheWrite
		agg.reasoning += res.reasoning
		agg.turns += res.turns

		if res.model != "" && res.total > 0 {
			agg.models[res.model] = agg.models[res.model].Add(thermal.ModelTokens{
				Input:      res.input,
				Output:     res.output,
				CacheRead:  res.cacheRead,
				CacheWrite: res.cacheWrite,
				Reasoning:  res.reasoning,
			})
		}

		if res.project != "" {
			key := projectDayKey{res.day, res.project}
			pd := byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{Project: res.project, Day: res.day}
				byProjectDay[key] = pd
			}
			pd.Input += res.input
			pd.Output += res.output
			pd.CacheRead += res.cacheRead
			pd.CacheWrite += res.cacheWrite
			pd.Reasoning += res.reasoning
			pd.Tokens += res.total
			pd.Turns += res.turns

			if res.model != "" && res.total > 0 {
				if projectModels[key] == nil {
					projectModels[key] = make(map[string]thermal.ModelTokens)
				}
				projectModels[key][res.model] = projectModels[key][res.model].Add(thermal.ModelTokens{
					Input:      res.input,
					Output:     res.output,
					CacheRead:  res.cacheRead,
					CacheWrite: res.cacheWrite,
					Reasoning:  res.reasoning,
				})
			}
		}
	}

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:       day,
			Tokens:    agg.input + agg.output + agg.cacheRead + agg.cacheWrite + agg.reasoning,
			Input:     agg.input,
			Output:    agg.output,
			Cache:     agg.cacheRead + agg.cacheWrite,
			Reasoning: agg.reasoning,
			Turns:     agg.turns,
			Models:    agg.models,
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

	sort.Strings(summary.Warnings)
	return summary, daily, projects, nil
}
