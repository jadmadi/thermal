// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// IsPeriodPreset reports whether token is a natural-language period preset
// recognized by the CLI.
func IsPeriodPreset(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "today", "yesterday",
		"lastweek", "last-week", "thisweek", "this-week",
		"lastmonth", "last-month", "thismonth", "this-month",
		"lastbilling", "last-billing", "billing", "thisbilling", "this-billing",
		"annual", "year", "ytd":
		return true
	}
	return false
}

// ResolvePeriodPreset maps a natural-language period token to a start date,
// end date (YYYY-MM-DD), and human-readable window label relative to now.
func ResolvePeriodPreset(token string, now time.Time, startOfWeek time.Weekday) (since, until, label string, ok bool) {
	if now.IsZero() {
		now = time.Now()
	}
	cur := midnight(now)
	todayStr := cur.Format("2006-01-02")

	switch strings.ToLower(strings.TrimSpace(token)) {
	case "today":
		return todayStr, todayStr, fmt.Sprintf("today (%s)", todayStr), true
	case "yesterday":
		y := cur.AddDate(0, 0, -1).Format("2006-01-02")
		return y, y, fmt.Sprintf("yesterday (%s)", y), true
	case "lastweek", "last-week":
		thisWeekStart := WeekStart(cur, startOfWeek)
		prevStart := thisWeekStart.AddDate(0, 0, -7)
		prevEnd := thisWeekStart.AddDate(0, 0, -1)
		s := prevStart.Format("2006-01-02")
		u := prevEnd.Format("2006-01-02")
		return s, u, fmt.Sprintf("last week (%s to %s)", s, u), true
	case "thisweek", "this-week":
		thisWeekStart := WeekStart(cur, startOfWeek)
		s := thisWeekStart.Format("2006-01-02")
		return s, todayStr, fmt.Sprintf("this week (%s to %s)", s, todayStr), true
	case "lastmonth", "last-month":
		firstThisMonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())
		lastPrevMonth := firstThisMonth.AddDate(0, 0, -1)
		firstPrevMonth := time.Date(lastPrevMonth.Year(), lastPrevMonth.Month(), 1, 0, 0, 0, 0, cur.Location())
		s := firstPrevMonth.Format("2006-01-02")
		u := lastPrevMonth.Format("2006-01-02")
		return s, u, fmt.Sprintf("last month (%s)", firstPrevMonth.Format("2006-01")), true
	case "thismonth", "this-month":
		firstThisMonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())
		s := firstThisMonth.Format("2006-01-02")
		return s, todayStr, fmt.Sprintf("this month (%s)", firstThisMonth.Format("2006-01")), true
	case "lastbilling", "last-billing":
		firstThisMonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())
		lastPrevMonth := firstThisMonth.AddDate(0, 0, -1)
		firstPrevMonth := time.Date(lastPrevMonth.Year(), lastPrevMonth.Month(), 1, 0, 0, 0, 0, cur.Location())
		s := firstPrevMonth.Format("2006-01-02")
		u := lastPrevMonth.Format("2006-01-02")
		return s, u, fmt.Sprintf("last billing (%s to %s)", s, u), true
	case "billing", "thisbilling", "this-billing":
		firstThisMonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())
		lastThisMonth := firstThisMonth.AddDate(0, 1, -1)
		s := firstThisMonth.Format("2006-01-02")
		u := lastThisMonth.Format("2006-01-02")
		return s, u, fmt.Sprintf("billing (%s to %s)", s, u), true
	case "annual", "year", "ytd":
		firstYear := time.Date(cur.Year(), time.January, 1, 0, 0, 0, 0, cur.Location())
		s := firstYear.Format("2006-01-02")
		return s, todayStr, fmt.Sprintf("annual (%d YTD)", cur.Year()), true
	}
	return "", "", "", false
}

