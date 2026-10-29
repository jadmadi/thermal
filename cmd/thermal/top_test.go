// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestTopApplicability(t *testing.T) {
	accepted := [][]string{
		{"--top", "5"},
		{"daily", "--top", "5"},
		{"weekly", "--top", "5"},
		{"monthly", "--top", "5"},
		{"projects", "--top", "5"},
		{"models", "--top", "5"},
		{"changelog", "--top", "5"},
		{"changelog", "--top", "0"},
		{"daily", "--top", "0"},
	}
	for _, args := range accepted {
		if _, err := parseArgsFrom(args); err != nil {
			t.Fatalf("%v: unexpected error: %v", args, err)
		}
	}

	rejected := [][]string{
		{"stats", "--top", "5"},
		{"mix", "--top", "5"},
		{"trend", "--top", "5"},
		{"audit", "--top", "5"},
		{"share", "--top", "5"},
		{"live", "--top", "5"},
		{"statusline", "--top", "5"},
		{"stats", "--top", "0"},
	}
	for _, args := range rejected {
		_, err := parseArgsFrom(args)
		if err == nil {
			t.Fatalf("%v: expected an error", args)
		}
		var ue *usageError
		if !errors.As(err, &ue) {
			t.Fatalf("%v: expected *usageError, got %T", args, err)
		}
		for _, want := range []string{"does not apply", "leaderboard, daily, weekly, monthly, projects, models, changelog"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("%v: error %q missing %q", args, err.Error(), want)
			}
		}
	}
}

func TestTopAppliesTable(t *testing.T) {
	cases := []struct {
		report string
		tool   string
		want   bool
	}{
		{"", "all", true},
		{"", "opencode", true},
		{"", "changelog", true},
		{"", "audit", false},
		{"", "share", false},
		{"daily", "", true},
		{"weekly", "", true},
		{"monthly", "", true},
		{"projects", "", true},
		{"models", "", true},
		{"changelog", "", true},
		{"stats", "", false},
		{"mix", "", false},
		{"trend", "", false},
		{"live", "", false},
		{"statusline", "", false},
	}
	for _, tc := range cases {
		if got := topApplies(tc.report, tc.tool); got != tc.want {
			t.Errorf("topApplies(%q,%q)=%v, want %v", tc.report, tc.tool, got, tc.want)
		}
	}
}

func TestChangelogLimit(t *testing.T) {
	if got := changelogLimit(thermal.Options{}); got != 3 {
		t.Fatalf("default limit=%d, want 3", got)
	}
	if got := changelogLimit(thermal.Options{TopSet: true, Top: 0}); got != 0 {
		t.Fatalf("explicit --top 0 limit=%d, want 0", got)
	}
	if got := changelogLimit(thermal.Options{TopSet: true, Top: 5}); got != 5 {
		t.Fatalf("--top 5 limit=%d, want 5", got)
	}
	if got := changelogLimit(thermal.Options{Last: 7}); got != 7 {
		t.Fatalf("--last 7 limit=%d, want 7", got)
	}
}
