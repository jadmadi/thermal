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

const (
	DefaultFlameCeilingTokens = 9_000_000
	DefaultFlameCeilingSteps  = 60
)

// LiveCostKind classifies how a live event's cost was obtained.
type LiveCostKind string

const (
	LiveCostRecorded  LiveCostKind = "recorded"
	LiveCostEstimated LiveCostKind = "estimated"
	LiveCostUnpriced  LiveCostKind = "unpriced"
)

// LiveEvent represents a single real-time token burn event captured during monitoring.
type LiveEvent struct {
	Timestamp   time.Time    `json:"timestamp"`
	Tool        string       `json:"tool"`
	Tokens      int64        `json:"tokens"`
	Turns       int          `json:"turns,omitempty"`
	Model       string       `json:"model,omitempty"`
	Project     string       `json:"project,omitempty"`
	Cost        float64      `json:"cost"`
	CostKind    LiveCostKind `json:"costKind,omitempty"`
	IsEstimated bool         `json:"isEstimated,omitempty"`
	IsBaseline  bool         `json:"isBaseline,omitempty"`
}

// LiveRollup is a today-scoped project or model aggregate for the idle view.
type LiveRollup struct {
	Name     string       `json:"name"`
	Tokens   int64        `json:"tokens"`
	Turns    int          `json:"turns,omitempty"`
	Cost     float64      `json:"cost"`
	CostKind LiveCostKind `json:"costKind,omitempty"`
}

// LiveEventGroup collapses repeated bursts for one tool/model/project into a
// single decision row.
type LiveEventGroup struct {
	Tool        string
	Model       string
	Project     string
	Count       int
	Latest      time.Time
	Tokens      int64
	Turns       int
	Cost        float64
	CostKind    LiveCostKind
	AnyUnpriced bool
	AnyOutlier  bool
}

// LiveSnapshot is the instantaneous state of live token burn, velocity, and recent events.
type LiveSnapshot struct {
	Timestamp        time.Time        `json:"timestamp"`
	TodayTokens      int64            `json:"todayTokens"`
	TodayCost        float64          `json:"todayCost"`
	TodayTurns       int              `json:"todayTurns"`
	TodayCacheHit    float64          `json:"todayCacheHit"`
	CarriedInTokens  int64            `json:"carriedInTokens"`
	CarriedInTurns   int              `json:"carriedInTurns"`
	CarriedInCost    float64          `json:"carriedInCost"`
	CarriedInAt      time.Time        `json:"carriedInAt"`
	SessionTokens    int64            `json:"sessionTokens"`
	SessionCost      float64          `json:"sessionCost"`
	SessionTurns     int              `json:"sessionTurns"`
	UnpricedTokens   int64            `json:"unpricedTokens"`
	UnpricedEvents   int              `json:"unpricedEvents"`
	BurnTokensPerMin float64          `json:"burnTokensPerMin"`
	BurnTokensPerSec float64          `json:"burnTokensPerSec"`
	BurnCostPerHr    float64          `json:"burnCostPerHr"`
	FlameIntensity   float64          `json:"flameIntensity"` // 0.0 to 1.0 of FlameCeiling
	FlameMetric      string           `json:"flameMetric"`    // tokens or steps
	FlameCeiling     float64          `json:"flameCeiling"`
	BurnTurnsPerMin  float64          `json:"burnTurnsPerMin,omitempty"`
	BurnTurnsPerSec  float64          `json:"burnTurnsPerSec,omitempty"`
	ActiveModel      string           `json:"activeModel,omitempty"`
	ActiveProject    string           `json:"activeProject,omitempty"`
	TodayProjects    []LiveRollup     `json:"todayProjects,omitempty"`
	TodayModels      []LiveRollup     `json:"todayModels,omitempty"`
	RecentEvents     []LiveEvent      `json:"recentEvents"`
	RollingTokens    [60]int64        `json:"rollingTokens"` // one-second buckets
	PeakTokensPerSec int64            `json:"peakTokensPerSec"`
	RollingTurns     [60]int          `json:"rollingTurns,omitempty"` // one-second buckets
	PeakStepsPerSec  int              `json:"peakStepsPerSec"`
	BucketSeconds    int              `json:"bucketSeconds"`
	WindowSeconds    int              `json:"windowSeconds"`
	ToolTotals       map[string]int64 `json:"toolTotals"`
}