// ResolveProjectScope resolves a user-supplied project query against known
// project days using exact match on the project slug, basename, or display
// name. When no match is found, it returns sorted candidate suggestions.
func ResolveProjectScope(query string, projects []ProjectDay) (matchedKeys []string, displayName string, suggestions []string, ok bool) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, "", nil, false
	}
	qLower := strings.ToLower(filepath.ToSlash(q))
	hasSlash := strings.Contains(qLower, "/")

	pathTokens := make(map[string]int64)
	var allPaths []string
	for _, pd := range projects {
		if pd.Project == "" {
			continue
		}
		if _, exists := pathTokens[pd.Project]; !exists {
			allPaths = append(allPaths, pd.Project)
		}
		pathTokens[pd.Project] += pd.Tokens + int64(pd.Turns)
	}
	sort.Strings(allPaths)
	dispNames := ProjectDisplayNames(allPaths)

	for _, p := range allPaths {
		cleanLower := strings.ToLower(filepath.ToSlash(filepath.Clean(p)))
		disp := strings.ToLower(filepath.ToSlash(dispNames[p]))
		if hasSlash {
			if qLower == disp || qLower == cleanLower || strings.HasSuffix(cleanLower, "/"+qLower) {
				matchedKeys = append(matchedKeys, p)
			}
			continue
		}
		base := strings.ToLower(filepath.Base(filepath.Clean(p)))
		slug := strings.ToLower(ProjectSlug(p))
		if qLower == slug || qLower == base || qLower == disp {
			matchedKeys = append(matchedKeys, p)
		}
	}

	if len(matchedKeys) > 0 {
		if len(matchedKeys) == 1 {
			displayName = dispNames[matchedKeys[0]]
			if displayName == "" {
				displayName = ProjectSlug(matchedKeys[0])
			}
		} else {
			displayName = ProjectSlug(matchedKeys[0])
		}
		return matchedKeys, displayName, nil, true
	}

	// Build candidate list for typo/mismatch suggestions.
	slugWeights := make(map[string]int64)
	for _, p := range allPaths {
		s := ProjectSlug(p)
		if s != "" {
			slugWeights[s] += pathTokens[p]
		}
	}
	var allSlugs []string
	for s := range slugWeights {
		allSlugs = append(allSlugs, s)
	}
	sort.Slice(allSlugs, func(i, j int) bool {
		if slugWeights[allSlugs[i]] != slugWeights[allSlugs[j]] {
			return slugWeights[allSlugs[i]] > slugWeights[allSlugs[j]]
		}
		return allSlugs[i] < allSlugs[j]
	})

	// Check Levenshtein or substring matches first.
	var closeMatches []string
	for _, s := range allSlugs {
		sLower := strings.ToLower(s)
		if strings.Contains(sLower, qLower) || strings.Contains(qLower, sLower) || editDistance(qLower, sLower) <= 2 {
			closeMatches = append(closeMatches, s)
		}
	}
	if len(closeMatches) > 0 {
		if len(closeMatches) > 5 {
			closeMatches = closeMatches[:5]
		}
		return nil, "", closeMatches, false
	}

	if len(allSlugs) > 5 {
		allSlugs = allSlugs[:5]
	}
	return nil, "", allSlugs, false
}

