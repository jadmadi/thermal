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

// LoadCommandCodeData reads command-code session transcripts in parallel using a bounded worker pool.
func LoadCommandCodeData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	projectsDir := filepath.Join(dataDir, "projects")

	var summary thermal.Summary
	summary.Tool = "command-code"

	pattern := filepath.Join(projectsDir, "*", "*.jsonl")
	allFiles, err := filepath.Glob(pattern)
	if err != nil {
		return summary, nil, err
	}

	var sessionFiles []string
	for _, p := range allFiles {
		base := filepath.Base(p)
		if strings.Contains(base, ".checkpoints.") || strings.HasPrefix(base, "hooks-audit-") {
			continue
		}
		sessionFiles = append(sessionFiles, p)
	}

	type fileResult struct {
		day         string
		msgCount    int
		model       string
		durationMs  int64
		hasSession  bool
		dayCounts   map[string]int
	}

	results := make(chan fileResult, len(sessionFiles))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8) // max 8 concurrent workers

	for _, path := range sessionFiles {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res := fileResult{
				dayCounts: make(map[string]int),
			}

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
					Timestamp string `json:"timestamp"`
					Role      string `json:"role"`
					Source    string `json:"source"`
					Version   string `json:"version"`
				}
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
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
				res.day = day
				res.msgCount++
				res.dayCounts[day]++
			}
			f.Close()

			// Read meta sidecar
			metaPath := strings.TrimSuffix(p, ".jsonl") + ".meta.json"
			if data, err := os.ReadFile(metaPath); err == nil {
				var meta struct {
					Model string `json:"model"`
				}
				if json.Unmarshal(data, &meta) == nil && meta.Model != "" {
					res.model = meta.Model
				}
			}

			if res.msgCount > 0 {
				res.hasSession = true
				res.durationMs = lastTs.Sub(firstTs).Milliseconds()
			}
			results <- res
		}(path)
	}

	wg.Wait()
	close(results)

	byDay := make(map[string]int)
	modelCounts := make(map[string]int64)

	for res := range results {
		if !res.hasSession {
			continue
		}
		summary.Sessions++
		if res.durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = res.durationMs
		}
		if res.model != "" && res.day != "" {
			modelCounts[res.model]++
		}
		for day, count := range res.dayCounts {
			byDay[day] += count
		}
	}

	var daily []thermal.DailyRow
	for day, count := range byDay {
		summary.LifetimeTokens += int64(count)
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

	if len(daily) == 0 {
		return loadJsonlData(dataDir, "t", true)
	}

	return summary, daily, nil
}
