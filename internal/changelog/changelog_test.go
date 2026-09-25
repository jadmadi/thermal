// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package changelog

import (
	"encoding/json"
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
