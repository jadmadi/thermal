// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// BuildWorkloadSnapshot aggregates historical token telemetry over the
// requested window for replay simulation.
func BuildWorkloadSnapshot(days []DailyRow, opts ReplayOptions, pricer Pricer) (WorkloadSnapshot, map[string]DailyRow, int) {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)

	byDay := make(map[string]DailyRow)
	var (
		uncachedInput      int64
		output             int64
		reasoning          int64
		cacheRead          int64
		cacheWrite         int64
		unclassifiedTokens int64
		totalTokens        int64
		actualSpend        float64
		isEstimated        bool
		minDay             string
		maxDay             string
	)

	for _, day := range days {
		t, ok := ParseDay(day.Day)
		if !ok || !inWindow(t, since, until, lastStart) {
			continue
		}
		if isActivityOnly(day) {
			continue
		}

		totalTokens += day.Tokens

		cost, est := dayCost(day, pricer)
		actualSpend += cost
		if est {
			isEstimated = true
		}

		existing := byDay[day.Day]
		existing.Day = day.Day
		existing.Tokens += day.Tokens
		existing.Input += day.Input
		existing.Output += day.Output
		existing.Reasoning += day.Reasoning
		existing.Cache += day.Cache
		existing.Turns += day.Turns
		if len(day.ReasoningEffort) > 0 {
			if existing.ReasoningEffort == nil {
				existing.ReasoningEffort = make(map[string]int)
			}
			for eff, n := range day.ReasoningEffort {
				existing.ReasoningEffort[eff] += n
			}
		}
		byDay[day.Day] = existing

		if minDay == "" || day.Day < minDay {
			minDay = day.Day
		}
		if maxDay == "" || day.Day > maxDay {
			maxDay = day.Day
		}

		// Category attribution: a day's model map is trusted only when its
		// disjoint categories cover the canonical total exactly. Otherwise the
		// canonical day categories are used, and any uncovered remainder is
		// counted as unclassified rather than invented input.
		modelsTotal := int64(0)
		for _, m := range day.Models {
			modelsTotal += m.Total()
		}
		dayTyped := day.Input + day.Output + day.Reasoning + day.Cache
		switch {
		case modelsTotal > 0 && modelsTotal == day.Tokens:
			for _, m := range day.Models {
				uncachedInput += m.Input
				output += m.Output
				reasoning += m.Reasoning
				cacheRead += m.CacheRead
				cacheWrite += m.CacheWrite
				unclassifiedTokens += m.Unclassified
			}
		case modelsTotal > day.Tokens || dayTyped > day.Tokens:
			// Inconsistent classification: preserve the canonical total and mark
			// the entire day's categories unclassified instead of scaling bad data.
			unclassifiedTokens += day.Tokens
		default:
			uncachedInput += day.Input
			output += day.Output
			reasoning += day.Reasoning
			cacheRead += day.Cache
			unclassifiedTokens += day.Tokens - dayTyped
		}
	}

	var dailyValues []int64
	var peakTokens int64
	var peakDay string
	for day, row := range byDay {
		val := row.Tokens
		if val > 0 {
			dailyValues = append(dailyValues, val)
			if val > peakTokens {
				peakTokens = val
				peakDay = day
			}
		}
	}
	sort.Slice(dailyValues, func(i, j int) bool { return dailyValues[i] < dailyValues[j] })

	activeDays := len(dailyValues)
	var medianTokens, p90Tokens int64
	if activeDays > 0 {
		medianTokens = dailyValues[activeDays/2]
		p90Idx := int(math.Ceil(float64(activeDays)*0.90)) - 1
		if p90Idx < 0 {
			p90Idx = 0
		}
		if p90Idx >= activeDays {
			p90Idx = activeDays - 1
		}
		p90Tokens = dailyValues[p90Idx]
	}

	var hitRate float64
	promptTotal := uncachedInput + cacheRead + cacheWrite
	if promptTotal > 0 {
		hitRate = float64(cacheRead) / float64(promptTotal)
	}

	// Calculate calendar span for normalization
	calendarDays := 30
	if minDay != "" && maxDay != "" {
		if t1, ok1 := ParseDay(minDay); ok1 {
			if t2, ok2 := ParseDay(maxDay); ok2 {
				span := int(t2.Sub(t1).Hours()/24) + 1
				if span > 0 {
					calendarDays = span
				}
			}
		}
	}
	if opts.Last > 0 {
		calendarDays = opts.Last
	}

	snap := WorkloadSnapshot{
		ActiveDays:         activeDays,
		TotalDays:          calendarDays,
		TotalTokens:        totalTokens,
		UncachedInput:      uncachedInput,
		Output:             output,
		Reasoning:          reasoning,
		CacheRead:          cacheRead,
		CacheWrite:         cacheWrite,
		UnclassifiedTokens: unclassifiedTokens,
		CacheHitRate:       hitRate,
		MedianDailyTokens:  medianTokens,
		P90DailyTokens:     p90Tokens,
		PeakDailyTokens:    peakTokens,
		PeakDay:            peakDay,
		ActualSpend:        actualSpend,
		IsEstimatedSpend:   isEstimated,
	}

	return snap, byDay, calendarDays
}

