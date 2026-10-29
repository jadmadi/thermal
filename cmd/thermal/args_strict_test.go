// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestParseArgsStrictValid(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		check func(*testing.T, thermal.Options)
	}{
		{"last after positional", []string{"daily", "--last", "7"}, func(t *testing.T, o thermal.Options) {
			if o.Report != "daily" || o.Last != 7 || !o.LastSet {
				t.Fatalf("got report=%q last=%d set=%v", o.Report, o.Last, o.LastSet)
			}
		}},
		{"last before positional", []string{"--last", "7", "daily"}, func(t *testing.T, o thermal.Options) {
			if o.Report != "daily" || o.Last != 7 {
				t.Fatalf("got report=%q last=%d", o.Report, o.Last)
			}
		}},
		{"tool report top", []string{"opencode", "projects", "--top", "5"}, func(t *testing.T, o thermal.Options) {
			if o.Tool != "opencode" || o.Report != "projects" || o.Top != 5 || !o.TopSet {
				t.Fatalf("got tool=%q report=%q top=%d set=%v", o.Tool, o.Report, o.Top, o.TopSet)
			}
		}},
		{"equal positional and flag tool", []string{"opencode", "--tool", "opencode", "daily"}, func(t *testing.T, o thermal.Options) {
			if o.Tool != "opencode" || o.Report != "daily" {
				t.Fatalf("got tool=%q report=%q", o.Tool, o.Report)
			}
		}},
		{"weeks inline", []string{"daily", "--weeks=12"}, func(t *testing.T, o thermal.Options) {
			if o.Weeks != 12 {
				t.Fatalf("weeks=%d", o.Weeks)
			}
		}},
		{"duration unit", []string{"live", "--interval", "500ms"}, func(t *testing.T, o thermal.Options) {
			if o.Interval != 500*time.Millisecond {
				t.Fatalf("interval=%v", o.Interval)
			}
		}},
		{"duration bare seconds", []string{"live", "--interval", "1.5"}, func(t *testing.T, o thermal.Options) {
			if o.Interval != 1500*time.Millisecond {
				t.Fatalf("interval=%v", o.Interval)
			}
		}},
		{"compact ceiling", []string{"live", "--flame-ceiling", "9M"}, func(t *testing.T, o thermal.Options) {
			if o.FlameCeiling != 9_000_000 || !o.FlameCeilingSet {
				t.Fatalf("ceiling=%v set=%v", o.FlameCeiling, o.FlameCeilingSet)
			}
		}},
		{"json flag", []string{"--json", "daily"}, func(t *testing.T, o thermal.Options) {
			if !o.JSON || o.Format != "json" {
				t.Fatalf("json=%v format=%q", o.JSON, o.Format)
			}
		}},
		{"verbose short", []string{"-v", "daily"}, func(t *testing.T, o thermal.Options) {
			if !o.Verbose {
				t.Fatal("verbose not set")
			}
		}},
		{"compact dates", []string{"daily", "--since", "20261001", "--until", "2026-10-31"}, func(t *testing.T, o thermal.Options) {
			if o.Since != "20261001" || o.Until != "2026-10-31" {
				t.Fatalf("since=%q until=%q", o.Since, o.Until)
			}
		}},
		{"top zero", []string{"projects", "--top", "0"}, func(t *testing.T, o thermal.Options) {
			if o.Top != 0 || !o.TopSet {
				t.Fatalf("top=%d set=%v", o.Top, o.TopSet)
			}
		}},
		{"double dash", []string{"--", "daily"}, func(t *testing.T, o thermal.Options) {
			if o.Report != "daily" {
				t.Fatalf("report=%q", o.Report)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, err := parseArgsFrom(tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tc.check(t, opts)
		})
	}
}

func TestParseArgsStrictInvalid(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"last text", []string{"daily", "--last", "abc"}, []string{"--last", "whole number"}},
		{"last float", []string{"daily", "--last", "1.5"}, []string{"--last", "whole number"}},
		{"last zero", []string{"daily", "--last", "0"}, []string{"--last", "at least 1"}},
		{"last negative", []string{"daily", "--last", "-1"}, []string{"--last", "at least 1"}},
		{"last missing", []string{"daily", "--last", "--json"}, []string{"--last", "needs a value"}},
		{"weeks text", []string{"daily", "--weeks", "abc"}, []string{"--weeks", "between 4 and 104"}},
		{"weeks range", []string{"daily", "--weeks", "3"}, []string{"--weeks", "between 4 and 104"}},
		{"top text", []string{"daily", "--top", "xyz"}, []string{"--top", "whole number"}},
		{"top negative", []string{"daily", "--top", "-1"}, []string{"--top", "cannot be negative"}},
		{"interval text", []string{"live", "--interval", "abc"}, []string{"--interval", "duration"}},
		{"interval zero", []string{"live", "--interval", "0"}, []string{"--interval", "positive"}},
		{"ceiling text", []string{"live", "--flame-ceiling", "abc"}, []string{"--flame-ceiling", "positive number"}},
		{"ceiling zero", []string{"live", "--flame-ceiling", "0"}, []string{"--flame-ceiling", "positive number"}},
		{"unknown flag", []string{"daily", "--metirc", "cost"}, []string{"unknown flag --metirc", "did you mean --metric?"}},
		{"empty tool", []string{"daily", "--tool="}, []string{"--tool", "non-empty"}},
		{"bad date", []string{"daily", "--since", "2026-13-01"}, []string{"--since", "YYYY-MM-DD"}},
		{"reversed dates", []string{"daily", "--since", "2026-10-29", "--until", "2026-10-01"}, []string{"is after --until", "did you mean"}},
		{"tool conflict", []string{"opencode", "--tool", "codex", "daily"}, []string{"conflicts with --tool", "opencode"}},
		{"third positional", []string{"daily", "extra", "third"}, []string{"unexpected argument", "third"}},
		{"bad order", []string{"daily", "--order", "sideways"}, []string{"--order", "asc or desc"}},
		{"bad weekday", []string{"daily", "--start-of-week", "funday"}, []string{"--start-of-week", "weekday"}},
		{"retired format", []string{"daily", "--format", "md"}, []string{"retired"}},
		{"top on stats", []string{"stats", "--top", "5"}, []string{"--top does not apply"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseArgsFrom(tc.args)
			if err == nil {
				t.Fatalf("expected an error for %v", tc.args)
			}
			var ue *usageError
			if !errors.As(err, &ue) {
				t.Fatalf("expected *usageError, got %T: %v", err, err)
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q does not contain %q", err.Error(), want)
				}
			}
		})
	}
}

