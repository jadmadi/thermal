// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
	_ "modernc.org/sqlite" // register sqlite driver
)

// LoadCodexData reads Codex session data from state_5.sqlite primary source
// or falls back to history.jsonl if stateDB doesn't exist.
func LoadCodexData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	stateDB := filepath.Join(dataDir, "state_5.sqlite")
	if _, err := os.Stat(stateDB); err == nil {
		return loadCodexFromStateDB(stateDB)
	}
	sum, daily, proj, err := loadJsonlData(dataDir, "ts", false)
	if err == nil {
		sum.MainlineTokens = sum.LifetimeTokens
		sum.RootSessions = sum.Sessions
	}
	return sum, daily, proj, err
}

func loadCodexFromStateDB(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot open %s: %w", dbPath, err)
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA mmap_size=268435456")

	// Query thread lineage DAG edges (parent-child relationships) if supported
	forkThreads := make(map[string]string)
	if hasTable(db, "thread_spawn_edges") {
		edgeRows, err := db.Query(`SELECT child_thread_id, parent_thread_id FROM thread_spawn_edges`)
		if err == nil {
			for edgeRows.Next() {
				var child, parent string
				if err := edgeRows.Scan(&child, &parent); err == nil {
					forkThreads[child] = parent
				}
			}
			_ = edgeRows.Err()
			edgeRows.Close()
		}
	}

	// Older Codex schemas lack the cwd column.
	cwdExpr := "''"
	if hasColumn(db, "threads", "cwd") {
		cwdExpr = "cwd"
	}
	threadSourceExpr := "''"
	if hasColumn(db, "threads", "thread_source") {
		threadSourceExpr = "thread_source"
	}
	rows, err := db.Query(`
		SELECT id, tokens_used, model, source, reasoning_effort, agent_role,
		       created_at, updated_at, rollout_path, ` + cwdExpr + `, ` + threadSourceExpr + `
		FROM threads
		WHERE archived = 0
		ORDER BY created_at
	`)
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot query threads: %w", err)
	}
	defer rows.Close()

	type dayAgg struct {
		tokens          int64
		turns           int
		input           int64
		output          int64
		reasoning       int64
		cache           int64
		linesAdded      int64
		linesDeleted    int64
		filesTouched    int64
		models          map[string]thermal.ModelTokens
		modelLines      map[string]thermal.LineDelta
		reasoningEffort map[string]int
		mainlineTokens  int64
		forkTokens      int64
	}
	byDay := make(map[string]*dayAgg)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
	projectModels := make(map[projectDayKey]map[string]thermal.ModelTokens)
	modelCounts := make(map[string]int64)
	sourceCounts := make(map[string]int)
	reasoningCounts := make(map[string]int)

	var summary thermal.Summary
	summary.Tool = "Codex"

	type threadInfo struct {
		id          string
		tokensUsed  int64
		rolloutPath string
		day         string
		model       string
		project     string
		isFork      bool
	}
	var threads []threadInfo

	for rows.Next() {
		var t threadInfo
		var model, source, reasoning, agentRole, rolloutPath, cwd, threadSource sql.NullString
		var createdAt, updatedAt int64

		if err := rows.Scan(&t.id, &t.tokensUsed, &model, &source, &reasoning,
			&agentRole, &createdAt, &updatedAt, &rolloutPath, &cwd, &threadSource); err != nil {
			continue
		}

		t.day = thermal.LocalDay(time.Unix(createdAt, 0).Local())
		t.project = thermal.ProjectKey(cwd.String)

		// Classify thread as exploratory fork / subagent vs mainline root session
		isFork := false
		if _, ok := forkThreads[t.id]; ok {
			isFork = true
		} else if strings.EqualFold(threadSource.String, "subagent") || strings.Contains(source.String, "parent_thread_id") {
			isFork = true
		}
		t.isFork = isFork

		durationMs := (updatedAt - createdAt) * 1000
		if durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = durationMs
		}

		canonicalModel := modelName(model.String)
		t.model = canonicalModel
		if canonicalModel != "" {
			modelCounts[canonicalModel] += t.tokensUsed
		}
		if source.String != "" {
			sourceCounts[source.String]++
		}
		if reasoning.String != "" {
			reasoningCounts[reasoning.String]++
		}

		summary.Sessions++
		summary.LifetimeTokens += t.tokensUsed

		if isFork {
			summary.ForkTokens += t.tokensUsed
			summary.ForkSessions++
		} else {
			summary.MainlineTokens += t.tokensUsed
			summary.RootSessions++
		}

		agg := byDay[t.day]
		if agg == nil {
			agg = &dayAgg{models: make(map[string]thermal.ModelTokens)}
			byDay[t.day] = agg
		}
		agg.tokens += t.tokensUsed
		agg.turns++
		if isFork {
			agg.forkTokens += t.tokensUsed
		} else {
			agg.mainlineTokens += t.tokensUsed
		}
		if reasoning.String != "" {
			if agg.reasoningEffort == nil {
				agg.reasoningEffort = make(map[string]int)
			}
			agg.reasoningEffort[reasoning.String]++
		}

		if t.project != "" {
			key := projectDayKey{t.day, t.project}
			pd := byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{Project: t.project, Day: t.day}
				byProjectDay[key] = pd
			}
			pd.Tokens += t.tokensUsed
			pd.Turns++
		}

		if rolloutPath.Valid {
			t.rolloutPath = rolloutPath.String
		}
		threads = append(threads, t)
	}
	if err := rows.Err(); err != nil {
		return thermal.Summary{}, nil, nil, err
	}

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}
	if len(sourceCounts) > 0 {
		summary.AgentBreakdown = sourceCounts
	}
	if len(reasoningCounts) > 0 {
		summary.ReasoningBreakdown = reasoningCounts
	}

	// Load incremental rollout cache
	rolloutCache := loadCodexRolloutCache(dbPath)
	var cacheMu sync.Mutex
	cacheDirty := false

	// Prune dead entries
	activeRollouts := make(map[string]bool)
	for _, t := range threads {
		if t.rolloutPath != "" {
			activeRollouts[t.rolloutPath] = true
		}
	}
	rolloutCache.mu.Lock()
	for p := range rolloutCache.Entries {
		if !activeRollouts[p] {
			delete(rolloutCache.Entries, p)
			cacheDirty = true
		}
	}
	rolloutCache.mu.Unlock()

	// Bounded worker pool for parallel rollout scanning
	results := make([]rolloutResult, len(threads))
	var wg sync.WaitGroup
	workerLimit := make(chan struct{}, 8) // max 8 concurrent workers

	for i, t := range threads {
		if t.rolloutPath == "" || t.tokensUsed == 0 {
			continue
		}
		workerLimit <- struct{}{}
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-workerLimit }()

			rolloutCache.mu.RLock()
			var cachedPtr *CodexRolloutEntry
			if entry, ok := rolloutCache.Entries[path]; ok {
				cachedPtr = &entry
			}
			rolloutCache.mu.RUnlock()

			res, newEntry, dirty := scanOrReuseRollout(path, cachedPtr)
			results[idx] = res

			if dirty {
				cacheMu.Lock()
				rolloutCache.mu.Lock()
				rolloutCache.Entries[path] = newEntry
				rolloutCache.mu.Unlock()
				cacheDirty = true
				cacheMu.Unlock()
			}
		}(i, t.rolloutPath)
	}
	wg.Wait()

	if cacheDirty {
		_ = saveCodexRolloutCache(dbPath, rolloutCache)
	}

	for i, res := range results {
		if res.warning != "" {
			summary.Warnings = append(summary.Warnings, res.warning)
		}
		summary.LinesAdded += res.linesAdded
		summary.LinesDeleted += res.linesDeleted
		summary.FilesTouched += res.filesTouched

		t := threads[i]
		agg := byDay[t.day]
		if agg != nil {
			agg.linesAdded += res.linesAdded
			agg.linesDeleted += res.linesDeleted
			agg.filesTouched += res.filesTouched
			if t.model != "" && (res.linesAdded > 0 || res.linesDeleted > 0 || res.filesTouched > 0) {
				if agg.modelLines == nil {
					agg.modelLines = make(map[string]thermal.LineDelta)
				}
				ml := agg.modelLines[t.model]
				ml.Added += res.linesAdded
				ml.Deleted += res.linesDeleted
				ml.Files += res.filesTouched
				agg.modelLines[t.model] = ml
			}
		}
		if t.project != "" {
			key := projectDayKey{t.day, t.project}
			if pd := byProjectDay[key]; pd != nil {
				pd.LinesAdded += res.linesAdded
				pd.LinesDeleted += res.linesDeleted
				pd.FilesTouched += res.filesTouched
			}
		}

		b := res.breakdown
		if b == nil {
			continue
		}
		summary.InputTokens += b.input
		summary.OutputTokens += b.output
		summary.ReasoningTokens += b.reasoning
		summary.CacheTokens += b.cache

		// The rollout total is cumulative for the thread and can drift from
		// threads.tokens_used, so scale the breakdown to tokens_used before
		// attributing it to the thread's day and model.
		if t.tokensUsed <= 0 {
			continue
		}
		breakdownTotal := b.input + b.output + b.reasoning + b.cache
		if breakdownTotal <= 0 {
			continue
		}
		ratio := float64(t.tokensUsed) / float64(breakdownTotal)
		scaled := tokenBreakdown{
			input:     int64(float64(b.input) * ratio),
			output:    int64(float64(b.output) * ratio),
			reasoning: int64(float64(b.reasoning) * ratio),
			cache:     int64(float64(b.cache) * ratio),
		}
		agg = byDay[t.day]
		if agg == nil {
			continue
		}
		agg.input += scaled.input
		agg.output += scaled.output
		agg.reasoning += scaled.reasoning
		agg.cache += scaled.cache
		if t.model != "" {
			if agg.models == nil {
				agg.models = make(map[string]thermal.ModelTokens)
			}
			agg.models[t.model] = agg.models[t.model].Add(thermal.ModelTokens{
				Input:     scaled.input,
				Output:    scaled.output,
				Reasoning: scaled.reasoning,
				CacheRead: scaled.cache,
			})
		}

		if t.project == "" {
			continue
		}
		key := projectDayKey{t.day, t.project}
		pd := byProjectDay[key]
		if pd == nil {
			continue
		}
		pd.Input += scaled.input
		pd.Output += scaled.output
		pd.Reasoning += scaled.reasoning
		pd.CacheRead += scaled.cache
		if t.model != "" {
			if projectModels[key] == nil {
				projectModels[key] = make(map[string]thermal.ModelTokens)
			}
			projectModels[key][t.model] = projectModels[key][t.model].Add(thermal.ModelTokens{
				Input:     scaled.input,
				Output:    scaled.output,
				Reasoning: scaled.reasoning,
				CacheRead: scaled.cache,
			})
		}
	}

	breakdownTotal := summary.InputTokens + summary.OutputTokens +
		summary.ReasoningTokens + summary.CacheTokens
	if breakdownTotal > 0 && breakdownTotal != summary.LifetimeTokens {
		ratio := float64(summary.LifetimeTokens) / float64(breakdownTotal)
		summary.InputTokens = int64(float64(summary.InputTokens) * ratio)
		summary.OutputTokens = int64(float64(summary.OutputTokens) * ratio)
		summary.ReasoningTokens = int64(float64(summary.ReasoningTokens) * ratio)
		summary.CacheTokens = int64(float64(summary.CacheTokens) * ratio)

		// Reconcile rounding remainder directly onto InputTokens
		remainder := summary.LifetimeTokens - (summary.InputTokens + summary.OutputTokens + summary.ReasoningTokens + summary.CacheTokens)
		summary.InputTokens += remainder
	} else if breakdownTotal == 0 {
		summary.InputTokens = summary.LifetimeTokens
	}

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:             day,
			Tokens:          agg.tokens,
			Input:           agg.input,
			Output:          agg.output,
			Reasoning:       agg.reasoning,
			Cache:           agg.cache,
			Turns:           agg.turns,
			Models:          agg.models,
			LinesAdded:      agg.linesAdded,
			LinesDeleted:    agg.linesDeleted,
			FilesTouched:    agg.filesTouched,
			ModelLines:      agg.modelLines,
			ReasoningEffort: agg.reasoningEffort,
			MainlineTokens:  agg.mainlineTokens,
			ForkTokens:      agg.forkTokens,
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

	sort.Strings(summary.Warnings)
	return summary, daily, projects, nil
}

type tokenBreakdown struct {
	input, output, reasoning, cache int64
}

func readLastTokenBreakdown(rolloutPath string) (*tokenBreakdown, int64, int64, int64, string) {
	b, la, ld, ft, w, _ := scanRolloutFull(rolloutPath)
	return b, la, ld, ft, w
}

func nonNegative(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}

// ScanCodexReceipts extracts verifiable work receipts from Codex state database and rollout logs.
func ScanCodexReceipts(dataDir string, pricer thermal.Pricer) []thermal.WorkReceipt {
	return thermal.ScanCodexReceipts(dataDir, pricer)
}
