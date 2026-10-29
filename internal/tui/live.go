// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/loaders"
	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

// LivePollFunc is the data retrieval function invoked on every tick.
type LivePollFunc func() ([]thermal.ToolResult, []thermal.ProjectDay, error)

type liveTickMsg time.Time

type liveCollectedMsg struct {
	seq       uint64
	results   []thermal.ToolResult
	projects  []thermal.ProjectDay
	err       error
	timestamp time.Time
}

func tickLive(d time.Duration) tea.Cmd {
	if d <= 0 {
		d = time.Second
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return liveTickMsg(t)
	})
}

// LiveModel is the Bubble Tea v2 model for the real-time token burn monitor.
type LiveModel struct {
	pollFn       LivePollFunc
	tracker      *thermal.LiveTracker
	snapshot     thermal.LiveSnapshot
	lastResults  []thermal.ToolResult
	lastProjects []thermal.ProjectDay
	pollRate     time.Duration
	filterTool   string
	palette      Palette
	width        int
	height       int
	paused       bool
	quitting     bool
	showHelp     bool
	startedAt    time.Time
	lastPollAt   time.Time
	lastEventAt  time.Time
	pollErr      error
	sortMode     string
	inFlight     bool
	pollSeq      uint64
	appliedSeq   uint64
}

// NewLiveModel constructs an initialized LiveModel.
func NewLiveModel(pollFn LivePollFunc, pollRate time.Duration, filterTool string, colorful bool, flameCeiling ...float64) LiveModel {
	tracker := thermal.NewLiveTracker()
	if len(flameCeiling) > 0 {
		tracker.SetFlameCeiling(flameCeiling[0])
	}
	now := time.Now()

	var initialSnap thermal.LiveSnapshot
	var initialResults []thermal.ToolResult
	var initialProjects []thermal.ProjectDay
	lastPollAt := now
	var pollErr error
	if pollFn != nil {
		results, projects, err := pollFn()
		if err == nil {
			snap, _ := tracker.Poll(results, projects, now)
			initialSnap = snap
			initialResults = results
			initialProjects = projects
		} else {
			lastPollAt = time.Time{}
			pollErr = err
		}
	}

	return LiveModel{
		pollFn:       pollFn,
		tracker:      tracker,
		snapshot:     initialSnap,
		lastResults:  initialResults,
		lastProjects: initialProjects,
		pollRate:     pollRate,
		filterTool:   filterTool,
		palette:      newPalette(colorful),
		width:        80,
		height:       24,
		startedAt:    now,
		lastPollAt:   lastPollAt,
		pollErr:      pollErr,
		sortMode:     "time",
	}
}

// Init satisfies tea.Model, kicking off the live tick loop.
func (m LiveModel) Init() tea.Cmd {
	return tickLive(m.pollRate)
}

// Update handles tick events, window resize, and keyboard navigation.
func (m LiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case liveTickMsg:
		now := time.Time(msg)
		if m.tracker != nil {
			snap, _ := m.tracker.Poll(m.lastResults, m.lastProjects, now)
			m.snapshot = snap
		}

		tickCmd := tickLive(m.pollRate)

		// At most one collection in flight; do not run synchronously
		if !m.paused && !m.inFlight && m.pollFn != nil {
			m.inFlight = true
			m.pollSeq++
			seq := m.pollSeq
			fn := m.pollFn
			collectCmd := func() tea.Msg {
				res, prj, err := fn()
				return liveCollectedMsg{
					seq:       seq,
					results:   res,
					projects:  prj,
					err:       err,
					timestamp: time.Now(),
				}
			}
			return m, tea.Batch(tickCmd, collectCmd)
		}
		return m, tickCmd

	case liveCollectedMsg:
		m.inFlight = false
		if m.quitting || m.paused || msg.seq < m.appliedSeq {
			return m, nil
		}
		m.appliedSeq = msg.seq
		if msg.err != nil {
			m.pollErr = msg.err
			return m, nil
		}
		m.pollErr = nil
		m.lastPollAt = msg.timestamp
		if m.tracker != nil {
			m.lastResults = msg.results
			m.lastProjects = msg.projects
			snap, newEvents := m.tracker.Poll(msg.results, msg.projects, msg.timestamp)
			m.snapshot = snap
			if len(newEvents) > 0 {
				m.lastEventAt = msg.timestamp
			}
		}
		return m, nil
	}

	pressed, ok := isPress(msg)
	if !ok {
		return m, nil
	}

	if m.showHelp {
		switch pressed {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "?", "esc", "q", "enter", " ", "space":
			m.showHelp = false
			return m, nil
		default:
			return m, nil
		}
	}

	switch pressed {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		m.quitting = true
		return m, tea.Quit

	case " ", "space":
		m.paused = !m.paused
		return m, nil

	case "c", "C":
		if m.tracker != nil {
			m.tracker.ResetSession()
			m.snapshot.SessionTokens = 0
			m.snapshot.SessionCost = 0
			m.snapshot.SessionTurns = 0
		}
		return m, nil

	case "s", "S":
		if m.sortMode == "cost" {
			m.sortMode = "time"
		} else {
			m.sortMode = "cost"
		}
		return m, nil

	case "?":
		m.showHelp = true
		return m, nil
	}

	return m, nil
}

