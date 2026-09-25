// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	mixSeriesShown  = 4
	analyticsNumber = 10
)

// RenderMix prints a period-by-series share table plus switching statistics.
func RenderMix(rep thermal.MixReport, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	gold := st.Gold

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s %s %s %s %s\n\n",
		highlight("Thermal"), dim("·"), highlight("mix"),
		dim("·"), highlight("by "+rep.By), dim("·"), highlight(rep.Metric)))

	if len(rep.Series) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No activity in the selected window.")))
		return sb.String()
	}

	shown := rep.Series
	otherLabel := ""
	var otherTotals map[string]float64
	if len(shown) > mixSeriesShown {
		shown = rep.Series[:mixSeriesShown]
		otherLabel = fmt.Sprintf("other (%d)", len(rep.Series)-mixSeriesShown)
		otherTotals = make(map[string]float64, len(rep.Buckets))
		for _, bucket := range rep.Buckets {
			var rest float64
			for i := mixSeriesShown; i < len(rep.Series); i++ {
				rest += bucketValue(bucket, i)
			}
			otherTotals[bucket.Period] = rest
		}
	}

	seriesWidth := 12
	headers := []string{"Period"}
	for _, s := range shown {
		headers = append(headers, truncate(s.Name, seriesWidth))
	}
	if otherLabel != "" {
		headers = append(headers, truncate(otherLabel, seriesWidth))
	}
	headers = append(headers, "Total")

	rule := 2 * (len(headers) - 1)
	for i := range headers {
		if i == 0 {
			rule += analyticsNumber + 2
		} else if i == len(headers)-1 {
			rule += analyticsNumber
		} else {
			rule += seriesWidth
		}
	}

	var cardLines []string

	var hsb strings.Builder
	hsb.WriteString(" ")
	for i, h := range headers {
		cell := thermal.PadRight(h, seriesWidth)
		if i == 0 {
			cell = thermal.PadRight(h, analyticsNumber+2)
		}
		if i == len(headers)-1 {
			cell = thermal.PadLeft(h, analyticsNumber)
		}
		hsb.WriteString(dim(cell))
		if i < len(headers)-1 {
			hsb.WriteString("  ")
		}
	}
	cardLines = append(cardLines, hsb.String())
	cardLines = append(cardLines, strings.Repeat("─", rule))

	formatRow := func(label string, shares []float64, total float64, style func(string) string) string {
		cells := []string{thermal.PadRight(label, analyticsNumber+2)}
		for _, share := range shares {
			cells = append(cells, thermal.PadRight(fmt.Sprintf("%.0f%%", share*100), seriesWidth))
		}
		cells = append(cells, thermal.PadLeft(formatMixValue(total, rep.Metric), analyticsNumber))
		var rsb strings.Builder
		rsb.WriteString(" ")
		for i, c := range cells {
			if style != nil {
				c = style(c)
			}
			rsb.WriteString(c)
			if i < len(cells)-1 {
				rsb.WriteString("  ")
			}
		}
		return rsb.String()
	}

	for _, bucket := range rep.Buckets {
		shares := make([]float64, 0, len(shown)+1)
		var total float64
		for _, s := range shown {
			total += bucketValue(bucket, indexOf(s.Name, rep.Series))
		}
		if otherLabel != "" {
			total += otherTotals[bucket.Period]
		}
		if total <= 0 {
			continue
		}
		for _, s := range shown {
			shares = append(shares, bucketValue(bucket, indexOf(s.Name, rep.Series))/total)
		}
		if otherLabel != "" {
			shares = append(shares, otherTotals[bucket.Period]/total)
		}
		cardLines = append(cardLines, formatRow(bucket.Period, shares, total, nil))
	}

	cardLines = append(cardLines, strings.Repeat("─", rule))
	totalShares := make([]float64, 0, len(shown)+1)
	for _, s := range shown {
		totalShares = append(totalShares, s.Share)
	}
	if otherLabel != "" {
		var rest float64
		for i := mixSeriesShown; i < len(rep.Series); i++ {
			rest += rep.Series[i].Share
		}
		totalShares = append(totalShares, rest)
	}
	cardLines = append(cardLines, formatRow("Total", totalShares, rep.Total, gold))

	title := "Tool Mix"
	if rep.By == "model" {
		title = "Model Mix"
	}

	sb.WriteString(RenderCard(CardOptions{
		Title:       title,
		RightHeader: rep.Metric + " share",
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
	}))

	if rep.By == "tool" {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim("Dominant  "), rep.Dominant))
		sb.WriteString(fmt.Sprintf("  %s %d %s\n", dim("Switches  "), rep.Switches, switchMonths(rep.SwitchesByMonth)))
		sb.WriteString(fmt.Sprintf("  %s %.2f (1 = one tool, 0 = even spread)\n", dim("Spread    "), rep.Concentration))
	}
	if rep.Estimated {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Cost mixes recorded values with estimates from models.dev list prices.")))
	}

	sb.WriteString("\n")
	return sb.String()
}

