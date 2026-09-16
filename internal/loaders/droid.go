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

// LoadDroidData scans Factory Droid session transcripts in parallel using a
// bounded worker pool. Session files carry no token or cost telemetry, so
// this is an activity-only loader: message records stand in for activity the
// way they do for command-code. Each *.jsonl file under sessions/ is one
// session; per-file *.settings.json sidecars carry no usage data.
func LoadDroidData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	var summary thermal.Summary
	summary.Tool = "Droid"

	pattern := filepath.Join(dataDir, "sessions", "*", "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return summary, nil, nil, err
	}

	type fileResult struct {
		dayCounts map[string]int
		duration  int64 // ms between first and last message
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

			res := fileResult{dayCounts: make(map[string]int)}

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
				}
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					continue
				}
				if rec.Type != "message" || rec.Timestamp == "" {
					continue
				}
				t, err := time.Parse(time.RFC3339, rec.Timestamp)
				if err != nil {
					continue
				}
				res.dayCounts[thermal.LocalDay(t.Local())]++
				if firstTs.IsZero() || t.Before(firstTs) {
					firstTs = t
				}
				if t.After(lastTs) {
					lastTs = t
				}
			}
			f.Close()

			if len(res.dayCounts) == 0 {
				results <- res
				return
			}
			res.duration = lastTs.Sub(firstTs).Milliseconds()
			results <- res
		}(path)
	}

	wg.Wait()
	close(results)

	byDay := make(map[string]int)
	for res := range results {
		if len(res.dayCounts) == 0 {
			continue
		}
		summary.Sessions++
		if res.duration > summary.LongestSessionMs {
			summary.LongestSessionMs = res.duration
		}
		for day, n := range res.dayCounts {
			byDay[day] += n
		}
	}

	var daily []thermal.DailyRow
	for day, n := range byDay {
		summary.LifetimeTokens += int64(n)
		daily = append(daily, thermal.DailyRow{Day: day, Tokens: int64(n), Turns: n})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	return summary, daily, nil, nil
}
