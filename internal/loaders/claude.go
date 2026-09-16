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

// LoadClaudeData scans Claude Code session transcripts in parallel using a
// bounded worker pool. Tokens come from assistant.message.usage
// (input_tokens, output_tokens, cache_read_input_tokens,
// cache_creation_input_tokens); each file under projects/ is one session and
// assistant messages are the countable turns. No cost fields exist in the
// transcripts, so Cost stays 0.
func LoadClaudeData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	var summary thermal.Summary
	summary.Tool = "Claude"

	pattern := filepath.Join(dataDir, "projects", "*", "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return summary, nil, err
	}

	type msgAgg struct {
		day        string
		input      int64
		output     int64
		cacheRead  int64
		cacheWrite int64
		model      string
	}
	type fileResult struct {
		msgs     []msgAgg
		model    string
		duration int64 // ms between first and last timestamp
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

			var firstTs, lastTs time.Time
			scanner := bufio.NewScanner(f)
			scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				var rec struct {
					Type      string `json:"type"`
					Timestamp string `json:"timestamp"`
					Message   struct {
						Model string `json:"model"`
						Usage struct {
							InputTokens              int64 `json:"input_tokens"`
							OutputTokens             int64 `json:"output_tokens"`
							CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
							CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
						} `json:"usage"`
					} `json:"message"`
				}
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					continue
				}
				if rec.Type != "assistant" {
					continue
				}
				day := ""
				if rec.Timestamp != "" {
					if t, err := time.Parse(time.RFC3339, rec.Timestamp); err == nil {
						day = thermal.LocalDay(t.Local())
						if firstTs.IsZero() || t.Before(firstTs) {
							firstTs = t
						}
						if t.After(lastTs) {
							lastTs = t
						}
					}
				}
				if day == "" {
					continue
				}
				res.msgs = append(res.msgs, msgAgg{
					day:        day,
					input:      rec.Message.Usage.InputTokens,
					output:     rec.Message.Usage.OutputTokens,
					cacheRead:  rec.Message.Usage.CacheReadInputTokens,
					cacheWrite: rec.Message.Usage.CacheCreationInputTokens,
					model:      rec.Message.Model,
				})
			}
			f.Close()

			if len(res.msgs) == 0 {
				results <- res
				return
			}
			res.duration = lastTs.Sub(firstTs).Milliseconds()
			results <- res
		}(path)
	}

	wg.Wait()
	close(results)

	type dayAgg struct {
		input, output         int64
		cacheRead, cacheWrite int64
		turns                 int
		models                map[string]thermal.ModelTokens
	}
	byDay := make(map[string]*dayAgg)
	modelCounts := make(map[string]int64)

	for res := range results {
		if len(res.msgs) == 0 {
			continue
		}
		summary.Sessions++
		if res.duration > summary.LongestSessionMs {
			summary.LongestSessionMs = res.duration
		}
		for _, m := range res.msgs {
			summary.InputTokens += m.input
			summary.OutputTokens += m.output
			summary.CacheTokens += m.cacheRead + m.cacheWrite
			summary.LifetimeTokens += m.input + m.output + m.cacheRead + m.cacheWrite
			if m.model != "" {
				modelCounts[m.model]++
			}
			agg := byDay[m.day]
			if agg == nil {
				agg = &dayAgg{models: make(map[string]thermal.ModelTokens)}
				byDay[m.day] = agg
			}
			agg.input += m.input
			agg.output += m.output
			agg.cacheRead += m.cacheRead
			agg.cacheWrite += m.cacheWrite
			agg.turns++
			if m.model != "" && (m.input+m.output+m.cacheRead+m.cacheWrite) > 0 {
				agg.models[m.model] = agg.models[m.model].Add(thermal.ModelTokens{
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
			Turns:  agg.turns,
			Models: agg.models,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}

	return summary, daily, nil
}
