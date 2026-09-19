package tui

import (
	"fmt"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

// Column targets for the Projects table. The layout measures the rendered
// header and drops optional columns, the same way the Overview table does.
const (
	pcolRank   = 4
	pcolName   = 22
	pcolTools  = 12
	pcolTokens = 8
	pcolCost   = 10
	pcolDays   = 5
	pcolLast   = 11
)

// renderProjects draws the ranked project table, windowed so the cursor row is
// always on screen. Names come from thermal.ProjectDisplayNames, the same rule
// the static report uses, so a directory and its row label never disagree
// between the two surfaces.
//
// height is the whole frame, not the table: the fixed header and footer are
// subtracted here, because only this function knows how tall its chrome is.
func renderProjects(pd ProjectsData, width, height int, cursor int, p Palette) string {
	names := projectLabels(pd.Rows)
	var b strings.Builder

	left := p.Emphasis.Render("Projects")
	right := fmt.Sprintf("%s to %s", pd.Start, pd.End)
	if pd.Start == "" {
		right = "no data"
	}
	if pd.ToolName != "" {
		right = pd.ToolName + " · " + right
	}
	b.WriteString(spread(left, p.Muted.Render(right), width))
	b.WriteString("\n\n")

	if len(pd.Rows) == 0 {
		b.WriteString(p.Muted.Render("No project activity in this window."))
		return b.String()
	}

	type optional struct {
		render func(int) string
		drop   int
	}
	var opts []optional
	opts = append(opts, optional{render: func(int) string { return padLeft("Days", pcolDays) }, drop: pcolDays})
	opts = append(opts, optional{render: func(int) string { return padLeft("Last", pcolLast) }, drop: pcolLast})
	opts = append(opts, optional{render: func(int) string { return padTo("Tools", pcolTools) }, drop: pcolTools})
	opts = append(opts, optional{render: func(barW int) string { return padLeft("Tokens", barW) }, drop: maxBarW})

	base := []string{padTo("#", pcolRank), padTo("Project", pcolName), padLeft("Tokens", pcolTokens), padLeft("Cost", pcolCost)}
	layout := func(include, barW int) string {
		cells := append([]string{}, base...)
		for i := 0; i < include; i++ {
			cells = append(cells, opts[i].render(barW))
		}
		return strings.Join(cells, tableGap)
	}

	include := len(opts)
	barW := maxBarW
	header := layout(include, barW)
	for lipglossWidth(header) > width && include > 0 {
		include--
		barW += opts[include].drop
		if include == 0 {
			barW = 0
		}
		header = layout(include, barW)
	}
	if slack := width - lipglossWidth(header); include >= 2 && slack > 0 {
		barW += slack
		if barW > maxBarW {
			barW = maxBarW
		}
		header = layout(include, barW)
	}

	b.WriteString(p.Muted.Render(header))
	b.WriteString("\n")
	b.WriteString(p.Dim.Render(strings.Repeat("─", minInt(width, lipglossWidth(header)))))
	b.WriteString("\n")

	// Rows the frame can hold: the frame is the tab bar and a blank line, the
	// title and a blank, the header and its rule, the table, then a rule, the
	// total, the note and the hint. Anything left over is rows.
	chrome := 15
	if pd.EstCost > 0 {
		chrome++
	}
	if ModelNoteForTools(pd.Rows, pd.Tokens) != "" {
		chrome++
	}
	first, last := viewportWindow(len(pd.Rows), cursor, height-chrome)

	for i := first; i < last; i++ {
		row := pd.Rows[i]
		cells := []string{
			padTo(fmt.Sprintf("%d.", i+1), pcolRank),
			padTo(truncateRunes(labelFor(names, row.Project), pcolName), pcolName),
			padLeft(thermal.CompactNumber(row.Tokens), pcolTokens),
			padLeft(Money(row.Cost), pcolCost),
		}
		if include >= 1 {
			cells = append(cells, padLeft(fmt.Sprintf("%d", row.Days), pcolDays))
		}
		if include >= 2 {
			last := row.LastDay
			if len(last) >= 10 {
				last = last[5:] // mm-dd is enough inside a known window
			}
			cells = append(cells, padLeft(last, pcolLast))
		}
		if include >= 3 {
			cells = append(cells, padTo(truncateRunes(joinTools(row.Tools, 2), pcolTools), pcolTools))
		}
		if include >= 4 {
			// Scale against the largest row overall, not the first visible one,
			// so bars keep their meaning as the window moves.
			cells = append(cells, p.bar(row.Tokens, pd.Rows[0].Tokens, barW))
		}
		b.WriteString(strings.Join(cells, tableGap))
		b.WriteString("\n")
	}

	b.WriteString(p.Dim.Render(strings.Repeat("─", minInt(width, lipglossWidth(header)))))
	b.WriteString("\n")
	if hidden := len(pd.Rows) - (last - first); hidden > 0 {
		above, below := first, len(pd.Rows)-last
		var parts []string
		if above > 0 {
			parts = append(parts, fmt.Sprintf("%d above", above))
		}
		if below > 0 {
			parts = append(parts, fmt.Sprintf("%d below", below))
		}
		b.WriteString(p.Dim.Render(fmt.Sprintf("  rows %d-%d of %d (%s)",
			first+1, last, len(pd.Rows), strings.Join(parts, ", "))))
		b.WriteString("\n")
	}
	b.WriteString(p.Emphasis.Render(padTo("Total", pcolRank+pcolName)) +
		padLeft(thermal.CompactNumber(pd.Tokens), pcolTokens) + tableGap +
		padLeft(Money(pd.Cost), pcolCost+len(tableGap)))
	if pd.ToolName == "" && len(pd.Rows) > 0 {
		b.WriteString(p.Muted.Render(fmt.Sprintf("   %d projects · %d active days", len(pd.Rows), pd.Days)))
	}
	b.WriteString("\n")

	if note := ModelNoteForTools(pd.Rows, pd.Tokens); note != "" {
		b.WriteString(p.Muted.Render(note))
		b.WriteString("\n")
	}
	if pd.EstCost > 0 {
		b.WriteString(p.Muted.Render(fmt.Sprintf("~ %s of the total is estimated from pricing data", Money(pd.EstCost))))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// viewportWindow returns the half-open row range to draw so that cursor is
// visible. It tries to keep the cursor centred, then clamps to the ends, which
// is what makes the table feel like a list rather than a page.
func viewportWindow(total, cursor, available int) (int, int) {
	if available < 1 {
		available = 1
	}
	if total <= available {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	first := cursor - available/2
	if first < 0 {
		first = 0
	}
	if first+available > total {
		first = total - available
	}
	return first, first + available
}

// renderProjectDetail draws the drill-down: totals, tool mix, model mix, the
// weekly trend, and the project heatmap.
func renderProjectDetail(d ProjectDetail, width int, p Palette) string {
	var b strings.Builder

	b.WriteString(spread(p.Emphasis.Render("Project · "+d.Label),
		p.Muted.Render(fmt.Sprintf("%s to %s", d.Start, d.End)), width))
	b.WriteString("\n\n")

	b.WriteString(p.Muted.Render("tokens") + " " + p.Emphasis.Render(thermal.CompactNumber(d.Tokens)+" tok") +
		"   " + p.Muted.Render("cost") + " " + p.Emphasis.Render(Money(d.Cost)) +
		"   " + p.Muted.Render("days") + " " + p.Emphasis.Render(fmt.Sprintf("%d", d.Days)) +
		"   " + p.Muted.Render("turns") + " " + p.Emphasis.Render(fmt.Sprintf("%d", d.Turns)))
	b.WriteString("\n")
	if d.FirstDay != "" {
		b.WriteString(p.Muted.Render(fmt.Sprintf("first %s · last %s", d.FirstDay, d.LastDay)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	colW := (width - 2) / 2
	if colW < 24 {
		colW = 24
	}
	b.WriteString(joinColumns([]string{
		detailShareList("Tools by tokens", d.Tools, colW, p),
		detailModelList("Models by tokens", d.Models, colW, p),
	}, colW))
	b.WriteString("\n\n")

	b.WriteString(p.Muted.Render("Weekly tokens, last 12 weeks"))
	b.WriteString("\n")
	b.WriteString(p.spark(d.Weekly) + "  " + p.Dim.Render(weekLabels(d.Weeks)))
	b.WriteString("\n\n")

	b.WriteString(p.Muted.Render(fmt.Sprintf("Heatmap, %d weeks", len(d.Heatmap)/7)))
	b.WriteString("\n")
	b.WriteString(renderHeat(d.Heatmap, minInt(width, 53), p))
	b.WriteString("\n")
	b.WriteString(p.Muted.Render("less · ░ ▒ ▓ █ more"))
	if d.ModelNote != "" {
		b.WriteString("\n\n")
		b.WriteString(p.Muted.Render(d.ModelNote))
	}
	if d.EstCost > 0 {
		b.WriteString("\n")
		b.WriteString(p.Muted.Render(fmt.Sprintf("~ %s of this cost is estimated", Money(d.EstCost))))
	}
	return strings.TrimRight(b.String(), "\n")
}

func detailShareList(title string, shares []ProjectShare, width int, p Palette) string {
	var b strings.Builder
	b.WriteString(p.Muted.Render(title))
	b.WriteString("\n")
	if len(shares) == 0 {
		b.WriteString("  —\n")
		return strings.TrimRight(b.String(), "\n")
	}
	limit := 5
	if len(shares) < limit {
		limit = len(shares)
	}
	nameW := 14
	barW := width - nameW - colShare - 4
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 32 {
		barW = 32
	}
	for _, s := range shares[:limit] {
		b.WriteString("  " + padTo(truncateRunes(s.Name, nameW), nameW) + "  " +
			p.bar(s.Value, shares[0].Value, barW) + "  " +
			padLeft(fmt.Sprintf("%.1f%%", s.Pct), colShare) + "\n")
	}
	if len(shares) > limit {
		b.WriteString(p.Dim.Render(fmt.Sprintf("  … %d more", len(shares)-limit)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func detailModelList(title string, shares []ModelShare, width int, p Palette) string {
	var b strings.Builder
	b.WriteString(p.Muted.Render(title))
	b.WriteString("\n")
	if len(shares) == 0 {
		b.WriteString(p.Muted.Render("  no model attribution") + "\n")
		return strings.TrimRight(b.String(), "\n")
	}
	limit := 5
	if len(shares) < limit {
		limit = len(shares)
	}
	nameW := 18
	barW := width - nameW - colShare - 4
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 32 {
		barW = 32
	}
	converted := make([]ProjectShare, len(shares))
	for i, m := range shares {
		converted[i] = ProjectShare{Name: m.Name, Value: m.Value, Pct: m.Pct}
	}
	for _, s := range converted[:limit] {
		b.WriteString("  " + padTo(truncateRunes(s.Name, nameW), nameW) + "  " +
			p.bar(s.Value, converted[0].Value, barW) + "  " +
			padLeft(fmt.Sprintf("%.1f%%", s.Pct), colShare) + "\n")
	}
	if len(shares) > limit {
		b.WriteString(p.Dim.Render(fmt.Sprintf("  … %d more", len(shares)-limit)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// renderHeat draws the 53 week grid, seven rows tall. Cells are filled a week
// at a time, so a column is one week and a row is one weekday. Filling
// day-by-day would silently transpose the grid, which is the kind of bug a
// sparse fixture hides.
func renderHeat(days []HeatDay, width int, p Palette) string {
	if len(days) == 0 {
		return ""
	}
	const weekdays = 7
	weeks := (len(days) + weekdays - 1) / weekdays

	// Glyph order matches the terminal heatmap in internal/render: the lightest
	// shade is level 1, so the level indexes the ramp one to one.
	glyphs := []string{"░", "▒", "▓", "█"}
	grid := make([][]string, weekdays)
	for r := range grid {
		grid[r] = make([]string, weeks)
		for c := range grid[r] {
			grid[r][c] = " "
		}
	}

	for i, d := range days {
		week := i / weekdays
		row := i % weekdays
		if week >= weeks {
			break
		}
		if d.Level <= 0 {
			grid[row][week] = p.Dim.Render("·")
			continue
		}
		g := glyphs[d.Level-1]
		if d.Level > len(glyphs) {
			g = glyphs[len(glyphs)-1]
		}
		if p.Colorful {
			grid[row][week] = p.Accent.Render(g)
		} else {
			grid[row][week] = g
		}
	}

	lastWeek := weeks
	if lastWeek > width {
		lastWeek = width
	}

	var b strings.Builder
	for r := 0; r < weekdays; r++ {
		line := strings.Join(grid[r][:lastWeek], "")
		b.WriteString(strings.TrimRight(line, " "))
		if r < weekdays-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// projectLabels resolves display names for a set of rows.
func projectLabels(rows []ProjectRow) map[string]string {
	keys := make([]string, 0, len(rows))
	for _, r := range rows {
		keys = append(keys, r.Project)
	}
	return thermal.ProjectDisplayNames(keys)
}

func labelFor(names map[string]string, key string) string {
	if name, ok := names[key]; ok {
		return name
	}
	return key
}

// weekLabels compresses twelve bucket labels into something that fits under the
// sparkline without repeating the month on every cell.
func weekLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	var parts []string
	lastMonth := ""
	for _, l := range labels {
		if len(l) < 5 {
			continue
		}
		month := l[:2]
		if month != lastMonth {
			parts = append(parts, l)
			lastMonth = month
			continue
		}
		parts = append(parts, "  ")
	}
	return strings.Join(parts, " ")
}

// spread places left and right text at the two edges of a line of the given
// width, falling back to a plain join when there is no room.
func spread(left, right string, width int) string {
	gap := width - lipglossWidth(left) - lipglossWidth(right)
	if gap < 1 {
		return left + " " + right
	}
	return left + strings.Repeat(" ", gap) + right
}
