// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/theme"
)

// CardOptions specifies properties for rendering a rounded, themed card container.
type CardOptions struct {
	Title       string
	RightHeader string
	Lines       []string
	Width       int // Target outer width; if 0, fits content automatically.
	MaxWidth    int // Maximum outer width allowed; if 0, no max cap unless Width is set.
	Indent      int // Number of leading spaces to indent the entire card.
	Colors      bool
	TitleColor  theme.Color
	BorderColor theme.Color
}

// RenderCard wraps content lines in a rounded border with an integrated title
// on the top left and optional header metadata on the top right.
//
// Example:
// ╭─ Title ────────────────────────────────────── Header ─╮
// │ content line                                           │
// ╰────────────────────────────────────────────────────────╯
func RenderCard(opts CardOptions) string {
	titleColor := opts.TitleColor
	if titleColor.Hex == "" {
		titleColor = theme.Primary
	}
	borderColor := opts.BorderColor
	if borderColor.Hex == "" {
		borderColor = theme.Border
	}

	titlePlain := ansi.Strip(opts.Title)
	rightPlain := ansi.Strip(opts.RightHeader)
	titleWidth := ansi.StringWidth(titlePlain)
	rightWidth := ansi.StringWidth(rightPlain)

	// Determine inner width (excluding the "│ " and " │" side borders)
	maxContentWidth := 0
	for _, line := range opts.Lines {
		w := ansi.StringWidth(line)
		if w > maxContentWidth {
			maxContentWidth = w
		}
	}

	minHeaderWidth := titleWidth + rightWidth + 4
	if opts.RightHeader != "" {
		minHeaderWidth += 2
	}

	innerWidth := maxContentWidth
	if minHeaderWidth > innerWidth {
		innerWidth = minHeaderWidth
	}
	if opts.Width > 4 {
		innerWidth = opts.Width - 4
		if minHeaderWidth > innerWidth {
			innerWidth = minHeaderWidth
		}
	} else if opts.MaxWidth > 4 && innerWidth > opts.MaxWidth-4 {
		innerWidth = opts.MaxWidth - 4
		if minHeaderWidth > innerWidth {
			innerWidth = minHeaderWidth
		}
	}

	lines := prepareCardLines(opts.Lines, innerWidth)

	indent := ""
	if opts.Indent > 0 {
		indent = strings.Repeat(" ", opts.Indent)
	}

	border := func(s string) string {
		return borderColor.Sprint(opts.Colors, s)
	}
	titleStyled := titleColor.SprintBold(opts.Colors, opts.Title)
	rightStyled := theme.TextMuted.Sprint(opts.Colors, opts.RightHeader)

	var sb strings.Builder

	// Render Top Border
	if opts.RightHeader != "" {
		ruleSpaces := innerWidth - 4 - titleWidth - rightWidth
		if ruleSpaces < 1 {
			ruleSpaces = 1
		}
		sb.WriteString(indent + border("╭─ ") + titleStyled + " " + border(strings.Repeat("─", ruleSpaces)) + " " + rightStyled + border(" ─╮\n"))
	} else {
		ruleSpaces := innerWidth - 2 - titleWidth
		if ruleSpaces < 1 {
			ruleSpaces = 1
		}
		sb.WriteString(indent + border("╭─ ") + titleStyled + " " + border(strings.Repeat("─", ruleSpaces)) + border("─╮\n"))
	}

	// Render Body Lines
	for _, line := range lines {
		w := ansi.StringWidth(line)
		pad := innerWidth - w
		if pad < 0 {
			pad = 0
		}
		sb.WriteString(indent + border("│ ") + line + strings.Repeat(" ", pad) + border(" │\n"))
	}

	// Render Bottom Border
	sb.WriteString(indent + border("╰"+strings.Repeat("─", innerWidth+2)+"╯\n"))

	return sb.String()
}

// prepareCardLines ensures all lines fit within innerWidth without pushing out borders.
func prepareCardLines(lines []string, innerWidth int) []string {
	var out []string
	for _, line := range lines {
		w := ansi.StringWidth(line)
		if w <= innerWidth {
			out = append(out, line)
			continue
		}

		stripped := ansi.Strip(line)
		trimmed := strings.TrimSpace(stripped)

		// 1. Horizontal rule / divider: resize to fit innerWidth exactly
		if len(trimmed) > 0 && (strings.Trim(trimmed, "─") == "" || strings.Trim(trimmed, "-") == "") {
			leadingSpaces := len(stripped) - len(strings.TrimLeft(stripped, " "))
			ruleLen := innerWidth - leadingSpaces
			if ruleLen < 1 {
				ruleLen = 1
			}
			char := "─"
			if strings.Contains(trimmed, "-") {
				char = "-"
			}
			out = append(out, strings.Repeat(" ", leadingSpaces)+strings.Repeat(char, ruleLen))
			continue
		}

		// 2. Bullet point line
		if bulletIdx := strings.Index(line, "• "); bulletIdx != -1 {
			prefix := line[:bulletIdx+len("• ")]
			text := strings.TrimSpace(line[bulletIdx+len("• "):])
			wrapped := WrapBullet(prefix, text, innerWidth)
			out = append(out, wrapped...)
			continue
		}

		// 3. Generic text line: wrap preserving leading indent
		leadingSpaces := len(stripped) - len(strings.TrimLeft(stripped, " "))
		prefix := ""
		if leadingSpaces > 0 && leadingSpaces < len(line) {
			prefix = line[:leadingSpaces]
		}
		text := strings.TrimSpace(line)
		wrapped := WrapBullet(prefix, text, innerWidth)
		out = append(out, wrapped...)
	}
	return out
}

// CardDivider returns an internal horizontal card divider matching innerWidth.
func CardDivider(innerWidth int, colors bool, borderColor ...theme.Color) string {
	bc := theme.Border
	if len(borderColor) > 0 && borderColor[0].Hex != "" {
		bc = borderColor[0]
	}
	return bc.Sprint(colors, strings.Repeat("─", innerWidth))
}
