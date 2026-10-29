// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

//go:build !windows

package main

import (
	"os/signal"
	"syscall"
)

// ignoreSIGPIPE turns a closed stdout consumer into an EPIPE write error so
// the NDJSON stream can exit cleanly instead of dying with SIGPIPE.
func ignoreSIGPIPE() {
	signal.Ignore(syscall.SIGPIPE)
}
