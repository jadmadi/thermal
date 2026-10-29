// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/thermal"
)

var errHelpRequested = errors.New("help requested")

// usageError is a parse/usage failure. It is printed as one cause line, an
// optional fix line, an optional example, and a help pointer.
type usageError struct {
	Message string
	Fix     string
	Example string
}

func (e *usageError) Error() string {
	var b strings.Builder
	b.WriteString(e.Message)
	if e.Fix != "" {
		b.WriteString("\n  " + e.Fix)
	}
	if e.Example != "" {
		b.WriteString("\n  example: " + e.Example)
	}
	return b.String()
}

// retiredError preserves the specific retirement message for flags that were
// removed from the CLI. It exits 1, not 2.
type retiredError struct{ Message string }

func (e *retiredError) Error() string { return e.Message }

type flagKind int

const (
	kindString flagKind = iota
	kindInt
	kindDuration
	kindCompactFloat
)

type flagMeta struct {
	Kind    flagKind
	Want    string
	Example string
}

// flagMetas is the single metadata table for value-taking flags. The flag
// registrations in parseArgsFrom supply the actual values; this table supplies
// validation, user-facing expectations, and examples.
var flagMetas = map[string]flagMeta{
	"tool":          {kindString, "a tool name", "thermal --tool opencode daily"},
	"db":            {kindString, "a database or directory path", "thermal --tool opencode --db <path> daily"},
	"weeks":         {kindInt, "a number between 4 and 104", "thermal daily --weeks 12"},
	"since":         {kindString, "a date (YYYY-MM-DD or YYYYMMDD)", "thermal daily --since 2026-10-01"},
	"until":         {kindString, "a date (YYYY-MM-DD or YYYYMMDD)", "thermal daily --until 2026-10-31"},
	"last":          {kindInt, "a whole number of days", "thermal daily --last 7"},
	"order":         {kindString, "asc or desc", "thermal daily --order asc"},
	"sort":          {kindString, "a sort key", "thermal --sort tokens"},
	"top":           {kindInt, "a non-negative whole number", "thermal projects --top 5"},
	"metric":        {kindString, "tokens or cost", "thermal stats --metric cost"},
	"by":            {kindString, "tool, model, or reasoning", "thermal mix --by model"},
	"grain":         {kindString, "day, week, or month", "thermal mix --grain month"},
	"start-of-week": {kindString, "a weekday name", "thermal daily --start-of-week monday"},
	"interval":      {kindDuration, "a duration like 500ms, 1s, or 2m", "thermal live --interval 1s"},
	"flame-ceiling": {kindCompactFloat, "a positive number like 9000000, 9M, or 60", "thermal live --flame-ceiling 9M"},
	"format":        {kindString, "text or json", "thermal daily --format json"},
}

// normalizeArgs separates flags from positionals, validates every value, and
// returns a flag-only argv for the stdlib FlagSet. It is the single strict
// parser: flags may appear before, between, or after positionals.
func normalizeArgs(args []string, fs *flag.FlagSet) (flagArgs, positionals []string, present map[string]bool, err error) {
	present = make(map[string]bool)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			positionals = append(positionals, arg)
			continue
		}
		if arg == "-h" || arg == "--help" {
			return nil, nil, present, errHelpRequested
		}
		if arg == "-fresh" || arg == "--zero" {
			arg = "--fresh"
		}
		if arg == "-license" {
			arg = "--license"
		}

		name, inline, hasInline := splitFlag(arg)
		switch name {
		case "--dense", "--distribution", "--dist", "--sparse":
			return nil, nil, present, &retiredError{
				Message: "the dense FinOps grid and distribution aliases have been retired.\n" +
					"  thermal stats now prints the distribution report; use thermal live for real-time monitoring.",
			}
		case "--port", "--host", "--open", "--no-open":
			return nil, nil, present, &retiredError{
				Message: "the embedded web server has been retired; web flags are no longer supported.\n" +
					"  For headless telemetry and streaming: thermal live --json --stream",
			}
		case "--against", "--compare":
			return nil, nil, present, &retiredError{
				Message: "subscription replay simulation has been retired; --against and --compare are no longer supported.\n" +
					"  For monthly volume and model pricing: thermal monthly, thermal models",
			}
		}

		key := strings.TrimLeft(name, "-")
		f := fs.Lookup(key)
		if f == nil {
			return nil, nil, present, unknownFlagError(name, fs)
		}
		present[key] = true

		if isBoolFlag(f) {
			if hasInline {
				if _, parseErr := strconv.ParseBool(inline); parseErr != nil {
					return nil, nil, present, &usageError{
						Message: fmt.Sprintf("%s wants true or false, got %q", name, inline),
						Example: "thermal " + strings.TrimLeft(name, "-"),
					}
				}
			}
			flagArgs = append(flagArgs, arg)
			continue
		}

		meta, ok := flagMetas[key]
		if !ok {
			meta = flagMeta{Kind: kindString, Want: "a value"}
		}

		var val string
		if hasInline {
			val = inline
			if val == "" {
				return nil, nil, present, emptyValueError(name, meta)
			}
		} else {
			if i+1 >= len(args) || (looksLikeFlag(args[i+1]) && !(meta.Kind != kindString && isNumericToken(args[i+1]))) {
				return nil, nil, present, missingValueError(name, meta)
			}
			val = args[i+1]
			i++
		}

		if err := validateFlagValue(name, val, meta); err != nil {
			return nil, nil, present, err
		}
		if meta.Kind == kindDuration {
			if _, parseErr := time.ParseDuration(val); parseErr != nil {
				sec, _ := strconv.ParseFloat(val, 64)
				val = time.Duration(sec * float64(time.Second)).String()
			}
		}
		if hasInline {
			flagArgs = append(flagArgs, name+"="+val)
		} else {
			flagArgs = append(flagArgs, name, val)
		}
	}

	return flagArgs, positionals, present, nil
}

