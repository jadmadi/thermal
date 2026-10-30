// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jadmadi/thermal/internal/version"
)

var errVersionRequested = errors.New("version requested")

// versionOutput returns the text or JSON version payload. It is deterministic
// apart from generatedAt and never touches data, pricing, caches, or network.
func versionOutput(asJSON bool) string {
	info := version.Current()
	if asJSON {
		b, err := marshalJSON(map[string]any{
			"type":        "version",
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
			"version":     info.Version,
			"stamped":     info.Stamped,
			"semver":      info.Semver,
			"commit":      info.Commit,
			"date":        info.Date,
		})
		if err != nil {
			return ""
		}
		return string(b) + "\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "thermal %s\n", info.Version)
	fmt.Fprintf(&b, "  commit: %s\n", info.Commit)
	fmt.Fprintf(&b, "  built:  %s\n", info.Date)
	return b.String()
}

func printVersion(asJSON bool) {
	fmt.Print(versionOutput(asJSON))
}

// resolveVersionTarget reports whether argv asks for version metadata. -V and
// --version win in any position; the first positional "version" also wins.
// -v/--verbose are deliberately not version requests.
func resolveVersionTarget(args []string) bool {
	firstPositional := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			if firstPositional == "" && i+1 < len(args) {
				firstPositional = args[i+1]
			}
			break
		}
		if a == "-V" || a == "--version" {
			return true
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			name := a
			if eq := strings.Index(a, "="); eq >= 0 {
				name = a[:eq]
			}
			key := strings.TrimLeft(name, "-")
			if _, ok := flagMetas[key]; ok && !strings.Contains(a, "=") {
				i++
			}
			continue
		}
		if firstPositional == "" {
			firstPositional = a
		}
	}
	return strings.EqualFold(firstPositional, "version")
}

// hasJSONFlag reports whether argv requests JSON, including the --format alias.
func hasJSONFlag(args []string) bool {
	for i, a := range args {
		if a == "--json" || a == "--format=json" {
			return true
		}
		if a == "--format" && i+1 < len(args) && strings.EqualFold(args[i+1], "json") {
			return true
		}
	}
	return false
}
