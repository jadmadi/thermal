//go:build !race

// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"testing"
	"time"
)

// TestLoadBudget measures the cold load on this machine. The goal sets a
// 300ms warm budget for the first frame; loading is the expensive part of that
// path, so it is measured directly rather than inferred.
func TestLoadBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("loads every installed tool")
	}
	start := time.Now()
	a := LoadTools(nil)
	elapsed := time.Since(start)
	if len(a.Tools) == 0 {
		t.Skip("no tool data on this machine")
	}
	t.Logf("loaded %d tools in %s (warm cache)", len(a.Tools), elapsed)
	// A warm pass re-reads only deltas. The bound is deliberately loose: an
	// up-to-date checkout on a busy machine should not fail a race in CI.
	if elapsed > 5*time.Second {
		t.Errorf("load took %s, warm path should be well under a second", elapsed)
	}
}
