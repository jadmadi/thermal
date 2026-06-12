package loaders

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func LoadCodewhaleData(dataDir string) (thermal.Summary, []thermal.DailyRow, error) {
	sessionsDir := filepath.Join(dataDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return thermal.Summary{}, nil, fmt.Errorf("cannot read %s: %w", sessionsDir, err)
	}

	dayCounts := make(map[string]int)
	var totalSessions int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name()))
		if err != nil {
			continue
		}
		var session struct {
			Metadata struct {
				CreatedAt string `json:"created_at"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.Metadata.CreatedAt == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, session.Metadata.CreatedAt)
		if err != nil {
			continue
		}
		day := thermal.LocalDay(t)
		dayCounts[day]++
		totalSessions++
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
		Tool:           "codewhale",
		Sessions:       totalSessions,
		LifetimeTokens: int64(totalSessions),
	}

	return summary, daily, nil
}
