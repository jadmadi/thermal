// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package thermal

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// LiveEvent represents a single real-time token burn event captured during monitoring.
type LiveEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Tool        string    `json:"tool"`
	Tokens      int64     `json:"tokens"`
	Turns       int       `json:"turns,omitempty"`
	Model       string    `json:"model,omitempty"`
	Project     string    `json:"project,omitempty"`
	Cost        float64   `json:"cost"`
	IsEstimated bool      `json:"isEstimated,omitempty"`
}

// LiveSnapshot is the instantaneous state of live token burn, velocity, and recent events.
type LiveSnapshot struct {
	Timestamp         time.Time        `json:"timestamp"`
	TodayTokens       int64            `json:"todayTokens"`
	TodayCost         float64          `json:"todayCost"`
	TodayTurns        int              `json:"todayTurns"`
	TodayCacheHit     float64          `json:"todayCacheHit"`
	SessionTokens     int64            `json:"sessionTokens"`
	SessionCost       float64          `json:"sessionCost"`
	SessionTurns      int              `json:"sessionTurns"`
	BurnTokensPerMin  float64          `json:"burnTokensPerMin"`
	BurnTokensPerSec  float64          `json:"burnTokensPerSec"`
	BurnCostPerHr     float64          `json:"burnCostPerHr"`
	FlameIntensity    float64          `json:"flameIntensity"` // 0.0 to 1.0 (token flame)
	BurnTurnsPerMin   float64          `json:"burnTurnsPerMin,omitempty"`
	BurnTurnsPerSec   float64          `json:"burnTurnsPerSec,omitempty"`
	ActivityIntensity float64          `json:"activityIntensity,omitempty"` // 0.0 to 1.0 (turn/step flame)
	ActiveModel       string           `json:"activeModel,omitempty"`
	ActiveProject     string           `json:"activeProject,omitempty"`
	RecentEvents      []LiveEvent      `json:"recentEvents"`
	RollingTokens     [60]int64        `json:"rollingTokens"` // 60 seconds rolling buckets
	PeakRollingTok    int64            `json:"peakRollingTok"`
	RollingTurns      [60]int          `json:"rollingTurns,omitempty"` // 60 seconds rolling turns
	PeakRollingTurn   int              `json:"peakRollingTurn,omitempty"`
	ToolTotals        map[string]int64 `json:"toolTotals"`
}

// LiveTracker tracks incremental tool token deltas, rolling burn velocity, and flame intensity.
type LiveTracker struct {
	mu                sync.Mutex
	initialized       bool
	seedToday         bool
	todayDay          string
	pricer            Pricer
	noEstimate        bool
	prevTotals        map[string]int64   // tool -> lifetime tokens seen
	prevTurns         map[string]int     // tool -> lifetime turns seen
	prevCosts         map[string]float64 // tool -> lifetime costs seen
	prevModels        map[string]map[string]int64
	prevProjTokens    map[string]int64 // tool\x00project -> tokens seen
	prevProjTurns     map[string]int   // tool\x00project -> turns seen
	sessionTokens     int64
	sessionCost       float64
	sessionTurns      int
	burnBuckets       [60]int64 // rolling 60 seconds of tokens
	turnBuckets       [60]int   // rolling 60 seconds of turns
	bucketIndex       int
	lastBucketTime    time.Time
	flameIntensity    float64
	activityIntensity float64
	activeModel       string
	activeProject     string
	events            []LiveEvent
	maxEvents         int
}

// NewLiveTracker creates a tracker configured to track live token burn events.
func NewLiveTracker() *LiveTracker {
	return &LiveTracker{
		prevTotals:     make(map[string]int64),
		prevTurns:      make(map[string]int),
		prevCosts:      make(map[string]float64),
		prevModels:     make(map[string]map[string]int64),
		prevProjTokens: make(map[string]int64),
		prevProjTurns:  make(map[string]int),
		maxEvents:      30,
	}
}

// NewLiveTrackerWithPricer creates a tracker configured with an explicit pricer and estimation settings.
func NewLiveTrackerWithPricer(p Pricer, noEst bool) *LiveTracker {
	lt := NewLiveTracker()
	lt.pricer = p
	lt.noEstimate = noEst
	return lt
}

