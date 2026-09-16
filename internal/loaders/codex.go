package loaders

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
	_ "modernc.org/sqlite"
)

// LoadCodexData reads Codex session data from state_5.sqlite primary source
// or falls back to history.jsonl if stateDB doesn't exist.
func LoadCodexData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	stateDB := filepath.Join(dataDir, "state_5.sqlite")
	if _, err := os.Stat(stateDB); err == nil {
		return loadCodexFromStateDB(dataDir, stateDB)
	}
	return loadJsonlData(dataDir, "ts", false)
}

func loadCodexFromStateDB(dataDir, dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot open %s: %w", dbPath, err)
	}
	defer db.Close()

	// Older Codex schemas lack the cwd column.
	cwdExpr := "''"
	if hasColumn(db, "threads", "cwd") {
		cwdExpr = "cwd"
	}
	rows, err := db.Query(`
		SELECT id, tokens_used, model, source, reasoning_effort, agent_role,
		       created_at, updated_at, rollout_path, ` + cwdExpr + `
		FROM threads
		WHERE archived = 0
		ORDER BY created_at
	`)
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot query threads: %w", err)
	}
	defer rows.Close()

	type dayAgg struct {
		tokens    int64
		turns     int
		input     int64
		output    int64
		reasoning int64
		cache     int64
		models    map[string]thermal.ModelTokens
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
	}
	var threads []threadInfo

	for rows.Next() {
		var t threadInfo
		var model, source, reasoning, agentRole, rolloutPath, cwd sql.NullString
		var createdAt, updatedAt int64

		if err := rows.Scan(&t.id, &t.tokensUsed, &model, &source, &reasoning,
			&agentRole, &createdAt, &updatedAt, &rolloutPath, &cwd); err != nil {
			continue
		}

		t.day = thermal.LocalDay(time.Unix(createdAt, 0).Local())
		t.project = thermal.ProjectKey(cwd.String)

		durationMs := (updatedAt - createdAt) * 1000
		if durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = durationMs
		}

		agg := byDay[t.day]
		if agg == nil {
			agg = &dayAgg{}
			byDay[t.day] = agg
		}
		agg.tokens += t.tokensUsed
		agg.turns++

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

		summary.Sessions++
		summary.LifetimeTokens += t.tokensUsed

		if model.Valid && model.String != "" {
			t.model = modelName(model.String)
			modelCounts[t.model]++
		}
		if source.Valid && source.String != "" {
			sourceCounts[source.String]++
		}
		if reasoning.Valid && reasoning.String != "" {
			reasoningCounts[reasoning.String]++
		}

		if rolloutPath.Valid {
			t.rolloutPath = rolloutPath.String
		}
		threads = append(threads, t)
	}

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}
	if len(sourceCounts) > 0 {
		summary.AgentBreakdown = sourceCounts
	}

	// Bounded worker pool for parallel rollout scanning
	breakdowns := make([]*tokenBreakdown, len(threads))
	var wg sync.WaitGroup
	workerLimit := make(chan struct{}, 8) // max 8 concurrent workers

	for i, t := range threads {
		if t.rolloutPath == "" || t.tokensUsed == 0 {
			continue
		}
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			workerLimit <- struct{}{}
			defer func() { <-workerLimit }()
			breakdowns[idx] = readLastTokenBreakdown(path)
		}(i, t.rolloutPath)
	}
	wg.Wait()

	for i, b := range breakdowns {
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
		t := threads[i]
		breakdownTotal := b.input + b.output + b.reasoning + b.cache
		if t.tokensUsed <= 0 || breakdownTotal <= 0 {
			continue
		}
		ratio := float64(t.tokensUsed) / float64(breakdownTotal)
		scaled := tokenBreakdown{
			input:     int64(float64(b.input) * ratio),
			output:    int64(float64(b.output) * ratio),
			reasoning: int64(float64(b.reasoning) * ratio),
			cache:     int64(float64(b.cache) * ratio),
		}
		agg := byDay[t.day]
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
			Day:       day,
			Tokens:    agg.tokens,
			Input:     agg.input,
			Output:    agg.output,
			Reasoning: agg.reasoning,
			Cache:     agg.cache,
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

	return summary, daily, projects, nil
}

type tokenBreakdown struct {
	input, output, reasoning, cache int64
}

func readLastTokenBreakdown(rolloutPath string) *tokenBreakdown {
	if _, err := os.Stat(rolloutPath); err != nil {
		return nil
	}

	f, err := os.Open(rolloutPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var last *tokenBreakdown
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.Contains(line, `"token_count"`) {
			continue
		}

		var rec struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if rec.Type != "event_msg" {
			continue
		}

		var ev struct {
			Type string `json:"type"`
			Info *struct {
				TotalTokenUsage *struct {
					InputTokens           int64 `json:"input_tokens"`
					CachedInputTokens     int64 `json:"cached_input_tokens"`
					OutputTokens          int64 `json:"output_tokens"`
					ReasoningOutputTokens int64 `json:"reasoning_output_tokens"`
				} `json:"total_token_usage"`
			} `json:"info"`
		}
		if json.Unmarshal(rec.Payload, &ev) != nil {
			continue
		}
		if ev.Type != "token_count" || ev.Info == nil || ev.Info.TotalTokenUsage == nil {
			continue
		}

		tu := ev.Info.TotalTokenUsage
		// OpenAI-style usage nests cached reads inside input_tokens and
		// reasoning inside output_tokens, and total_tokens equals
		// input plus output. Subtract the nested parts so the four token
		// types are disjoint and add up to the recorded total.
		last = &tokenBreakdown{
			input:     nonNegative(tu.InputTokens - tu.CachedInputTokens),
			output:    nonNegative(tu.OutputTokens - tu.ReasoningOutputTokens),
			reasoning: tu.ReasoningOutputTokens,
			cache:     tu.CachedInputTokens,
		}
	}

	return last
}

func nonNegative(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}
