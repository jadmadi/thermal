// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package loaders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestLoadUserConfig_INI(t *testing.T) {
	fakeHome := t.TempDir()
	cfgDir := filepath.Join(fakeHome, ".config", "thermal")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `
# Thermal Tool Paths Override
[paths]
agy = ~/my-custom-antigravity
opencode = /custom/path/opencode.db
codex = ~/my-codex
hermes: /custom/hermes/state.db
`
	if err := os.WriteFile(filepath.Join(cfgDir, "thermal.config"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	overrides := LoadUserConfig(fakeHome)
	if len(overrides) != 4 {
		t.Fatalf("expected 4 overrides, got %d", len(overrides))
	}

	wantAgy := filepath.Join(fakeHome, "my-custom-antigravity")
	if overrides[thermal.ToolAgy] != wantAgy {
		t.Errorf("ToolAgy = %q, want %q", overrides[thermal.ToolAgy], wantAgy)
	}

	if overrides[thermal.ToolOpenCode] != filepath.Clean("/custom/path/opencode.db") {
		t.Errorf("ToolOpenCode = %q, want %q", overrides[thermal.ToolOpenCode], filepath.Clean("/custom/path/opencode.db"))
	}

	wantCodex := filepath.Join(fakeHome, "my-codex")
	if overrides[thermal.ToolCodex] != wantCodex {
		t.Errorf("ToolCodex = %q, want %q", overrides[thermal.ToolCodex], wantCodex)
	}

	if overrides[thermal.ToolHermes] != filepath.Clean("/custom/hermes/state.db") {
		t.Errorf("ToolHermes = %q, want %q", overrides[thermal.ToolHermes], filepath.Clean("/custom/hermes/state.db"))
	}
}

func TestLoadUserConfig_StructuredJSON(t *testing.T) {
	fakeHome := t.TempDir()
	cfgDir := filepath.Join(fakeHome, ".config", "thermal")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `{
  "paths": {
    "antigravity": "~/.gemini/custom-agy",
    "oc": "/opt/opencode/opencode.db",
    "cmd": "~/my-command-code"
  }
}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	overrides := LoadUserConfig(fakeHome)
	if len(overrides) != 3 {
		t.Fatalf("expected 3 overrides, got %d", len(overrides))
	}

	wantAgy := filepath.Join(fakeHome, ".gemini", "custom-agy")
	if overrides[thermal.ToolAgy] != wantAgy {
		t.Errorf("ToolAgy = %q, want %q", overrides[thermal.ToolAgy], wantAgy)
	}
	if overrides[thermal.ToolOpenCode] != filepath.Clean("/opt/opencode/opencode.db") {
		t.Errorf("ToolOpenCode = %q, want %q", overrides[thermal.ToolOpenCode], filepath.Clean("/opt/opencode/opencode.db"))
	}
	wantCmd := filepath.Join(fakeHome, "my-command-code")
	if overrides[thermal.ToolCommandCode] != wantCmd {
		t.Errorf("ToolCommandCode = %q, want %q", overrides[thermal.ToolCommandCode], wantCmd)
	}
}

func TestLoadUserConfig_FlatJSON(t *testing.T) {
	fakeHome := t.TempDir()
	cfgDir := filepath.Join(fakeHome, ".config", "thermal")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `{
  "grok": "~/custom-grok",
  "claude": "/custom/claude"
}`
	if err := os.WriteFile(filepath.Join(cfgDir, "thermal.config"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	overrides := LoadUserConfig(fakeHome)
	if len(overrides) != 2 {
		t.Fatalf("expected 2 overrides, got %d", len(overrides))
	}

	wantGrok := filepath.Join(fakeHome, "custom-grok")
	if overrides[thermal.ToolGrok] != wantGrok {
		t.Errorf("ToolGrok = %q, want %q", overrides[thermal.ToolGrok], wantGrok)
	}
	if overrides[thermal.ToolClaude] != filepath.Clean("/custom/claude") {
		t.Errorf("ToolClaude = %q, want %q", overrides[thermal.ToolClaude], filepath.Clean("/custom/claude"))
	}
}

func TestLoadUserConfig_EnvOverride(t *testing.T) {
	fakeHome := t.TempDir()
	customFile := filepath.Join(fakeHome, "custom-thermal.ini")
	content := "mimocode = /custom/mimo.db\n"
	if err := os.WriteFile(customFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("THERMAL_CONFIG", customFile)
	overrides := LoadUserConfig(fakeHome)
	if overrides[thermal.ToolMiMoCode] != filepath.Clean("/custom/mimo.db") {
		t.Errorf("THERMAL_CONFIG override failed: got %v", overrides)
	}
}

func TestLoadUserConfig_EmptyWhenMissing(t *testing.T) {
	fakeHome := t.TempDir()
	overrides := LoadUserConfig(fakeHome)
	if len(overrides) != 0 {
		t.Errorf("expected empty overrides for missing config, got %v", overrides)
	}
}

func TestAllTools_ConfigOverrides(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome)

	cfgDir := filepath.Join(fakeHome, ".config", "thermal")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	customAgy := filepath.Join(fakeHome, "special-antigravity")
	if err := os.MkdirAll(filepath.Join(customAgy, "brain"), 0755); err != nil {
		t.Fatal(err)
	}
	customOC := filepath.Join(fakeHome, "special-opencode", "opencode.db")
	if err := os.MkdirAll(filepath.Dir(customOC), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(customOC, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	content := "agy = ~/special-antigravity\nopencode = ~/special-opencode/opencode.db\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "thermal.config"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tools := AllTools()
	agyInfo := tools[thermal.ToolAgy]
	if agyInfo.DataDir != customAgy {
		t.Errorf("agy DataDir = %q, want %q", agyInfo.DataDir, customAgy)
	}
	ocInfo := tools[thermal.ToolOpenCode]
	if ocInfo.DBPath != customOC {
		t.Errorf("opencode DBPath = %q, want %q", ocInfo.DBPath, customOC)
	}
}
