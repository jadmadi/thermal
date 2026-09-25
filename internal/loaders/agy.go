// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jadmadi/thermal/internal/thermal"
)

// ResolveAgyBrainDir accepts either the Antigravity data root (containing
// brain/) or the brain dir itself. If caller specifies legacy antigravity-cli,
// it prefers the primary unified hub (~/.gemini/antigravity/brain) if available,
// falling back to legacy archive if primary does not exist.
func ResolveAgyBrainDir(dataDir string) string {
	if filepath.Base(dataDir) == "brain" {
		return dataDir
	}

	// Prefer primary unified hub if caller passed legacy root
	if filepath.Base(dataDir) == "antigravity-cli" {
		primaryHub := filepath.Join(filepath.Dir(dataDir), "antigravity", "brain")
		if st, err := os.Stat(primaryHub); err == nil && st.IsDir() {
			return primaryHub
		}
	}

	primary := filepath.Join(dataDir, "brain")
	if st, err := os.Stat(primary); err == nil && st.IsDir() {
		return primary
	}

	// Fallback to legacy archive if primary doesn't exist
	if filepath.Base(dataDir) == "antigravity" {
		legacy := filepath.Join(filepath.Dir(dataDir), "antigravity-cli", "brain")
		if st, err := os.Stat(legacy); err == nil && st.IsDir() {
			return legacy
		}
	}
	return primary
}

func parseWorkspaceURIs(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return ""
	}
	var uris []string
	if err := json.Unmarshal([]byte(raw), &uris); err == nil {
		for _, u := range uris {
			cand := strings.TrimSpace(u)
			cand = strings.TrimPrefix(cand, "file://")
			if unescaped, err := url.PathUnescape(cand); err == nil {
				cand = unescaped
			}
			if filepath.IsAbs(cand) {
				return filepath.Clean(cand)
			}
		}
	}
	cand := strings.TrimPrefix(raw, "file://")
	if unescaped, err := url.PathUnescape(cand); err == nil {
		cand = unescaped
	}
	if filepath.IsAbs(cand) {
		return filepath.Clean(cand)
	}
	return ""
}

func loadAgyWorkspaces(dataDir string) map[string]string {
	wsMap := make(map[string]string)
	var candidates []string
	if filepath.Base(dataDir) == "brain" {
		candidates = append(candidates, filepath.Join(filepath.Dir(dataDir), "conversation_summaries.db"))
	}
	candidates = append(candidates,
		filepath.Join(dataDir, "conversation_summaries.db"),
		filepath.Join(filepath.Dir(dataDir), "antigravity-cli", "conversation_summaries.db"),
		filepath.Join(filepath.Dir(dataDir), "antigravity", "conversation_summaries.db"),
	)

	seen := make(map[string]bool)
	for _, dbPath := range candidates {
		if seen[dbPath] {
			continue
		}
		seen[dbPath] = true
		if fi, err := os.Stat(dbPath); err != nil || fi.IsDir() {
			continue
		}

		db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=268435456")
		if err != nil {
			continue
		}

		rows, err := db.Query("SELECT conversation_id, workspace_uris FROM conversation_summaries")
		if err != nil {
			db.Close()
			continue
		}

		for rows.Next() {
			var convID, rawURIs string
			if err := rows.Scan(&convID, &rawURIs); err == nil {
				if ws := parseWorkspaceURIs(rawURIs); ws != "" {
					if wsMap[convID] == "" {
						wsMap[convID] = thermal.ProjectKey(ws)
					}
				}
			}
		}
		_ = rows.Err()
		db.Close()
	}
	return wsMap
}

type sessionResult struct {
	project     string
	steps       int
	firstTs     time.Time
	lastTs      time.Time
	dayCounts   map[string]int
	dayModels   map[string]map[string]int
	modelCounts map[string]int64
	activeModel string
	warnings    []string
}

