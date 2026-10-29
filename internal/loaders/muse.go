// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite" // register sqlite driver
)

type museFacetDetail struct {
	Model string `json:"model,omitempty"`
}

type museFacet struct {
	Detail *museFacetDetail `json:"detail,omitempty"`
}

type museEventDetails struct {
	Facets []museFacet `json:"facets,omitempty"`
}

type museEvent struct {
	Kind         string            `json:"kind"`
	Model        string            `json:"model,omitempty"`
	DurationMs   int64             `json:"duration_ms,omitempty"`
	FinishReason string            `json:"finish_reason,omitempty"`
	Details      *museEventDetails `json:"details,omitempty"`
	Usage        struct {
		InputTokens      int64 `json:"input_tokens"`
		OutputTokens     int64 `json:"output_tokens"`
		ReasoningTokens  int64 `json:"reasoning_tokens"`
		CacheReadTokens  int64 `json:"cache_read_tokens"`
		CacheWriteTokens int64 `json:"cache_write_tokens"`
		CachedTokens     int64 `json:"cached_tokens"`
	} `json:"usage"`
}

type museRecord struct {
	WorkspaceRoot string `json:"workspace_root,omitempty"`
	Cwd           any    `json:"cwd,omitempty"`
	Model         string `json:"model,omitempty"`
}

type musePayload struct {
	Kind   string      `json:"kind,omitempty"`
	Event  *museEvent  `json:"event,omitempty"`
	Record *museRecord `json:"record,omitempty"`
}

type museStream struct {
	Kind string `json:"kind,omitempty"`
	ID   string `json:"id,omitempty"`
}

type museLogLine struct {
	RecordedAt  int64       `json:"recorded_at"`
	Stream      museStream  `json:"stream"`
	PayloadType string      `json:"payload_type"`
	Payload     musePayload `json:"payload"`
}

type museTurn struct {
	day        string
	rawTs      int64
	model      string
	input      int64
	output     int64
	reasoning  int64
	cacheRead  int64
	cacheWrite int64
	tokens     int64
	durationMs int64
}

type museFileResult struct {
	sessionID    string
	dateFromPath string
	project      string
	firstTs      int64
	lastTs       int64
	turns        []museTurn
	warning      string
}

type museDayAgg struct {
	tokens    int64
	input     int64
	output    int64
	reasoning int64
	cache     int64
	turns     int
	models    map[string]thermal.ModelTokens
}

type museProjDayAgg struct {
	tokens     int64
	input      int64
	output     int64
	reasoning  int64
	cacheRead  int64
	cacheWrite int64
	turns      int
	models     map[string]thermal.ModelTokens
}

func extractMuseDateFromPath(sessionsDir, path string) string {
	rel, err := filepath.Rel(sessionsDir, path)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) >= 3 && len(parts[0]) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2 {
		return parts[0] + "-" + parts[1] + "-" + parts[2]
	}
	return ""
}

func extractMuseSessionID(sessionsDir, path string) string {
	rel, err := filepath.Rel(sessionsDir, path)
	if err == nil {
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) >= 4 {
			return parts[3]
		}
		if len(parts) >= 2 {
			return parts[0]
		}
	}
	dir := filepath.Dir(path)
	base := filepath.Base(dir)
	if base == "subagent" || base == "approval-review" {
		return filepath.Base(filepath.Dir(dir))
	}
	return base
}

func parseCwd(raw any) string {
	if s, ok := raw.(string); ok {
		return s
	}
	if m, ok := raw.(map[string]any); ok {
		if text, ok := m["text"].(string); ok {
			return text
		}
	}
	return ""
}

