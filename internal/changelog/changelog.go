// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package changelog

import (
	"bufio"
	_ "embed"
	"os"
	"regexp"
	"strings"

	"github.com/jadmadi/thermal/internal/version"
)

//go:embed CHANGELOG.md
var embeddedChangelog string

// Release represents a single release entry from the changelog.
type Release struct {
	Version  string   `json:"version"`
	Date     string   `json:"date,omitempty"`
	URL      string   `json:"url,omitempty"`
	Features []string `json:"features,omitempty"`
	Fixes    []string `json:"fixes,omitempty"`
	Breaking []string `json:"breaking,omitempty"`
	Other    []string `json:"other,omitempty"`
}

// ChangelogReport represents the structured report of all or selected releases.
type ChangelogReport struct {
	CurrentVersion string    `json:"currentVersion"`
	TotalReleases  int       `json:"totalReleases"`
	Releases       []Release `json:"releases"`
}

var (
	headerRegex = regexp.MustCompile(`^##\s+\[?([^\]\s]+)\]?(?:\(([^\)]+)\))?\s*(?:\(([^)]+)\))?`)
	commitRegex = regexp.MustCompile(`\s*\(\[[0-9a-fA-F]+\]\([^\)]+\)\)`)
)

// Parse parses a Keep a Changelog formatted markdown string into a slice of Release objects.
func Parse(content string) ([]Release, error) {
	var releases []Release
	var current *Release
	var currentSection string

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Detect Release header: e.g. ## [0.14.0](https://...) (2026-09-22) or ## [Unreleased]
		if strings.HasPrefix(line, "## ") {
			if current != nil && (len(current.Features) > 0 || len(current.Fixes) > 0 || len(current.Breaking) > 0 || len(current.Other) > 0) {
				releases = append(releases, *current)
			}

			matches := headerRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				v := matches[1]
				url := ""
				date := ""
				if len(matches) > 2 {
					url = matches[2]
				}
				if len(matches) > 3 {
					date = matches[3]
				}
				current = &Release{
					Version: v,
					URL:     url,
					Date:    date,
				}
				currentSection = ""
				continue
			}
		}

		if current == nil {
			continue
		}

		// Detect section headers: ### Features, ### Bug Fixes, etc.
		if strings.HasPrefix(line, "### ") {
			secTitle := strings.ToLower(strings.TrimPrefix(line, "### "))
			switch {
			case strings.Contains(secTitle, "feature"):
				currentSection = "features"
			case strings.Contains(secTitle, "fix"):
				currentSection = "fixes"
			case strings.Contains(secTitle, "breaking"):
				currentSection = "breaking"
			default:
				currentSection = "other"
			}
			continue
		}

		// Detect bullet points: * ... or - ...
		if strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "- ") {
			bullet := strings.TrimSpace(line[2:])
			// Strip markdown commit link at the end: ([2c387fb](https://...))
			bullet = commitRegex.ReplaceAllString(bullet, "")
			// Convert markdown bold prefix **scope:** to scope:
			if strings.HasPrefix(bullet, "**") {
				if idx := strings.Index(bullet[2:], "**"); idx != -1 {
					scope := bullet[2 : 2+idx]
					rest := strings.TrimSpace(bullet[2+idx+2:])
					bullet = scope + " " + rest
				}
			}

			switch currentSection {
			case "features":
				current.Features = append(current.Features, bullet)
			case "fixes":
				current.Fixes = append(current.Fixes, bullet)
			case "breaking":
				current.Breaking = append(current.Breaking, bullet)
			default:
				if bullet != "" {
					current.Other = append(current.Other, bullet)
				}
			}
		}
	}

	if current != nil && (len(current.Features) > 0 || len(current.Fixes) > 0 || len(current.Breaking) > 0 || len(current.Other) > 0) {
		releases = append(releases, *current)
	}

	return releases, scanner.Err()
}

// LoadContent loads the changelog markdown from local disk if available,
// falling back to embedded compile-time changelog.
func LoadContent() string {
	// Try local file first (helpful during local development / testing)
	candidates := []string{
		"CHANGELOG.md",
		"../CHANGELOG.md",
		"../../CHANGELOG.md",
	}
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
			return string(data)
		}
	}
	return embeddedChangelog
}

// GetReport parses the changelog and returns a ChangelogReport.
func GetReport(limit int) ChangelogReport {
	raw := LoadContent()
	all, _ := Parse(raw)

	total := len(all)
	selected := all
	if limit > 0 && limit < len(all) {
		selected = all[:limit]
	}

	return ChangelogReport{
		CurrentVersion: version.Version,
		TotalReleases:  total,
		Releases:       selected,
	}
}
