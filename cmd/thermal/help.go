// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/render"
)

// toolHelpEntry is the help-facing identity of a supported tool.
type toolHelpEntry struct {
	Canonical string
	Display   string
	Alias     string
}

var toolHelpEntries = []toolHelpEntry{
	{"opencode", "OpenCode", "oc"},
	{"devin", "Devin", ""},
	{"codex", "Codex", ""},
	{"claude", "Claude", ""},
	{"zcode", "ZCode", "zc"},
	{"grok", "Grok", ""},
	{"mimocode", "MiMoCode", "mimo"},
	{"codewhale", "codewhale", "whale"},
	{"dsh", "DeepSeek (DSH)", "dsh"},
	{"hermes", "Nous Hermes", "hermes"},
	{"command-code", "CMD", "cmd"},
	{"zed", "Zed", ""},
	{"agy", "Agy", "agy"},
	{"muse", "Muse", "muse"},
	{"droid", "Droid", "droid"},
}

func findToolEntry(name string) (toolHelpEntry, bool) {
	t, ok := loaders.ResolveTool(name)
	if !ok {
		return toolHelpEntry{}, false
	}
	canonical := string(t)
	for _, e := range toolHelpEntries {
		if e.Canonical == canonical {
			return e, true
		}
	}
	return toolHelpEntry{}, false
}

func baseReportFlags() []render.HelpFlag {
	return []render.HelpFlag{
		{"--since <date>", "Start date (YYYY-MM-DD or YYYYMMDD)"},
		{"--until <date>", "End date (YYYY-MM-DD or YYYYMMDD)"},
		{"--last <N>", "Last N days, weeks, or months (N >= 1)"},
		{"--json", "Emit the JSON envelope"},
		{"--require-data", "Exit 1 when the report has no data"},
		{"--no-color", "Disable ANSI colors"},
		{"--tool <name>", "Scope to one tool (canonical name or alias)"},
		{"--db <path>", "Read one tool's data from a file or directory (requires --tool)"},
		{"--offline", "Use cached pricing only, never fetch"},
		{"--no-estimate", "Recorded cost only; disable estimates"},
		{"--format <text|json>", "Output format (json is an alias for --json)"},
	}
}

func withFlags(extra ...render.HelpFlag) []render.HelpFlag {
	return append(extra, baseReportFlags()...)
}

