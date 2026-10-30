// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sourceFixtures() []thermal.SourceInfo {
	items := make([]thermal.SourceInfo, 0, 15)
	states := []thermal.SourceState{
		thermal.SourceFound, thermal.SourceFound, thermal.SourceEmpty,
		thermal.SourceMissing, thermal.SourceMissing, thermal.SourceUnreadable,
		thermal.SourceWrongShape, thermal.SourceMissing, thermal.SourceFound,
		thermal.SourceFound, thermal.SourceEmpty, thermal.SourceMissing,
		thermal.SourceFound, thermal.SourceMissing, thermal.SourceFound,
	}
	for i := 0; i < 15; i++ {
		items = append(items, thermal.SourceInfo{
			Tool:        "tool-" + string(rune('a'+i)),
			Name:        "Tool " + string(rune('A'+i)),
			Path:        "/home/tester/.local/share/tool-" + string(rune('a'+i)) + "/session-store-with-a-long-name.db",
			DisplayPath: "~/.local/share/tool-" + string(rune('a'+i)) + "/session-store-with-a-long-name.db",
			Shape:       "file",
			State:       states[i],
			Detail:      "expected file",
		})
	}
	return items
}

func TestRenderSourcesFitsEightyColumns(t *testing.T) {
	out := RenderSources(sourceFixtures(), true)
	for _, line := range strings.Split(out, "\n") {
		if len(line) > 80 {
			t.Fatalf("sources line exceeds 80 columns (%d): %q", len(line), line)
		}
		if strings.Contains(line, "\x1b") {
			t.Fatalf("sources output contains ANSI: %q", line)
		}
	}
	for _, want := range []string{"Thermal · sources", "Summary:", "Legend:", "Next:", "missing", "wrong-shape"} {
		if !strings.Contains(out, want) {
			t.Fatalf("sources output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderSourcesJSON(t *testing.T) {
	now := time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)
	raw := RenderSourcesJSON(sourceFixtures(), "/home/tester", now)
	var payload struct {
		Type    string                `json:"type"`
		Home    string                `json:"home"`
		Summary thermal.SourceSummary `json:"summary"`
		Sources []thermal.SourceInfo  `json:"sources"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, raw)
	}
	if payload.Type != "sources" || payload.Home != "/home/tester" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if len(payload.Sources) != 15 || payload.Summary.Total != 15 {
		t.Fatalf("sources=%d summary=%+v", len(payload.Sources), payload.Summary)
	}
	if !strings.Contains(raw, `"generatedAt": "2026-10-30T00:00:00Z"`) {
		t.Fatalf("missing deterministic generatedAt:\n%s", raw)
	}
}

func TestSanitizeText(t *testing.T) {
	cases := map[string]string{
		"new\nline":   "new\\nline",
		"tab\there":   "tab\\there",
		"esc\x1b[31m": `esc\x1b[31m`,
	}
	for in, want := range cases {
		if got := SanitizeText(in); got != want {
			t.Fatalf("SanitizeText(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestSourcesNarrowWidth(t *testing.T) {
	SetWidth(40)
	defer SetWidth(0)
	items := []thermal.SourceInfo{{
		Name: "OpenCode", DisplayPath: "~/.local/share/opencode/opencode.db",
		State: thermal.SourceMissing, Detail: "expected file",
	}}
	out := RenderSources(items, true)
	for _, line := range strings.Split(out, "\n") {
		if len([]rune(line)) > 40 {
			t.Fatalf("line exceeds 40 columns: %q", line)
		}
	}
}

func TestTruncateSourceHelpers(t *testing.T) {
	if got := truncateMiddle("abcdefghij", 6); got != "abc…ij" {
		t.Fatalf("truncateMiddle=%q", got)
	}
	if got := truncateSourceRight("abcdefghij", 6); got != "abcde…" {
		t.Fatalf("truncateSourceRight=%q", got)
	}
	if got := truncateMiddle("short", 10); got != "short" {
		t.Fatalf("truncateMiddle short=%q", got)
	}
}
