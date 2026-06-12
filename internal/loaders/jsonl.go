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

	"thermal/internal/thermal"
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

func LoadCommandCodeData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	summary, daily, err := loadJsonlData(dataDir, "t", true)
	summary.Tool = "command-code"
	return summary, daily, err
}

func LoadCodexData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	summary, daily, err := loadJsonlData(dataDir, "ts", false)
	summary.Tool = "Codex"
	return summary, daily, err
}

func LoadAgyData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	brainDir := filepath.Join(dataDir, "brain")
	entries, err := os.ReadDir(brainDir)
	if err != nil {
		return thermal.Summary{}, nil, fmt.Errorf("cannot read %s: %w", brainDir, err)
	}

	dayCounts := make(map[string]int)
	var totalEntries int64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		transcriptPath := filepath.Join(brainDir, entry.Name(), ".system_generated", "logs", "transcript.jsonl")
		f, err := os.Open(transcriptPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var record struct {
				CreatedAt string `json:"created_at"`
			}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}
			if record.CreatedAt == "" {
				continue
			}
			t, err := time.Parse(time.RFC3339, record.CreatedAt)
			if err != nil {
				continue
			}
			day := thermal.LocalDay(t)
			dayCounts[day]++
			totalEntries++
		}
		f.Close()
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
		Tool:           "Agy",
		Sessions:       0,
		LifetimeTokens: totalEntries,
	}

	return summary, daily, nil
}
