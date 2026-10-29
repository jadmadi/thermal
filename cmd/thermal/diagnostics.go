// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jadmadi/thermal/internal/loaders"
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
