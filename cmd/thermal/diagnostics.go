// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/render"
)

// printToolConfigWarning surfaces a non-fatal config-override fallback only
// with --verbose, so normal output stays quiet.
func printToolConfigWarning(info loaders.ToolInfo, verbose bool) {
	if !verbose || info.Warning == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "thermal: warning: %s: %s\n", info.Name, info.Warning)
}

// printLoadError prints the user-facing loader error and, with --verbose, the
// raw driver error behind a --db path.
func printLoadError(err error, verbose bool) {
	fmt.Fprintf(os.Stderr, "thermal: %v\n", err)
	if !verbose {
		return
	}
	var dbErr *loaders.DBPathError
	if errors.As(err, &dbErr) && dbErr.Err != nil {
		fmt.Fprintf(os.Stderr, "thermal: verbose: %v\n", dbErr.Err)
	}
}

// unknownToolError carries the unresolved token so the caller can render a
// slot-aware diagnostic instead of a generic loader error.
type unknownToolError struct{ Token string }

func (e *unknownToolError) Error() string { return "unknown tool: " + e.Token }

type tokenSlot int

const (
	slotCommand tokenSlot = iota
	slotTool
)

type tokenClass int

const (
	classKnown tokenClass = iota
	classUnknown
	classWrongSlot
	classPath
	classRetired
	classEmpty
)

type tokenDiagnostic struct {
	Token            string
	Slot             tokenSlot
	Class            tokenClass
	Suggestion       string
	SuggestIsCommand bool
	IsReport         bool
	Message          string
	ExitCode         int
}

var diagnosticCommandNames = []string{
	"daily", "weekly", "monthly", "projects", "models", "stats", "mix", "trend",
	"live", "statusline", "prompt", "status", "share", "sources", "audit", "changelog",
	"upgrade", "version", "license", "help",
}

var diagnosticToolNames = func() []string {
	names := []string{"all", "auto"}
	for _, e := range toolHelpEntries {
		names = append(names, e.Canonical)
		if e.Alias != "" {
			names = append(names, e.Alias)
		}
	}
	names = append(names, "deepseek", "nous", "ccode", "factory", "zed-editor", "antigravity")
	return names
}()

const diagnosticCommandList = "daily, weekly, monthly, projects, models, stats, mix, trend, live, statusline, share, audit, changelog, upgrade, version, license"

var diagnosticToolList = func() string {
	parts := make([]string, 0, len(toolHelpEntries))
	for _, e := range toolHelpEntries {
		parts = append(parts, e.Display)
	}
	return strings.Join(parts, ", ")
}()

func isKnownCommandToken(t string) bool {
	if _, ok := commandHelps[t]; ok {
		return true
	}
	if isReportWord(t) {
		return true
	}
	return t == "prompt" || t == "status" || t == "help"
}

func isReportCommandToken(t string) bool {
	return isReportWord(t)
}

func looksLikePath(t string) bool {
	return strings.ContainsAny(t, `/\`) || strings.HasPrefix(t, ".") || strings.HasPrefix(t, "~")
}

// classifyToken assigns a slot-aware diagnosis to an unknown token.
func classifyToken(token string, slot tokenSlot) tokenDiagnostic {
	d := tokenDiagnostic{Token: token, Slot: slot, ExitCode: 2}
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		d.Class = classEmpty
		return d
	}
	lower := strings.ToLower(trimmed)
	d.Token = lower

	if msg, ok := retiredHelp[lower]; ok {
		d.Class = classRetired
		d.Message = msg
		d.ExitCode = 1
		return d
	}
	if isKnownCommandToken(lower) {
		if slot == slotCommand {
			d.Class = classKnown
			d.ExitCode = 0
			return d
		}
		d.Class = classWrongSlot
		d.IsReport = isReportCommandToken(lower)
		return d
	}
	if _, ok := loaders.ResolveTool(lower); ok {
		if slot == slotTool {
			d.Class = classKnown
			d.ExitCode = 0
			return d
		}
		d.Class = classWrongSlot
		d.Message = "is a tool"
		return d
	}
	if looksLikePath(lower) {
		d.Class = classPath
		return d
	}

	if suggestion, ok := helpSynonyms[lower]; ok && suggestion != "" {
		d.Class = classUnknown
		d.Suggestion = suggestion
		d.SuggestIsCommand = true
		return d
	}
	// Commands win ties so a typo like "today" or "daly" is explained as a
	// command, not a tool.
	if suggestion, ok := nearestName(lower, diagnosticCommandNames); ok {
		d.Class = classUnknown
		d.Suggestion = suggestion
		d.SuggestIsCommand = true
		return d
	}
	if slot == slotTool {
		if suggestion, ok := nearestName(lower, diagnosticToolNames); ok {
			d.Class = classUnknown
			d.Suggestion = suggestion
			return d
		}
	}
	d.Class = classUnknown
	return d
}

func renderTokenDiagnostic(d tokenDiagnostic) string {
	var b strings.Builder
	switch d.Class {
	case classKnown:
		return ""
	case classEmpty:
		b.WriteString("thermal: empty argument is not a command or tool\n")
		return b.String()
	case classRetired:
		fmt.Fprintf(&b, "thermal: %s\n", d.Message)
		see := "thermal changelog"
		if d.Token == "dashboard" {
			see = "thermal live"
		}
		fmt.Fprintf(&b, "  See also: %s\n", see)
		return b.String()
	case classWrongSlot:
		kind := "command"
		if d.IsReport {
			kind = "report command"
		}
		if d.Message == "is a tool" {
			kind = "tool"
		}
		fmt.Fprintf(&b, "thermal: %q is a %s\n", d.Token, kind)
		fmt.Fprintf(&b, "  run thermal %s\n", d.Token)
		return b.String()
	case classPath:
		fmt.Fprintf(&b, "thermal: %q looks like a path\n", d.Token)
		b.WriteString("  thermal commands do not take paths; use --db <path> for a data source\n")
		return b.String()
	}

	label := "tool"
	if d.SuggestIsCommand {
		label = "command"
	}
	fmt.Fprintf(&b, "thermal: unknown %s %q\n", label, d.Token)
	if d.Suggestion != "" {
		if d.SuggestIsCommand {
			fmt.Fprintf(&b, "  did you mean: thermal %s?\n", d.Suggestion)
		} else {
			fmt.Fprintf(&b, "  did you mean: %q?\n", d.Suggestion)
		}
	}
	prefix := "Valid tools: "
	list := diagnosticToolList
	if label == "command" {
		prefix = "Valid commands: "
		list = diagnosticCommandList
	}
	for _, line := range render.WrapText(prefix+list, 76) {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	if label == "command" {
		b.WriteString("  run thermal --help for the command list\n")
	} else {
		b.WriteString("  run thermal --help for supported tools\n")
	}
	return b.String()
}

// printUnknownToken prints a slot-aware diagnostic to stderr and returns the
// process exit code for the caller.
func printUnknownToken(token string) int {
	d := classifyToken(token, slotTool)
	if d.Class == classKnown {
		return 0
	}
	fmt.Fprint(os.Stderr, renderTokenDiagnostic(d))
	return d.ExitCode
}