// applyReplayCostComparison fills the cost delta only for complete price
// comparisons. Partial and missing price rows keep zero placeholders and are
// never eligible for a savings recommendation.
func applyReplayCostComparison(row *PlanReplayRow, normActualSpend float64) {
	if row.PricingStatus != "complete" {
		row.CostDelta = 0
		row.CostDeltaPercent = 0
		row.IsRecommended = false
		return
	}
	row.CostDelta = row.MonthlyCost - normActualSpend
	if normActualSpend > 0 {
		row.CostDeltaPercent = (row.CostDelta / normActualSpend) * 100
	}
}

// SimulatePlan evaluates a subscription plan or API tier against a workload snapshot.
func SimulatePlan(snap WorkloadSnapshot, byDay map[string]DailyRow, calendarDays int, plan SubscriptionPlan, pricer ReplayPricer) PlanReplayRow {
	if calendarDays <= 0 {
		calendarDays = 30
	}
	normFactor := 30.0 / float64(calendarDays)

	// Normalized monthly actual spend
	normActualSpend := snap.ActualSpend * normFactor

	row := PlanReplayRow{
		ID:            plan.ID,
		Name:          plan.Name,
		Type:          string(plan.Type),
		TotalDays:     snap.ActiveDays,
		PricingStatus: "complete",
	}

	if plan.Type == PlanTypePayAsYouGo {
		row.PricingStatus = "missing"
		var periodCost float64
		if pricer != nil && plan.DefaultModel != "" {
			inRate, outRate, cacheReadRate, cacheWriteRate, ok := pricer.LookupPrice(plan.DefaultModel)
			if ok {
				periodCost = (float64(snap.UncachedInput)*inRate +
					float64(snap.Output+snap.Reasoning)*outRate +
					float64(snap.CacheRead)*cacheReadRate +
					float64(snap.CacheWrite)*cacheWriteRate) / 1e6
				if snap.UnclassifiedTokens > 0 {
					row.PricingStatus = "partial"
				} else {
					row.PricingStatus = "complete"
				}
			}
		}
		row.MonthlyCost = periodCost * normFactor
		row.CapacityVerdict = "PASS"
		row.VerdictDetail = "0% throttled (unmetered API)"
	} else {
		row.MonthlyCost = plan.MonthlyFee
		throttled := 0

		multiplier := plan.BurstMultiplier
		if multiplier <= 0 {
			multiplier = 1.0
		}

		for _, dayRow := range byDay {
			dayTokens := dayRow.Tokens
			effectiveBurstTokens := dayTokens
			if multiplier > 1.0 && dayRow.Reasoning > 0 {
				effectiveBurstTokens = (dayTokens - dayRow.Reasoning) + int64(float64(dayRow.Reasoning)*multiplier)
			}

			// Only the plan's configured daily ceiling is assessable from daily
			// rows. A day's sum is not a rolling-window peak: intraday timestamps
			// are required to observe one, so rolling limits are never compared
			// against a daily total and burst spikes stay unasserted.
			if plan.DailyTokenLimit > 0 && effectiveBurstTokens > plan.DailyTokenLimit {
				throttled++
			}
		}

		row.ThrottledDays = throttled
		row.BurstThrottles = 0
		if snap.ActiveDays > 0 {
			row.ThrottleRate = float64(throttled) / float64(snap.ActiveDays)
		}

		rollingNote := ""
		if plan.RollingWindowLimit > 0 {
			rollingNote = " · rolling spikes unasserted"
		}

		switch {
		case throttled == 0 && plan.RollingWindowLimit > 0:
			row.CapacityVerdict = "UNKNOWN"
			row.VerdictDetail = "rolling-window capacity cannot be assessed from daily totals; intraday evidence unavailable"
		case throttled == 0:
			row.CapacityVerdict = "PASS"
			row.VerdictDetail = "0% throttled"
		case row.ThrottleRate <= 0.20:
			row.CapacityVerdict = "DEGRADED"
			row.VerdictDetail = fmt.Sprintf("throttled %d/%d days (%.0f%%)%s", throttled, snap.ActiveDays, row.ThrottleRate*100, rollingNote)
		default:
			row.CapacityVerdict = "FAIL"
			row.VerdictDetail = fmt.Sprintf("throttled %d/%d days (%.0f%%)%s", throttled, snap.ActiveDays, row.ThrottleRate*100, rollingNote)
		}
	}

	applyReplayCostComparison(&row, normActualSpend)

	return row
}

