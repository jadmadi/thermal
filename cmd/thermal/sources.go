// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"
)

var errNoData = errors.New("no supported tool data found")

// runSources prints the stat-only inventory and always exits 0 when it can be
// produced, even when every source is missing.
func runSources(opts thermal.Options) {
	home := thermal.HomeDir()
	items := loaders.Inventory(home)
	if opts.JSON {
		fmt.Print(render.RenderSourcesJSON(items, home, time.Now()))
		return
	}
	fmt.Print(render.RenderSources(items, opts.NoColor))
}

// printEmptyInventory prints the inventory for an empty-data path. Human output
// goes to stderr; --json emits the sources object on stdout and still exits 1.
func printEmptyInventory(opts thermal.Options, command string) int {
	home := thermal.HomeDir()
	items := loaders.Inventory(home)
	if opts.JSON {
		fmt.Print(render.RenderSourcesJSON(items, home, time.Now()))
		return 1
	}
	fmt.Fprint(os.Stderr, render.RenderSources(items, opts.NoColor))
	if command != "" && command != "all" {
		fmt.Fprintf(os.Stderr, "\n  Run thermal %s --help for this command's flags.\n", command)
	}
	return 1
}