// parseAgyModelSetting extracts a canonical model name from an Antigravity user settings change.
func parseAgyModelSetting(content string) string {
	const needle = "Model Selection"
	idx := strings.Index(content, needle)
	if idx < 0 {
		return ""
	}
	rest := content[idx+len(needle):]
	toIdx := strings.Index(rest, " to ")
	if toIdx < 0 {
		return ""
	}
	target := rest[toIdx+4:]
	if dotIdx := strings.Index(target, ". "); dotIdx >= 0 {
		target = target[:dotIdx]
	}
	if endIdx := strings.IndexAny(target, "\r\n<\";"); endIdx >= 0 {
		target = target[:endIdx]
	}
	target = strings.TrimSuffix(target, ".")
	target = strings.TrimSpace(target)
	if target == "" || target == "None" || len(target) > 50 {
		return ""
	}
	return modelName(target)
}

// countAgySteps parses one JSONL log (overview.txt or transcript.jsonl — same
// schema) and folds step counts, per-day buckets, and session bounds into res.
// It returns the number of steps counted in this file.
func countAgySteps(path string, res *sessionResult) int {
	f, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			res.warnings = append(res.warnings, formatScanWarning(path, err))
		}
		return 0
	}
	defer f.Close()

	counted := 0
	curModel := res.activeModel
	scanner := newJSONLScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec struct {
			CreatedAt string `json:"created_at"`
			Source    string `json:"source"`
			Type      string `json:"type"`
			Content   string `json:"content"`
			ToolCalls []struct {
				Args struct {
					Cwd string `json:"Cwd"`
				} `json:"args"`
			} `json:"tool_calls"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}

		// Fallback project detection from transcript content if not resolved from conversation_summaries.db
		if res.project == "" {
			for _, tc := range rec.ToolCalls {
				cwd := strings.Trim(strings.TrimSpace(tc.Args.Cwd), "\"")
				if cwd != "" && filepath.IsAbs(cwd) {
					res.project = thermal.ProjectKey(cwd)
					break
				}
			}
			if res.project == "" && strings.Contains(rec.Content, " -> ") {
				idx := strings.Index(rec.Content, " -> ")
				start := strings.LastIndex(rec.Content[:idx], "\n")
				cand := strings.TrimSpace(rec.Content[start+1 : idx])
				cand = strings.TrimPrefix(cand, "file://")
				if filepath.IsAbs(cand) {
					res.project = thermal.ProjectKey(cand)
				}
			}
		}

		// Detect genuine Antigravity user settings changes
		if (rec.Source == "USER_EXPLICIT" || rec.Type == "USER_INPUT" || rec.Type == "SYSTEM") &&
			(strings.Contains(rec.Content, "<USER_SETTINGS_CHANGE>") || strings.Contains(rec.Content, "changed setting") || strings.Contains(rec.Content, "changed from")) &&
			strings.Contains(rec.Content, "Model Selection") {
			if m := parseAgyModelSetting(rec.Content); m != "" {
				curModel = m
				res.activeModel = m
			}
		}

		if rec.CreatedAt == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, rec.CreatedAt)
		if err != nil {
			continue
		}
		day := thermal.LocalDay(t)
		res.dayCounts[day]++
		res.steps++
		counted++
		if curModel != "" {
			res.modelCounts[curModel]++
			if res.dayModels[day] == nil {
				res.dayModels[day] = make(map[string]int)
			}
			res.dayModels[day][curModel]++
		}
		if res.firstTs.IsZero() || t.Before(res.firstTs) {
			res.firstTs = t
		}
		if t.After(res.lastTs) {
			res.lastTs = t
		}
	}
	if err := scanner.Err(); err != nil {
		res.warnings = append(res.warnings, formatScanWarning(path, err))
	}

	// Backfill steps in this session if model was declared after initial steps
	if res.activeModel != "" {
		for day, count := range res.dayCounts {
			dayAssigned := 0
			if res.dayModels[day] != nil {
				for _, c := range res.dayModels[day] {
					dayAssigned += c
				}
			}
			unassigned := count - dayAssigned
			if unassigned > 0 {
				res.modelCounts[res.activeModel] += int64(unassigned)
				if res.dayModels[day] == nil {
					res.dayModels[day] = make(map[string]int)
				}
				res.dayModels[day][res.activeModel] += unassigned
			}
		}
	}

	return counted
}

func probeAgyModels(path string, res *sessionResult) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := newJSONLScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec struct {
			Source  string `json:"source"`
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if (rec.Source == "USER_EXPLICIT" || rec.Type == "USER_INPUT" || rec.Type == "SYSTEM") &&
			(strings.Contains(rec.Content, "<USER_SETTINGS_CHANGE>") || strings.Contains(rec.Content, "changed setting") || strings.Contains(rec.Content, "changed from")) &&
			strings.Contains(rec.Content, "Model Selection") {
			if m := parseAgyModelSetting(rec.Content); m != "" {
				res.activeModel = m
			}
		}
	}
	if res.activeModel != "" {
		for day, count := range res.dayCounts {
			res.modelCounts[res.activeModel] += int64(count)
			if res.dayModels[day] == nil {
				res.dayModels[day] = make(map[string]int)
			}
			res.dayModels[day][res.activeModel] += count
		}
	}
}

// LoadAgyData reads Google Antigravity session data in parallel using a bounded worker pool.
func LoadAgyData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	brainDir := ResolveAgyBrainDir(dataDir)
	entries, err := os.ReadDir(brainDir)
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot read %s: %w", brainDir, err)
	}

	results := make(chan sessionResult, len(entries))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8) // max 8 concurrent workers

	workspaces := loadAgyWorkspaces(dataDir)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		entryName := entry.Name()
		sessionDir := filepath.Join(brainDir, entryName)
		project := workspaces[entryName]
		sem <- struct{}{}
		wg.Add(1)

		go func(sDir, proj string) {
			defer wg.Done()
			defer func() { <-sem }()

			res := sessionResult{
				project:     proj,
				dayCounts:   make(map[string]int),
				dayModels:   make(map[string]map[string]int),
				modelCounts: make(map[string]int64),
			}

			logsDir := filepath.Join(sDir, ".system_generated", "logs")

			// 1. Process overview.txt (legacy) and transcript.jsonl (current).
			// If overview exists, it supplies steps; then probe transcript.jsonl for models.
			// In modern sessions, transcript.jsonl parses both in a single pass.
			if countAgySteps(filepath.Join(logsDir, "overview.txt"), &res) == 0 {
				countAgySteps(filepath.Join(logsDir, "transcript.jsonl"), &res)
			} else {
				probeAgyModels(filepath.Join(logsDir, "transcript.jsonl"), &res)
			}

			if res.steps > 0 || len(res.warnings) > 0 {
				results <- res
			}
		}(sessionDir, project)
	}

	wg.Wait()
	close(results)

	byDay := make(map[string]int)
	type projectDayKey struct {
		day     string
		project string
	}
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
	modelCounts := make(map[string]int64)
	var summary thermal.Summary
	summary.Tool = "Agy"

	for res := range results {
		summary.Warnings = append(summary.Warnings, res.warnings...)
		if res.steps == 0 {
			continue
		}
		summary.Sessions++
		summary.LifetimeTokens += int64(res.steps)

		durationMs := res.lastTs.Sub(res.firstTs).Milliseconds()
		if durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = durationMs
		}

		for day, count := range res.dayCounts {
			byDay[day] += count
		}
		for model, count := range res.modelCounts {
			modelCounts[model] += count
		}
		if res.project != "" {
			for day, count := range res.dayCounts {
				key := projectDayKey{day: day, project: res.project}
				pd := byProjectDay[key]
				if pd == nil {
					pd = &thermal.ProjectDay{
						Project: res.project,
						Day:     day,
						Tool:    "Agy",
					}
					byProjectDay[key] = pd
				}
				pd.Turns += count
			}
		}
	}

	var daily []thermal.DailyRow
	for day, count := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: int64(count),
			Turns:  count,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	var projectDays []thermal.ProjectDay
	for _, pd := range byProjectDay {
		projectDays = append(projectDays, *pd)
	}
	sort.Slice(projectDays, func(i, j int) bool {
		if projectDays[i].Day != projectDays[j].Day {
			return projectDays[i].Day < projectDays[j].Day
		}
		return projectDays[i].Project < projectDays[j].Project
	})

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}

	sort.Strings(summary.Warnings)
	return summary, daily, projectDays, nil
}