func scanMuseJSONL(sessionsDir, path string) museFileResult {
	res := museFileResult{
		sessionID:    extractMuseSessionID(sessionsDir, path),
		dateFromPath: extractMuseDateFromPath(sessionsDir, path),
	}

	f, err := os.Open(path)
	if err != nil {
		res.warning = formatScanWarning(path, err)
		return res
	}
	defer f.Close()

	inApprovalReview := filepath.Base(filepath.Dir(path)) == "approval-review"
	var activeModel string

	scanner := newJSONLScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry museLogLine
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		// Approval-review sandboxes emit a deterministic clock starting at
		// 1780531500000000 + sequence. Ignore synthetic timestamps for session
		// duration and resolve their calendar day from the parent session or path.
		isSynthetic := (entry.Stream.Kind == "approval.review.runtime" || inApprovalReview) &&
			entry.RecordedAt >= 1780531500000000 && entry.RecordedAt < 1780531501000000

		if entry.RecordedAt > 0 && !isSynthetic {
			if res.firstTs == 0 || entry.RecordedAt < res.firstTs {
				res.firstTs = entry.RecordedAt
			}
			if entry.RecordedAt > res.lastTs {
				res.lastTs = entry.RecordedAt
			}
		}

		if rec := entry.Payload.Record; rec != nil {
			if rec.WorkspaceRoot != "" && res.project == "" {
				res.project = rec.WorkspaceRoot
			} else if rec.Cwd != nil && res.project == "" {
				if cwd := parseCwd(rec.Cwd); cwd != "" {
					res.project = cwd
				}
			}
			if rec.Model != "" {
				activeModel = rec.Model
			}
		}

		ev := entry.Payload.Event
		if ev != nil {
			if ev.Details != nil {
				for _, facet := range ev.Details.Facets {
					if facet.Detail != nil && facet.Detail.Model != "" {
						activeModel = facet.Detail.Model
					}
				}
			}
			if ev.Model != "" {
				activeModel = ev.Model
			}
		}

		if ev != nil && ev.Kind == "model_completed" {
			u := ev.Usage
			netInput := u.InputTokens - (u.CacheReadTokens + u.CacheWriteTokens)
			if netInput < 0 {
				netInput = 0
			}
			netOutput := u.OutputTokens - u.ReasoningTokens
			if netOutput < 0 {
				netOutput = 0
			}
			turnTokens := netInput + netOutput + u.ReasoningTokens + u.CacheReadTokens + u.CacheWriteTokens
			if turnTokens == 0 && u.InputTokens+u.OutputTokens > 0 {
				turnTokens = u.InputTokens + u.OutputTokens
				netInput = turnTokens
			}

			model := ev.Model
			if model == "" {
				model = activeModel
			}
			if model == "" {
				model = "muse-spark-1.3-contributor"
			}

			var day string
			if !isSynthetic && entry.RecordedAt > 0 {
				day = thermal.UnixDay(entry.RecordedAt / 1000000)
			}
			res.turns = append(res.turns, museTurn{
				day:        day,
				rawTs:      entry.RecordedAt,
				model:      modelName(model),
				input:      netInput,
				output:     netOutput,
				reasoning:  u.ReasoningTokens,
				cacheRead:  u.CacheReadTokens,
				cacheWrite: u.CacheWriteTokens,
				tokens:     turnTokens,
				durationMs: ev.DurationMs,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		res.warning = formatScanWarning(path, err)
	}
	return res
}

// LoadMuseData reads Meta Muse sessions from ~/.local/share/muse/sessions/
// logs and folds in any unindexed legacy sessions from session-index.db.
// It parses completed model calls (input, output, reasoning, cache tokens,
// model names, durations, and local model-catalog pricing), attributes projects
// through ProjectKey, and enforces disjoint token accounting.
func LoadMuseData(path string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	var summary thermal.Summary
	summary.Tool = "Muse"
	summary.ModelBreakdown = make(map[string]int64)

	var dbPath, dataDir string
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		dbPath = path
		dataDir = filepath.Dir(path)
	} else {
		dataDir = path
		dbPath = filepath.Join(path, "session-index.db")
	}

	sessionsDir := filepath.Join(dataDir, "sessions")
	var files []string
	if _, err := os.Stat(sessionsDir); err == nil {
		_ = filepath.WalkDir(sessionsDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(d.Name(), ".jsonl") {
				files = append(files, p)
			}
			return nil
		})
	}
	sort.Strings(files)

	byDay := make(map[string]*museDayAgg)
	byProjDay := make(map[string]*museProjDayAgg) // key: project + "\x00" + day

	scannedSessions := make(map[string]bool)
	sessionTimes := make(map[string][2]int64)
	sessionProjects := make(map[string]string)

	if len(files) > 0 {
		results := make(chan museFileResult, len(files))
		const maxWorkers = 16
		workers := len(files)
		if workers > maxWorkers {
			workers = maxWorkers
		}
		sem := make(chan struct{}, workers)
		var wg sync.WaitGroup

		for _, file := range files {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				results <- scanMuseJSONL(sessionsDir, p)
			}(file)
		}

		wg.Wait()
		close(results)

		var fileResults []museFileResult
		for r := range results {
			if r.warning != "" {
				summary.Warnings = append(summary.Warnings, r.warning)
			}
			if r.sessionID != "" {
				scannedSessions[r.sessionID] = true
				if r.firstTs > 0 {
					st, exists := sessionTimes[r.sessionID]
					if !exists {
						sessionTimes[r.sessionID] = [2]int64{r.firstTs, r.lastTs}
					} else {
						if r.firstTs < st[0] {
							st[0] = r.firstTs
						}
						if r.lastTs > st[1] {
							st[1] = r.lastTs
						}
						sessionTimes[r.sessionID] = st
					}
				}
				if r.project != "" && sessionProjects[r.sessionID] == "" {
					sessionProjects[r.sessionID] = r.project
				}
			}
			fileResults = append(fileResults, r)
		}

		// Merge known projects and unindexed sessions from session-index.db
		if fi, err := os.Stat(dbPath); err == nil && !fi.IsDir() {
			if err := loadMuseIndexDB(dbPath, scannedSessions, sessionProjects, &summary, byDay, byProjDay); err != nil {
				return thermal.Summary{}, nil, nil, err
			}
		}

		// Aggregate turns
		for _, r := range fileResults {
			proj := r.project
			if proj == "" {
				proj = sessionProjects[r.sessionID]
			}
			canonProj := ""
			if proj != "" {
				canonProj = thermal.ProjectKey(proj)
			}

			for _, t := range r.turns {
				day := t.day
				if day == "" {
					if st, ok := sessionTimes[r.sessionID]; ok && st[0] > 0 {
						day = thermal.UnixDay(st[0] / 1000000)
					} else if r.dateFromPath != "" {
						day = r.dateFromPath
					} else {
						day = thermal.UnixDay(t.rawTs / 1000000)
					}
				}

				summary.ModelBreakdown[t.model]++

				da, exists := byDay[day]
				if !exists {
					da = &museDayAgg{models: make(map[string]thermal.ModelTokens)}
					byDay[day] = da
				}
				da.tokens += t.tokens
				da.input += t.input
				da.output += t.output
				da.reasoning += t.reasoning
				da.cache += (t.cacheRead + t.cacheWrite)
				da.turns++

				mt := da.models[t.model]
				mt.Input += t.input
				mt.Output += t.output
				mt.Reasoning += t.reasoning
				mt.CacheRead += t.cacheRead
				mt.CacheWrite += t.cacheWrite
				da.models[t.model] = mt

				if canonProj != "" {
					pKey := canonProj + "\x00" + day
					pda, pExists := byProjDay[pKey]
					if !pExists {
						pda = &museProjDayAgg{models: make(map[string]thermal.ModelTokens)}
						byProjDay[pKey] = pda
					}
					pda.tokens += t.tokens
					pda.input += t.input
					pda.output += t.output
					pda.reasoning += t.reasoning
					pda.cacheRead += t.cacheRead
					pda.cacheWrite += t.cacheWrite
					pda.turns++

					pmt := pda.models[t.model]
					pmt.Input += t.input
					pmt.Output += t.output
					pmt.Reasoning += t.reasoning
					pmt.CacheRead += t.cacheRead
					pmt.CacheWrite += t.cacheWrite
					pda.models[t.model] = pmt
				}
			}
		}

		for _, st := range sessionTimes {
			if st[1] >= st[0] {
				durMs := (st[1] - st[0]) / 1000
				if durMs > summary.LongestSessionMs {
					summary.LongestSessionMs = durMs
				}
			}
		}
	} else {
		// No JSONL files found; pure database mode
		if fi, err := os.Stat(dbPath); err == nil && !fi.IsDir() {
			if err := loadMuseIndexDB(dbPath, scannedSessions, sessionProjects, &summary, byDay, byProjDay); err != nil {
				return thermal.Summary{}, nil, nil, err
			}
		}
	}

	summary.Sessions = len(scannedSessions)

	var daily []thermal.DailyRow
	for day, da := range byDay {
		if da.tokens > 0 || da.turns > 0 {
			daily = append(daily, thermal.DailyRow{
				Day:       day,
				Tokens:    da.tokens,
				Input:     da.input,
				Output:    da.output,
				Reasoning: da.reasoning,
				Cache:     da.cache,
				Turns:     da.turns,
				Models:    da.models,
			})
			summary.LifetimeTokens += da.tokens
			summary.InputTokens += da.input
			summary.OutputTokens += da.output
			summary.ReasoningTokens += da.reasoning
			summary.CacheTokens += da.cache
		}
	}
	sort.Slice(daily, func(i, j int) bool {
		return daily[i].Day < daily[j].Day
	})

	var projects []thermal.ProjectDay
	for pKey, pda := range byProjDay {
		parts := strings.Split(pKey, "\x00")
		proj, day := parts[0], parts[1]
		projects = append(projects, thermal.ProjectDay{
			Tool:       "Muse",
			Project:    proj,
			Day:        day,
			Tokens:     pda.tokens,
			Input:      pda.input,
			Output:     pda.output,
			Reasoning:  pda.reasoning,
			CacheRead:  pda.cacheRead,
			CacheWrite: pda.cacheWrite,
			Turns:      pda.turns,
			Models:     pda.models,
		})
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Day != projects[j].Day {
			return projects[i].Day < projects[j].Day
		}
		return projects[i].Project < projects[j].Project
	})

	return summary, daily, projects, nil
}

