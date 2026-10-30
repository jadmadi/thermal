// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package version

import (
	"strconv"
	"strings"
)

// These are set at build time via ldflags:
//
//	go build -ldflags "-X github.com/jadmadi/thermal/internal/version.Version=v0.2.0"
//
// GoReleaser injects them automatically via the ldflags section in
// .goreleaser.yml. When building from source without ldflags, Version
// defaults to "dev" and Commit/Date to "unknown".
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Info is the honest build metadata payload.
type Info struct {
	Version string
	Semver  string
	Commit  string
	Date    string
	Stamped bool
}

// String returns a human-readable version string.
func String() string {
	if Version == "dev" {
		return "dev (built from source; version not stamped)"
	}
	return Version
}

// Current returns the build metadata with an explicit stamped flag.
func Current() Info {
	semver := ""
	if Version != "dev" && strings.HasPrefix(Version, "v") {
		semver = Version
	}
	return Info{
		Version: String(),
		Semver:  semver,
		Commit:  Commit,
		Date:    Date,
		Stamped: Version != "dev" && Commit != "unknown" && Date != "unknown",
	}
}

// IsPast reports whether the current release has reached target. A source
// build (semver empty) is unknown and never claims a deadline has passed.
func IsPast(target string) bool {
	cur := Current().Semver
	if cur == "" || target == "" {
		return false
	}
	return compareSemver(cur, target) >= 0
}

func compareSemver(a, b string) int {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, "-", 2)
	nums := strings.Split(parts[0], ".")
	var out [3]int
	for i := 0; i < 3 && i < len(nums); i++ {
		n, _ := strconv.Atoi(nums[i])
		out[i] = n
	}
	return out
}
