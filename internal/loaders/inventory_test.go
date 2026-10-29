// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func isolateInventoryEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"HERMES_HOME", "DSH_HOME", "CODEX_HOME", "OPENCODE_HOME", "ZED_HOME", "AGY_HOME", "XDG_DATA_HOME", "XDG_CONFIG_HOME"} {
		t.Setenv(key, "")
	}
}

func findSource(items []thermal.SourceInfo, tool thermal.Tool) thermal.SourceInfo {
	for _, item := range items {
		if item.Tool == string(tool) {
			return item
		}
	}
	return thermal.SourceInfo{}
}

func TestInventoryEmptyHome(t *testing.T) {
	isolateInventoryEnv(t)
	home := t.TempDir()
	items := Inventory(home)
	if len(items) != 15 {
		t.Fatalf("inventory has %d items, want 15", len(items))
	}
	for _, item := range items {
		if item.State != thermal.SourceMissing {
			t.Errorf("%s state=%s, want missing", item.Tool, item.State)
		}
	}
	if summary := thermal.SummarizeSources(items); summary.Missing != 15 || summary.Found != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestInventoryStableOrder(t *testing.T) {
	isolateInventoryEnv(t)
	items := Inventory(t.TempDir())
	if len(items) != len(InventoryOrder) {
		t.Fatalf("inventory length %d, order length %d", len(items), len(InventoryOrder))
	}
	for i, tool := range InventoryOrder {
		if items[i].Tool != string(tool) {
			t.Fatalf("inventory[%d]=%q, want %q", i, items[i].Tool, tool)
		}
	}
}

func TestInventoryStates(t *testing.T) {
	isolateInventoryEnv(t)
	home := t.TempDir()
	tools := AllToolsForHome(home)

	oc := tools[thermal.ToolOpenCode]
	if err := os.MkdirAll(filepath.Dir(oc.DBPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data := append([]byte("SQLite format 3\x00"), make([]byte, 512)...)
	if err := os.WriteFile(oc.DBPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	claude := tools[thermal.ToolClaude]
	if err := os.MkdirAll(claude.DataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	droid := tools[thermal.ToolDroid]
	if err := os.MkdirAll(filepath.Dir(droid.DataDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(droid.DataDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	items := Inventory(home)
	if got := findSource(items, thermal.ToolOpenCode); got.State != thermal.SourceFound {
		t.Fatalf("OpenCode state=%s, want found", got.State)
	}
	if got := findSource(items, thermal.ToolClaude); got.State != thermal.SourceEmpty {
		t.Fatalf("Claude state=%s, want empty", got.State)
	}
	if got := findSource(items, thermal.ToolDroid); got.State != thermal.SourceWrongShape {
		t.Fatalf("Droid state=%s, want wrong-shape", got.State)
	}
	if got := findSource(items, thermal.ToolDevin); got.State != thermal.SourceMissing {
		t.Fatalf("Devin state=%s, want missing", got.State)
	}
}

func TestInventoryEnvOverride(t *testing.T) {
	isolateInventoryEnv(t)
	home := t.TempDir()
	hermesHome := filepath.Join(t.TempDir(), "hermes-fixture")
	if err := os.MkdirAll(hermesHome, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERMES_HOME", hermesHome)
	items := Inventory(home)
	got := findSource(items, thermal.ToolHermes)
	if !strings.HasPrefix(got.Path, hermesHome) {
		t.Fatalf("Hermes path=%q, want prefix %q", got.Path, hermesHome)
	}
}

func TestInventoryPrivacy(t *testing.T) {
	isolateInventoryEnv(t)
	home := t.TempDir()
	tools := AllToolsForHome(home)
	claude := tools[thermal.ToolClaude]
	projects := filepath.Join(claude.DataDir, "projects", "secret-project")
	if err := os.MkdirAll(projects, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := "SENTINEL-PROMPT-TEXT"
	if err := os.WriteFile(filepath.Join(projects, "session.jsonl"), []byte(`{"prompt":"`+sentinel+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	items := Inventory(home)
	rendered := fmt.Sprintf("%+v", items)
	if strings.Contains(rendered, sentinel) || strings.Contains(rendered, "secret-project") {
		t.Fatalf("inventory leaked session content or project name: %s", rendered)
	}
}

func TestInventorySummary(t *testing.T) {
	items := []thermal.SourceInfo{
		{State: thermal.SourceFound},
		{State: thermal.SourceFound},
		{State: thermal.SourceEmpty},
		{State: thermal.SourceMissing},
		{State: thermal.SourceUnreadable},
		{State: thermal.SourceWrongShape},
	}
	summary := thermal.SummarizeSources(items)
	if summary.Total != 6 || summary.Found != 2 || summary.Empty != 1 || summary.Missing != 1 || summary.Unreadable != 1 || summary.WrongShape != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestDisplayHomePath(t *testing.T) {
	home := "/home/tester"
	cases := map[string]string{
		home:                       "~",
		home + "/.codex":           "~/.codex",
		home + "/.local/share/zed": "~/.local/share/zed",
		"/opt/thermal/opencode.db": "/opt/thermal/opencode.db",
	}
	for path, want := range cases {
		if got := displayHomePath(path, home); got != want {
			t.Errorf("displayHomePath(%q)=%q, want %q", path, got, want)
		}
	}
}