// renderStackedFrame is the 60-79 column layout: one KPI per line, no
// sparkline or flame bar, and a controls line that always keeps [q] Quit.
func (m LiveModel) renderStackedFrame(width int) string {
	cardWidth := width - 2
	if cardWidth < 40 {
		cardWidth = 40
	}
	inner := cardWidth - 4
	colors := m.palette.Colorful
	border := func(s string) string { return theme.Border.Sprint(colors, s) }

	status := "● LIVE"
	if m.paused {
		status = "⏸ PAUSED"
	}
	title := "🔥 THERMAL LIVE"
	rule := inner - 4 - ansi.StringWidth(title) - ansi.StringWidth(status)
	if rule < 1 {
		rule = 1
	}

	var b strings.Builder
	b.WriteString(border(" ╭─ ") + theme.Primary.SprintBold(colors, title) + " " +
		border(strings.Repeat("─", rule)) + " " + border(status) + border(" ─╮") + "\n")
	line := func(s string) {
		b.WriteString(" " + border("│ ") + padLine(s, inner) + border(" │") + "\n")
	}

	rate := fmt.Sprintf("%s tok/m", thermal.CompactNumber(int64(m.snapshot.BurnTokensPerMin)))
	if m.snapshot.BurnTokensPerMin <= 0 && m.snapshot.BurnTurnsPerMin > 0 {
		rate = fmt.Sprintf("%s step/m", thermal.CompactNumber(int64(m.snapshot.BurnTurnsPerMin)))
	}
	today := fmt.Sprintf("%s tok", thermal.CompactNumber(m.snapshot.TodayTokens))
	if m.snapshot.TodayTokens == 0 && m.snapshot.TodayTurns > 0 {
		today = fmt.Sprintf("%d turns", m.snapshot.TodayTurns)
	}
	session := fmt.Sprintf("%s tok", thermal.CompactNumber(m.snapshot.SessionTokens))
	if m.snapshot.SessionTokens == 0 && m.snapshot.SessionTurns > 0 {
		session = fmt.Sprintf("%d turns", m.snapshot.SessionTurns)
	}

	line(fmt.Sprintf("Rate     %s", rate))
	line(fmt.Sprintf("Today    %s  %s", today, formatCost(m.snapshot.TodayCost, colors)))
	line(fmt.Sprintf("Session  %s  %s", session, formatCost(m.snapshot.SessionCost, colors)))
	line(fmt.Sprintf("Cache    %.1f%% hit", m.snapshot.TodayCacheHit))
	line("")
	line("[q] Quit  [p] Pause  [s] Sort  [?] Help")
	b.WriteString(" " + border("╰"+strings.Repeat("─", inner+2)+"╯") + "\n")
	return b.String()
}

