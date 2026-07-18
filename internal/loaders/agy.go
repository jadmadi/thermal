package loaders

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// LoadAgyData reads Google Antigravity session data in parallel using a bounded worker pool.
func LoadAgyData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	brainDir := filepath.Join(dataDir, "brain")
	entries, err := os.ReadDir(brainDir)
	if err != nil {
		return thermal.Summary{}, nil, fmt.Errorf("cannot read %s: %w", brainDir, err)
	}

	type sessionResult struct {
		steps       int
		firstTs     time.Time
		lastTs      time.Time
		dayCounts   map[string]int
		modelCounts map[string]int64
	}

	results := make(chan sessionResult, len(entries))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8) // max 8 concurrent workers

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionDir := filepath.Join(brainDir, entry.Name())
		wg.Add(1)

		go func(sDir string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res := sessionResult{
				dayCounts:   make(map[string]int),
				modelCounts: make(map[string]int64),
			}

			logsDir := filepath.Join(sDir, ".system_generated", "logs")

			// 1. Process overview.txt
			overviewPath := filepath.Join(logsDir, "overview.txt")
			if f, err := os.Open(overviewPath); err == nil {
				scanner := bufio.NewScanner(f)
				scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
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
					}
					if json.Unmarshal([]byte(line), &rec) != nil {
						continue
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

					if res.firstTs.IsZero() || t.Before(res.firstTs) {
						res.firstTs = t
					}
					if t.After(res.lastTs) {
						res.lastTs = t
					}
				}
				f.Close()
			}

			// 2. Process transcript.jsonl for model extraction
			transcriptPath := filepath.Join(logsDir, "transcript.jsonl")
			if f, err := os.Open(transcriptPath); err == nil {
				scanner := bufio.NewScanner(f)
				scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" {
						continue
					}
					if strings.Contains(line, "Model Selection") {
						var rec struct {
							Content string `json:"content"`
						}
						if json.Unmarshal([]byte(line), &rec) == nil {
							if idx := strings.Index(rec.Content, "Model Selection"); idx >= 0 {
								rest := rec.Content[idx:]
								if toIdx := strings.Index(rest, " to "); toIdx >= 0 {
									model := strings.TrimSpace(rest[toIdx+4:])
									if idx := strings.IndexAny(model, "\r\n\""); idx >= 0 {
										model = strings.TrimSpace(model[:idx])
									}
									model = strings.TrimSuffix(model, ".")
									model = strings.TrimSpace(model)
									if model != "" && model != "None" {
										res.modelCounts[model]++
									}
								}
							}
						}
					}
				}
				f.Close()
			}

			if res.steps > 0 {
				results <- res
			}
		}(sessionDir)
	}

	wg.Wait()
	close(results)

	byDay := make(map[string]int)
	modelCounts := make(map[string]int64)
	var summary thermal.Summary
	summary.Tool = "Agy"

	for res := range results {
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

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}

	return summary, daily, nil
}
