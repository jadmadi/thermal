// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	"github.com/jadmadi/thermal/internal/audit"
	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/pricing"
	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/share"
	"github.com/jadmadi/thermal/internal/thermal"
	"github.com/jadmadi/thermal/internal/tui"
	"github.com/jadmadi/thermal/internal/version"
)

func usage() string {
	return `Usage: thermal [options] [tool] [report]

Don't break the streak.
Terminal usage profile for AI coding tools.

Commands:
  dashboard      Interactive dashboard (needs a terminal)
  daily          Daily report (tokens and cost per day)
  weekly         Weekly report
  monthly        Monthly report
  projects       Tokens and cost per project, ranked
  models         Tokens and estimated cost per model, ranked
  mix            Tool or model mix over time, with switching stats
  stats          Daily distribution: percentiles, weekday, outliers
  trend          Daily trend fit with a month-end projection
  replay         Simulate workload against subscriptions & API pricing
  yield          Token yield & code output delta telemetry
  receipt        Verifiable work receipts & session verification outcomes
  share          Generate a stateless, private share URL for your streak
  audit          Audit local agent setup and context health
  license        Show license, dual-licensing & commercial terms
  upgrade        Self-upgrade to the latest release
  version        Show version info

Reports accept an optional tool: "thermal opencode weekly",
"thermal weekly" (all tools). Tool defaults to all.

Cost: recorded cost comes from the tool source; estimated cost
(prefixed with ~) is calculated from models.dev pricing for tools
that record model usage without costs. Reports name the split
under the total. Run with --no-estimate to see recorded cost alone.
Projects collapse to the nearest git root and merge across tools.
--sort picks the ranking: streak|tokens|cost for the leaderboard,
tokens|cost|days|recent for projects, tokens|cost for models.
--metric tokens|cost applies to mix, stats, and trend.
--by tool|model and --grain day|week|month apply to mix.
--against <model> and --compare <plans> apply to replay.

Supported tools:
  all           Show all tools as leaderboard (default)
  mimocode      MiMoCode
  opencode      OpenCode
  codex         Codex CLI
  devin         Devin
  agy           Agy (Antigravity)
  command-code  command-code-ai
  codewhale     codewhale
  zcode         ZCode
  grok          Grok
  muse          Muse
  claude        Claude Code
  droid         Droid
  dsh           DeepSeek (DSH)
  hermes        Nous Hermes

Options:
  --tool <name>      Select single tool (default: all)
  --db <path>        Override database/data path
  --weeks <N>        Heatmap width in weeks (default: 52)
  --since <date>     Report start date (YYYY-MM-DD)
  --until <date>     Report end date (YYYY-MM-DD)
  --last <N>         Report last N days, weeks, or months
  --order <asc|desc> Sort order for report rows (default: desc)
  --sort <field>     Ranking: streak|tokens|cost (leaderboard), tokens|cost|days|recent (projects), tokens|cost (models)
  --top <N>          Limit output rows for projects and models (default: all)
  --metric <field>   Analytics metric: tokens or cost (default: tokens)
  --by <tool|model>  Mix breakdown dimension (default: tool)
  --grain <grain>    Mix aggregation window: day, week, or month (default: week)
  --against <model>  Target model for replay simulation
  --compare <plans>  Plans to compare in replay: comma-separated or all
  --breakdown        Show per-model rows under each period (daily, weekly, monthly)
  --chart            Print bar rows under the table (daily, weekly, monthly, projects, models)
  --dense            High-density 9-box FinOps grid view (stats)
  --start-of-week    Week start day, sunday-saturday (default: sunday)
  --offline          Use cached pricing only, never fetch
  --no-estimate      Recorded cost only, no pricing estimates (leaderboard and reports)
  --license          Show license and commercial terms
  --json             Output JSON instead of dashboard
  --no-color         Disable ANSI colors
  --verbose          Enable verbose diagnostic warnings on stderr
  -h, --help         Show this help`
}

