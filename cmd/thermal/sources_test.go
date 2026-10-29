// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSourcesCommandParsing(t *testing.T) {
	opts, err := parseArgsFrom([]string{"sources"})
	if err != nil {
		t.Fatalf("sources parse: %v", err)
	}
	if opts.Tool != "sources" {
		t.Fatalf("tool=%q, want sources", opts.Tool)
	}

	opts, err = parseArgsFrom([]string{"sources", "--json"})
	if err != nil {
		t.Fatalf("sources --json parse: %v", err)
	}
	if !opts.JSON {
		t.Fatal("sources --json did not set JSON")
	}
}

func TestSourcesRejectsTopAndDB(t *testing.T) {
	for _, args := range [][]string{
		{"sources", "--top", "5"},
		{"sources", "--db", "/tmp"},
	} {
		_, err := parseArgsFrom(args)
		if err == nil {
			t.Fatalf("%v: expected an error", args)
		}
		var ue *usageError
		if !errors.As(err, &ue) {
			t.Fatalf("%v: expected *usageError, got %T", args, err)
		}
	}
}

func TestSourcesHelp(t *testing.T) {
	out, code := helpOutput("sources", "", "", true)
	if code != 0 {
		t.Fatalf("sources help code=%d", code)
	}
	for _, want := range []string{"thermal sources [options]", "--json", "stat-only"} {
		if !strings.Contains(strings.ToLower(out), strings.ToLower(want)) {
			t.Fatalf("sources help missing %q:\n%s", want, out)
		}
	}
}
