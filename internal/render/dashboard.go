package render

import (
	"fmt"
	"os"
	"strings"

	"thermal/internal/thermal"
)

func RenderDashboard(toolName string, summary thermal.Summary, daily []thermal.DailyRow, dbPath string, weeks int, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

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

	muted := func(s string) string { return ColorCode(colors, "38;5;245", s) }
	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }

	var sb strings.Builder
	sb.WriteString("\n")

	tokenLabel := "tokens"
	switch toolName {
	case "command-code":
		tokenLabel = "commands"
	case "codewhale":
		tokenLabel = "sessions"
	case "Devin":
		tokenLabel = "sessions"
	case "Codex":
		tokenLabel = "messages"
	case "Agy":
		tokenLabel = "steps"
	}

	sb.WriteString(fmt.Sprintf("  %s  %s %s / %d weeks  %s\n",
		highlight(toolName+" activity"),
		highlight(thermal.CompactNumber(visibleTokens)),
		tokenLabel,
		weeks,
		muted(thermal.FormatPath(dbPath)),
	))
	sb.WriteString("\n")

	for _, line := range RenderHeatmap(activity, weeks, colors) {
		sb.WriteString(line + "\n")
	}

	allTime := summary.LifetimeTokens

	sb.WriteString(fmt.Sprintf("  %d active days  %s  %d day streak  %s  %d best  %s  %s all-time\n",
		visibleActive, muted("|"), current, muted("|"), longest, muted("|"), thermal.CompactNumber(allTime),
	))
	sb.WriteString("\n")

	return sb.String()
}