func parseArgs() thermal.Options {
	if len(os.Args) == 2 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" || arg == "version" {
			fmt.Printf("thermal %s\n", version.String())
			if version.Commit != "unknown" {
				fmt.Printf("  commit: %s\n", version.Commit)
			}
			if version.Date != "unknown" {
				fmt.Printf("  built:  %s\n", version.Date)
			}
			os.Exit(0)
		}
		if arg == "--license" || arg == "license" {
			if arg == "--license" {
				emitDeprecationWarning("--license", "thermal license", "0.14.0", "deprecated-license-flag", os.Getenv("NO_COLOR") != "")
			}
			runLicense(thermal.Options{})
			os.Exit(0)
		}
	}

	var opts thermal.Options
	flag.StringVar(&opts.Tool, "tool", "all", "Tool: all, mimocode, opencode, codex, devin, agy, command-code, codewhale, zcode, grok, muse, claude, droid, dsh, hermes")
	flag.StringVar(&opts.DBPath, "db", "", "Override database/data path")
	flag.IntVar(&opts.Weeks, "weeks", 52, "Heatmap width in weeks (4-104)")
	flag.StringVar(&opts.Since, "since", "", "Report window start (YYYY-MM-DD or YYYYMMDD)")
	flag.StringVar(&opts.Until, "until", "", "Report window end (YYYY-MM-DD or YYYYMMDD)")
	flag.IntVar(&opts.Last, "last", 0, "Last N days/weeks/months for reports")
	flag.StringVar(&opts.Order, "order", "desc", "Report sort order: asc or desc")
	flag.StringVar(&opts.Sort, "sort", "", "Ranking: streak|tokens|cost for the leaderboard, tokens|cost|days|recent for projects, tokens|cost for models")
	flag.IntVar(&opts.Top, "top", 0, "Project or model rows to print, 0 for all")
	flag.StringVar(&opts.Metric, "metric", "tokens", "Analytics metric: tokens or cost")
	flag.StringVar(&opts.By, "by", "tool", "Mix dimension: tool or model")
	flag.StringVar(&opts.Grain, "grain", "week", "Mix bucket size: day, week, or month")
	flag.StringVar(&opts.Against, "against", "", "Target model for replay simulation")
	flag.StringVar(&opts.Compare, "compare", "", "Plans to compare in replay: comma-separated or all")
	flag.BoolVar(&opts.Breakdown, "breakdown", false, "Show per-model rows in reports")
	flag.BoolVar(&opts.Chart, "chart", false, "Print bar rows under report tables")
	flag.BoolVar(&opts.Dense, "dense", false, "High-density 9-box FinOps grid view (stats)")
	flag.StringVar(&opts.StartOfWeek, "start-of-week", "sunday", "Week start day: sunday-saturday")
	flag.BoolVar(&opts.Offline, "offline", false, "Use cached pricing only, never fetch")
	flag.BoolVar(&opts.NoEstimate, "no-estimate", false, "Recorded cost only, no pricing estimates")
	flag.BoolVar(&opts.NoUpdateCheck, "no-update-check", false, "Disable daily automatic update check")
	var showLicense bool
	flag.BoolVar(&showLicense, "license", false, "Show license, dual-licensing & commercial terms")
	flag.BoolVar(&opts.JSON, "json", false, "Output JSON instead of dashboard")
	flag.BoolVar(&opts.NoColor, "no-color", false, "Disable ANSI colors")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose warning diagnostics on stderr")
	flag.BoolVar(&opts.Verbose, "v", false, "Enable verbose warning diagnostics on stderr (shorthand)")
	flag.BoolVar(&opts.Verbose, "d", false, "Enable verbose warning diagnostics on stderr (shorthand)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage())
	}
	flag.Parse()
	if showLicense {
		opts.Tool = "license"
	}

	remaining := flag.Args()
	var positionals []string
	for i := 0; i < len(remaining); i++ {
		arg := remaining[i]
		if !strings.HasPrefix(arg, "-") {
			positionals = append(positionals, arg)
			continue
		}

		name, inline := arg, ""
		hasInline := false
		if eq := strings.Index(arg, "="); eq >= 0 {
			name = arg[:eq]
			inline = arg[eq+1:]
			hasInline = true
		}
		takeValue := func() (string, bool) {
			if hasInline {
				return inline, true
			}
			if i+1 < len(remaining) {
				next := remaining[i+1]
				// Accept negative numbers so --last -1 reaches validation
				// instead of being silently ignored.
				if !strings.HasPrefix(next, "-") || isNegativeNumber(next) {
					i++
					return next, true
				}
			}
			return "", false
		}

		switch name {
		case "--no-color":
			opts.NoColor = true
		case "--json":
			opts.JSON = true
		case "--license":
			opts.Tool = "license"
		case "--breakdown":
			opts.Breakdown = true
		case "--chart":
			opts.Chart = true
		case "--dense":
			opts.Dense = true
		case "--offline":
			opts.Offline = true
		case "--no-estimate":
			opts.NoEstimate = true
		case "--verbose", "-v", "-d":
			opts.Verbose = true
		case "--weeks":
			if v, ok := takeValue(); ok {
				if n, err := strconv.Atoi(v); err == nil {
					opts.Weeks = n
				}
			}
		case "--db":
			if v, ok := takeValue(); ok {
				opts.DBPath = v
			}
		case "--tool":
			if v, ok := takeValue(); ok {
				opts.Tool = v
			}
		case "--since":
			if v, ok := takeValue(); ok {
				opts.Since = v
			}
		case "--until":
			if v, ok := takeValue(); ok {
				opts.Until = v
			}
		case "--last":
			if v, ok := takeValue(); ok {
				if n, err := strconv.Atoi(v); err == nil {
					opts.Last = n
				}
			}
		case "--order":
			if v, ok := takeValue(); ok {
				opts.Order = v
			}
		case "--sort":
			if v, ok := takeValue(); ok {
				opts.Sort = v
			}
		case "--top":
			if v, ok := takeValue(); ok {
				if n, err := strconv.Atoi(v); err == nil {
					opts.Top = n
				}
			}
		case "--metric":
			if v, ok := takeValue(); ok {
				opts.Metric = v
			}
		case "--by":
			if v, ok := takeValue(); ok {
				opts.By = v
			}
		case "--grain":
			if v, ok := takeValue(); ok {
				opts.Grain = v
			}
		case "--start-of-week":
			if v, ok := takeValue(); ok {
				opts.StartOfWeek = v
			}
		case "--against":
			if v, ok := takeValue(); ok {
				opts.Against = v
			}
		case "--compare":
			if v, ok := takeValue(); ok {
				opts.Compare = v
			}
		}
	}

	if len(positionals) > 0 {
		if isReportWord(positionals[0]) {
			opts.Report = strings.ToLower(positionals[0])
			if len(positionals) > 1 {
				opts.Tool = positionals[1]
			}
		} else {
			opts.Tool = positionals[0]
			if len(positionals) > 1 && isReportWord(positionals[1]) {
				opts.Report = strings.ToLower(positionals[1])
			}
		}
	}

	if err := validateReportFlags(opts); err != nil {
		fmt.Fprintf(os.Stderr, "thermal: %v\n", err)
		os.Exit(1)
	}

	if opts.Weeks < 4 || opts.Weeks > 104 {
		fmt.Fprintf(os.Stderr, "thermal: --weeks must be between 4 and 104\n")
		os.Exit(1)
	}

	return opts
}

func isReportWord(s string) bool {
	switch strings.ToLower(s) {
	case "daily", "weekly", "monthly", "projects", "models", "trend", "mix", "stats", "replay", "yield", "receipt":
		return true
	}
	return false
}