func loadMuseIndexDB(dbPath string, scannedSessions map[string]bool, sessionProjects map[string]string, summary *thermal.Summary, byDay map[string]*museDayAgg, byProjDay map[string]*museProjDayAgg) error {
	db, err := openBoundedStore(dbPath, 30000000000)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := rejectSubstitutableObjects(db, "sessions"); err != nil {
		return err
	}

	hasWorkspace := false
	if tiRows, err := queryStore(db, `PRAGMA table_info(sessions)`); err == nil {
		for tiRows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt any
			if err := tiRows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil {
				if name == "workspace_root" {
					hasWorkspace = true
				}
			}
		}
		_ = tiRows.Err()
		tiRows.Close()
	}

	workspaceCol := "'' AS workspace_root"
	if hasWorkspace {
		workspaceCol = "COALESCE(workspace_root, '')"
	}

	query := `
		SELECT
			session_id,
			COALESCE(model_id, ''),
			` + workspaceCol + `,
			COALESCE(prompt_count, 0),
			COALESCE(created_at_us, 0),
			COALESCE(updated_at_us, 0)
		FROM sessions
	`
	rows, err := queryStore(db, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID, modelID, workspaceRoot string
		var promptCount, createdAtUs, updatedAtUs int64
		if err := rows.Scan(&sessionID, &modelID, &workspaceRoot, &promptCount, &createdAtUs, &updatedAtUs); err != nil {
			return err
		}

		if workspaceRoot != "" && sessionProjects[sessionID] == "" {
			sessionProjects[sessionID] = workspaceRoot
		}

		if scannedSessions[sessionID] {
			continue // Already accounted for in JSONL
		}
		scannedSessions[sessionID] = true

		if modelID != "" {
			summary.ModelBreakdown[modelName(modelID)]++
		}

		durMs := (updatedAtUs - createdAtUs) / 1000
		if durMs > summary.LongestSessionMs {
			summary.LongestSessionMs = durMs
		}

		if promptCount > 0 {
			ts := updatedAtUs
			if ts == 0 {
				ts = createdAtUs
			}
			day := thermal.UnixDay(ts / 1000000)
			da, exists := byDay[day]
			if !exists {
				da = &museDayAgg{models: make(map[string]thermal.ModelTokens)}
				byDay[day] = da
			}
			da.tokens += promptCount
			da.turns += int(promptCount)

			if workspaceRoot != "" {
				canonProj := thermal.ProjectKey(workspaceRoot)
				pKey := canonProj + "\x00" + day
				pda, pExists := byProjDay[pKey]
				if !pExists {
					pda = &museProjDayAgg{models: make(map[string]thermal.ModelTokens)}
					byProjDay[pKey] = pda
				}
				pda.tokens += promptCount
				pda.turns += int(promptCount)
			}
		}
	}
	return rows.Err()
}
