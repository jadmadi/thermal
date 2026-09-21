package loaders

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite" // register sqlite driver
)

// hermesModelUsage captures per-model token breakdown if session_model_usage is present.
type hermesModelUsage struct {
	model      string
	input      int64
	output     int64
	reasoning  int64
	cacheRead  int64
	cacheWrite int64
}

// parseHermesTimestamp parses Unix timestamps (seconds, milliseconds, microseconds)
// or RFC3339 / SQLite date strings from Hermes session records.
func parseHermesTimestamp(v any) (time.Time, bool) {
	switch val := v.(type) {
	case float64:
		return parseHermesEpoch(int64(val), val)
	case int64:
		return parseHermesEpoch(val, float64(val))
	case int:
		return parseHermesEpoch(int64(val), float64(val))
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
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local); err == nil {
			return t, true
		}
		if t, err := time.ParseInLocation("2006-01-02", val, time.Local); err == nil {
			return t, true
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return parseHermesEpoch(int64(f), f)
		}
	case []byte:
		return parseHermesTimestamp(string(val))
	}
	return time.Time{}, false
}

func parseHermesEpoch(n int64, f float64) (time.Time, bool) {
	if n <= 0 {
		return time.Time{}, false
	}
	if n > 1e14 { // microseconds
		return time.UnixMicro(n), true
	}
	if n > 1e11 { // milliseconds
		return time.UnixMilli(n), true
	}
	// seconds (support fractional part if present)
	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	return time.Unix(sec, nsec), true
}

