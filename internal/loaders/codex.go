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
func LoadCodexData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	stateDB := filepath.Join(dataDir, "state_5.sqlite")
	if _, err := os.Stat(stateDB); err == nil {
		return loadCodexFromStateDB(dataDir, stateDB)
	}
	return loadJsonlData(dataDir, "ts", false)
}

func loadCodexFromStateDB(dataDir, dbPath string) (thermal.Summary, []thermal.DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return thermal.Summary{}, nil, fmt.Errorf("cannot open %s: %w", dbPath, err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, tokens_used, model, source, reasoning_effort, agent_role,
		       created_at, updated_at, rollout_path
		FROM threads
		WHERE archived = 0
		ORDER BY created_at
	`)
	if err != nil {
		return thermal.Summary{}, nil, fmt.Errorf("cannot query threads: %w", err)
	}
	defer rows.Close()

	type dayAgg struct {
		tokens int64
		turns  int
	}
	byDay := make(map[string]*dayAgg)
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
	}
	var threads []threadInfo

	for rows.Next() {
		var t threadInfo
		var model, source, reasoning, agentRole, rolloutPath sql.NullString
		var createdAt, updatedAt int64

		if err := rows.Scan(&t.id, &t.tokensUsed, &model, &source, &reasoning,
			&agentRole, &createdAt, &updatedAt, &rolloutPath); err != nil {
			continue
		}

		t.day = thermal.LocalDay(time.Unix(createdAt, 0).Local())

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

		summary.Sessions++
		summary.LifetimeTokens += t.tokensUsed

		if model.Valid && model.String != "" {
			modelCounts[model.String]++
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

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: agg.tokens,
			Turns:  agg.turns,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

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

	for _, b := range breakdowns {
		if b == nil {
			continue
		}
		summary.InputTokens += b.input
		summary.OutputTokens += b.output
		summary.ReasoningTokens += b.reasoning
		summary.CacheTokens += b.cache
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

	return summary, daily, nil
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
		last = &tokenBreakdown{
			input:     tu.InputTokens,
			output:    tu.OutputTokens,
			reasoning: tu.ReasoningOutputTokens,
			cache:     tu.CachedInputTokens,
		}
	}

	return last
}