// LiveTracker tracks incremental tool token deltas, rolling burn velocity, and flame intensity.
type LiveTracker struct {
	mu              sync.Mutex
	initialized     bool
	todayDay        string
	pricer          Pricer
	noEstimate      bool
	prevTotals      map[string]int64   // tool -> lifetime tokens seen
	prevTurns       map[string]int     // tool -> lifetime turns seen
	prevCosts       map[string]float64 // tool -> lifetime costs seen
	prevModels      map[string]map[string]int64
	prevProjTokens  map[string]int64 // tool\x00project -> tokens seen
	prevProjTurns   map[string]int   // tool\x00project -> turns seen
	carriedInTokens int64
	carriedInTurns  int
	carriedInCost   float64
	carriedInAt     time.Time
	sessionTokens   int64
	sessionCost     float64
	sessionTurns    int
	unpricedTokens  int64
	unpricedEvents  int
	todayProjects   []LiveRollup
	todayModels     []LiveRollup
	burnBuckets     [60]int64 // rolling 60 seconds of tokens
	turnBuckets     [60]int   // rolling 60 seconds of turns
	bucketIndex     int
	lastBucketTime  time.Time
	flameCeiling    float64
	activeModel     string
	activeProject   string
	events          []LiveEvent
	maxEvents       int
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

// SetFlameCeiling sets the flame ceiling in the active metric's units per
// minute. A non-positive value keeps the metric-specific default.
func (lt *LiveTracker) SetFlameCeiling(ceiling float64) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.flameCeiling = ceiling
}

// ResetSession zeroes session-accumulated tokens and costs without affecting lifetime baselines.
func (lt *LiveTracker) ResetSession() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.sessionTokens = 0
	lt.sessionCost = 0
	lt.sessionTurns = 0
	lt.unpricedTokens = 0
	lt.unpricedEvents = 0
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

