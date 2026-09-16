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

// LoadCodewhaleData reads CodeWhale session JSON files from
// ~/.codewhale/sessions/*.json. Each file has a metadata block with
// total_tokens, cost (session_cost_usd), model, mode, message_count, and
// cumulative_turn_secs — all pre-aggregated at the session level.
//
// The old loader only counted sessions (1 session = 1 activity unit). This
// surfaces the real token volume and cost, moving codewhale from Activity
// Hunters to Token Warriors.
func LoadCodewhaleData(dataDir string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	sessionsDir := filepath.Join(dataDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return thermal.Summary{}, nil, nil, fmt.Errorf("cannot read %s: %w", sessionsDir, err)
	}

	type dayAgg struct {
		tokens int64
		cost   float64
		turns  int
		models map[string]thermal.ModelTokens
	}
	byDay := make(map[string]*dayAgg)
	modelCounts := make(map[string]int64)
	modeCounts := make(map[string]int)

	var summary thermal.Summary
	summary.Tool = "codewhale"

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
				CreatedAt          string  `json:"created_at"`
				UpdatedAt          string  `json:"updated_at"`
				MessageCount       int     `json:"message_count"`
				TotalTokens        int64   `json:"total_tokens"`
				Model              string  `json:"model"`
				Mode               string  `json:"mode"`
				Workspace          string  `json:"workspace"`
				CumulativeTurnSecs int64   `json:"cumulative_turn_secs"`
				Cost               struct {
					SessionCostUSD float64 `json:"session_cost_usd"`
				} `json:"cost"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		md := session.Metadata
		if md.CreatedAt == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, md.CreatedAt)
		if err != nil {
			continue
		}
		day := thermal.LocalDay(t)

		agg := byDay[day]
		if agg == nil {
			agg = &dayAgg{models: make(map[string]thermal.ModelTokens)}
			byDay[day] = agg
		}
		agg.tokens += md.TotalTokens
		agg.cost += md.Cost.SessionCostUSD
		agg.turns += md.MessageCount
		if md.Model != "" && md.TotalTokens > 0 {
			// codewhale records a session total with no token type split, so
			// the tokens land in the unclassified bucket.
			agg.models[md.Model] = agg.models[md.Model].Add(thermal.ModelTokens{Unclassified: md.TotalTokens})
		}

		summary.Sessions++
		summary.LifetimeTokens += md.TotalTokens
		summary.Cost += md.Cost.SessionCostUSD

		// Session duration (seconds → ms).
		durationMs := md.CumulativeTurnSecs * 1000
		if durationMs > summary.LongestSessionMs {
			summary.LongestSessionMs = durationMs
		}

		if md.Model != "" {
			modelCounts[md.Model]++
		}
		if md.Mode != "" {
			modeCounts[md.Mode]++
		}
	}

	// Build daily rows.
	var daily []thermal.DailyRow
	for day, agg := range byDay {
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: agg.tokens,
			Cost:   agg.cost,
			Turns:  agg.turns,
			Models: agg.models,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	if len(modelCounts) > 0 {
		summary.ModelBreakdown = modelCounts
	}
	if len(modeCounts) > 0 {
		summary.AgentBreakdown = modeCounts
	}

	return summary, daily, nil, nil
}
