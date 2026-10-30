// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/loaders"
)

func TestIntentAndToolSuggestions(t *testing.T) {
	for _, token := range []string{"doctor", "config", "setup", "history", "usage"} {
		d := classifyToken(token, slotTool)
		if d.Message != "intent" || !strings.Contains(renderTokenDiagnostic(d), "thermal sources") {
			t.Fatalf("%s diagnostic = %+v", token, d)
		}
	}
	top := renderTokenDiagnostic(classifyToken("top", slotTool))
	if !strings.Contains(top, "--top") || strings.Contains(top, `"oc"`) {
		t.Fatalf("top diagnostic = %q", top)
	}
	open := renderTokenDiagnostic(classifyToken("open", slotTool))
	if !strings.Contains(open, `OpenCode (alias "oc")`) {
		t.Fatalf("open diagnostic = %q", open)
	}
	code := renderTokenDiagnostic(classifyToken("code", slotTool))
	for _, want := range []string{"Claude", "OpenCode", "CMD"} {
		if !strings.Contains(code, want) {
			t.Fatalf("code diagnostic missing %q: %q", want, code)
		}
	}
}

func TestClassifyToken(t *testing.T) {
	cases := []struct {
		name             string
		token            string
		slot             tokenSlot
		class            tokenClass
		suggestion       string
		suggestIsCommand bool
	}{
		{"today intent", "today", slotTool, classUnknown, "daily", true},
		{"daly typo", "daly", slotTool, classUnknown, "daily", true},
		{"weak typo", "weak", slotTool, classUnknown, "weekly", true},
		{"cost intent", "cost", slotTool, classUnknown, "stats", true},
		{"report in tool slot", "weekly", slotTool, classWrongSlot, "", false},
		{"stats in tool slot", "stats", slotTool, classWrongSlot, "", false},
		{"path", "/tmp", slotTool, classPath, "", false},
		{"home path", "~/data", slotTool, classPath, "", false},
		{"known tool", "opencode", slotTool, classKnown, "", false},
		{"known all", "all", slotTool, classKnown, "", false},
		{"retired", "dashboard", slotTool, classRetired, "", false},
		{"unknown tool", "zzzz", slotTool, classUnknown, "", false},
		{"known command", "daily", slotCommand, classKnown, "", false},
		{"empty", "", slotTool, classEmpty, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := classifyToken(tc.token, tc.slot)
			if d.Class != tc.class {
				t.Fatalf("class=%v, want %v (%+v)", d.Class, tc.class, d)
			}
			if d.Suggestion != tc.suggestion {
				t.Fatalf("suggestion=%q, want %q", d.Suggestion, tc.suggestion)
			}
			if d.SuggestIsCommand != tc.suggestIsCommand {
				t.Fatalf("suggestIsCommand=%v, want %v", d.SuggestIsCommand, tc.suggestIsCommand)
			}
		})
	}
}

func TestRenderTokenDiagnostic(t *testing.T) {
	today := renderTokenDiagnostic(classifyToken("today", slotTool))
	if !strings.Contains(today, `unknown command "today"`) {
		t.Fatalf("missing command label:\n%s", today)
	}
	if !strings.Contains(today, "did you mean: thermal daily?") {
		t.Fatalf("missing suggestion:\n%s", today)
	}
	if strings.Contains(today, "unknown tool") {
		t.Fatalf("command token blamed as a tool:\n%s", today)
	}
	if lines := strings.Count(today, "\n"); lines > 5 {
		t.Fatalf("diagnostic has %d lines:\n%s", lines, today)
	}

	nope := renderTokenDiagnostic(classifyToken("nope", slotTool))
	if !strings.Contains(nope, `unknown tool "nope"`) || !strings.Contains(nope, "Valid tools:") {
		t.Fatalf("unexpected unknown-tool diagnostic:\n%s", nope)
	}

	wrong := renderTokenDiagnostic(classifyToken("weekly", slotTool))
	if !strings.Contains(wrong, `"weekly" is a report command`) || !strings.Contains(wrong, "run thermal weekly") {
		t.Fatalf("unexpected wrong-slot diagnostic:\n%s", wrong)
	}

	path := renderTokenDiagnostic(classifyToken("/tmp", slotTool))
	if !strings.Contains(path, "looks like a path") || !strings.Contains(path, "--db") {
		t.Fatalf("unexpected path diagnostic:\n%s", path)
	}

	retired := renderTokenDiagnostic(classifyToken("dashboard", slotTool))
	if !strings.Contains(retired, "See also: thermal live") {
		t.Fatalf("unexpected retired diagnostic:\n%s", retired)
	}
}

func TestDiagnosticCandidateCoverage(t *testing.T) {
	for name := range commandHelps {
		if !isKnownCommandToken(name) {
			t.Errorf("command %q is not recognized", name)
		}
		found := false
		for _, candidate := range diagnosticCommandNames {
			if candidate == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("command %q missing from suggestion candidates", name)
		}
	}
	for _, entry := range toolHelpEntries {
		for _, name := range []string{entry.Canonical, entry.Alias} {
			if name == "" {
				continue
			}
			if _, ok := loaders.ResolveTool(name); !ok {
				t.Errorf("tool %q does not resolve", name)
			}
			found := false
			for _, candidate := range diagnosticToolNames {
				if candidate == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("tool %q missing from suggestion candidates", name)
			}
		}
	}
	for _, name := range diagnosticCommandNames {
		if !isKnownCommandToken(name) {
			t.Errorf("suggestion candidate %q is not a known command", name)
		}
	}
	for _, name := range diagnosticToolNames {
		if _, ok := loaders.ResolveTool(name); !ok {
			t.Errorf("suggestion candidate %q is not a known tool", name)
		}
	}
}
