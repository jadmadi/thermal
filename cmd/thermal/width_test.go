// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"strings"
	"testing"
)

func TestLiveWidthRefusal(t *testing.T) {
	if err := liveWidthError(59); err == nil {
		t.Fatal("59 columns should be rejected")
	} else if !strings.Contains(err.Error(), "at least 60 columns") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := liveWidthError(60); err != nil {
		t.Fatalf("60 columns should be accepted: %v", err)
	}
	if err := liveWidthError(80); err != nil {
		t.Fatalf("80 columns should be accepted: %v", err)
	}
	if err := liveWidthError(120); err != nil {
		t.Fatalf("120 columns should be accepted: %v", err)
	}
}

func TestWidthPrecedence(t *testing.T) {
	opts, err := parseArgsFrom([]string{"--width", "60", "daily"})
	if err != nil {
		t.Fatalf("--width 60: %v", err)
	}
	if opts.Width != 60 || !opts.WidthSet {
		t.Fatalf("width=%d set=%v", opts.Width, opts.WidthSet)
	}
	opts, err = parseArgsFrom([]string{"daily", "--width=100"})
	if err != nil {
		t.Fatalf("--width=100: %v", err)
	}
	if opts.Width != 100 || !opts.WidthSet {
		t.Fatalf("width=%d set=%v", opts.Width, opts.WidthSet)
	}
	for _, args := range [][]string{{"--width", "39", "daily"}, {"--width=241", "daily"}} {
		if _, err := parseArgsFrom(args); err == nil {
			t.Fatalf("%v: expected a range error", args)
		}
	}
	if got, ok := scanWidth([]string{"--width", "60", "daily"}); !ok || got != 60 {
		t.Fatalf("scanWidth space=%d,%v", got, ok)
	}
	if got, ok := scanWidth([]string{"--width=100", "daily"}); !ok || got != 100 {
		t.Fatalf("scanWidth equals=%d,%v", got, ok)
	}
	if _, ok := scanWidth([]string{"--width", "39"}); ok {
		t.Fatal("scanWidth accepted 39")
	}
}

func TestCommandMinimumWidths(t *testing.T) {
	cases := []struct {
		report, tool, command string
		min                   int
	}{
		{"daily", "", "daily", 56},
		{"weekly", "", "weekly", 56},
		{"monthly", "", "monthly", 56},
		{"projects", "", "projects", 48},
		{"models", "", "models", 48},
		{"audit", "", "audit", 56},
		{"changelog", "", "changelog", 48},
		{"", "all", "leaderboard", 40},
		{"", "opencode", "opencode", 40},
	}
	for _, tc := range cases {
		command, min := commandMinWidth(tc.report, tc.tool)
		if command != tc.command || min != tc.min {
			t.Errorf("commandMinWidth(%q,%q)=(%q,%d), want (%q,%d)", tc.report, tc.tool, command, min, tc.command, tc.min)
		}
	}
}
