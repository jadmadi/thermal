// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FormatCost renders a money value with the product-wide precision policy.
// Estimated values carry the single tilde that means "includes estimated
// pricing"; recorded values never do. Zero is always $0.00.
func FormatCost(v float64, estimated bool) string {
	prefix := ""
	if estimated {
		prefix = "~"
	}
	if v == 0 {
		return prefix + "$0.00"
	}
	if v < 0.01 {
		return fmt.Sprintf("%s$%.4f", prefix, v)
	}
	return fmt.Sprintf("%s$%.2f", prefix, v)
}

// FormatCostOrDash is FormatCost with the report/table convention that an
// absent cost is an em dash rather than a fabricated zero.
func FormatCostOrDash(v float64, estimated bool) string {
	if v <= 0 {
		return "—"
	}
	return FormatCost(v, estimated)
}

// FormatTotalCost renders a report total. A total with recorded money is not
// prefixed with the estimate tilde; the footer names the recorded and
// estimated components. An all-estimated total keeps the tilde.
func FormatTotalCost(total, recorded, estimated float64) string {
	if total <= 0 {
		return "—"
	}
	return FormatCost(total, recorded <= 0 && estimated > 0)
}

// FormatPercent renders a percentage with a fixed number of decimals.
func FormatPercent(v float64, decimals int) string {
	return fmt.Sprintf("%.*f%%", decimals, v)
}

// FormatCacheHit renders the cache-hit rate with one decimal, or an em dash
// when there is no prompt volume to measure.
func FormatCacheHit(v float64) string {
	if v <= 0 {
		return "—"
	}
	return FormatPercent(v, 1)
}

// FormatReportWindow labels a report window without implying that an active
// period count is a calendar span.
func FormatReportWindow(active int, unit string, last int, since, until string) string {
	plural := unit + "s"
	switch {
	case last > 0:
		if last == 1 {
			return fmt.Sprintf("last 1 %s", unit)
		}
		return fmt.Sprintf("last %d %s", last, plural)
	case since != "" || until != "":
		if since != "" && until != "" {
			return fmt.Sprintf("%s to %s", since, until)
		}
		if since != "" {
			return "since " + since
		}
		return "through " + until
	default:
		activeUnit := plural
		if active == 1 {
			activeUnit = unit
		}
		return fmt.Sprintf("all time (%d active %s)", active, activeUnit)
	}
}

func CompactNumber(v int64) string {
	if v >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(v)/1_000_000_000)
	}
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(v)/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(v)/1_000)
	}
	return strconv.FormatInt(v, 10)
}

func PadRight(s string, width int) string {
	if n := width - utf8.RuneCountInString(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func PadLeft(s string, width int) string {
	if n := width - utf8.RuneCountInString(s); n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

func FormatPath(p string) string {
	home := HomeDir()
	if home != "" {
		if strings.HasPrefix(p, home+"/") || strings.HasPrefix(p, home+`\`) {
			return "~" + p[len(home):]
		}
	}
	return p
}
