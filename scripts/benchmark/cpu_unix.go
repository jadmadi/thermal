// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

//go:build unix

package main

import (
	"os"
	"syscall"
	"time"
)

// processCPU returns the child's user+system CPU time from rusage.
func processCPU(ps *os.ProcessState) (time.Duration, string, bool) {
	if ps == nil {
		return 0, "unavailable", false
	}
	r, ok := ps.SysUsage().(*syscall.Rusage)
	if !ok || r == nil {
		return 0, "unavailable", false
	}
	user := time.Duration(r.Utime.Sec)*time.Second + time.Duration(r.Utime.Usec)*time.Microsecond
	sys := time.Duration(r.Stime.Sec)*time.Second + time.Duration(r.Stime.Usec)*time.Microsecond
	return user + sys, "rusage", true
}
