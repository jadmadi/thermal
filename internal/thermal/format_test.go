// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompactNumber(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1540, "1.5K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{44000000, "44.0M"},
		{1000000000, "1.0B"},
		{1850000000, "1.9B"},
	}
	for _, tc := range tests {
		if got := CompactNumber(tc.input); got != tc.want {
			t.Errorf("CompactNumber(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatCostContract(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		estimated bool
		want      string
	}{
		{"recorded zero", 0, false, "$0.00"},
		{"estimated zero", 0, true, "~$0.00"},
		{"recorded dollars", 12.34, false, "$12.34"},
		{"estimated dollars", 12.34, true, "~$12.34"},
		{"recorded sub-cent", 0.005, false, "$0.0050"},
		{"estimated sub-cent", 0.005, true, "~$0.0050"},
	}
	for _, tc := range tests {
		if got := FormatCost(tc.value, tc.estimated); got != tc.want {
			t.Errorf("%s: FormatCost(%v,%v) = %q, want %q", tc.name, tc.value, tc.estimated, got, tc.want)
		}
	}
	if got := FormatCostOrDash(0, false); got != "—" {
		t.Errorf("FormatCostOrDash(0) = %q, want —", got)
	}
	if got := FormatCostOrDash(0.5, true); got != "~$0.50" {
		t.Errorf("FormatCostOrDash estimated = %q", got)
	}
}

func TestFormatTotalCostContract(t *testing.T) {
	if got := FormatTotalCost(10, 4, 6); got != "$10.00" {
		t.Errorf("mixed total = %q, want $10.00 (no tilde)", got)
	}
	if got := FormatTotalCost(10, 0, 10); got != "~$10.00" {
		t.Errorf("all-estimated total = %q, want ~$10.00", got)
	}
	if got := FormatTotalCost(0, 0, 0); got != "—" {
		t.Errorf("empty total = %q, want —", got)
	}
}

func TestFormatPercentAndCacheHit(t *testing.T) {
	if got := FormatPercent(96.777, 1); got != "96.8%" {
		t.Errorf("FormatPercent = %q", got)
	}
	if got := FormatCacheHit(96.777); got != "96.8%" {
		t.Errorf("FormatCacheHit = %q", got)
	}
	if got := FormatCacheHit(0); got != "—" {
		t.Errorf("FormatCacheHit(0) = %q", got)
	}
}

func TestFormatReportWindow(t *testing.T) {
	if got := FormatReportWindow(96, "day", 0, "", ""); got != "all time (96 active days)" {
		t.Errorf("all-time window = %q", got)
	}
	if got := FormatReportWindow(1, "day", 0, "", ""); got != "all time (1 active day)" {
		t.Errorf("singular active window = %q", got)
	}
	if got := FormatReportWindow(7, "day", 7, "", ""); got != "last 7 days" {
		t.Errorf("last window = %q", got)
	}
	if got := FormatReportWindow(1, "day", 1, "", ""); got != "last 1 day" {
		t.Errorf("singular last window = %q", got)
	}
	if got := FormatReportWindow(3, "week", 0, "2026-01-01", "2026-01-31"); got != "2026-01-01 to 2026-01-31" {
		t.Errorf("date window = %q", got)
	}
}

func TestPadding(t *testing.T) {
	if got := PadRight("abc", 5); got != "abc  " {
		t.Errorf("PadRight = %q, want %q", got, "abc  ")
	}
	if got := PadRight("abcde", 3); got != "abcde" {
		t.Errorf("PadRight truncation = %q, want %q", got, "abcde")
	}
	if got := PadLeft("abc", 5); got != "  abc" {
		t.Errorf("PadLeft = %q, want %q", got, "  abc")
	}
	if got := PadLeft("—", 5); got != "    —" {
		t.Errorf("PadLeft em dash = %q, want %q", got, "    —")
	}
	if got := PadRight("—", 5); got != "—    " {
		t.Errorf("PadRight em dash = %q, want %q", got, "—    ")
	}
}

func TestFormatPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	sub := filepath.Join(home, ".local", "share", "opencode")
	got := FormatPath(sub)
	if got != "~/.local/share/opencode" && got != "~\\.local\\share\\opencode" {
		t.Errorf("FormatPath(%q) = %q", sub, got)
	}
}