func (lt *LiveTracker) addRollingDelta(tokens int64, turns int, elapsed time.Duration) {
	if tokens <= 0 && turns <= 0 {
		return
	}
	seconds := int(elapsed / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	windowTokens := tokens
	windowTurns := turns
	if seconds > 60 {
		windowTokens = tokens * 60 / int64(seconds)
		windowTurns = turns * 60 / seconds
		seconds = 60
	}
	baseTokens := windowTokens / int64(seconds)
	remTokens := windowTokens % int64(seconds)
	baseTurns := windowTurns / seconds
	remTurns := windowTurns % seconds
	for i := 0; i < seconds; i++ {
		idx := (lt.bucketIndex - i + 60) % 60
		addTokens := baseTokens
		if int64(i) < remTokens {
			addTokens++
		}
		lt.burnBuckets[idx] += addTokens
		addTurns := baseTurns
		if i < remTurns {
			addTurns++
		}
		lt.turnBuckets[idx] += addTurns
	}
}

func addRollupCost(r *LiveRollup, cost float64, estimated bool, tokenBearing bool) {
	r.Cost += cost
	if cost > 0 && estimated {
		r.CostKind = LiveCostEstimated
	} else if cost > 0 && r.CostKind != LiveCostEstimated {
		r.CostKind = LiveCostRecorded
	}
	if r.CostKind == "" && tokenBearing {
		r.CostKind = LiveCostUnpriced
	}
}

func sortedLiveRollups(m map[string]*LiveRollup, limit int) []LiveRollup {
	out := make([]LiveRollup, 0, len(m))
	for _, r := range m {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cost != out[j].Cost {
			return out[i].Cost > out[j].Cost
		}
		if out[i].Tokens != out[j].Tokens {
			return out[i].Tokens > out[j].Tokens
		}
		return out[i].Name < out[j].Name
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func buildTodayRollups(results []ToolResult, projDays []ProjectDay, todayDay string, projSlugs map[string]string, pricer Pricer) ([]LiveRollup, []LiveRollup) {
	projects := make(map[string]*LiveRollup)
	for _, pd := range projDays {
		if pd.Day != todayDay {
			continue
		}
		slug := projSlugs[pd.Project]
		if slug == "" {
			slug = ProjectSlug(pd.Project)
		}
		if slug == "" {
			continue
		}
		r := projects[slug]
		if r == nil {
			r = &LiveRollup{Name: slug}
			projects[slug] = r
		}
		r.Tokens += pd.Tokens
		r.Turns += pd.Turns
		cost, estimated := dayCost(DailyRow{
			Day:       pd.Day,
			Tokens:    pd.Tokens,
			Input:     pd.Input,
			Output:    pd.Output,
			Reasoning: pd.Reasoning,
			Cache:     pd.CacheRead + pd.CacheWrite,
			Cost:      pd.Cost,
			Models:    pd.Models,
		}, pricer)
		addRollupCost(r, cost, estimated, pd.Tokens > 0)
	}

	models := make(map[string]*LiveRollup)
	for _, res := range results {
		for _, day := range res.Daily {
			if day.Day != todayDay {
				continue
			}
			for name, mt := range day.Models {
				tokens := mt.Total()
				if tokens == 0 {
					continue
				}
				r := models[name]
				if r == nil {
					r = &LiveRollup{Name: name}
					models[name] = r
				}
				r.Tokens += tokens
				if pricer != nil {
					cost, missing := pricer.PriceDay(DailyRow{
						Tokens: tokens,
						Models: map[string]ModelTokens{name: mt},
					})
					if cost > 0 {
						addRollupCost(r, cost, true, true)
					} else if len(missing) > 0 || cost == 0 {
						addRollupCost(r, 0, false, true)
					}
				} else {
					addRollupCost(r, 0, false, true)
				}
			}
		}
	}

	return sortedLiveRollups(projects, 3), sortedLiveRollups(models, 3)
}

// Poll advances the time window, checks for tool token deltas, and returns an updated snapshot.
func (lt *LiveTracker) Poll(results []ToolResult, projDays []ProjectDay, now time.Time) (LiveSnapshot, []LiveEvent) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	wasInitialized := lt.initialized
	dayChanged := false
	currentDay := LocalDay(now)
	if lt.todayDay != "" && currentDay != lt.todayDay {
		// Midnight transition: day boundary crossed
		lt.todayDay = currentDay
		lt.activeProject = ""
		lt.activeModel = ""
		dayChanged = true
	} else if lt.todayDay == "" {
		lt.todayDay = currentDay
	}

	// 1. Advance one-second buckets and decay flame intensity.
	elapsed := time.Duration(0)
	if !lt.lastBucketTime.IsZero() {
		elapsed = now.Sub(lt.lastBucketTime)
		rawSteps := int(elapsed / time.Second)
		steps := rawSteps
		if steps > 60 {
			steps = 60
		}
		if steps > 0 {
			for s := 0; s < steps; s++ {
				lt.bucketIndex = (lt.bucketIndex + 1) % 60
				lt.burnBuckets[lt.bucketIndex] = 0
				lt.turnBuckets[lt.bucketIndex] = 0
			}
			if rawSteps > 60 {
				lt.lastBucketTime = now
			} else {
				lt.lastBucketTime = lt.lastBucketTime.Add(time.Duration(rawSteps) * time.Second)
			}
		} else if rawSteps < 0 {
			lt.lastBucketTime = now
			elapsed = 0
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

	// Capture the opening balance once at launch and again on a local-day
	// rollover. It is display history, never a live delta.
	if !lt.initialized || dayChanged {
		lt.carriedInTokens = todayTok
		lt.carriedInTurns = todayTurns
		lt.carriedInCost = todayCost
		lt.carriedInAt = now
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

		lt.initialized = true
	} else {
		for _, res := range results {
			curTok := res.Summary.LifetimeTokens
			if isActivityTool(res.Tool) {
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
				toolProjTokens := make(map[string]int64)
				toolProjTurns := make(map[string]int)
				for _, pd := range projDays {
					if pd.Tool == res.Name {
						key := pd.Tool + "\x00" + pd.Project
						toolProjTokens[key] += pd.Tokens
						toolProjTurns[key] += pd.Turns
					}
				}
				for key, tok := range toolProjTokens {
					lt.prevProjTokens[key] = tok
					lt.prevProjTurns[key] = toolProjTurns[key]
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
					}
					if turnSeen {
						pDTurns = acc.turns - pPrevTurns
					} else {
						pDTurns = acc.turns
					}

					if pDTok < 0 {
						pDTok = 0
					}
					if pDTok > deltaTok {
						pDTok = deltaTok
					}
					if pDTurns < 0 {
						pDTurns = 0
					}
					if pDTurns > deltaTurns {
						pDTurns = deltaTurns
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
						if pDeltas[i].deltaTurns != pDeltas[j].deltaTurns {
							return pDeltas[i].deltaTurns > pDeltas[j].deltaTurns
						}
						return pDeltas[i].slug < pDeltas[j].slug
					})
				}

				if totalProjectTokDelta > deltaTok || totalProjectTurnDelta > deltaTurns {
					var clamped []projectDeltaInfo
					var allocTok int64
					var allocTurns int
					for _, pd := range pDeltas {
						availTok := deltaTok - allocTok
						availTurns := deltaTurns - allocTurns
						if availTok <= 0 && availTurns <= 0 {
							break
						}
						tokToAssign := pd.deltaTok
						if tokToAssign > availTok {
							tokToAssign = availTok
						}
						if tokToAssign < 0 {
							tokToAssign = 0
						}
						turnsToAssign := pd.deltaTurns
						if turnsToAssign > availTurns {
							turnsToAssign = availTurns
						}
						if turnsToAssign < 0 {
							turnsToAssign = 0
						}
						if tokToAssign > 0 || turnsToAssign > 0 {
							clamped = append(clamped, projectDeltaInfo{
								project:    pd.project,
								slug:       pd.slug,
								deltaTok:   tokToAssign,
								deltaTurns: turnsToAssign,
							})
							allocTok += tokToAssign
							allocTurns += turnsToAssign
						}
					}
					pDeltas = clamped
					totalProjectTokDelta = allocTok
					totalProjectTurnDelta = allocTurns
				}

				isEst := !lt.noEstimate && res.Summary.Cost == 0 && res.EstimatedCost > 0
				costKind := LiveCostUnpriced
				if res.Summary.Cost > 0 {
					costKind = LiveCostRecorded
				} else if isEst {
					costKind = LiveCostEstimated
				}

				// If specific project deltas were identified, emit events for each active project
				if len(pDeltas) > 0 {
					var emittedTurns int
					var emittedTok int64
					for _, pd := range pDeltas {
						tokShare := pd.deltaTok
						if tokShare <= 0 && deltaTok > 0 && totalProjectTurnDelta > 0 && totalProjectTokDelta == 0 {
							tokShare = deltaTok * int64(pd.deltaTurns) / int64(totalProjectTurnDelta)
						}
						if tokShare <= 0 && deltaTok > 0 && len(pDeltas) == 1 {
							tokShare = deltaTok
						}
						turnShare := pd.deltaTurns
						if turnShare <= 0 && deltaTurns > 0 && len(pDeltas) == 1 {
							turnShare = deltaTurns
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
							Turns:       turnShare,
							Model:       bestModel,
							Project:     pd.slug,
							Cost:        costShare,
							CostKind:    costKind,
							IsEstimated: costKind == LiveCostEstimated,
						}
						newEvents = append(newEvents, ev)
						lt.events = append([]LiveEvent{ev}, lt.events...)
						if len(lt.events) > lt.maxEvents {
							lt.events = lt.events[:lt.maxEvents]
						}
						lt.activeProject = pd.slug
						emittedTurns += turnShare
						emittedTok += tokShare
					}

					// Remainder unattributed to any project
					remTok := deltaTok - emittedTok
					if remTok < 0 {
						remTok = 0
					}
					remTurns := deltaTurns - emittedTurns
					if remTurns < 0 {
						remTurns = 0
					}
					if remTok > 0 || remTurns > 0 {
						remCost := 0.0
						if deltaTok > 0 && remTok > 0 {
							remCost = deltaCost * (float64(remTok) / float64(deltaTok))
						}
						ev := LiveEvent{
							Timestamp:   now,
							Tool:        res.Name,
							Tokens:      remTok,
							Turns:       remTurns,
							Model:       bestModel,
							Project:     "",
							Cost:        remCost,
							CostKind:    costKind,
							IsEstimated: costKind == LiveCostEstimated,
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
						CostKind:    costKind,
						IsEstimated: costKind == LiveCostEstimated,
					}
					newEvents = append(newEvents, ev)
					lt.events = append([]LiveEvent{ev}, lt.events...)
					if len(lt.events) > lt.maxEvents {
						lt.events = lt.events[:lt.maxEvents]
					}
				}

				if costKind == LiveCostUnpriced && deltaTok > 0 {
					lt.unpricedTokens += deltaTok
					lt.unpricedEvents++
				}

				lt.sessionTokens += deltaTok
				lt.sessionCost += deltaCost
				lt.sessionTurns += deltaTurns

				lt.addRollingDelta(deltaTok, deltaTurns, elapsed)
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
				toolProjTokens := make(map[string]int64)
				toolProjTurns := make(map[string]int)
				for _, pd := range projDays {
					if pd.Tool == res.Name {
						key := pd.Tool + "\x00" + pd.Project
						toolProjTokens[key] += pd.Tokens
						toolProjTurns[key] += pd.Turns
					}
				}
				for key, tok := range toolProjTokens {
					lt.prevProjTokens[key] = tok
					lt.prevProjTurns[key] = toolProjTurns[key]
				}
			}
		}
	}

	if !wasInitialized || dayChanged || len(newEvents) > 0 {
		lt.todayProjects, lt.todayModels = buildTodayRollups(results, projDays, lt.todayDay, projSlugs, effectivePricer)
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

	flameMetric := "tokens"
	flameRate := burnPerMin
	flameCeiling := float64(DefaultFlameCeilingTokens)
	if burnPerMin <= 0 && turnPerMin > 0 {
		flameMetric = "steps"
		flameRate = turnPerMin
		flameCeiling = DefaultFlameCeilingSteps
	} else if burnPerMin <= 0 && turnPerMin <= 0 && todayTok == 0 && todayTurns > 0 {
		flameMetric = "steps"
		flameRate = 0
		flameCeiling = DefaultFlameCeilingSteps
	}
	if lt.flameCeiling > 0 {
		flameCeiling = lt.flameCeiling
	}
	flameIntensity := 0.0
	if flameCeiling > 0 {
		flameIntensity = math.Min(1.0, flameRate/flameCeiling)
	}

	snap := LiveSnapshot{
		Timestamp:        now,
		TodayTokens:      todayTok,
		TodayCost:        todayCost,
		TodayTurns:       todayTurns,
		TodayCacheHit:    cacheHit,
		CarriedInTokens:  lt.carriedInTokens,
		CarriedInTurns:   lt.carriedInTurns,
		CarriedInCost:    lt.carriedInCost,
		CarriedInAt:      lt.carriedInAt,
		SessionTokens:    lt.sessionTokens,
		SessionCost:      lt.sessionCost,
		SessionTurns:     lt.sessionTurns,
		UnpricedTokens:   lt.unpricedTokens,
		UnpricedEvents:   lt.unpricedEvents,
		BurnTokensPerMin: burnPerMin,
		BurnTokensPerSec: burnPerSec,
		BurnCostPerHr:    burnCostHr,
		FlameIntensity:   flameIntensity,
		FlameMetric:      flameMetric,
		FlameCeiling:     flameCeiling,
		BurnTurnsPerMin:  turnPerMin,
		BurnTurnsPerSec:  turnPerSec,
		ActiveModel:      lt.activeModel,
		ActiveProject:    lt.activeProject,
		TodayProjects:    lt.todayProjects,
		TodayModels:      lt.todayModels,
		RecentEvents:     append([]LiveEvent{}, lt.events...),
		RollingTokens:    lt.burnBuckets,
		PeakTokensPerSec: peakRolling,
		RollingTurns:     lt.turnBuckets,
		PeakStepsPerSec:  peakTurns,
		BucketSeconds:    1,
		WindowSeconds:    60,
		ToolTotals:       toolTotals,
	}

	return snap, newEvents
}

func eventCostKind(ev LiveEvent) LiveCostKind {
	if ev.CostKind != "" {
		return ev.CostKind
	}
	if ev.IsEstimated {
		return LiveCostEstimated
	}
	if ev.Cost > 0 {
		return LiveCostRecorded
	}
	return LiveCostUnpriced
}

func mergeLiveEventGroup(g *LiveEventGroup, ev LiveEvent) {
	if g.Count == 0 {
		g.Tool = ev.Tool
		g.Model = ev.Model
		g.Project = ev.Project
	}
	g.Count++
	g.Tokens += ev.Tokens
	g.Turns += ev.Turns
	g.Cost += ev.Cost
	if ev.Timestamp.After(g.Latest) {
		g.Latest = ev.Timestamp
	}
	kind := eventCostKind(ev)
	if kind == LiveCostUnpriced {
		g.AnyUnpriced = true
	}
	if kind == LiveCostEstimated {
		g.CostKind = LiveCostEstimated
	} else if kind == LiveCostRecorded && g.CostKind != LiveCostEstimated {
		g.CostKind = LiveCostRecorded
	}
	if g.CostKind == "" {
		g.CostKind = kind
	}
}

// GroupLiveEvents groups non-baseline events by tool, model, and project when
// they fall within window of the newest event. Older events remain individual
// rows, and baseline rows are excluded.
func GroupLiveEvents(events []LiveEvent, window time.Duration, now time.Time) []LiveEventGroup {
	if now.IsZero() {
		now = time.Now()
	}
	nonBaseline := make([]LiveEvent, 0, len(events))
	newest := time.Time{}
	for _, ev := range events {
		if ev.IsBaseline {
			continue
		}
		if ev.Timestamp.After(now) {
			ev.Timestamp = now
		}
		nonBaseline = append(nonBaseline, ev)
		if ev.Timestamp.After(newest) {
			newest = ev.Timestamp
		}
	}
	if len(nonBaseline) == 0 {
		return nil
	}

	groups := make([]LiveEventGroup, 0, len(nonBaseline))
	index := make(map[string]int)
	for _, ev := range nonBaseline {
		withinWindow := window > 0 && newest.Sub(ev.Timestamp) <= window
		if withinWindow {
			key := ev.Tool + "\x00" + ev.Model + "\x00" + ev.Project
			if i, ok := index[key]; ok {
				mergeLiveEventGroup(&groups[i], ev)
				continue
			}
			index[key] = len(groups)
		}
		g := LiveEventGroup{}
		mergeLiveEventGroup(&g, ev)
		groups = append(groups, g)
	}
	return groups
}

// MedianOutlierFlags marks values strictly greater than factor times the median
// of the positive values. It returns no flags until minSamples positive values
// exist, so a two-row ticker never labels one row as anomalous.
func MedianOutlierFlags(values []int64, factor float64, minSamples int) []bool {
	flags := make([]bool, len(values))
	if factor <= 0 || minSamples <= 0 {
		return flags
	}
	positives := make([]int64, 0, len(values))
	for _, v := range values {
		if v > 0 {
			positives = append(positives, v)
		}
	}
	if len(positives) < minSamples {
		return flags
	}
	sort.Slice(positives, func(i, j int) bool { return positives[i] < positives[j] })
	n := len(positives)
	var median float64
	if n%2 == 1 {
		median = float64(positives[n/2])
	} else {
		median = float64(positives[n/2-1]+positives[n/2]) / 2
	}
	if median <= 0 {
		return flags
	}
	threshold := factor * median
	for i, v := range values {
		if v > 0 && float64(v) > threshold {
			flags[i] = true
		}
	}
	return flags
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
			sb.WriteRune('·')
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