func splitFlag(arg string) (name, inline string, hasInline bool) {
	if eq := strings.Index(arg, "="); eq >= 0 {
		return arg[:eq], arg[eq+1:], true
	}
	return arg, "", false
}

func looksLikeFlag(s string) bool {
	return strings.HasPrefix(s, "-") && s != "-"
}

func isNumericToken(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	dot := false
	for _, r := range s {
		if r == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isBoolFlag(f *flag.Flag) bool {
	bf, ok := f.Value.(interface{ IsBoolFlag() bool })
	return ok && bf.IsBoolFlag()
}

func validateFlagValue(name, val string, meta flagMeta) error {
	switch meta.Kind {
	case kindInt:
		n, err := strconv.Atoi(val)
		if err != nil {
			return &usageError{
				Message: fmt.Sprintf("%s wants %s, got %q", name, meta.Want, val),
				Example: meta.Example,
			}
		}
		switch name {
		case "--weeks":
			if n < 4 || n > 104 {
				return &usageError{Message: fmt.Sprintf("--weeks wants a number between 4 and 104, got %d", n), Example: meta.Example}
			}
		case "--last":
			if n < 1 {
				return &usageError{Message: fmt.Sprintf("--last must be at least 1, got %d", n), Example: meta.Example}
			}
		case "--top":
			if n < 0 {
				return &usageError{Message: fmt.Sprintf("--top cannot be negative, got %d", n), Example: meta.Example}
			}
		}
	case kindDuration:
		if d, err := time.ParseDuration(val); err == nil {
			if d <= 0 {
				return &usageError{Message: fmt.Sprintf("--interval must be positive, got %q", val), Example: meta.Example}
			}
			return nil
		}
		sec, err := strconv.ParseFloat(val, 64)
		if err != nil || math.IsNaN(sec) || math.IsInf(sec, 0) || sec <= 0 {
			return &usageError{Message: fmt.Sprintf("--interval wants %s, got %q", meta.Want, val), Example: meta.Example}
		}
	case kindCompactFloat:
		v, err := parseCompactFloat(val)
		if err != nil || v <= 0 {
			return &usageError{Message: fmt.Sprintf("--flame-ceiling wants %s, got %q", meta.Want, val), Example: meta.Example}
		}
	default:
		if strings.TrimSpace(val) == "" {
			return emptyValueError(name, meta)
		}
	}
	return nil
}

func missingValueError(name string, meta flagMeta) error {
	return &usageError{
		Message: fmt.Sprintf("%s needs a value (%s)", name, meta.Want),
		Example: meta.Example,
	}
}

func emptyValueError(name string, meta flagMeta) error {
	return &usageError{
		Message: fmt.Sprintf("%s requires a non-empty value", name),
		Example: meta.Example,
	}
}

func unknownFlagError(name string, fs *flag.FlagSet) error {
	key := strings.TrimLeft(name, "-")
	var names []string
	fs.VisitAll(func(f *flag.Flag) { names = append(names, f.Name) })
	best, ok := nearestName(key, names)
	fix := ""
	if ok {
		prefix := "--"
		if strings.HasPrefix(name, "-") && !strings.HasPrefix(name, "--") {
			prefix = "-"
		}
		fix = fmt.Sprintf("did you mean %s%s?", prefix, best)
	}
	return &usageError{Message: fmt.Sprintf("unknown flag %s", name), Fix: fix}
}

func nearestName(key string, names []string) (string, bool) {
	if len(key) < 3 {
		return "", false
	}
	best, bestDist := "", 3
	for _, n := range names {
		if n == key {
			continue
		}
		d := levenshtein(strings.ToLower(key), strings.ToLower(n))
		if d < bestDist || (d == bestDist && n < best) {
			best, bestDist = n, d
		}
	}
	if bestDist <= 2 {
		return best, true
	}
	return "", false
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := 0; j <= len(br); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func printUsageError(w io.Writer, e *usageError) {
	fmt.Fprintf(w, "thermal: %s\n", e.Message)
	if e.Fix != "" {
		fmt.Fprintf(w, "  %s\n", e.Fix)
	}
	if e.Example != "" {
		fmt.Fprintf(w, "  example: %s\n", e.Example)
	}
	if e.Example == "" {
		fmt.Fprintln(w, "  run thermal --help for the flag list")
	}
}

// parseArgsFrom parses arguments without touching os.Args or exiting. The
// dispatcher and tests both use it.
func parseArgsFrom(args []string) (thermal.Options, error) {
	var opts thermal.Options
	fs := flag.NewFlagSet("thermal", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	fs.StringVar(&opts.Tool, "tool", "all", "Tool: all, mimocode, opencode, codex, devin, agy, command-code, codewhale, zcode, grok, muse, claude, droid, dsh, hermes, zed")
	fs.StringVar(&opts.DBPath, "db", "", "Override database/data path")
	fs.IntVar(&opts.Weeks, "weeks", 52, "Heatmap width in weeks (4-104)")
	fs.StringVar(&opts.Since, "since", "", "Report window start (YYYY-MM-DD or YYYYMMDD)")
	fs.StringVar(&opts.Until, "until", "", "Report window end (YYYY-MM-DD or YYYYMMDD)")
	fs.IntVar(&opts.Last, "last", 0, "Last N days/weeks/months for reports")
	fs.StringVar(&opts.Order, "order", "desc", "Report sort order: asc or desc")
	fs.StringVar(&opts.Sort, "sort", "", "Ranking: streak|tokens|cost for the leaderboard, tokens|cost|days|recent for projects, tokens|cost for models")
	fs.IntVar(&opts.Top, "top", 0, "Project or model rows to print, 0 for all")
	fs.StringVar(&opts.Metric, "metric", "tokens", "Analytics metric: tokens or cost")
	fs.StringVar(&opts.By, "by", "tool", "Mix dimension: tool or model")
	fs.StringVar(&opts.Grain, "grain", "week", "Mix bucket size: day, week, or month")
	fs.BoolVar(&opts.Breakdown, "breakdown", false, "Show per-model rows in reports")
	fs.BoolVar(&opts.Chart, "chart", false, "Print bar rows under report tables")
	fs.StringVar(&opts.StartOfWeek, "start-of-week", "sunday", "Week start day: sunday-saturday")
	fs.BoolVar(&opts.Offline, "offline", false, "Use cached pricing only, never fetch")
	fs.BoolVar(&opts.NoEstimate, "no-estimate", false, "Recorded cost only, no pricing estimates")
	fs.BoolVar(&opts.NoUpdateCheck, "no-update-check", false, "Disable daily automatic update check")
	fs.DurationVar(&opts.Interval, "interval", time.Second, "Polling interval for live monitor (default: 1s)")
	var flameCeilingRaw string
	fs.StringVar(&flameCeilingRaw, "flame-ceiling", "", "Live flame ceiling in active units per minute (e.g. 9M or 60)")
	fs.BoolVar(&opts.Stream, "stream", false, "Stream continuous NDJSON live events (live verb)")
	fs.BoolVar(&opts.Stream, "follow", false, "Stream continuous NDJSON live events (live verb shorthand)")
	fs.BoolVar(&opts.Stream, "f", false, "Stream continuous NDJSON live events (live verb shorthand)")
	fs.BoolVar(&opts.Fresh, "fresh", false, "Bypass cached statusline data; live sessions always start at launch")
	var showLicense bool
	fs.BoolVar(&showLicense, "license", false, "Show license, dual-licensing & commercial terms")
	fs.StringVar(&opts.Format, "format", "text", "Output format: text or json")
	fs.BoolVar(&opts.Nerd, "nerd", false, "Use Nerd Font glyphs in statusline")
	fs.BoolVar(&opts.Plain, "plain", false, "Plain text output without glyphs or icons")
	fs.BoolVar(&opts.JSON, "json", false, "Output JSON instead of the terminal UI")
	fs.BoolVar(&opts.NoColor, "no-color", false, "Disable ANSI colors")
	fs.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose warning diagnostics on stderr")
	fs.BoolVar(&opts.Verbose, "v", false, "Enable verbose warning diagnostics on stderr (shorthand)")
	fs.BoolVar(&opts.Verbose, "d", false, "Enable verbose warning diagnostics on stderr (shorthand)")
	var showVersion bool
	fs.BoolVar(&showVersion, "V", false, "Print version metadata (shorthand)")
	fs.BoolVar(&showVersion, "version", false, "Print version metadata")

	flagArgs, positionals, present, err := normalizeArgs(args, fs)
	if err != nil {
		return opts, err
	}
	if err := fs.Parse(flagArgs); err != nil {
		return opts, &usageError{Message: err.Error()}
	}
	if showVersion {
		return opts, errVersionRequested
	}
	opts.LastSet = present["last"]
	opts.TopSet = present["top"]

	if flameCeilingRaw != "" {
		v, parseErr := parseCompactFloat(flameCeilingRaw)
		if parseErr != nil {
			return opts, &usageError{
				Message: fmt.Sprintf("--flame-ceiling wants a positive number like 9000000, 9M, or 60, got %q", flameCeilingRaw),
				Example: "thermal live --flame-ceiling 9M",
			}
		}
		opts.FlameCeiling = v
		opts.FlameCeilingSet = true
	}

	opts.Format = strings.ToLower(strings.TrimSpace(opts.Format))
	if opts.Format == "" {
		opts.Format = "text"
	}
	if opts.JSON && (opts.Format == "md" || opts.Format == "markdown") {
		return opts, &usageError{Message: fmt.Sprintf("--json cannot be combined with --format %s", opts.Format)}
	}
	if opts.JSON {
		opts.Format = "json"
	} else if opts.Format == "json" {
		opts.JSON = true
	}

	if len(positionals) > 0 {
		if len(positionals) > 2 {
			return opts, &usageError{Message: fmt.Sprintf("unexpected argument %q", positionals[2])}
		}
		posTool := ""
		if isReportWord(positionals[0]) {
			opts.Report = strings.ToLower(positionals[0])
			if len(positionals) > 1 {
				posTool = positionals[1]
			}
		} else {
			posTool = positionals[0]
			if len(positionals) > 1 {
				if !isReportWord(positionals[1]) {
					return opts, &usageError{Message: fmt.Sprintf("unexpected argument %q", positionals[1])}
				}
				opts.Report = strings.ToLower(positionals[1])
			}
		}
		if posTool != "" {
			if present["tool"] && !strings.EqualFold(opts.Tool, posTool) {
				return opts, &usageError{
					Message: fmt.Sprintf("positional tool %q conflicts with --tool %s", posTool, opts.Tool),
					Fix:     "remove one of them, or make both the same",
				}
			}
			opts.Tool = posTool
		}
	}

	if showLicense {
		opts.Tool = "license"
	}
	if opts.DBPath != "" && (opts.Tool == "all" || opts.Tool == "auto" || opts.Tool == "") {
		return opts, &usageError{
			Message: "--db applies to one tool; add --tool <name>",
			Example: "thermal --tool opencode --db <path> daily",
		}
	}
	if opts.Report == "statusline" || opts.Report == "prompt" || opts.Report == "status" || opts.Tool == "statusline" || opts.Tool == "prompt" || opts.Tool == "status" {
		opts.NoUpdateCheck = true
	}

	if opts.Weeks < 4 || opts.Weeks > 104 {
		return opts, &usageError{Message: fmt.Sprintf("--weeks wants a number between 4 and 104, got %d", opts.Weeks), Example: "thermal daily --weeks 12"}
	}
	if err := validateReportFlags(opts); err != nil {
		return opts, &usageError{Message: err.Error()}
	}
	return opts, nil
}
