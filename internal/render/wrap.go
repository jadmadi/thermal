// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

const (
	// DefaultTerminalWidth is used when terminal size detection is unavailable.
	DefaultTerminalWidth = 100
	// DefaultCardWidth is the standard outer width for terminal cards.
	DefaultCardWidth = 88
	// MinCardWidth is the minimum outer width to preserve readable column spacing.
	MinCardWidth = 72
	// MaxCardWidth is the maximum outer width to prevent line-length blowouts (MD013 / terminal standard).
	MaxCardWidth = 96
)

// TerminalWidth returns the current terminal width in columns, or DefaultTerminalWidth.
func TerminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if w, _, err := term.GetSize(int(os.Stdin.Fd())); err == nil && w > 0 {
		return w
	}
	return DefaultTerminalWidth
}

// BoundedCardWidth calculates a standardized, responsive card outer width
// that honors terminal boundaries while capping line lengths to standard bounds (72-96 cols).
func BoundedCardWidth(indent int) int {
	w := TerminalWidth()
	avail := w - indent*2
	if avail < MinCardWidth {
		return MinCardWidth
	}
	if avail > MaxCardWidth {
		return MaxCardWidth
	}
	return avail
}

// WrapText wraps text into lines that do not exceed limit display cells.
// It wraps strictly at whitespace boundaries so flags like --chart or identifiers
// like sub-10ms are never broken in the middle of hyphens. It preserves ANSI escape
// sequences across wrapped lines. If any single unbroken word exceeds limit, it
// hard breaks that word to ensure the line strictly stays within limit.
func WrapText(s string, limit int) []string {
	if limit <= 0 {
		return []string{s}
	}
	if ansi.StringWidth(s) <= limit && !strings.Contains(s, "\n") {
		return []string{s}
	}

	var result []string
	paragraphs := strings.Split(s, "\n")
	for _, p := range paragraphs {
		words := strings.Fields(p)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		var curLine strings.Builder
		curWidth := 0

		for _, word := range words {
			wWidth := ansi.StringWidth(word)

			// If a single word exceeds limit, hardwrap it
			if wWidth > limit {
				if curWidth > 0 {
					result = append(result, curLine.String())
					curLine.Reset()
					curWidth = 0
				}
				for _, part := range strings.Split(ansi.Hardwrap(word, limit, true), "\n") {
					if part != "" {
						result = append(result, part)
					}
				}
				continue
			}

			if curWidth == 0 {
				curLine.WriteString(word)
				curWidth = wWidth
			} else if curWidth+1+wWidth <= limit {
				curLine.WriteString(" ")
				curLine.WriteString(word)
				curWidth += 1 + wWidth
			} else {
				result = append(result, curLine.String())
				curLine.Reset()
				curLine.WriteString(word)
				curWidth = wWidth
			}
		}
		if curWidth > 0 {
			result = append(result, curLine.String())
		}
	}
	return result
}

// WrapBullet wraps a bullet item with hanging indentation.
// The first line starts with prefix, and subsequent lines start with
// continuation spaces matching the display width of prefix.
// Optional styleFn is applied to the wrapped text of each line.
func WrapBullet(prefix, text string, limit int, styleFn ...func(string) string) []string {
	prefixWidth := ansi.StringWidth(prefix)
	avail := limit - prefixWidth
	if avail < 12 {
		avail = 12
	}

	continuationIndent := strings.Repeat(" ", prefixWidth)
	words := WrapText(text, avail)
	if len(words) == 0 {
		return []string{prefix}
	}

	fn := func(s string) string { return s }
	if len(styleFn) > 0 && styleFn[0] != nil {
		fn = styleFn[0]
	}

	var result []string
	for i, w := range words {
		if i == 0 {
			result = append(result, prefix+fn(w))
		} else {
			result = append(result, continuationIndent+fn(w))
		}
	}
	return result
}
