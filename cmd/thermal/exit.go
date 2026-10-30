// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"io"
)

const (
	ExitOK      = 0
	ExitFailure = 1
	ExitUsage   = 2
	ExitRefused = 3
)

// printExitCodes writes the stable process contract shared by help pages.
func printExitCodes(w io.Writer) {
	fmt.Fprintln(w, "Exit codes: 0 ok · 1 data/IO error · 2 usage error · 3 refused")
}
