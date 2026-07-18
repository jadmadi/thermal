package render

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
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

	// Extra analytics line: cost, code changes, sessions, agent breakdown.
	var extra []string
	if summary.Cost > 0 {
		extra = append(extra, fmt.Sprintf("$%.2f spent", summary.Cost))
	}
	if summary.LinesAdded > 0 || summary.LinesDeleted > 0 {
		extra = append(extra, fmt.Sprintf("%s+ / %s- lines", thermal.CompactNumber(summary.LinesAdded), thermal.CompactNumber(summary.LinesDeleted)))
	}
	if summary.FilesTouched > 0 {
		extra = append(extra, fmt.Sprintf("%s files", thermal.CompactNumber(summary.FilesTouched)))
	}
	if summary.Sessions > 0 && summary.LongestSessionMs > 0 {
		extra = append(extra, fmt.Sprintf("%d sessions", summary.Sessions))
	}
	if len(extra) > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(extra, muted("  ·  "))))
	}

	// Agent breakdown line (top 3 agents by count).
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
		sb.WriteString(fmt.Sprintf("  %s agents  %s\n", muted("·"), strings.Join(parts, muted("  "))))
	}

	sb.WriteString("\n")

	return sb.String()
}
