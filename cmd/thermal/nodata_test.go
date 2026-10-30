// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestNoDataOnboarding(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	printNoDataTo(&stdout, &stderr, thermal.Options{}, "")
	out := stdout.String()
	for _, want := range []string{
		"No AI coding sessions found yet.",
		"Run one prompt in OpenCode, Claude, Codex, or any supported tool",
		"thermal sources",
		"thermal --help",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("onboarding missing %q:\n%s", want, out)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
	if lines := strings.Count(strings.TrimRight(out, "\n"), "\n") + 1; lines > 12 {
		t.Fatalf("onboarding has %d lines, want <=12", lines)
	}
}

func TestNoDataLeaderboardEnvelope(t *testing.T) {
	payload := markNoDataPayload(leaderboardJSONPayloadWithMeta(nil,
		time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC), "", true, 0))
	raw, err := marshalJSON(payload)
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	var decoded struct {
		Type    string           `json:"type"`
		State   string           `json:"state"`
		Results []map[string]any `json:"results"`
		Next    *noDataNext      `json:"next"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Type != "leaderboard" || decoded.State != "no-data" {
		t.Fatalf("type/state = %q/%q", decoded.Type, decoded.State)
	}
	if decoded.Results == nil || len(decoded.Results) != 0 {
		t.Fatalf("results = %#v, want empty array", decoded.Results)
	}
	if decoded.Next == nil || decoded.Next.Inventory == "" || decoded.Next.Help == "" {
		t.Fatalf("next = %#v", decoded.Next)
	}
}

func TestNoDataReportArrays(t *testing.T) {
	stats := thermal.AggregateStats(nil, thermal.StatsOptions{}, nil)
	if stats.Weekday == nil || stats.TopDays == nil || stats.Outliers == nil || stats.Histogram == nil {
		t.Fatalf("stats arrays must be non-nil: %#v", stats)
	}
	trend := thermal.AggregateTrend(nil, thermal.TrendOptions{}, nil)
	if trend.Points == nil {
		t.Fatal("trend points must be non-nil")
	}
}

func TestRequireDataFlagParses(t *testing.T) {
	opts, err := parseArgsFrom([]string{"daily", "--require-data"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !opts.RequireData {
		t.Fatal("--require-data did not set RequireData")
	}
}

func TestNoDataNextInfo(t *testing.T) {
	next := noDataNextInfo()
	if next.Inventory != "thermal sources --json" || next.Help != "thermal --help" {
		t.Fatalf("next = %#v", next)
	}
}
