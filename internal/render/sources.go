// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

const sourceTableWidth = 71

// RenderSources prints the stat-only source inventory. It never includes file
// contents, session titles, project names, or prompts.
func RenderSources(items []thermal.SourceInfo, noColor bool) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  Thermal · sources · %d supported tools\n\n", len(items)))

	sb.WriteString(fmt.Sprintf("  %-14s %-5s %-22s %-12s %-14s\n",
		"TOOL", "ALIAS", "PATH", "STATE", "DETAIL"))
	sb.WriteString("  " + strings.Repeat("-", sourceTableWidth) + "\n")
	for _, item := range items {
		alias := item.Alias
		if alias == "" {
			alias = "—"
		}
		path := truncateLeft(item.DisplayPath, 22)
		detail := truncateSourceRight(item.Detail, 14)
		sb.WriteString(fmt.Sprintf("  %-14s %-5s %-22s %-12s %-14s\n",
			truncateSourceRight(item.Name, 14),
			alias,
			path,
			string(item.State),
			detail,
		))
	}
	sb.WriteString("\n")

	summary := thermal.SummarizeSources(items)
	sb.WriteString(fmt.Sprintf("  Summary: %d found · %d empty · %d missing · %d unreadable",
		summary.Found, summary.Empty, summary.Missing, summary.Unreadable))
	if summary.WrongShape > 0 {
		sb.WriteString(fmt.Sprintf(" · %d wrong-shape", summary.WrongShape))
	}
	sb.WriteString("\n")
	sb.WriteString("  Legend: found=data · empty=no sessions · missing=path absent\n")
	sb.WriteString("          unreadable=permissions · wrong-shape=unexpected file/directory\n\n")
	sb.WriteString("  Next: run one prompt in any supported tool, then re-run thermal.\n")
	sb.WriteString("        thermal sources --json   machine-readable inventory\n")
	sb.WriteString("        thermal --help           supported tools and flags\n")
	return sb.String()
}

// RenderSourcesJSON returns the stable machine-readable inventory. The caller
// supplies the home and clock so tests are deterministic.
func RenderSourcesJSON(items []thermal.SourceInfo, home string, now time.Time) string {
	if items == nil {
		items = []thermal.SourceInfo{}
	}
	payload := struct {
		Type        string                `json:"type"`
		GeneratedAt string                `json:"generatedAt"`
		Home        string                `json:"home"`
		Summary     thermal.SourceSummary `json:"summary"`
		Sources     []thermal.SourceInfo  `json:"sources"`
	}{
		Type:        "sources",
		GeneratedAt: now.UTC().Format(time.RFC3339),
		Home:        home,
		Summary:     thermal.SummarizeSources(items),
		Sources:     items,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(b) + "\n"
}

func truncateLeft(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	return "…" + string(r[len(r)-(width-1):])
}

func truncateSourceRight(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	return string(r[:width-1]) + "…"
}
