// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package theme

import (
	"os"
	"strings"
	"testing"
)

func TestTheme_Sprint_NoColor(t *testing.T) {
	text := "Thermal Telemetry"
	for _, c := range []Color{Primary, Secondary, Accent, Success, Warning, Error, Text, TextMuted, Border} {
		out := c.Sprint(false, text)
		if out != text {
			t.Fatalf("Sprint(false) = %q, want %q", out, text)
		}
		if strings.Contains(out, "\033[") {
			t.Fatalf("Sprint(false) contained ANSI escape sequence: %q", out)
		}

		outBold := c.SprintBold(false, text)
		if outBold != text {
			t.Fatalf("SprintBold(false) = %q, want %q", outBold, text)
		}
		if strings.Contains(outBold, "\033[") {
			t.Fatalf("SprintBold(false) contained ANSI escape sequence: %q", outBold)
		}
	}
}

func TestTheme_Sprint_WithColor(t *testing.T) {
	text := "Thermal Telemetry"
	out := Primary.Sprint(true, text)
	if !strings.Contains(out, "\033[") {
		t.Fatalf("Sprint(true) missing ANSI escape sequence: %q", out)
	}
	if !strings.Contains(out, text) {
		t.Fatalf("Sprint(true) missing content text: %q", out)
	}
	if !strings.HasSuffix(out, "\033[0m") {
		t.Fatalf("Sprint(true) missing reset suffix: %q", out)
	}

	outBold := Primary.SprintBold(true, text)
	if !strings.Contains(outBold, "\033[1;") {
		t.Fatalf("SprintBold(true) missing bold code: %q", outBold)
	}
}

func TestTheme_Enabled(t *testing.T) {
	// 1. Explicit CLI flag disables color
	if Enabled(true) {
		t.Fatalf("Enabled(true) should be false")
	}

	// 2. NO_COLOR env var disables color
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	os.Setenv("NO_COLOR", "1")
	if Enabled(false) {
		t.Fatalf("Enabled(false) with NO_COLOR=1 should be false")
	}

	os.Unsetenv("NO_COLOR")

	// 3. CLICOLOR=0 disables color unless CLICOLOR_FORCE=1
	origCli := os.Getenv("CLICOLOR")
	origForce := os.Getenv("CLICOLOR_FORCE")
	defer func() {
		os.Setenv("CLICOLOR", origCli)
		os.Setenv("CLICOLOR_FORCE", origForce)
	}()

	os.Setenv("CLICOLOR", "0")
	os.Unsetenv("CLICOLOR_FORCE")
	if Enabled(false) {
		t.Fatalf("Enabled(false) with CLICOLOR=0 should be false")
	}

	os.Setenv("CLICOLOR_FORCE", "1")
	if !Enabled(false) {
		t.Fatalf("Enabled(false) with CLICOLOR_FORCE=1 should be true")
	}
}

func TestTheme_HeatmapRamp(t *testing.T) {
	if len(HeatmapLevels) != 5 {
		t.Fatalf("expected 5 heatmap levels, got %d", len(HeatmapLevels))
	}
	for i, lvl := range HeatmapLevels {
		if lvl.Hex == "" || lvl.ANSI256 == "" {
			t.Fatalf("level %d missing Hex or ANSI256", i)
		}
		out := lvl.Sprint(true, "■")
		if !strings.Contains(out, "■") {
			t.Fatalf("level %d sprint output corrupted: %q", i, out)
		}
	}
}

func TestTheme_HexToRGB(t *testing.T) {
	r, g, b := hexToRGB("#FF6A00")
	if r != 255 || g != 106 || b != 0 {
		t.Fatalf("hexToRGB(#FF6A00) = (%d, %d, %d), want (255, 106, 0)", r, g, b)
	}

	r, g, b = hexToRGB("4fd6be")
	if r != 79 || g != 214 || b != 190 {
		t.Fatalf("hexToRGB(4fd6be) = (%d, %d, %d), want (79, 214, 190)", r, g, b)
	}
}
