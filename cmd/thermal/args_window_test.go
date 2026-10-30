// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"strings"
	"testing"
)

func TestParseLastValue(t *testing.T) {
	cases := []struct {
		raw      string
		n        int
		unit     string
		explicit bool
		wantErr  string
	}{
		{"7", 7, "", false, ""},
		{"14d", 14, "day", true, ""},
		{"4w", 4, "week", true, ""},
		{"3m", 3, "month", true, ""},
		{"0", 0, "", false, "at least 1"},
		{"-1", 0, "", false, "at least 1"},
		{"1.5", 0, "", false, "whole number"},
		{"7x", 0, "", false, "d, w, or m"},
	}
	for _, tc := range cases {
		n, unit, explicit, err := parseLastValue(tc.raw)
		if tc.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("%q: err=%v, want %q", tc.raw, err, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", tc.raw, err)
		}
		if n != tc.n || unit != tc.unit || explicit != tc.explicit {
			t.Fatalf("%q: got (%d,%q,%v), want (%d,%q,%v)", tc.raw, n, unit, explicit, tc.n, tc.unit, tc.explicit)
		}
	}
}

func TestExplicitLastUsesSinceAndLabel(t *testing.T) {
	opts, err := parseArgsFrom([]string{"weekly", "--last", "14d"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !opts.LastExplicit || opts.LastUnit != "day" || opts.Last != 0 {
		t.Fatalf("opts = %+v", opts)
	}
	if opts.Since == "" || opts.WindowLabel != "last 14 days" {
		t.Fatalf("since/label = %q/%q", opts.Since, opts.WindowLabel)
	}
}

func TestRepeatedFlagWarnings(t *testing.T) {
	warnings := repeatedFlagWarnings([]string{"--tool", "claude", "--tool", "codewhale"})
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v", warnings)
	}
	for _, want := range []string{"--tool", "codewhale", "claude"} {
		if !strings.Contains(warnings[0], want) {
			t.Fatalf("warning %q missing %q", warnings[0], want)
		}
	}
	if got := repeatedFlagWarnings([]string{"--json", "--json"}); len(got) != 0 {
		t.Fatalf("boolean repeats warned: %v", got)
	}
}

func TestDoubleDashDiagnostic(t *testing.T) {
	msg := doubleDashDiagnostic([]string{"weekly", "--", "--offline"})
	if !strings.Contains(msg, `"--offline" was passed after --`) {
		t.Fatalf("msg = %q", msg)
	}
	if got := doubleDashDiagnostic([]string{"weekly", "--", "opencode"}); got != "" {
		t.Fatalf("unexpected diagnostic: %q", got)
	}
}
