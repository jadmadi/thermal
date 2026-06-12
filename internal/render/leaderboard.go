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

func RenderLeaderboard(results []thermal.ToolResult, weeks int, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }
	green := func(s string) string { return ColorCode(colors, "38;5;40", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }

	tokenResults := make([]thermal.ToolResult, 0)
	activityResults := make([]thermal.ToolResult, 0)
	for _, r := range results {
		switch r.Tool {
		case thermal.ToolMiMoCode, thermal.ToolOpenCode:
			tokenResults = append(tokenResults, r)
		default:
			activityResults = append(activityResults, r)
		}
	}

	sort.Slice(tokenResults, func(i, j int) bool {
		if tokenResults[i].CurrentStreak != tokenResults[j].CurrentStreak {
			return tokenResults[i].CurrentStreak > tokenResults[j].CurrentStreak
		}
		if tokenResults[i].LongestStreak != tokenResults[j].LongestStreak {
			return tokenResults[i].LongestStreak > tokenResults[j].LongestStreak
		}
		return tokenResults[i].TotalActivity > tokenResults[j].TotalActivity
	})

	sort.Slice(activityResults, func(i, j int) bool {
		if activityResults[i].CurrentStreak != activityResults[j].CurrentStreak {
			return activityResults[i].CurrentStreak > activityResults[j].CurrentStreak
		}
		if activityResults[i].LongestStreak != activityResults[j].LongestStreak {
			return activityResults[i].LongestStreak > activityResults[j].LongestStreak
		}
		return activityResults[i].TotalActivity > activityResults[j].TotalActivity
	})

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
		sb.WriteString(fmt.Sprintf("   %s  %s %s  %s  %s   %s\n",
			thermal.PadRight("#", 3), thermal.PadRight("Tool", 14), thermal.PadRight("Strk", 6), thermal.PadRight("Best", 6), thermal.PadRight("Days", 6), "Tokens",
		))
		sb.WriteString(fmt.Sprintf("   %s\n", dim(strings.Repeat("─", 55))))

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
			activityStr := fmt.Sprintf("%s tok", thermal.CompactNumber(r.TotalActivity))

			sb.WriteString(fmt.Sprintf("  %s %s %s  %s  %s   %s\n",
				medal, nameStr, streakStr, bestStr, activeStr, activityStr,
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
			case thermal.ToolDevin, thermal.ToolCodewhale:
				actLabel = "sess"
			case thermal.ToolCommandCode:
				actLabel = "cmd"
			case thermal.ToolCodex:
				actLabel = "msg"
			case thermal.ToolAgy:
				actLabel = "step"
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

	sb.WriteString(fmt.Sprintf("\n  %s\n\n", dim("Keep the heat going. Don't break the streak.")))

	return sb.String()
}
