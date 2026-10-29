// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package changelog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleChangelog = `# Changelog

All notable changes to this project will be documented in this file.

## [0.14.0](https://github.com/jadmadi/thermal/compare/v0.13.0...v0.14.0) (2026-09-22)

### Features

* **cli:** add daily cached background update check and upgrade notification ([2c387fb](https://github.com/...))
* **discovery:** declare Content-Signal directives in robots.txt ([fd5ca83](https://github.com/...))

### Bug Fixes

* **docs:** sanitize project names and paths across documentation ([345afbe](https://github.com/...))

## [0.13.0](https://github.com/jadmadi/thermal/compare/v0.12.0...v0.13.0) (2026-09-22)

### Features

* **init:** initial release of Thermal under GNU AGPL-3.0 ([23a7733](https://github.com/...))
`

func TestParse(t *testing.T) {
	releases, err := Parse(sampleChangelog)
	if err != nil {
		t.Fatalf("unexpected error parsing changelog: %v", err)
	}

	if len(releases) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(releases))
	}

	r1 := releases[0]
	if r1.Version != "0.14.0" {
		t.Errorf("expected version 0.14.0, got %s", r1.Version)
	}
	if r1.Date != "2026-09-22" {
		t.Errorf("expected date 2026-09-22, got %s", r1.Date)
	}
	if len(r1.Features) != 2 {
		t.Fatalf("expected 2 features in r1, got %d", len(r1.Features))
	}
	if !strings.HasPrefix(r1.Features[0], "cli: add daily cached") {
		t.Errorf("expected clean scope prefix, got %s", r1.Features[0])
	}
	if strings.Contains(r1.Features[0], "2c387fb") {
		t.Errorf("expected commit hash stripped, got %s", r1.Features[0])
	}

	if len(r1.Fixes) != 1 {
		t.Fatalf("expected 1 fix in r1, got %d", len(r1.Fixes))
	}

	r2 := releases[1]
	if r2.Version != "0.13.0" {
		t.Errorf("expected version 0.13.0, got %s", r2.Version)
	}
	if len(r2.Features) != 1 {
		t.Fatalf("expected 1 feature in r2, got %d", len(r2.Features))
	}
}

func TestRenderChangelog(t *testing.T) {
	releases, err := Parse(sampleChangelog)
	if err != nil {
		t.Fatal(err)
	}

	rep := ChangelogReport{
		CurrentVersion: "0.14.0",
		TotalReleases:  len(releases),
		Releases:       releases,
	}

	out := RenderChangelog(rep, true)
	if strings.Contains(out, "\033[") {
		t.Errorf("expected zero ANSI escape sequences with noColor=true")
	}
	if !strings.Contains(out, "Thermal · changelog") {
		t.Errorf("expected title banner in output")
	}
	if !strings.Contains(out, "v0.14.0") {
		t.Errorf("expected v0.14.0 in output")
	}
	if !strings.Contains(out, "Features:") {
		t.Errorf("expected Features: in output")
	}
	if !strings.Contains(out, "cli: add daily cached") {
		t.Errorf("expected feature bullet in output")
	}

	// Test JSON
	jsonStr := RenderChangelogJSON(rep)
	var decoded ChangelogReport
	if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
		t.Fatalf("failed to unmarshal changelog JSON: %v", err)
	}
	if len(decoded.Releases) != 2 {
		t.Errorf("expected 2 releases in decoded JSON, got %d", len(decoded.Releases))
	}
}

func TestGetReport(t *testing.T) {
	rep := GetReport(3)
	if rep.TotalReleases == 0 {
		t.Errorf("expected total releases > 0 from embedded/local changelog")
	}
	if len(rep.Releases) > 3 {
		t.Errorf("expected at most 3 releases with limit=3, got %d", len(rep.Releases))
	}
}

func TestParseStripsGoalTokensAndDeduplicatesReleases(t *testing.T) {
	raw := `# Changelog

## [0.21.0](https://github.com/jadmadi/thermal/compare/v0.20.1...v0.21.0) (2026-10-29)

### Features

* **pricing:** add cost provenance reconciliation (goals/cost-provenance-display-and-reconciliation/goal.md) [goal: cost-provenance-display-and-reconciliation] ([a7034d1](https://github.com/jadmadi/thermal/commit/a7034d1))
* **live:** separate baseline carry-in (goals/live-baseline-separation/goal.md) (goals/live-ticker-unit-and-cost-clarity/goal.md)

## [0.21.0](https://github.com/jadmadi/thermal/compare/v0.20.1...v0.21.0) (2026-10-29)

### Features

* **duplicate:** should be skipped because 0.21.0 was already parsed
`
	releases, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 deduplicated release, got %d", len(releases))
	}
	for _, feat := range releases[0].Features {
		if strings.Contains(feat, "goals/") || strings.Contains(feat, "[goal:") {
			t.Errorf("expected goal tokens stripped from bullet, got %q", feat)
		}
		if strings.Contains(feat, "duplicate:") {
			t.Errorf("expected duplicate release block to be skipped, got %q", feat)
		}
	}
}

func TestLoadContentIgnoresWorkingDirectoryChangelog(t *testing.T) {
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	tmpDir := t.TempDir()
	hijack := "# Changelog\n\n## [99.99.99] (2099-01-01)\n\n### Features\n\n* **hijack:** foreign repo changelog\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "CHANGELOG.md"), []byte(hijack), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	content := LoadContent()
	if strings.Contains(content, "99.99.99") || strings.Contains(content, "foreign repo changelog") {
		t.Errorf("LoadContent() read CHANGELOG.md from working directory instead of embedded changelog")
	}
}

func TestEmbeddedChangelogCleanliness(t *testing.T) {
	releases, err := Parse(LoadContent())
	if err != nil {
		t.Fatalf("unexpected error parsing embedded changelog: %v", err)
	}
	seen := make(map[string]bool)
	for _, rel := range releases {
		if seen[rel.Version] {
			t.Errorf("duplicate release version %q in embedded changelog", rel.Version)
		}
		seen[rel.Version] = true
		allBullets := append(append(append([]string{}, rel.Features...), rel.Fixes...), rel.Breaking...)
		allBullets = append(allBullets, rel.Other...)
		for _, b := range allBullets {
			if strings.Contains(b, "goals/") || strings.Contains(b, "[goal:") {
				t.Errorf("release %s bullet contains leaked goal token: %q", rel.Version, b)
			}
		}
	}
}
