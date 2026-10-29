// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

//go:build windows

package main

// ignoreSIGPIPE is a no-op on Windows; writes to a closed pipe return an
// error instead of raising a SIGPIPE signal.
func ignoreSIGPIPE() {}