// SimulateModel prices the workload against a standalone model rate card.
func SimulateModel(snap WorkloadSnapshot, calendarDays int, modelID string, pricer ReplayPricer) PlanReplayRow {
	if calendarDays <= 0 {
		calendarDays = 30
	}
	normFactor := 30.0 / float64(calendarDays)
	normActualSpend := snap.ActualSpend * normFactor

	row := PlanReplayRow{
		ID:              modelID,
		Name:            fmt.Sprintf("%s (API)", modelID),
		Type:            string(PlanTypePayAsYouGo),
		TotalDays:       snap.ActiveDays,
		CapacityVerdict: "PASS",
		VerdictDetail:   "0% throttled (unmetered API)",
		PricingStatus:   "missing",
	}

	var periodCost float64
	if pricer != nil {
		inRate, outRate, cacheReadRate, cacheWriteRate, ok := pricer.LookupPrice(modelID)
		if ok {
			periodCost = (float64(snap.UncachedInput)*inRate +
				float64(snap.Output+snap.Reasoning)*outRate +
				float64(snap.CacheRead)*cacheReadRate +
				float64(snap.CacheWrite)*cacheWriteRate) / 1e6
			if snap.UnclassifiedTokens > 0 {
				row.PricingStatus = "partial"
			} else {
				row.PricingStatus = "complete"
			}
		}
	}
	row.MonthlyCost = periodCost * normFactor
	applyReplayCostComparison(&row, normActualSpend)
	return row
}

// AggregateReplay coordinates workload extraction, plan evaluation, and
// recommendation synthesis.
func AggregateReplay(days []DailyRow, opts ReplayOptions, pricer ReplayPricer) ReplayReport {
	// Default to last 30 days if no time bounds specified for monthly billing parity
	if opts.Since == "" && opts.Until == "" && opts.Last == 0 {
		opts.Last = 30
	}

	snap, byDay, calendarDays := BuildWorkloadSnapshot(days, opts, pricer)

	report := ReplayReport{
		Type:     "replay",
		Since:    opts.Since,
		Until:    opts.Until,
		Workload: snap,
	}

	// 1. If --against is specified, replay specifically against that model
	if opts.Against != "" {
		row := SimulateModel(snap, calendarDays, opts.Against, pricer)
		report.Plans = append(report.Plans, row)
	}

	// 2. Replay candidate comparison plans
	for _, plan := range opts.Compare {
		row := SimulatePlan(snap, byDay, calendarDays, plan, pricer)
		report.Plans = append(report.Plans, row)
	}

	// 3. Mark best recommended plan and synthesize guidance
	var bestPlanIdx = -1
	var minPassingCost = math.MaxFloat64

	for i, p := range report.Plans {
		if p.CapacityVerdict == "PASS" && p.PricingStatus == "complete" && p.MonthlyCost < minPassingCost {
			minPassingCost = p.MonthlyCost
			bestPlanIdx = i
		}
	}
	if bestPlanIdx >= 0 {
		report.Plans[bestPlanIdx].IsRecommended = true
	}

	// Build recommendation narrative
	var rec strings.Builder
	if snap.ActiveDays > 0 {
		p90Formatted := CompactNumber(snap.P90DailyTokens)
		medianFormatted := CompactNumber(snap.MedianDailyTokens)

		hasSubThrottle := false
		for _, p := range report.Plans {
			if p.Type == string(PlanTypeSubscription) && p.CapacityVerdict == "FAIL" {
				hasSubThrottle = true
				break
			}
		}

		if hasSubThrottle {
			rec.WriteString(fmt.Sprintf("Your median daily volume (%s) and p90 burst (%s) exceed flat $20/mo rate limits on heavy days.\n", medianFormatted, p90Formatted))
		}
		if bestPlanIdx >= 0 {
			bp := report.Plans[bestPlanIdx]
			if bp.CostDelta < 0 {
				rec.WriteString(fmt.Sprintf("Switching to %s would cost ~$%.2f/mo with 0%% throttling, saving $%.2f/mo.", bp.Name, bp.MonthlyCost, math.Abs(bp.CostDelta)))
			} else {
				rec.WriteString(fmt.Sprintf("%s provides full unthrottled throughput for your workload at ~$%.2f/mo.", bp.Name, bp.MonthlyCost))
			}
		}
	} else {
		rec.WriteString("No active coding sessions found in the selected window.")
	}

	report.Recommendation = strings.TrimSpace(rec.String())
	return report
}
