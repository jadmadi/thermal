// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// ActivityCategory classifies developer operations into standard engineering activity buckets.
type ActivityCategory string

const (
	ActivityCoding        ActivityCategory = "Coding"
	ActivityConversation  ActivityCategory = "Conversation"
	ActivityExploration   ActivityCategory = "Exploration"
	ActivityDelegation    ActivityCategory = "Delegation"
	ActivityFeatureDev    ActivityCategory = "Feature Dev"
	ActivityTesting       ActivityCategory = "Testing"
	ActivityBuildDeploy   ActivityCategory = "Build/Deploy"
	ActivityBrainstorming ActivityCategory = "Brainstorming"
	ActivityDebugging     ActivityCategory = "Debugging"
)

var (
	cachedReceiptsMu      sync.Mutex
	cachedReceiptsSummary FinOpsReceiptSummary
	cachedReceiptsPr      Pricer
	cachedReceiptsValid   bool
)

// ActivityShare holds token and percentage metrics for an activity category.
type ActivityShare struct {
	Category ActivityCategory `json:"category"`
	Tokens   int64            `json:"tokens"`
	Cost     float64          `json:"cost"`
	Turns    int              `json:"turns"`
	Percent  float64          `json:"percent"`
	OneShot  string           `json:"oneShot,omitempty"`
}

// SubToolCall records execution count and share for an extracted shell utility.
type SubToolCall struct {
	Name  string  `json:"name"`
	Calls int     `json:"calls"`
	Share float64 `json:"share"`
}

// CacheMetrics captures prompt caching performance and FinOps cost savings.
type CacheMetrics struct {
	CacheReadTokens  int64   `json:"cacheReadTokens"`
	CacheWriteTokens int64   `json:"cacheWriteTokens"`
	InputTokens      int64   `json:"inputTokens"`
	HitRate          float64 `json:"hitRate"`          // percentage: read / (input + read) * 100
	EstimatedSavings float64 `json:"estimatedSavings"` // estimated USD saved by cache hits vs baseline input
}

// MCPMetrics tracks Model Context Protocol tool overhead and payload usage.
type MCPMetrics struct {
	ServerCalls    int   `json:"serverCalls"`
	OverheadTokens int64 `json:"overheadTokens"`
	ServersActive  int   `json:"serversActive"`
}

// FinOpsDayRow records daily activity for the dense FinOps grid.
type FinOpsDayRow struct {
	Date  string  `json:"date"`
	Cost  float64 `json:"cost"`
	Calls int     `json:"calls"`
}

// FinOpsProjectRow records per-project metrics.
type FinOpsProjectRow struct {
	Name     string  `json:"name"`
	Cost     float64 `json:"cost"`
	AvgCost  float64 `json:"avgCost"`
	Sessions int     `json:"sessions"`
	Overhead int64   `json:"overhead"`
	Tokens   int64   `json:"tokens"`
}

// FinOpsModelRow records per-model metrics.
type FinOpsModelRow struct {
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	CachePct    float64 `json:"cachePct"`
	Calls       int     `json:"calls"`
	OneShot     string  `json:"oneShot"`
	TokPerS     string  `json:"tokPerS"`
	IsEstimated bool    `json:"isEstimated,omitempty"`
}

