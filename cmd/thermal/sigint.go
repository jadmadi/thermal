// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// isNormalLiveStop reports whether a Bubble Tea error is a user interrupt or
// cancellation rather than a renderer failure.
func isNormalLiveStop(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "program was killed") ||
		strings.Contains(msg, "program was interrupted") ||
		strings.Contains(msg, "interrupted")
}

// installLiveSignalWatchdog restores the terminal and exits 130 when a second
// signal arrives while shutdown is still in progress. The returned function
// stops the watchdog on a clean exit.
func installLiveSignalWatchdog() func() {
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
		case <-done:
			return
		}
		select {
		case <-sigCh:
			fmt.Fprint(os.Stderr, "\033[?1049l\033[?25h\033[?2004l")
			os.Exit(130)
		case <-done:
		}
	}()
	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