// isNegativeNumber reports whether s is a negative integer. Numeric flag
// values like --last -1 must not be mistaken for the next flag.
func isNegativeNumber(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for _, ch := range s[1:] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// validateReportFlags rejects report options that would otherwise be silently
// ignored, and checks the values themselves. Sort defaults depend on the
// command: the leaderboard ranks by streak, projects and models by tokens.
func validateReportFlags(opts thermal.Options) error {
	sortKey := strings.ToLower(opts.Sort)
	metricKey := strings.ToLower(opts.Metric)
	if metricKey == "" {
		metricKey = "tokens"
	}
	byKey := strings.ToLower(opts.By)
	if byKey == "" {
		byKey = "tool"
	}
	grainKey := strings.ToLower(opts.Grain)
	if grainKey == "" {
		grainKey = "week"
	}

	if opts.Last < 0 {
		return fmt.Errorf("--last cannot be negative")
	}
	if opts.Top < 0 {
		return fmt.Errorf("--top cannot be negative")
	}
	if opts.Last > 0 && (opts.Since != "" || opts.Until != "") {
		return fmt.Errorf("--last cannot be combined with --since or --until")
	}
	if opts.Order != "asc" && opts.Order != "desc" {
		return fmt.Errorf("--order must be asc or desc")
	}
	if _, ok := thermal.ParseWeekday(opts.StartOfWeek); !ok {
		return fmt.Errorf("--start-of-week must be a weekday name, sunday through saturday")
	}
	if opts.Since != "" {
		if _, ok := thermal.ParseDay(opts.Since); !ok {
			return fmt.Errorf("--since must be YYYY-MM-DD or YYYYMMDD")
		}
	}
	if opts.Until != "" {
		if _, ok := thermal.ParseDay(opts.Until); !ok {
			return fmt.Errorf("--until must be YYYY-MM-DD or YYYYMMDD")
		}
	}

	if opts.Dense && opts.Report != "stats" {
		return fmt.Errorf("--dense only applies to the stats command")
	}

	if opts.Report == "" {
		if opts.Tool == "audit" || opts.Tool == "share" {
			if opts.Chart || opts.Breakdown || opts.Since != "" || opts.Until != "" || opts.Last != 0 || opts.Top != 0 {
				return fmt.Errorf("report options do not apply to the %s command", opts.Tool)
			}
			if opts.Against != "" || opts.Compare != "" {
				return fmt.Errorf("--against and --compare only apply to the replay command")
			}
			return nil
		}
		// Leaderboard flags. Sort is optional here and means streak when unset.
		if sortKey != "" && sortKey != "streak" && sortKey != "tokens" && sortKey != "cost" {
			return fmt.Errorf("--sort must be streak, tokens, or cost for the leaderboard")
		}
		if opts.Top != 0 {
			return fmt.Errorf("--top only applies to the projects and models commands")
		}
		if metricKey != "tokens" || byKey != "tool" || grainKey != "week" {
			return fmt.Errorf("--metric, --by, and --grain only apply to the trend, mix, and stats commands")
		}
		// --no-estimate and --offline also apply to the leaderboard, which is
		// the one place a reader compares recorded cost against reports.
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		if opts.Since != "" || opts.Until != "" || opts.Last != 0 ||
			opts.Breakdown || opts.Chart ||
			opts.Order != "desc" || opts.StartOfWeek != "sunday" {
			return fmt.Errorf("report options (--since, --until, --last, --breakdown, --chart, --order, --start-of-week) need a report command: daily, weekly, monthly, projects, or models")
		}
		return nil
	}

	switch opts.Report {
	case "replay":
		if sortKey != "" {
			return fmt.Errorf("--sort does not apply to the replay command")
		}
		if opts.Top != 0 {
			return fmt.Errorf("--top only applies to the projects and models commands")
		}
		if metricKey != "tokens" || byKey != "tool" || grainKey != "week" {
			return fmt.Errorf("--metric, --by, and --grain only apply to the trend, mix, and stats commands")
		}
		if opts.Breakdown || opts.Chart {
			return fmt.Errorf("--breakdown and --chart do not apply to the replay command")
		}
	case "yield":
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		if opts.Breakdown || opts.Chart {
			return fmt.Errorf("--breakdown and --chart do not apply to the yield command")
		}
		if metricKey != "tokens" || byKey != "tool" || grainKey != "week" {
			return fmt.Errorf("--metric, --by, and --grain only apply to the trend, mix, and stats commands")
		}
		switch sortKey {
		case "", "tokens", "lines", "yield":
		default:
			return fmt.Errorf("--sort must be tokens, lines, or yield for the yield command")
		}
	case "receipt":
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		if opts.Breakdown || opts.Chart {
			return fmt.Errorf("--breakdown and --chart do not apply to the receipt command")
		}
		if metricKey != "tokens" || byKey != "tool" || grainKey != "week" {
			return fmt.Errorf("--metric, --by, and --grain only apply to the trend, mix, and stats commands")
		}
		switch sortKey {
		case "", "tokens", "verified", "cost", "rate":
		default:
			return fmt.Errorf("--sort must be tokens, verified, cost, or rate for the receipt command")
		}
	case "projects":
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		switch sortKey {
		case "", "tokens", "cost", "days", "recent":
		default:
			return fmt.Errorf("--sort must be tokens, cost, days, or recent for projects")
		}
	case "models":
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		switch sortKey {
		case "", "tokens", "cost":
		default:
			return fmt.Errorf("--sort must be tokens or cost for models")
		}
	case "trend", "mix", "stats":
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		if sortKey != "" {
			return fmt.Errorf("--sort does not apply to the analytics commands")
		}
		if opts.Top != 0 {
			return fmt.Errorf("--top only applies to the projects and models commands")
		}
		switch metricKey {
		case "tokens", "cost":
		default:
			return fmt.Errorf("--metric must be tokens or cost")
		}
		if opts.Report == "mix" {
			switch byKey {
			case "tool", "model":
			default:
				return fmt.Errorf("--by must be tool or model")
			}
			switch grainKey {
			case "day", "week", "month":
			default:
				return fmt.Errorf("--grain must be day, week, or month")
			}
			break
		}
		if byKey != "tool" {
			return fmt.Errorf("--by only applies to the mix command")
		}
		if grainKey != "week" {
			return fmt.Errorf("--grain only applies to the mix command")
		}
	default:
		if opts.Against != "" || opts.Compare != "" {
			return fmt.Errorf("--against and --compare only apply to the replay command")
		}
		if sortKey != "" {
			return fmt.Errorf("--sort only applies to the leaderboard, projects, and models")
		}
		if opts.Top != 0 {
			return fmt.Errorf("--top only applies to the projects and models commands")
		}
		if metricKey != "tokens" || byKey != "tool" || grainKey != "week" {
			return fmt.Errorf("--metric, --by, and --grain only apply to the trend, mix, and stats commands")
		}
	}
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--check-update-bg" {
		runBackgroundUpdateCheck()
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	_ = ctx

	opts := parseArgs()
	defer maybeCheckForUpdate(opts)

	if usedLicenseFlag() {
		emitDeprecationWarning("--license", "thermal license", "0.14.0", "deprecated-license-flag", opts.NoColor)
	}
	if opts.Tool == "nous" || opts.Tool == "nous-hermes" {
		emitDeprecationWarning(opts.Tool, "hermes", "0.14.0", "deprecated-tool-nous", opts.NoColor)
	}

	// Subcommands that don't go through the normal tool flow.
	switch opts.Tool {
	case "upgrade":
		os.Exit(runUpgrade())
	case "dashboard":
		os.Exit(runDashboard(opts))
	case "yield":
		opts.Tool = "all"
		opts.Report = "yield"
		runYieldReport(opts)
		return
	case "audit":
		runAudit(opts)
		return
	case "share":
		runShare(opts)
		return
	case "license", "--license":
		runLicense(opts)
		return
	case "version", "--version", "-v":
		fmt.Printf("thermal %s\n", version.String())
		if version.Commit != "unknown" {
			fmt.Printf("  commit: %s\n", version.Commit)
		}
		if version.Date != "unknown" {
			fmt.Printf("  built:  %s\n", version.Date)
		}
		return
	}

	if opts.Report != "" {
		switch opts.Report {
		case "projects":
			runProjectReport(opts)
		case "models":
			runModelReport(opts)
		case "mix":
			runMixReport(opts)
		case "stats":
			runStatsReport(opts)
		case "trend":
			runTrendReport(opts)
		case "replay":
			runReplayReport(opts)
		case "yield":
			runYieldReport(opts)
		case "receipt":
			runReceiptReport(opts)
		case "audit":
			runAudit(opts)
		case "share":
			runShare(opts)
		default:
			runReport(opts)
		}
		return
	}

	if opts.Tool == "all" {
		tools := loaders.AllTools()
		var results []thermal.ToolResult
		pricer := newPricer(opts)

		for _, t := range allToolOrder {
			info := tools[t]
			if !toolHasData(info) {
				continue
			}

			data, err := loaders.LoadToolData(t, info, "")
			if err != nil {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "thermal: warning: failed loading %s: %v\n", info.Name, err)
				}
				continue
			}
			printToolWarnings(info.Name, data.Summary.Warnings, opts.Verbose)

			var estCost float64
			if pricer != nil && data.Summary.Cost == 0 {
				for _, d := range data.Daily {
					if len(d.Models) > 0 {
						c, _ := pricer.PriceDay(d)
						estCost += c
					}
				}
			}

			activeDays := make(map[string]bool)
			for _, d := range data.Daily {
				if d.Turns > 0 {
					activeDays[d.Day] = true
				}
			}
			current, longest := thermal.ComputeStreaks(activeDays)

			results = append(results, thermal.ToolResult{
				Tool:          t,
				Name:          info.Name,
				Summary:       data.Summary,
				Daily:         data.Daily,
				CurrentStreak: current,
				LongestStreak: longest,
				ActiveDays:    len(activeDays),
				TotalActivity: data.Summary.LifetimeTokens,
				DataPath:      data.Path,
				EstimatedCost: estCost,
			})
		}

		if len(results) == 0 {
			fmt.Fprintf(os.Stderr, "thermal: no supported tool data found\n")
			os.Exit(1)
		}

		if opts.JSON {
			type jsonResult struct {
				thermal.ToolResult
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

		fmt.Print(render.RenderLeaderboard(results, opts.Weeks, opts.NoColor, opts.Sort, !opts.NoEstimate))
		return
	}

	var tool thermal.Tool
	if opts.Tool == "auto" {
		tool = loaders.DetectTool(opts.Tool)
	} else {
		var ok bool
		tool, ok = loaders.ResolveTool(opts.Tool)
		if !ok {
			fmt.Fprintf(os.Stderr, "thermal: unknown tool: %s\n", opts.Tool)
			os.Exit(1)
		}
	}
	tools := loaders.AllTools()
	info := tools[tool]

	data, err := loaders.LoadToolData(tool, info, opts.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "thermal: %v\n", err)
		os.Exit(1)
	}
	printToolWarnings(info.Name, data.Summary.Warnings, opts.Verbose)

	activeDays := make(map[string]bool)
	for _, d := range data.Daily {
		if d.Turns > 0 {
			activeDays[d.Day] = true
		}
	}
	current, longest := thermal.ComputeStreaks(activeDays)

	if opts.JSON {
		type jsonSummary struct {
			thermal.Summary
			CurrentStreak int `json:"currentStreak"`
			LongestStreak int `json:"longestStreak"`
		}
		out := map[string]interface{}{
			"tool":        info.Name,
			"dataPath":    data.Path,
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
			"summary":     jsonSummary{Summary: data.Summary, CurrentStreak: current, LongestStreak: longest},
			"daily":       data.Daily,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(out)
		return
	}

	var estCost float64
	pricer := newPricer(opts)
	if pricer != nil && data.Summary.Cost == 0 {
		for _, d := range data.Daily {
			if len(d.Models) > 0 {
				c, _ := pricer.PriceDay(d)
				estCost += c
			}
		}
	}

	fmt.Print(render.RenderDashboard(info.Name, data.Summary, data.Daily, data.Path, opts.Weeks, opts.NoColor, estCost))
}

// allToolOrder is the stable display order shared by the leaderboard and the
// all-tools reports.
var allToolOrder = []thermal.Tool{
	thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolDevin,
	thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale, thermal.ToolZCode,
	thermal.ToolGrok, thermal.ToolMuse, thermal.ToolClaude, thermal.ToolDroid,
	thermal.ToolDsh, thermal.ToolHermes,
}

// toolHasData reports whether the tool's database or data directory exists.
func toolHasData(info loaders.ToolInfo) bool {
	if info.DBPath != "" {
		_, err := os.Stat(info.DBPath)
		return err == nil
	}
	if info.DataDir != "" {
		_, err := os.Stat(info.DataDir)
		return err == nil
	}
	return false
}

// usageSet is the loaded data behind the report, project, and model commands.
type usageSet struct {
	days     []thermal.DailyRow
	projects []thermal.ProjectDay
	byTool   []thermal.ToolDays
	results  []thermal.ToolResult
	toolName string
}

// loadUsage loads one tool or every tool with data. All-tools mode skips
// tools that fail and reports them on stderr when verbose; a named tool exits
// on failure.
func loadUsage(opts thermal.Options) usageSet {
	var set usageSet

	if opts.Tool == "all" || opts.Tool == "auto" {
		tools := loaders.AllTools()
		for _, t := range allToolOrder {
			info := tools[t]
			if !toolHasData(info) {
				continue
			}
			data, err := loaders.LoadToolData(t, info, "")
			if err != nil {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "thermal: warning: failed loading %s: %v\n", info.Name, err)
				}
				continue
			}
			printToolWarnings(info.Name, data.Summary.Warnings, opts.Verbose)
			set.days = append(set.days, data.Daily...)
			for i := range data.Projects {
				data.Projects[i].Tool = info.Name
			}
			set.projects = append(set.projects, data.Projects...)
			set.byTool = append(set.byTool, thermal.ToolDays{Tool: info.Name, Days: data.Daily})
			set.results = append(set.results, thermal.ToolResult{
				Tool:    t,
				Name:    info.Name,
				Summary: data.Summary,
				Daily:   data.Daily,
			})
		}
		if len(set.days) == 0 && len(set.projects) == 0 {
			fmt.Fprintf(os.Stderr, "thermal: no supported tool data found\n")
			os.Exit(1)
		}
		return set
	}

	tool, ok := loaders.ResolveTool(opts.Tool)
	if !ok {
		fmt.Fprintf(os.Stderr, "thermal: unknown tool: %s\n", opts.Tool)
		os.Exit(1)
	}
	info := loaders.AllTools()[tool]
	data, err := loaders.LoadToolData(tool, info, opts.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "thermal: %v\n", err)
		os.Exit(1)
	}
	printToolWarnings(info.Name, data.Summary.Warnings, opts.Verbose)
	set.days = data.Daily
	set.projects = data.Projects
	for i := range set.projects {
		set.projects[i].Tool = info.Name
	}
	set.byTool = []thermal.ToolDays{{Tool: info.Name, Days: data.Daily}}
	set.toolName = info.Name
	set.results = []thermal.ToolResult{{
		Tool:    tool,
		Name:    info.Name,
		Summary: data.Summary,
		Daily:   data.Daily,
	}}
	return set
}

