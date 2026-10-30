// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"strings"
	"testing"
)

func TestResolveHelpTarget(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		verb    string
		tool    string
		unknown string
		help    bool
	}{
		{"global short", []string{"-h"}, "", "", "", true},
		{"global long", []string{"--help"}, "", "", "", true},
		{"global help word", []string{"help"}, "", "", "", true},
		{"scoped after verb", []string{"daily", "--help"}, "daily", "", "", true},
		{"scoped before verb", []string{"--help", "daily"}, "daily", "", "", true},
		{"help command", []string{"help", "daily"}, "daily", "", "", true},
		{"live no tui", []string{"live", "--help"}, "live", "", "", true},
		{"upgrade safe", []string{"upgrade", "--help"}, "upgrade", "", "", true},
		{"utility audit", []string{"audit", "--help"}, "audit", "", "", true},
		{"utility version", []string{"version", "--help"}, "version", "", "", true},
		{"tool page", []string{"opencode", "--help"}, "", "opencode", "", true},
		{"tool and verb", []string{"opencode", "weekly", "--help"}, "weekly", "opencode", "", true},
		{"verb then tool", []string{"weekly", "opencode", "--help"}, "weekly", "opencode", "", true},
		{"help tool", []string{"help", "opencode"}, "", "opencode", "", true},
		{"unknown help", []string{"help", "nope"}, "", "", "nope", true},
		{"unknown with flag", []string{"nope", "--help"}, "", "", "nope", true},
		{"typo command", []string{"today", "--help"}, "", "", "today", true},
		{"retired", []string{"dashboard", "--help"}, "dashboard", "", "", true},
		{"statusline alias prompt", []string{"prompt", "--help"}, "prompt", "", "", true},
		{"statusline alias status", []string{"status", "--help"}, "status", "", "", true},
		{"flags before help", []string{"--json", "daily", "--help"}, "daily", "", "", true},
		{"value flag skipped", []string{"daily", "--help", "--since", "not-a-date"}, "daily", "", "", true},
		{"extra token ignored", []string{"daily", "--help", "extra"}, "daily", "", "", true},
		{"help wins over unknown flag", []string{"--help", "--bogus"}, "", "", "", true},
		{"help help", []string{"help", "help"}, "help", "", "", true},
		{"no help", []string{"daily"}, "", "", "", false},
		{"no help flags", []string{"--json", "daily"}, "", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verb, tool, unknown, help := resolveHelpTarget(tc.args)
			if verb != tc.verb || tool != tc.tool || unknown != tc.unknown || help != tc.help {
				t.Fatalf("got (%q,%q,%q,%v), want (%q,%q,%q,%v)", verb, tool, unknown, help, tc.verb, tc.tool, tc.unknown, tc.help)
			}
		})
	}
}

func TestHelpPagesFitEightyColumns(t *testing.T) {
	verbs := []string{"daily", "weekly", "monthly", "projects", "models", "stats", "mix", "trend", "live", "statusline", "share", "audit", "changelog", "upgrade", "version", "license", "help"}
	for _, verb := range verbs {
		t.Run(verb, func(t *testing.T) {
			out, code := helpOutput(verb, "", "", true)
			if code != 0 {
				t.Fatalf("helpOutput(%q) code=%d", verb, code)
			}
			if !strings.Contains(out, "Usage:") {
				t.Fatalf("%s page has no Usage section", verb)
			}
			for _, line := range strings.Split(out, "\n") {
				if len(line) > 80 {
					t.Fatalf("%s page line exceeds 80 columns: %q", verb, line)
				}
				if strings.Contains(line, "\x1b") {
					t.Fatalf("%s page contains ANSI: %q", verb, line)
				}
			}
		})
	}
}

func TestHelpPageSpecifics(t *testing.T) {
	daily, _ := helpOutput("daily", "", "", true)
	for _, want := range []string{"--since", "--until", "--last", "--breakdown", "--chart", "--json", "JSON:", "See also:"} {
		if !strings.Contains(daily, want) {
			t.Fatalf("daily help missing %q", want)
		}
	}
	if strings.Contains(daily, "Daily Usage") {
		t.Fatal("daily help unexpectedly contains a data table")
	}

	live, _ := helpOutput("live", "", "", true)
	for _, want := range []string{"--interval", "--flame-ceiling", "--stream", "--json"} {
		if !strings.Contains(live, want) {
			t.Fatalf("live help missing %q", want)
		}
	}

	upgrade, _ := helpOutput("upgrade", "", "", true)
	if !strings.Contains(upgrade, "Self-upgrade") {
		t.Fatal("upgrade help missing summary")
	}
}

func TestHelpAliasesAndTools(t *testing.T) {
	prompt, code := helpOutput("prompt", "", "", true)
	if code != 0 {
		t.Fatalf("prompt help code=%d", code)
	}
	if !strings.Contains(prompt, "thermal prompt [options]") {
		t.Fatalf("prompt help did not keep the alias in Usage:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Alias: prompt") {
		t.Fatalf("prompt help missing alias note:\n%s", prompt)
	}

	tool, code := helpOutput("", "opencode", "", true)
	if code != 0 {
		t.Fatalf("tool help code=%d", code)
	}
	if !strings.Contains(tool, "thermal opencode [report] [options]") {
		t.Fatalf("tool help has wrong usage:\n%s", tool)
	}

	scoped, code := helpOutput("weekly", "opencode", "", true)
	if code != 0 {
		t.Fatalf("scoped help code=%d", code)
	}
	if !strings.Contains(scoped, "thermal opencode weekly [options]") {
		t.Fatalf("scoped help has wrong usage:\n%s", scoped)
	}
}

func TestGlossaryHelp(t *testing.T) {
	for _, term := range []string{"streak", "token", "warriors", "activity", "glossary"} {
		out, code := helpOutput("", "", term, true)
		if code != 0 {
			t.Fatalf("help %s code=%d\n%s", term, code, out)
		}
		if !strings.Contains(strings.ToLower(out), term) {
			t.Fatalf("help %s missing term:\n%s", term, out)
		}
	}
	if out, code := helpOutput("", "", "doctor", true); code != 2 || !strings.Contains(out, "thermal sources") {
		t.Fatalf("doctor help code=%d\n%s", code, out)
	}
}

func TestHelpUnknownAndRetired(t *testing.T) {
	unknown, code := helpOutput("", "", "today", true)
	if code != 2 {
		t.Fatalf("unknown help code=%d", code)
	}
	if !strings.Contains(unknown, `unknown command "today"`) || !strings.Contains(unknown, "did you mean: thermal daily?") {
		t.Fatalf("unknown help missing suggestion:\n%s", unknown)
	}
	if strings.Contains(unknown, "unknown tool") {
		t.Fatalf("unknown help blamed a tool:\n%s", unknown)
	}

	retired, code := helpOutput("dashboard", "", "", true)
	if code != 1 {
		t.Fatalf("retired help code=%d", code)
	}
	if !strings.Contains(retired, "retired") || !strings.Contains(retired, "See also: thermal live") {
		t.Fatalf("retired help missing See also:\n%s", retired)
	}
}

func TestHelpHelpPage(t *testing.T) {
	out, code := helpOutput("help", "", "", true)
	if code != 0 {
		t.Fatalf("help-help code=%d", code)
	}
	if !strings.Contains(out, "thermal help [command]") || !strings.Contains(out, "thermal <command> --help") {
		t.Fatalf("help-help missing both forms:\n%s", out)
	}
}
