// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"math"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestLeaderboardJSONRanking(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "OpenCode", TotalActivity: 1000,
			Summary:        thermal.Summary{Cost: 10.50, ModelBreakdown: map[string]int64{"mystery-model": 2500}},
			EstimatedCost:  2.25,
			MissingPricing: []string{"mystery-model"},
		},
	}
	payload := leaderboardJSONPayloadWithMeta(results, time.Now(), "tokens", true, 0)
	if payload["window"] != "all-time" {
		t.Fatalf("window=%v", payload["window"])
	}
	ranking, ok := payload["ranking"].(map[string]any)
	if !ok || ranking["key"] != "tokens" || ranking["order"] != "desc" {
		t.Fatalf("ranking=%v", payload["ranking"])
	}
	costBasis, ok := payload["costBasis"].(map[string]any)
	if !ok {
		t.Fatalf("costBasis=%v", payload["costBasis"])
	}
	recorded, _ := costBasis["recorded"].(float64)
	estimated, _ := costBasis["estimated"].(float64)
	shown, _ := costBasis["shown"].(float64)
	if math.Abs(recorded+estimated-shown) > 0.0001 {
		t.Fatalf("cost basis does not add up: %v + %v != %v", recorded, estimated, shown)
	}
	unpriced, ok := payload["unpricedModels"].([]map[string]any)
	if !ok || len(unpriced) != 1 || unpriced[0]["name"] != "mystery-model" {
		t.Fatalf("unpricedModels=%v", payload["unpricedModels"])
	}

	disabled := leaderboardJSONPayloadWithMeta(results, time.Now(), "", false, 0)
	cb := disabled["costBasis"].(map[string]any)
	if cb["estimationDisabled"] != true || cb["estimated"].(float64) != 0 {
		t.Fatalf("disabled costBasis=%v", cb)
	}
	rank := disabled["ranking"].(map[string]any)
	if rank["key"] != "streak" {
		t.Fatalf("default ranking=%v", rank)
	}
}