var commandHelps = map[string]render.CommandHelp{
	"daily": {
		Title:   "daily",
		Usage:   "thermal daily [tool] [options]",
		Summary: "Daily token volume and cost breakdown",
		Description: []string{
			"Answers: what did I use on each day, and what did it cost?",
			"Reads local session data read-only. Days with no classified tokens, cost, or model attribution are omitted.",
		},
		Flags: withFlags(
			render.HelpFlag{"--top <N>", "Cap period rows (0 = all)"},
			render.HelpFlag{"--breakdown", "Show per-model rows under each period"},
			render.HelpFlag{"--chart", "Print a bar chart under the table"},
		),
		Examples: []string{"thermal daily", "thermal daily --last 7", "thermal daily opencode --json"},
		JSON:     "keys: type, window, data, totals, costBasis, generatedAt",
		Related:  []string{"thermal weekly", "thermal stats"},
	},
	"weekly": {
		Title:   "weekly",
		Usage:   "thermal weekly [tool] [options]",
		Summary: "Weekly aggregated activity and spend",
		Description: []string{
			"Folds daily rows into calendar weeks. The week start follows --start-of-week.",
		},
		Flags: withFlags(
			render.HelpFlag{"--top <N>", "Cap period rows (0 = all)"},
			render.HelpFlag{"--start-of-week <day>", "First day of week (sunday-saturday, default sunday)"},
			render.HelpFlag{"--breakdown", "Show per-model rows under each period"},
			render.HelpFlag{"--chart", "Print a bar chart under the table"},
		),
		Examples: []string{"thermal weekly", "thermal weekly --last 4", "thermal weekly --start-of-week monday --json"},
		JSON:     "keys: type, window, data, totals, costBasis, generatedAt",
		Related:  []string{"thermal daily", "thermal monthly"},
	},
	"monthly": {
		Title:   "monthly",
		Usage:   "thermal monthly [tool] [options]",
		Summary: "Monthly summary and model breakdown",
		Description: []string{
			"Folds daily rows into calendar months, with optional per-model rows.",
		},
		Flags: withFlags(
			render.HelpFlag{"--top <N>", "Cap period rows (0 = all)"},
			render.HelpFlag{"--breakdown", "Show per-model rows under each period"},
			render.HelpFlag{"--chart", "Print a bar chart under the table"},
		),
		Examples: []string{"thermal monthly", "thermal monthly --last 3", "thermal monthly --breakdown --json"},
		JSON:     "keys: type, window, data, totals, costBasis, generatedAt",
		Related:  []string{"thermal weekly", "thermal models"},
	},
	"projects": {
		Title:   "projects",
		Usage:   "thermal projects [tool] [options]",
		Summary: "Tokens and cost per project, folded to the git root",
		Description: []string{
			"Groups sessions by project, resolving symlinks and folding subdirectories into the nearest git root.",
		},
		Flags: withFlags(
			render.HelpFlag{"--sort <field>", "Rank by tokens|cost|days|recent (default tokens)"},
			render.HelpFlag{"--top <N>", "Cap table rows (0 = all)"},
			render.HelpFlag{"--order <asc|desc>", "Sort order (default desc)"},
			render.HelpFlag{"--breakdown", "Show tool and top-model rows under each project"},
			render.HelpFlag{"--chart", "Print a bar chart under the table"},
		),
		Examples: []string{"thermal projects", "thermal projects --top 10", "thermal projects --sort cost --json"},
		JSON:     "keys: type, data, totals, costBasis, generatedAt",
		Related:  []string{"thermal models", "thermal daily"},
	},
	"models": {
		Title:   "models",
		Usage:   "thermal models [tool] [options]",
		Summary: "Tokens and estimated cost per model, ranked",
		Description: []string{
			"Ranks canonical model ids across tools. Recorded session cost is not attributable to one model, so cost is estimated from the pricing catalog.",
		},
		Flags: withFlags(
			render.HelpFlag{"--sort <field>", "Rank by tokens|cost (default tokens)"},
			render.HelpFlag{"--top <N>", "Cap table rows (0 = all)"},
			render.HelpFlag{"--order <asc|desc>", "Sort order (default desc)"},
			render.HelpFlag{"--chart", "Print a bar chart under the table"},
		),
		Examples: []string{"thermal models", "thermal models --top 10", "thermal models --sort cost --json"},
		JSON:     "keys: type, data, totals, costBasis, generatedAt",
		Related:  []string{"thermal projects", "thermal mix"},
	},
	"stats": {
		Title:   "stats",
		Usage:   "thermal stats [tool] [options]",
		Summary: "Daily distribution, percentiles, and token composition",
		Description: []string{
			"Shows active-day distribution, percentiles, weekday profile, token composition, and outliers. Mean, median, p90, and max use active days; trend uses calendar days.",
		},
		Flags: withFlags(
			render.HelpFlag{"--metric <field>", "Measure tokens or cost (default tokens)"},
		),
		Examples: []string{"thermal stats", "thermal stats --metric cost", "thermal stats --json"},
		JSON:     "keys: type, metric, activeDays, total, mean, median, p90, max, weekday, composition, histogram, topDays, generatedAt",
		Related:  []string{"thermal trend", "thermal mix"},
	},
	"mix": {
		Title:   "mix",
		Usage:   "thermal mix [tool] [options]",
		Summary: "Tool and model market share and switching frequency",
		Description: []string{
			"Shows share by tool or model over time, the dominant tool, switch count, and spread.",
		},
		Flags: withFlags(
			render.HelpFlag{"--by <dimension>", "Break down by tool|model|reasoning (default tool)"},
			render.HelpFlag{"--grain <grain>", "Bucket by day|week|month (default week)"},
			render.HelpFlag{"--metric <field>", "Measure tokens or cost (default tokens)"},
		),
		Examples: []string{"thermal mix", "thermal mix --by model --grain month", "thermal mix --json"},
		JSON:     "keys: type, metric, by, grain, rows, total, dominant, switches, spread, generatedAt",
		Related:  []string{"thermal models", "thermal stats"},
	},
	"trend": {
		Title:   "trend",
		Usage:   "thermal trend [tool] [options]",
		Summary: "Daily trend fit with a month-end projection",
		Description: []string{
			"Fits a least-squares line over calendar days, including zero-activity days, and projects to month end.",
		},
		Flags: withFlags(
			render.HelpFlag{"--metric <field>", "Measure tokens or cost (default tokens)"},
		),
		Examples: []string{"thermal trend", "thermal trend --last 30", "thermal trend --metric cost --json"},
		JSON:     "keys: type, metric, points, mean, slope, projection, generatedAt",
		Related:  []string{"thermal stats", "thermal daily"},
	},
	"live": {
		Title:   "live",
		Usage:   "thermal live [tool] [options]",
		Summary: "Real-time token burn monitor",
		Description: []string{
			"Interactive monitor with burn velocity, flame intensity, rolling activity, and a recent-burst ticker. Requires a terminal; use --json or --json --stream for scripts.",
		},
		Flags: []render.HelpFlag{
			{"--interval <duration>", "Polling interval (default 1s)"},
			{"--flame-ceiling <N>", "Flame ceiling in active units per minute (e.g. 9M or 60)"},
			{"--stream", "Stream continuous NDJSON events (alias: --follow, -f)"},
			{"--json", "Emit a single JSON snapshot"},
			{"--no-color", "Disable ANSI colors"},
			{"--fresh", "Start a new session baseline at launch"},
			{"--tool <name>", "Scope to one tool"},
			{"--db <path>", "Read one tool's data from a file or directory (requires --tool)"},
			{"--verbose", "Show loader diagnostics on stderr"},
		},
		Examples: []string{"thermal live", "thermal live --json", "thermal live --json --stream"},
		JSON:     "snapshot: burnTokensPerSec, burnTokensPerMin, flameIntensity, sessionTokens, todayTokens, todayProjects, todayModels, recentEvents",
		Related:  []string{"thermal statusline", "thermal stats"},
	},
	"statusline": {
		Title:   "statusline",
		Usage:   "thermal statusline [options]",
		Summary: "Instant shell statusline and prompt helper",
		Description: []string{
			"Prints one line for a shell prompt or multiplexer. Aliases: prompt, status. With no data it prints nothing and exits 0.",
		},
		Flags: []render.HelpFlag{
			{"--plain", "Plain text without emoji or glyphs"},
			{"--nerd", "Use the Nerd Font flame glyph"},
			{"--json", "Emit the statusline JSON object"},
			{"--no-color", "Disable ANSI colors"},
			{"--fresh", "Bypass the cached snapshot"},
			{"--tool <name>", "Scope to one tool"},
			{"--db <path>", "Read one tool's data from a file or directory (requires --tool)"},
		},
		Examples: []string{"thermal statusline", "thermal statusline --plain", "thermal statusline --json"},
		JSON:     "keys: streak, scope, longestStreak, todayTokens, todayTurns, activeTool, day, updatedAt",
		Related:  []string{"thermal live", "thermal share"},
	},
	"share": {
		Title:   "share",
		Usage:   "thermal share [options]",
		Summary: "Stateless, private URL card for your streak",
		Description: []string{
			"Encodes a client-verifiable card in the URL fragment. Zero database or cloud storage; no paths or prompts are included.",
		},
		Flags: []render.HelpFlag{
			{"--json", "Emit the full URL and snapshot"},
			{"--no-color", "Disable ANSI colors"},
			{"--tool <name>", "Scope to one tool"},
			{"--db <path>", "Read one tool's data from a file or directory (requires --tool)"},
			{"--offline", "Use cached pricing only"},
			{"--no-estimate", "Recorded cost only"},
		},
		Examples: []string{"thermal share", "thermal share --json"},
		JSON:     "keys: type, snapshot, url, generatedAt",
		Related:  []string{"thermal statusline", "thermal daily"},
	},
	"sources": {
		Title:   "sources",
		Usage:   "thermal sources [options]",
		Summary: "Stat-only inventory of supported data sources",
		Description: []string{
			"Lists each supported tool's expected file or directory and whether it is found, empty, missing, unreadable, or the wrong shape.",
			"Reads only path metadata: no database, transcript, or prompt is opened. This command always exits 0 when it can produce the inventory.",
		},
		Flags: []render.HelpFlag{
			{"--json", "Emit the sources JSON object"},
			{"--no-color", "Disable ANSI colors"},
			{"--verbose", "Show additional diagnostics on stderr"},
		},
		Examples: []string{"thermal sources", "thermal sources --json"},
		JSON:     "keys: type, generatedAt, home, summary, sources",
		Related:  []string{"thermal audit", "thermal --help"},
	},
	"audit": {
		Title:   "audit",
		Usage:   "thermal audit [options]",
		Summary: "Local setup and context health",
		Description: []string{
			"Audits instruction files, MCP server density, context tax, and readiness. Scans the current repository, not agent session data.",
		},
		Flags: []render.HelpFlag{
			{"--json", "Emit the audit JSON object"},
			{"--no-color", "Disable ANSI colors"},
			{"--verbose", "Show additional diagnostics on stderr"},
		},
		Examples: []string{"thermal audit", "thermal audit --json"},
		JSON:     "keys: type, score, grade, contextTax, rules, mcp, readiness, recommendations",
		Related:  []string{"thermal changelog", "thermal license"},
	},
	"changelog": {
		Title:   "changelog",
		Usage:   "thermal changelog [options]",
		Summary: "Release history and new features",
		Description: []string{
			"Reads the embedded changelog. The default shows three releases; --top 0 shows all.",
		},
		Flags: []render.HelpFlag{
			{"--top <N>", "Number of releases to show (0 = all, default 3)"},
			{"--json", "Emit the changelog JSON object"},
			{"--no-color", "Disable ANSI colors"},
		},
		Examples: []string{"thermal changelog", "thermal changelog --top 5", "thermal changelog --top 0 --json"},
		JSON:     "keys: type, currentVersion, releases, generatedAt",
		Related:  []string{"thermal version", "thermal upgrade"},
	},
	"upgrade": {
		Title:   "upgrade",
		Usage:   "thermal upgrade [options]",
		Summary: "Self-upgrade thermal to the latest release",
		Description: []string{
			"Replaces the running binary with the latest release. --check and --dry-run never download or mutate; a mutation requires a terminal confirmation or --yes.",
			"Non-interactive mutation without --yes exits 3. --json emits exactly one object on stdout.",
		},
		Flags: []render.HelpFlag{
			{"--check", "Report current, latest, and update availability; no mutation"},
			{"--dry-run", "Resolve the asset and target; no download or mutation"},
			{"--json", "Emit one upgrade JSON object on stdout"},
			{"--yes", "Confirm mutation without a prompt (required for non-TTY)"},
			{"--force", "Reinstall the same version or downgrade to the latest release"},
			{"-h, --help", "Show this page without upgrading"},
		},
		Examples: []string{"thermal upgrade --check", "thermal upgrade --dry-run --json", "thermal upgrade --yes"},
		JSON:     "keys: type, action, current, latest, asset, target, mutated, migration, error",
		Note:     "A failed download, extraction, or rename leaves the existing binary intact.",
		Related:  []string{"thermal version", "thermal changelog"},
	},
	"version": {
		Title:   "version",
		Usage:   "thermal version [options]",
		Summary: "SemVer, commit hash, and build timestamp",
		Description: []string{
			"Prints build metadata. Read-only; no data scan, network call, or update check.",
		},
		Flags: []render.HelpFlag{
			{"--json", "Emit the version JSON object"},
			{"-V, --version", "Short forms of this command"},
			{"-h, --help", "Show this page"},
		},
		Examples: []string{"thermal version", "thermal -V --json", "thermal --version"},
		JSON:     "keys: type, generatedAt, version, commit, date",
		Note:     "Short forms: thermal -V and thermal --version. The -v flag means verbose, not version.",
		Related:  []string{"thermal changelog", "thermal upgrade"},
	},
	"license": {
		Title:   "license",
		Usage:   "thermal license [options]",
		Summary: "AGPL-3.0 and commercial enterprise licensing",
		Description: []string{
			"Prints the dual-licensing terms, commercial pricing, and covenant.",
		},
		Flags: []render.HelpFlag{
			{"--json", "Emit the license JSON object"},
			{"-h, --help", "Show this page"},
		},
		Examples: []string{"thermal license", "thermal license --json"},
		JSON:     "keys: type, name, author, license, covenant, commercial_licensing",
		Related:  []string{"thermal version", "thermal audit"},
	},
}

