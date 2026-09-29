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

func TestRenderStatusline(t *testing.T) {
	data := StatuslineData{
		Streak:        14,
		LongestStreak: 25,
		TodayTokens:   42500,
		TodayTurns:    18,
		ActiveTool:    "Codex",
		Day:           "2026-09-29",
		UpdatedAt:     time.Now().Unix(),
	}

	// 1. Default (Unicode flame, colored)
	out := RenderStatusline(data, StatuslineOptions{NoColor: true})
	if !strings.Contains(out, "🔥 14d · 42.5K tok · Codex") {
		t.Errorf("unexpected default statusline output: %q", out)
	}

	// 2. Nerd Font
	nerdOut := RenderStatusline(data, StatuslineOptions{Nerd: true, NoColor: true})
	if !strings.Contains(nerdOut, "󰈸 14d · 42.5K tok · Codex") {
		t.Errorf("unexpected nerd font statusline output: %q", nerdOut)
	}

	// 3. Plain (no emoji)
	plainOut := RenderStatusline(data, StatuslineOptions{Plain: true, NoColor: true})
	if !strings.Contains(plainOut, "14d · 42.5K tok · Codex") || strings.Contains(plainOut, "🔥") {
		t.Errorf("unexpected plain statusline output: %q", plainOut)
	}

	// 4. JSON
	jsonOut := RenderStatusline(data, StatuslineOptions{JSON: true})
	var parsed StatuslineData
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("failed unmarshaling json statusline: %v", err)
	}
	if parsed.Streak != 14 || parsed.TodayTokens != 42500 || parsed.ActiveTool != "Codex" {
		t.Errorf("parsed JSON mismatch: %+v", parsed)
	}

	// 5. Zero streak
	zeroData := StatuslineData{
		Streak:      0,
		TodayTokens: 0,
		Day:         "2026-09-29",
	}
	zeroOut := RenderStatusline(zeroData, StatuslineOptions{NoColor: true})
	if !strings.Contains(zeroOut, "💤 0d · 0 tok") {
		t.Errorf("unexpected zero streak statusline: %q", zeroOut)
	}
}

func TestStatuslineCache_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	today := thermal.LocalDay(time.Now())
	original := StatuslineData{
		Streak:        7,
		LongestStreak: 12,
		TodayTokens:   1500000,
		TodayTurns:    35,
		ActiveTool:    "Devin",
		Day:           today,
		UpdatedAt:     time.Now().Unix(),
	}

	if err := SaveStatuslineCache(tempDir, original); err != nil {
		t.Fatalf("failed saving cache: %v", err)
	}

	loaded, ok := LoadStatuslineCache(tempDir)
	if !ok {
		t.Fatalf("expected cache hit, got miss")
	}

	if loaded.Streak != original.Streak || loaded.TodayTokens != original.TodayTokens || loaded.ActiveTool != original.ActiveTool {
		t.Errorf("loaded data mismatch: got %+v, want %+v", loaded, original)
	}
}
