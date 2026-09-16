package loaders

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// ticksToUSD converts Grok's costUsdTicks to dollars (1 tick = 1e-10 USD,
// per the ccusage Grok source documentation).
const ticksToUSD = 1e-10

// grokUsage mirrors the usage object on turn_completed updates. inputTokens
// follows OpenAI convention and already includes cached reads, and reasoning
// tokens are a subset of output tokens. The loader subtracts the nested parts
// so the stored token types are disjoint and add up to the recorded total.
type grokUsage struct {
	InputTokens         int64 `json:"inputTokens"`
	OutputTokens        int64 `json:"outputTokens"`
	TotalTokens         int64 `json:"totalTokens"`
	CachedReadTokens    int64 `json:"cachedReadTokens"`
	CacheCreationTokens int64 `json:"cacheCreationTokens"`
	ReasoningTokens     int64 `json:"reasoningTokens"`
	CostUsdTicks        int64 `json:"costUsdTicks"`
	ModelUsage          map[string]struct {
		ModelCalls int64 `json:"modelCalls"`
	} `json:"modelUsage"`
}

// LoadGrokData scans Grok CLI session logs in parallel using a bounded
// worker pool. The primary source is updates.jsonl: only turn_completed
// updates carry usable per-turn usage. summary.json sidecars supply the agent
// name and session duration. Sessions killed mid-turn never record usage and
// are invisible, matching upstream reporter behavior.
func LoadGrokData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	var summary thermal.Summary
	summary.Tool = "Grok"

	pattern := filepath.Join(dataDir, "sessions", "*", "*", "updates.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return summary, nil, nil, err
	}

	type turnAgg struct {
		day        string
		input      int64
		output     int64
		reason     int64
		cacheRead  int64
		cacheWrite int64
		total      int64
		cost       float64
		models     map[string]int64
		firstTs    int64
		lastTs     int64
	}
	type fileResult struct {
		sessionID string
		agent     string
		turns     []turnAgg
		duration  int64 // ms, from summary.json sidecar
	}

	results := make(chan fileResult, len(files))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8) // max 8 concurrent workers

	for _, path := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var res fileResult

			f, err := os.Open(p)
			if err != nil {
				results <- res
				return
			}

			scanner := bufio.NewScanner(f)
			scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				var rec struct {
					Timestamp int64  `json:"timestamp"`
					Method    string `json:"method"`
					Params    struct {
						SessionID string `json:"sessionId"`
						Update    struct {
							SessionUpdate string    `json:"sessionUpdate"`
							Usage         grokUsage `json:"usage"`
						} `json:"update"`
					} `json:"params"`
				}
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					continue
				}
				if rec.Method != "_x.ai/session/update" ||
					rec.Params.Update.SessionUpdate != "turn_completed" {
					continue
				}
				u := rec.Params.Update.Usage
				if u.TotalTokens == 0 && u.InputTokens == 0 && u.OutputTokens == 0 {
					continue
				}
				if res.sessionID == "" {
					res.sessionID = rec.Params.SessionID
				}
				total := u.TotalTokens
				if total == 0 {
					total = u.InputTokens + u.OutputTokens
				}
				models := make(map[string]int64)
				for name, m := range u.ModelUsage {
					if name != "" && m.ModelCalls > 0 {
						models[name] += m.ModelCalls
					}
				}
				res.turns = append(res.turns, turnAgg{
					day:        thermal.UnixDay(rec.Timestamp),
					input:      nonNegative(u.InputTokens - u.CachedReadTokens - u.CacheCreationTokens),
					output:     nonNegative(u.OutputTokens - u.ReasoningTokens),
					reason:     u.ReasoningTokens,
					cacheRead:  u.CachedReadTokens,
					cacheWrite: u.CacheCreationTokens,
					total:      total,
					cost:       float64(u.CostUsdTicks) * ticksToUSD,
					models:     models,
					firstTs:    rec.Timestamp,
					lastTs:     rec.Timestamp,
				})
			}
			f.Close()

			if len(res.turns) == 0 {
				results <- res
				return
			}

			// Agent name + duration from the summary.json sidecar.
			if data, err := os.ReadFile(filepath.Join(filepath.Dir(p), "summary.json")); err == nil {
				var meta struct {
					AgentName string `json:"agent_name"`
					CreatedAt string `json:"created_at"`
					UpdatedAt string `json:"updated_at"`
				}
				if json.Unmarshal(data, &meta) == nil {
					res.agent = meta.AgentName
					if start, err := time.Parse(time.RFC3339, meta.CreatedAt); err == nil {
						if end, err := time.Parse(time.RFC3339, meta.UpdatedAt); err == nil {
							res.duration = end.Sub(start).Milliseconds()
						}
					}
				}
			}
			results <- res
		}(path)
	}

	wg.Wait()
	close(results)

	type dayAgg struct {
		tokens     int64
		input      int64
		output     int64
		reason     int64
		cacheRead  int64
		cacheWrite int64
		cost       float64
		turns      int
		models     map[string]thermal.ModelTokens
	}
	byDay := make(map[string]*dayAgg)
	modelCounts := make(map[string]int64)
	agentCounts := make(map[string]int)

	for res := range results {
		if len(res.turns) == 0 {
			continue
		}
		summary.Sessions++
		if res.duration > summary.LongestSessionMs {
			summary.LongestSessionMs = res.duration
		}
		if res.agent != "" {
			agentCounts[res.agent]++
		}
		for _, t := range res.turns {
			summary.InputTokens += t.input
			summary.OutputTokens += t.output
			summary.ReasoningTokens += t.reason
			summary.CacheTokens += t.cacheRead + t.cacheWrite
			summary.LifetimeTokens += t.total
			summary.Cost += t.cost
			for name, n := range t.models {
				modelCounts[name] += n
			}
			agg := byDay[t.day]
			if agg == nil {
				agg = &dayAgg{models: make(map[string]thermal.ModelTokens)}
				byDay[t.day] = agg
			}
			agg.tokens += t.total
			agg.input += t.input
			agg.output += t.output
			agg.reason += t.reason
			agg.cacheRead += t.cacheRead
			agg.cacheWrite += t.cacheWrite
			agg.cost += t.cost
			agg.turns++
			// modelUsage carries call counts, not tokens. Attribute the turn's
			// tokens only when the turn used exactly one model.
			if len(t.models) == 1 {
				for name := range t.models {
					agg.models[name] = agg.models[name].Add(thermal.ModelTokens{
						Input:      t.input,
						Output:     t.output,
						Reasoning:  t.reason,
						CacheRead:  t.cacheRead,
						CacheWrite: t.cacheWrite,
					})
				}
			}
		}
	}

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:       day,
			Tokens:    agg.tokens,
			Input:     agg.input,
			Output:    agg.output,
			Reasoning: agg.reason,
			Cache:     agg.cacheRead + agg.cacheWrite,
			Cost:      agg.cost,
			Turns:     agg.turns,
			Models:    agg.models,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}
	if len(agentCounts) > 0 {
		summary.AgentBreakdown = agentCounts
	}

	return summary, daily, nil, nil
}
