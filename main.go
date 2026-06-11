package main

import (
	"bufio"
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

type Tool string

const (
	ToolAll         Tool = "all"
	ToolAuto        Tool = "auto"
	ToolMiMoCode    Tool = "mimocode"
	ToolOpenCode    Tool = "opencode"
	ToolCodex       Tool = "codex"
	ToolDevin       Tool = "devin"
	ToolAgy         Tool = "agy"
	ToolCommandCode Tool = "command-code"
	ToolCodewhale   Tool = "codewhale"
)

type Options struct {
	tool    string
	dbPath  string
	weeks   int
	json    bool
	noColor bool
}

type Summary struct {
	Tool             string  `json:"tool"`
	Sessions         int     `json:"sessions"`
	LifetimeTokens   int64   `json:"lifetimeTokens"`
	InputTokens      int64   `json:"inputTokens"`
	OutputTokens     int64   `json:"outputTokens"`
	ReasoningTokens  int64   `json:"reasoningTokens"`
	CacheTokens      int64   `json:"cacheTokens"`
	Cost             float64 `json:"cost"`
	LongestSessionMs int64   `json:"longestSessionMs"`
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

type ToolResult struct {
	Tool         Tool
	Name         string
	Summary      Summary
	Daily        []DailyRow
	CurrentStreak int
	LongestStreak int
	ActiveDays    int
	TotalActivity int64
	DataPath      string
}

var reset = "\033[0m"

func usage() string {
	return `Usage: thermal [options]

Don't break the streak.
Terminal usage profile for AI coding tools.

Supported tools:
  all           Show all tools as leaderboard (default)
  mimocode      MiMoCode
  opencode      OpenCode
  codex         Codex CLI
  devin         Devin
  agy           Agy (Antigravity)
  command-code  command-code-ai
  codewhale     codewhale

Options:
  --tool <name>   Tool to show (default: all)
  --db <path>     Override database/data path
  --weeks <num>   Heatmap width in weeks, 4-104 (default: 52)
  --json          Output JSON instead of dashboard
  --no-color      Disable ANSI colors
  -h, --help      Show this help`
}

func parseArgs() Options {
	var opts Options
	flag.StringVar(&opts.tool, "tool", "all", "Tool: all, mimocode, opencode, codex, agy, command-code, codewhale")
	flag.StringVar(&opts.dbPath, "db", "", "Override database/data path")
	flag.IntVar(&opts.weeks, "weeks", 52, "Heatmap width in weeks (4-104)")
	flag.BoolVar(&opts.json, "json", false, "Output JSON instead of dashboard")
	flag.BoolVar(&opts.noColor, "no-color", false, "Disable ANSI colors")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage())
	}
	flag.Parse()

	// Handle flags that appear after positional arg
	remaining := flag.Args()
	if len(remaining) > 1 {
		for i := 1; i < len(remaining); i++ {
			arg := remaining[i]
			switch {
			case arg == "--no-color":
				opts.noColor = true
			case arg == "--json":
				opts.json = true
			case strings.HasPrefix(arg, "--weeks="):
				if v, err := strconv.Atoi(strings.TrimPrefix(arg, "--weeks=")); err == nil {
					opts.weeks = v
				}
			case arg == "--weeks" && i+1 < len(remaining):
				if v, err := strconv.Atoi(remaining[i+1]); err == nil {
					opts.weeks = v
					i++
				}
			case strings.HasPrefix(arg, "--db="):
				opts.dbPath = strings.TrimPrefix(arg, "--db=")
			case arg == "--db" && i+1 < len(remaining):
				opts.dbPath = remaining[i+1]
				i++
			case strings.HasPrefix(arg, "--tool="):
				opts.tool = strings.TrimPrefix(arg, "--tool=")
			case arg == "--tool" && i+1 < len(remaining):
				opts.tool = remaining[i+1]
				i++
			}
		}
	}

	// Treat first positional arg as tool name
	if len(remaining) > 0 && !strings.HasPrefix(remaining[0], "-") {
		opts.tool = remaining[0]
	}

	if opts.weeks < 4 || opts.weeks > 104 {
		fmt.Fprintf(os.Stderr, "thermal: --weeks must be between 4 and 104\n")
		os.Exit(1)
	}

	return opts
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

type toolInfo struct {
	dbPath  string
	dataDir string
	name    string
	loader  func(string) (Summary, []DailyRow, error)
}

func allTools() map[Tool]toolInfo {
	home := homeDir()
	return map[Tool]toolInfo{
		ToolMiMoCode: {
			dbPath:  filepath.Join(home, ".local", "share", "mimocode", "mimocode.db"),
			dataDir: filepath.Join(home, ".local", "share", "mimocode"),
			name:    "MiMoCode",
			loader: func(p string) (Summary, []DailyRow, error) { return loadSqliteData(p) },
		},
		ToolOpenCode: {
			dbPath:  filepath.Join(home, ".local", "share", "opencode", "opencode.db"),
			dataDir: filepath.Join(home, ".local", "share", "opencode"),
			name:    "OpenCode",
			loader: func(p string) (Summary, []DailyRow, error) { return loadSqliteData(p) },
		},
		ToolCodex: {
			dataDir: filepath.Join(home, ".codex"),
			name:    "Codex",
			loader:  loadCodexData,
		},
		ToolDevin: {
			dbPath:  filepath.Join(home, ".local", "share", "devin", "cli", "sessions.db"),
			dataDir: filepath.Join(home, ".local", "share", "devin", "cli"),
			name:    "Devin",
			loader:  loadDevinData,
		},
		ToolAgy: {
			dataDir: filepath.Join(home, ".gemini", "antigravity"),
			name:    "Agy",
			loader:  loadAgyData,
		},
		ToolCommandCode: {
			dataDir: filepath.Join(home, ".commandcode"),
			name:    "command-code",
			loader:  loadCommandCodeData,
		},
		ToolCodewhale: {
			dataDir: filepath.Join(home, ".codewhale"),
			name:    "codewhale",
			loader:  loadCodewhaleData,
		},
	}
}

var toolAliases = map[string]Tool{
	"mimo":   ToolMiMoCode,
	"mimo-":  ToolMiMoCode,
	"mimocode": ToolMiMoCode,
	"oc":     ToolOpenCode,
	"opencode": ToolOpenCode,
	"codex":  ToolCodex,
	"devin":  ToolDevin,
	"agy":    ToolAgy,
	"cmd":    ToolCommandCode,
	"cc":     ToolCommandCode,
	"commandcode": ToolCommandCode,
	"command-code": ToolCommandCode,
	"whale":  ToolCodewhale,
	"codewhale": ToolCodewhale,
	"all":    ToolAll,
	"auto":   ToolAuto,
}

func resolveTool(name string) (Tool, bool) {
	if t, ok := toolAliases[strings.ToLower(name)]; ok {
		return t, true
	}
	return "", false
}

func loadToolData(t Tool, info toolInfo, dbPath string) (Summary, []DailyRow, string, error) {
	switch t {
	case ToolMiMoCode, ToolOpenCode, ToolDevin:
		p := dbPath
		if p == "" {
			p = info.dbPath
		}
		if p == "" {
			return Summary{}, nil, "", fmt.Errorf("no database path configured for %s", info.name)
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return Summary{}, nil, "", fmt.Errorf("database not found: %s", p)
		}
		s, d, err := info.loader(p)
		s.Tool = info.name
		return s, d, info.dbPath, err
	default:
		dir := info.dataDir
		if dbPath != "" {
			dir = dbPath
		}
		if dir == "" {
			return Summary{}, nil, "", fmt.Errorf("no data directory configured for %s", info.name)
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return Summary{}, nil, "", fmt.Errorf("data directory not found: %s", dir)
		}
		s, d, err := info.loader(dir)
		s.Tool = info.name
		dataPath := filepath.Join(dir, "history.jsonl")
		if t == ToolAgy {
			dataPath = filepath.Join(dir, "brain")
		} else if t == ToolCodewhale {
			dataPath = filepath.Join(dir, "sessions")
		}
		return s, d, dataPath, err
	}
}

func detectTool(opts *Options) Tool {
	if opts.tool != "auto" && opts.tool != "all" {
		t, ok := resolveTool(opts.tool)
		if !ok {
			fmt.Fprintf(os.Stderr, "thermal: unknown tool: %s\n\n", opts.tool)
			fmt.Fprintln(os.Stderr, "Available tools:")
			fmt.Fprintln(os.Stderr, "  mimocode, mimo       MiMoCode")
			fmt.Fprintln(os.Stderr, "  opencode, oc         OpenCode")
			fmt.Fprintln(os.Stderr, "  codex                Codex CLI")
			fmt.Fprintln(os.Stderr, "  devin                Devin")
			fmt.Fprintln(os.Stderr, "  agy                  Agy (Antigravity)")
			fmt.Fprintln(os.Stderr, "  command-code, cmd    command-code-ai")
			fmt.Fprintln(os.Stderr, "  codewhale, whale     codewhale")
			fmt.Fprintln(os.Stderr, "  all                  Show leaderboard (default)")
			os.Exit(1)
		}
		return t
	}

	// auto: find first installed tool
	tools := allTools()
	for _, t := range []Tool{ToolMiMoCode, ToolOpenCode, ToolCodex, ToolAgy, ToolCommandCode, ToolCodewhale} {
		info := tools[t]
		if info.dbPath != "" {
			if _, err := os.Stat(info.dbPath); err == nil {
				return t
			}
		}
		if info.dataDir != "" {
			if _, err := os.Stat(info.dataDir); err == nil {
				return t
			}
		}
	}

	fmt.Fprintf(os.Stderr, "thermal: no supported tool data found\n")
	os.Exit(1)
	return ""
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

func padRight(s string, width int) string {
	if n := width - len(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func padLeft(s string, width int) string {
	if n := width - len(s); n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

func formatPath(p string) string {
	home := homeDir()
	if home != "" && strings.HasPrefix(p, home+"/") {
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

	// Current streak: walk backward from latest active day (not necessarily today)
	current = 0
	if len(sorted) > 0 {
		cursor, _ := time.Parse("2006-01-02", sorted[len(sorted)-1])
		for {
			if !days[localDay(cursor)] {
				break
			}
			current++
			cursor = cursor.Add(-24 * time.Hour)
		}
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
				a := activity[localDay(date)]
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

// --- SQLite-based loaders (MiMoCode, OpenCode) ---

func loadSqliteData(dbPath string) (Summary, []DailyRow, error) {
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

// --- JSONL loader (command-code, codex) ---

func loadJsonlData(dataDir string, fieldTimestamp string, useMillis bool) (Summary, []DailyRow, error) {
	historyPath := filepath.Join(dataDir, "history.jsonl")
	f, err := os.Open(historyPath)
	if err != nil {
		return Summary{}, nil, fmt.Errorf("cannot open %s: %w", historyPath, err)
	}
	defer f.Close()

	dayCounts := make(map[string]int)
	var totalCommands int64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		var ts int64
		if v, ok := raw[fieldTimestamp].(float64); ok {
			ts = int64(v)
		}
		if ts == 0 {
			continue
		}

		var t time.Time
		if useMillis {
			t = time.UnixMilli(ts)
		} else {
			t = time.Unix(ts, 0)
		}
		day := localDay(t)
		dayCounts[day]++
		totalCommands++
	}

	var daily []DailyRow
	for day, count := range dayCounts {
		daily = append(daily, DailyRow{
			Day:   day,
			Tokens: int64(count),
			Turns:  count,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	summary := Summary{
		Sessions:      0,
		LifetimeTokens: totalCommands,
	}

	return summary, daily, nil
}

func loadCommandCodeData(dataDir string) (Summary, []DailyRow, error) {
	summary, daily, err := loadJsonlData(dataDir, "t", true)
	summary.Tool = "command-code"
	return summary, daily, err
}

func loadCodexData(dataDir string) (Summary, []DailyRow, error) {
	summary, daily, err := loadJsonlData(dataDir, "ts", false)
	summary.Tool = "Codex"
	return summary, daily, err
}

// --- Devin loader (SQLite sessions) ---

func loadDevinData(dbPath string) (Summary, []DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return Summary{}, nil, err
	}
	defer db.Close()

	var summary Summary
	err = db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&summary.Sessions)
	if err != nil {
		return Summary{}, nil, err
	}
	summary.LifetimeTokens = int64(summary.Sessions)

	rows, err := db.Query(`
		SELECT
			date(created_at, 'unixepoch', 'localtime') AS day,
			COUNT(*) AS turns
		FROM sessions
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
		if err := rows.Scan(&r.Day, &r.Turns); err != nil {
			return Summary{}, nil, err
		}
		r.Tokens = int64(r.Turns)
		daily = append(daily, r)
	}

	return summary, daily, nil
}

// --- Agy loader (transcript JSONL files) ---

func loadAgyData(dataDir string) (Summary, []DailyRow, error) {
	brainDir := filepath.Join(dataDir, "brain")
	entries, err := os.ReadDir(brainDir)
	if err != nil {
		return Summary{}, nil, fmt.Errorf("cannot read %s: %w", brainDir, err)
	}

	dayCounts := make(map[string]int)
	var totalEntries int64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		transcriptPath := filepath.Join(brainDir, entry.Name(), ".system_generated", "logs", "transcript.jsonl")
		f, err := os.Open(transcriptPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var record struct {
				CreatedAt string `json:"created_at"`
			}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}
			if record.CreatedAt == "" {
				continue
			}
			t, err := time.Parse(time.RFC3339, record.CreatedAt)
			if err != nil {
				continue
			}
			day := localDay(t)
			dayCounts[day]++
			totalEntries++
		}
		f.Close()
	}

	var daily []DailyRow
	for day, count := range dayCounts {
		daily = append(daily, DailyRow{
			Day:   day,
			Tokens: int64(count),
			Turns:  count,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	summary := Summary{
		Tool:          "Agy",
		Sessions:      0,
		LifetimeTokens: totalEntries,
	}

	return summary, daily, nil
}

// --- JSON loader (codewhale) ---

func loadCodewhaleData(dataDir string) (Summary, []DailyRow, error) {
	sessionsDir := filepath.Join(dataDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return Summary{}, nil, fmt.Errorf("cannot read %s: %w", sessionsDir, err)
	}

	dayCounts := make(map[string]int)
	var totalSessions int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name()))
		if err != nil {
			continue
		}
		var session struct {
			Metadata struct {
				CreatedAt string `json:"created_at"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.Metadata.CreatedAt == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, session.Metadata.CreatedAt)
		if err != nil {
			continue
		}
		day := localDay(t)
		dayCounts[day]++
		totalSessions++
	}

	var daily []DailyRow
	for day, count := range dayCounts {
		daily = append(daily, DailyRow{
			Day:   day,
			Tokens: int64(count),
			Turns:  count,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })

	summary := Summary{
		Tool:          "codewhale",
		Sessions:      totalSessions,
		LifetimeTokens: int64(totalSessions),
	}

	return summary, daily, nil
}

// --- Single tool dashboard ---

func renderDashboard(toolName string, summary Summary, daily []DailyRow, dbPath string, weeks int, noColor bool) string {
	colors := !noColor && isTerminal() && os.Getenv("NO_COLOR") == ""

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
	firstSunday := thisSunday.AddDate(0, 0, -(weeks-1)*7)
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
		highlight(compactNumber(visibleTokens)),
		tokenLabel,
		weeks,
		muted(formatPath(dbPath)),
	))
	sb.WriteString("\n")

	for _, line := range renderHeatmap(activity, weeks, colors) {
		sb.WriteString(line + "\n")
	}

	allTime := summary.LifetimeTokens

	sb.WriteString(fmt.Sprintf("  %d active days  %s  %d day streak  %s  %d best  %s  %s all-time\n",
		visibleActive, muted("|"), current, muted("|"), longest, muted("|"), compactNumber(allTime),
	))
	sb.WriteString("\n")

	return sb.String()
}

// --- Leaderboard ---

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

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func renderLeaderboard(results []ToolResult, weeks int, noColor bool) string {
	colors := !noColor && isTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return colorCode(colors, "1;38;5;255", s) }
	gold := func(s string) string { return colorCode(colors, "1;33", s) }
	green := func(s string) string { return colorCode(colors, "38;5;40", s) }
	dim := func(s string) string { return colorCode(colors, "38;5;239", s) }

	// Split into token-based and activity-based tools
	tokenResults := make([]ToolResult, 0)
	activityResults := make([]ToolResult, 0)
	for _, r := range results {
		switch r.Tool {
		case ToolMiMoCode, ToolOpenCode:
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

	// --- Token Warriors ---
	if len(tokenResults) > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", highlight("Token Warriors")))
		sb.WriteString(fmt.Sprintf("   %s  %s %s  %s  %s   %s\n",
			padRight("#", 3), padRight("Tool", 14), padRight("Strk", 6), padRight("Best", 6), padRight("Days", 6), "Tokens",
		))
		sb.WriteString(fmt.Sprintf("   %s\n", dim(strings.Repeat("─", 55))))

		for i, r := range tokenResults {
			rank := i + 1
			medal := medals(rank, colors)

			nameStr := padRight(r.Name, 14)
			if rank == 1 {
				nameStr = gold(r.Name) + strings.Repeat(" ", 14-len(r.Name))
			}

			streakPlain := fmt.Sprintf("%dd", r.CurrentStreak)
			streakStr := padLeft(streakPlain, 6)
			if r.CurrentStreak > 0 {
				streakStr = strings.Repeat(" ", 6-len(streakPlain)) + green(streakPlain)
			}

			bestStr := padLeft(fmt.Sprintf("%dd", r.LongestStreak), 6)
			activeStr := padLeft(fmt.Sprintf("%dd", r.ActiveDays), 6)
			activityStr := fmt.Sprintf("%s tok", compactNumber(r.TotalActivity))

			sb.WriteString(fmt.Sprintf("  %s %s %s  %s  %s   %s\n",
				medal, nameStr, streakStr, bestStr, activeStr, activityStr,
			))
		}
		sb.WriteString("\n")
	}

	// --- Activity Hunters ---
	if len(activityResults) > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", highlight("Activity Hunters")))
		sb.WriteString(fmt.Sprintf("   %s  %s %s  %s  %s   %s\n",
			padRight("#", 3), padRight("Tool", 14), padRight("Strk", 6), padRight("Best", 6), padRight("Days", 6), "Activity",
		))
		sb.WriteString(fmt.Sprintf("   %s\n", dim(strings.Repeat("─", 55))))

		for i, r := range activityResults {
			rank := i + 1
			medal := medals(rank, colors)

			nameStr := padRight(r.Name, 14)
			if rank == 1 {
				nameStr = gold(r.Name) + strings.Repeat(" ", 14-len(r.Name))
			}

			streakPlain := fmt.Sprintf("%dd", r.CurrentStreak)
			streakStr := padLeft(streakPlain, 6)
			if r.CurrentStreak > 0 {
				streakStr = strings.Repeat(" ", 6-len(streakPlain)) + green(streakPlain)
			}

			bestStr := padLeft(fmt.Sprintf("%dd", r.LongestStreak), 6)
			activeStr := padLeft(fmt.Sprintf("%dd", r.ActiveDays), 6)

			var actLabel string
			switch r.Tool {
			case ToolDevin, ToolCodewhale:
				actLabel = "sess"
			case ToolCommandCode:
				actLabel = "cmd"
			case ToolCodex:
				actLabel = "msg"
			case ToolAgy:
				actLabel = "step"
			default:
				actLabel = "act"
			}
			activityStr := fmt.Sprintf("%s %s", compactNumber(r.TotalActivity), actLabel)

			sb.WriteString(fmt.Sprintf("  %s %s %s  %s  %s   %s\n",
				medal, nameStr, streakStr, bestStr, activeStr, activityStr,
			))
		}
		sb.WriteString("\n")
	}

	// Winner callout
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

func main() {
	opts := parseArgs()

	if opts.tool == "all" {
		// Leaderboard mode: load all installed tools
		tools := allTools()
		var results []ToolResult

		toolOrder := []Tool{ToolMiMoCode, ToolOpenCode, ToolCodex, ToolDevin, ToolAgy, ToolCommandCode, ToolCodewhale}
		for _, t := range toolOrder {
			info := tools[t]
			// Check if tool data exists
			if info.dbPath != "" {
				if _, err := os.Stat(info.dbPath); os.IsNotExist(err) {
					continue
				}
			} else if info.dataDir != "" {
				if _, err := os.Stat(info.dataDir); os.IsNotExist(err) {
					continue
				}
			} else {
				continue
			}

			summary, daily, dataPath, err := loadToolData(t, info, "")
			if err != nil {
				continue
			}

			activeDays := make(map[string]bool)
			for _, d := range daily {
				if d.Turns > 0 {
					activeDays[d.Day] = true
				}
			}
			current, longest := computeStreaks(activeDays)

			results = append(results, ToolResult{
				Tool:          t,
				Name:          info.name,
				Summary:       summary,
				Daily:         daily,
				CurrentStreak: current,
				LongestStreak: longest,
				ActiveDays:    len(activeDays),
				TotalActivity: summary.LifetimeTokens,
				DataPath:      dataPath,
			})
		}

		if len(results) == 0 {
			fmt.Fprintf(os.Stderr, "thermal: no supported tool data found\n")
			os.Exit(1)
		}

		if opts.json {
			type jsonResult struct {
				ToolResult
				CurrentStreak int `json:"currentStreak"`
				LongestStreak int `json:"longestStreak"`
			}
			var jsonResults []jsonResult
			for _, r := range results {
				jsonResults = append(jsonResults, jsonResult{
					ToolResult:    r,
					CurrentStreak: r.CurrentStreak,
					LongestStreak: r.LongestStreak,
				})
			}
			out := map[string]interface{}{
				"generatedAt": time.Now().UTC().Format(time.RFC3339),
				"results":     jsonResults,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(out)
			return
		}

		fmt.Print(renderLeaderboard(results, opts.weeks, opts.noColor))
		return
	}

	// Single tool mode
	var tool Tool
	if opts.tool == "auto" {
		tool = detectTool(&opts)
	} else {
		var ok bool
		tool, ok = resolveTool(opts.tool)
		if !ok {
			fmt.Fprintf(os.Stderr, "thermal: unknown tool: %s\n", opts.tool)
			os.Exit(1)
		}
	}
	tools := allTools()
	info := tools[tool]

	summary, daily, dataPath, err := loadToolData(tool, info, opts.dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "thermal: %v\n", err)
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
			CurrentStreak int `json:"currentStreak"`
			LongestStreak int `json:"longestStreak"`
		}
		out := map[string]interface{}{
			"tool":        info.name,
			"dataPath":    dataPath,
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
			"summary":     jsonSummary{Summary: summary, CurrentStreak: current, LongestStreak: longest},
			"daily":       daily,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return
	}

	fmt.Print(renderDashboard(info.name, summary, daily, dataPath, opts.weeks, opts.noColor))
}
