// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import "fmt"

var resolvedWidth int

// WidthRefusal is the exact stderr message for a command below its minimum.
func WidthRefusal(command string, min, width int) string {
	return fmt.Sprintf("thermal: %s needs at least %d columns (terminal is %d)\n  widen the terminal, pipe to a file, or set --width 100\n", command, min, width)
}

// SetWidth pins the invocation's resolved text width. Zero restores automatic
// resolution (COLUMNS, then the terminal ioctl, then the default).
func SetWidth(w int) {
	if w >= 40 && w <= 240 {
		resolvedWidth = w
		return
	}
	resolvedWidth = 0
}

// CardWidthFor returns an explicit card outer width at narrow terminals and 0
// (fit-to-content) at the default 96+ layout, preserving the established look.
func CardWidthFor(width int) int {
	if width < 96 {
		return width - 2
	}
	return 0
}

// ContentWidth is the width available to renderable content. It uses the real
// terminal width (or COLUMNS), caps at 96 columns to keep lines pasteable, and
// floors at 40 so a very narrow terminal still receives a compact, unclipped
// table.
func ContentWidth() int {
	w := TerminalWidth()
	if w > MaxCardWidth {
		w = MaxCardWidth
	}
	if w < 40 {
		w = 40
	}
	return w
}
