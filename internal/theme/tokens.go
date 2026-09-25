// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package theme

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// Color defines a multi-target color representation supporting TrueColor hex,
// ANSI 256, ANSI 16, and Lip Gloss styling.
type Color struct {
	Hex     string
	ANSI256 string
	ANSI16  string
}

// Sprint formats text with this color if enabled is true.
// If enabled is false or text is empty, returns text uncolored.
func (c Color) Sprint(enabled bool, text string) string {
	if !enabled || text == "" {
		return text
	}
	if SupportsTrueColor() && c.Hex != "" {
		r, g, b := hexToRGB(c.Hex)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
	}
	if c.ANSI256 != "" {
		return fmt.Sprintf("\033[38;5;%sm%s\033[0m", c.ANSI256, text)
	}
	if c.ANSI16 != "" {
		return fmt.Sprintf("\033[%sm%s\033[0m", c.ANSI16, text)
	}
	return text
}

// SprintBold formats text bold with this color if enabled.
func (c Color) SprintBold(enabled bool, text string) string {
	if !enabled || text == "" {
		return text
	}
	if SupportsTrueColor() && c.Hex != "" {
		r, g, b := hexToRGB(c.Hex)
		return fmt.Sprintf("\033[1;38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
	}
	if c.ANSI256 != "" {
		return fmt.Sprintf("\033[1;38;5;%sm%s\033[0m", c.ANSI256, text)
	}
	if c.ANSI16 != "" {
		return fmt.Sprintf("\033[1;%sm%s\033[0m", c.ANSI16, text)
	}
	return text
}

// LipGloss returns the color as image/color.Color accepted by Lip Gloss styles.
func (c Color) LipGloss() color.Color {
	return lipgloss.Color(c.Hex)
}

// Style returns a basic Lip Gloss foreground style.
func (c Color) Style() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c.LipGloss())
}

// BoldStyle returns a bold Lip Gloss foreground style.
func (c Color) BoldStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c.LipGloss()).Bold(true)
}

// Mimocode 12-Step Luminance Scale
var (
	DarkStep1  = Color{Hex: "#0a0a0a", ANSI256: "232", ANSI16: "40"}   // Background void
	DarkStep2  = Color{Hex: "#141414", ANSI256: "233", ANSI16: "40"}   // Panel surface
	DarkStep3  = Color{Hex: "#1e1e1e", ANSI256: "234", ANSI16: "40"}   // Card element
	DarkStep4  = Color{Hex: "#282828", ANSI256: "235", ANSI16: "30"}   // Input background
	DarkStep5  = Color{Hex: "#323232", ANSI256: "236", ANSI16: "30"}   // Cold/inactive heatmap cell
	DarkStep6  = Color{Hex: "#3c3c3c", ANSI256: "237", ANSI16: "30"}   // Subtle borders
	DarkStep7  = Color{Hex: "#484848", ANSI256: "239", ANSI16: "30"}   // Standard borders
	DarkStep8  = Color{Hex: "#606060", ANSI256: "241", ANSI16: "37"}   // Active/hover borders
	DarkStep9  = Color{Hex: "#FF6A00", ANSI256: "208", ANSI16: "1;33"} // Blaze Amber (Primary)
	DarkStep10 = Color{Hex: "#FF8A3C", ANSI256: "215", ANSI16: "33"}   // Warm Amber (Secondary)
	DarkStep11 = Color{Hex: "#808080", ANSI256: "244", ANSI16: "37"}   // Muted text
	DarkStep12 = Color{Hex: "#eeeeee", ANSI256: "255", ANSI16: "1;37"} // High-contrast text
)

// Semantic Accents (Mimocode + Thermal)
var (
	Primary      = DarkStep9                                             // #FF6A00 (Blaze Amber)
	Secondary    = DarkStep10                                            // #FF8A3C (Warm Amber)
	Accent       = Color{Hex: "#818CF8", ANSI256: "105", ANSI16: "1;34"} // #818CF8 (Soft Indigo)
	Success      = Color{Hex: "#4fd6be", ANSI256: "79", ANSI16: "1;32"}  // #4fd6be (Electric Mint)
	Warning      = Color{Hex: "#FBBF24", ANSI256: "220", ANSI16: "1;33"} // #FBBF24 (Amber Gold)
	Error        = Color{Hex: "#FB7185", ANSI256: "204", ANSI16: "1;31"} // #FB7185 (Coral Rose)
	Cyan         = Color{Hex: "#38bdf8", ANSI256: "81", ANSI16: "36"}    // #38bdf8 (Slate Sky)
	Text         = DarkStep12                                            // #eeeeee (Values, primary text)
	TextMuted    = DarkStep11                                            // #808080 (Secondary labels, dates)
	Border       = DarkStep7                                             // #484848 (Box frames, dividers)
	BorderSubtle = DarkStep6                                             // #3c3c3c (Nested dividers)
	BorderActive = DarkStep9                                             // #FF6A00 (Active selection, header frame)
	Bg           = DarkStep1                                             // #0a0a0a (Dark void)
)

// HeatmapLevels defines the 5-step thermal radiation ramp:
// Level 0: Inactive / Cold (#323232)
// Level 1: Warm Ember      (#8C3A00)
// Level 2: Ignition        (#C45200)
// Level 3: Active Fire     (#FF6A00)
// Level 4: Solar White-Hot (#FF8A3C)
var HeatmapLevels = [5]Color{
	DarkStep5,
	Color{Hex: "#8C3A00", ANSI256: "130", ANSI16: "33"},
	Color{Hex: "#C45200", ANSI256: "166", ANSI16: "1;33"},
	DarkStep9,
	DarkStep10,
}

// SupportsTrueColor reports whether the current terminal supports 24-bit RGB TrueColor.
func SupportsTrueColor() bool {
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	if ct == "truecolor" || ct == "24bit" {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	return strings.Contains(term, "direct") || strings.Contains(term, "truecolor")
}

// Enabled reports whether color output should be enabled, adhering to:
// 1. Explicit CLI --no-color flag
// 2. NO_COLOR standard (https://no-color.org)
// 3. CLICOLOR and CLICOLOR_FORCE specifications (https://bixense.com/clicolors)
func Enabled(cliNoColor bool) bool {
	if cliNoColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("CLICOLOR") == "0" && os.Getenv("CLICOLOR_FORCE") != "1" {
		return false
	}
	return true
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 255, 255, 255
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 255, 255, 255
	}
	return uint8(v >> 16), uint8((v >> 8) & 0xff), uint8(v & 0xff)
}
