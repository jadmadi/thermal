package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/thermal"
)

// Layout constants. The table drops columns from the right as the terminal
// narrows, and the names truncate last, so the numbers stay readable at 80
// columns.
const (
	colRank    = 3
	colName    = 14
	colTokens  = 9
	colCost    = 10
	colStreak  = 8
	colShare   = 6
	sparkCells = 12
	minBarW    = 8
	maxBarW    = 28
	tableGap   = "  "
)

// renderOverview draws the Overview view for a window. It is pure, so the
// golden tests can call it with fixed input and compare text.
//
// height is the whole frame. Blocks are dropped from the bottom when the
// terminal cannot hold them all, which is what keeps the frame from scrolling
// the tab bar off screen.
func renderOverview(ov Overview, width, height int, p Palette) string {
	// Two lines are the tab bar and its blank, two more close the frame.
	budget := height - 4
	if budget < 1 {
		budget = 1
	}
	return fitBlocks(budget,
		titleLine(ov, width, p),
		statCards(ov, width, p),
		toolTable(ov, width, p),
		shareStrip(ov, width, p),
	)
}

func titleLine(ov Overview, width int, p Palette) string {
	left := p.Emphasis.Render("Overview")
	period := fmt.Sprintf("%s to %s", ov.Start, ov.End)
	if ov.Start == "" {
		period = "no data"
	}
	right := p.Muted.Render(fmt.Sprintf("range %s · %s", ov.Range, period))
	gap := width - lipglossWidth(left) - lipglossWidth(right) - 1
	if gap < 1 {
		return left + " " + right
	}
	return left + strings.Repeat(" ", gap) + right
}

// statCards prints the five numbers a reader checks first. Every card is
// label plus value, never a colour on its own.
func statCards(ov Overview, width int, p Palette) string {
	cards := [][2]string{
		{"tokens", thermal.CompactNumber(ov.Tokens) + " tok"},
		{"cost", Money(ov.Cost)},
		{"tools", fmt.Sprintf("%d", ov.Tools)},
		{"projects", fmt.Sprintf("%d", ov.Projects)},
		{"streak", fmt.Sprintf("%dd / best %dd", ov.Current, ov.Longest)},
	}

	// Cards sit in one row when there is room, otherwise two, otherwise one
	// per line. A narrow terminal never truncates a number.
	perRow := 5
	if width < 110 {
		perRow = 3
	}
	if width < 70 {
		perRow = 1
	}

	var lines []string
	for i := 0; i < len(cards); i += perRow {
		end := i + perRow
		if end > len(cards) {
			end = len(cards)
		}
		var cells []string
		for _, c := range cards[i:end] {
			label := p.Muted.Render(c[0])
			value := p.Emphasis.Render(c[1])
			cells = append(cells, label+"\n"+value)
		}
		colW := (width - (len(cells)-1)*2) / len(cells)
		if colW < 8 {
			colW = 8
		}
		lines = append(lines, joinColumns(cells, colW))
	}
	return strings.Join(lines, "\n\n")
}

// joinColumns lays out pre-rendered cells side by side, padding each to a fixed
// width so multi-line cells stay aligned.
func joinColumns(cells []string, colW int) string {
	split := make([][]string, len(cells))
	height := 0
	for i, c := range cells {
		split[i] = strings.Split(c, "\n")
		if len(split[i]) > height {
			height = len(split[i])
		}
	}
	var rows []string
	for r := 0; r < height; r++ {
		var parts []string
		for i := range split {
			line := ""
			if r < len(split[i]) {
				line = split[i][r]
			}
			parts = append(parts, padTo(line, colW))
		}
		rows = append(rows, strings.TrimRight(strings.Join(parts, "  "), " "))
	}
	return strings.Join(rows, "\n")
}