// RenderStats prints the daily distribution summary with a text histogram.
func RenderStats(rep thermal.StatsReport, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s %s %s\n\n",
		highlight("Thermal"), dim("·"), highlight("stats"), dim("·"), highlight(rep.Metric)))

	if rep.Days == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No activity in the selected window.")))
		return sb.String()
	}

	metric := func(v float64) string { return formatMixValue(v, rep.Metric) }
	rows := [][2]string{
		{"Active days", fmt.Sprintf("%d", rep.Days)},
		{"Total", metric(rep.Total)},
		{"Mean", metric(rep.Mean)},
		{"Median", metric(rep.Median)},
		{"p90", metric(rep.P90)},
		{"Max", metric(rep.Max)},
	}
	for _, row := range rows {
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight(row[0], 12)), row[1]))
	}

	maxWeekdayMean := 0.0
	for _, w := range rep.Weekday {
		if w.Mean > maxWeekdayMean {
			maxWeekdayMean = w.Mean
		}
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Weekday profile")))
	for _, w := range rep.Weekday {
		if w.Days == 0 {
			continue
		}
		bar := chartBar(w.Mean, maxWeekdayMean, 16, colors)
		daysLabel := fmt.Sprintf("%d days", w.Days)
		if w.Days == 1 {
			daysLabel = "1 day"
		}
		sb.WriteString(fmt.Sprintf("  %s %s  %s  %s\n",
			dim(thermal.PadRight(w.Weekday, 11)),
			thermal.PadLeft(metric(w.Mean), 10),
			bar,
			dim(thermal.PadLeft(daysLabel, 8)),
		))
	}

	if rep.Composition != nil && rep.Composition.Total > 0 {
		c := rep.Composition
		totalF := float64(c.Total)

		maxPart := float64(c.CacheRead)
		for _, v := range []int64{c.CacheWrite, c.UncachedInput, c.Output, c.Reasoning} {
			if float64(v) > maxPart {
				maxPart = float64(v)
			}
		}
		if maxPart <= 0 {
			maxPart = 1
		}

		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Token composition")))
		type part struct {
			label string
			val   int64
		}
		parts := []part{
			{"Cache read", c.CacheRead},
			{"Cache write", c.CacheWrite},
			{"Uncached input", c.UncachedInput},
			{"Output", c.Output},
			{"Reasoning", c.Reasoning},
		}
		for _, p := range parts {
			share := 0.0
			if totalF > 0 {
				share = float64(p.val) / totalF
			}
			bar := chartBar(float64(p.val), maxPart, 16, colors)
			shareLabel := fmt.Sprintf("%5.1f%%", share*100)
			sb.WriteString(fmt.Sprintf("  %s %s  %s  %s\n",
				dim(thermal.PadRight(p.label, 15)),
				thermal.PadLeft(metric(float64(p.val)), 10),
				bar,
				dim(shareLabel),
			))
		}
		if c.CacheRead > 0 || c.UncachedInput > 0 || c.CacheWrite > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %.1f%% of prompt tokens read from cache\n",
				dim("Cache hit rate"), c.CacheHitRate*100))
		}
	}

	if len(rep.Histogram) > 0 {
		maxCount := 0
		maxRangeW := 0
		for _, b := range rep.Histogram {
			if b.Count > maxCount {
				maxCount = b.Count
			}
			labelLen := len(fmt.Sprintf("%s — %s", metric(b.From), metric(b.To)))
			if labelLen > maxRangeW {
				maxRangeW = labelLen
			}
		}
		if maxRangeW < 16 {
			maxRangeW = 16
		}

		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Distribution")))
		for _, b := range rep.Histogram {
			bar := chartBar(float64(b.Count), float64(maxCount), 24, colors)
			rangeLabel := fmt.Sprintf("%s — %s", metric(b.From), metric(b.To))
			countLabel := fmt.Sprintf("%d", b.Count)
			sb.WriteString(fmt.Sprintf("  %s  %s  %s\n",
				dim(thermal.PadLeft(rangeLabel, maxRangeW)),
				bar,
				thermal.PadLeft(countLabel, 4),
			))
		}
	}

	topLimit := 5
	if len(rep.TopDays) < topLimit {
		topLimit = len(rep.TopDays)
	}
	displayedTopDays := rep.TopDays[:topLimit]

	allDisplayedAreOutliers := len(displayedTopDays) > 0
	for _, td := range displayedTopDays {
		if td.Value <= rep.OutlierThreshold {
			allDisplayedAreOutliers = false
			break
		}
	}

	if len(rep.TopDays) > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Top days")))
		for _, d := range displayedTopDays {
			tag := ""
			if rep.OutlierThreshold > 0 && d.Value > rep.OutlierThreshold {
				tag = fmt.Sprintf(" %s", dim("(outlier)"))
			}
			sb.WriteString(fmt.Sprintf("  %s %s%s\n", dim(thermal.PadRight(d.Day, 12)), thermal.PadLeft(metric(d.Value), 10), tag))
		}
	}

	if len(rep.Outliers) > 0 {
		header := fmt.Sprintf("Outliers above %s (median + 2 MAD)", metric(rep.OutlierThreshold))
		if allDisplayedAreOutliers {
			if len(rep.Outliers) > topLimit {
				extra := len(rep.Outliers) - topLimit
				sb.WriteString(fmt.Sprintf("\n  %s: %s\n", dim(header), fmt.Sprintf("all %d top days above, plus %d more", topLimit, extra)))
			} else if topLimit == 1 {
				sb.WriteString(fmt.Sprintf("\n  %s: %s\n", dim(header), "top day above"))
			} else {
				sb.WriteString(fmt.Sprintf("\n  %s: %s\n", dim(header), fmt.Sprintf("all %d top days above", topLimit)))
			}
		} else {
			sb.WriteString(fmt.Sprintf("\n  %s\n", dim(header)))
			for i, d := range rep.Outliers {
				if i == 5 {
					break
				}
				sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight(d.Day, 12)), thermal.PadLeft(metric(d.Value), 10)))
			}
		}
	}

	if rep.Estimated {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Cost mixes recorded values with estimates from models.dev list prices.")))
	}

	sb.WriteString("\n")
	return sb.String()
}