func printToolWarnings(toolName string, warnings []string, verbose bool) {
	printToolWarningsTo(os.Stderr, toolName, warnings, verbose)
}

func printToolWarningsTo(w io.Writer, toolName string, warnings []string, verbose bool) {
	if !verbose {
		return
	}
	for _, wStr := range warnings {
		fmt.Fprintf(w, "thermal: warning: %s: %s\n", toolName, wStr)
	}
}

func usedLicenseFlag() bool {
	for _, a := range os.Args[1:] {
		if a == "--license" || a == "-license" || strings.HasPrefix(a, "--license=") || strings.HasPrefix(a, "-license=") {
			return true
		}
	}
	return false
}

func emitDeprecationWarning(item, replacement, removalVersion, anchor string, noColor bool) {
	emitDeprecationWarningTo(os.Stderr, item, replacement, removalVersion, anchor, noColor)
}

func emitDeprecationWarningTo(w io.Writer, item, replacement, removalVersion, anchor string, noColor bool) {
	colorEnabled := !noColor && os.Getenv("NO_COLOR") == ""
	if f, ok := w.(*os.File); ok {
		if !isatty.IsTerminal(f.Fd()) && !isatty.IsCygwinTerminal(f.Fd()) {
			colorEnabled = false
		}
	}

	msg := fmt.Sprintf("%s is deprecated and will be removed in v%s; use %s (see docs/MIGRATION.md#%s)",
		item, removalVersion, replacement, anchor)

	if colorEnabled {
		fmt.Fprintf(w, "\033[33mthermal: warning:\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(w, "thermal: warning: %s\n", msg)
	}
}

// newPricer builds the cost estimator unless estimation is disabled. It never
// touches the network when offline, and a missing catalog leaves stored cost
// untouched.
func newPricer(opts thermal.Options) thermal.Pricer {
	if opts.NoEstimate {
		return nil
	}
	cat := pricing.Load(pricing.DefaultCachePath(), opts.Offline)
	if cat.Len() > 0 {
		return cat
	}
	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "thermal: warning: no pricing data available; showing stored cost only")
	}
	return nil
}

