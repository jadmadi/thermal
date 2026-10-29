// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

// jsonPrecision is the documented maximum number of decimal places for
// serialized floats. Costs, percentages, shares, and means all use this one
// boundary rule so no raw float noise reaches a machine consumer.
const jsonPrecision = 4

func roundJSONValue(v any) any {
	switch x := v.(type) {
	case float64:
		if math.Abs(x) < 0.00005 {
			return float64(0)
		}
		return math.Round(x*math.Pow10(jsonPrecision)) / math.Pow10(jsonPrecision)
	case []any:
		for i := range x {
			x[i] = roundJSONValue(x[i])
		}
		return x
	case map[string]any:
		for key, value := range x {
			x[key] = roundJSONValue(value)
		}
		return x
	default:
		return v
	}
}

func roundedJSON(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, err
	}
	return roundJSONValue(generic), nil
}

func marshalJSON(v any) ([]byte, error) {
	generic, err := roundedJSON(v)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(generic, "", "  ")
}

func marshalJSONCompact(v any) ([]byte, error) {
	generic, err := roundedJSON(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(generic)
}

func writeJSON(v any) {
	b, err := marshalJSON(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "thermal: failed writing JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

// writeReportJSON prints any report payload with the standard indentation and
// the shared precision policy.
func writeReportJSON(v interface{}) {
	writeJSON(v)
}

func leaderboardJSONPayload(results []thermal.ToolResult, now time.Time, topN ...int) map[string]interface{} {
	top := 0
	if len(topN) > 0 {
		top = topN[0]
	}
	type leaderboardTotals struct {
		Tools         int     `json:"tools"`
		ActiveDays    int     `json:"activeDays"`
		Tokens        int64   `json:"totalTokens"`
		RecordedCost  float64 `json:"storedCost"`
		EstimatedCost float64 `json:"estimatedCost"`
		Cost          float64 `json:"cost"`
		CostBasis     string  `json:"costBasis"`
	}
	active := make(map[string]bool)
	totals := leaderboardTotals{Tools: len(results), CostBasis: string(thermal.CostBasisTotal)}
	for _, r := range results {
		totals.RecordedCost += r.Summary.Cost
		totals.EstimatedCost += r.EstimatedCost
		for _, d := range r.Daily {
			if !thermal.IsActivityOnly(d) {
				totals.Tokens += d.Tokens
			}
			if d.Turns > 0 {
				active[d.Day] = true
			}
		}
	}
	totals.ActiveDays = len(active)
	totals.Cost = totals.RecordedCost + totals.EstimatedCost
	display := results
	if top > 0 && len(display) > top {
		display = display[:top]
	}
	return map[string]interface{}{
		"type":        "leaderboard",
		"generatedAt": now.UTC().Format(time.RFC3339),
		"results":     display,
		"totals":      totals,
	}
}