// RenderTrend prints the fitted slope and the month-end projection.
func RenderTrend(rep thermal.TrendReport, noColor bool) string {
	st := NewStyle(noColor)
	colors := st.Colors
	highlight := st.Highlight
	dim := st.Dim
	green := st.Success
	yellow := st.Warning
	faint := st.Faint

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s %s %s\n\n",
		highlight("Thermal"), dim("·"), highlight("trend"), dim("·"), highlight(rep.Metric)))

	if len(rep.Points) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No activity in the selected window.")))
		return sb.String()
	}

	metric := func(v float64) string { return formatMixValue(v, rep.Metric) }

	dir := trendDirection(rep)
	relPct := 0.0
	if rep.Mean > 0 {
		relPct = (rep.Slope / rep.Mean) * 100.0
	}

	var badge string
	var badgeColor func(string) string
	switch dir {
	case "rising":
		badge = fmt.Sprintf("RISING ↗ %s (%.1f%%/day)", signedValue(rep.Slope, rep.Metric), math.Abs(relPct))
		badgeColor = green
	case "falling":
		badge = fmt.Sprintf("FALLING ↘ %s (%.1f%%/day)", signedValue(rep.Slope, rep.Metric), math.Abs(relPct))
		badgeColor = yellow
	default:
		badge = "FLAT → steady state"
		badgeColor = faint
	}

	sb.WriteString(fmt.Sprintf("  Trajectory: [%s]  Window: [%d Days · %s to %s]  Daily Mean: ~%s\n\n",
		badgeColor(badge),
		len(rep.Points),
		rep.FirstDay,
		rep.LastDay,
		highlight(metric(rep.Mean)),
	))

	cardWidth := BoundedCardWidth(2)
	innerWidth := cardWidth - 4

	rows := [][2]string{
		{"Range", fmt.Sprintf("%s to %s", rep.FirstDay, rep.LastDay)},
		{"Days", fmt.Sprintf("%d", len(rep.Points))},
		{"Mean", metric(rep.Mean)},
		{"Slope", fmt.Sprintf("%s per day", signedValue(rep.Slope, rep.Metric))},
		{"Direction", dir},
	}

	var cardLines []string
	for _, row := range rows {
		cardLines = append(cardLines, fmt.Sprintf(" %s %s", dim(thermal.PadRight(row[0], 12)), row[1]))
	}

	// Sparkline if we have points
	if len(rep.Points) > 1 {
		var maxVal float64
		for _, pt := range rep.Points {
			if pt.Value > maxVal {
				maxVal = pt.Value
			}
		}
		peakLabel := fmt.Sprintf("(peak: %s)", metric(maxVal))
		availCells := innerWidth - 14 - len(peakLabel) - 2
		if availCells > 40 {
			availCells = 40
		}
		if availCells < 8 {
			availCells = 8
		}

		var vals []float64
		if len(rep.Points) <= availCells {
			vals = make([]float64, len(rep.Points))
			for i, pt := range rep.Points {
				vals[i] = pt.Value
			}
		} else {
			vals = make([]float64, availCells)
			binSize := float64(len(rep.Points)) / float64(availCells)
			for i := 0; i < availCells; i++ {
				start := int(float64(i) * binSize)
				end := int(float64(i+1) * binSize)
				if end > len(rep.Points) {
					end = len(rep.Points)
				}
				if start >= end {
					start = end - 1
				}
				sum := 0.0
				cnt := 0
				for j := start; j < end; j++ {
					sum += rep.Points[j].Value
					cnt++
				}
				if cnt > 0 {
					vals[i] = sum / float64(cnt)
				}
			}
		}

		sparkChars := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
		var sparkBuilder strings.Builder
		for _, v := range vals {
			if maxVal <= 0 || v <= 0 {
				sparkBuilder.WriteRune(' ')
				continue
			}
			idx := int((v / maxVal) * 7.0)
			if idx > 7 {
				idx = 7
			}
			if idx < 0 {
				idx = 0
			}
			sparkBuilder.WriteRune(sparkChars[idx])
		}
		cardLines = append(cardLines, fmt.Sprintf(" %s %s  %s",
			dim(thermal.PadRight("Activity", 12)),
			highlight(sparkBuilder.String()),
			dim(peakLabel),
		))
	}

	if rep.Projection != nil {
		p := rep.Projection
		cardLines = append(cardLines, strings.Repeat("─", innerWidth))
		cardLines = append(cardLines, " "+dim(fmt.Sprintf("Projection to %s (%d days)", p.To, p.Days)))
		cardLines = append(cardLines, fmt.Sprintf(" %s %s", dim(thermal.PadRight("Expected", 12)), metric(p.Expected)))
		cardLines = append(cardLines, fmt.Sprintf(" %s %s", dim(thermal.PadRight("Band", 12)),
			fmt.Sprintf("%s to %s", metric(p.Low), metric(p.High))))
	}

	sb.WriteString(RenderCard(CardOptions{
		Title:       "Activity Trend",
		RightHeader: rep.Metric,
		Lines:       cardLines,
		Indent:      2,
		Colors:      colors,
		TitleColor:  theme.Primary,
		BorderColor: theme.Border,
		MaxWidth:    cardWidth,
	}))

	// Actionable Insights section (like audit)
	sb.WriteString("\n  Actionable Insights:\n")
	if dir == "rising" {
		bullet := fmt.Sprintf("Workload velocity is accelerating by %s daily (%.1f%% growth).", signedValue(rep.Slope, rep.Metric), math.Abs(relPct))
		if rep.Projection != nil {
			bullet += fmt.Sprintf(" Projected volume through %s is ~%s.", rep.Projection.To, metric(rep.Projection.Expected))
		}
		for _, line := range WrapBullet("  • ", bullet, cardWidth) {
			sb.WriteString(line + "\n")
		}
	} else if dir == "falling" {
		bullet := fmt.Sprintf("Workload velocity is contracting by %s daily (%.1f%% decline). Activity is stabilizing.", signedValue(rep.Slope, rep.Metric), math.Abs(relPct))
		for _, line := range WrapBullet("  • ", bullet, cardWidth) {
			sb.WriteString(line + "\n")
		}
	} else {
		bullet := fmt.Sprintf("Workload is holding steady at ~%s daily with consistent utilization.", metric(rep.Mean))
		for _, line := range WrapBullet("  • ", bullet, cardWidth) {
			sb.WriteString(line + "\n")
		}
	}

	if rep.Estimated {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Cost mixes recorded values with estimates from models.dev list prices.")))
	}

	sb.WriteString("\n")
	return sb.String()
}

