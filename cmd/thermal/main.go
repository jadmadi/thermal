package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"
)

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

func parseArgs() thermal.Options {
	var opts thermal.Options
	flag.StringVar(&opts.Tool, "tool", "all", "Tool: all, mimocode, opencode, codex, agy, command-code, codewhale")
	flag.StringVar(&opts.DBPath, "db", "", "Override database/data path")
	flag.IntVar(&opts.Weeks, "weeks", 52, "Heatmap width in weeks (4-104)")
	flag.BoolVar(&opts.JSON, "json", false, "Output JSON instead of dashboard")
	flag.BoolVar(&opts.NoColor, "no-color", false, "Disable ANSI colors")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage())
	}
	flag.Parse()

	remaining := flag.Args()
	if len(remaining) > 1 {
		for i := 1; i < len(remaining); i++ {
			arg := remaining[i]
			switch {
			case arg == "--no-color":
				opts.NoColor = true
			case arg == "--json":
				opts.JSON = true
			case strings.HasPrefix(arg, "--weeks="):
				if v, err := strconv.Atoi(strings.TrimPrefix(arg, "--weeks=")); err == nil {
					opts.Weeks = v
				}
			case arg == "--weeks" && i+1 < len(remaining):
				if v, err := strconv.Atoi(remaining[i+1]); err == nil {
					opts.Weeks = v
					i++
				}
			case strings.HasPrefix(arg, "--db="):
				opts.DBPath = strings.TrimPrefix(arg, "--db=")
			case arg == "--db" && i+1 < len(remaining):
				opts.DBPath = remaining[i+1]
				i++
			case strings.HasPrefix(arg, "--tool="):
				opts.Tool = strings.TrimPrefix(arg, "--tool=")
			case arg == "--tool" && i+1 < len(remaining):
				opts.Tool = remaining[i+1]
				i++
			}
		}
	}

	if len(remaining) > 0 && !strings.HasPrefix(remaining[0], "-") {
		opts.Tool = remaining[0]
	}

	if opts.Weeks < 4 || opts.Weeks > 104 {
		fmt.Fprintf(os.Stderr, "thermal: --weeks must be between 4 and 104\n")
		os.Exit(1)
	}

	return opts
}

func main() {
	opts := parseArgs()

	if opts.Tool == "all" {
		tools := loaders.AllTools()
		var results []thermal.ToolResult

		toolOrder := []thermal.Tool{thermal.ToolMiMoCode, thermal.ToolOpenCode, thermal.ToolCodex, thermal.ToolDevin, thermal.ToolAgy, thermal.ToolCommandCode, thermal.ToolCodewhale}
		for _, t := range toolOrder {
			info := tools[t]
			if info.DBPath != "" {
				if _, err := os.Stat(info.DBPath); os.IsNotExist(err) {
					continue
				}
			} else if info.DataDir != "" {
				if _, err := os.Stat(info.DataDir); os.IsNotExist(err) {
					continue
				}
			} else {
				continue
			}

			summary, daily, dataPath, err := loaders.LoadToolData(t, info, "")
			if err != nil {
				continue
			}

			activeDays := make(map[string]bool)
			for _, d := range daily {
				if d.Turns > 0 {
					activeDays[d.Day] = true
				}
			}
			current, longest := thermal.ComputeStreaks(activeDays)

			results = append(results, thermal.ToolResult{
				Tool:          t,
				Name:          info.Name,
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

		fmt.Print(render.RenderLeaderboard(results, opts.Weeks, opts.NoColor))
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

	summary, daily, dataPath, err := loaders.LoadToolData(tool, info, opts.DBPath)
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
	current, longest := thermal.ComputeStreaks(activeDays)

	if opts.JSON {
		type jsonSummary struct {
			thermal.Summary
			CurrentStreak int `json:"currentStreak"`
			LongestStreak int `json:"longestStreak"`
		}
		out := map[string]interface{}{
			"tool":        info.Name,
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

	fmt.Print(render.RenderDashboard(info.Name, summary, daily, dataPath, opts.Weeks, opts.NoColor))
}
