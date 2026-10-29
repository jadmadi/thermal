// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
)

// HelpItem represents a single command or option entry in help output.
type HelpItem struct {
	Name string
	Desc string
}

// HelpSection represents a categorized group of commands or options.
type HelpSection struct {
	Title string
	Items []HelpItem
}

// padStyled pads a styled string with trailing spaces based on plain text length.
func padStyled(styled string, plainLen int, targetWidth int) string {
	if targetWidth > plainLen {
		return styled + strings.Repeat(" ", targetWidth-plainLen)
	}
	return styled
}

// RenderHelp returns the formatted, categorized, and themed help manual.
func RenderHelp(noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	secTitle := func(s string) string {
		return theme.Primary.SprintBold(colors, s)
	}
	cmdName := func(s string) string {
		return theme.Secondary.SprintBold(colors, s)
	}
	flagName := func(s string) string {
		// Style flag name in secondary, and dim the placeholder (e.g. <name>)
		if idx := strings.Index(s, " <"); idx != -1 {
			return theme.Secondary.Sprint(colors, s[:idx]) + " " + dim(s[idx+1:])
		}
		return theme.Secondary.Sprint(colors, s)
	}

	termWidth := TerminalWidth()
	if termWidth > MaxCardWidth {
		termWidth = MaxCardWidth
	}
	if termWidth < MinCardWidth {
		termWidth = MinCardWidth
	}

	var sb strings.Builder

	// 1. Header Banner
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("usage profile & activity heatmaps for AI coding tools"),
	))

	// 2. Usage Section
	sb.WriteString(fmt.Sprintf("  %s\n", secTitle("Usage:")))
	usageItems := []struct {
		cmd  string
		desc string
	}{
		{"thermal [options]", "All-tool leaderboard & streak heatmap"},
		{"thermal [command] [options]", "Analytics, audit, and utility commands"},
		{"thermal [tool] [report] [options]", "Scoped report (e.g. opencode weekly)"},
		{"thermal help [command]", "Global help or one command's help page"},
	}
	for _, u := range usageItems {
		padded := padStyled(cmdName(u.cmd), len(u.cmd), 36)
		avail := termWidth - (4 + 36 + 1)
		wrapped := WrapText(u.desc, avail)
		for i, line := range wrapped {
			if i == 0 {
				sb.WriteString(fmt.Sprintf("    %s %s\n", padded, dim(line)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", 4+36+1), dim(line)))
			}
		}
	}
	sb.WriteString("\n")

	// 3. Command Categories
	cmdSections := []HelpSection{
		{
			Title: "Dashboards & Views:",
			Items: []HelpItem{
				{Name: "live", Desc: "Real-time token burn monitor with velocity, flame & sparkline"},
				{Name: "statusline", Desc: "Instant shell statusline and prompt helper (<3ms, alias: prompt)"},
				{Name: "share", Desc: "Generate a stateless, private URL card for your streak"},
			},
		},
		{
			Title: "Reports & Rankings:",
			Items: []HelpItem{
				{Name: "daily", Desc: "Daily token volume and cost breakdown"},
				{Name: "weekly", Desc: "Weekly aggregated activity and spend"},
				{Name: "monthly", Desc: "Monthly summary and model breakdown"},
				{Name: "projects", Desc: "Tokens and cost per project, folded to git root"},
				{Name: "models", Desc: "Tokens and estimated cost per model, ranked"},
			},
		},
		{
			Title: "FinOps & Analytics:",
			Items: []HelpItem{
				{Name: "stats", Desc: "Daily distribution, percentiles, and token composition"},
				{Name: "mix", Desc: "Tool and model market-share & switching frequency"},
				{Name: "trend", Desc: "Daily trend fit with month-end volume & spend projection"},
			},
		},
		{
			Title: "Diagnostics & Utilities:",
			Items: []HelpItem{
				{Name: "help", Desc: "Global help or one command's help page (e.g. thermal help daily)"},
				{Name: "audit", Desc: "Audit local agent instruction files & MCP context tax"},
				{Name: "changelog", Desc: "Show release history, newly added commands & arguments"},
				{Name: "upgrade", Desc: "Self-upgrade thermal to the latest release"},
				{Name: "version", Desc: "Print SemVer, commit hash, and build timestamp"},
				{Name: "license", Desc: "Display AGPL-3.0 and commercial enterprise licensing"},
			},
		},
	}

	for _, sec := range cmdSections {
		sb.WriteString(fmt.Sprintf("  %s\n", secTitle(sec.Title)))
		for _, item := range sec.Items {
			padded := padStyled(cmdName(item.Name), len(item.Name), 14)
			avail := termWidth - 19
			wrapped := WrapText(item.Desc, avail)
			for i, line := range wrapped {
				if i == 0 {
					sb.WriteString(fmt.Sprintf("    %s %s\n", padded, line))
				} else {
					sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", 19), line))
				}
			}
		}
		sb.WriteString("\n")
	}

	// 4. Common Examples
	sb.WriteString(fmt.Sprintf("  %s\n", secTitle("Common Examples:")))
	examples := []struct {
		cmd  string
		desc string
	}{
		{"thermal", "All-tool leaderboard & streak heatmap"},
		{"thermal opencode weekly --chart", "Weekly breakdown with terminal bar chart"},
		{"thermal daily --help", "Scoped help for the daily report"},
		{"thermal help daily", "Same page via the help command"},
		{"thermal stats", "Daily distribution and token composition"},
		{"thermal audit", "Audit instruction files & MCP context tax"},
		{"thermal changelog --top 5", "View latest new features & arguments"},
	}
	for _, ex := range examples {
		styled := theme.Text.Sprint(colors, ex.cmd)
		padded := padStyled(styled, len(ex.cmd), 34)
		avail := termWidth - (4 + 34 + 1)
		wrapped := WrapText(ex.desc, avail)
		for i, line := range wrapped {
			if i == 0 {
				sb.WriteString(fmt.Sprintf("    %s %s\n", padded, dim(line)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", 4+34+1), dim(line)))
			}
		}
	}
	sb.WriteString("\n")

	// 5. Option Categories
	optSections := []HelpSection{
		{
			Title: "Filter & Window Options:",
			Items: []HelpItem{
				{Name: "--tool <name>", Desc: "Filter to a single tool (default: all)"},
				{Name: "--interval <dur>", Desc: "Polling interval for live monitor (default: 1s)"},
				{Name: "--flame-ceiling <N>", Desc: "Live flame ceiling in active units per minute (e.g. 9M or 60)"},
				{Name: "--since <date>", Desc: "Start date filter (YYYY-MM-DD or YYYYMMDD)"},
				{Name: "--until <date>", Desc: "End date filter (YYYY-MM-DD or YYYYMMDD)"},
				{Name: "--last <N>", Desc: "Limit to last N days, weeks, or months"},
				{Name: "--order <asc|desc>", Desc: "Sort order for report rows (default: desc)"},
				{Name: "--weeks <N>", Desc: "Heatmap width in weeks (default: 52)"},
				{Name: "--start-of-week <day>", Desc: "First day of week: sunday-saturday (default: sunday)"},
			},
		},
		{
			Title: "Ranking & Analytics Options:",
			Items: []HelpItem{
				{Name: "--sort <field>", Desc: "Leaderboard: streak|tokens|cost; projects: tokens|cost|days|recent; models: tokens|cost"},
				{Name: "--top <N>", Desc: "Cap table rows to top N items (default: all)"},
				{Name: "--metric <field>", Desc: "Analytics metric: tokens or cost (default: tokens)"},
				{Name: "--by <tool|model>", Desc: "Mix breakdown dimension (default: tool)"},
				{Name: "--grain <grain>", Desc: "Mix window: day, week, or month (default: week)"},
			},
		},
		{
			Title: "Display & Formatting Options:",
			Items: []HelpItem{
				{Name: "--chart", Desc: "Charts for daily, weekly, monthly, projects, and models reports"},
				{Name: "--breakdown", Desc: "Show per-model token breakdown under each period"},
				{Name: "--nerd", Desc: "Use Nerd Font flame glyph (󰈸) in statusline"},
				{Name: "--plain", Desc: "Plain text statusline output without emoji or glyphs"},
				{Name: "--no-estimate", Desc: "Recorded cost only; disable models.dev pricing estimates"},
				{Name: "--format <text|json>", Desc: "Output format (json is an alias for --json; md/markdown retired)"},
			},
		},
		{
			Title: "Global & Engine Options:",
			Items: []HelpItem{
				{Name: "--db <path>", Desc: "Override database/data directory path"},
				{Name: "--offline", Desc: "Use cached models.dev pricing catalog without network"},
				{Name: "--fresh", Desc: "Bypass cached statusline data; live sessions always start at launch"},
				{Name: "--stream, -f", Desc: "Stream continuous NDJSON live events (live verb)"},
				{Name: "--json", Desc: "Output structured JSON (never launches a terminal UI)"},
				{Name: "--no-color", Desc: "Disable all ANSI escape sequences"},
				{Name: "--verbose, -v", Desc: "Output diagnostic warnings to stderr"},
				{Name: "-h, --help", Desc: "Show global or command-scoped help"},
			},
		},
	}

	for _, sec := range optSections {
		sb.WriteString(fmt.Sprintf("  %s\n", secTitle(sec.Title)))
		for _, item := range sec.Items {
			styled := flagName(item.Name)
			padded := padStyled(styled, len(item.Name), 22)
			avail := termWidth - 27
			wrapped := WrapText(item.Desc, avail)
			for i, line := range wrapped {
				if i == 0 {
					sb.WriteString(fmt.Sprintf("    %s %s\n", padded, line))
				} else {
					sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", 27), line))
				}
			}
		}
		sb.WriteString("\n")
	}

	// 6. Supported Tools
	sb.WriteString(fmt.Sprintf("  %s\n", secTitle("Supported Tools:")))
	twLabel := "Token Warriors:"
	twPadded := padStyled(theme.Secondary.SprintBold(colors, twLabel), len(twLabel), 18)
	twList := "OpenCode · Devin · Codex · Claude · ZCode · Grok · MiMoCode (mimo) · codewhale · DeepSeek DSH (dsh) · Nous Hermes (hermes) · CMD (command-code) · Zed"
	avail := termWidth - 23
	twWrapped := WrapText(twList, avail)
	for i, line := range twWrapped {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("    %s %s\n", twPadded, dim(line)))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", 23), dim(line)))
		}
	}

	ahLabel := "Activity Hunters:"
	ahPadded := padStyled(theme.Secondary.SprintBold(colors, ahLabel), len(ahLabel), 18)
	ahList := "Agy (agy) · Muse (muse) · Droid (droid)"
	ahWrapped := WrapText(ahList, avail)
	for i, line := range ahWrapped {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("    %s %s\n\n", ahPadded, dim(line)))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n\n", strings.Repeat(" ", 23), dim(line)))
		}
	}

	// 7. Footer
	footerText := "Documentation & guides: https://jadmadi.net/project/thermal/ · AGPL-3.0 / Commercial"
	for _, line := range WrapText(footerText, termWidth-4) {
		sb.WriteString(fmt.Sprintf("  %s\n", dim(line)))
	}
	sb.WriteString("\n")

	return sb.String()
}

