// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"
)

func TestTableFormatters(t *testing.T) {
	// 1. padLeft and padRight with unicode em-dash
	if got := padLeft("—", 5); got != "    —" {
		t.Errorf("padLeft('—', 5) = %q, want '    —'", got)
	}
	if got := padRight("—", 5); got != "—    " {
		t.Errorf("padRight('—', 5) = %q, want '—    '", got)
	}

	// 2. formatCostOrDash
	if got := formatCostOrDash(0); got != "—" {
		t.Errorf("formatCostOrDash(0) = %q, want '—'", got)
	}
	if got := formatCostOrDash(-5); got != "—" {
		t.Errorf("formatCostOrDash(-5) = %q, want '—'", got)
	}
	if got := formatCostOrDash(12.34); got != "$12.34" {
		t.Errorf("formatCostOrDash(12.34) = %q, want '$12.34'", got)
	}
	if got := formatCostOrDash(0.005); got != "$0.0050" {
		t.Errorf("formatCostOrDash(0.005) = %q, want '$0.0050'", got)
	}

	// 3. formatRankCell
	if got := formatRankCell(1, 3, false); got != " 1." {
		t.Errorf("formatRankCell(1, 3, false) = %q, want ' 1.'", got)
	}
	if got := formatRankCell(10, 4, false); got != " 10." {
		t.Errorf("formatRankCell(10, 4, false) = %q, want ' 10.'", got)
	}

	// 4. formatHeaderRow and tableRule
	headers := []string{"#", "Name", "Total"}
	widths := []int{3, 10, 8}
	alignRight := []bool{true, false, true}

	hdrNoColor := formatHeaderRow(headers, widths, alignRight, false)
	if !strings.Contains(hdrNoColor, "  #") || !strings.Contains(hdrNoColor, "Name      ") || !strings.Contains(hdrNoColor, "   Total") {
		t.Errorf("unexpected header output: %q", hdrNoColor)
	}

	rule := tableRule(widths)
	expectedRuleLen := widths[0] + widths[1] + widths[2] + 2*2 // 3 + 10 + 8 + 4 = 25
	if len([]rune(rule)) != expectedRuleLen {
		t.Errorf("tableRule length = %d, want %d", len([]rune(rule)), expectedRuleLen)
	}
}
