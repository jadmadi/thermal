// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestDBOverrideRequiresOneTool(t *testing.T) {
	cases := [][]string{
		{"daily", "--db", "/tmp"},
		{"--tool", "all", "daily", "--db", "/tmp"},
		{"--tool", "auto", "daily", "--db", "/tmp"},
		{"live", "--db", "/tmp"},
		{"statusline", "--db", "/tmp"},
		{"--db", "/tmp", "daily"},
		{"--db=/tmp", "projects"},
	}
	for _, args := range cases {
		_, err := parseArgsFrom(args)
		if err == nil {
			t.Fatalf("%v: expected a scope error", args)
		}
		var ue *usageError
		if !errors.As(err, &ue) {
			t.Fatalf("%v: expected *usageError, got %T", args, err)
		}
		if !strings.Contains(err.Error(), "applies to one tool") {
			t.Fatalf("%v: unexpected message: %v", args, err)
		}
	}
}

func TestDBOverrideAllowsSingleTool(t *testing.T) {
	cases := [][]string{
		{"opencode", "daily", "--db", "/tmp"},
		{"--tool", "opencode", "daily", "--db", "/tmp"},
		{"daily", "opencode", "--db", "/tmp"},
		{"--tool", "muse", "daily", "--db", "/tmp"},
		{"--tool", "opencode", "--db=/tmp", "daily"},
		{"opencode", "--db", "/tmp", "daily"},
		{"--tool", "opencode", "live", "--db", "/tmp"},
	}
	for _, args := range cases {
		if _, err := parseArgsFrom(args); err != nil {
			t.Fatalf("%v: unexpected error: %v", args, err)
		}
	}
}