// SetPricer sets the pricing estimator for live cost tracking.
func (lt *LiveTracker) SetPricer(p Pricer) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.pricer = p
}

// SetNoEstimate controls whether unpriced tokens should have estimated cost computed.
func (lt *LiveTracker) SetNoEstimate(noEst bool) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.noEstimate = noEst
}

// SetSeedToday controls whether session counters initialize seeded with today's accumulated totals.
func (lt *LiveTracker) SetSeedToday(seed bool) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.seedToday = seed
}

// ResetSession zeroes session-accumulated tokens and costs without affecting lifetime baselines.
func (lt *LiveTracker) ResetSession() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.sessionTokens = 0
	lt.sessionCost = 0
	lt.sessionTurns = 0
}

func toolTotalTurns(res ToolResult) int {
	var tot int
	for _, d := range res.Daily {
		tot += d.Turns
	}
	if tot == 0 && res.Summary.Sessions > 0 {
		tot = res.Summary.Sessions
	}
	return tot
}

// Poll advances the time window, checks for tool token deltas, and returns an updated snapshot.
func (lt *LiveTracker) Poll(results []ToolResult, projDays []ProjectDay, now time.Time) (LiveSnapshot, []LiveEvent) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	currentDay := LocalDay(now)
	if lt.todayDay != "" && currentDay != lt.todayDay {
		// Midnight transition: day boundary crossed
		lt.todayDay = currentDay
		lt.prevProjTokens = make(map[string]int64)
		lt.prevProjTurns = make(map[string]int)
		for _, pd := range projDays {
			if pd.Day == currentDay {
				key := pd.Tool + "\x00" + pd.Project
				lt.prevProjTokens[key] = pd.Tokens
				lt.prevProjTurns[key] = pd.Turns
			}
		}
		lt.activeProject = ""
		lt.activeModel = ""
	} else if lt.todayDay == "" {
		lt.todayDay = currentDay
	}

	// 1. Advance rolling seconds & decay flame intensity
	if !lt.lastBucketTime.IsZero() {
		elapsed := now.Sub(lt.lastBucketTime)
		steps := int(elapsed.Seconds())
		if steps > 0 {
			if steps > 60 {
				steps = 60
			}
			for s := 0; s < steps; s++ {
				lt.bucketIndex = (lt.bucketIndex + 1) % 60
				lt.burnBuckets[lt.bucketIndex] = 0
				lt.turnBuckets[lt.bucketIndex] = 0
				lt.flameIntensity *= 0.85
				lt.activityIntensity *= 0.85
			}
			if lt.flameIntensity < 0.02 {
				lt.flameIntensity = 0
			}
			if lt.activityIntensity < 0.02 {
				lt.activityIntensity = 0
			}
			lt.lastBucketTime = now
		} else if steps < 0 {
			lt.lastBucketTime = now
		}
	} else {
		lt.lastBucketTime = now
	}

	// 2. Aggregate Today's volume across results
	var todayTok, todayIn, todayOut, todayCacheRead, todayCacheWrite int64
	var todayCost float64
	var todayTurns int
	toolTotals := make(map[string]int64)

	var effectivePricer Pricer
	if !lt.noEstimate {
		effectivePricer = lt.pricer
	}

	// Pre-build project display slugs
	var projPaths []string
	for _, pd := range projDays {
		if pd.Project != "" {
			projPaths = append(projPaths, pd.Project)
		}
	}
	projSlugs := ProjectDisplayNames(projPaths)

	for _, res := range results {
		toolTotals[res.Name] = res.Summary.LifetimeTokens
		isAct := isActivityTool(res.Tool)

		for _, d := range res.Daily {
			if d.Day == lt.todayDay {
				todayTurns += d.Turns

				if isAct || isActivityOnly(d) {
					continue
				}

				todayTok += d.Tokens
				if len(d.Models) > 0 {
					for _, m := range d.Models {
						todayIn += m.Input
						todayOut += m.Output
						todayCacheRead += m.CacheRead
						todayCacheWrite += m.CacheWrite
					}
				} else {
					todayIn += d.Input
					todayOut += d.Output
					todayCacheRead += d.Cache
					typedSum := d.Input + d.Output + d.Reasoning + d.Cache
					if d.Tokens > typedSum {
						todayIn += (d.Tokens - typedSum)
					}
				}

				c, _ := dayCost(d, effectivePricer)
				todayCost += c
			}
		}
	}

	var cacheHit float64
	promptTotal := todayIn + todayCacheRead + todayCacheWrite
	if promptTotal > 0 {
		cacheHit = float64(todayCacheRead) / float64(promptTotal) * 100.0
	}

	// 3. Delta detection
	var newEvents []LiveEvent

	if !lt.initialized {
		// First poll: initialize baselines so historical data is not treated as a live burst
		for _, res := range results {
			curTok := res.Summary.LifetimeTokens
			if isActivityTool(res.Tool) {
				curTok = 0
			}
			lt.prevTotals[res.Name] = curTok
			lt.prevTurns[res.Name] = toolTotalTurns(res)

			cost := res.Summary.Cost
			if cost == 0 && !lt.noEstimate {
				cost = res.EstimatedCost
			}
			lt.prevCosts[res.Name] = cost

			modMap := make(map[string]int64)
			for m, c := range res.Summary.ModelBreakdown {
				modMap[m] = c
			}
			lt.prevModels[res.Name] = modMap
		}

		// Initialize activeModel to today's active model across daily rows
		if lt.activeModel == "" {
			var bestDayCount int64
			for _, res := range results {
				for _, d := range res.Daily {
					if d.Day == lt.todayDay && len(d.Models) > 0 {
						for m, mt := range d.Models {
							tot := mt.Total()
							if tot > bestDayCount {
								bestDayCount = tot
								lt.activeModel = m
							}
						}
					}
				}
			}
			if lt.activeModel == "" {
				// Fallback to most recent daily row or ModelBreakdown
				var latestDay string
				var bestCount int64
				for _, res := range results {
					for _, d := range res.Daily {
						if (d.Day > latestDay || (d.Day == latestDay && d.Tokens > bestCount)) && len(d.Models) > 0 {
							for m, mt := range d.Models {
								if mt.Total() > bestCount {
									bestCount = mt.Total()
									latestDay = d.Day
									lt.activeModel = m
								}
							}
						}
					}
					if lt.activeModel == "" {
						var maxM int64
						for m, c := range res.Summary.ModelBreakdown {
							if c > maxM {
								maxM = c
								lt.activeModel = m
							}
						}
					}
				}
			}
		}

		// Initialize project baselines across all project days
		for _, pd := range projDays {
			key := pd.Tool + "\x00" + pd.Project
			lt.prevProjTokens[key] += pd.Tokens
			lt.prevProjTurns[key] += pd.Turns
		}

		// Initialize activeProject to today's active project across projDays
		if lt.activeProject == "" {
			var bestActivity int64
			for _, pd := range projDays {
				if pd.Day == lt.todayDay {
					slug := projSlugs[pd.Project]
					if slug == "" {
						slug = ProjectSlug(pd.Project)
					}
					act := pd.Tokens
					if act <= 0 {
						act = int64(pd.Turns)
					}
					if slug != "" && (act > bestActivity || lt.activeProject == "") {
						bestActivity = act
						lt.activeProject = slug
					}
				}
			}
			if lt.activeProject == "" && len(projDays) > 0 {
				var latestDay string
				var bestAct int64
				for _, pd := range projDays {
					if pd.Day > latestDay || (pd.Day == latestDay && pd.Tokens > bestAct) {
						latestDay = pd.Day
						bestAct = pd.Tokens
						slug := projSlugs[pd.Project]
						if slug == "" {
							slug = ProjectSlug(pd.Project)
						}
						if slug != "" {
							lt.activeProject = slug
						}
					}
				}
			}
		}

		if lt.seedToday {
			lt.sessionTokens = todayTok
			lt.sessionTurns = todayTurns
			lt.sessionCost = todayCost
			if todayTok > 0 {
				lt.flameIntensity = 0.3
			}
			if todayTurns > 0 {
				lt.activityIntensity = 0.3
			}

			// Seed baseline events for today's active tools so the event ticker is populated
			for _, res := range results {
				var dayTok int64
				var dayTurns int
				var dCost float64
				var toolModel string
				var toolProj string

				for _, d := range res.Daily {
					if d.Day == lt.todayDay {
						dayTurns += d.Turns
						if !isActivityTool(res.Tool) && !isActivityOnly(d) {
							dayTok += d.Tokens
							c, _ := dayCost(d, effectivePricer)
							dCost += c
						}
						if toolModel == "" && len(d.Models) > 0 {
							var maxM int64
							for m, mt := range d.Models {
								if mt.Total() > maxM {
									maxM = mt.Total()
									toolModel = m
								}
							}
						}
					}
				}
				if toolModel == "" {
					var maxM int64
					for m, c := range res.Summary.ModelBreakdown {
						if c > maxM {
							maxM = c
							toolModel = m
						}
					}
				}
				if toolModel == "" && len(res.Daily) > 0 {
					for i := len(res.Daily) - 1; i >= 0; i-- {
						d := res.Daily[i]
						if len(d.Models) > 0 {
							var maxM int64
							for m, mt := range d.Models {
								if mt.Total() > maxM {
									maxM = mt.Total()
									toolModel = m
								}
							}
							if toolModel != "" {
								break
							}
						}
					}
				}
				for _, pd := range projDays {
					if pd.Tool == res.Name && pd.Day == lt.todayDay {
						slug := projSlugs[pd.Project]
						if slug == "" {
							slug = ProjectSlug(pd.Project)
						}
						toolProj = slug
						break
					}
				}
				if toolProj == "" {
					var latestDay string
					for _, pd := range projDays {
						if pd.Tool == res.Name && pd.Day > latestDay {
							latestDay = pd.Day
							slug := projSlugs[pd.Project]
							if slug == "" {
								slug = ProjectSlug(pd.Project)
							}
							toolProj = slug
						}
					}
				}

				if dayTok > 0 || dayTurns > 0 {
					isEst := !lt.noEstimate && res.Summary.Cost == 0 && dCost > 0
					ev := LiveEvent{
						Timestamp:   now,
						Tool:        res.Name,
						Tokens:      dayTok,
						Turns:       dayTurns,
						Model:       toolModel,
						Project:     toolProj,
						Cost:        dCost,
						IsEstimated: isEst,
					}
					lt.events = append(lt.events, ev)
				}
			}
			if len(lt.events) > lt.maxEvents {
				lt.events = lt.events[:lt.maxEvents]
			}
		}
		lt.initialized = true
	} else {
		for _, res := range results {
			curTok := res.Summary.LifetimeTokens
			isAct := isActivityTool(res.Tool)
			if isAct {
				curTok = 0
			}
			curTurns := toolTotalTurns(res)
			curCost := res.Summary.Cost
			if curCost == 0 && !lt.noEstimate {
				curCost = res.EstimatedCost
			}

			prevTok, tokSeen := lt.prevTotals[res.Name]
			prevTurns, turnSeen := lt.prevTurns[res.Name]
			prevCost, costSeen := lt.prevCosts[res.Name]

			// First appearance of a newly discovered tool after initialization
			if !tokSeen && !turnSeen {
				lt.prevTotals[res.Name] = curTok
				lt.prevTurns[res.Name] = curTurns
				lt.prevCosts[res.Name] = curCost
				modMap := make(map[string]int64)
				for m, c := range res.Summary.ModelBreakdown {
					modMap[m] = c
				}
				lt.prevModels[res.Name] = modMap
				for _, pd := range projDays {
					if pd.Tool == res.Name {
						key := pd.Tool + "\x00" + pd.Project
						lt.prevProjTokens[key] += pd.Tokens
						lt.prevProjTurns[key] += pd.Turns
					}
				}
				continue
			}

			deltaTok := curTok - prevTok
			if deltaTok < 0 {
				// Counter reset or log truncation
				deltaTok = 0
				lt.prevTotals[res.Name] = curTok
			}

			deltaTurns := curTurns - prevTurns
			if deltaTurns < 0 {
				// Turn counter reset
				deltaTurns = 0
				lt.prevTurns[res.Name] = curTurns
			}

			var deltaCost float64
			if costSeen {
				deltaCost = curCost - prevCost
				if deltaCost < 0 {
					deltaCost = 0
					lt.prevCosts[res.Name] = curCost
				}
			}

			if deltaTok > 0 || deltaTurns > 0 {
				// Detect model that saw an increase
				var bestModel string
				var bestModelInc int64
				prevM := lt.prevModels[res.Name]
				for m, curCount := range res.Summary.ModelBreakdown {
					pCount := prevM[m]
					if curCount > pCount && (curCount-pCount) > bestModelInc {
						bestModelInc = curCount - pCount
						bestModel = m
					}
				}
				if bestModel == "" {
					// Check today's DailyRow models for this tool
					var todayMax int64
					for _, d := range res.Daily {
						if d.Day == lt.todayDay && len(d.Models) > 0 {
							for m, mt := range d.Models {
								tot := mt.Total()
								if tot > todayMax {
									todayMax = tot
									bestModel = m
								}
							}
						}
					}
				}
				if bestModel == "" && len(res.Daily) > 0 {
					// Fallback to most recent daily row's top model
					for i := len(res.Daily) - 1; i >= 0; i-- {
						d := res.Daily[i]
						if len(d.Models) > 0 {
							var maxCount int64
							for m, mt := range d.Models {
								if mt.Total() > maxCount {
									maxCount = mt.Total()
									bestModel = m
								}
							}
							if bestModel != "" {
								break
							}
						}
					}
				}
				if bestModel == "" {
					// Fallback to top model overall for this tool
					var maxM int64
					for m, c := range res.Summary.ModelBreakdown {
						if c > maxM {
							maxM = c
							bestModel = m
						}
					}
				}
				if bestModel == "" {
					bestModel = lt.activeModel
				}
				if bestModel != "" {
					lt.activeModel = bestModel
				}

				// Compute current cumulative project totals for this tool across all project days
				type projAccum struct {
					project string
					slug    string
					tokens  int64
					turns   int
				}
				toolProjTotals := make(map[string]*projAccum)
				for _, pd := range projDays {
					if pd.Tool == res.Name {
						key := pd.Tool + "\x00" + pd.Project
						acc := toolProjTotals[key]
						if acc == nil {
							slug := projSlugs[pd.Project]
							if slug == "" {
								slug = ProjectSlug(pd.Project)
							}
							acc = &projAccum{
								project: pd.Project,
								slug:    slug,
							}
							toolProjTotals[key] = acc
						}
						acc.tokens += pd.Tokens
						acc.turns += pd.Turns
					}
				}

				// Find which project(s) had deltas for this tool
				type projectDeltaInfo struct {
					project    string
					slug       string
					deltaTok   int64
					deltaTurns int
				}
				var pDeltas []projectDeltaInfo
				var totalProjectTokDelta int64
				var totalProjectTurnDelta int

				for key, acc := range toolProjTotals {
					pPrevTok, tokSeen := lt.prevProjTokens[key]
					pPrevTurns, turnSeen := lt.prevProjTurns[key]

					var pDTok int64
					var pDTurns int
					if tokSeen {
						pDTok = acc.tokens - pPrevTok
					} else {
						pDTok = acc.tokens
						if pDTok > deltaTok {
							pDTok = deltaTok
						}
					}
					if turnSeen {
						pDTurns = acc.turns - pPrevTurns
					} else {
						pDTurns = acc.turns
						if pDTurns > deltaTurns {
							pDTurns = deltaTurns
						}
					}

					if pDTok < 0 {
						pDTok = 0
					}
					if pDTurns < 0 {
						pDTurns = 0
					}

					if pDTok > 0 || pDTurns > 0 {
						pDeltas = append(pDeltas, projectDeltaInfo{
							project:    acc.project,
							slug:       acc.slug,
							deltaTok:   pDTok,
							deltaTurns: pDTurns,
						})
						totalProjectTokDelta += pDTok
						totalProjectTurnDelta += pDTurns
					}

					// Update project baseline
					lt.prevProjTokens[key] = acc.tokens
					lt.prevProjTurns[key] = acc.turns
				}

				if len(pDeltas) > 1 {
					sort.Slice(pDeltas, func(i, j int) bool {
						if pDeltas[i].deltaTok != pDeltas[j].deltaTok {
							return pDeltas[i].deltaTok > pDeltas[j].deltaTok
						}
						return pDeltas[i].slug < pDeltas[j].slug
					})
				}

				isEst := !lt.noEstimate && res.Summary.Cost == 0 && res.EstimatedCost > 0

				// If specific project deltas were identified, emit events for each active project
				if len(pDeltas) > 0 {
					for _, pd := range pDeltas {
						tokShare := pd.deltaTok
						if tokShare <= 0 && deltaTok > 0 && totalProjectTurnDelta > 0 {
							tokShare = deltaTok * int64(pd.deltaTurns) / int64(totalProjectTurnDelta)
						}
						if tokShare <= 0 && deltaTok > 0 && len(pDeltas) == 1 {
							tokShare = deltaTok
						}
						costShare := 0.0
						if deltaTok > 0 && tokShare > 0 {
							costShare = deltaCost * (float64(tokShare) / float64(deltaTok))
						} else if len(pDeltas) == 1 {
							costShare = deltaCost
						}

						ev := LiveEvent{
							Timestamp:   now,
							Tool:        res.Name,
							Tokens:      tokShare,
							Turns:       pd.deltaTurns,
							Model:       bestModel,
							Project:     pd.slug,
							Cost:        costShare,
							IsEstimated: isEst,
						}
						newEvents = append(newEvents, ev)
						lt.events = append([]LiveEvent{ev}, lt.events...)
						if len(lt.events) > lt.maxEvents {
							lt.events = lt.events[:lt.maxEvents]
						}
						lt.activeProject = pd.slug
					}

					// Remainder unattributed to any project
					if deltaTok > totalProjectTokDelta && totalProjectTokDelta > 0 {
						remTok := deltaTok - totalProjectTokDelta
						remCost := deltaCost * (float64(remTok) / float64(deltaTok))
						ev := LiveEvent{
							Timestamp:   now,
							Tool:        res.Name,
							Tokens:      remTok,
							Model:       bestModel,
							Project:     "",
							Cost:        remCost,
							IsEstimated: isEst,
						}
						newEvents = append(newEvents, ev)
						lt.events = append([]LiveEvent{ev}, lt.events...)
						if len(lt.events) > lt.maxEvents {
							lt.events = lt.events[:lt.maxEvents]
						}
					}
				} else {
					// Tool delta without specific project delta
					ev := LiveEvent{
						Timestamp:   now,
						Tool:        res.Name,
						Tokens:      deltaTok,
						Turns:       deltaTurns,
						Model:       bestModel,
						Project:     "",
						Cost:        deltaCost,
						IsEstimated: isEst,
					}
					newEvents = append(newEvents, ev)
					lt.events = append([]LiveEvent{ev}, lt.events...)
					if len(lt.events) > lt.maxEvents {
						lt.events = lt.events[:lt.maxEvents]
					}
				}

				lt.sessionTokens += deltaTok
				lt.sessionCost += deltaCost
				lt.sessionTurns += deltaTurns

				if deltaTok > 0 {
					lt.burnBuckets[lt.bucketIndex] += deltaTok
					lt.flameIntensity = math.Min(1.0, lt.flameIntensity+0.5)
				}
				if deltaTurns > 0 {
					lt.turnBuckets[lt.bucketIndex] += deltaTurns
					lt.activityIntensity = math.Min(1.0, lt.activityIntensity+0.4)
				}
				if bestModel != "" {
					lt.activeModel = bestModel
				}

				// Update baselines
				lt.prevTotals[res.Name] = curTok
				lt.prevTurns[res.Name] = curTurns
				lt.prevCosts[res.Name] = curCost

				modMap := make(map[string]int64)
				for m, c := range res.Summary.ModelBreakdown {
					modMap[m] = c
				}
				lt.prevModels[res.Name] = modMap
			} else {
				// Keep project baselines updated even without tool-level delta
				for _, pd := range projDays {
					if pd.Tool == res.Name && pd.Day == lt.todayDay {
						key := pd.Tool + "\x00" + pd.Project
						lt.prevProjTokens[key] = pd.Tokens
						lt.prevProjTurns[key] = pd.Turns
					}
				}
			}
		}
	}

	// 4. Calculate rolling 60s burn velocity (tokens and turns)
	var sumRolling, peakRolling int64
	for _, b := range lt.burnBuckets {
		sumRolling += b
		if b > peakRolling {
			peakRolling = b
		}
	}

	burnPerMin := float64(sumRolling)
	burnPerSec := burnPerMin / 60.0

	var sumTurns, peakTurns int
	for _, b := range lt.turnBuckets {
		sumTurns += b
		if b > peakTurns {
			peakTurns = b
		}
	}

	turnPerMin := float64(sumTurns)
	turnPerSec := turnPerMin / 60.0

	// Estimate cost per hour based on session spend rate or typical catalog blend ($3/M)
	var burnCostHr float64
	if lt.sessionTokens > 0 && lt.sessionCost > 0 {
		costPerTok := lt.sessionCost / float64(lt.sessionTokens)
		burnCostHr = burnPerMin * 60.0 * costPerTok
	} else if todayTok > 0 {
		// Blended assumption: ~$3.50 per million tokens
		burnCostHr = (burnPerMin * 60.0 / 1_000_000.0) * 3.50
	}

	snap := LiveSnapshot{
		Timestamp:         now,
		TodayTokens:       todayTok,
		TodayCost:         todayCost,
		TodayTurns:        todayTurns,
		TodayCacheHit:     cacheHit,
		SessionTokens:     lt.sessionTokens,
		SessionCost:       lt.sessionCost,
		SessionTurns:      lt.sessionTurns,
		BurnTokensPerMin:  burnPerMin,
		BurnTokensPerSec:  burnPerSec,
		BurnCostPerHr:     burnCostHr,
		FlameIntensity:    lt.flameIntensity,
		BurnTurnsPerMin:   turnPerMin,
		BurnTurnsPerSec:   turnPerSec,
		ActivityIntensity: lt.activityIntensity,
		ActiveModel:       lt.activeModel,
		ActiveProject:     lt.activeProject,
		RecentEvents:      append([]LiveEvent{}, lt.events...),
		RollingTokens:     lt.burnBuckets,
		PeakRollingTok:    peakRolling,
		RollingTurns:      lt.turnBuckets,
		PeakRollingTurn:   peakTurns,
		ToolTotals:        toolTotals,
	}

	return snap, newEvents
}

