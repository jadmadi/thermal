package render

import (
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

func heatCell(level int, colors bool) string {
	if !colors {
		return []string{"□", "░", "▒", "▓", "█"}[level]
	}
	palette := []string{"38;5;238", "38;5;22", "38;5;28", "38;5;34", "38;5;40"}
	return ColorCode(true, palette[level], "■")
}

func RenderHeatmap(activity map[string]thermal.DayActivity, weeks int, colors bool) []string {
	today := thermal.StartOfToday()
	thisSunday := today.AddDate(0, 0, -int(today.Weekday()))
	firstSunday := thisSunday.AddDate(0, 0, -(weeks-1)*7)
	thresholds := thermal.ActivityThresholds(activity)

	monthLine := make([]string, weeks)
	for i := range monthLine {
		monthLine[i] = " "
	}
	lastMonth := -1
	for w := 0; w < weeks; w++ {
		date := firstSunday.AddDate(0, 0, w*7)
		m := int(date.Month()) - 1
		if m != lastMonth {
			monthAbbr := date.Format("Jan")
			for i, ch := range monthAbbr {
				if w+i < weeks {
					monthLine[w+i] = string(ch)
				}
			}
			lastMonth = m
		}
	}

	labels := []string{"   ", "Mon", "   ", "Wed", "   ", "Fri", "   "}
	lines := []string{"      " + strings.Join(monthLine, "")}

	for weekday := 0; weekday < 7; weekday++ {
		row := "  " + labels[weekday] + " "
		for w := 0; w < weeks; w++ {
			date := firstSunday.AddDate(0, 0, w*7+weekday)
			if date.After(today) {
				row += " "
			} else {
				a := activity[thermal.LocalDay(date)]
				row += heatCell(thermal.ActivityLevel(a, thresholds), colors)
			}
		}
		lines = append(lines, row)
	}

	legend := ""
	for i := 0; i < 5; i++ {
		legend += heatCell(i, colors)
	}
	lines = append(lines, "      Less "+legend+" More")
	return lines
}