// View satisfies tea.Model, rendering the live monitor dashboard frame.
func (m LiveModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.showHelp {
		v := tea.NewView("\n" + m.renderHelp() + "\n")
		v.AltScreen = true
		return v
	}

	w := m.width
	if w <= 0 {
		w = 80
	}
	h := m.height
	if h <= 0 {
		h = 24
	}

	frame := m.renderLiveFrame(w, h)
	var content string
	if h >= 18 {
		content = "\n" + frame + "\n"
	} else {
		content = frame
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderLiveFrame assembles the themed, bounded card for the live monitor.
func (m LiveModel) renderLiveFrame(width, height int) string {
	if width < 80 {
		return m.renderStackedFrame(width)
	}
	colors := m.palette.Colorful

	// Responsive outer card width: adapts up to 200 columns for ultra-wide displays
	cardWidth := width - 4
	if cardWidth > 200 {
		cardWidth = 200
	}
	if cardWidth < 74 {
		cardWidth = 74
	}
	innerWidth := cardWidth - 4

	// 1. Header Line
	headerTitle := "🔥 THERMAL LIVE"
	if m.filterTool != "" {
		headerTitle += fmt.Sprintf(" [%s]", m.filterTool)
	}

	now := time.Now()
	currentTime := now.Format("15:04:05")
	rateSec := fmt.Sprintf("%.1fs", m.pollRate.Seconds())
	var statusIndicator string
	switch m.feedState(now) {
	case "paused":
		if colors {
			statusIndicator = theme.Warning.SprintBold(true, "⏸ PAUSED") + " · " + theme.TextMuted.Sprint(true, currentTime)
		} else {
			statusIndicator = "⏸ PAUSED · " + currentTime
		}
	case "offline":
		if colors {
			statusIndicator = theme.Error.SprintBold(true, "● OFFLINE") + " · " + theme.TextMuted.Sprint(true, currentTime)
		} else {
			statusIndicator = "● OFFLINE · " + currentTime
		}
	case "stale":
		if colors {
			statusIndicator = theme.Warning.SprintBold(true, "● STALE") + " · " + theme.TextMuted.Sprint(true, currentTime)
		} else {
			statusIndicator = "● STALE · " + currentTime
		}
	default:
		if colors {
			statusIndicator = theme.Success.SprintBold(true, "● LIVE") + " " + theme.TextMuted.Sprint(true, "("+rateSec+") · "+currentTime)
		} else {
			statusIndicator = "● LIVE (" + rateSec + ") · " + currentTime
		}
	}

	titleStyled := theme.Primary.SprintBold(colors, headerTitle)
	border := func(s string) string {
		return theme.Border.Sprint(colors, s)
	}

	// Exact rule length formula ensures top border matches cardWidth + 1 exactly:
	// " ╭─ " (4) + title + " " (1) + rule + " " (1) + status + " ─╮" (4) = innerWidth + 5 + 1
	ruleLen := innerWidth - 4 - ansi.StringWidth(headerTitle) - ansi.StringWidth(statusIndicator)
	if ruleLen < 1 {
		ruleLen = 1
	}
	topLine := border(" ╭─ ") + titleStyled + " " + border(strings.Repeat("─", ruleLen)) + " " + statusIndicator + border(" ─╮")

	// 2. Top 3-Box Metric KPI Grid
	colW := (innerWidth - 4) / 3
	if colW < 22 {
		colW = 22
	}

	isActivity := (m.snapshot.TodayTokens == 0 && m.snapshot.TodayTurns > 0) || (m.snapshot.SessionTokens == 0 && m.snapshot.SessionTurns > 0)
	hasTokenEvents := false
	hasActivityEvents := false
	for _, ev := range m.snapshot.RecentEvents {
		if ev.Tokens > 0 {
			hasTokenEvents = true
		}
		if ev.Tokens == 0 && ev.Turns > 0 {
			hasActivityEvents = true
		}
	}
	isHybrid := hasTokenEvents && hasActivityEvents

	burnRate := m.snapshot.BurnTokensPerMin
	burnSpeed := m.snapshot.BurnTokensPerSec
	intensity := m.snapshot.FlameIntensity
	rateIsActivity := isActivity

	if !isActivity && m.snapshot.BurnTokensPerMin <= 0 && m.snapshot.BurnTurnsPerMin > 0 {
		burnRate = m.snapshot.BurnTurnsPerMin
		burnSpeed = m.snapshot.BurnTurnsPerSec
		rateIsActivity = true
	} else if isActivity {
		burnRate = m.snapshot.BurnTurnsPerMin
		burnSpeed = m.snapshot.BurnTurnsPerSec
		rateIsActivity = true
	}

	// Column 1: Burn Velocity
	c1Title := theme.Primary.SprintBold(colors, "BURN VELOCITY")
	c1L1 := fmt.Sprintf("Rate:     %s", formatRate(burnRate, m.snapshot.BurnTurnsPerMin, rateIsActivity, colW, colors))
	c1L2 := fmt.Sprintf("Speed:    %s", formatSpeed(burnSpeed, m.snapshot.BurnTurnsPerSec, rateIsActivity, colW, colors))
	c1L3 := fmt.Sprintf("Burn/hr now: %s", formatCostHr(m.snapshot.BurnCostPerHr, colors))

	// Column 2: Today's Usage
	c2Title := theme.Secondary.SprintBold(colors, "TODAY'S USAGE")
	c2L1 := fmt.Sprintf("Volume:   %s", formatVolume(m.snapshot.TodayTokens, m.snapshot.TodayTurns, colW, colors))
	c2L2 := fmt.Sprintf("Cost:     %s", formatCost(m.snapshot.TodayCost, colors))
	c2L3 := formatCacheHit(m.snapshot.TodayCacheHit, colW, colors)

	// Column 3: Live Session
	c3Title := theme.Accent.SprintBold(colors, "LIVE SESSION")
	sessionEmpty := m.snapshot.SessionTokens == 0 && m.snapshot.SessionTurns == 0 && m.snapshot.SessionCost == 0
	sameAsToday := m.snapshot.TodayTokens == m.snapshot.SessionTokens &&
		m.snapshot.TodayTurns == m.snapshot.SessionTurns &&
		m.snapshot.TodayCost == m.snapshot.SessionCost &&
		(m.snapshot.TodayTokens > 0 || m.snapshot.TodayTurns > 0 || m.snapshot.TodayCost > 0)
	c3L1 := fmt.Sprintf("Burned:   %s", formatDelta(m.snapshot.SessionTokens, m.snapshot.SessionTurns, colW, colors))
	c3L2 := fmt.Sprintf("Spend:    %s", formatCost(m.snapshot.SessionCost, colors))
	c3L3 := fmt.Sprintf("State:    %s", formatIntensityBadge(intensity, colors))
	if sessionEmpty {
		c3L1 = "Since launch: 0 tok"
		c3L2 = theme.TextMuted.Sprint(colors, "Spend:    "+thermal.FormatCost(0, false))
		c3L3 = "State:    " + formatIntensityBadge(intensity, colors)
	}
	if sameAsToday {
		c3L1 = theme.TextMuted.Sprint(colors, "= Today's usage")
		c3L2 = ""
		c3L3 = ""
	}

	gridRows := []string{
		join3(c1Title, c2Title, c3Title, colW),
		join3(c1L1, c2L1, c3L1, colW),
		join3(c1L2, c2L2, c3L2, colW),
		join3(c1L3, c2L3, c3L3, colW),
	}
	carriedIn := formatCarriedIn(m.snapshot.CarriedInTokens, m.snapshot.CarriedInTurns, m.snapshot.CarriedInCost, colors)
	todayRollups := formatTodayRollups(m.snapshot.TodayProjects, m.snapshot.TodayModels)
	if height < 20 {
		todayRollups = nil
	}
	if !rateIsActivity && height >= 20 && (m.snapshot.BurnTurnsPerMin > 0 || m.snapshot.BurnTurnsPerSec > 0) {
		rateBoth := liveCellFits("Rate:     ", formatRateNumber(burnRate)+" tok/m · "+formatRateNumber(m.snapshot.BurnTurnsPerMin)+" step/m", colW)
		speedBoth := liveCellFits("Speed:    ", formatSpeedNumber(burnSpeed)+" tok/s · "+formatSpeedNumber(m.snapshot.BurnTurnsPerSec)+" step/s", colW)
		if !rateBoth || !speedBoth {
			todayRollups = append(todayRollups, fmt.Sprintf("  Steps: %s step/m · %s step/s",
				formatRateNumber(m.snapshot.BurnTurnsPerMin), formatSpeedNumber(m.snapshot.BurnTurnsPerSec)))
		}
	}

	// 3. Flame Intensity Gauge
	flameLine := formatFlameLine(intensity, m.snapshot.FlameMetric, m.snapshot.FlameCeiling, m.snapshot.ActiveModel, innerWidth, colors)

	// 4. Rolling 60s Sparkline
	hasTokenSpark := m.snapshot.PeakTokensPerSec > 0
	hasTurnSpark := m.snapshot.PeakStepsPerSec > 0

	var sparkBuckets [60]int64
	var peakVal int64
	isSparkActivity := isActivity
	if !isActivity && !hasTokenSpark && hasTurnSpark {
		isSparkActivity = true
	}

	if isSparkActivity {
		for i, v := range m.snapshot.RollingTurns {
			sparkBuckets[i] = int64(v)
		}
		peakVal = int64(m.snapshot.PeakStepsPerSec)
	} else {
		sparkBuckets = m.snapshot.RollingTokens
		peakVal = m.snapshot.PeakTokensPerSec
	}
	sparkLine := formatSparkLine(sparkBuckets, peakVal, innerWidth, isSparkActivity, colors)

	// 5. Divider helper - mathematically matches innerWidth + 5 exactly
	divider := func(label string) string {
		labStyled := theme.Border.Sprint(colors, label)
		dRule := innerWidth - 2 - ansi.StringWidth(label)
		if dRule < 1 {
			dRule = 1
		}
		return border(" ├─ ") + labStyled + " " + border(strings.Repeat("─", dRule)) + border("─┤")
	}

	// 6. Recent Token & Activity Bursts Ticker
	tickerLabel := "RECENT TOKEN BURSTS"
	if isActivity {
		tickerLabel = "RECENT ACTIVITY BURSTS"
	} else if isHybrid {
		tickerLabel = "RECENT TOKEN & ACTIVITY BURSTS"
	}
	tickerHeader := formatTickerHeader(innerWidth, colors)

	allEvents := make([]thermal.LiveEvent, 0, len(m.snapshot.RecentEvents))
	for _, ev := range m.snapshot.RecentEvents {
		if ev.IsBaseline {
			continue
		}
		allEvents = append(allEvents, ev)
	}
	groups := thermal.GroupLiveEvents(allEvents, 60*time.Second, now)
	sortLiveGroups(groups, m.sortMode)
	topSpendersLine := ""
	if m.sortMode == "cost" && len(groups) > 0 {
		topSpendersLine = formatTopSpenders(groups, colors)
	}

	// Available lines for events based on terminal height
	showSparkline := height >= 16
	fixedLines := 13
	if !showSparkline {
		fixedLines = 11
	}
	if carriedIn != "" {
		fixedLines++
	}
	fixedLines++
	fixedLines += len(todayRollups)
	if topSpendersLine != "" {
		fixedLines++
	}
	if height >= 18 {
		fixedLines += 2
	}
	availEvents := height - fixedLines
	if availEvents < 1 {
		availEvents = 1
	}
	if availEvents > 20 {
		availEvents = 20
	}
	if len(groups) > availEvents {
		groups = groups[:availEvents]
	}

	var eventLines []string
	if len(groups) == 0 {
		msg := "Idle: no new bursts in the last 60s"
		if m.lastEventAt.IsZero() {
			msg = "Idle: no events since launch"
		}
		if colors {
			msg = theme.TextMuted.Sprint(true, msg)
		}
		eventLines = append(eventLines, padLine("  "+msg, innerWidth))
	} else {
		tokValues := make([]int64, len(groups))
		stepValues := make([]int64, len(groups))
		for i, g := range groups {
			tokValues[i] = g.Tokens
			stepValues[i] = int64(g.Turns)
		}
		tokOutliers := thermal.MedianOutlierFlags(tokValues, 3.0, 3)
		stepOutliers := thermal.MedianOutlierFlags(stepValues, 3.0, 3)
		for i, g := range groups {
			g.AnyOutlier = tokOutliers[i] || stepOutliers[i]
			eventLines = append(eventLines, formatGroupRow(g, tokOutliers[i], stepOutliers[i], now, innerWidth, colors))
		}
	}

	// 7. Footer - exactly matches cardWidth
	bottomLine := border(" ╰" + strings.Repeat("─", cardWidth-2) + "╯")
	footerControls := theme.TextMuted.Sprint(colors, " • Controls: ") +
		theme.DarkStep12.Sprint(colors, "[Space]") + theme.TextMuted.Sprint(colors, " Pause  ·  ") +
		theme.DarkStep12.Sprint(colors, "[c]") + theme.TextMuted.Sprint(colors, " Reset  ·  ") +
		theme.DarkStep12.Sprint(colors, "[s]") + theme.TextMuted.Sprint(colors, " Sort: "+m.sortMode+"  ·  ") +
		theme.DarkStep12.Sprint(colors, "[?]") + theme.TextMuted.Sprint(colors, " Help  ·  ") +
		theme.DarkStep12.Sprint(colors, "[q]") + theme.TextMuted.Sprint(colors, " Quit")

	var sb strings.Builder
	sb.WriteString(topLine + "\n")
	for _, gr := range gridRows {
		sb.WriteString(border(" │ ") + padLine(gr, innerWidth) + border(" │\n"))
	}
	if carriedIn != "" {
		sb.WriteString(border(" │ ") + padLine(carriedIn, innerWidth) + border(" │\n"))
	}
	for _, line := range todayRollups {
		sb.WriteString(border(" │ ") + padLine(line, innerWidth) + border(" │\n"))
	}
	sb.WriteString(divider("FLAME INTENSITY") + "\n")
	sb.WriteString(border(" │ ") + padLine(flameLine, innerWidth) + border(" │\n"))
	if showSparkline {
		sb.WriteString(divider("ROLLING ACTIVITY · 60s / 1s buckets") + "\n")
		sb.WriteString(border(" │ ") + padLine(sparkLine, innerWidth) + border(" │\n"))
	}
	sb.WriteString(divider(tickerLabel) + "\n")
	sb.WriteString(border(" │ ") + padLine(tickerHeader, innerWidth) + border(" │\n"))
	sb.WriteString(border(" │ ") + padLine(m.formatLastEventLine(colors), innerWidth) + border(" │\n"))
	if topSpendersLine != "" {
		sb.WriteString(border(" │ ") + padLine(topSpendersLine, innerWidth) + border(" │\n"))
	}
	for _, el := range eventLines {
		sb.WriteString(border(" │ ") + padLine(el, innerWidth) + border(" │\n"))
	}
	sb.WriteString(bottomLine + "\n")
	sb.WriteString(" " + footerControls)

	return sb.String()
}

func (m LiveModel) renderHelp() string {
	colors := m.palette.Colorful
	w := m.width - 4
	if w > 86 {
		w = 86
	}
	if w < 72 {
		w = 72
	}
	inner := w - 4

	border := func(s string) string { return theme.Border.Sprint(colors, s) }
	title := theme.Primary.SprintBold(colors, "THERMAL LIVE · CONTROLS & MONITOR GUIDE")
	ruleLen := inner - 2 - ansi.StringWidth("THERMAL LIVE · CONTROLS & MONITOR GUIDE")
	if ruleLen < 1 {
		ruleLen = 1
	}

	top := border("╭─ ") + title + " " + border(strings.Repeat("─", ruleLen)) + border("─╮\n")
	bot := border("╰" + strings.Repeat("─", w-2) + "╯\n")

	lines := []string{
		"",
		theme.Secondary.SprintBold(colors, "  Keyboard Navigation & Actions:"),
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[Space]"), "Pause / resume live event polling"),
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[c]"), "Clear and reset live session counters"),
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[s]"), "Toggle time/cost sorting"),
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[?]"), "Toggle this help screen"),
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[q] / [Esc]"), "Quit live monitor"),
		"",
		theme.Secondary.SprintBold(colors, "  Real-Time Metrics & Indicators:"),
		"    Burn Velocity  Calculated over a rolling 60-second activity window",
		"    Flame Meter    Rate / ceiling; tokens default 9M/min, steps default 60/min",
		"    Rolling 60s    Second-by-second activity sparkline across the past minute",
		"    Event Ticker   Separate TOKENS/STEPS columns; TOOL×N groups N bursts",
		"    Ticker Marks   * = older than 60s; ! = over 3x the visible median",
		"    Costs          ~ = estimated; — = unpriced; ≥ = partial known; $0.00 = recorded zero",
		"",
		theme.Secondary.SprintBold(colors, "  Headless & Status Bar Automation:"),
		"    thermal live --fresh           Accepted for compatibility; sessions always start at launch",
		"    thermal live --json            Instant JSON snapshot of live metrics",
		"    thermal live --json --stream   Stream continuous NDJSON live events",
		"",
	}

	var sb strings.Builder
	sb.WriteString(" " + top)
	for _, l := range lines {
		pad := inner - ansi.StringWidth(l)
		if pad < 0 {
			pad = 0
		}
		sb.WriteString(" " + border("│ ") + l + strings.Repeat(" ", pad) + border(" │\n"))
	}
	sb.WriteString(" " + bot)
	sb.WriteString("  " + theme.TextMuted.Sprint(colors, "Press [?] or [Esc] to return to live monitor"))
	return sb.String()
}

// Helpers for formatted lines

// padLine ensures that s never exceeds width cells (truncating if needed),
// and pads with trailing spaces so the visual display width is EXACTLY width.
func padLine(s string, width int) string {
	sw := ansi.StringWidth(s)
	if sw > width {
		s = ansi.Truncate(s, width, "")
		sw = ansi.StringWidth(s)
	}
	if sw < width {
		s += strings.Repeat(" ", width-sw)
	}
	return s
}

// padCell pads or truncates text to an exact visual cell width.
func padCell(text string, width int, alignRight bool) string {
	sw := ansi.StringWidth(text)
	if sw > width {
		text = ansi.Truncate(text, width, "…")
		sw = ansi.StringWidth(text)
	}
	gap := width - sw
	if gap <= 0 {
		return text
	}
	if alignRight {
		return strings.Repeat(" ", gap) + text
	}
	return text + strings.Repeat(" ", gap)
}

func cleanModelName(m string) string {
	m = strings.TrimSpace(m)
	if dotIdx := strings.Index(m, ". "); dotIdx >= 0 {
		m = strings.TrimSpace(m[:dotIdx])
	}
	if semiIdx := strings.Index(m, ";"); semiIdx >= 0 {
		m = strings.TrimSpace(m[:semiIdx])
	}
	if nlIdx := strings.IndexAny(m, "\r\n\""); nlIdx >= 0 {
		m = strings.TrimSpace(m[:nlIdx])
	}
	m = strings.TrimSuffix(m, ".")
	return m
}

func join3(c1, c2, c3 string, colW int) string {
	return padCell(c1, colW, false) + "  " + padCell(c2, colW, false) + "  " + padCell(c3, colW, false)
}

func formatRateNumber(v float64) string {
	if v <= 0 {
		return "0"
	}
	if v < 10 {
		return fmt.Sprintf("%.1f", v)
	}
	return thermal.CompactNumber(int64(math.Round(v)))
}

func formatSpeedNumber(v float64) string {
	if v <= 0 {
		return "0"
	}
	if v < 1 {
		return fmt.Sprintf("%.3f", v)
	}
	if v < 10 {
		return fmt.Sprintf("%.2f", v)
	}
	return thermal.CompactNumber(int64(math.Round(v)))
}

func liveCellFits(label, value string, colW int) bool {
	return colW <= 0 || ansi.StringWidth(label+value) <= colW
}

func formatRate(tokRate, turnRate float64, isActivity bool, colW int, colors bool) string {
	if isActivity {
		if turnRate <= 0 {
			return theme.TextMuted.Sprint(colors, "0 step/m")
		}
		return theme.Primary.SprintBold(colors, formatRateNumber(turnRate)+" step/m")
	}
	if tokRate <= 0 && turnRate <= 0 {
		return theme.TextMuted.Sprint(colors, "0 tok/m")
	}
	if tokRate <= 0 {
		return theme.Primary.SprintBold(colors, formatRateNumber(turnRate)+" step/m")
	}
	token := formatRateNumber(tokRate) + " tok/m"
	if turnRate > 0 {
		both := token + " · " + formatRateNumber(turnRate) + " step/m"
		if liveCellFits("Rate:     ", both, colW) {
			return theme.Primary.SprintBold(colors, both)
		}
	}
	return theme.Primary.SprintBold(colors, token)
}

func formatSpeed(tokSpeed, turnSpeed float64, isActivity bool, colW int, colors bool) string {
	if isActivity {
		if turnSpeed <= 0 {
			return theme.TextMuted.Sprint(colors, "0 step/s")
		}
		return formatSpeedNumber(turnSpeed) + " step/s"
	}
	if tokSpeed <= 0 && turnSpeed <= 0 {
		return theme.TextMuted.Sprint(colors, "0 tok/s")
	}
	if tokSpeed > 0 {
		token := formatSpeedNumber(tokSpeed) + " tok/s"
		if turnSpeed > 0 {
			both := token + " · " + formatSpeedNumber(turnSpeed) + " step/s"
			if liveCellFits("Speed:    ", both, colW) {
				return both
			}
		}
		return token
	}
	return formatSpeedNumber(turnSpeed) + " step/s"
}

func formatCostHr(costHr float64, colors bool) string {
	if costHr <= 0 {
		return theme.TextMuted.Sprint(colors, "$0.00/hr")
	}
	if costHr < 0.01 {
		return "<$0.01/hr"
	}
	return thermal.FormatCost(costHr, true) + "/hr"
}

func formatVolume(tokens int64, turns int, colW int, colors bool) string {
	if tokens <= 0 && turns > 0 {
		str := fmt.Sprintf("%s step", thermal.CompactNumber(int64(turns)))
		return theme.Secondary.SprintBold(colors, str)
	}
	if tokens <= 0 {
		return theme.TextMuted.Sprint(colors, "0 tok")
	}
	if turns > 0 {
		full := fmt.Sprintf("%s tok · %s step", thermal.CompactNumber(tokens), thermal.CompactNumber(int64(turns)))
		if colW <= 0 || ansi.StringWidth("Volume:   "+full) <= colW {
			return theme.Secondary.SprintBold(colors, full)
		}
		compact := fmt.Sprintf("%s tok (%s step)", thermal.CompactNumber(tokens), thermal.CompactNumber(int64(turns)))
		if ansi.StringWidth("Volume:   "+compact) <= colW {
			return theme.Secondary.SprintBold(colors, compact)
		}
		short := fmt.Sprintf("%s tok %s step", thermal.CompactNumber(tokens), thermal.CompactNumber(int64(turns)))
		if ansi.StringWidth("Volume:   "+short) <= colW {
			return theme.Secondary.SprintBold(colors, short)
		}
	}
	str := fmt.Sprintf("%s tok", thermal.CompactNumber(tokens))
	return theme.Secondary.SprintBold(colors, str)
}

func formatDelta(tokens int64, turns int, colW int, colors bool) string {
	if tokens <= 0 && turns > 0 {
		str := fmt.Sprintf("+%s step", thermal.CompactNumber(int64(turns)))
		return theme.Accent.SprintBold(colors, str)
	}
	if tokens <= 0 {
		return theme.TextMuted.Sprint(colors, "+0 tok")
	}
	if turns > 0 {
		full := fmt.Sprintf("+%s tok · +%s step", thermal.CompactNumber(tokens), thermal.CompactNumber(int64(turns)))
		if colW <= 0 || ansi.StringWidth("Burned:   "+full) <= colW {
			return theme.Accent.SprintBold(colors, full)
		}
		compact := fmt.Sprintf("+%s tok (+%s step)", thermal.CompactNumber(tokens), thermal.CompactNumber(int64(turns)))
		if ansi.StringWidth("Burned:   "+compact) <= colW {
			return theme.Accent.SprintBold(colors, compact)
		}
		short := fmt.Sprintf("+%s tok +%s step", thermal.CompactNumber(tokens), thermal.CompactNumber(int64(turns)))
		if ansi.StringWidth("Burned:   "+short) <= colW {
			return theme.Accent.SprintBold(colors, short)
		}
	}
	str := fmt.Sprintf("+%s tok", thermal.CompactNumber(tokens))
	return theme.Accent.SprintBold(colors, str)
}

func (m LiveModel) feedState(now time.Time) string {
	if m.paused {
		return "paused"
	}
	if m.pollErr != nil || m.lastPollAt.IsZero() {
		return "offline"
	}
	threshold := 3 * m.pollRate
	if threshold < 5*time.Second {
		threshold = 5 * time.Second
	}
	if now.Sub(m.lastPollAt) > threshold {
		return "stale"
	}
	return "live"
}

func formatAge(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

func (m LiveModel) formatLastEventLine(colors bool) string {
	if m.lastEventAt.IsZero() {
		return theme.TextMuted.Sprint(colors, "  Last event: none · * >60s · ! >3x")
	}
	age := time.Since(m.lastEventAt)
	label := "  Last event: " + formatAge(age) + " ago · * >60s · ! >3x"
	if age > 60*time.Second {
		return theme.Warning.Sprint(colors, label)
	}
	return theme.TextMuted.Sprint(colors, label)
}

func formatRollupCost(r thermal.LiveRollup) string {
	return formatLiveCostPlain(thermal.LiveEvent{Cost: r.Cost, CostKind: r.CostKind})
}

func formatTodayRollups(projects, models []thermal.LiveRollup) []string {
	var lines []string
	if len(projects) > 0 {
		parts := make([]string, 0, 2)
		for i, p := range projects {
			if i >= 2 {
				break
			}
			parts = append(parts, p.Name+" "+formatRollupCost(p))
		}
		lines = append(lines, "  Today by project: "+strings.Join(parts, " · "))
	}
	if len(models) > 0 {
		top := models[0]
		label := "  Top model: "
		if top.CostKind == thermal.LiveCostEstimated {
			label = "  Top model (est. cost): "
		}
		lines = append(lines, label+top.Name+" "+formatRollupCost(top))
	}
	return lines
}

func formatCarriedIn(tokens int64, turns int, cost float64, colors bool) string {
	var parts []string
	if tokens > 0 {
		parts = append(parts, thermal.CompactNumber(tokens)+" tok")
	}
	if turns > 0 {
		parts = append(parts, thermal.CompactNumber(int64(turns))+" step")
	}
	if cost > 0 {
		parts = append(parts, formatCost(cost, colors))
	}
	if len(parts) == 0 {
		return ""
	}
	label := theme.TextMuted.Sprint(colors, "Carried in: ")
	suffix := theme.TextMuted.Sprint(colors, " (baseline)")
	return label + strings.Join(parts, " · ") + suffix
}

func formatCost(c float64, colors bool) string {
	if c <= 0 {
		return theme.TextMuted.Sprint(colors, thermal.FormatCost(0, false))
	}
	return thermal.FormatCost(c, true)
}

func formatCacheHit(hit float64, availWidth int, colors bool) string {
	if hit <= 0 {
		return theme.TextMuted.Sprint(colors, "Cache hit: —")
	}
	full := "Cache hit: " + thermal.FormatCacheHit(hit)
	if availWidth <= 0 || ansi.StringWidth(full) <= availWidth {
		return full
	}
	compact := thermal.FormatPercent(hit, 1) + " cache"
	if ansi.StringWidth(compact) <= availWidth {
		return compact
	}
	mini := fmt.Sprintf("CH %.0f%%", hit)
	if ansi.StringWidth(mini) <= availWidth {
		return mini
	}
	return fmt.Sprintf("%.0f%%", hit)
}

func gaugeWidth(innerWidth int) int {
	w := innerWidth / 4
	if w > 32 {
		w = 32
	}
	if w < 16 {
		w = 16
	}
	return w
}

func formatIntensityBadge(intensity float64, colors bool) string {
	switch {
	case intensity >= 0.6:
		return theme.Error.SprintBold(colors, "BURNING 🔥")
	case intensity >= 0.2:
		return theme.Primary.SprintBold(colors, "WARM ⚡")
	case intensity > 0.02:
		return theme.Warning.Sprint(colors, "COOLING 💨")
	default:
		return theme.TextMuted.Sprint(colors, "IDLE 💤")
	}
}

func formatFlameLine(intensity float64, metric string, ceiling float64, model string, innerWidth int, colors bool) string {
	barW := gaugeWidth(innerWidth)

	var barStr string
	var pct int
	if intensity <= 0 {
		barStr = strings.Repeat("·", barW)
	} else {
		barStr, pct = thermal.FlameBar(intensity, barW)
	}
	pctStr := fmt.Sprintf("%3d%%", pct)
	state := formatIntensityBadge(intensity, colors)

	var barStyled string
	if colors {
		filled := int(math.Round(intensity * float64(barW)))
		if filled > barW {
			filled = barW
		}
		var fillCol theme.Color
		switch {
		case intensity >= 0.6:
			fillCol = theme.Error
		case intensity >= 0.25:
			fillCol = theme.Primary
		default:
			fillCol = theme.Warning
		}
		runes := []rune(barStr)
		if filled > len(runes) {
			filled = len(runes)
		}
		partFill := fillCol.Sprint(true, string(runes[:filled]))
		partEmpty := theme.DarkStep6.Sprint(true, string(runes[filled:]))
		barStyled = "[" + partFill + partEmpty + "]"
	} else {
		barStyled = "[" + barStr + "]"
	}

	ceilingStr := thermal.CompactNumber(int64(math.Round(ceiling))) + " tok/m"
	if metric == "steps" {
		ceilingStr = fmt.Sprintf("%.0f step/m", ceiling)
	}
	prefix := fmt.Sprintf("  Flame: %s %s of %s  %s", barStyled, pctStr, ceilingStr, state)
	prefixWidth := ansi.StringWidth(prefix)

	modClean := loaders.CanonicalModelName(cleanModelName(model))
	if modClean != "" {
		avail := innerWidth - prefixWidth - 5
		if avail >= 8 {
			modTrunc := ansi.Truncate(modClean, avail, "…")
			modStyled := theme.Secondary.Sprint(colors, modTrunc)
			prefix += "  ·  " + modStyled
		}
	}

	return prefix
}

func formatSparkLine(buckets [60]int64, peak int64, innerWidth int, isActivity bool, colors bool) string {
	prefix := "  Spark: ["
	suffix := "]"

	unitSec := " tok/s"
	unitSum := " tok"
	label60 := "60s Burn: "
	if isActivity {
		unitSec = " step/s"
		unitSum = " step"
		label60 = "60s Steps: "
	}

	peakStr := ""
	if peak > 0 {
		formatted := thermal.CompactNumber(peak) + unitSec + " (1s bucket)"
		if colors {
			peakStr = "  Peak: " + theme.DarkStep12.Sprint(true, formatted)
		} else {
			peakStr = "  Peak: " + formatted
		}
	}

	sparkW := gaugeWidth(innerWidth)

	spark := thermal.Sparkline(buckets, 59, sparkW)
	allZero := true
	for _, b := range buckets {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		spark = strings.Repeat("·", sparkW)
		if peakStr == "" {
			peakStr = "  no activity (60s)"
		}
	}
	sparkStyled := spark
	if colors {
		sparkStyled = theme.Cyan.Sprint(true, spark)
	}

	line := fmt.Sprintf("%s%s%s%s", prefix, sparkStyled, suffix, peakStr)

	// If there's generous room (wide terminal), show 60s burn sum
	if innerWidth >= 110 {
		var sum60 int64
		for _, b := range buckets {
			sum60 += b
		}
		if sum60 > 0 {
			volStr := fmt.Sprintf("  ·  %s%s%s", label60, thermal.CompactNumber(sum60), unitSum)
			if ansi.StringWidth(line)+ansi.StringWidth(volStr) < innerWidth-2 {
				if colors {
					volStr = "  ·  " + theme.TextMuted.Sprint(true, label60) + theme.Primary.SprintBold(true, thermal.CompactNumber(sum60)+unitSum)
				}
				line += volStr
			}
		}
	}

	return line
}

type tickerLayout struct {
	timeW, ageW, toolW, tokW, stepW, modelW, projectW, costW int
	showModel, showProject                                   bool
	compact                                                  bool
}

func calcColWidths(innerWidth int) tickerLayout {
	l := tickerLayout{
		timeW: 8,
		ageW:  4,
		toolW: 10,
		tokW:  10,
		stepW: 8,
		costW: 10,
	}
	// Fixed columns: indent 2 + time 8 + age 4 + tool 10 + tokens 10 +
	// steps 8 + cost 10 + five 2-cell gaps = 62 cells.
	remaining := innerWidth - 62
	if remaining >= 24 {
		l.showModel = true
		l.showProject = true
		available := remaining - 4 // two optional-column gaps
		l.modelW = available * 55 / 100
		if l.modelW < 12 {
			l.modelW = 12
		}
		l.projectW = available - l.modelW
		if l.projectW < 8 {
			l.projectW = 8
		}
	} else if remaining >= 14 {
		l.showModel = true
		l.modelW = remaining - 2 // optional-column gap
		if l.modelW < 12 {
			l.modelW = 12
		}
	}
	l.compact = innerWidth < 90
	return l
}

func formatTickerHeader(innerWidth int, colors bool) string {
	l := calcColWidths(innerWidth)
	tokLabel, stepLabel := "TOKENS", "STEPS"
	if l.compact {
		tokLabel, stepLabel = "TOK", "STEP"
	}

	parts := []string{
		padCell("TIME", l.timeW, false),
		padCell("AGE", l.ageW, false),
		padCell("TOOL", l.toolW, false),
		padCell(tokLabel, l.tokW, true),
		padCell(stepLabel, l.stepW, true),
	}
	if l.showModel {
		parts = append(parts, padCell("MODEL", l.modelW, false))
	}
	if l.showProject {
		parts = append(parts, padCell("PROJECT", l.projectW, false))
	}
	parts = append(parts, padCell("COST", l.costW, true))

	hdr := "  " + strings.Join(parts, "  ")
	if colors {
		return theme.TextMuted.Sprint(true, hdr)
	}
	return hdr
}

func formatLiveCostPlain(ev thermal.LiveEvent) string {
	if ev.CostKind == thermal.LiveCostUnpriced {
		return "—"
	}
	if ev.Cost <= 0 {
		if ev.CostKind == thermal.LiveCostRecorded || (ev.CostKind == "" && !ev.IsEstimated) {
			return "$0.00"
		}
		return "—"
	}
	prefix := ""
	if ev.CostKind == thermal.LiveCostEstimated || (ev.CostKind == "" && ev.IsEstimated) {
		prefix = "~"
	}
	if ev.Cost < 0.00005 {
		return prefix + "<$0.0001"
	}
	if ev.Cost < 0.01 {
		return fmt.Sprintf("%s$%.4f", prefix, ev.Cost)
	}
	if ev.Cost >= 1000 {
		return fmt.Sprintf("%s$%.0f", prefix, ev.Cost)
	}
	return fmt.Sprintf("%s$%.2f", prefix, ev.Cost)
}

func sortLiveGroups(groups []thermal.LiveEventGroup, mode string) {
	sort.SliceStable(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		if mode == "cost" {
			aPriced, bPriced := a.Cost > 0, b.Cost > 0
			if aPriced != bPriced {
				return aPriced
			}
			if a.Cost != b.Cost {
				return a.Cost > b.Cost
			}
			if a.Tokens != b.Tokens {
				return a.Tokens > b.Tokens
			}
			return a.Latest.After(b.Latest)
		}
		if !a.Latest.Equal(b.Latest) {
			return a.Latest.After(b.Latest)
		}
		if a.Tokens != b.Tokens {
			return a.Tokens > b.Tokens
		}
		return a.Tool+a.Model+a.Project < b.Tool+b.Model+b.Project
	})
}

func formatTopSpenders(groups []thermal.LiveEventGroup, colors bool) string {
	type spend struct {
		name     string
		cost     float64
		unpriced bool
		kind     thermal.LiveCostKind
	}
	byProject := make(map[string]*spend)
	for _, g := range groups {
		name := g.Project
		if name == "" {
			name = "unattributed"
		}
		s := byProject[name]
		if s == nil {
			s = &spend{name: name}
			byProject[name] = s
		}
		s.cost += g.Cost
		s.unpriced = s.unpriced || g.AnyUnpriced
		if g.CostKind == thermal.LiveCostEstimated {
			s.kind = thermal.LiveCostEstimated
		} else if g.CostKind == thermal.LiveCostRecorded && s.kind != thermal.LiveCostEstimated {
			s.kind = thermal.LiveCostRecorded
		}
	}
	list := make([]*spend, 0, len(byProject))
	for _, s := range byProject {
		list = append(list, s)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].cost != list[j].cost {
			return list[i].cost > list[j].cost
		}
		return list[i].name < list[j].name
	})
	parts := make([]string, 0, 3)
	for _, s := range list {
		if s.cost <= 0 {
			continue
		}
		costStr := formatRollupCost(thermal.LiveRollup{Cost: s.cost, CostKind: s.kind})
		if s.unpriced {
			costStr = "≥" + strings.TrimPrefix(costStr, "~")
		}
		parts = append(parts, s.name+" "+costStr)
		if len(parts) >= 3 {
			break
		}
	}
	if len(parts) == 0 {
		return "  Top spenders (60s): —"
	}
	return "  Top spenders (60s): " + strings.Join(parts, " · ")
}

