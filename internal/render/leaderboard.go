package render

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

func medals(rank int, colors bool) string {
	prefix := fmt.Sprintf(" %d.", rank)
	if !colors {
		return prefix
	}
	switch rank {
	case 1:
		return "\033[1;33m" + prefix + reset
	case 2:
		return "\033[37m" + prefix + reset
	case 3:
		return "\033[33m" + prefix + reset
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
				bar += "\033[38;5;40m" + "█" + reset
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
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }
	green := func(s string) string { return ColorCode(colors, "38;5;40", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }

	tokenResults := make([]thermal.ToolResult, 0)
	activityResults := make([]thermal.ToolResult, 0)
	for _, r := range results {
		switch r.Tool {
		case thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolDevin, thermal.ToolCodex, thermal.ToolCodewhale, thermal.ToolZCode, thermal.ToolGrok, thermal.ToolClaude:
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
				if a.Summary.Cost != b.Summary.Cost {
					return a.Summary.Cost > b.Summary.Cost
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
		sb.WriteString(fmt.Sprintf("  %s\n", highlight("Token Warriors")))
		sb.WriteString(fmt.Sprintf("   %s  %s %s  %s  %s   %s  %s\n",
			thermal.PadRight("#", 3), thermal.PadRight("Tool", 14), thermal.PadRight("Strk", 6), thermal.PadRight("Best", 6), thermal.PadRight("Days", 6), thermal.PadRight("Tokens", 10), "Cost",
		))
		sb.WriteString(fmt.Sprintf("   %s\n", dim(strings.Repeat("─", 65))))

		for i, r := range tokenResults {
			rank := i + 1
			medal := medals(rank, colors)

			nameStr := thermal.PadRight(r.Name, 14)
			if rank == 1 {
				nameStr = gold(r.Name) + strings.Repeat(" ", 14-len(r.Name))
			}

			streakPlain := fmt.Sprintf("%dd", r.CurrentStreak)
			streakStr := thermal.PadLeft(streakPlain, 6)
			if r.CurrentStreak > 0 {
				streakStr = strings.Repeat(" ", 6-len(streakPlain)) + green(streakPlain)
			}

			bestStr := thermal.PadLeft(fmt.Sprintf("%dd", r.LongestStreak), 6)
			activeStr := thermal.PadLeft(fmt.Sprintf("%dd", r.ActiveDays), 6)
			activityStr := thermal.PadRight(fmt.Sprintf("%s tok", thermal.CompactNumber(r.TotalActivity)), 10)

			costStr := "—"
			if r.Summary.Cost > 0 {
				if r.Summary.Cost < 0.01 {
					costStr = fmt.Sprintf("$%.4f", r.Summary.Cost)
				} else {
					costStr = fmt.Sprintf("$%.2f", r.Summary.Cost)
				}
			}

			sb.WriteString(fmt.Sprintf("  %s %s %s  %s  %s   %s  %s\n",
				medal, nameStr, streakStr, bestStr, activeStr, activityStr, costStr,
			))
		}
		sb.WriteString("\n")
	}

	if len(activityResults) > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", highlight("Activity Hunters")))
		sb.WriteString(fmt.Sprintf("   %s  %s %s  %s  %s   %s\n",
			thermal.PadRight("#", 3), thermal.PadRight("Tool", 14), thermal.PadRight("Strk", 6), thermal.PadRight("Best", 6), thermal.PadRight("Days", 6), "Activity",
		))
		sb.WriteString(fmt.Sprintf("   %s\n", dim(strings.Repeat("─", 55))))

		for i, r := range activityResults {
			rank := i + 1
			medal := medals(rank, colors)

			nameStr := thermal.PadRight(r.Name, 14)
			if rank == 1 {
				nameStr = gold(r.Name) + strings.Repeat(" ", 14-len(r.Name))
			}

			streakPlain := fmt.Sprintf("%dd", r.CurrentStreak)
			streakStr := thermal.PadLeft(streakPlain, 6)
			if r.CurrentStreak > 0 {
				streakStr = strings.Repeat(" ", 6-len(streakPlain)) + green(streakPlain)
			}

			bestStr := thermal.PadLeft(fmt.Sprintf("%dd", r.LongestStreak), 6)
			activeStr := thermal.PadLeft(fmt.Sprintf("%dd", r.ActiveDays), 6)

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
			activityStr := fmt.Sprintf("%s %s", thermal.CompactNumber(r.TotalActivity), actLabel)

			sb.WriteString(fmt.Sprintf("  %s %s %s  %s  %s   %s\n",
				medal, nameStr, streakStr, bestStr, activeStr, activityStr,
			))
		}
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

	// Cost provenance. The leaderboard prints recorded cost only, which is not
	// the number a report prints: reports estimate the days a source left
	// blank. Saying so here stops the two from looking like a bug.
	var recorded float64
	recording := 0
	for _, r := range results {
		if r.Summary.Cost > 0 {
			recorded += r.Summary.Cost
			recording++
		}
	}
	if recording > 0 {
		line := fmt.Sprintf("Recorded cost: %s from %d tools.", formatCost(recorded), recording)
		if estimate {
			line += " Reports estimate the rest from pricing data; the leaderboard does not."
		} else {
			line += " Estimates are off. Reports would show the same figure."
		}
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim(line)))
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n\n", dim("Keep the heat going. Don't break the streak.")))

	return sb.String()
}