func helpHelpPage() render.CommandHelp {
	return render.CommandHelp{
		Title:   "help",
		Usage:   "thermal help [command]  |  thermal <command> --help",
		Summary: "Global and command-scoped help",
		Description: []string{
			"Help is side-effect free: it never reads agent data, opens a TUI, calls the network, writes a cache, or replaces the binary.",
			"A help request wins over argument validation and command execution, so thermal upgrade --help is always safe.",
		},
		Examples: []string{"thermal --help", "thermal help daily", "thermal daily --help"},
		JSON:     "not supported",
		Related:  []string{"thermal version", "thermal license"},
	}
}

func toolHelpPage(name string) (render.CommandHelp, bool) {
	entry, ok := findToolEntry(name)
	if !ok {
		return render.CommandHelp{}, false
	}
	desc := []string{
		"Reads only " + entry.Display + "'s local session data. Add a report word to scope the aggregation.",
		"Reports: daily, weekly, monthly, projects, models, stats, mix, trend.",
	}
	if entry.Alias != "" {
		desc = append(desc, "CLI alias: "+entry.Alias+".")
	}
	return render.CommandHelp{
		Title:       entry.Display,
		Usage:       "thermal " + entry.Canonical + " [report] [options]",
		Summary:     "Scoped activity, tokens, and cost",
		Description: desc,
		Flags: []render.HelpFlag{
			{"--db <path>", "Override the data file or directory"},
			{"--weeks <N>", "Heatmap width in weeks (4-104, default 52; auto-fits narrow terminals)"},
			{"--width <N>", "Text render width in columns (40-240; overrides COLUMNS)"},
			{"--json", "Emit the scoped JSON envelope"},
			{"--no-color", "Disable ANSI colors"},
			{"--verbose", "Show loader diagnostics on stderr"},
			{"--offline", "Use cached pricing only"},
			{"--no-estimate", "Recorded cost only"},
		},
		Examples: []string{
			"thermal " + entry.Canonical,
			"thermal " + entry.Canonical + " daily --last 7",
			"thermal " + entry.Canonical + " --json",
		},
		JSON:    "keys: type, tool, summary, daily, projects, generatedAt",
		Related: []string{"thermal daily", "thermal models"},
	}, true
}