// FinOpsToolRow records tool/assistant usage share and streak.
type FinOpsToolRow struct {
	Name     string  `json:"name"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
	Share    float64 `json:"share"`
	Streak   int     `json:"streak"`
	Sessions int     `json:"sessions"`
}

// FinOpsReceiptSummary captures factual test and linter execution outcomes.
type FinOpsReceiptSummary struct {
	VerifiedCount   int     `json:"verifiedCount"`
	ClaimedCount    int     `json:"claimedCount"`
	FailedCount     int     `json:"failedCount"`
	TokensVerified  int64   `json:"tokensVerified"`
	VerifiedRate    float64 `json:"verifiedRate"`
	SpendEfficiency string  `json:"spendEfficiency"`
	TestTools       string  `json:"testTools,omitempty"`
}

// FinOpsReplaySummary captures subscription capacity simulations.
type FinOpsReplaySummary struct {
	CapacityVerdict string             `json:"capacityVerdict"`
	Plans           []FinOpsReplayPlan `json:"plans"`
}

// FinOpsReplayPlan records individual subscription/model simulation outcome.
type FinOpsReplayPlan struct {
	Name            string  `json:"name"`
	MonthlyCost     float64 `json:"monthlyCost"`
	CostDelta       float64 `json:"costDelta"`
	CapacityVerdict string  `json:"capacityVerdict"`
	ThrottleRate    float64 `json:"throttleRate"`
}

// FinOpsCompositionSummary tracks disjoint token type shares.
type FinOpsCompositionSummary struct {
	InputTokens      int64   `json:"inputTokens"`
	OutputTokens     int64   `json:"outputTokens"`
	ReasoningTokens  int64   `json:"reasoningTokens"`
	CacheReadTokens  int64   `json:"cacheReadTokens"`
	CacheWriteTokens int64   `json:"cacheWriteTokens"`
	ReasoningRatio   float64 `json:"reasoningRatio"`
}

// CoreToolCall records execution count for backward compatibility.
type CoreToolCall struct {
	Name  string `json:"name"`
	Calls int    `json:"calls"`
}

// SkillAgentCall records uses and cost for backward compatibility.
type SkillAgentCall struct {
	Name string  `json:"name"`
	Uses int     `json:"uses"`
	Cost float64 `json:"cost"`
}

// WorkflowMetrics tracks developer iteration metrics.
type WorkflowMetrics struct {
	Corrections string `json:"corrections"`
	FirstEdit   string `json:"firstEdit"`
	Rework      string `json:"rework"`
	Coverage    string `json:"coverage"`
}

// TodaySummary aggregates key metrics for the headline card.
type TodaySummary struct {
	Period       string  `json:"period,omitempty"`
	Project      string  `json:"project,omitempty"`
	Cost         float64 `json:"cost"`
	Calls        int     `json:"calls"`
	Sessions     int     `json:"sessions"`
	CacheHitRate float64 `json:"cacheHitRate"`
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
	CachedTokens int64   `json:"cachedTokens"`
	WriteTokens  int64   `json:"writeTokens"`
}

// FinOpsGridPayload is the data model powering the 9-box dense FinOps dashboard.
type FinOpsGridPayload struct {
	Period        string       `json:"period,omitempty"`
	ActiveProject string       `json:"activeProject,omitempty"`
	Today         TodaySummary `json:"today"`

	// 1. Daily Activity (last 10 days)
	DailyHistory     []FinOpsDayRow `json:"dailyHistory"`
	TotalDaysScanned int            `json:"totalDaysScanned"`

	// 2. By Project
	ProjectBreakdown []FinOpsProjectRow `json:"projectBreakdown"`

	// 3. Assistant Mix (Box 3)
	ToolMix []FinOpsToolRow `json:"toolMix"`

	// 4. By Model (Box 4)
	ModelBreakdown []FinOpsModelRow `json:"modelBreakdown"`

	// 5. Prompt Cache & Savings (Box 5)
	Cache CacheMetrics `json:"cache"`

	// 6. Code Yield & Velocity (Box 6)
	Yield YieldRow `json:"yield"`

	// 7. Work Receipts & Tests (Box 7)
	Receipt FinOpsReceiptSummary `json:"receipt"`

	// 8. Subscription Replay (Box 8)
	Replay FinOpsReplaySummary `json:"replay"`

	// 9. Token Composition (Box 9)
	Composition FinOpsCompositionSummary `json:"composition"`

	// Token Share / Taxonomy
	Taxonomy []ActivityShare `json:"taxonomy"`

	// Backward-compatibility fields
	MCP                   MCPMetrics       `json:"mcp,omitempty"`
	CoreTools             []CoreToolCall   `json:"coreTools,omitempty"`
	SubTools              []SubToolCall    `json:"subTools,omitempty"`
	SkillsAgents          []SkillAgentCall `json:"skillsAgents,omitempty"`
	Workflow              WorkflowMetrics  `json:"workflow,omitempty"`
	TotalTokens           int64            `json:"totalTokens"`
	TotalCost             float64          `json:"totalCost"`
	ActiveDays            int              `json:"activeDays"`
	AvgDailyTokens        int64            `json:"avgDailyTokens"`
	AvgDailyCost          float64          `json:"avgDailyCost"`
	SpendEfficiency       string           `json:"spendEfficiency"`
	TopModels             []YieldRow       `json:"topModels,omitempty"`
	TopProjects           []YieldRow       `json:"topProjects,omitempty"`
	MonthEndRunRateTokens int64            `json:"monthEndRunRateTokens"`
	MonthEndRunRateCost   float64          `json:"monthEndRunRateCost"`
	CapacityVerdict       string           `json:"capacityVerdict"`
}

// ClassifyCommand classifies a shell command line into an ActivityCategory and utility name
// using zero heap allocations.
func ClassifyCommand(cmd string) (ActivityCategory, string) {
	cmdClean := strings.TrimSpace(cmd)
	if strings.HasPrefix(cmdClean, "\"") && strings.HasSuffix(cmdClean, "\"") && len(cmdClean) >= 2 {
		cmdClean = strings.TrimSpace(cmdClean[1 : len(cmdClean)-1])
	}
	cmdLower := strings.ToLower(cmdClean)

	// Utility extraction: extract first token
	var util string
	if idx := strings.IndexAny(cmdClean, " \t|;&"); idx >= 0 {
		util = cmdClean[:idx]
	} else {
		util = cmdClean
	}
	utilLower := strings.ToLower(util)

	// 1. Testing
	if strings.HasPrefix(cmdLower, "go test") || strings.Contains(cmdLower, "cargo test") ||
		strings.Contains(cmdLower, "pytest") || strings.Contains(cmdLower, "npm test") ||
		strings.Contains(cmdLower, "npm t ") || strings.Contains(cmdLower, "pnpm test") ||
		strings.Contains(cmdLower, "jest") || strings.Contains(cmdLower, "vitest") ||
		strings.Contains(cmdLower, "golangci-lint") || strings.Contains(cmdLower, "eslint") ||
		strings.Contains(cmdLower, "ruff") || strings.Contains(cmdLower, "cargo clippy") {
		return ActivityTesting, utilLower
	}

	// 2. Debugging
	if strings.Contains(cmdLower, "log") || strings.Contains(cmdLower, "dmesg") ||
		strings.Contains(cmdLower, "journalctl") || strings.Contains(cmdLower, "gdb") ||
		strings.Contains(cmdLower, "lldb") || strings.Contains(cmdLower, "strace") ||
		strings.Contains(cmdLower, "pstack") || strings.Contains(cmdLower, "pprof") ||
		strings.Contains(cmdLower, "debug") {
		return ActivityDebugging, utilLower
	}

	// 3. Coding (writes, git commits, edits, builds)
	if strings.HasPrefix(cmdLower, "git commit") || strings.HasPrefix(cmdLower, "git merge") ||
		strings.Contains(cmdLower, "git checkout -b") || strings.Contains(cmdLower, "go build") ||
		strings.Contains(cmdLower, "cargo build") || strings.Contains(cmdLower, "npm run build") ||
		strings.Contains(cmdLower, "make") || strings.Contains(cmdLower, "patch") {
		return ActivityCoding, utilLower
	}

	// 4. Exploration (searches, directory listings, inspections)
	if strings.HasPrefix(cmdLower, "git status") || strings.HasPrefix(cmdLower, "git diff") ||
		strings.HasPrefix(cmdLower, "git log") || strings.HasPrefix(cmdLower, "ls") ||
		strings.HasPrefix(cmdLower, "find") || strings.HasPrefix(cmdLower, "grep") ||
		strings.HasPrefix(cmdLower, "rg") || strings.HasPrefix(cmdLower, "cat") ||
		strings.HasPrefix(cmdLower, "head") || strings.HasPrefix(cmdLower, "tail") ||
		strings.HasPrefix(cmdLower, "curl") || strings.HasPrefix(cmdLower, "which") {
		return ActivityExploration, utilLower
	}

	return ActivityCoding, utilLower
}

// ComputeFinOpsGrid compiles the multi-dimensional 9-box FinOps grid metrics across loaded tools for Today.
func ComputeFinOpsGrid(days []DailyRow, results []ToolResult, projects []ProjectDay, yieldRep YieldReport, pricer Pricer) FinOpsGridPayload {
	return computeFinOpsGridCore(days, results, projects, yieldRep, pricer, "Today", "")
}

// ComputeFinOpsGridForPeriod compiles the 9-box FinOps grid metrics for a specified period and optional project filter.
func ComputeFinOpsGridForPeriod(days []DailyRow, results []ToolResult, projects []ProjectDay, yieldRep YieldReport, pricer Pricer, period string, projectFilter string) FinOpsGridPayload {
	if period == "" {
		period = "Today"
	}

	var uniqueDates []string
	dateSet := make(map[string]bool)
	for _, d := range days {
		if (d.Tokens > 0 || d.Turns > 0) && !dateSet[d.Day] {
			dateSet[d.Day] = true
			uniqueDates = append(uniqueDates, d.Day)
		}
	}
	sort.Slice(uniqueDates, func(i, j int) bool {
		return uniqueDates[i] > uniqueDates[j]
	})

	var latestDay string
	if len(uniqueDates) > 0 {
		latestDay = uniqueDates[0]
	} else {
		latestDay = time.Now().Format("2006-01-02")
	}

	latestTime, err := time.Parse("2006-01-02", latestDay)
	if err != nil {
		latestTime = time.Now()
	}

	var cutoff string
	switch period {
	case "Today":
		cutoff = latestDay
	case "7 Days":
		cutoff = latestTime.AddDate(0, 0, -6).Format("2006-01-02")
	case "30 Days":
		cutoff = latestTime.AddDate(0, 0, -29).Format("2006-01-02")
	case "This Month":
		cutoff = latestTime.Format("2006-01") + "-01"
	case "6 Months":
		cutoff = latestTime.AddDate(0, -6, 0).Format("2006-01-02")
	case "Lifetime":
		cutoff = ""
	default:
		cutoff = latestDay
	}

	var filteredDays []DailyRow
	for _, d := range days {
		if period == "Today" {
			if d.Day == latestDay {
				filteredDays = append(filteredDays, d)
			}
		} else if cutoff != "" {
			if d.Day >= cutoff && d.Day <= latestDay {
				filteredDays = append(filteredDays, d)
			}
		} else {
			filteredDays = append(filteredDays, d)
		}
	}

	var filteredProjects []ProjectDay
	for _, p := range projects {
		if projectFilter != "" && p.Project != projectFilter && ProjectSlug(p.Project) != projectFilter {
			continue
		}
		if period == "Today" {
			if p.Day == latestDay {
				filteredProjects = append(filteredProjects, p)
			}
		} else if cutoff != "" {
			if p.Day >= cutoff && p.Day <= latestDay {
				filteredProjects = append(filteredProjects, p)
			}
		} else {
			filteredProjects = append(filteredProjects, p)
		}
	}

	var filteredResults []ToolResult
	for _, res := range results {
		resCopy := res
		var resDays []DailyRow
		for _, d := range res.Daily {
			if period == "Today" {
				if d.Day == latestDay {
					resDays = append(resDays, d)
				}
			} else if cutoff != "" {
				if d.Day >= cutoff && d.Day <= latestDay {
					resDays = append(resDays, d)
				}
			} else {
				resDays = append(resDays, d)
			}
		}
		resCopy.Daily = resDays
		filteredResults = append(filteredResults, resCopy)
	}

	if projectFilter != "" {
		byDay := make(map[string]*DailyRow)
		for _, p := range filteredProjects {
			dr, ok := byDay[p.Day]
			if !ok {
				dr = &DailyRow{
					Day:    p.Day,
					Models: make(map[string]ModelTokens),
				}
				byDay[p.Day] = dr
			}
			dr.Tokens += p.Tokens
			dr.Turns += p.Turns
			dr.Input += p.Input
			dr.Output += p.Output
			dr.Reasoning += p.Reasoning
			dr.Cache += (p.CacheRead + p.CacheWrite)
			dr.Cost += p.Cost
			if len(p.Models) > 0 {
				for m, mt := range p.Models {
					dr.Models[m] = dr.Models[m].Add(mt)
				}
			}
		}
		filteredDays = nil
		for _, dr := range byDay {
			filteredDays = append(filteredDays, *dr)
		}
		sort.Slice(filteredDays, func(i, j int) bool {
			return filteredDays[i].Day > filteredDays[j].Day
		})

		for i := range filteredResults {
			var projResDays []DailyRow
			for _, pd := range filteredProjects {
				if pd.Tool == filteredResults[i].Name {
					projResDays = append(projResDays, DailyRow{
						Day:       pd.Day,
						Tokens:    pd.Tokens,
						Turns:     pd.Turns,
						Input:     pd.Input,
						Output:    pd.Output,
						Reasoning: pd.Reasoning,
						Cache:     pd.CacheRead + pd.CacheWrite,
						Cost:      pd.Cost,
						Models:    pd.Models,
					})
				}
			}
			filteredResults[i].Daily = projResDays
		}
	}

	grid := computeFinOpsGridCore(filteredDays, filteredResults, filteredProjects, yieldRep, pricer, period, projectFilter)
	grid.Period = period
	grid.ActiveProject = ProjectSlug(projectFilter)
	grid.Today.Period = period
	grid.Today.Project = ProjectSlug(projectFilter)
	return grid
}

func computeFinOpsGridCore(days []DailyRow, results []ToolResult, projects []ProjectDay, yieldRep YieldReport, pricer Pricer, period string, projectFilter string) FinOpsGridPayload {
	var (
		totalTokens     int64
		totalCost       float64
		inputTokens     int64
		outputTokens    int64
		reasoningTokens int64
		cacheRead       int64
		cacheWrite      int64
		totalCalls      int
		activeDays      int
	)

	// Sort days descending by date
	daysDesc := make([]DailyRow, len(days))
	copy(daysDesc, days)
	sort.Slice(daysDesc, func(i, j int) bool {
		return daysDesc[i].Day > daysDesc[j].Day
	})

	for _, d := range days {
		if d.Tokens > 0 || d.Turns > 0 {
			activeDays++
		}
		totalTokens += d.Tokens
		totalCalls += d.Turns

		dayCost := d.Cost
		if dayCost <= 0 && pricer != nil && len(d.Models) > 0 {
			c, _ := pricer.PriceDay(d)
			dayCost = c
		}
		totalCost += dayCost

		if len(d.Models) > 0 {
			for _, m := range d.Models {
				inputTokens += m.Input
				outputTokens += m.Output
				reasoningTokens += m.Reasoning
				cacheRead += m.CacheRead
				cacheWrite += m.CacheWrite
			}
		} else {
			inputTokens += d.Input
			outputTokens += d.Output
			reasoningTokens += d.Reasoning
			cacheRead += d.Cache
		}
	}

	// If disjoint tokens were not broken down by the source, assign remainder to input
	disjointSum := inputTokens + outputTokens + reasoningTokens + cacheRead + cacheWrite
	if disjointSum < totalTokens {
		inputTokens += (totalTokens - disjointSum)
	}

	// Default call count if not recorded
	if totalCalls == 0 && totalTokens > 0 {
		totalCalls = int(totalTokens / 2500)
		if totalCalls < 1 {
			totalCalls = 1
		}
	}

	var avgDailyTokens int64
	var avgDailyCost float64
	if activeDays > 0 {
		avgDailyTokens = totalTokens / int64(activeDays)
		avgDailyCost = totalCost / float64(activeDays)
	}

	// Aggregate daily rows across tools by date
	type dayAgg struct {
		Date            string
		Cost            float64
		Calls           int
		Tokens          int64
		InputTokens     int64
		OutputTokens    int64
		ReasoningTokens int64
		CachedTokens    int64
		WriteTokens     int64
	}
	dayAggMap := make(map[string]*dayAgg)
	var uniqueDates []string
	for _, d := range days {
		if d.Tokens == 0 && d.Turns == 0 {
			continue
		}
		agg, ok := dayAggMap[d.Day]
		if !ok {
			agg = &dayAgg{Date: d.Day}
			dayAggMap[d.Day] = agg
			uniqueDates = append(uniqueDates, d.Day)
		}
		c := d.Cost
		if c <= 0 && pricer != nil && len(d.Models) > 0 {
			c, _ = pricer.PriceDay(d)
		}
		agg.Cost += c
		calls := d.Turns
		if calls <= 0 && d.Tokens > 0 {
			calls = max(1, int(d.Tokens/2500))
		}
		agg.Calls += calls
		agg.Tokens += d.Tokens

		var dayIn, dayOut, dayReas, dayCR, dayCW int64
		if len(d.Models) > 0 {
			for _, m := range d.Models {
				dayIn += m.Input
				dayOut += m.Output
				dayReas += m.Reasoning
				dayCR += m.CacheRead
				dayCW += m.CacheWrite
			}
		} else {
			dayIn = d.Input
			dayOut = d.Output
			dayReas = d.Reasoning
			dayCR = d.Cache
		}
		dayDisjoint := dayIn + dayOut + dayReas + dayCR + dayCW
		if dayDisjoint < d.Tokens {
			dayIn += (d.Tokens - dayDisjoint)
		}
		agg.InputTokens += dayIn
		agg.OutputTokens += dayOut
		agg.ReasoningTokens += dayReas
		agg.CachedTokens += dayCR
		agg.WriteTokens += dayCW
	}

	sort.Slice(uniqueDates, func(i, j int) bool {
		return uniqueDates[i] > uniqueDates[j]
	})

	// 1. Headline summary
	var todaySummary TodaySummary
	if period == "Today" && len(uniqueDates) > 0 {
		latest := dayAggMap[uniqueDates[0]]
		var hitRate float64
		if latest.InputTokens+latest.CachedTokens > 0 {
			hitRate = (float64(latest.CachedTokens) / float64(latest.InputTokens+latest.CachedTokens)) * 100.0
		}

		todaySummary = TodaySummary{
			Period:       period,
			Project:      ProjectSlug(projectFilter),
			Cost:         latest.Cost,
			Calls:        max(1, latest.Calls),
			Sessions:     max(1, latest.Calls/4),
			CacheHitRate: hitRate,
			InputTokens:  latest.InputTokens,
			OutputTokens: latest.OutputTokens,
			CachedTokens: latest.CachedTokens,
			WriteTokens:  latest.WriteTokens,
		}
	} else if len(uniqueDates) > 0 {
		var hitRate float64
		if inputTokens+cacheRead > 0 {
			hitRate = (float64(cacheRead) / float64(inputTokens+cacheRead)) * 100.0
		}
		todaySummary = TodaySummary{
			Period:       period,
			Project:      ProjectSlug(projectFilter),
			Cost:         totalCost,
			Calls:        max(1, totalCalls),
			Sessions:     max(1, totalCalls/4),
			CacheHitRate: hitRate,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			CachedTokens: cacheRead,
			WriteTokens:  cacheWrite,
		}
	} else {
		todaySummary = TodaySummary{
			Period:       period,
			Project:      ProjectSlug(projectFilter),
			Cost:         totalCost,
			Calls:        max(1, totalCalls),
			Sessions:     1,
			CacheHitRate: 0.0,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			CachedTokens: cacheRead,
			WriteTokens:  cacheWrite,
		}
	}

	// 2. Box 1: Daily Activity (last 10 unique days scanned, newest first)
	var dailyHistory []FinOpsDayRow
	for i := 0; i < len(uniqueDates) && i < 10; i++ {
		agg := dayAggMap[uniqueDates[i]]
		dailyHistory = append(dailyHistory, FinOpsDayRow{
			Date:  agg.Date,
			Cost:  agg.Cost,
			Calls: max(1, agg.Calls),
		})
	}

	// 3. Box 2: Project Breakdown
	projMap := make(map[string]*FinOpsProjectRow)
	for _, p := range projects {
		if period == "Today" && len(uniqueDates) > 0 && p.Day != uniqueDates[0] {
			continue
		}
		name := ProjectSlug(p.Project)
		if name == "" {
			name = "default"
		}
		row, ok := projMap[name]
		if !ok {
			row = &FinOpsProjectRow{Name: name}
			projMap[name] = row
		}
		row.Tokens += p.Tokens
		row.Sessions++
		pCost := p.Cost
		if pCost <= 0 && pricer != nil && len(p.Models) > 0 {
			c, _ := pricer.PriceDay(DailyRow{
				Day:       p.Day,
				Tokens:    p.Tokens,
				Input:     p.Input,
				Output:    p.Output,
				Reasoning: p.Reasoning,
				Cache:     p.CacheRead + p.CacheWrite,
				Models:    p.Models,
			})
			pCost = c
		}
		row.Cost += pCost
		row.Overhead += int64(float64(p.Tokens) * 0.02)
	}

	var projList []FinOpsProjectRow
	for _, pr := range projMap {
		if pr.Sessions > 0 {
			pr.AvgCost = pr.Cost / float64(pr.Sessions)
		}
		projList = append(projList, *pr)
	}
	sort.Slice(projList, func(i, j int) bool {
		if projList[i].Cost != projList[j].Cost {
			return projList[i].Cost > projList[j].Cost
		}
		return projList[i].Tokens > projList[j].Tokens
	})
	if len(projList) > 6 {
		projList = projList[:6]
	}

	// 4. Box 3: Assistant Mix (Tool breakdown across results)
	var toolMix []FinOpsToolRow
	for _, res := range results {
		var (
			tTokens int64
			tCost   float64
			tCalls  int
			tDays   = make(map[string]bool)
		)
		for _, d := range res.Daily {
			if d.Tokens == 0 && d.Turns == 0 {
				continue
			}
			tTokens += d.Tokens
			c := d.Cost
			if c <= 0 && pricer != nil && len(d.Models) > 0 {
				c, _ = pricer.PriceDay(d)
			}
			tCost += c
			tCalls += d.Turns
			tDays[d.Day] = true
		}
		if tTokens == 0 && tCalls == 0 {
			continue
		}
		streak, _ := ComputeStreaks(tDays)
		share := 0.0
		if totalTokens > 0 {
			share = (float64(tTokens) / float64(totalTokens)) * 100.0
		}
		toolMix = append(toolMix, FinOpsToolRow{
			Name:     res.Name,
			Tokens:   tTokens,
			Cost:     tCost,
			Share:    share,
			Streak:   streak,
			Sessions: max(1, tCalls/4),
		})
	}
	sort.Slice(toolMix, func(i, j int) bool {
		if toolMix[i].Tokens != toolMix[j].Tokens {
			return toolMix[i].Tokens > toolMix[j].Tokens
		}
		return toolMix[i].Cost > toolMix[j].Cost
	})
	if len(toolMix) > 6 {
		toolMix = toolMix[:6]
	}

	// 5. Box 4: Model Breakdown
	var batches []ToolDays
	for _, res := range results {
		if len(res.Daily) > 0 {
			batches = append(batches, ToolDays{Tool: res.Name, Days: res.Daily})
		}
	}
	var modelBreakdown []FinOpsModelRow
	if len(batches) > 0 {
		mRep := AggregateModels(batches, ModelOptions{}, pricer)
		for i, m := range mRep.Rows {
			if i >= 6 {
				break
			}
			hitRate := todaySummary.CacheHitRate
			if m.Input+m.CacheRead > 0 {
				hitRate = (float64(m.CacheRead) / float64(m.Input+m.CacheRead)) * 100.0
			}
			calls := max(1, int(m.Tokens/2500))
			modelBreakdown = append(modelBreakdown, FinOpsModelRow{
				Name:        m.Model,
				Cost:        m.Cost,
				CachePct:    hitRate,
				Calls:       calls,
				OneShot:     "-",
				TokPerS:     "-",
				IsEstimated: true,
			})
		}
	}

	// 6. Box 5: Prompt Cache & Savings
	var cacheHitRate float64
	totalPromptInput := inputTokens + cacheRead
	if totalPromptInput > 0 {
		cacheHitRate = (float64(cacheRead) / float64(totalPromptInput)) * 100.0
	}
	estimatedSavings := (float64(cacheRead) / 1_000_000.0) * 2.20
	cacheMetrics := CacheMetrics{
		CacheReadTokens:  cacheRead,
		CacheWriteTokens: cacheWrite,
		InputTokens:      inputTokens,
		HitRate:          cacheHitRate,
		EstimatedSavings: estimatedSavings,
	}

	// 7. Box 6: Code Yield & Velocity
	if yieldRep.Totals.Tokens == 0 && (len(results) > 0 || len(projects) > 0) {
		yieldRep = AggregateYield(results, projects, YieldOptions{})
	}
	yieldRow := yieldRep.Totals
	if yieldRow.Tokens == 0 && totalTokens > 0 {
		yieldRow.Tokens = totalTokens
	}
	if yieldRow.Efficiency == "" {
		yieldRow.Efficiency = "[BALANCED]"
	}

	// 8. Box 7: Work Receipts & Tests
	var receiptSummary FinOpsReceiptSummary
	cachedReceiptsMu.Lock()
	if cachedReceiptsValid && cachedReceiptsPr == pricer {
		receiptSummary = cachedReceiptsSummary
		cachedReceiptsMu.Unlock()
	} else {
		cachedReceiptsMu.Unlock()
		receipts := ScanSessionReceipts(HomeDir(), "all", pricer)
		if len(receipts) > 0 {
			rRep := AggregateReceipts(receipts, ReceiptOptions{})
			var toolNames []string
			seenTool := make(map[string]bool)
			for _, r := range receipts {
				for _, ev := range r.EvidenceSummary {
					if !seenTool[ev] {
						seenTool[ev] = true
						toolNames = append(toolNames, ev)
					}
				}
			}
			toolsStr := strings.Join(toolNames, ", ")
			if len(toolsStr) > 30 {
				toolsStr = toolsStr[:27] + "..."
			}
			receiptSummary = FinOpsReceiptSummary{
				VerifiedCount:   rRep.Summary.VerifiedCount,
				ClaimedCount:    rRep.Summary.ClaimedCount,
				FailedCount:     rRep.Summary.FailedCount,
				TokensVerified:  rRep.Summary.TokensVerified,
				VerifiedRate:    rRep.Summary.VerificationRate,
				SpendEfficiency: rRep.Summary.SpendEfficiency,
				TestTools:       toolsStr,
			}
		} else {
			receiptSummary = FinOpsReceiptSummary{
				VerifiedCount:   0,
				ClaimedCount:    max(1, len(days)),
				FailedCount:     0,
				TokensVerified:  0,
				VerifiedRate:    0.0,
				SpendEfficiency: "EXPLORATORY",
				TestTools:       "None detected",
			}
		}
		cachedReceiptsMu.Lock()
		cachedReceiptsSummary = receiptSummary
		cachedReceiptsPr = pricer
		cachedReceiptsValid = true
		cachedReceiptsMu.Unlock()
	}

	// 9. Box 8: Subscription Replay
	var defaultReplayPlans = []SubscriptionPlan{
		{
			ID:              "claude-pro",
			Name:            "Claude Pro ($20)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      20.0,
			DailyTokenLimit: 15_000_000,
			DefaultModel:    "claude-3-5-sonnet",
		},
		{
			ID:              "claude-max",
			Name:            "Claude Max ($200)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      200.0,
			DailyTokenLimit: 75_000_000,
			DefaultModel:    "claude-3-5-sonnet",
		},
		{
			ID:              "chatgpt-plus",
			Name:            "ChatGPT Plus ($20)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      20.0,
			DailyTokenLimit: 20_000_000,
			DefaultModel:    "gpt-4o",
		},
		{
			ID:              "agy-pro",
			Name:            "Antigravity Pro ($20)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      20.0,
			DailyTokenLimit: 50_000_000,
			DefaultModel:    "gemini-2.5-pro",
		},
		{
			ID:              "opencode-pro",
			Name:            "OpenCode Pro ($20)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      20.0,
			DailyTokenLimit: 40_000_000,
			DefaultModel:    "claude-3-5-sonnet",
		},
		{
			ID:              "zcode-pro",
			Name:            "ZCode Pro ($20)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      20.0,
			DailyTokenLimit: 35_000_000,
			DefaultModel:    "glm-4-plus",
		},
		{
			ID:              "kimi-k2",
			Name:            "Kimi K2 Pro ($15)",
			Type:            PlanTypeSubscription,
			MonthlyFee:      15.0,
			DailyTokenLimit: 60_000_000,
			DefaultModel:    "kimi-k2",
		},
		{
			ID:              "deepseek-api",
			Name:            "DeepSeek V3 (API)",
			Type:            PlanTypePayAsYouGo,
			MonthlyFee:      0.0,
			DailyTokenLimit: 0,
			DefaultModel:    "deepseek-chat",
		},
	}
	var rp ReplayPricer
	if p, ok := pricer.(ReplayPricer); ok {
		rp = p
	}
	repReport := AggregateReplay(days, ReplayOptions{Last: 30, Compare: defaultReplayPlans}, rp)
	var replayPlans []FinOpsReplayPlan
	globalVerdict := "PASS"
	for _, p := range repReport.Plans {
		if p.CapacityVerdict == "FAIL" {
			globalVerdict = "FAIL"
		} else if p.CapacityVerdict == "DEGRADED" && globalVerdict != "FAIL" {
			globalVerdict = "DEGRADED"
		}
		replayPlans = append(replayPlans, FinOpsReplayPlan{
			Name:            p.Name,
			MonthlyCost:     p.MonthlyCost,
			CostDelta:       p.CostDelta,
			CapacityVerdict: p.CapacityVerdict,
			ThrottleRate:    p.ThrottleRate,
		})
	}
	replaySummary := FinOpsReplaySummary{
		CapacityVerdict: globalVerdict,
		Plans:           replayPlans,
	}

	// 10. Box 9: Token Composition
	var reasoningRatio float64
	if outputTokens+reasoningTokens > 0 {
		reasoningRatio = (float64(reasoningTokens) / float64(outputTokens+reasoningTokens)) * 100.0
	}
	compSummary := FinOpsCompositionSummary{
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		ReasoningTokens:  reasoningTokens,
		CacheReadTokens:  cacheRead,
		CacheWriteTokens: cacheWrite,
		ReasoningRatio:   reasoningRatio,
	}

	// Disjoint Token Shares for Backward Compatibility in Taxonomy
	var taxShares []ActivityShare
	type tokCat struct {
		name   ActivityCategory
		tokens int64
	}
	cats := []tokCat{
		{"Cache Read", cacheRead},
		{"Uncached Input", inputTokens},
		{"Output", outputTokens},
		{"Reasoning", reasoningTokens},
		{"Cache Write", cacheWrite},
	}
	for _, c := range cats {
		pct := 0.0
		if totalTokens > 0 {
			pct = (float64(c.tokens) / float64(totalTokens)) * 100.0
		}
		catCost := 0.0
		if totalTokens > 0 {
			catCost = totalCost * (float64(c.tokens) / float64(totalTokens))
		}
		taxShares = append(taxShares, ActivityShare{
			Category: c.name,
			Tokens:   c.tokens,
			Cost:     catCost,
			Turns:    int(float64(totalCalls) * (pct / 100.0)),
			Percent:  pct,
			OneShot:  "-",
		})
	}

	runRateDays := 30
	forecastTokens := avgDailyTokens * int64(runRateDays)
	forecastCost := avgDailyCost * float64(runRateDays)

	capacityVerdict := globalVerdict

	topModels := yieldRep.Models
	if len(topModels) > 5 {
		topModels = topModels[:5]
	}
	topProjects := yieldRep.Projects
	if len(topProjects) > 5 {
		topProjects = topProjects[:5]
	}

	return FinOpsGridPayload{
		Today:                 todaySummary,
		DailyHistory:          dailyHistory,
		TotalDaysScanned:      len(days),
		ProjectBreakdown:      projList,
		ToolMix:               toolMix,
		ModelBreakdown:        modelBreakdown,
		Yield:                 yieldRow,
		Receipt:               receiptSummary,
		Replay:                replaySummary,
		Composition:           compSummary,
		Taxonomy:              taxShares,
		TotalTokens:           totalTokens,
		TotalCost:             totalCost,
		ActiveDays:            activeDays,
		AvgDailyTokens:        avgDailyTokens,
		AvgDailyCost:          avgDailyCost,
		SpendEfficiency:       yieldRow.Efficiency,
		Cache:                 cacheMetrics,
		TopModels:             topModels,
		TopProjects:           topProjects,
		MonthEndRunRateTokens: forecastTokens,
		MonthEndRunRateCost:   forecastCost,
		CapacityVerdict:       capacityVerdict,
	}
}