func editDistance(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := 0; j <= len(br); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

// BuildLeaderboardResults computes ranked ToolResult entries for the
// leaderboard. When no window or project filter is set in opts, it uses each
// tool's lifetime Summary and full Daily history. When a date window or
// project scope is set, it filters rows and recomputes streaks, active days,
// token/activity volume, cost, and unpriced disclosures strictly within that
// scope.
func BuildLeaderboardResults(loaded []ToolResult, projects []ProjectDay, opts LeaderboardOptions, pricer Pricer) []ToolResult {
	hasWindow := opts.Since != "" || opts.Until != "" || opts.Last > 0
	hasProject := opts.Project != ""

	if !hasWindow && !hasProject {
		out := make([]ToolResult, 0, len(loaded))
		for _, r := range loaded {
			if !hasDailyOrProjectActivity(r.Daily) && r.Summary.LifetimeTokens == 0 {
				continue
			}
			var estCost float64
			var toolMissing []string
			var toolUnattr int64
			if pricer != nil && opts.Estimate {
				for _, d := range r.Daily {
					if d.Cost == 0 && d.Tokens > 0 && !isActivityOnly(d) {
						if len(d.Models) == 0 {
							toolUnattr += d.Tokens
						} else {
							c, missing := pricer.PriceDay(d)
							estCost += c
							for _, m := range missing {
								if !containsString(toolMissing, m) {
									toolMissing = append(toolMissing, m)
								}
							}
						}
					}
				}
			}
			activeDays := make(map[string]bool)
			for _, d := range r.Daily {
				if d.Turns > 0 {
					activeDays[d.Day] = true
				}
			}
			current, longest := ComputeStreaks(activeDays)
			cp := r
			cp.CurrentStreak = current
			cp.LongestStreak = longest
			cp.ActiveDays = len(activeDays)
			cp.TotalActivity = r.Summary.LifetimeTokens
			cp.EstimatedCost = estCost
			cp.MissingPricing = toolMissing
			cp.UnattributedTokens = toolUnattr
			out = append(out, cp)
		}
		return out
	}

	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)
	singleDayWindow := (!since.IsZero() && !until.IsZero() && since.Equal(until)) || opts.Last == 1
	recordedDays := opts.RecordedCostDays
	if recordedDays == nil {
		recordedDays = RecordedCostDaysForResults(loaded)
	}

	if hasProject {
		targetKeys := make(map[string]bool)
		for _, k := range strings.Split(opts.Project, "\x00") {
			if k != "" {
				targetKeys[k] = true
			}
		}

		// Group matching ProjectDay rows by tool name -> day (both windowed and full project history).
		byToolDay := make(map[string]map[string]*DailyRow)
		byToolAllDay := make(map[string]map[string]*DailyRow)
		for _, pd := range projects {
			if !targetKeys[pd.Project] {
				continue
			}
			t, ok := ParseDay(pd.Day)
			if !ok {
				continue
			}
			tok := pd.Tokens
			if tok == 0 && pd.Turns > 0 && (pd.Tool == "Agy" || pd.Tool == "Droid" || (pd.Tool == "Muse" && pd.Input == 0 && pd.Output == 0)) {
				tok = int64(pd.Turns)
			}

			allMap := byToolAllDay[pd.Tool]
			if allMap == nil {
				allMap = make(map[string]*DailyRow)
				byToolAllDay[pd.Tool] = allMap
			}
			allDr := allMap[pd.Day]
			if allDr == nil {
				allDr = &DailyRow{Day: pd.Day}
				allMap[pd.Day] = allDr
			}
			allDr.Tokens += tok
			allDr.Turns += pd.Turns
			allDr.Cost += pd.Cost

			if !inWindow(t, since, until, lastStart) {
				continue
			}
			toolMap := byToolDay[pd.Tool]
			if toolMap == nil {
				toolMap = make(map[string]*DailyRow)
				byToolDay[pd.Tool] = toolMap
			}
			dr := toolMap[pd.Day]
			if dr == nil {
				dr = &DailyRow{
					Day:    pd.Day,
					Models: make(map[string]ModelTokens),
				}
				toolMap[pd.Day] = dr
			}
			dr.Tokens += tok
			dr.Turns += pd.Turns
			dr.Input += pd.Input
			dr.Output += pd.Output
			dr.Reasoning += pd.Reasoning
			dr.Cache += pd.CacheRead + pd.CacheWrite
			dr.Cost += pd.Cost
			dr.LinesAdded += pd.LinesAdded
			dr.LinesDeleted += pd.LinesDeleted
			dr.FilesTouched += pd.FilesTouched
			for m, counts := range pd.Models {
				dr.Models[m] = dr.Models[m].Add(counts)
			}
		}

		var out []ToolResult
		for _, r := range loaded {
			toolMap := byToolDay[r.Name]
			if len(toolMap) == 0 {
				continue
			}
			days := make([]DailyRow, 0, len(toolMap))
			for _, dr := range toolMap {
				if len(dr.Models) == 0 {
					dr.Models = nil
				}
				days = append(days, *dr)
			}
			sort.Slice(days, func(i, j int) bool { return days[i].Day < days[j].Day })

			var histDays []DailyRow
			for _, dr := range byToolAllDay[r.Name] {
				histDays = append(histDays, *dr)
			}
			sort.Slice(histDays, func(i, j int) bool { return histDays[i].Day < histDays[j].Day })

			if res, ok := buildFilteredToolResult(r, days, histDays, singleDayWindow, opts.Estimate, pricer, recordedDays); ok {
				out = append(out, res)
			}
		}
		return out
	}

	// Date-windowed leaderboard across all projects.
	var out []ToolResult
	for _, r := range loaded {
		var days []DailyRow
		for _, d := range r.Daily {
			t, ok := ParseDay(d.Day)
			if !ok || !inWindow(t, since, until, lastStart) {
				continue
			}
			days = append(days, d)
		}
		if res, ok := buildFilteredToolResult(r, days, r.Daily, singleDayWindow, opts.Estimate, pricer, nil); ok {
			out = append(out, res)
		}
	}
	return out
}

