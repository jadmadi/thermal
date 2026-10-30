// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"
)

// noDataNext is the machine-readable next step shared by every empty report
// envelope. It points at the inventory and the help page, never at a report
// that would fail the same way.
type noDataNext struct {
	Inventory string `json:"inventory"`
	Help      string `json:"help"`
}

func noDataNextInfo() *noDataNext {
	return &noDataNext{
		Inventory: "thermal sources --json",
		Help:      "thermal --help",
	}
}

// markNoDataPayload adds the state and next keys to a map-based envelope. The
// requested type and all existing keys stay untouched.
func markNoDataPayload(payload map[string]interface{}) map[string]interface{} {
	payload["state"] = "no-data"
	payload["next"] = noDataNextInfo()
	return payload
}

// printNoDataTo writes the short first-run block to stdout and, under
// --verbose, appends the stat-only source inventory to stderr. It never opens
// a database or reads a transcript.
func printNoDataTo(stdout, stderr io.Writer, opts thermal.Options, command string) {
	items := loaders.Inventory(thermal.HomeDir())
	summary := thermal.SummarizeSources(items)

	fmt.Fprintln(stdout, "  No AI coding sessions found yet.")
	if summary.Empty == 1 {
		for _, item := range items {
			if item.State == thermal.SourceEmpty {
				fmt.Fprintf(stdout, "  %s is installed but has no sessions yet.\n", item.Name)
				break
			}
		}
	}
	if summary.Unreadable == 1 {
		fmt.Fprintln(stdout, "  1 source is unreadable: run thermal sources --verbose.")
	} else if summary.Unreadable > 1 {
		fmt.Fprintf(stdout, "  %d sources are unreadable: run thermal sources --verbose.\n", summary.Unreadable)
	}
	fmt.Fprintln(stdout, "  Run one prompt in OpenCode, Claude, Codex, or any supported tool, then")
	fmt.Fprintln(stdout, "  re-run thermal.")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "  Next: thermal sources   inspect all 15 source paths and states")
	fmt.Fprintln(stdout, "        thermal --help    supported tools and flags")

	if opts.Verbose {
		if command != "" && command != "all" {
			fmt.Fprintf(stderr, "  Run thermal %s --help for this command's flags.\n", command)
		}
		fmt.Fprintln(stderr)
		fmt.Fprint(stderr, render.RenderSources(items, opts.NoColor))
	}
}

func printNoData(opts thermal.Options, command string) {
	printNoDataTo(os.Stdout, os.Stderr, opts, command)
}

// requireDataExit preserves the strict exit for callers that opt into
// --require-data without changing the payload they already received.
func requireDataExit() {
	fmt.Fprintln(os.Stderr, "thermal: no data found (--require-data)")
	os.Exit(1)
}

// printShareNoData keeps share a deliberate failure without dumping the full
// source inventory into a one-command error.
func printShareNoData() {
	fmt.Fprintln(os.Stderr, "thermal: no data to share yet")
	fmt.Fprintln(os.Stderr, "  run one prompt in a supported tool, then: thermal share")
	os.Exit(1)
}

// dailyHasActivity reports whether a source has real session activity. A
// source that exists but only contains empty sessions must not become a
// zero-value leaderboard row.
func dailyHasActivity(days []thermal.DailyRow, projects []thermal.ProjectDay) bool {
	for _, day := range days {
		if day.Tokens > 0 || day.Turns > 0 {
			return true
		}
	}
	return len(projects) > 0
}

// loadUsageForReport returns noData instead of exiting so a JSON caller can
// receive the envelope it asked for. A real load error still exits 1.
func loadUsageForReport(opts thermal.Options) (usageSet, bool) {
	set, err := loadUsage(opts)
	if err == nil {
		return set, false
	}
	if errors.Is(err, errNoData) {
		return usageSet{}, true
	}
	var unknown *unknownToolError
	if errors.As(err, &unknown) {
		os.Exit(printUnknownToken(unknown.Token))
	}
	printLoadError(err, opts.Verbose)
	os.Exit(1)
	return usageSet{}, false
}