// LoadHermesData reads the Nous Hermes SQLite DB (~/.hermes/state.db).
// Hermes is an open-source, local-first persistent coding agent developed by
// Nous Research. It stores session lifecycle metadata, token usage, tool calls,
// and cost estimates in state.db.
func LoadHermesData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	if dbPath == "" {
		if env := os.Getenv("HERMES_HOME"); env != "" {
			dbPath = filepath.Join(env, "state.db")
		} else {
			dbPath = filepath.Join(thermal.HomeDir(), ".hermes", "state.db")
		}
	}

	if _, err := os.Stat(dbPath); err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("hermes: database not found: %s", dbPath)
	}

	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=268435456")
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 268435456;")

	sessionTable := ""
	var hasSessions bool
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'`).Scan(&hasSessions)
	if hasSessions {
		sessionTable = "sessions"
	} else {
		var hasSession bool
		_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='session'`).Scan(&hasSession)
		if hasSession {
			sessionTable = "session"
		}
	}

	if sessionTable == "" {
		return thermal.Summary{}, nil, nil, fmt.Errorf("hermes: sessions table not found in %s", dbPath)
	}

	// Probe columns dynamically to remain resilient across Hermes schema revisions.
	timeCol := ""
	for _, c := range []string{"started_at", "created_at", "time_created", "timestamp"} {
		if hasColumn(db, sessionTable, c) {
			timeCol = c
			break
		}
	}
	if timeCol == "" {
		return thermal.Summary{}, nil, nil, fmt.Errorf("hermes: no timestamp column found in table %s", sessionTable)
	}

	endTimeCol := ""
	for _, c := range []string{"ended_at", "updated_at", "time_updated"} {
		if hasColumn(db, sessionTable, c) {
			endTimeCol = c
			break
		}
	}

	cwdCol := "''"
	for _, c := range []string{"cwd", "git_repo_root", "directory", "working_directory", "project"} {
		if hasColumn(db, sessionTable, c) {
			cwdCol = "COALESCE(" + c + ", '')"
			break
		}
	}

	modelCol := "''"
	for _, c := range []string{"model", "model_id"} {
		if hasColumn(db, sessionTable, c) {
			modelCol = "COALESCE(" + c + ", '')"
			break
		}
	}

	inputCol := "0"
	for _, c := range []string{"input_tokens", "tokens_input"} {
		if hasColumn(db, sessionTable, c) {
			inputCol = "COALESCE(" + c + ", 0)"
			break
		}
	}

	outputCol := "0"
	for _, c := range []string{"output_tokens", "tokens_output"} {
		if hasColumn(db, sessionTable, c) {
			outputCol = "COALESCE(" + c + ", 0)"
			break
		}
	}

	cacheReadCol := "0"
	for _, c := range []string{"cache_read_tokens", "tokens_cache_read"} {
		if hasColumn(db, sessionTable, c) {
			cacheReadCol = "COALESCE(" + c + ", 0)"
			break
		}
	}

	cacheWriteCol := "0"
	for _, c := range []string{"cache_write_tokens", "tokens_cache_write"} {
		if hasColumn(db, sessionTable, c) {
			cacheWriteCol = "COALESCE(" + c + ", 0)"
			break
		}
	}

	reasoningCol := "0"
	for _, c := range []string{"reasoning_tokens", "tokens_reasoning"} {
		if hasColumn(db, sessionTable, c) {
			reasoningCol = "COALESCE(" + c + ", 0)"
			break
		}
	}

	costCol := "0.0"
	for _, c := range []string{"estimated_cost_usd", "cost", "cost_usd"} {
		if hasColumn(db, sessionTable, c) {
			costCol = "COALESCE(" + c + ", 0.0)"
			break
		}
	}

	turnsCol := "1"
	for _, c := range []string{"message_count", "turn_count", "turns", "tool_call_count"} {
		if hasColumn(db, sessionTable, c) {
			turnsCol = "COALESCE(" + c + ", 1)"
			break
		}
	}

	// Check for optional session_model_usage table.
	modelUsages := make(map[string][]hermesModelUsage)
	var hasSMU bool
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='session_model_usage'`).Scan(&hasSMU)
	if hasSMU {
		smuModelCol := "model"
		if !hasColumn(db, "session_model_usage", smuModelCol) && hasColumn(db, "session_model_usage", "model_id") {
			smuModelCol = "model_id"
		}
		smuInCol := "0"
		if hasColumn(db, "session_model_usage", "input_tokens") {
			smuInCol = "COALESCE(input_tokens, 0)"
		}
		smuOutCol := "0"
		if hasColumn(db, "session_model_usage", "output_tokens") {
			smuOutCol = "COALESCE(output_tokens, 0)"
		}
		smuReasonCol := "0"
		if hasColumn(db, "session_model_usage", "reasoning_tokens") {
			smuReasonCol = "COALESCE(reasoning_tokens, 0)"
		}
		smuCRCol := "0"
		if hasColumn(db, "session_model_usage", "cache_read_tokens") {
			smuCRCol = "COALESCE(cache_read_tokens, 0)"
		}
		smuCWCol := "0"
		if hasColumn(db, "session_model_usage", "cache_write_tokens") {
			smuCWCol = "COALESCE(cache_write_tokens, 0)"
		}

		smuQuery := fmt.Sprintf(`
			SELECT session_id, %s, %s, %s, %s, %s, %s
			FROM session_model_usage
		`, smuModelCol, smuInCol, smuOutCol, smuReasonCol, smuCRCol, smuCWCol)

		if smuRows, err := db.Query(smuQuery); err == nil {
			for smuRows.Next() {
				var sessID, m string
				var in, out, reason, cr, cw int64
				if err := smuRows.Scan(&sessID, &m, &in, &out, &reason, &cr, &cw); err == nil {
					modelUsages[sessID] = append(modelUsages[sessID], hermesModelUsage{
						model:      modelName(m),
						input:      in,
						output:     out,
						reasoning:  reason,
						cacheRead:  cr,
						cacheWrite: cw,
					})
				}
			}
			smuRows.Close()
		}
	}

	endSelect := "NULL"
	if endTimeCol != "" {
		endSelect = endTimeCol
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s
		FROM %s
	`, timeCol, endSelect, cwdCol, modelCol, inputCol, outputCol, reasoningCol, cacheReadCol, cacheWriteCol, costCol, turnsCol, sessionTable)

	rows, err := db.Query(query)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer rows.Close()

	var summary thermal.Summary
	summary.ModelBreakdown = make(map[string]int64)

	byDay := make(map[string]*thermal.DailyRow)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)

	for rows.Next() {
		var id string
		var rawTime, rawEndTime any
		var cwd, rawModel string
		var input, output, reasoning, cacheRead, cacheWrite int64
		var cost float64
		var turns int

		if err := rows.Scan(&id, &rawTime, &rawEndTime, &cwd, &rawModel, &input, &output, &reasoning, &cacheRead, &cacheWrite, &cost, &turns); err != nil {
			continue
		}

		startT, ok := parseHermesTimestamp(rawTime)
		if !ok || startT.IsZero() {
			continue
		}
		day := thermal.LocalDay(startT.Local())

		if rawEndTime != nil {
			if endT, ok := parseHermesTimestamp(rawEndTime); ok && endT.After(startT) {
				durMs := endT.Sub(startT).Milliseconds()
				if durMs > summary.LongestSessionMs {
					summary.LongestSessionMs = durMs
				}
			}
		}

		if turns <= 0 {
			turns = 1
		}

		projectKey := ""
		if cwd != "" {
			projectKey = thermal.ProjectKey(cwd)
		}

		row := byDay[day]
		if row == nil {
			row = &thermal.DailyRow{Day: day, Models: make(map[string]thermal.ModelTokens)}
			byDay[day] = row
		}

		var pd *thermal.ProjectDay
		if projectKey != "" {
			key := projectDayKey{day: day, project: projectKey}
			pd = byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{Project: projectKey, Day: day, Models: make(map[string]thermal.ModelTokens)}
				byProjectDay[key] = pd
			}
		}

		summary.Sessions++
		summary.Cost += cost

		row.Turns += turns
		row.Cost += cost
		if pd != nil {
			pd.Turns += turns
			pd.Cost += cost
		}

		// If session_model_usage has entries for this session, attribute tokens per model.
		if smuList, ok := modelUsages[id]; ok && len(smuList) > 0 {
			for _, smu := range smuList {
				mTotal := smu.input + smu.output + smu.reasoning + smu.cacheRead + smu.cacheWrite
				summary.LifetimeTokens += mTotal
				summary.InputTokens += smu.input
				summary.OutputTokens += smu.output
				summary.ReasoningTokens += smu.reasoning
				summary.CacheTokens += smu.cacheRead + smu.cacheWrite

				row.Tokens += mTotal
				row.Input += smu.input
				row.Output += smu.output
				row.Reasoning += smu.reasoning
				row.Cache += smu.cacheRead + smu.cacheWrite

				if pd != nil {
					pd.Tokens += mTotal
					pd.Input += smu.input
					pd.Output += smu.output
					pd.Reasoning += smu.reasoning
					pd.CacheRead += smu.cacheRead
					pd.CacheWrite += smu.cacheWrite
				}

				if smu.model != "" {
					summary.ModelBreakdown[smu.model]++
					row.Models[smu.model] = row.Models[smu.model].Add(thermal.ModelTokens{
						Input:      smu.input,
						Output:     smu.output,
						Reasoning:  smu.reasoning,
						CacheRead:  smu.cacheRead,
						CacheWrite: smu.cacheWrite,
					})
					if pd != nil {
						pd.Models[smu.model] = pd.Models[smu.model].Add(thermal.ModelTokens{
							Input:      smu.input,
							Output:     smu.output,
							Reasoning:  smu.reasoning,
							CacheRead:  smu.cacheRead,
							CacheWrite: smu.cacheWrite,
						})
					}
				}
			}
		} else {
			model := modelName(rawModel)
			tot := input + output + reasoning + cacheRead + cacheWrite

			summary.LifetimeTokens += tot
			summary.InputTokens += input
			summary.OutputTokens += output
			summary.ReasoningTokens += reasoning
			summary.CacheTokens += cacheRead + cacheWrite

			row.Tokens += tot
			row.Input += input
			row.Output += output
			row.Reasoning += reasoning
			row.Cache += cacheRead + cacheWrite

			if pd != nil {
				pd.Tokens += tot
				pd.Input += input
				pd.Output += output
				pd.Reasoning += reasoning
				pd.CacheRead += cacheRead
				pd.CacheWrite += cacheWrite
			}

			if model != "" {
				summary.ModelBreakdown[model]++
				row.Models[model] = row.Models[model].Add(thermal.ModelTokens{
					Input:      input,
					Output:     output,
					Reasoning:  reasoning,
					CacheRead:  cacheRead,
					CacheWrite: cacheWrite,
				})
				if pd != nil {
					pd.Models[model] = pd.Models[model].Add(thermal.ModelTokens{
						Input:      input,
						Output:     output,
						Reasoning:  reasoning,
						CacheRead:  cacheRead,
						CacheWrite: cacheWrite,
					})
				}
			}
		}
	}

	daily := make([]thermal.DailyRow, 0, len(byDay))
	for _, r := range byDay {
		daily = append(daily, *r)
	}
	sort.Slice(daily, func(i, j int) bool {
		return daily[i].Day < daily[j].Day
	})

	projects := make([]thermal.ProjectDay, 0, len(byProjectDay))
	for _, p := range byProjectDay {
		projects = append(projects, *p)
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Day != projects[j].Day {
			return projects[i].Day < projects[j].Day
		}
		return projects[i].Project < projects[j].Project
	})

	return summary, daily, projects, nil
}
