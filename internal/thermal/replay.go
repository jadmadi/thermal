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
func BuildWorkloadSnapshot(days []DailyRow, opts ReplayOptions, pricer Pricer) (WorkloadSnapshot, map[string]int64, int) {
	since, until, lastStart := windowBounds(opts.Since, opts.Until, opts.Last, opts.Now)

	byDay := make(map[string]int64)
	var (
		uncachedInput int64
		output        int64
		reasoning     int64
		cacheRead     int64
		cacheWrite    int64
		actualSpend   float64
		isEstimated   bool
		minDay        string
		maxDay        string
	)

	for _, day := range days {
		t, ok := ParseDay(day.Day)
		if !ok || !inWindow(t, since, until, lastStart) {
			continue
		}
		if isActivityOnly(day) {
			continue
		}

		cost, est := dayCost(day, pricer)
		actualSpend += cost
		if est {
			isEstimated = true
		}

		byDay[day.Day] += day.Tokens
		if minDay == "" || day.Day < minDay {
			minDay = day.Day
		}
		if maxDay == "" || day.Day > maxDay {
			maxDay = day.Day
		}

		if len(day.Models) > 0 {
			for _, m := range day.Models {
				uncachedInput += m.Input
				output += m.Output
				reasoning += m.Reasoning
				cacheRead += m.CacheRead
				cacheWrite += m.CacheWrite
			}
		} else {
			uncachedInput += day.Input
			output += day.Output
			reasoning += day.Reasoning
			cacheRead += day.Cache
			typedSum := day.Input + day.Output + day.Reasoning + day.Cache
			if day.Tokens > typedSum {
				uncachedInput += (day.Tokens - typedSum)
			}
		}
	}

	var dailyValues []int64
	var peakTokens int64
	var peakDay string
	for day, val := range byDay {
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

	totalTokens := uncachedInput + output + reasoning + cacheRead + cacheWrite
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
		ActiveDays:        activeDays,
		TotalDays:         calendarDays,
		TotalTokens:       totalTokens,
		UncachedInput:     uncachedInput,
		Output:            output,
		Reasoning:         reasoning,
		CacheRead:         cacheRead,
		CacheWrite:        cacheWrite,
		CacheHitRate:      hitRate,
		MedianDailyTokens: medianTokens,
		P90DailyTokens:    p90Tokens,
		PeakDailyTokens:   peakTokens,
		PeakDay:           peakDay,
		ActualSpend:       actualSpend,
		IsEstimatedSpend:  isEstimated,
	}

	return snap, byDay, calendarDays
}

// SimulatePlan evaluates a subscription plan or API tier against a workload snapshot.
func SimulatePlan(snap WorkloadSnapshot, byDay map[string]int64, calendarDays int, plan SubscriptionPlan, pricer ReplayPricer) PlanReplayRow {
	if calendarDays <= 0 {
		calendarDays = 30
	}
	normFactor := 30.0 / float64(calendarDays)

	// Normalized monthly actual spend
	normActualSpend := snap.ActualSpend * normFactor

	row := PlanReplayRow{
		ID:        plan.ID,
		Name:      plan.Name,
		Type:      string(plan.Type),
		TotalDays: snap.ActiveDays,
	}

	if plan.Type == PlanTypePayAsYouGo {
		var periodCost float64
		if pricer != nil && plan.DefaultModel != "" {
			inRate, outRate, cacheReadRate, cacheWriteRate, ok := pricer.LookupPrice(plan.DefaultModel)
			if ok {
				periodCost = (float64(snap.UncachedInput)*inRate +
					float64(snap.Output+snap.Reasoning)*outRate +
					float64(snap.CacheRead)*cacheReadRate +
					float64(snap.CacheWrite)*cacheWriteRate) / 1e6
			}
		}
		row.MonthlyCost = periodCost * normFactor
		row.CapacityVerdict = "PASS"
		row.VerdictDetail = "0% throttled (unmetered API)"
	} else {
		row.MonthlyCost = plan.MonthlyFee
		throttled := 0
		if plan.DailyTokenLimit > 0 {
			for _, tokens := range byDay {
				if tokens > plan.DailyTokenLimit {
					throttled++
				}
			}
		}
		row.ThrottledDays = throttled
		if snap.ActiveDays > 0 {
			row.ThrottleRate = float64(throttled) / float64(snap.ActiveDays)
		}
		if throttled == 0 {
			row.CapacityVerdict = "PASS"
			row.VerdictDetail = "0% throttled"
		} else if row.ThrottleRate <= 0.20 {
			row.CapacityVerdict = "DEGRADED"
			row.VerdictDetail = fmt.Sprintf("throttled %d/%d days (%.0f%%)", throttled, snap.ActiveDays, row.ThrottleRate*100)
		} else {
			row.CapacityVerdict = "FAIL"
			row.VerdictDetail = fmt.Sprintf("throttled %d/%d days (%.0f%%)", throttled, snap.ActiveDays, row.ThrottleRate*100)
		}
	}

	row.CostDelta = row.MonthlyCost - normActualSpend
	if normActualSpend > 0 {
		row.CostDeltaPercent = (row.CostDelta / normActualSpend) * 100
	}

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
	}

	var periodCost float64
	if pricer != nil {
		inRate, outRate, cacheReadRate, cacheWriteRate, ok := pricer.LookupPrice(modelID)
		if ok {
			periodCost = (float64(snap.UncachedInput)*inRate +
				float64(snap.Output+snap.Reasoning)*outRate +
				float64(snap.CacheRead)*cacheReadRate +
				float64(snap.CacheWrite)*cacheWriteRate) / 1e6
		}
	}
	row.MonthlyCost = periodCost * normFactor
	row.CostDelta = row.MonthlyCost - normActualSpend
	if normActualSpend > 0 {
		row.CostDeltaPercent = (row.CostDelta / normActualSpend) * 100
	}
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
		if p.CapacityVerdict == "PASS" && p.MonthlyCost < minPassingCost {
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
