// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/theme"
)

func TestRenderCard_Alignment(t *testing.T) {
	lines := []string{
		"Line 1 with text",
		"Line 2 is longer than the first one",
		"Short",
	}

	card := RenderCard(CardOptions{
		Title:       "Test Card",
		RightHeader: "metadata info",
		Lines:       lines,
		Colors:      false,
	})

	cardLines := strings.Split(strings.TrimSuffix(card, "\n"), "\n")
	if len(cardLines) != len(lines)+2 {
		t.Fatalf("expected %d card lines, got %d", len(lines)+2, len(cardLines))
	}

	expectedWidth := ansi.StringWidth(cardLines[0])
	for i, l := range cardLines {
		w := ansi.StringWidth(l)
		if w != expectedWidth {
			t.Errorf("line %d width = %d, expected %d. Line:\n%q", i, w, expectedWidth, l)
		}
	}
}

func TestRenderCard_NoRightHeader(t *testing.T) {
	lines := []string{"Content A", "Content B"}
	card := RenderCard(CardOptions{
		Title:  "Header Only",
		Lines:  lines,
		Colors: false,
	})

	cardLines := strings.Split(strings.TrimSuffix(card, "\n"), "\n")
	expectedWidth := ansi.StringWidth(cardLines[0])
	for i, l := range cardLines {
		w := ansi.StringWidth(l)
		if w != expectedWidth {
			t.Errorf("line %d width = %d, expected %d. Line:\n%q", i, w, expectedWidth, l)
		}
	}
}

func TestRenderCard_NoColor(t *testing.T) {
	lines := []string{"Plain text"}
	card := RenderCard(CardOptions{
		Title:       "Title",
		RightHeader: "Right",
		Lines:       lines,
		Colors:      false,
	})

	if strings.Contains(card, "\033[") {
		t.Errorf("expected no ANSI codes when Colors is false, got:\n%s", card)
	}
}

func TestRenderCard_WithColors(t *testing.T) {
	lines := []string{"Colored text"}
	card := RenderCard(CardOptions{
		Title:       "Title",
		RightHeader: "Right",
		Lines:       lines,
		Colors:      true,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	})

	if !strings.Contains(card, "\033[") {
		t.Errorf("expected ANSI escape codes when Colors is true, got:\n%s", card)
	}
}

func TestRenderCard_WidthEnforcementWithLongBullet(t *testing.T) {
	lines := []string{
		"  • This is a very long bullet point that would previously blow out the card width horizontally across the terminal and destroy all formatting.",
	}
	card := RenderCard(CardOptions{
		Title:       "Wrapped Card",
		RightHeader: "v1.0.0",
		Lines:       lines,
		Width:       60,
		Colors:      false,
	})

	cardLines := strings.Split(strings.TrimSuffix(card, "\n"), "\n")
	if len(cardLines) <= 3 {
		t.Fatalf("expected bullet to wrap into multiple lines, got %d card lines", len(cardLines))
	}

	expectedWidth := 60
	for i, l := range cardLines {
		w := ansi.StringWidth(l)
		if w != expectedWidth {
			t.Errorf("line %d width = %d, expected %d. Line:\n%q", i, w, expectedWidth, l)
		}
	}
}

func TestRenderCard_DividerClamping(t *testing.T) {
	lines := []string{
		"Short line",
		strings.Repeat("─", 150),
		"Another short line",
	}
	card := RenderCard(CardOptions{
		Title:  "Divider Test",
		Lines:  lines,
		Width:  50,
		Colors: false,
	})

	cardLines := strings.Split(strings.TrimSuffix(card, "\n"), "\n")
	expectedWidth := 50
	for i, l := range cardLines {
		w := ansi.StringWidth(l)
		if w != expectedWidth {
			t.Errorf("line %d width = %d, expected %d. Line:\n%q", i, w, expectedWidth, l)
		}
	}
}
