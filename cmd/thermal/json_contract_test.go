// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/share"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestJSONPrecision(t *testing.T) {
	payload := map[string]any{
		"cost":  1.9774259520000008,
		"pct":   96.77737513675537,
		"share": 0.5038565852808624,
		"mean":  396960162.5520833,
	}
	b, err := marshalJSON(payload)
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	var decoded map[string]float64
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for key, value := range decoded {
		text := strconv.FormatFloat(value, 'f', -1, 64)
		parts := strings.SplitN(text, ".", 2)
		if len(parts) == 2 && len(parts[1]) > jsonPrecision {
			t.Errorf("%s serialized with %d decimals: %s", key, len(parts[1]), text)
		}
	}
	for key, want := range map[string]float64{"cost": 1.9774, "pct": 96.7774, "share": 0.5039, "mean": 396960162.5521} {
		if math.Abs(decoded[key]-want) > 1e-9 {
			t.Errorf("%s = %v, want %v", key, decoded[key], want)
		}
	}
}

func TestJSONNoAliasKeys(t *testing.T) {
	result := thermal.ToolResult{
		Tool:    thermal.ToolOpenCode,
		Name:    "OpenCode",
		Summary: thermal.Summary{Tool: "OpenCode", LifetimeTokens: 10},
	}
	b, err := marshalJSON(map[string]any{"results": []thermal.ToolResult{result}})
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	var decoded struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, alias := range []string{"Tool", "Name", "Summary"} {
		if _, ok := decoded.Results[0][alias]; ok {
			t.Errorf("alias key %q remains in leaderboard result", alias)
		}
	}
	for _, canonical := range []string{"tool", "name", "summary"} {
		if _, ok := decoded.Results[0][canonical]; !ok {
			t.Errorf("canonical key %q missing", canonical)
		}
	}
}

func TestJSONEnvelopeContract(t *testing.T) {
	payload := leaderboardJSONPayload([]thermal.ToolResult{{
		Tool:          thermal.ToolOpenCode,
		Name:          "OpenCode",
		Summary:       thermal.Summary{LifetimeTokens: 100},
		EstimatedCost: 0.5,
		Daily:         []thermal.DailyRow{{Day: "2026-01-01", Tokens: 100, Turns: 1}},
	}}, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	b, err := marshalJSON(payload)
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded["type"] != "leaderboard" {
		t.Errorf("type = %v, want leaderboard", decoded["type"])
	}
	if decoded["generatedAt"] == "" || decoded["results"] == nil || decoded["totals"] == nil {
		t.Errorf("leaderboard envelope missing keys: %v", decoded)
	}
	totals := decoded["totals"].(map[string]any)
	for _, key := range []string{"tools", "activeDays", "totalTokens", "storedCost", "estimatedCost", "cost", "costBasis"} {
		if _, ok := totals[key]; !ok {
			t.Errorf("leaderboard totals missing %q", key)
		}
	}
}

func TestShortenShareURL(t *testing.T) {
	long := "https://thermal.example/share?v=v1." + strings.Repeat("x", 200)
	short := shortenURL(long, 96)
	if len([]rune(short)) != 96 || !strings.Contains(short, "…") {
		t.Errorf("shortenURL length/ellipsis = %d/%v", len([]rune(short)), short)
	}
	if shortenURL("short", 96) != "short" {
		t.Errorf("short URL was changed")
	}
}

func TestJSONShareFields(t *testing.T) {
	snap := share.BuildShareSnapshot("all", []thermal.DailyRow{
		{Day: "2026-01-01", Tokens: 100, Turns: 1},
	}, 1.25, 2.5, "all time (1 active days)")
	b, err := marshalJSON(map[string]any{
		"type":        "share",
		"generatedAt": "2026-01-02T00:00:00Z",
		"snapshot":    snap,
	})
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	var decoded struct {
		Type     string                     `json:"type"`
		Snapshot map[string]json.RawMessage `json:"snapshot"`
	}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Type != "share" {
		t.Errorf("type = %q", decoded.Type)
	}
	for _, key := range []string{"scope", "window", "streakBasis", "activeDays", "recordedCost", "estimatedCost", "totalCost", "coverage"} {
		raw, ok := decoded.Snapshot[key]
		if !ok || string(raw) == "null" {
			t.Errorf("share snapshot field %q missing or null", key)
		}
	}
}