// Sparkline60 renders an adaptive UTF-8 sparkline string from 60 seconds rolling tokens.
func Sparkline60(buckets [60]int64, currentIndex int) string {
	return Sparkline(buckets, currentIndex, 60)
}

// Sparkline renders an adaptive UTF-8 sparkline string of up to count seconds.
func Sparkline(buckets [60]int64, currentIndex int, count int) string {
	if count <= 0 {
		return ""
	}
	if count > 60 {
		count = 60
	}
	var maxVal int64
	for _, v := range buckets {
		if v > maxVal {
			maxVal = v
		}
	}

	levels := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var sb strings.Builder

	// Render chronologically from oldest to newest for requested count
	startOffset := 60 - count + 1
	for i := startOffset; i <= 60; i++ {
		idx := (currentIndex + i) % 60
		val := buckets[idx]
		if val <= 0 || maxVal <= 0 {
			sb.WriteRune(' ')
			continue
		}
		ratio := float64(val) / float64(maxVal)
		lIdx := int(math.Floor(ratio * float64(len(levels)-1)))
		if lIdx < 0 {
			lIdx = 0
		}
		if lIdx >= len(levels) {
			lIdx = len(levels) - 1
		}
		sb.WriteRune(levels[lIdx])
	}

	return sb.String()
}

// FlameBar renders an ASCII/Unicode flame meter of a given cell width.
func FlameBar(intensity float64, width int) (bar string, percent int) {
	if width <= 0 {
		return "", 0
	}
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1.0 {
		intensity = 1.0
	}

	percent = int(math.Round(intensity * 100))
	filled := int(math.Round(intensity * float64(width)))
	if filled > width {
		filled = width
	}
	empty := width - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty), percent
}
