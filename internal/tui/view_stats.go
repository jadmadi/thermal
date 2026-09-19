package tui

import (
	"fmt"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

// renderStats draws the Stats tab: a distribution, a weekday profile, the top
// days, the outliers, and the projection band.
func renderStats(sv StatsView, width int, p Palette) string {
	var b strings.Builder

	right := fmt.Sprintf("%s · %s to %s", sv.Metric, sv.Start, sv.End)
	b.WriteString(spread(p.Emphasis.Render("Stats"), p.Muted.Render(right), width))
	b.WriteString("\n\n")

	if sv.Days == 0 {
		b.WriteString(p.Muted.Render("No activity in this window."))
		return b.String()
	}

	b.WriteString(statSummary(sv, width, p))
	b.WriteString("\n\n")

	// Two blocks side by side only when both fit. At 80 columns the histogram
	// alone needs most of the frame, so they stack instead of overflowing.
	half := (width - 3) / 2
	if half < 26 {
		half = 26
	}
	hist := histogramBlock(sv, width, p)
	rhythm := rhythmBlock(sv, half, p)
	if maxLineWidth(hist)+maxLineWidth(rhythm)+2 <= width {
		b.WriteString(joinColumns([]string{hist, rhythm}, half))
	} else {
		b.WriteString(hist)
		b.WriteString("\n\n")
		b.WriteString(rhythm)
	}
	b.WriteString("\n\n")
	b.WriteString(dayLists(sv, width, p))

	if sv.Project != nil {
		b.WriteString("\n\n")
		b.WriteString(projectionBlock(sv, width, p))
	}
	if sv.Estimated {
		b.WriteString("\n\n")
		b.WriteString(p.Muted.Render("cost is estimated from pricing data for days with no recorded cost"))
	}
	return strings.TrimRight(b.String(), "\n")
}

func statSummary(sv StatsView, width int, p Palette) string {
	unit := "tokens"
	if sv.Metric == MetricCost {
		unit = "cost"
	}
	cells := [][2]string{
		{"days", fmt.Sprintf("%d", sv.Days)},
		{"mean", shortMetric(sv.Mean, sv.Metric)},
		{"median", shortMetric(sv.Median, sv.Metric)},
		{"p90", shortMetric(sv.P90, sv.Metric)},
		{"max", shortMetric(sv.Max, sv.Metric)},
	}
	perRow := 5
	if width < 100 {
		perRow = 3
	}
	var rows []string
	for i := 0; i < len(cells); i += perRow {
		end := i + perRow
		if end > len(cells) {
			end = len(cells)
		}
		var rendered []string
		for _, c := range cells[i:end] {
			rendered = append(rendered, p.Muted.Render(c[0])+"\n"+p.Emphasis.Render(c[1]))
		}
		colW := (width - (len(rendered)-1)*2) / len(rendered)
		rows = append(rows, joinColumns(rendered, colW))
	}
	out := strings.Join(rows, "\n\n")
	return out + "\n" + p.Dim.Render("daily "+unit+" for this window")
}

// histogramBlock prints the distribution as horizontal bars, because a
// histogram read in a terminal needs a scale on the left and a count on the
// right, and vertical bars cannot carry both at 80 columns.
func histogramBlock(sv StatsView, width int, p Palette) string {
	var b strings.Builder
	title := "Distribution"
	if sv.UseLog {
		title += " (log scale)"
	}
	b.WriteString(p.Muted.Render(title))
	b.WriteString("\n")
	if len(sv.Histogram) == 0 {
		b.WriteString("  —\n")
		return strings.TrimRight(b.String(), "\n")
	}

	var maxCount int
	labelW := 0
	for _, bin := range sv.Histogram {
		if bin.Count > maxCount {
			maxCount = bin.Count
		}
		if len(bin.Label) > labelW {
			labelW = len(bin.Label)
		}
	}
	barW := width - labelW - 10
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 36 {
		barW = 36
	}

	// Skip empty bins at the ends: an empty leading bin is noise, and dropping
	// it keeps the interesting shape on screen.
	from, to := 0, len(sv.Histogram)
	for from < to && sv.Histogram[from].Count == 0 {
		from++
	}
	for to > from && sv.Histogram[to-1].Count == 0 {
		to--
	}
	if from == to {
		b.WriteString("  every day fell in one bin\n")
		return strings.TrimRight(b.String(), "\n")
	}

	for _, bin := range sv.Histogram[from:to] {
		count := fmt.Sprintf("%d", bin.Count)
		b.WriteString("  " + padTo(bin.Label, labelW) + "  " +
			p.bar(int64(bin.Count), int64(maxCount), barW) + "  " +
			padLeft(count, 4) + "\n")
	}
	if sv.LogReason != "" {
		b.WriteString(p.Dim.Render("  " + sv.LogReason))
	}
	return strings.TrimRight(b.String(), "\n")
}

// rhythmBlock prints the weekday profile, which answers whether the work is a
// weekday habit or a weekend habit.
func rhythmBlock(sv StatsView, width int, p Palette) string {
	var b strings.Builder
	b.WriteString(p.Muted.Render("Weekday profile, mean per active day"))
	b.WriteString("\n")
	if len(sv.Weekday) == 0 {
		b.WriteString("  —\n")
		return strings.TrimRight(b.String(), "\n")
	}
	var max float64
	for _, wd := range sv.Weekday {
		if wd.Mean > max {
			max = wd.Mean
		}
	}
	labelW := 4
	barW := width - labelW - 14
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 24 {
		barW = 24
	}
	for _, wd := range sv.Weekday {
		label := wd.Weekday
		if len(label) > 3 {
			label = label[:3]
		}
		b.WriteString("  " + padTo(label, labelW) + "  " +
			p.bar(int64(wd.Mean), int64(max), barW) + "  " +
			padLeft(shortMetric(wd.Mean, sv.Metric), 7) + "  " +
			p.Dim.Render(fmt.Sprintf("%dd", wd.Days)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// dayLists prints the top days and the outliers, side by side when they fit.
func dayLists(sv StatsView, width int, p Palette) string {
	half := (width - 3) / 2
	if half < 26 {
		half = 26
	}
	top := dayList("Top days", sv.TopDays, sv.Metric, half, p)
	out := dayList(outlierTitle(sv), sv.Outliers, sv.Metric, half, p)
	if maxLineWidth(top)+maxLineWidth(out)+2 <= width {
		return joinColumns([]string{top, out}, half)
	}
	return top + "\n\n" + out
}

// maxLineWidth reports the widest line in a block, which is what decides
// whether two blocks can share a row.
func maxLineWidth(block string) int {
	max := 0
	for _, line := range strings.Split(block, "\n") {
		if n := lipglossWidth(line); n > max {
			max = n
		}
	}
	return max
}

func outlierTitle(sv StatsView) string {
	if sv.Threshold <= 0 {
		return "Outliers, median plus two MAD"
	}
	return fmt.Sprintf("Outliers above %s", shortMetric(sv.Threshold, sv.Metric))
}

func dayList(title string, days []thermal.DayValue, m Metric, width int, p Palette) string {
	var b strings.Builder
	b.WriteString(p.Muted.Render(title))
	b.WriteString("\n")
	if len(days) == 0 {
		b.WriteString("  none\n")
		return strings.TrimRight(b.String(), "\n")
	}
	limit := 6
	if len(days) < limit {
		limit = len(days)
	}
	var max float64
	for _, d := range days[:limit] {
		if d.Value > max {
			max = d.Value
		}
	}
	barW := width - 18
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 20 {
		barW = 20
	}
	for _, d := range days[:limit] {
		b.WriteString("  " + padTo(d.Day, 11) + " " +
			p.bar(int64(d.Value), int64(max), barW) + "  " +
			padLeft(shortMetric(d.Value, m), 8) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// projectionBlock draws the month-end band. It is a range, never a single
// number, because a least squares fit through noisy usage days does not
// support a confident point.
func projectionBlock(sv StatsView, width int, p Palette) string {
	pr := sv.Project
	var b strings.Builder
	b.WriteString(p.Muted.Render(fmt.Sprintf("Projection to %s, from the least squares fit over the window", pr.To)))
	b.WriteString("\n")

	labelW := 12
	barW := width - labelW - 22
	if barW < minBarW {
		barW = minBarW
	}
	if barW > 40 {
		barW = 40
	}
	scale := pr.High
	if scale <= 0 {
		scale = pr.Expected
	}
	if scale <= 0 {
		b.WriteString("  " + p.Muted.Render("not enough days to project") + "\n")
		return strings.TrimRight(b.String(), "\n")
	}

	low := pr.Low
	if low < 0 {
		low = 0
	}
	b.WriteString("  " + padTo("expected", labelW) +
		p.bar(int64(pr.Expected), int64(scale), barW) + "  " +
		padLeft(shortMetric(pr.Expected, sv.Metric), 9) + "\n")
	// The band is a range, so it is printed as text rather than drawn as a bar:
	// a bar would imply a point, and the whole point of the band is that the
	// fit does not support one. The outline shows where the range sits against
	// the scale.
	lo := int64(low / scale * float64(barW))
	hi := int64(pr.High / scale * float64(barW))
	if hi <= lo {
		hi = lo + 1
	}
	bandBar := strings.Repeat(" ", int(lo)) +
		p.Dim.Render(strings.Repeat("━", int(hi-lo))) +
		strings.Repeat(" ", barW-int(hi))
	b.WriteString("  " + padTo("band", labelW) + bandBar + "  " +
		padLeft(shortMetric(low, sv.Metric)+" to "+shortMetric(pr.High, sv.Metric), 9) + "\n")
	b.WriteString("  " + padTo("slope", labelW) +
		p.Muted.Render(fmt.Sprintf("%+.3g per day over %d days left", sv.Slope, pr.Days)) + "\n")
	return strings.TrimRight(b.String(), "\n")
}
