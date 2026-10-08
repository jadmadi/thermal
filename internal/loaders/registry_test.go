// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestResolveTool_Aliases(t *testing.T) {
	tests := []struct {
		alias string
		want  thermal.Tool
		ok    bool
	}{
		{"mimo", thermal.ToolMiMoCode, true},
		{"oc", thermal.ToolOpenCode, true},
		{"cmd", thermal.ToolCommandCode, true},
		{"whale", thermal.ToolCodewhale, true},
		{"zcode", thermal.ToolZCode, true},
		{"zc", thermal.ToolZCode, true},
		{"grok", thermal.ToolGrok, true},
		{"muse", thermal.ToolMuse, true},
		{"claude", thermal.ToolClaude, true},
		{"ccode", thermal.ToolClaude, true},
		{"droid", thermal.ToolDroid, true},
		{"factory", thermal.ToolDroid, true},
		{"dsh", thermal.ToolDsh, true},
		{"deepseek", thermal.ToolDsh, true},
		{"deepseek-harness", thermal.ToolDsh, true},
		{"hermes", thermal.ToolHermes, true},
		{"nous", thermal.ToolHermes, true},
		{"nous-hermes", thermal.ToolHermes, true},
		{"zed", thermal.ToolZed, true},
		{"zed-editor", thermal.ToolZed, true},
		{"agy", thermal.ToolAgy, true},
		{"antigravity", thermal.ToolAgy, true},
		{"nonexistent", "", false},
	}
	for _, tc := range tests {
		got, ok := ResolveTool(tc.alias)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ResolveTool(%q) = (%q, %v), want (%q, %v)", tc.alias, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAgyHomeDir(t *testing.T) {
	fakeHome := t.TempDir()
	unifiedDir := filepath.Join(fakeHome, ".gemini", "antigravity")
	cliDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(filepath.Join(unifiedDir, "brain"), 0755); err != nil {
		t.Fatalf("mkdir unified brain failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(cliDir, "brain"), 0755); err != nil {
		t.Fatalf("mkdir cli brain failed: %v", err)
	}

	// 1. Relative ANTIGRAVITY_APP_DATA_DIR="antigravity" (official Antigravity CLI/Desktop 2.0 convention)
	t.Setenv("ANTIGRAVITY_APP_DATA_DIR", "antigravity")
	t.Setenv("AGY_HOME", "")
	got := agyHomeDir(fakeHome)
	if got != unifiedDir {
		t.Errorf("agyHomeDir with relative 'antigravity' = %q, want %q", got, unifiedDir)
	}

	// 2. Relative ANTIGRAVITY_APP_DATA_DIR="antigravity-cli"
	t.Setenv("ANTIGRAVITY_APP_DATA_DIR", "antigravity-cli")
	got = agyHomeDir(fakeHome)
	if got != cliDir {
		t.Errorf("agyHomeDir with relative 'antigravity-cli' = %q, want %q", got, cliDir)
	}

	// 3. Tilde path ANTIGRAVITY_APP_DATA_DIR="~/.gemini/antigravity"
	t.Setenv("ANTIGRAVITY_APP_DATA_DIR", "~/.gemini/antigravity")
	got = agyHomeDir(fakeHome)
	if got != unifiedDir {
		t.Errorf("agyHomeDir with tilde path = %q, want %q", got, unifiedDir)
	}

	// 4. Absolute path
	t.Setenv("ANTIGRAVITY_APP_DATA_DIR", unifiedDir)
	got = agyHomeDir(fakeHome)
	if got != unifiedDir {
		t.Errorf("agyHomeDir with absolute path = %q, want %q", got, unifiedDir)
	}

	// 5. Unset -> defaults to auto-discovery preferring unified hub
	t.Setenv("ANTIGRAVITY_APP_DATA_DIR", "")
	got = agyHomeDir(fakeHome)
	if got != unifiedDir {
		t.Errorf("agyHomeDir with unset env = %q, want %q", got, unifiedDir)
	}

	// 6. When unified hub does not exist, auto-discovers cliDir
	fakeHome2 := t.TempDir()
	cliOnlyDir := filepath.Join(fakeHome2, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(filepath.Join(cliOnlyDir, "brain"), 0755); err != nil {
		t.Fatalf("mkdir cliOnly brain failed: %v", err)
	}
	got = agyHomeDir(fakeHome2)
	if got != cliOnlyDir {
		t.Errorf("agyHomeDir fallback to cliDir = %q, want %q", got, cliOnlyDir)
	}
}

func TestAllTools_Config(t *testing.T) {
	tools := AllTools()
	if len(tools) == 0 {
		t.Fatalf("expected tools in registry")
	}
	for id, info := range tools {
		if info.Name == "" || info.Loader == nil {
			t.Errorf("tool %s has incomplete ToolInfo configuration", id)
		}
	}
}

func TestProbePlatformDataPath(t *testing.T) {
	fakeHome := t.TempDir()

	// 1. None exists -> returns primary Unix fallback
	db, dir := probePlatformDataPath(fakeHome, "opencode", "opencode.db")
	expectedDB := filepath.Join(fakeHome, ".local", "share", "opencode", "opencode.db")
	expectedDir := filepath.Join(fakeHome, ".local", "share", "opencode")
	if db != expectedDB || dir != expectedDir {
		t.Fatalf("expected fallback (%q, %q), got (%q, %q)", expectedDB, expectedDir, db, dir)
	}

	// 2. macOS Library/Application Support exists
	macOSDir := filepath.Join(fakeHome, "Library", "Application Support", "opencode")
	if err := os.MkdirAll(macOSDir, 0755); err != nil {
		t.Fatalf("mkdir macOS failed: %v", err)
	}
	macOSDB := filepath.Join(macOSDir, "opencode.db")
	if err := os.WriteFile(macOSDB, []byte("mock"), 0644); err != nil {
		t.Fatalf("write macOS db failed: %v", err)
	}
	db, dir = probePlatformDataPath(fakeHome, "opencode", "opencode.db")
	if db != macOSDB || dir != macOSDir {
		t.Fatalf("expected macOS discovery (%q, %q), got (%q, %q)", macOSDB, macOSDir, db, dir)
	}

	// 3. Windows APPDATA takes precedence when Unix and macOS do not exist
	fakeHomeWin := t.TempDir()
	appDataDir := filepath.Join(t.TempDir(), "AppDataRoaming")
	t.Setenv("APPDATA", appDataDir)
	winAppDir := filepath.Join(appDataDir, "opencode")
	if err := os.MkdirAll(winAppDir, 0755); err != nil {
		t.Fatalf("mkdir winAppDir failed: %v", err)
	}
	winDB := filepath.Join(winAppDir, "opencode.db")
	if err := os.WriteFile(winDB, []byte("mock"), 0644); err != nil {
		t.Fatalf("write winDB failed: %v", err)
	}
	db, dir = probePlatformDataPath(fakeHomeWin, "opencode", "opencode.db")
	if db != winDB || dir != winAppDir {
		t.Fatalf("expected APPDATA discovery (%q, %q), got (%q, %q)", winDB, winAppDir, db, dir)
	}
}