func scopeHelpPage(page render.CommandHelp, name string) render.CommandHelp {
	entry, ok := findToolEntry(name)
	if !ok {
		return page
	}
	page.Title = page.Title + " · " + entry.Display
	usage := strings.Replace(page.Usage, "[tool] ", "", 1)
	page.Usage = "thermal " + entry.Canonical + " " + strings.TrimPrefix(usage, "thermal ")
	page.Description = append([]string{"Scoped to " + entry.Display + "."}, page.Description...)
	page.Examples = append([]string{"thermal " + entry.Canonical + " " + strings.TrimPrefix(page.Examples[0], "thermal ")}, page.Examples...)
	return page
}

var retiredHelp = map[string]string{
	"dashboard": "the interactive dashboard has been retired in favor of 'thermal live' and static report commands.",
	"serve":     "the embedded web server has been retired; use thermal live --json --stream.",
	"web":       "the embedded web server has been retired; use thermal live --json --stream.",
	"replay":    "subscription replay simulation has been retired; use thermal monthly and thermal models.",
	"yield":     "token yield efficiency reporting has been retired; use thermal stats and thermal models.",
	"receipt":   "work receipts have been retired; use thermal share.",
}

var helpSynonyms = map[string]string{
	"today":     "daily",
	"yesterday": "daily",
	"week":      "weekly",
	"weak":      "weekly",
	"month":     "monthly",
	"repo":      "projects",
	"project":   "projects",
	"model":     "models",
	"cost":      "stats",
	"spend":     "stats",
	"status":    "statusline",
	"prompt":    "statusline",
}