func toolTable(ov Overview, width int, p Palette) string {
	if len(ov.Rows) == 0 {
		return p.Muted.Render("No activity in this window.")
	}

	// Optional columns are dropped from the right until the header fits. The
	// widths above are targets, not promises: the header text is measured, so
	// the frame cannot exceed the terminal because a label was long.
	type optional struct {
		render func(barW int) string
		drop   int // cells recovered by dropping it
	}
	var opts []optional
	opts = append(opts, optional{
		render: func(int) string { return padLeft("Share", colShare) },
		drop:   colShare,
	})
	opts = append(opts, optional{
		render: func(int) string { return padRight("Last 12 weeks", sparkCells) },
		drop:   sparkCells,
	})
	opts = append(opts, optional{
		render: func(barW int) string {
			label := "Tokens"
			if ov.Metric == MetricCost {
				label = "Cost"
			}
			return padLeft(label, barW)
		},
		drop: maxBarW,
	})

	base := []string{padTo("#", colRank), padTo("Tool", colName), padLeft("Tokens", colTokens), padLeft("Cost", colCost)}
	layout := func(include int, barW int) (string, int) {
		cells := append([]string{}, base...)
		for i := 0; i < include; i++ {
			cells = append(cells, opts[i].render(barW))
		}
		cells = append(cells, padLeft("Streak", colStreak))
		return strings.Join(cells, tableGap), len(cells)
	}

	include := len(opts)
	barW := maxBarW
	header, _ := layout(include, barW)
	for lipglossWidth(header) > width && include > 0 {
		include--
		barW += opts[include].drop
		if include == 0 {
			barW = 0
		}
		header, _ = layout(include, barW)
	}
	// Whatever is left over goes to the bar, capped so the table does not
	// stretch across an ultra wide terminal.
	slack := width - lipglossWidth(header)
	if include >= 2 {
		barW += slack
		if barW > maxBarW {
			barW = maxBarW
		}
		header, _ = layout(include, barW)
	}

	var b strings.Builder
	b.WriteString(p.Muted.Render(header))
	b.WriteString("\n")
	b.WriteString(p.Dim.Render(strings.Repeat("─", minInt(width, lipglossWidth(header)))))
	b.WriteString("\n")

	maxTokens := ov.Tokens
	maxCost := ov.Cost
	for i, row := range ov.Rows {
		cells := []string{
			padTo(fmt.Sprintf("%d.", i+1), colRank),
			padTo(truncateRunes(row.Tool, colName), colName),
			padLeft(thermal.CompactNumber(row.Tokens), colTokens),
			padLeft(Money(row.Cost), colCost),
		}
		if include >= 1 {
			cells = append(cells, padLeft(fmt.Sprintf("%.1f%%", row.SharePct), colShare))
		}
		if include >= 2 {
			cells = append(cells, p.spark(row.Spark))
		}
		if include >= 3 {
			var value, max int64
			if ov.Metric == MetricCost {
				value = int64(row.Cost * 100)
				max = int64(maxCost * 100)
			} else {
				value = row.Tokens
				max = maxTokens
			}
			cells = append(cells, p.bar(value, max, barW))
		}
		cells = append(cells, padLeft(fmt.Sprintf("%dd", row.Current), colStreak))
		b.WriteString(strings.Join(cells, tableGap))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func padRight(s string, w int) string {
	n := w - len([]rune(ansi.Strip(s)))
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

// shareStrip folds the whole window into one composition bar, so the reader
// sees the split in a single glance. Segments use different runes because the
// palette has one accent: a reader with no colour still distinguishes them, and
// the legend carries the numbers.
//
// It is deliberately not a second bar chart. The table already has per-tool
// bars; this answers a different question, which is how the parts make up the
// whole.
func shareStrip(ov Overview, width int, p Palette) string {
	if len(ov.Shares) == 0 {
		return ""
	}

	// Fold everything past the legend limit into one trailing segment so the
	// bar still represents the full window.
	const legendRows = 5
	segments := ov.Shares
	hiddenPct := 0.0
	hiddenTools := 0
	if len(segments) > legendRows {
		for _, s := range segments[legendRows:] {
			hiddenPct += s.Pct
		}
		hiddenTools = len(segments) - legendRows
		segments = segments[:legendRows]
	}

	runes := []rune{'█', '▓', '▒', '░', '▚', '▪'}
	bar := stripBar(segments, hiddenPct, hiddenTools, runes, width, p)
	if bar == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(p.Muted.Render("Share of the window"))
	b.WriteString("\n\n")
	b.WriteString(bar)
	b.WriteString("\n")

	for i, s := range segments {
		if i >= len(runes) {
			break
		}
		if s.Pct <= 0 {
			continue
		}
		marker := string(runes[i])
		if p.Colorful {
			marker = p.Accent.Render(marker)
		}
		b.WriteString("  " + marker + " " +
			padTo(truncateRunes(s.Tool, 14), 16) +
			padLeft(fmt.Sprintf("%.1f%%", s.Pct), 7) + "\n")
	}
	if hiddenTools > 0 && hiddenPct > 0 {
		b.WriteString(p.Dim.Render(fmt.Sprintf("  … %d smaller tools", hiddenTools)) +
			padLeft(fmt.Sprintf("%.1f%%", hiddenPct), 8) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// stripBar builds the composition bar. Each segment gets at least one cell when
// its share rounds to zero, so a visible tool never vanishes from the picture.
func stripBar(segments []Share, hiddenPct float64, hiddenTools int, runes []rune, width int, p Palette) string {
	total := 0.0
	for _, s := range segments {
		total += s.Pct
	}
	total += hiddenPct
	if total <= 0 {
		return ""
	}

	barW := width - 2
	if barW < 20 {
		barW = 20
	}
	if barW > 72 {
		barW = 72
	}

	// Convert percentages to whole cells, giving every visible segment a cell.
	cells := make([]int, len(segments))
	used := 0
	for i, s := range segments {
		if s.Pct <= 0 {
			continue
		}
		n := int(s.Pct / total * float64(barW))
		if n == 0 {
			n = 1
		}
		cells[i] = n
		used += n
	}
	if hiddenPct > 0 {
		n := int(hiddenPct / total * float64(barW))
		if n == 0 {
			n = 1
		}
		cells = append(cells, n)
		used += n
		runes = append(runes, '▪')
		_ = hiddenTools
	}
	// Trim or grow to the bar width so the strip always matches the frame.
	for used > barW {
		longest := -1
		for i, n := range cells {
			if n > 1 && (longest == -1 || n > cells[longest]) {
				longest = i
			}
		}
		if longest == -1 {
			break
		}
		cells[longest]--
		used--
	}
	for used < barW {
		cells[0]++
		used++
	}

	var b strings.Builder
	for i, n := range cells {
		if n <= 0 {
			continue
		}
		r := runes[0]
		if i < len(runes) {
			r = runes[i]
		}
		seg := strings.Repeat(string(r), n)
		if r == '█' && i == 0 {
			if b.Len() == 0 {
				seg = p.Accent.Render(seg)
			}
		}
		b.WriteString(seg)
	}
	return b.String()
}

// --- small text helpers -----------------------------------------------------
//
// Width here means display cells for the ASCII output the dashboard produces.
// Model and tool names are ASCII, so rune count is a safe proxy; the helpers
// exist so a future unicode name has one place to fix.

func lipglossWidth(s string) int {
	return len([]rune(ansi.Strip(s)))
}

func padTo(s string, w int) string {
	n := w - len([]rune(ansi.Strip(s)))
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

func padLeft(s string, w int) string {
	n := w - len([]rune(ansi.Strip(s)))
	if n <= 0 {
		return s
	}
	return strings.Repeat(" ", n) + s
}

func truncateRunes(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w <= 1 {
		return string(r[:w])
	}
	return string(r[:w-1]) + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// fitBlocks joins rendered blocks with a blank line between them, dropping
// whole blocks from the end until the result fits the frame. Blocks carry their
// own priority: the caller orders them, so the least important is last and is
// the first to go. At least one block always prints, so a very short terminal
// shows a partial view rather than nothing.
func fitBlocks(budget int, blocks ...string) string {
	var kept []string
	used := 0
	for _, block := range blocks {
		block = strings.TrimRight(block, "\n")
		if block == "" {
			continue
		}
		h := len(strings.Split(block, "\n"))
		need := h
		if len(kept) > 0 {
			need += 2 // the blank line between blocks
		}
		if used+need > budget && len(kept) > 0 {
			break
		}
		kept = append(kept, block)
		used += need
	}
	return strings.Join(kept, "\n\n")
}