// bucketValue reads one series value from a bucket, tolerating a shorter value
// slice so a hand-built report cannot panic the renderer.
func bucketValue(b thermal.MixBucket, i int) float64 {
	if i < 0 || i >= len(b.Values) {
		return 0
	}
	return b.Values[i]
}

// indexOf finds a series by name in a ranked series list.
func indexOf(name string, series []thermal.MixSeries) int {
	for i, s := range series {
		if s.Name == name {
			return i
		}
	}
	return 0
}

func switchMonths(byMonth map[string]int) string {
	if len(byMonth) == 0 {
		return ""
	}
	months := make([]string, 0, len(byMonth))
	for m := range byMonth {
		months = append(months, m)
	}
	sort.Strings(months)
	parts := make([]string, 0, len(months))
	for _, m := range months {
		parts = append(parts, fmt.Sprintf("%d in %s", byMonth[m], m))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func trendDirection(rep thermal.TrendReport) string {
	if rep.Mean == 0 {
		return "flat"
	}
	relative := rep.Slope / rep.Mean
	switch {
	case relative > 0.01:
		return "rising"
	case relative < -0.01:
		return "falling"
	default:
		return "flat"
	}
}

func signedValue(v float64, metric string) string {
	sign := ""
	if v > 0 {
		sign = "+"
	}
	return sign + formatMixValue(v, metric)
}

// formatMixValue formats a mix or stats value in the report's metric unit.
func formatMixValue(v float64, metric string) string {
	if metric == "cost" {
		return formatCost(v)
	}
	return thermal.CompactNumber(int64(v))
}