func buildFilteredToolResult(base ToolResult, days []DailyRow, historyDays []DailyRow, singleDayWindow bool, estimate bool, pricer Pricer, recordedDays map[string]bool) (ToolResult, bool) {
	var totalTokens, inputTokens, outputTokens, reasoningTokens, cacheTokens int64
	var totalTurns int
	var totalCost, estCost float64
	var toolUnattr int64
	var toolMissing []string
	modelBreakdown := make(map[string]int64)
	activeDays := make(map[string]bool)
	var maxWindowDay string

	isActHunter := base.Tool == ToolAgy || base.Tool == ToolDroid ||
		(base.Tool == ToolMuse && base.Summary.InputTokens == 0 && base.Summary.OutputTokens == 0 && base.Summary.LifetimeTokens <= int64(base.Summary.Sessions))

	for _, d := range days {
		if d.Tokens <= 0 && d.Turns <= 0 && d.Cost <= 0 {
			continue
		}
		if !isActHunter && d.Tokens <= 0 && d.Cost <= 0 {
			continue
		}
		totalTokens += d.Tokens
		inputTokens += d.Input
		outputTokens += d.Output
		reasoningTokens += d.Reasoning
		cacheTokens += d.Cache
		totalTurns += d.Turns
		totalCost += d.Cost
		if d.Tokens > 0 || d.Cost > 0 || (isActHunter && d.Turns > 0) {
			activeDays[d.Day] = true
			if d.Day > maxWindowDay {
				maxWindowDay = d.Day
			}
		}
		for m, counts := range d.Models {
			if counts.Total() > 0 {
				modelBreakdown[m] += counts.Total()
			}
		}

		if !isActivityOnly(d) && d.Cost == 0 && d.Tokens > 0 {
			skipEst := false
			if recordedDays != nil && (recordedDays[string(base.Tool)+"|"+d.Day] || recordedDays[base.Name+"|"+d.Day]) {
				skipEst = true
			}
			if len(d.Models) == 0 {
				if !skipEst {
					toolUnattr += d.Tokens
				}
			} else if pricer != nil && estimate && !skipEst {
				c, missing := pricer.PriceDay(d)
				estCost += c
				for _, m := range missing {
					if !containsString(toolMissing, m) {
						toolMissing = append(toolMissing, m)
					}
				}
			}
		}
	}

	if len(activeDays) == 0 || (totalTokens == 0 && totalCost == 0) {
		return ToolResult{}, false
	}

	current, longest := ComputeStreaks(activeDays)
	if len(historyDays) > 0 && maxWindowDay != "" {
		histActive := make(map[string]bool)
		for _, d := range historyDays {
			if d.Day > maxWindowDay {
				continue
			}
			if d.Tokens > 0 || d.Cost > 0 || (isActHunter && d.Turns > 0) {
				histActive[d.Day] = true
			}
		}
		if len(histActive) > 0 {
			histCur, histLong := ComputeStreaks(histActive)
			current = histCur
			if singleDayWindow {
				longest = histLong
			} else if current > longest {
				longest = current
			}
		}
	}
	if len(modelBreakdown) == 0 {
		modelBreakdown = nil
	}

	summary := Summary{
		Tool:            base.Name,
		Sessions:        totalTurns,
		LifetimeTokens:  totalTokens,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		ReasoningTokens: reasoningTokens,
		CacheTokens:     cacheTokens,
		Cost:            totalCost,
		ModelBreakdown:  modelBreakdown,
	}

	return ToolResult{
		Tool:               base.Tool,
		Name:               base.Name,
		Summary:            summary,
		Daily:              days,
		CurrentStreak:      current,
		LongestStreak:      longest,
		ActiveDays:         len(activeDays),
		TotalActivity:      totalTokens,
		DataPath:           base.DataPath,
		EstimatedCost:      estCost,
		MissingPricing:     toolMissing,
		UnattributedTokens: toolUnattr,
	}, true
}

func hasDailyOrProjectActivity(days []DailyRow) bool {
	for _, d := range days {
		if d.Tokens > 0 || d.Turns > 0 {
			return true
		}
	}
	return false
}
