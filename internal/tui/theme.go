package tui

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// Palette reuses the render package's 256-colour choices so a dashboard and a
// static report read as the same product. Colours are named by role, not by
// value, because the no-colour path swaps them out.
type Palette struct {
	Accent    lipgloss.Style // titles, selected tab, bars
	Muted     lipgloss.Style // secondary labels
	Emphasis  lipgloss.Style // values that matter
	Dim       lipgloss.Style // separators and hints
	Colorful  bool
	barCells  []rune
	emptyCell rune
}

// newPalette builds the palette. When colorful is false every style is the
// identity style, so rendering never depends on colour being present and the
// golden tests can compare plain text.
func newPalette(colorful bool) Palette {
	p := Palette{
		Colorful:  colorful,
		barCells:  []rune("▁▂▃▄▅▆▇█"),
		emptyCell: '░',
	}
	if !colorful {
		return p
	}
	p.Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	p.Muted = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	p.Emphasis = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	p.Dim = lipgloss.NewStyle().Foreground(lipgloss.Color("239"))
	return p
}

// bar renders a proportional bar. The value is scaled against maxVal, so the
// largest bar fills width cells. A bar carries no meaning on its own: callers
// always print the number beside it.
func (p Palette) bar(value, maxVal int64, width int) string {
	if maxVal <= 0 || width <= 0 {
		return ""
	}
	filled := int(float64(value) / float64(maxVal) * float64(width))
	if filled > width {
		filled = width
	}
	// Half a cell is the floor: a whole cell of accent at 0.4% would overstate
	// the share, and the number is printed beside the bar anyway.
	if filled == 0 && float64(value)/float64(maxVal)*float64(width) >= 0.5 {
		filled = 1
	}
	full := strings.Repeat("█", filled)
	rest := strings.Repeat("·", width-filled)
	if p.Colorful {
		return p.Accent.Render(full) + p.Dim.Render(rest)
	}
	return full + rest
}

// spark renders a compact rune series, one rune per value, using the block
// ramp so twelve weeks fit in twelve cells.
//
// Scaling is per row and square rooted. A global scale crushes every tool
// against the largest one, and a linear scale hides the shape of a tool whose
// weeks are all small. The ramp shows each tool's own trend, so the number
// beside it stays the source of magnitude.
func (p Palette) spark(values []int64) string {
	if len(values) == 0 {
		return ""
	}
	var maxVal int64
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	var b strings.Builder
	for _, v := range values {
		if maxVal == 0 || v == 0 {
			b.WriteRune(p.emptyCell)
			continue
		}
		ratio := float64(v) / float64(maxVal)
		idx := int(math.Sqrt(ratio) * float64(len(p.barCells)-1))
		if v > 0 && idx == 0 {
			idx = 1 // any activity is visible as at least a low block
		}
		if idx < 0 {
			idx = 0
		}
		if idx >= len(p.barCells) {
			idx = len(p.barCells) - 1
		}
		b.WriteRune(p.barCells[idx])
	}
	out := b.String()
	if p.Colorful {
		return p.Accent.Render(out)
	}
	return out
}
