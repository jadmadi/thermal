package loaders

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func loadJsonlData(dataDir string, fieldTimestamp string, useMillis bool) (thermal.Summary, []thermal.DailyRow, error) {
	historyPath := filepath.Join(dataDir, "history.jsonl")
	f, err := os.Open(historyPath)
	if err != nil {
		return thermal.Summary{}, nil, fmt.Errorf("cannot open %s: %w", historyPath, err)
	}
	defer f.Close()

	dayCounts := make(map[string]int)
	var totalCommands int64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		var ts int64
		if v, ok := raw[fieldTimestamp].(float64); ok {
			ts = int64(v)
		}
		if ts == 0 {
			continue
		}

		var t time.Time
		if useMillis {
			t = time.UnixMilli(ts)
		} else {
			t = time.Unix(ts, 0)
		}
		day := thermal.LocalDay(t)
		dayCounts[day]++
		totalCommands++
	}

	var daily []thermal.DailyRow
	for day, count := range dayCounts {
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: int64(count),
			Turns:  count,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	summary := thermal.Summary{
		Sessions:       0,
		LifetimeTokens: totalCommands,
	}

	return summary, daily, nil
}