// HelpFlag is one flag row in a command-scoped help page.
type HelpFlag struct {
	Name string
	Desc string
}

// CommandHelp is the content of a command-scoped help page.
type CommandHelp struct {
	Title       string
	Usage       string
	Summary     string
	Description []string
	Flags       []HelpFlag
	Examples    []string
	JSON        string
	Note        string
	Related     []string
}

// RenderCommandHelp renders a plain, width-safe command help page. It never
// reads data, writes files, opens a TUI, or touches the network.
func RenderCommandHelp(p CommandHelp) string {
	const maxWidth = 80
	var sb strings.Builder

	sb.WriteString("\n")
	if p.Summary != "" {
		sb.WriteString(fmt.Sprintf("  Thermal · %s · %s\n\n", p.Title, p.Summary))
	} else {
		sb.WriteString(fmt.Sprintf("  Thermal · %s\n\n", p.Title))
	}

	sb.WriteString("  Usage:\n")
	for _, line := range WrapText(p.Usage, maxWidth-4) {
		sb.WriteString("    " + line + "\n")
	}
	sb.WriteString("\n")

	if len(p.Description) > 0 {
		sb.WriteString("  Description:\n")
		for _, para := range p.Description {
			for _, line := range WrapText(para, maxWidth-4) {
				sb.WriteString("    " + line + "\n")
			}
		}
		sb.WriteString("\n")
	}

	if len(p.Flags) > 0 {
		sb.WriteString("  Flags:\n")
		const nameWidth = 24
		for _, f := range p.Flags {
			prefix := fmt.Sprintf("    %-*s ", nameWidth, f.Name)
			avail := maxWidth - len(prefix)
			if avail < 20 {
				avail = 20
			}
			for i, line := range WrapText(f.Desc, avail) {
				if i == 0 {
					sb.WriteString(prefix + line + "\n")
				} else {
					sb.WriteString(strings.Repeat(" ", len(prefix)) + line + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}

	if len(p.Examples) > 0 {
		sb.WriteString("  Examples:\n")
		for _, ex := range p.Examples {
			for _, line := range WrapText(ex, maxWidth-4) {
				sb.WriteString("    " + line + "\n")
			}
		}
		sb.WriteString("\n")
	}

	if p.JSON != "" {
		sb.WriteString("  JSON:\n")
		for _, line := range WrapText(p.JSON, maxWidth-4) {
			sb.WriteString("    " + line + "\n")
		}
		sb.WriteString("\n")
	}

	if p.Note != "" {
		for _, line := range WrapText("Note: "+p.Note, maxWidth-4) {
			sb.WriteString("  " + line + "\n")
		}
		sb.WriteString("\n")
	}

	if len(p.Related) > 0 {
		sb.WriteString("  See also:\n")
		for _, line := range WrapText(strings.Join(p.Related, " · "), maxWidth-4) {
			sb.WriteString("    " + line + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
