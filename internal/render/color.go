// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"os"

	"github.com/jadmadi/thermal/internal/theme"
)

var reset = "\033[0m"

func ColorCode(enabled bool, code string, value string) string {
	if enabled {
		return "\033[" + code + "m" + value + reset
	}
	return value
}

func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// Style encapsulates unified terminal formatting functions adhering to the
// central Mimocode theme and strict NO_COLOR standards.
type Style struct {
	Colors    bool
	Highlight func(s string) string
	Text      func(s string) string
	Muted     func(s string) string
	Dim       func(s string) string
	Primary   func(s string) string
	Secondary func(s string) string
	Accent    func(s string) string
	Success   func(s string) string
	Warning   func(s string) string
	Error     func(s string) string
	Gold      func(s string) string
	Green     func(s string) string
	Red       func(s string) string
	Cyan      func(s string) string
	Faint     func(s string) string
}

// NewStyle constructs a Style configured for the current terminal and options.
func NewStyle(noColor bool) Style {
	colors := !noColor && (IsTerminal() || os.Getenv("CLICOLOR_FORCE") == "1") && theme.Enabled(noColor)
	return Style{
		Colors:    colors,
		Highlight: func(s string) string { return theme.Text.SprintBold(colors, s) },
		Text:      func(s string) string { return theme.Text.Sprint(colors, s) },
		Muted:     func(s string) string { return theme.TextMuted.Sprint(colors, s) },
		Dim:       func(s string) string { return theme.Border.Sprint(colors, s) },
		Faint:     func(s string) string { return theme.TextMuted.Sprint(colors, s) },
		Primary:   func(s string) string { return theme.Primary.Sprint(colors, s) },
		Secondary: func(s string) string { return theme.Secondary.Sprint(colors, s) },
		Accent:    func(s string) string { return theme.Accent.Sprint(colors, s) },
		Success:   func(s string) string { return theme.Success.Sprint(colors, s) },
		Warning:   func(s string) string { return theme.Warning.Sprint(colors, s) },
		Error:     func(s string) string { return theme.Error.Sprint(colors, s) },
		Gold:      func(s string) string { return theme.Secondary.Sprint(colors, s) },
		Green:     func(s string) string { return theme.Success.Sprint(colors, s) },
		Red:       func(s string) string { return theme.Error.Sprint(colors, s) },
		Cyan:      func(s string) string { return theme.Accent.Sprint(colors, s) },
	}
}