func formatGroupRow(g thermal.LiveEventGroup, tokOutlier, stepOutlier bool, now time.Time, innerWidth int, colors bool) string {
	return formatTickerRow(thermal.LiveEvent{
		Timestamp: g.Latest,
		Tool:      g.Tool,
		Tokens:    g.Tokens,
		Turns:     g.Turns,
		Model:     g.Model,
		Project:   g.Project,
		Cost:      g.Cost,
		CostKind:  g.CostKind,
	}, g.Count, g.AnyUnpriced, tokOutlier, stepOutlier, now, innerWidth, colors)
}

func formatEventRow(ev thermal.LiveEvent, tokOutlier, stepOutlier bool, now time.Time, innerWidth int, colors bool) string {
	return formatTickerRow(ev, 1, false, tokOutlier, stepOutlier, now, innerWidth, colors)
}

func formatTickerRow(ev thermal.LiveEvent, count int, anyUnpriced bool, tokOutlier, stepOutlier bool, now time.Time, innerWidth int, colors bool) string {
	l := calcColWidths(innerWidth)

	timeStr := ev.Timestamp.Format("15:04:05")
	age := now.Sub(ev.Timestamp)
	if age < 0 {
		age = 0
	}
	stale := age > 60*time.Second
	ageStr := formatAge(age)
	if stale {
		ageStr += "*"
	}

	toolStr := ev.Tool
	if toolStr == "" {
		toolStr = "—"
	} else if count > 1 {
		toolStr = fmt.Sprintf("%s×%d", toolStr, count)
	}
	tokStr := "—"
	if ev.Tokens > 0 {
		tokStr = "+" + thermal.CompactNumber(ev.Tokens)
		if tokOutlier {
			tokStr += "!"
		}
	}
	stepStr := "—"
	if ev.Turns > 0 {
		stepStr = "+" + thermal.CompactNumber(int64(ev.Turns))
		if stepOutlier {
			stepStr += "!"
		}
	}
	modStr := loaders.CanonicalModelName(cleanModelName(ev.Model))
	if modStr == "" {
		modStr = "—"
	}
	projStr := ev.Project
	if projStr == "" {
		projStr = "—"
	}
	costStr := formatLiveCostPlain(ev)
	if anyUnpriced {
		if ev.Cost <= 0 {
			costStr = "—"
		} else {
			costStr = "≥" + strings.TrimPrefix(costStr, "~")
		}
	}

	timeCell := padCell(timeStr, l.timeW, false)
	ageCell := padCell(ageStr, l.ageW, false)
	toolCell := padCell(toolStr, l.toolW, false)
	tokCell := padCell(tokStr, l.tokW, true)
	stepCell := padCell(stepStr, l.stepW, true)
	modCell := padCell(modStr, l.modelW, false)
	projCell := padCell(projStr, l.projectW, false)
	costCell := padCell(costStr, l.costW, true)

	if colors {
		if stale {
			timeCell = theme.TextMuted.Sprint(true, timeCell)
			ageCell = theme.TextMuted.Sprint(true, ageCell)
			toolCell = theme.TextMuted.Sprint(true, toolCell)
			tokCell = theme.TextMuted.Sprint(true, tokCell)
			stepCell = theme.TextMuted.Sprint(true, stepCell)
			modCell = theme.TextMuted.Sprint(true, modCell)
			projCell = theme.TextMuted.Sprint(true, projCell)
			costCell = theme.TextMuted.Sprint(true, costCell)
		} else {
			timeCell = theme.TextMuted.Sprint(true, timeCell)
			ageCell = theme.TextMuted.Sprint(true, ageCell)
			toolCell = theme.Secondary.Sprint(true, toolCell)
			if tokOutlier {
				tokCell = theme.Warning.SprintBold(true, tokCell)
			} else {
				tokCell = theme.Primary.SprintBold(true, tokCell)
			}
			if stepOutlier {
				stepCell = theme.Warning.SprintBold(true, stepCell)
			} else {
				stepCell = theme.Accent.Sprint(true, stepCell)
			}
			modCell = theme.DarkStep12.Sprint(true, modCell)
			projCell = theme.Accent.Sprint(true, projCell)
			costCell = theme.TextMuted.Sprint(true, costCell)
		}
	}

	parts := []string{timeCell, ageCell, toolCell, tokCell, stepCell}
	if l.showModel {
		parts = append(parts, modCell)
	}
	if l.showProject {
		parts = append(parts, projCell)
	}
	parts = append(parts, costCell)
	return "  " + strings.Join(parts, "  ")
}
