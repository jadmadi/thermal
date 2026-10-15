// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

//go:build !unix

package main

import (
	"os"
	"time"
)

// processCPU reports CPU accounting as unavailable on platforms without rusage.
func processCPU(_ *os.ProcessState) (time.Duration, string, bool) {
	return 0, "unavailable", false
}