func TestParseArgsLastWins(t *testing.T) {
	opts, err := parseArgsFrom([]string{"daily", "--last", "7", "--last", "3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Last != 3 {
		t.Fatalf("last=%d, want 3 (last occurrence wins)", opts.Last)
	}
}

func TestNearestName(t *testing.T) {
	if got, ok := nearestName("metirc", []string{"metric", "models", "sort"}); !ok || got != "metric" {
		t.Fatalf("nearestName(metirc) = %q, %v", got, ok)
	}
	if _, ok := nearestName("zz", []string{"metric", "sort"}); ok {
		t.Fatal("short token should not get a suggestion")
	}
}

func FuzzParseArgsStrict(f *testing.F) {
	for _, seed := range []string{
		"daily --last 7", "daily --last abc", "opencode --tool codex daily",
		"--json daily", "live --interval 1.5", "live --flame-ceiling 9M",
		"daily --since 2026-10-29 --until 2026-10-01", "-- top",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		args := strings.Fields(raw)
		_, err := parseArgsFrom(args)
		if err == nil {
			return
		}
		var ue *usageError
		var re *retiredError
		if !errors.As(err, &ue) && !errors.As(err, &re) && !errors.Is(err, errHelpRequested) {
			t.Fatalf("untyped error for %q: %T %v", raw, err, err)
		}
	})
}
