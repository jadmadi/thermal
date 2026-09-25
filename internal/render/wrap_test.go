// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWrapText_ShortLine(t *testing.T) {
	text := "short text"
	lines := WrapText(text, 50)
	if len(lines) != 1 || lines[0] != text {
		t.Fatalf("expected 1 line %q, got %v", text, lines)
	}
}

func TestWrapText_LongLine(t *testing.T) {
	text := "Daylight Saving Time (DST) streak accuracy — replaced 24-hour time subtraction in streak logic with calendar date arithmetic to guarantee accuracy during transitions."
	limit := 50
	lines := WrapText(text, limit)
	if len(lines) <= 1 {
		t.Fatalf("expected multiple lines, got %d", len(lines))
	}
	for i, l := range lines {
		w := ansi.StringWidth(l)
		if w > limit {
			t.Errorf("line %d width %d exceeds limit %d: %q", i, w, limit, l)
		}
	}
}

func TestWrapBullet_HangingIndent(t *testing.T) {
	prefix := "  • "
	text := "Daylight Saving Time (DST) streak accuracy — replaced 24-hour time subtraction in streak logic with calendar date arithmetic to guarantee accuracy."
	limit := 60
	lines := WrapBullet(prefix, text, limit)

	if len(lines) <= 1 {
		t.Fatalf("expected multiple wrapped lines, got %d", len(lines))
	}

	if !strings.HasPrefix(lines[0], prefix) {
		t.Errorf("expected first line to have prefix %q, got %q", prefix, lines[0])
	}

	continuationIndent := "    "
	for i := 1; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], continuationIndent) {
			t.Errorf("line %d does not have continuation indent: %q", i, lines[i])
		}
		w := ansi.StringWidth(lines[i])
		if w > limit {
			t.Errorf("line %d width %d exceeds limit %d", i, w, limit)
		}
	}
}

func TestWrapBullet_WithStyleFn(t *testing.T) {
	prefix := "  • "
	text := "This is a styled bullet with multiple words wrapping."
	limit := 30
	styled := WrapBullet(prefix, text, limit, func(s string) string {
		return "[" + s + "]"
	})

	for _, l := range styled {
		if !strings.Contains(l, "[") || !strings.Contains(l, "]") {
			t.Errorf("expected styleFn applied, got line %q", l)
		}
	}
}

func TestBoundedCardWidth(t *testing.T) {
	w := BoundedCardWidth(2)
	if w < MinCardWidth || w > MaxCardWidth {
		t.Errorf("expected BoundedCardWidth between %d and %d, got %d", MinCardWidth, MaxCardWidth, w)
	}
}