// startOfWeek resolves the report week start from the flag.
func startOfWeek(opts thermal.Options) time.Weekday {
	if d, ok := thermal.ParseWeekday(opts.StartOfWeek); ok {
		return d
	}
	return time.Sunday
}

// writeReportJSON prints any report payload with the standard indentation.
func writeReportJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// runReport folds the loaded days into the requested grain and prints a table
// or JSON.
func runReport(opts thermal.Options) {
	aggOpts := thermal.AggregateOptions{
		StartOfWeek: startOfWeek(opts),
		Since:       opts.Since,
		Until:       opts.Until,
		Last:        opts.Last,
		Order:       opts.Order,
	}

	set := loadUsage(opts)

	rep := thermal.Aggregate(set.days, thermal.Grain(opts.Report), aggOpts, newPricer(opts))
	rep.Tool = set.toolName

	if opts.JSON {
		type jsonReport struct {
			thermal.Report
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReport{
			Report:      rep,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if opts.Breakdown {
		fmt.Print(render.RenderReportBreakdown(rep, opts.NoColor))
		if opts.Chart {
			fmt.Print(render.RenderPeriodChart(rep, opts.Sort, reportWidth(), opts.NoColor))
		}
		return
	}
	fmt.Print(render.RenderReport(rep, opts.NoColor))
	if opts.Chart {
		fmt.Print(render.RenderPeriodChart(rep, "tokens", reportWidth(), opts.NoColor))
	}
}

// projectDisplayNames resolves the labels the projects table used, so charts
// and tables name a project the same way.
func projectDisplayNames(rep thermal.ProjectReport) map[string]string {
	paths := make([]string, 0, len(rep.Rows))
	for _, row := range rep.Rows {
		paths = append(paths, row.Project)
	}
	return thermal.ProjectDisplayNames(paths)
}

// reportWidth is the width the terminal gives a table. Charts size themselves
// against it so a narrow terminal gets shorter bars rather than wrapped lines.
func reportWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if w, _, err := term.GetSize(int(os.Stdin.Fd())); err == nil && w > 0 {
		return w
	}
	return 100
}

// runDashboard launches the interactive dashboard. Two guards keep scripts and
// pipes safe: a non-TTY stdout is told to use the static commands, and --json
// never opens a terminal UI.
func runDashboard(opts thermal.Options) int {
	if opts.JSON {
		fmt.Fprintln(os.Stderr, "thermal: --json does not apply to the dashboard; use a report command, for example: thermal weekly --json")
		return 1
	}
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		fmt.Println("thermal dashboard is interactive and needs a terminal.")
		fmt.Println("For scripts and pipes use the static commands:")
		fmt.Println()
		fmt.Println("  thermal                  leaderboard across every tool")
		fmt.Println("  thermal weekly           weekly tokens and cost")
		fmt.Println("  thermal projects         tokens and cost per repository")
		fmt.Println("  thermal models           tokens and estimated cost per model")
		fmt.Println("  thermal weekly --json    the same numbers, machine readable")
		return 0
	}

	adapter := tui.LoadTools(newPricer(opts))
	if note := tui.LoadNote(adapter); note != "" && adapter.Latest == "" {
		fmt.Fprintf(os.Stderr, "thermal: %s\n", note)
		return 1
	}

	colorful := !opts.NoColor && os.Getenv("NO_COLOR") == ""
	p := tea.NewProgram(tui.New(adapter, colorful))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "thermal: dashboard failed: %v\n", err)
		return 1
	}
	return 0
}

