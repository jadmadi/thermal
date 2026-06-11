package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Options struct {
	dbPath  string
	weeks   int
	json    bool
	noColor bool
}

type Summary struct {
	Sessions        int     `json:"sessions"`
	LifetimeTokens  int64   `json:"lifetimeTokens"`
	InputTokens     int64   `json:"inputTokens"`
	OutputTokens    int64   `json:"outputTokens"`
	ReasoningTokens int64   `json:"reasoningTokens"`
	CacheTokens     int64   `json:"cacheTokens"`
	Cost            float64 `json:"cost"`
	LongestSessionMs int64  `json:"longestSessionMs"`
}

type DailyRow struct {
	Day   string `json:"day"`
	Tokens int64 `json:"tokens"`
	Turns  int   `json:"turns"`
}

type DayActivity struct {
	Tokens int64
	Turns  int
}

var reset = "\033[0m"

func usage() string {
	return `Usage: mimo-streak [options]

Terminal usage profile generated from the MiMoCode SQLite database.

Options:
  --db <path>       SQLite database path (default: ~/.local/share/mimocode/mimocode.db)
  --weeks <number>  Heatmap width in weeks, from 4 to 104 (default: 52)
  --json            Print computed data as JSON instead of the dashboard
  --no-color        Disable ANSI colors
  -h, --help        Show this help`
}

func parseArgs() Options {
	var opts Options
	flag.StringVar(&opts.dbPath, "db", defaultDbPath(), "SQLite database path")
	flag.IntVar(&opts.weeks, "weeks", 52, "Heatmap width in weeks (4-104)")
	flag.BoolVar(&opts.json, "json", false, "Output JSON instead of dashboard")
	flag.BoolVar(&opts.noColor, "no-color", false, "Disable ANSI colors")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage())
	}
	flag.Parse()

	if opts.weeks < 4 || opts.weeks > 104 {
		fmt.Fprintf(os.Stderr, "mimo-streak: --weeks must be between 4 and 104\n")
		os.Exit(1)
	}

	return opts
}

func defaultDbPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "mimocode", "mimocode.db")
}

func localDay(t time.Time) string {
	return t.Format("2006-01-02")
}

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func compactNumber(v int64) string {
	if v >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(v)/1_000_000_000)
	}
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(v)/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(v)/1_000)
	}
	return strconv.FormatInt(v, 10)
}

func formatPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if strings.HasPrefix(p, home+"/") {
		return "~" + p[len(home):]
	}
	return p
}

func computeStreaks(days map[string]bool) (current int, longest int) {
	sorted := make([]string, 0, len(days))
	for d := range days {
		sorted = append(sorted, d)
	}
	sort.Strings(sorted)

	longest = 0
	run := 0
	var prev time.Time
	for _, d := range sorted {
		t, _ := time.Parse("2006-01-02", d)
		if !prev.IsZero() && t.Sub(prev) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
		prev = t
	}

	current = 0
	cursor := startOfToday()
	for {
		key := localDay(cursor)
		if !days[key] {
			break
		}
		current++
		cursor = cursor.Add(-24 * time.Hour)
	}

	return current, longest
}

func colorCode(enabled bool, code string, value string) string {
	if enabled {
		return "\033[" + code + "m" + value + reset
	}
	return value
}

func activityThresholds(activity map[string]DayActivity) [3]int64 {
	var nonzero []int64
	for _, a := range activity {
		if a.Tokens > 0 {
			nonzero = append(nonzero, a.Tokens)
		}
	}
	if len(nonzero) == 0 {
		return [3]int64{0, 0, 0}
	}
	sort.Slice(nonzero, func(i, j int) bool { return nonzero[i] < nonzero[j] })
	at := func(q float64) int64 {
		idx := int(math.Floor(float64(len(nonzero)-1) * q))
		return nonzero[idx]
	}
	return [3]int64{at(0.25), at(0.5), at(0.75)}
}

func activityLevel(a DayActivity, thresholds [3]int64) int {
	if a.Tokens <= 0 {
		return 1
	}
	if a.Tokens <= thresholds[0] {
		return 1
	}
	if a.Tokens <= thresholds[1] {
		return 2
	}
	if a.Tokens <= thresholds[2] {
		return 3
	}
	return 4
}

func heatCell(level int, colors bool) string {
	if !colors {
		return []string{"□", "░", "▒", "▓", "█"}[level]
	}
	palette := []string{"38;5;238", "38;5;22", "38;5;28", "38;5;34", "38;5;40"}
	return colorCode(true, palette[level], "■")
}

func renderHeatmap(activity map[string]DayActivity, weeks int, colors bool) []string {
	today := startOfToday()
	thisSunday := today.AddDate(0, 0, -int(today.Weekday()))
	firstSunday := thisSunday.AddDate(0, 0, -(weeks-1)*7)
	thresholds := activityThresholds(activity)

	monthLine := make([]string, weeks)
	for i := range monthLine {
		monthLine[i] = " "
	}
	lastMonth := -1
	for w := 0; w < weeks; w++ {
		date := firstSunday.AddDate(0, 0, w*7)
		m := int(date.Month()) - 1
		if m != lastMonth {
			label := date.Format("Jan")[:1] // first char
			monthLine[w] = label
			if len(date.Format("Jan")) > 1 {
				if w+1 < weeks {
					monthLine[w] = date.Format("Jan")[:min(2, len(date.Format("Jan")))]
				}
			}
			// Actually let's use full month abbreviation placed at week position
			monthLine[w] = ""
			monthAbbr := date.Format("Jan")
			for i, ch := range monthAbbr {
				if w+i < weeks {
					monthLine[w+i] = string(ch)
				}
			}
			lastMonth = m
		}
	}

	labels := []string{"   ", "Mon", "   ", "Wed", "   ", "Fri", "   "}
	lines := []string{"      " + strings.Join(monthLine, "")}

	for weekday := 0; weekday < 7; weekday++ {
		row := "  " + labels[weekday] + " "
		for w := 0; w < weeks; w++ {
			date := firstSunday.AddDate(0, 0, w*7+weekday)
			if date.After(today) {
				row += " "
			} else {
				a, ok := activity[localDay(date)]
				if !ok {
					a = DayActivity{}
				}
				row += heatCell(activityLevel(a, thresholds), colors)
			}
		}
		lines = append(lines, row)
	}

	legend := ""
	for i := 0; i < 5; i++ {
		legend += heatCell(i, colors)
	}
	lines = append(lines, "      Less "+legend+" More")
	return lines
}

