// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"
)

func TestRenderHelp_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CLICOLOR_FORCE", "0")
	out := RenderHelp(true)

	if strings.Contains(out, "\033[") {
		t.Errorf("expected zero ANSI escape codes when noColor is true, got:\n%s", out)
	}

	// Verify required sections
	requiredSections := []string{
		"Thermal · usage profile",
		"Usage:",
		"Dashboards & Views:",
		"Reports & Rankings:",
		"FinOps & Analytics:",
		"Diagnostics & Utilities:",
		"Common Examples:",
		"Filter & Window Options:",
		"Ranking & Analytics Options:",
		"Display & Formatting Options:",
		"Global & Engine Options:",
		"Supported Tools:",
		"Token Warriors:",
		"Activity Hunters:",
	}

	for _, sec := range requiredSections {
		if !strings.Contains(out, sec) {
			t.Errorf("expected help output to contain %q", sec)
		}
	}

	// Verify key commands are listed
	commands := []string{
		"dashboard", "daily", "weekly", "monthly", "projects", "models",
		"mix", "stats", "trend", "replay", "yield", "receipt",
		"share", "web", "serve", "audit", "changelog", "upgrade", "version", "license",
	}

	for _, cmd := range commands {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected help output to list command %q", cmd)
		}
	}
}

func TestRenderHelp_WithColors(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "1")
	out := RenderHelp(false)

	if !strings.Contains(out, "\033[") {
		t.Errorf("expected ANSI escape codes when noColor is false and CLICOLOR_FORCE=1, got plain text")
	}

	if !strings.Contains(out, "Thermal") {
		t.Errorf("expected Thermal title in colored output")
	}
}

func TestRenderHelp_Alignment(t *testing.T) {
	out := RenderHelp(true)
	lines := strings.Split(out, "\n")

	for _, l := range lines {
		if strings.HasPrefix(l, "    dashboard") {
			idx := strings.Index(l, "Interactive")
			if idx != 19 {
				t.Errorf("expected 'Interactive' to start at col 19, got %d in line: %q", idx, l)
			}
		}
		if strings.HasPrefix(l, "    --tool <name>") {
			idx := strings.Index(l, "Filter")
			if idx != 27 {
				t.Errorf("expected 'Filter' to start at col 27, got %d in line: %q", idx, l)
			}
		}
		if strings.HasPrefix(l, "    thermal [options]") {
			idx := strings.Index(l, "All-tool")
			if idx != 41 {
				t.Errorf("expected 'All-tool' to start at col 41, got %d in line: %q", idx, l)
			}
		}
		if strings.HasPrefix(l, "    thermal opencode weekly --chart") {
			idx := strings.Index(l, "Weekly")
			if idx != 39 {
				t.Errorf("expected 'Weekly' to start at col 39, got %d in line: %q", idx, l)
			}
		}
	}
}
