// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAnalyticsJSONBasis(t *testing.T) {
	statsOut, statsErr, code := runSim(t, "stats", "--json", "--offline")
	if code != 0 {
		t.Fatalf("stats --json exit=%d stderr=%s", code, statsErr)
	}
	var stats map[string]any
	if err := json.Unmarshal([]byte(statsOut), &stats); err != nil {
		t.Fatalf("stats JSON: %v\n%s", err, statsOut)
	}
	if stats["meanBasis"] != "active-day" {
		t.Fatalf("stats meanBasis=%v", stats["meanBasis"])
	}

	trendOut, trendErr, code := runSim(t, "trend", "--json", "--offline")
	if code != 0 {
		t.Fatalf("trend --json exit=%d stderr=%s", code, trendErr)
	}
	var trend map[string]any
	if err := json.Unmarshal([]byte(trendOut), &trend); err != nil {
		t.Fatalf("trend JSON: %v\n%s", err, trendOut)
	}
	if trend["meanBasis"] != "calendar-day" {
		t.Fatalf("trend meanBasis=%v", trend["meanBasis"])
	}
	if _, ok := trend["activeDays"].(float64); !ok {
		t.Fatalf("trend activeDays is not a number: %v", trend["activeDays"])
	}
	if projection, ok := trend["projection"].(map[string]any); ok {
		if projection["bandMethod"] != "residual-sd-1.96" {
			t.Fatalf("trend bandMethod=%v", projection["bandMethod"])
		}
	}
}

func TestAnalyticsTextBasis(t *testing.T) {
	statsOut, _, code := runSim(t, "stats", "--no-color", "--offline")
	if code != 0 {
		t.Fatalf("stats exit=%d", code)
	}
	if !strings.Contains(statsOut, "Mean / active day") || !strings.Contains(statsOut, "Basis:") {
		t.Fatalf("stats basis labels missing:\n%s", statsOut)
	}

	trendOut, _, code := runSim(t, "trend", "--no-color", "--offline")
	if code != 0 {
		t.Fatalf("trend exit=%d", code)
	}
	if !strings.Contains(trendOut, "Mean / calendar day") || !strings.Contains(trendOut, "Basis:") {
		t.Fatalf("trend basis labels missing:\n%s", trendOut)
	}
}
