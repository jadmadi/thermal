// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

func medals(rank int, colors bool) string {
	prefix := fmt.Sprintf(" %d.", rank)
	if !colors {
		return prefix
	}
	switch rank {
	case 1:
		return theme.Secondary.SprintBold(colors, prefix)
	case 2:
		return theme.Text.Sprint(colors, prefix)
	case 3:
		return theme.Warning.Sprint(colors, prefix)
	default:
		return prefix
	}
}

func streakBar(streak int, maxStreak int, colors bool) string {
	if maxStreak == 0 || streak == 0 {
		return ""
	}
	width := 20
	filled := 0
	if maxStreak > 0 {
		filled = (streak * width) / maxStreak
	}
	if filled == 0 && streak > 0 {
		filled = 1
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			if colors {
				bar += theme.Primary.Sprint(colors, "█")
			} else {
				bar += "█"
			}
		} else {
			bar += " "
		}
	}
	return bar
}

func fireEmoji(streak int) string {
	switch {
	case streak >= 30:
		return "🔥🔥🔥"
	case streak >= 14:
		return "🔥🔥"
	case streak >= 7:
		return "🔥"
	default:
		return ""
	}
}

// RenderLeaderboard ranks tools. sortKey picks the primary ranking: "streak"
// (default) keeps the contribution-streak order, "tokens" ranks by token
// volume, and "cost" by recorded cost. Streak remains the tiebreaker.
// RenderLeaderboard prints the ranked tools. When estimate is false the frame
// adds a line stating the recorded cost total, because the Cost column only
// ever shows what a tool recorded and a reader should not have to add it up.
func RenderLeaderboard(results []thermal.ToolResult, weeks int, noColor bool, sortKey string, estimate bool, topN ...int) string {
	top := 0
	if len(topN) > 0 {
		top = topN[0]
	}
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	gold := st.Gold
	green := st.Primary // Streaks glow in Mimocode Blaze Amber
	dim := st.Dim

	tokenResults := make([]thermal.ToolResult, 0, len(results))
	activityResults := make([]thermal.ToolResult, 0, len(results))
	for _, r := range results {
		switch r.Tool {
		case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolCodex, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode, thermal.ToolGrok, thermal.ToolClaude, thermal.ToolDsh, thermal.ToolHermes, thermal.ToolZed:
			tokenResults = append(tokenResults, r)
		case thermal.ToolMuse:
			if r.Summary.LifetimeTokens > int64(r.Summary.Sessions) || r.Summary.InputTokens > 0 || r.Summary.OutputTokens > 0 {
				tokenResults = append(tokenResults, r)
			} else {
				activityResults = append(activityResults, r)
			}
		default:
			activityResults = append(activityResults, r)
		}
	}

	rankBy := func(list []thermal.ToolResult) {
		sort.Slice(list, func(i, j int) bool {
			a, b := list[i], list[j]
			switch strings.ToLower(sortKey) {
			case "tokens":
				if a.TotalActivity != b.TotalActivity {
					return a.TotalActivity > b.TotalActivity
				}
			case "cost":
				aCost := a.Summary.Cost
				if aCost == 0 && estimate {
					aCost = a.EstimatedCost
				}
				bCost := b.Summary.Cost
				if bCost == 0 && estimate {
					bCost = b.EstimatedCost
				}
				if aCost != bCost {
					return aCost > bCost
				}
			}
			if a.CurrentStreak != b.CurrentStreak {
				return a.CurrentStreak > b.CurrentStreak
			}
			if a.LongestStreak != b.LongestStreak {
				return a.LongestStreak > b.LongestStreak
			}
			return a.TotalActivity > b.TotalActivity
		})
	}
	rankBy(tokenResults)
	rankBy(activityResults)
	if top > 0 {
		if len(tokenResults) > top {
			tokenResults = tokenResults[:top]
		}
		if len(activityResults) > top {
			activityResults = activityResults[:top]
		}
	}

	maxTokenStreak := 0
	for _, r := range tokenResults {
		if r.CurrentStreak > maxTokenStreak {
			maxTokenStreak = r.CurrentStreak
		}
	}
	maxActStreak := 0
	for _, r := range activityResults {
		if r.CurrentStreak > maxActStreak {
			maxActStreak = r.CurrentStreak
		}
	}

	width := ContentWidth()
	compact := width < 70
	tiny := width < 56

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s  %s\n\n", highlight("THERMAL"), dim("— Don't break the streak.")))

	sortLabel := "current streak"
	switch strings.ToLower(sortKey) {
	case "tokens":
		sortLabel = "tokens"
	case "cost":
		sortLabel = "cost"
	}
	for _, line := range WrapText(fmt.Sprintf("Ranked by %s · all time", sortLabel), width-4) {
		sb.WriteString("  " + dim(line) + "\n")
	}
	sb.WriteString("\n")

	if len(tokenResults) > 0 {
		var tokenLines []string
		tokenHeaders := []string{"#", "Tool", "Strk", "Best", "Days", "Tokens", "Cost"}
		tokenAlign := []bool{true, false, true, true, true, true, true}
		tokenWidths := []int{3, 14, 6, 6, 6, 11, 11}
		tokenTableWidths := tokenWidths
		if tiny {
			tokenHeaders = []string{"#", "Tool", "Tokens"}
			tokenAlign = []bool{true, false, true}
			tokenTableWidths = []int{3, 14, 11}
		} else if compact {
			tokenHeaders = []string{"#", "Tool", "Strk", "Tokens", "Cost"}
			tokenAlign = []bool{true, false, true, true, true}
			tokenTableWidths = []int{3, 14, 6, 11, 11}
		}

		tokenLines = append(tokenLines, formatHeaderRow(tokenHeaders, tokenTableWidths, tokenAlign, colors))
		tokenLines = append(tokenLines, " "+tableRule(tokenTableWidths))

		for i, r := range tokenResults {
			rank := i + 1
			rankCell := formatRankCell(rank, tokenWidths[0], colors)

			nameStr := padRight(r.Name, tokenWidths[1])
			if colors {
				if rank == 1 {
					nameStr = padRightStyled(r.Name, tokenWidths[1], func(s string) string { return gold(s) })
				} else {
					nameStr = padRightStyled(r.Name, tokenWidths[1], func(s string) string { return theme.Text.Sprint(true, s) })
				}
			}

			streakPlain := fmt.Sprintf("%dd", r.CurrentStreak)
			var streakStr string
			if colors && r.CurrentStreak > 0 {
				streakStr = padLeftStyled(streakPlain, tokenWidths[2], func(s string) string { return green(s) })
			} else if colors {
				streakStr = padLeftStyled(streakPlain, tokenWidths[2], func(s string) string { return theme.TextMuted.Sprint(true, s) })
			} else {
				streakStr = padLeft(streakPlain, tokenWidths[2])
			}

			bestPlain := fmt.Sprintf("%dd", r.LongestStreak)
			var bestStr string
			if colors {
				bestStr = padLeftStyled(bestPlain, tokenWidths[3], func(s string) string { return theme.Text.Sprint(true, s) })
			} else {
				bestStr = padLeft(bestPlain, tokenWidths[3])
			}

			activePlain := fmt.Sprintf("%dd", r.ActiveDays)
			var activeStr string
			if colors {
				activeStr = padLeftStyled(activePlain, tokenWidths[4], func(s string) string { return theme.Text.Sprint(true, s) })
			} else {
				activeStr = padLeft(activePlain, tokenWidths[4])
			}

			tokPlain := fmt.Sprintf("%s tok", thermal.CompactNumber(r.TotalActivity))
			var tokStr string
			if colors {
				if r.TotalActivity >= 1_000_000_000 {
					tokStr = padLeftStyled(tokPlain, tokenWidths[5], func(s string) string { return theme.Primary.SprintBold(true, s) })
				} else if r.TotalActivity >= 100_000_000 {
					tokStr = padLeftStyled(tokPlain, tokenWidths[5], func(s string) string { return theme.Secondary.Sprint(true, s) })
				} else {
					tokStr = padLeftStyled(tokPlain, tokenWidths[5], func(s string) string { return theme.Text.Sprint(true, s) })
				}
			} else {
				tokStr = padLeft(tokPlain, tokenWidths[5])
			}

			costPlain := formatCostOrDash(r.Summary.Cost)
			if r.Summary.Cost > 0 && estimate && r.EstimatedCost > 0 {
				costPlain = formatEstimatedCost(r.Summary.Cost + r.EstimatedCost)
			} else if r.Summary.Cost <= 0 && estimate && r.EstimatedCost > 0 {
				costPlain = formatEstimatedCost(r.EstimatedCost)
			}
			var costStr string
			if colors {
				if costPlain == "—" {
					costStr = padLeftStyled(costPlain, tokenWidths[6], func(s string) string { return theme.TextMuted.Sprint(true, s) })
				} else if strings.HasPrefix(costPlain, "~") {
					costStr = padLeftStyled(costPlain, tokenWidths[6], func(s string) string { return theme.Secondary.Sprint(true, s) })
				} else {
					costStr = padLeftStyled(costPlain, tokenWidths[6], func(s string) string { return theme.Text.Sprint(true, s) })
				}
			} else {
				costStr = padLeft(costPlain, tokenWidths[6])
			}

			if tiny {
				tokenLines = append(tokenLines, fmt.Sprintf(" %s  %s  %s",
					rankCell, nameStr, tokStr,
				))
			} else if compact {
				tokenLines = append(tokenLines, fmt.Sprintf(" %s  %s  %s  %s  %s",
					rankCell, nameStr, streakStr, tokStr, costStr,
				))
			} else {
				tokenLines = append(tokenLines, fmt.Sprintf(" %s  %s  %s  %s  %s  %s  %s",
					rankCell, nameStr, streakStr, bestStr, activeStr, tokStr, costStr,
				))
			}
		}

		sb.WriteString(RenderCard(CardOptions{
			Title:       "Token Warriors",
			RightHeader: "tokens & spend",
			Lines:       tokenLines,
			Width:       CardWidthFor(width),
			Indent:      2,
			Colors:      colors,
			TitleColor:  theme.Primary,
			BorderColor: theme.Border,
		}))
		sb.WriteString("\n")
	}

	if len(activityResults) > 0 {
		var actLines []string
		actHeaders := []string{"#", "Tool", "Strk", "Best", "Days", "Activity"}
		actAlign := []bool{true, false, true, true, true, true}
		actWidths := []int{3, 14, 6, 6, 6, 12}
		actTableWidths := actWidths
		if tiny {
			actHeaders = []string{"#", "Tool", "Activity"}
			actAlign = []bool{true, false, true}
			actTableWidths = []int{3, 14, 12}
		} else if compact {
			actHeaders = []string{"#", "Tool", "Strk", "Activity"}
			actAlign = []bool{true, false, true, true}
			actTableWidths = []int{3, 14, 6, 12}
		}

		actLines = append(actLines, formatHeaderRow(actHeaders, actTableWidths, actAlign, colors))
		actLines = append(actLines, " "+tableRule(actTableWidths))

		for i, r := range activityResults {
			rank := i + 1
			rankCell := formatRankCell(rank, actWidths[0], colors)

			nameStr := padRight(r.Name, actWidths[1])
			if colors {
				if rank == 1 {
					nameStr = padRightStyled(r.Name, actWidths[1], func(s string) string { return gold(s) })
				} else {
					nameStr = padRightStyled(r.Name, actWidths[1], func(s string) string { return theme.Text.Sprint(true, s) })
				}
			}

			streakPlain := fmt.Sprintf("%dd", r.CurrentStreak)
			var streakStr string
			if colors && r.CurrentStreak > 0 {
				streakStr = padLeftStyled(streakPlain, actWidths[2], func(s string) string { return green(s) })
			} else if colors {
				streakStr = padLeftStyled(streakPlain, actWidths[2], func(s string) string { return theme.TextMuted.Sprint(true, s) })
			} else {
				streakStr = padLeft(streakPlain, actWidths[2])
			}

			bestPlain := fmt.Sprintf("%dd", r.LongestStreak)
			var bestStr string
			if colors {
				bestStr = padLeftStyled(bestPlain, actWidths[3], func(s string) string { return theme.Text.Sprint(true, s) })
			} else {
				bestStr = padLeft(bestPlain, actWidths[3])
			}

			activePlain := fmt.Sprintf("%dd", r.ActiveDays)
			var activeStr string
			if colors {
				activeStr = padLeftStyled(activePlain, actWidths[4], func(s string) string { return theme.Text.Sprint(true, s) })
			} else {
				activeStr = padLeft(activePlain, actWidths[4])
			}

			var actLabel string
			switch r.Tool {
			case thermal.ToolAgy:
				actLabel = "step"
			case thermal.ToolMuse:
				actLabel = "prompt"
			case thermal.ToolDroid:
				actLabel = "msg"
			default:
				actLabel = "act"
			}
			actPlain := fmt.Sprintf("%s %s", thermal.CompactNumber(r.TotalActivity), actLabel)
			var actStr string
			if colors {
				actStr = padLeftStyled(actPlain, actWidths[5], func(s string) string { return theme.Text.Sprint(true, s) })
			} else {
				actStr = padLeft(actPlain, actWidths[5])
			}

			if tiny {
				actLines = append(actLines, fmt.Sprintf(" %s  %s  %s",
					rankCell, nameStr, actStr,
				))
			} else if compact {
				actLines = append(actLines, fmt.Sprintf(" %s  %s  %s  %s",
					rankCell, nameStr, streakStr, actStr,
				))
			} else {
				actLines = append(actLines, fmt.Sprintf(" %s  %s  %s  %s  %s  %s",
					rankCell, nameStr, streakStr, bestStr, activeStr, actStr,
				))
			}
		}

		sb.WriteString(RenderCard(CardOptions{
			Title:       "Activity Hunters",
			RightHeader: "messages & steps",
			Lines:       actLines,
			Width:       CardWidthFor(width),
			Indent:      2,
			Colors:      colors,
			TitleColor:  theme.Secondary,
			BorderColor: theme.Border,
		}))
		sb.WriteString("\n")
		for _, line := range WrapText("Activity counts are steps and messages, not comparable to token totals.", width-4) {
			sb.WriteString("  " + dim(line) + "\n")
		}
		sb.WriteString("\n")
	}

	for _, line := range WrapText("Legend: Strk = current streak · Best = longest streak · Days = active days", width-4) {
		sb.WriteString("  " + dim(line) + "\n")
	}
	for _, line := range WrapText("        Cost: "+CostLegendText(), width-4) {
		sb.WriteString("  " + dim(line) + "\n")
	}
	sb.WriteString("\n")

	allResults := append(tokenResults, activityResults...)
	if len(allResults) > 0 {
		best := allResults[0]
		for _, r := range allResults {
			if r.CurrentStreak > best.CurrentStreak {
				best = r
			}
		}
		if best.CurrentStreak > 0 {
			line := fmt.Sprintf("%s %s is on fire with a %d-day streak!",
				gold(">>"), highlight(best.Name), best.CurrentStreak)
			for _, wl := range WrapText(line, width-4) {
				sb.WriteString("  " + wl + "\n")
			}
		} else {
			sb.WriteString(fmt.Sprintf("  %s No active streaks. Time to code!\n", dim("--")))
		}
	}

	// Cost provenance: one arithmetic line so the reader never has to add up
	// recorded and estimated values in their head.
	var recorded, estimated float64
	recording, estimating := 0, 0
	for _, r := range results {
		if r.Summary.Cost > 0 {
			recorded += r.Summary.Cost
			recording++
		}
		if estimate && r.EstimatedCost > 0 {
			estimated += r.EstimatedCost
			estimating++
		}
	}
	shown := recorded + estimated
	costLine := fmt.Sprintf("Cost: %s recorded + %s estimated = %s shown", formatCost(recorded), formatEstimatedCost(estimated), formatEstimatedCost(shown))
	if !estimate {
		costLine = fmt.Sprintf("Cost: %s recorded + estimation disabled = %s shown", formatCost(recorded), formatCost(recorded))
	}
	if recording > 0 || estimating > 0 {
		for _, line := range WrapText(costLine, width-4) {
			sb.WriteString("  " + dim(line) + "\n")
		}
	} else {
		for _, line := range WrapText("Cost: $0.00 recorded + $0.00 estimated = $0.00 shown", width-4) {
			sb.WriteString("  " + dim(line) + "\n")
		}
	}

	var totalUnattr int64
	unpriced := make(map[string]int64)
	if estimate {
		for _, r := range results {
			totalUnattr += r.UnattributedTokens
			for _, m := range r.MissingPricing {
				unpriced[m] += r.Summary.ModelBreakdown[m]
			}
		}
	}
	if note := unpriceableNote(totalUnattr); note != "" {
		for _, line := range WrapText(note, width-4) {
			sb.WriteString("  " + dim(line) + "\n")
		}
	}
	if len(unpriced) > 0 {
		names := make([]string, 0, len(unpriced))
		for name := range unpriced {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, 0, 4)
		for i, name := range names {
			if i == 4 {
				parts = append(parts, fmt.Sprintf("+%d more", len(names)-4))
				break
			}
			parts = append(parts, fmt.Sprintf("%s (%s tok)", name, thermal.CompactNumber(unpriced[name])))
		}
		note := "Unpriced models: " + strings.Join(parts, ", ")
		for _, line := range WrapText(note, width-4) {
			sb.WriteString("  " + dim(line) + "\n")
		}
	}
	sb.WriteString("\n")

	for _, line := range WrapText("Keep the heat going. Don't break the streak.", width-4) {
		sb.WriteString("  " + dim(line) + "\n")
	}
	sb.WriteString("\n")

	return sb.String()
}
