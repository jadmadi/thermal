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
func RenderLeaderboard(results []thermal.ToolResult, weeks int, noColor bool, sortKey string, estimate bool) string {
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
		case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolCodex, thermal.ToolCodewhale, thermal.ToolZCode, thermal.ToolGrok, thermal.ToolClaude, thermal.ToolDsh, thermal.ToolHermes:
			tokenResults = append(tokenResults, r)
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

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s  %s\n\n", highlight("THERMAL"), dim("— Don't break the streak.")))

	if len(tokenResults) > 0 {
		var tokenLines []string
		tokenHeaders := []string{"#", "Tool", "Strk", "Best", "Days", "Tokens", "Cost"}
		tokenAlign := []bool{true, false, true, true, true, true, true}
		tokenWidths := []int{3, 14, 6, 6, 6, 11, 11}

		tokenLines = append(tokenLines, formatHeaderRow(tokenHeaders, tokenWidths, tokenAlign, colors))
		tokenLines = append(tokenLines, " "+tableRule(tokenWidths))

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

			costPlain := "—"
			if r.Summary.Cost > 0 {
				if r.Summary.Cost < 0.01 {
					costPlain = fmt.Sprintf("$%.4f", r.Summary.Cost)
				} else {
					costPlain = fmt.Sprintf("$%.2f", r.Summary.Cost)
				}
			} else if estimate && r.EstimatedCost > 0 {
				if r.EstimatedCost < 0.01 {
					costPlain = fmt.Sprintf("~$%.4f", r.EstimatedCost)
				} else {
					costPlain = fmt.Sprintf("~$%.2f", r.EstimatedCost)
				}
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

			tokenLines = append(tokenLines, fmt.Sprintf(" %s  %s  %s  %s  %s  %s  %s",
				rankCell, nameStr, streakStr, bestStr, activeStr, tokStr, costStr,
			))
		}

		sb.WriteString(RenderCard(CardOptions{
			Title:       "Token Warriors",
			RightHeader: "tokens & spend",
			Lines:       tokenLines,
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

		actLines = append(actLines, formatHeaderRow(actHeaders, actWidths, actAlign, colors))
		actLines = append(actLines, " "+tableRule(actWidths))

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
			case thermal.ToolCommandCode:
				actLabel = "msg"
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

			actLines = append(actLines, fmt.Sprintf(" %s  %s  %s  %s  %s  %s",
				rankCell, nameStr, streakStr, bestStr, activeStr, actStr,
			))
		}

		sb.WriteString(RenderCard(CardOptions{
			Title:       "Activity Hunters",
			RightHeader: "messages & steps",
			Lines:       actLines,
			Indent:      2,
			Colors:      colors,
			TitleColor:  theme.Secondary,
			BorderColor: theme.Border,
		}))
		sb.WriteString("\n")
	}

	allResults := append(tokenResults, activityResults...)
	if len(allResults) > 0 {
		best := allResults[0]
		for _, r := range allResults {
			if r.CurrentStreak > best.CurrentStreak {
				best = r
			}
		}
		if best.CurrentStreak > 0 {
			sb.WriteString(fmt.Sprintf("  %s %s is on fire with a %d-day streak!\n",
				gold(">>"), highlight(best.Name), best.CurrentStreak,
			))
		} else {
			sb.WriteString(fmt.Sprintf("  %s No active streaks. Time to code!\n", dim("--")))
		}
	}

	// Cost provenance: distinguish recorded cost from estimated cost.
	var recorded float64
	var estimated float64
	recording := 0
	estimating := 0
	for _, r := range results {
		if r.Summary.Cost > 0 {
			recorded += r.Summary.Cost
			recording++
		} else if estimate && r.EstimatedCost > 0 {
			estimated += r.EstimatedCost
			estimating++
		}
	}
	if recording > 0 || estimating > 0 {
		var parts []string
		if recording > 0 {
			parts = append(parts, fmt.Sprintf("Recorded cost: %s from %d tools.", formatCost(recorded), recording))
		}
		if estimate {
			if estimating > 0 {
				parts = append(parts, fmt.Sprintf("Estimated: ~%s across %d tools (~ prefix).", formatCost(estimated), estimating))
			} else {
				parts = append(parts, "Reports estimate the rest from pricing data; the leaderboard does not.")
			}
		} else {
			parts = append(parts, "Estimates are off. Reports would show the same figure.")
		}
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim(strings.Join(parts, " "))))
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n\n", dim("Keep the heat going. Don't break the streak.")))

	return sb.String()
}
