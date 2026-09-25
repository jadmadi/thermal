// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

func RenderDashboard(toolName string, summary thermal.Summary, daily []thermal.DailyRow, dbPath string, weeks int, noColor bool, estimatedCost ...float64) string {
	st := NewStyle(noColor)
	colors := st.Colors
	muted := st.Muted
	highlight := st.Highlight
	dim := st.Dim

	tokenLabel := "tokens"
	switch toolName {
	case "command-code":
		tokenLabel = "messages"
	case "Agy":
		tokenLabel = "steps"
	case "Muse":
		tokenLabel = "prompts"
	case "Droid":
		tokenLabel = "messages"
	}

	activeDays := make(map[string]bool)
	for _, d := range daily {
		if d.Turns > 0 {
			activeDays[d.Day] = true
		}
	}
	current, longest := thermal.ComputeStreaks(activeDays)

	activity := make(map[string]thermal.DayActivity)
	for _, d := range daily {
		activity[d.Day] = thermal.DayActivity{Tokens: d.Tokens, Turns: d.Turns}
	}

	today := thermal.StartOfToday()
	thisSunday := today.AddDate(0, 0, -int(today.Weekday()))
	firstSunday := thisSunday.AddDate(0, 0, -(weeks-1)*7)
	firstDay := thermal.LocalDay(firstSunday)
	todayStr := thermal.LocalDay(today)

	var visibleTokens int64
	visibleActive := 0
	for _, d := range daily {
		if d.Day >= firstDay && d.Day <= todayStr {
			visibleTokens += d.Tokens
			if d.Turns > 0 {
				visibleActive++
			}
		}
	}

	var cardLines []string

	allTime := summary.LifetimeTokens

	// Executive KPI strip
	kpiLine := fmt.Sprintf(" %s %s / %d weeks   ·   %d active days  |  %d day streak  |  %d best  |  %s all-time",
		highlight(thermal.CompactNumber(visibleTokens)),
		tokenLabel,
		weeks,
		visibleActive,
		current,
		longest,
		thermal.CompactNumber(allTime),
	)
	cardLines = append(cardLines, kpiLine)
	cardLines = append(cardLines, "")

	for _, line := range RenderHeatmap(activity, weeks, colors) {
		cardLines = append(cardLines, line)
	}

	// Extra analytics line: cost, code changes, sessions, agent breakdown.
	var extra []string
	var est float64
	if len(estimatedCost) > 0 {
		est = estimatedCost[0]
	}
	if summary.Cost > 0 {
		if summary.Cost < 0.01 {
			extra = append(extra, fmt.Sprintf("$%.4f spent", summary.Cost))
		} else {
			extra = append(extra, fmt.Sprintf("$%.2f spent", summary.Cost))
		}
	} else if est > 0 {
		if est < 0.01 {
			extra = append(extra, fmt.Sprintf("~$%.4f spent (est)", est))
		} else {
			extra = append(extra, fmt.Sprintf("~$%.2f spent (est)", est))
		}
	}
	if summary.LinesAdded > 0 || summary.LinesDeleted > 0 {
		extra = append(extra, fmt.Sprintf("%s+ / %s- lines", thermal.CompactNumber(summary.LinesAdded), thermal.CompactNumber(summary.LinesDeleted)))
	}
	if summary.FilesTouched > 0 {
		extra = append(extra, fmt.Sprintf("%s files", thermal.CompactNumber(summary.FilesTouched)))
	}
	if summary.LifetimeTokens > 0 && summary.CacheTokens > 0 {
		cachePct := float64(summary.CacheTokens) / float64(summary.LifetimeTokens) * 100
		extra = append(extra, fmt.Sprintf("%.0f%% cache", cachePct))
	}
	if summary.Sessions > 0 {
		extra = append(extra, fmt.Sprintf("%d sessions", summary.Sessions))
	}

	if len(extra) > 0 || len(summary.AgentBreakdown) > 0 {
		maxW := 0
		for _, l := range cardLines {
			if w := ansi.StringWidth(l); w > maxW {
				maxW = w
			}
		}
		if maxW < 60 {
			maxW = 60
		}
		cardLines = append(cardLines, " "+dim(strings.Repeat("─", maxW)))
		if len(extra) > 0 {
			cardLines = append(cardLines, " "+strings.Join(extra, muted("  ·  ")))
		}
		if len(summary.AgentBreakdown) > 0 {
			type agentCount struct {
				agent string
				n     int
			}
			var agents []agentCount
			for a, n := range summary.AgentBreakdown {
				agents = append(agents, agentCount{a, n})
			}
			sort.Slice(agents, func(i, j int) bool { return agents[i].n > agents[j].n })
			limit := 3
			if len(agents) < limit {
				limit = len(agents)
			}
			var parts []string
			for i := 0; i < limit; i++ {
				parts = append(parts, fmt.Sprintf("%s: %d", agents[i].agent, agents[i].n))
			}
			cardLines = append(cardLines, fmt.Sprintf(" %s agents  %s", muted("·"), strings.Join(parts, muted("  "))))
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(RenderCard(CardOptions{
		Title:       toolName + " Activity",
		RightHeader: thermal.FormatPath(dbPath),
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))
	sb.WriteString("\n")

	return sb.String()
}
