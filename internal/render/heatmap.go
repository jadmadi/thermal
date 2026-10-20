// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"time"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

func heatCell(level int, colors bool) string {
	if !colors {
		return []string{"□", "░", "▒", "▓", "█"}[level]
	}
	if level < 0 {
		level = 0
	}
	if level >= len(theme.HeatmapLevels) {
		level = len(theme.HeatmapLevels) - 1
	}
	return theme.HeatmapLevels[level].Sprint(true, "■")
}

// monthHeader places month abbreviations without overlapping them. A label
// needs three cells plus at least one separator; if the natural week slot is
// too close to the previous label, it shifts right. A label that cannot fit is
// omitted rather than truncated into a misleading fragment like "ONov".
func monthHeader(firstSunday time.Time, weeks int) string {
	if weeks <= 0 {
		return ""
	}
	line := make([]rune, weeks)
	for i := range line {
		line[i] = ' '
	}
	lastMonth := -1
	nextFree := 0
	for w := 0; w < weeks; w++ {
		date := firstSunday.AddDate(0, 0, w*7)
		m := int(date.Month()) - 1
		if m == lastMonth {
			continue
		}
		lastMonth = m
		monthAbbr := date.Format("Jan")
		start := w
		if start < nextFree {
			start = nextFree
		}
		if start+len(monthAbbr) > weeks {
			continue
		}
		for i, ch := range monthAbbr {
			line[start+i] = ch
		}
		nextFree = start + len(monthAbbr) + 1
	}
	return string(line)
}

func RenderHeatmap(activity map[string]thermal.DayActivity, weeks int, colors bool) []string {
	today := thermal.StartOfToday()
	thisSunday := today.AddDate(0, 0, -int(today.Weekday()))
	firstSunday := thisSunday.AddDate(0, 0, -(weeks-1)*7)
	thresholds := thermal.ActivityThresholds(activity)

	labels := []string{"   ", "Mon", "   ", "Wed", "   ", "Fri", "   "}
	lines := []string{"      " + monthHeader(firstSunday, weeks)}

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
