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
