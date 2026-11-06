// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"bytes"
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
	"github.com/klauspost/compress/zstd"
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
	SessionStats    *dshSessionStatsRow    `json:"sessionStats"`
	ModelSelection  *dshModelSelectionRow  `json:"modelSelection"`
	TokenUsage      *dshTokenUsageRow      `json:"tokenUsage"`
	TurnBoundary    *dshTurnBoundaryRow    `json:"turnBoundary"`
	WorkingActivity *dshWorkingActivityRow `json:"workingActivity"`
	TimeContext     *dshTimeContextRow     `json:"timeContext"`
}

type dshWorkingActivityRow struct {
	Val *struct {
		UpdatedAt any `json:"updatedAt"`
	} `json:"val"`
}

type dshTimeContextRow struct {
	Val *struct {
		LastMessageTime any `json:"lastMessageTime"`
	} `json:"val"`
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

type dshTurnBoundaryRow struct {
	Val *dshTurnBoundaryVal `json:"val"`
}

type dshTurnBoundaryVal struct {
	LastTurn int `json:"lastTurn"`
}

type dshTokenTotals struct {
	UncachedInputTokens int64 `json:"uncachedInputTokens"`
	InputTokens         int64 `json:"inputTokens"`
	OutputTokens        int64 `json:"outputTokens"`
	ReasoningTokens     int64 `json:"reasoningTokens"`
	CacheReadTokens     int64 `json:"cacheReadTokens"`
	CacheWriteTokens    int64 `json:"cacheWriteTokens"`
	TotalTokens         int64 `json:"totalTokens"`
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
	id               string
	day              string
	endDay           string
	project          string
	model            string
	input            int64
	output           int64
	cacheRead        int64
	cacheWrite       int64
	reasoning        int64
	total            int64
	turns            int
	durationMs       int64
	active           bool
	skipSessionCount bool
	warning          string
}

func extractDshSession(id string, rec dshRecord) dshSessionResult {
	t, ok := parseDshTimestamp(rec.Identity.CreatedAt)
	if !ok || t.IsZero() {
		return dshSessionResult{id: id}
	}
	day := thermal.LocalDay(t.Local())

	var endT time.Time
	if rec.Rows.WorkingActivity != nil && rec.Rows.WorkingActivity.Val != nil {
		if ut, ok := parseDshTimestamp(rec.Rows.WorkingActivity.Val.UpdatedAt); ok && ut.After(endT) {
			endT = ut
		}
	}
	if rec.Rows.TimeContext != nil && rec.Rows.TimeContext.Val != nil {
		if mt, ok := parseDshTimestamp(rec.Rows.TimeContext.Val.LastMessageTime); ok && mt.After(endT) {
			endT = mt
		}
	}
	var endDay string
	if !endT.IsZero() {
		endDay = thermal.LocalDay(endT.Local())
	}

	var turns int
	var durationMs int64
	if rec.Rows.SessionStats != nil && rec.Rows.SessionStats.Val != nil {
		turns = rec.Rows.SessionStats.Val.Turns
		durationMs = rec.Rows.SessionStats.Val.LLMMs + rec.Rows.SessionStats.Val.ToolMs
	}
	if durationMs == 0 && endT.After(t) {
		durationMs = endT.Sub(t).Milliseconds()
	}
	// Version 7 stopped writing sessionStats. turnBoundary.lastTurn carries the
	// turn index the session UI shows, so take the larger value when both exist.
	if tb := rec.Rows.TurnBoundary; tb != nil && tb.Val != nil && tb.Val.LastTurn > turns {
		turns = tb.Val.LastTurn
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
		endDay:     endDay,
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

const (
	dshModelLogLineCap     = 2000
	dshModelLogWorkerLimit = 16
)

// dshLogEvent is the subset of a session event needed to recover the model.
// DeepSeek records it twice: flat on request/context and nested on
// request/header.
type dshLogEvent struct {
	Type string `json:"type"`
	Data struct {
		Model  string `json:"model"`
		Header struct {
			Config struct {
				Model string `json:"model"`
			} `json:"config"`
		} `json:"header"`
	} `json:"data"`
}

// dshSessionLogPath picks the event log inside one session directory.
func dshSessionLogPath(dir string) string {
	preferred := filepath.Join(dir, "session.v4.jsonl.zstd")
	if fi, err := os.Stat(preferred); err == nil && fi.Mode().IsRegular() {
		return preferred
	}
	matches, err := filepath.Glob(filepath.Join(dir, "session*.jsonl.zstd"))
	if err != nil || len(matches) == 0 {
		return ""
	}
	sort.Strings(matches)
	return matches[0]
}

// dshModelLogIndex maps session ids to their event log. DeepSeek stores logs
// at sessions/<flattened-cwd>/<session-id>/session*.jsonl.zstd.
func dshModelLogIndex(sessionsRoot string) map[string]string {
	index := make(map[string]string)
	workspaces, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return index
	}
	for _, ws := range workspaces {
		if !ws.IsDir() {
			continue
		}
		wsDir := filepath.Join(sessionsRoot, ws.Name())
		sessions, err := os.ReadDir(wsDir)
		if err != nil {
			continue
		}
		for _, s := range sessions {
			if !s.IsDir() {
				continue
			}
			if _, exists := index[s.Name()]; exists {
				continue
			}
			if p := dshSessionLogPath(filepath.Join(wsDir, s.Name())); p != "" {
				index[s.Name()] = p
			}
		}
	}
	return index
}

// dshModelFromLog head-decodes a session event log and returns the first model
// recorded in a request event. It stops at the first assistant message so a log
// without request metadata cannot stream the whole file.
func dshModelFromLog(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec, err := zstd.NewReader(f)
	if err != nil {
		return "", err
	}
	defer dec.Close()

	scanner := newJSONLScanner(dec)
	for lines := 0; scanner.Scan() && lines < dshModelLogLineCap; lines++ {
		line := scanner.Bytes()
		if bytes.Contains(line, []byte(`"assistant/message"`)) {
			break
		}
		if !bytes.Contains(line, []byte(`"request/context"`)) &&
			!bytes.Contains(line, []byte(`"request/header"`)) {
			continue
		}
		var event dshLogEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		if event.Data.Model != "" {
			return modelName(event.Data.Model), nil
		}
		if event.Data.Header.Config.Model != "" {
			return modelName(event.Data.Header.Config.Model), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

type dshAssistantMessageEvent struct {
	Time any `json:"time"`
	Data struct {
		Turn    int `json:"turn"`
		Message struct {
			Source struct {
				Model string `json:"model"`
			} `json:"source"`
		} `json:"message"`
		Usage *dshTokenTotals `json:"usage"`
	} `json:"data"`
}

// dshDailySlicesFromLog streams a multi-day session event log and partitions
// its token and turn usage across the calendar days on which assistant messages
// occurred. Any positive remainder from the projection totals is reconciled
// onto the session's final active day so lifetime totals stay exact.
func dshDailySlicesFromLog(path string, base dshSessionResult) ([]dshSessionResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec, err := zstd.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	type dayBucket struct {
		input, output, cacheRead, cacheWrite, reasoning int64
		model                                           string
		turns                                           map[int]struct{}
		msgCount                                        int
	}
	byDay := make(map[string]*dayBucket)
	var dayOrder []string
	defaultModel := base.model

	scanner := newJSONLScanner(dec)
	for scanner.Scan() {
		line := scanner.Bytes()
		if defaultModel == "" && (bytes.Contains(line, []byte(`"request/context"`)) ||
			bytes.Contains(line, []byte(`"request/header"`))) {
			var event dshLogEvent
			if err := json.Unmarshal(line, &event); err == nil {
				if event.Data.Model != "" {
					defaultModel = modelName(event.Data.Model)
				} else if event.Data.Header.Config.Model != "" {
					defaultModel = modelName(event.Data.Header.Config.Model)
				}
			}
			continue
		}
		if !bytes.Contains(line, []byte(`"assistant/message"`)) {
			continue
		}
		var ev dshAssistantMessageEvent
		if err := json.Unmarshal(line, &ev); err != nil || ev.Data.Usage == nil {
			continue
		}
		evDay := base.day
		if t, ok := parseDshTimestamp(ev.Time); ok && !t.IsZero() {
			evDay = thermal.LocalDay(t.Local())
		}
		m := modelName(ev.Data.Message.Source.Model)
		if m == "" {
			m = defaultModel
		}

		u := ev.Data.Usage
		cr := u.CacheReadTokens
		cw := u.CacheWriteTokens
		reas := u.ReasoningTokens
		out := u.OutputTokens
		if reas > 0 && out > reas {
			out -= reas
		}
		var in int64
		if u.UncachedInputTokens > 0 {
			in = u.UncachedInputTokens
		} else if u.InputTokens > 0 {
			if u.TotalTokens > 0 && u.InputTokens+u.OutputTokens+cr+cw == u.TotalTokens {
				in = u.InputTokens
			} else if u.InputTokens >= cr {
				in = u.InputTokens - cr
			} else {
				in = u.InputTokens
			}
		}

		b := byDay[evDay]
		if b == nil {
			b = &dayBucket{turns: make(map[int]struct{})}
			byDay[evDay] = b
			dayOrder = append(dayOrder, evDay)
		}
		b.input += in
		b.output += out
		b.cacheRead += cr
		b.cacheWrite += cw
		b.reasoning += reas
		if m != "" {
			b.model = m
		}
		if ev.Data.Turn > 0 {
			b.turns[ev.Data.Turn] = struct{}{}
		} else {
			b.msgCount++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(dayOrder) == 0 {
		return nil, nil
	}

	sort.Strings(dayOrder)
	var sumIn, sumOut, sumCR, sumCW, sumReas int64
	for _, d := range dayOrder {
		b := byDay[d]
		sumIn += b.input
		sumOut += b.output
		sumCR += b.cacheRead
		sumCW += b.cacheWrite
		sumReas += b.reasoning
	}

	// Reconcile any positive projection remainder onto the latest day bucket.
	lastBucket := byDay[dayOrder[len(dayOrder)-1]]
	if base.input > sumIn {
		lastBucket.input += base.input - sumIn
	}
	if base.output > sumOut {
		lastBucket.output += base.output - sumOut
	}
	if base.cacheRead > sumCR {
		lastBucket.cacheRead += base.cacheRead - sumCR
	}
	if base.cacheWrite > sumCW {
		lastBucket.cacheWrite += base.cacheWrite - sumCW
	}
	if base.reasoning > sumReas {
		lastBucket.reasoning += base.reasoning - sumReas
	}

	slices := make([]dshSessionResult, 0, len(dayOrder))
	for i, d := range dayOrder {
		b := byDay[d]
		tot := b.input + b.output + b.cacheRead + b.cacheWrite + b.reasoning
		turns := len(b.turns) + b.msgCount
		if turns <= 0 && tot > 0 {
			turns = 1
		}
		if tot == 0 && turns == 0 {
			continue
		}
		m := b.model
		if m == "" {
			m = defaultModel
		}
		dur := int64(0)
		if i == 0 {
			dur = base.durationMs
		}
		slices = append(slices, dshSessionResult{
			id:               base.id,
			day:              d,
			endDay:           base.endDay,
			project:          base.project,
			model:            m,
			input:            b.input,
			output:           b.output,
			cacheRead:        b.cacheRead,
			cacheWrite:       b.cacheWrite,
			reasoning:        b.reasoning,
			total:            tot,
			turns:            turns,
			durationMs:       dur,
			active:           true,
			skipSessionCount: len(slices) > 0,
		})
	}
	return slices, nil
}

// fillDshModels attributes a model to sessions whose projection carries none
// and splits multi-day sessions across their active calendar days using the
// zstd event log DeepSeek writes beside the projection cache.
func fillDshModels(sessionsRoot string, results []dshSessionResult) ([]dshSessionResult, []string) {
	var pending []int
	for i := range results {
		r := results[i]
		if !r.active || r.total <= 0 {
			continue
		}
		multiDay := r.endDay != "" && r.endDay != r.day
		if r.model == "" || multiDay {
			pending = append(pending, i)
		}
	}
	if len(pending) == 0 {
		return results, nil
	}
	index := dshModelLogIndex(sessionsRoot)
	if len(index) == 0 {
		return results, nil
	}

	var mu sync.Mutex
	var warnings []string
	splitSlices := make(map[int][]dshSessionResult)
	var wg sync.WaitGroup
	sem := make(chan struct{}, dshModelLogWorkerLimit)

	for _, i := range pending {
		logPath, ok := index[results[i].id]
		if !ok {
			continue
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-sem }()

			r := results[idx]
			if r.endDay != "" && r.endDay != r.day {
				slices, err := dshSplitFromLogOrModel(path, r)
				if err != nil {
					mu.Lock()
					warnings = append(warnings, formatScanWarning(path, err))
					mu.Unlock()
					return
				}
				if len(slices) > 0 {
					mu.Lock()
					splitSlices[idx] = slices
					mu.Unlock()
					return
				}
			}

			model, err := dshModelFromLog(path)
			if err != nil {
				mu.Lock()
				warnings = append(warnings, formatScanWarning(path, err))
				mu.Unlock()
				return
			}
			if model != "" {
				results[idx].model = model
			}
		}(i, logPath)
	}
	wg.Wait()

	if len(splitSlices) == 0 {
		return results, warnings
	}
	expanded := make([]dshSessionResult, 0, len(results)+len(splitSlices))
	for i, r := range results {
		if slices, ok := splitSlices[i]; ok {
			expanded = append(expanded, slices...)
		} else {
			expanded = append(expanded, r)
		}
	}
	return expanded, warnings
}

func dshSplitFromLogOrModel(path string, base dshSessionResult) ([]dshSessionResult, error) {
	slices, err := dshDailySlicesFromLog(path, base)
	if err != nil {
		return nil, err
	}
	return slices, nil
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

	if info.IsDir() {
		// Version 7 projections no longer carry the model; the zstd event log
		// does, and multi-day sessions split usage by message timestamp.
		var logWarnings []string
		sessionResults, logWarnings = fillDshModels(filepath.Join(dataDir, "sessions"), sessionResults)
		summary.Warnings = append(summary.Warnings, logWarnings...)
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
		if !res.skipSessionCount {
			summary.Sessions++
			if res.durationMs > summary.LongestSessionMs {
				summary.LongestSessionMs = res.durationMs
			}
			if res.model != "" {
				modelCounts[res.model]++
			}
		}
		summary.InputTokens += res.input
		summary.OutputTokens += res.output
		summary.CacheTokens += res.cacheRead + res.cacheWrite
		summary.ReasoningTokens += res.reasoning
		summary.LifetimeTokens += res.total

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
