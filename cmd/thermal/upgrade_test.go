// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestFindAsset_ExactMatching(t *testing.T) {
	assets := []asset{
		{Name: "thermal_0.3.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64"},
		{Name: "thermal_0.3.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_arm64"},
	}

	name, _, err := findAsset(assets)
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		if err != nil || name != "thermal_0.3.0_linux_amd64.tar.gz" {
			t.Errorf("failed matching linux_amd64: %v", err)
		}
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		if err != nil || name != "thermal_0.3.0_darwin_arm64.tar.gz" {
			t.Errorf("failed matching darwin_arm64: %v", err)
		}
	} else {
		// If on windows or other arch, must return clean error without picking wrong architecture
		if err == nil {
			t.Errorf("expected error when no matching architecture exists, got %q", name)
		}
	}
}

func TestIsBreakingUpgrade(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"0.13.0", "v0.14.0", true},
		{"0.13.0", "0.14.0", true},
		{"0.13.0", "v0.13.1", false},
		{"1.0.0", "v2.0.0", true},
		{"1.0.0", "v1.1.0", false},
		{"1.0.0", "v1.0.1", false},
		{"dev", "v0.14.0", true},
		{"", "v0.14.0", true},
		{"0.13.0", "0.13.0", false},
		{"v0.13.0", "v0.13.0", false},
	}

	for _, tt := range tests {
		got := isBreakingUpgrade(tt.current, tt.latest)
		if got != tt.expected {
			t.Errorf("isBreakingUpgrade(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.expected)
		}
	}
}

func TestPrintUpgradeSuccess(t *testing.T) {
	// 1. Breaking upgrade should include migration guide notice
	var buf strings.Builder
	printUpgradeSuccess(&buf, "0.13.0", "v0.14.0")
	out := buf.String()
	if !strings.Contains(out, "✓ upgraded to v0.14.0") {
		t.Errorf("expected upgrade confirmation, got %q", out)
	}
	if !strings.Contains(out, "migration guide: https://thermal.jadmadi.net/migration or docs/MIGRATION.md") {
		t.Errorf("expected migration guide notice in breaking upgrade, got %q", out)
	}

	// 2. Patch upgrade should NOT include migration guide notice
	buf.Reset()
	printUpgradeSuccess(&buf, "0.13.0", "v0.13.1")
	out = buf.String()
	if !strings.Contains(out, "✓ upgraded to v0.13.1") {
		t.Errorf("expected upgrade confirmation, got %q", out)
	}
	if strings.Contains(out, "migration guide") {
		t.Errorf("expected NO migration guide notice in patch upgrade, got %q", out)
	}
}
