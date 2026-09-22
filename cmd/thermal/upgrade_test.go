// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
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

func TestSemverCompare(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"v0.14.0", "v0.13.0", 1},
		{"v0.13.0", "v0.14.0", -1},
		{"0.13.0", "0.13.0", 0},
		{"v1.2.3", "1.2.3", 0},
		{"v1.2.3", "v1.2.4", -1},
		{"v1.3.0", "v1.2.9", 1},
		{"v2.0.0", "v1.9.9", 1},
		{"dev", "v1.0.0", -1},
		{"v1.0.0", "dev", 1},
		{"", "v1.0.0", -1},
		{"v1.0.0", "", 1},
		{"dev", "dev", 0},
		{"v0.13.0", "v0.13.0-rc1", 0},
	}

	for _, tt := range tests {
		got := semverCompare(tt.v1, tt.v2)
		if got != tt.want {
			t.Errorf("semverCompare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestUpdateCachePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Initially cache does not exist
	_, err := loadUpdateCache()
	if err == nil {
		t.Errorf("expected error reading non-existent update cache, got nil")
	}

	// Save cache
	now := time.Now().UTC().Truncate(time.Second)
	testCache := updateCache{
		CheckedAt:     now,
		LatestVersion: "v0.14.0",
	}
	if err := saveUpdateCache(testCache); err != nil {
		t.Fatalf("saveUpdateCache failed: %v", err)
	}

	// Verify file was written to ~/.cache/thermal/update.json
	expectedPath := filepath.Join(tmpDir, ".cache", "thermal", "update.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected cache file at %s, but stat returned not exist", expectedPath)
	}

	// Load and verify
	loaded, err := loadUpdateCache()
	if err != nil {
		t.Fatalf("loadUpdateCache failed: %v", err)
	}
	if loaded.LatestVersion != "v0.14.0" {
		t.Errorf("expected LatestVersion 'v0.14.0', got %q", loaded.LatestVersion)
	}
	if !loaded.CheckedAt.Equal(now) {
		t.Errorf("expected CheckedAt %v, got %v", now, loaded.CheckedAt)
	}
}

func TestMaybeCheckForUpdate_Guards(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// All of these should return immediately without panicking or creating cache files
	cases := []struct {
		name string
		opts thermal.Options
		env  map[string]string
	}{
		{
			name: "json output disables check",
			opts: thermal.Options{JSON: true},
		},
		{
			name: "offline mode disables check",
			opts: thermal.Options{Offline: true},
		},
		{
			name: "no update check option disables check",
			opts: thermal.Options{NoUpdateCheck: true},
		},
		{
			name: "upgrade subcommand disables check",
			opts: thermal.Options{Tool: "upgrade"},
		},
		{
			name: "bg worker subcommand disables check",
			opts: thermal.Options{Tool: "--check-update-bg"},
		},
		{
			name: "THERMAL_NO_UPDATE_CHECK env disables check",
			opts: thermal.Options{},
			env:  map[string]string{"THERMAL_NO_UPDATE_CHECK": "1"},
		},
		{
			name: "CI env disables check",
			opts: thermal.Options{},
			env:  map[string]string{"CI": "true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			maybeCheckForUpdate(tc.opts)
		})
	}
}
