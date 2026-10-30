// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package version

import (
	"strings"
	"testing"
)

func preserve() func() {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	return func() { Version, Commit, Date = oldVersion, oldCommit, oldDate }
}

func TestCurrentUnstamped(t *testing.T) {
	defer preserve()()
	Version, Commit, Date = "dev", "unknown", "unknown"
	info := Current()
	if info.Stamped || info.Semver != "" {
		t.Fatalf("unstamped info = %+v", info)
	}
	if !strings.Contains(info.Version, "version not stamped") {
		t.Fatalf("version string = %q", info.Version)
	}
	if IsPast("v0.14.0") {
		t.Fatal("dev must not claim a deadline has passed")
	}
}

func TestCurrentStampedAndIsPast(t *testing.T) {
	defer preserve()()
	Version, Commit, Date = "v0.21.0", "abc123", "2026-10-30T00:00:00Z"
	info := Current()
	if !info.Stamped || info.Semver != "v0.21.0" || info.Commit != "abc123" {
		t.Fatalf("stamped info = %+v", info)
	}
	if !IsPast("v0.14.0") {
		t.Fatal("v0.14.0 should be past")
	}
	if IsPast("v0.22.0") {
		t.Fatal("v0.22.0 should be in the future")
	}
	if IsPast("") {
		t.Fatal("empty target must not be past")
	}
}
