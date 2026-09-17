package render

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	mixSeriesShown  = 4
	analyticsNumber = 10
)

// RenderMix prints a period-by-series share table plus switching statistics.
func RenderMix(rep thermal.MixReport, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }

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

	sb.WriteString("  ")
	for i, h := range headers {
		cell := thermal.PadRight(h, seriesWidth)
		if i == 0 {
			cell = thermal.PadRight(h, analyticsNumber+2)
		}
		if i == len(headers)-1 {
			cell = thermal.PadLeft(h, analyticsNumber)
		}
		sb.WriteString(dim(cell))
		if i < len(headers)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

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
	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")

	printRow := func(label string, shares []float64, total float64, style func(string) string) {
		cells := []string{thermal.PadRight(label, analyticsNumber+2)}
		for _, share := range shares {
			cells = append(cells, thermal.PadRight(fmt.Sprintf("%.0f%%", share*100), seriesWidth))
		}
		cells = append(cells, thermal.PadLeft(formatMixValue(total, rep.Metric), analyticsNumber))
		sb.WriteString("  ")
		for i, c := range cells {
			if style != nil {
				c = style(c)
			}
			sb.WriteString(c)
			if i < len(cells)-1 {
				sb.WriteString("  ")
			}
		}
		sb.WriteString("\n")
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
		printRow(bucket.Period, shares, total, nil)
	}

	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")
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
	printRow("Total", totalShares, rep.Total, gold)

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
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }

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

	sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Weekday profile")))
	for _, w := range rep.Weekday {
		if w.Days == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("  %s %s %s\n",
			dim(thermal.PadRight(w.Weekday, 12)),
			thermal.PadLeft(metric(w.Mean), 10),
			dim(fmt.Sprintf("%d days", w.Days)),
		))
	}

	if len(rep.Histogram) > 0 {
		maxCount := 0
		for _, b := range rep.Histogram {
			if b.Count > maxCount {
				maxCount = b.Count
			}
		}
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Distribution")))
		for _, b := range rep.Histogram {
			barWidth := 0
			if maxCount > 0 {
				barWidth = b.Count * 24 / maxCount
			}
			sb.WriteString(fmt.Sprintf("  %s %s %s\n",
				dim(thermal.PadLeft(metric(b.From), 10)),
				thermal.PadRight(strings.Repeat("#", barWidth), 24),
				fmt.Sprintf("%d", b.Count),
			))
		}
	}

	if len(rep.TopDays) > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim("Top days")))
		for i, d := range rep.TopDays {
			if i == 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight(d.Day, 12)), thermal.PadLeft(metric(d.Value), 10)))
		}
	}

	if len(rep.Outliers) > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim(fmt.Sprintf("Outliers above %s (median + 2 MAD)", metric(rep.OutlierThreshold)))))
		for i, d := range rep.Outliers {
			if i == 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight(d.Day, 12)), thermal.PadLeft(metric(d.Value), 10)))
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
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s %s %s\n\n",
		highlight("Thermal"), dim("·"), highlight("trend"), dim("·"), highlight(rep.Metric)))

	if len(rep.Points) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No activity in the selected window.")))
		return sb.String()
	}

	metric := func(v float64) string { return formatMixValue(v, rep.Metric) }
	rows := [][2]string{
		{"Range", fmt.Sprintf("%s to %s", rep.FirstDay, rep.LastDay)},
		{"Days", fmt.Sprintf("%d", len(rep.Points))},
		{"Mean", metric(rep.Mean)},
		{"Slope", fmt.Sprintf("%s per day", signedValue(rep.Slope, rep.Metric))},
		{"Direction", trendDirection(rep)},
	}
	for _, row := range rows {
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight(row[0], 12)), row[1]))
	}

	if rep.Projection != nil {
		p := rep.Projection
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim(fmt.Sprintf("Projection to %s (%d days)", p.To, p.Days))))
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight("Expected", 12)), metric(p.Expected)))
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim(thermal.PadRight("Band", 12)),
			fmt.Sprintf("%s to %s", metric(p.Low), metric(p.High))))
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