var helpCandidates = func() []string {
	names := []string{
		"daily", "weekly", "monthly", "projects", "models", "stats", "mix", "trend",
		"live", "statusline", "share", "audit", "changelog", "upgrade", "version",
		"license", "help", "prompt", "status",
	}
	for _, e := range toolHelpEntries {
		names = append(names, e.Canonical)
		if e.Alias != "" {
			names = append(names, e.Alias)
		}
	}
	return names
}()

func nearestHelpName(token string) (string, bool) {
	t := strings.ToLower(token)
	if v, ok := helpSynonyms[t]; ok {
		return v, true
	}
	return nearestName(t, helpCandidates)
}

// resolveHelpTarget scans raw argv for a help request and resolves the page
// context. It only inspects tokens; it never validates values or reads data.
func resolveHelpTarget(args []string) (verb, tool, unknown string, help bool) {
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if a == "-h" || a == "--help" {
			help = true
			continue
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			name := a
			if eq := strings.Index(a, "="); eq >= 0 {
				name = a[:eq]
			}
			key := strings.TrimLeft(name, "-")
			if _, ok := flagMetas[key]; ok && !strings.Contains(a, "=") {
				i++
			}
			continue
		}
		pos = append(pos, a)
	}
	if len(pos) > 0 && strings.EqualFold(pos[0], "help") {
		help = true
		pos = pos[1:]
	}
	if !help {
		return "", "", "", false
	}
	if len(pos) == 0 {
		return "", "", "", true
	}

	p0 := strings.ToLower(pos[0])
	if p0 == "help" {
		return "help", "", "", true
	}
	if _, ok := retiredHelp[p0]; ok {
		return p0, "", "", true
	}
	if p0 == "status" {
		return "status", "", "", true
	}
	if p0 == "all" || p0 == "auto" {
		return "", "", "", true
	}
	if isReportWord(p0) {
		if len(pos) > 1 {
			if _, ok := findToolEntry(pos[1]); ok {
				tool = pos[1]
			}
		}
		return p0, tool, "", true
	}
	if _, ok := commandHelps[p0]; ok {
		return p0, "", "", true
	}
	if _, ok := findToolEntry(p0); ok {
		tool = p0
		if len(pos) > 1 && isReportWord(pos[1]) {
			verb = strings.ToLower(pos[1])
		}
		return verb, tool, "", true
	}
	return "", "", p0, true
}