func loadData(dbPath string) (Summary, []DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return Summary{}, nil, err
	}
	defer db.Close()

	var summary Summary
	err = db.QueryRow(`
		SELECT
			COUNT(DISTINCT session_id),
			COALESCE(SUM(
				COALESCE(CAST(json_extract(data, '$.tokens.input') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.output') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.reasoning') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.write') AS INTEGER), 0)
			), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.tokens.input') AS INTEGER), 0)), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.tokens.output') AS INTEGER), 0)), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.tokens.reasoning') AS INTEGER), 0)), 0),
			COALESCE(SUM(
				COALESCE(CAST(json_extract(data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.write') AS INTEGER), 0)
			), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.cost') AS REAL), 0)), 0)
		FROM message
		WHERE json_extract(data, '$.role') = 'assistant'
	`).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.InputTokens,
		&summary.OutputTokens, &summary.ReasoningTokens, &summary.CacheTokens, &summary.Cost)
	if err != nil {
		return Summary{}, nil, err
	}

	db.QueryRow(`SELECT COALESCE(MAX(time_updated - time_created), 0) FROM session`).
		Scan(&summary.LongestSessionMs)

	rows, err := db.Query(`
		SELECT
			date(time_created / 1000, 'unixepoch', 'localtime') AS day,
			SUM(
				COALESCE(CAST(json_extract(data, '$.tokens.input') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.output') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.reasoning') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.write') AS INTEGER), 0)
			) AS tokens,
			COUNT(*) AS turns
		FROM message
		WHERE json_extract(data, '$.role') = 'assistant'
		GROUP BY day
		ORDER BY day
	`)
	if err != nil {
		return Summary{}, nil, err
	}
	defer rows.Close()

	var daily []DailyRow
	for rows.Next() {
		var r DailyRow
		if err := rows.Scan(&r.Day, &r.Tokens, &r.Turns); err != nil {
			return Summary{}, nil, err
		}
		daily = append(daily, r)
	}

	return summary, daily, nil
}

func renderDashboard(opts Options, summary Summary, daily []DailyRow) string {
	colors := !opts.noColor && os.Getenv("NO_COLOR") == ""
	if os.Getenv("TERM") == "" {
		colors = false
	}

	activeDays := make(map[string]bool)
	for _, d := range daily {
		if d.Turns > 0 {
			activeDays[d.Day] = true
		}
	}
	current, longest := computeStreaks(activeDays)

	activity := make(map[string]DayActivity)
	for _, d := range daily {
		activity[d.Day] = DayActivity{Tokens: d.Tokens, Turns: d.Turns}
	}

	today := startOfToday()
	thisSunday := today.AddDate(0, 0, -int(today.Weekday()))
	firstSunday := thisSunday.AddDate(0, 0, -(opts.weeks-1)*7)
	firstDay := localDay(firstSunday)
	todayStr := localDay(today)

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

	muted := func(s string) string { return colorCode(colors, "38;5;245", s) }
	highlight := func(s string) string { return colorCode(colors, "1;38;5;255", s) }

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s  %s tokens / %d weeks  %s\n",
		highlight("MiMoCode activity"),
		highlight(compactNumber(visibleTokens)),
		opts.weeks,
		muted(formatPath(opts.dbPath)),
	))
	sb.WriteString("\n")

	for _, line := range renderHeatmap(activity, opts.weeks, colors) {
		sb.WriteString(line + "\n")
	}

	sb.WriteString(fmt.Sprintf("  %d active days  %s  %d day streak  %s  %d best  %s  %s all-time\n",
		visibleActive, muted("|"), current, muted("|"), longest, muted("|"), compactNumber(summary.LifetimeTokens),
	))
	sb.WriteString("\n")

	return sb.String()
}

func main() {
	opts := parseArgs()

	if _, err := os.Stat(opts.dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "mimo-streak: database not found: %s\n", opts.dbPath)
		os.Exit(1)
	}

	summary, daily, err := loadData(opts.dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mimo-streak: %v\n", err)
		os.Exit(1)
	}

	activeDays := make(map[string]bool)
	for _, d := range daily {
		if d.Turns > 0 {
			activeDays[d.Day] = true
		}
	}
	current, longest := computeStreaks(activeDays)

	if opts.json {
		type jsonSummary struct {
			Summary
			CurrentStreak  int `json:"currentStreak"`
			LongestStreak  int `json:"longestStreak"`
		}
		out := map[string]interface{}{
			"database":    opts.dbPath,
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
			"summary":     jsonSummary{Summary: summary, CurrentStreak: current, LongestStreak: longest},
			"daily":       daily,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return
	}

	fmt.Print(renderDashboard(opts, summary, daily))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