// runProjectReport ranks projects by token usage across tools and time.
func runProjectReport(opts thermal.Options) {
	set := loadUsage(opts)

	projectOpts := thermal.ProjectOptions{
		Since: opts.Since,
		Until: opts.Until,
		Last:  opts.Last,
		Sort:  opts.Sort,
		Order: opts.Order,
	}
	rep := thermal.AggregateProjects(set.projects, projectOpts, newPricer(opts))

	if opts.JSON {
		type jsonReport struct {
			thermal.ProjectReport
			GeneratedAt string `json:"generatedAt"`
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(jsonReport{
			ProjectReport: rep,
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if opts.Breakdown {
		fmt.Print(render.RenderProjectsBreakdown(rep, opts.Top, opts.NoColor))
		if opts.Chart {
			fmt.Print(render.RenderProjectChart(rep, projectDisplayNames(rep), reportWidth(), opts.NoColor))
		}
		return
	}
	fmt.Print(render.RenderProjects(rep, opts.Top, opts.NoColor))
	if opts.Chart {
		fmt.Print(render.RenderProjectChart(rep, projectDisplayNames(rep), reportWidth(), opts.NoColor))
	}
}

// runModelReport ranks models across tools and time. Cost here is always an
// estimate because recorded cost is not attributable to a single model.
func runModelReport(opts thermal.Options) {
	set := loadUsage(opts)

	modelOpts := thermal.ModelOptions{
		Since: opts.Since,
		Until: opts.Until,
		Last:  opts.Last,
		Sort:  opts.Sort,
		Order: opts.Order,
	}
	rep := thermal.AggregateModels(set.byTool, modelOpts, newPricer(opts))

	if opts.JSON {
		type jsonReport struct {
			thermal.ModelReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReport{
			ModelReport: rep,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderModels(rep, opts.Top, opts.NoColor))
	if opts.Chart {
		fmt.Print(render.RenderModelChart(rep, reportWidth(), opts.NoColor))
	}
}

// runMixReport shows the tool or model mix over time with switching stats.
func runMixReport(opts thermal.Options) {
	set := loadUsage(opts)

	mixOpts := thermal.MixOptions{
		Since:       opts.Since,
		Until:       opts.Until,
		Last:        opts.Last,
		Grain:       thermal.Grain(strings.ToLower(opts.Grain)),
		By:          strings.ToLower(opts.By),
		Metric:      strings.ToLower(opts.Metric),
		StartOfWeek: startOfWeek(opts),
	}
	pricer := newPricer(opts)
	var rep thermal.MixReport
	if mixOpts.By == "model" {
		rep = thermal.AggregateModelMix(set.days, mixOpts, pricer)
	} else {
		rep = thermal.AggregateToolMix(set.byTool, mixOpts, pricer)
	}

	if opts.JSON {
		type jsonReport struct {
			thermal.MixReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReport{
			MixReport:   rep,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderMix(rep, opts.NoColor))
}

// runStatsReport summarises the daily distribution of tokens or cost,
// or renders the high-density 9-box FinOps grid if --dense is requested.
func runStatsReport(opts thermal.Options) {
	set := loadUsage(opts)
	pricer := newPricer(opts)

	if opts.Dense {
		yieldOpts := thermal.YieldOptions{
			Since: opts.Since,
			Until: opts.Until,
			Last:  opts.Last,
		}
		yieldRep := thermal.AggregateYield(set.results, set.projects, yieldOpts)
		grid := thermal.ComputeFinOpsGrid(set.days, set.results, set.projects, yieldRep, pricer)

		if opts.JSON {
			type jsonDenseReport struct {
				thermal.FinOpsGridPayload
				GeneratedAt string `json:"generatedAt"`
			}
			writeReportJSON(jsonDenseReport{
				FinOpsGridPayload: grid,
				GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		colorful := !opts.NoColor && render.IsTerminal() && os.Getenv("NO_COLOR") == ""
		if render.IsTerminal() && !opts.NoColor {
			p := tea.NewProgram(tui.NewDense(grid, colorful))
			if _, err := p.Run(); err == nil {
				return
			}
		}
		fmt.Print(tui.RenderDenseFinOps(grid, reportWidth(), colorful))
		return
	}

	statsOpts := thermal.StatsOptions{
		Since:  opts.Since,
		Until:  opts.Until,
		Last:   opts.Last,
		Metric: strings.ToLower(opts.Metric),
	}
	rep := thermal.AggregateStats(set.days, statsOpts, pricer)

	if opts.JSON {
		type jsonReport struct {
			thermal.StatsReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReport{
			StatsReport: rep,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderStats(rep, opts.NoColor))
}

// runTrendReport fits a daily trend and projects it to month end.
func runTrendReport(opts thermal.Options) {
	set := loadUsage(opts)

	trendOpts := thermal.TrendOptions{
		Since:  opts.Since,
		Until:  opts.Until,
		Last:   opts.Last,
		Metric: strings.ToLower(opts.Metric),
	}
	rep := thermal.AggregateTrend(set.days, trendOpts, newPricer(opts))

	if opts.JSON {
		type jsonReport struct {
			thermal.TrendReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReport{
			TrendReport: rep,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderTrend(rep, opts.NoColor))
}

// newReplayPricer returns a ReplayPricer unless estimation is disabled.
func newReplayPricer(opts thermal.Options) thermal.ReplayPricer {
	if opts.NoEstimate {
		return nil
	}
	cat := pricing.Load(pricing.DefaultCachePath(), opts.Offline)
	if cat.Len() > 0 {
		return cat
	}
	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "thermal: warning: no pricing data available; replay cannot estimate models")
	}
	return nil
}

// runReplayReport simulates the historical workload against subscriptions and API pricing.
func runReplayReport(opts thermal.Options) {
	set := loadUsage(opts)
	pricer := newReplayPricer(opts)

	var plans []thermal.SubscriptionPlan
	if opts.Compare != "" {
		if strings.ToLower(opts.Compare) == "all" {
			plans = pricing.StandardPlans()
		} else {
			for _, key := range strings.Split(opts.Compare, ",") {
				key = strings.TrimSpace(key)
				if key == "" {
					continue
				}
				p, ok := pricing.LookupPlan(key)
				if !ok {
					p = thermal.SubscriptionPlan{
						ID:           key,
						Name:         fmt.Sprintf("%s (API)", key),
						Type:         thermal.PlanTypePayAsYouGo,
						DefaultModel: key,
					}
				}
				plans = append(plans, p)
			}
		}
	} else if opts.Against == "" {
		for _, id := range pricing.DefaultComparisonPlans() {
			if p, ok := pricing.LookupPlan(id); ok {
				plans = append(plans, p)
			}
		}
	}

	replayOpts := thermal.ReplayOptions{
		Since:   opts.Since,
		Until:   opts.Until,
		Last:    opts.Last,
		Against: opts.Against,
		Compare: plans,
	}
	rep := thermal.AggregateReplay(set.days, replayOpts, pricer)

	if opts.JSON {
		type jsonReport struct {
			thermal.ReplayReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReport{
			ReplayReport: rep,
			GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderReplay(rep, opts.NoColor))
}

func runLicense(opts thermal.Options) {
	if opts.JSON {
		payload := map[string]any{
			"name":                   "thermal",
			"author":                 "Jad Madi",
			"license":                "AGPL-3.0-only",
			"spdx":                   "AGPL-3.0-only",
			"dual_license_available": true,
			"description":            "GNU Affero General Public License v3.0 (Community) / Commercial Enterprise License",
			"commercial_licensing": map[string]any{
				"available": true,
				"contact":   "contact@jadmadi.net",
				"github":    "https://github.com/jadmadi/thermal",
				"use_cases": []string{
					"proprietary_embedding",
					"closed_source_saas",
					"enterprise_ospo_compliance",
					"dedicated_support_sla",
				},
			},
			"covenant": map[string]any{
				"irrevocable_agpl":                 true,
				"non_relicensing_of_contributions": true,
				"no_artificial_feature_gating":     true,
				"description":                      "Thermal guarantees that the core CLI, agent loaders, TUI, and local analytics engine will remain available under OSI-approved copyleft (GNU AGPL-3.0) in perpetuity, without retroactive conversion of community contributions to proprietary terms.",
			},
			"third_party_notices": "NOTICES.md",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return
	}

	fmt.Println("Thermal — AI Agent Telemetry & Streak Engine")
	fmt.Println("Copyright (C) 2026 Jad Madi. All rights reserved.")
	fmt.Println()
	fmt.Println("Open Source & Community Edition (GNU AGPLv3):")
	fmt.Println("  Thermal is free and open-source software under the terms of the")
	fmt.Println("  GNU Affero General Public License v3.0 (AGPL-3.0).")
	fmt.Println("  Modifications distributed as binaries or served over a network")
	fmt.Println("  (hosted dashboards, team leaderboards, SaaS) require full source")
	fmt.Println("  code disclosure under AGPLv3 Section 13.")
	fmt.Println()
	fmt.Println("Community Reciprocity Covenant:")
	fmt.Println("  Thermal guarantees that the core CLI, loaders, TUI, and telemetry engine")
	fmt.Println("  will remain free software under GNU AGPL-3.0 in perpetuity. Community")
	fmt.Println("  contributions will never be retroactively relicensed into proprietary")
	fmt.Println("  or closed-source products without contributor consent.")
	fmt.Println()
	fmt.Println("Commercial & Enterprise Dual-Licensing:")
	fmt.Println("  A commercial license is available for organizations and vendors requiring:")
	fmt.Println("  • Embedding Thermal into proprietary software or IDE extensions")
	fmt.Println("  • Hosted multi-tenant cloud/SaaS leaderboards without source disclosure")
	fmt.Println("  • Compliance with corporate OSPO policies restricting AGPL")
	fmt.Println("  • Commercial support SLAs, custom agent loaders, and warranty")
	fmt.Println()
	fmt.Println("Commercial Inquiries:")
	fmt.Println("  Email:   contact@jadmadi.net")
	fmt.Println("  GitHub:  https://github.com/jadmadi/thermal")
	fmt.Println("  Terms:   DUAL-LICENSE.md")
	fmt.Println()
	fmt.Println("Third-Party Notices:")
	fmt.Println("  Incorporates permissive open-source components (MIT, BSD-3, Public Domain).")
	fmt.Println("  See NOTICES.md in the source repository for complete attributions.")
}

// runAudit executes the non-destructive local setup and context health diagnostic.
func runAudit(opts thermal.Options) {
	rep := audit.RunAudit("")
	if opts.JSON {
		fmt.Println(render.RenderAuditJSON(rep))
		return
	}
	fmt.Print(render.RenderAudit(rep, opts.NoColor))
}

// runShare generates a stateless, zero-database share URL for user streaks and telemetry.
func runShare(opts thermal.Options) {
	targetTool := "all"
	if opts.Report == "share" && opts.Tool != "" && opts.Tool != "share" {
		targetTool = opts.Tool
	} else if opts.Tool != "" && opts.Tool != "share" {
		targetTool = opts.Tool
	}
	opts.Tool = targetTool
	set := loadUsage(opts)
	toolLabel := targetTool
	pricer := newPricer(opts)
	var totalCost float64
	var estimated bool
	for _, d := range set.days {
		if d.Cost > 0 {
			totalCost += d.Cost
		} else if pricer != nil && len(d.Models) > 0 {
			c, _ := pricer.PriceDay(d)
			if c > 0 {
				totalCost += c
				estimated = true
			}
		}
	}

	snap := share.BuildShareSnapshot(toolLabel, set.days, totalCost, estimated)
	url, token, err := share.ShareURL(snap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "thermal: failed generating share token: %v\n", err)
		os.Exit(1)
	}

	if opts.JSON {
		res := map[string]any{
			"url":      url,
			"token":    token,
			"snapshot": snap,
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
		return
	}

	colors := !opts.NoColor && render.IsTerminal() && os.Getenv("NO_COLOR") == ""
	highlight := func(s string) string { return render.ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return render.ColorCode(colors, "38;5;239", s) }
	gold := func(s string) string { return render.ColorCode(colors, "1;33", s) }
	cyan := func(s string) string { return render.ColorCode(colors, "1;36", s) }

	fmt.Println()
	fmt.Printf("  %s %s %s\n\n", highlight("Thermal"), dim("·"), highlight("share · stateless streak card"))
	fmt.Printf("  %s\n", highlight("Share URL:"))
	fmt.Printf("  %s\n\n", cyan(url))
	fmt.Printf("  %s\n", highlight("Encoded Payload:"))
	fmt.Printf("  • Current Streak: %s days (Longest: %s days)\n", gold(fmt.Sprintf("%d", snap.CurrentStreak)), fmt.Sprintf("%d", snap.LongestStreak))
	fmt.Printf("  • Active Days:    %d days\n", snap.ActiveDays)
	fmt.Printf("  • Total Volume:   %s tokens\n", thermal.CompactNumber(snap.TotalTokens))
	if snap.TotalCost > 0 {
		costStr := fmt.Sprintf("$%.2f", snap.TotalCost)
		if snap.EstimatedCost {
			costStr = "~" + costStr
		}
		fmt.Printf("  • Spend Profile:  %s\n", costStr)
	}
	fmt.Printf("  • Privacy Notice: 100%% zero-database, client-only URL fragment. No file paths or prompt data encoded.\n\n")
}

// runYieldReport correlates token spend with code output delta telemetry.
func runYieldReport(opts thermal.Options) {
	if opts.Tool == "yield" {
		opts.Tool = "all"
	}
	yieldOpts := thermal.YieldOptions{
		Since: opts.Since,
		Until: opts.Until,
		Last:  opts.Last,
		Sort:  opts.Sort,
		Top:   opts.Top,
	}

	set := loadUsage(opts)
	rep := thermal.AggregateYield(set.results, set.projects, yieldOpts)
	rep.Tool = set.toolName

	if opts.JSON {
		type jsonYieldReport struct {
			thermal.YieldReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonYieldReport{
			YieldReport: rep,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderYield(rep, opts.Top, opts.NoColor))
}

// runReceiptReport links agent token spend to concrete verification outcomes.
func runReceiptReport(opts thermal.Options) {
	receiptOpts := thermal.ReceiptOptions{
		Since: opts.Since,
		Until: opts.Until,
		Last:  opts.Last,
		Sort:  opts.Sort,
		Top:   opts.Top,
	}

	targetTool := opts.Tool
	if targetTool == "receipt" {
		targetTool = "all"
	}
	opts.Tool = targetTool

	home := thermal.HomeDir()
	pricer := newPricer(opts)

	// Scan real session transcripts from local tool stores
	receipts := thermal.ScanSessionReceipts(home, targetTool, pricer)

	// If no transcript files found (e.g. blank environment or tools storing aggregates only),
	// synthesize receipts from loaded tool sessions
	if len(receipts) == 0 {
		set := loadUsage(opts)
		for _, res := range set.results {
			for _, day := range res.Daily {
				if day.Turns == 0 && day.Tokens == 0 {
					continue
				}
				receipts = append(receipts, thermal.WorkReceipt{
					SessionID: fmt.Sprintf("%s-%s", strings.ToLower(res.Name), day.Day),
					Tool:      res.Name,
					Day:       day.Day,
					Tokens:    day.Tokens,
					Cost:      day.Cost,
					Tier:      thermal.Tier2Claimed,
					Status:    "CLAIMED",
				})
			}
		}
	}

	rep := thermal.AggregateReceipts(receipts, receiptOpts)
	rep.Tool = targetTool

	if opts.JSON {
		type jsonReceiptReport struct {
			thermal.ReceiptReport
			GeneratedAt string `json:"generatedAt"`
		}
		writeReportJSON(jsonReceiptReport{
			ReceiptReport: rep,
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	fmt.Print(render.RenderReceipt(rep, opts.Top, opts.NoColor))
}