func noColorFromArgs(args []string) bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	for _, a := range args {
		if a == "--no-color" || strings.HasPrefix(a, "--no-color=") {
			return true
		}
	}
	return false
}

// helpOutput returns the text and exit code for a help request without writing.
func helpOutput(verb, tool, unknown string, noColor bool) (string, int) {
	if unknown != "" {
		var b strings.Builder
		fmt.Fprintf(&b, "thermal: unknown command %q\n", unknown)
		if s, ok := nearestHelpName(unknown); ok {
			fmt.Fprintf(&b, "  did you mean: thermal %s?\n", s)
		}
		b.WriteString("  run thermal --help for the command list\n")
		return b.String(), 1
	}
	if msg, ok := retiredHelp[verb]; ok {
		see := "See also: thermal changelog"
		if verb == "dashboard" {
			see = "See also: thermal live"
		}
		return fmt.Sprintf("thermal: %s\n  %s\n", msg, see), 1
	}
	if verb == "" && tool == "" {
		return render.RenderHelp(noColor), 0
	}
	if verb == "help" {
		return render.RenderCommandHelp(helpHelpPage()), 0
	}
	if tool != "" && verb == "" {
		if page, ok := toolHelpPage(tool); ok {
			return render.RenderCommandHelp(page), 0
		}
		return fmt.Sprintf("thermal: unknown command %q\n", tool), 1
	}

	alias := ""
	pageName := verb
	if verb == "prompt" || verb == "status" {
		alias = verb
		pageName = "statusline"
	}
	page, ok := commandHelps[pageName]
	if !ok {
		return fmt.Sprintf("thermal: unknown command %q\n", verb), 1
	}
	if tool != "" {
		page = scopeHelpPage(page, tool)
	}
	if alias != "" {
		page.Usage = strings.Replace(page.Usage, "thermal statusline", "thermal "+alias, 1)
		page.Note = "Alias: " + alias
	}
	return render.RenderCommandHelp(page), 0
}

func printHelpFor(verb, tool, unknown string, noColor bool) int {
	if width := render.TerminalWidth(); width < 48 {
		fmt.Fprint(os.Stderr, render.WidthRefusal("help", 48, width))
		return 2
	}
	out, code := helpOutput(verb, tool, unknown, noColor)
	if code == 0 {
		fmt.Print(out)
	} else {
		fmt.Fprint(os.Stderr, out)
	}
	return code
}
