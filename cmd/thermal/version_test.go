// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestResolveVersionTarget(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"short version", []string{"-V"}, true},
		{"long version", []string{"--version"}, true},
		{"version command", []string{"version"}, true},
		{"short verbose is not version", []string{"-v"}, false},
		{"long verbose is not version", []string{"--verbose"}, false},
		{"verbose daily", []string{"-v", "daily"}, false},
		{"daily verbose", []string{"daily", "-v"}, false},
		{"version after positional", []string{"daily", "-V"}, true},
		{"version before positional", []string{"-V", "daily"}, true},
		{"version beats verbose", []string{"-v", "-V"}, true},
		{"version beats validation", []string{"-V", "--bogus"}, true},
		{"json before version", []string{"--json", "version"}, true},
		{"double dash version", []string{"--", "version"}, true},
		{"double dash short flag is positional", []string{"--", "-V"}, false},
		{"none", []string{}, false},
		{"normal report", []string{"daily", "--last", "7"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveVersionTarget(tc.args); got != tc.want {
				t.Fatalf("resolveVersionTarget(%v)=%v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestVersionOutput(t *testing.T) {
	text := versionOutput(false)
	if !strings.Contains(text, "thermal ") {
		t.Fatalf("text version missing prefix: %q", text)
	}
	if strings.Contains(text, "\x1b") {
		t.Fatal("text version contains ANSI")
	}

	raw := versionOutput(true)
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("version JSON invalid: %v\n%s", err, raw)
	}
	if payload["type"] != "version" {
		t.Fatalf("type=%v", payload["type"])
	}
	for _, key := range []string{"version", "commit", "date", "generatedAt"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("version JSON missing %q", key)
		}
	}
}

func TestVerboseIsNotVersion(t *testing.T) {
	opts, err := parseArgsFrom([]string{"-v", "--json"})
	if err != nil {
		t.Fatalf("parseArgsFrom(-v --json): %v", err)
	}
	if !opts.Verbose || !opts.JSON {
		t.Fatalf("verbose=%v json=%v", opts.Verbose, opts.JSON)
	}
	if resolveVersionTarget([]string{"-v", "--json"}) {
		t.Fatal("-v --json was treated as a version request")
	}
}

func TestVersionFlagParserDefense(t *testing.T) {
	for _, args := range [][]string{{"-V"}, {"--version"}, {"daily", "-V"}} {
		_, err := parseArgsFrom(args)
		if !errors.Is(err, errVersionRequested) {
			t.Fatalf("parseArgsFrom(%v) error=%v, want errVersionRequested", args, err)
		}
	}
}

func TestVersionJSONParity(t *testing.T) {
	strip := func(raw string) map[string]any {
		var m map[string]any
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("json: %v", err)
		}
		delete(m, "generatedAt")
		return m
	}
	a := strip(versionOutput(true))
	b := strip(versionOutput(true))
	if len(a) != len(b) {
		t.Fatalf("key count differs: %d vs %d", len(a), len(b))
	}
	for key := range a {
		av, bv := a[key], b[key]
		if av != bv {
			t.Fatalf("%s differs: %v vs %v", key, av, bv)
		}
	}
}

func TestVersionHelpPage(t *testing.T) {
	out, code := helpOutput("version", "", "", true)
	if code != 0 {
		t.Fatalf("version help code=%d", code)
	}
	for _, want := range []string{"-V, --version", "Short forms", "-v flag means"} {
		if !strings.Contains(out, want) {
			t.Fatalf("version help missing %q:\n%s", want, out)
		}
	}
}
