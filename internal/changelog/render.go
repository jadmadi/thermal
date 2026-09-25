// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package changelog

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/theme"
)

// RenderChangelog formats the changelog report for terminal display.
func RenderChangelog(rep ChangelogReport, noColor bool) string {
	st := render.NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	faint := st.Faint
	green := st.Success
	coral := st.Error

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("changelog · release history & new features"),
	))

	curVer := rep.CurrentVersion
	if curVer == "" || curVer == "dev" {
		curVer = "dev (source build)"
	} else if !strings.HasPrefix(curVer, "v") {
		curVer = "v" + curVer
	}

	shown := len(rep.Releases)
	sb.WriteString(fmt.Sprintf("  Current: %s  ·  Showing %d of %d releases\n\n",
		highlight(curVer),
		shown,
		rep.TotalReleases,
	))

	if len(rep.Releases) == 0 {
		sb.WriteString("  " + faint("No changelog entries found.") + "\n\n")
		return sb.String()
	}

	cardWidth := render.BoundedCardWidth(2)
	innerWidth := cardWidth - 4

	formatBullet := func(s string) string {
		if !colors || !strings.Contains(s, "`") {
			return s
		}
		var out strings.Builder
		remaining := s
		for {
			start := strings.Index(remaining, "`")
			if start == -1 {
				out.WriteString(remaining)
				break
			}
			end := strings.Index(remaining[start+1:], "`")
			if end == -1 {
				out.WriteString(remaining)
				break
			}
			out.WriteString(remaining[:start])
			token := remaining[start+1 : start+1+end]
			out.WriteString(theme.Primary.SprintBold(colors, token))
			remaining = remaining[start+1+end+1:]
		}
		return out.String()
	}

	for i, rel := range rep.Releases {
		var lines []string

		// 1. Breaking Changes
		if len(rel.Breaking) > 0 {
			lines = append(lines, " "+coral("Breaking Changes:"))
			for _, b := range rel.Breaking {
				lines = append(lines, render.WrapBullet("  • ", formatBullet(b), innerWidth, coral)...)
			}
			lines = append(lines, "")
		}

		// 2. Features
		if len(rel.Features) > 0 {
			lines = append(lines, " "+green("Features:"))
			for _, feat := range rel.Features {
				lines = append(lines, render.WrapBullet("  • ", formatBullet(feat), innerWidth)...)
			}
		}

		// 3. Bug Fixes (if present and room allows)
		if len(rel.Fixes) > 0 {
			if len(rel.Features) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, " "+faint("Bug Fixes:"))
			for _, fix := range rel.Fixes {
				lines = append(lines, render.WrapBullet("  • ", formatBullet(fix), innerWidth, faint)...)
			}
		}

		// 4. Other
		if len(rel.Other) > 0 && len(rel.Features) == 0 && len(rel.Fixes) == 0 {
			lines = append(lines, " "+dim("Changes:"))
			for _, o := range rel.Other {
				lines = append(lines, render.WrapBullet("  • ", formatBullet(o), innerWidth)...)
			}
		}

		// Clean up trailing empty line if any
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		title := rel.Version
		if title != "Unreleased" && !strings.HasPrefix(title, "v") {
			title = "v" + title
		}

		titleColor := theme.Secondary
		if i == 0 {
			titleColor = theme.Primary
		}

		sb.WriteString(render.RenderCard(render.CardOptions{
			Title:       title,
			RightHeader: rel.Date,
			Lines:       lines,
			Width:       cardWidth,
			Indent:      2,
			Colors:      colors,
			TitleColor:  titleColor,
			BorderColor: theme.Border,
		}))
		sb.WriteString("\n")
	}

	if shown < rep.TotalReleases {
		remainder := rep.TotalReleases - shown
		sb.WriteString(fmt.Sprintf("  %s\n\n",
			dim(fmt.Sprintf("… and %d more releases (use --top %d to view all, or --json for full history)", remainder, rep.TotalReleases)),
		))
	}

	return sb.String()
}

// RenderChangelogJSON formats the changelog report as indented JSON.
func RenderChangelogJSON(rep ChangelogReport) string {
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
