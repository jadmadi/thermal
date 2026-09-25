// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"math"
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
	inFlight     bool
	pollSeq      uint64
	appliedSeq   uint64
}

// NewLiveModel constructs an initialized LiveModel.
func NewLiveModel(pollFn LivePollFunc, pollRate time.Duration, filterTool string, colorful bool, seedToday ...bool) LiveModel {
	tracker := thermal.NewLiveTracker()
	seed := true
	if len(seedToday) > 0 {
		seed = seedToday[0]
	}
	tracker.SetSeedToday(seed)
	now := time.Now()

	var initialSnap thermal.LiveSnapshot
	var initialResults []thermal.ToolResult
	var initialProjects []thermal.ProjectDay
	if pollFn != nil {
		results, projects, err := pollFn()
		if err == nil {
			snap, _ := tracker.Poll(results, projects, now)
			initialSnap = snap
			initialResults = results
			initialProjects = projects
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
		if msg.err == nil && m.tracker != nil {
			m.lastResults = msg.results
			m.lastProjects = msg.projects
			snap, _ := m.tracker.Poll(msg.results, msg.projects, msg.timestamp)
			m.snapshot = snap
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

	case "?":
		m.showHelp = true
		return m, nil
	}

	return m, nil
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

	content := "\n" + m.renderLiveFrame(w, h) + "\n"
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderLiveFrame assembles the themed, bounded card for the live monitor.
func (m LiveModel) renderLiveFrame(width, height int) string {
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

	currentTime := time.Now().Format("15:04:05")
	var statusIndicator string
	if m.paused {
		if colors {
			statusIndicator = theme.Warning.SprintBold(true, "⏸ PAUSED") + " · " + theme.TextMuted.Sprint(true, currentTime)
		} else {
			statusIndicator = "⏸ PAUSED · " + currentTime
		}
	} else {
		rateSec := fmt.Sprintf("%.1fs", m.pollRate.Seconds())
		if colors {
			statusIndicator = theme.Success.SprintBold(true, "● STREAMING") + " " + theme.TextMuted.Sprint(true, "("+rateSec+") · "+currentTime)
		} else {
			statusIndicator = "● STREAMING (" + rateSec + ") · " + currentTime
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

	burnRate := m.snapshot.BurnTokensPerMin
	burnSpeed := m.snapshot.BurnTokensPerSec
	intensity := m.snapshot.FlameIntensity
	if isActivity {
		burnRate = m.snapshot.BurnTurnsPerMin
		burnSpeed = m.snapshot.BurnTurnsPerSec
		intensity = m.snapshot.ActivityIntensity
	}

	// Column 1: Burn Velocity
	c1Title := theme.Primary.SprintBold(colors, "BURN VELOCITY")
	c1L1 := fmt.Sprintf("Rate:     %s", formatRate(burnRate, isActivity, colors))
	c1L2 := fmt.Sprintf("Speed:    %s", formatSpeed(burnSpeed, isActivity, colors))
	c1L3 := fmt.Sprintf("Burn/Hr:  %s", formatCostHr(m.snapshot.BurnCostPerHr, colors))

	// Column 2: Today's Usage
	c2Title := theme.Secondary.SprintBold(colors, "TODAY'S USAGE")
	c2L1 := fmt.Sprintf("Volume:   %s", formatVolume(m.snapshot.TodayTokens, m.snapshot.TodayTurns, colors))
	c2L2 := fmt.Sprintf("Cost:     %s", formatCost(m.snapshot.TodayCost, colors))
	c2L3 := fmt.Sprintf("Turns:    %s", formatTurnsAndHit(m.snapshot.TodayTurns, m.snapshot.TodayCacheHit, colors))

	// Column 3: Live Session
	c3Title := theme.Accent.SprintBold(colors, "LIVE SESSION")
	c3L1 := fmt.Sprintf("Burned:   %s", formatDelta(m.snapshot.SessionTokens, m.snapshot.SessionTurns, colors))
	c3L2 := fmt.Sprintf("Spend:    %s", formatCost(m.snapshot.SessionCost, colors))
	c3L3 := fmt.Sprintf("State:    %s", formatIntensityBadge(intensity, colors))

	gridRows := []string{
		join3(c1Title, c2Title, c3Title, colW),
		join3(c1L1, c2L1, c3L1, colW),
		join3(c1L2, c2L2, c3L2, colW),
		join3(c1L3, c2L3, c3L3, colW),
	}

	// 3. Flame Intensity Gauge
	flameLine := formatFlameLine(intensity, m.snapshot.ActiveModel, innerWidth, colors)

	// 4. Rolling 60s Sparkline
	var sparkBuckets [60]int64
	var peakVal int64
	if isActivity {
		for i, v := range m.snapshot.RollingTurns {
			sparkBuckets[i] = int64(v)
		}
		peakVal = int64(m.snapshot.PeakRollingTurn)
	} else {
		sparkBuckets = m.snapshot.RollingTokens
		peakVal = m.snapshot.PeakRollingTok
	}
	sparkLine := formatSparkLine(sparkBuckets, peakVal, innerWidth, isActivity, colors)

	// 5. Divider helper - mathematically matches innerWidth + 5 exactly
	divider := func(label string) string {
		labStyled := theme.Border.Sprint(colors, label)
		dRule := innerWidth - 2 - ansi.StringWidth(label)
		if dRule < 1 {
			dRule = 1
		}
		return border(" ├─ ") + labStyled + " " + border(strings.Repeat("─", dRule)) + border("─┤")
	}

	// 6. Recent Token Bursts Ticker
	tickerHeader := formatTickerHeader(innerWidth, colors)

	// Available lines for events based on terminal height
	fixedLines := 15
	availEvents := height - fixedLines
	if availEvents < 3 {
		availEvents = 3
	}
	if availEvents > 20 {
		availEvents = 20
	}

	var eventLines []string
	if len(m.snapshot.RecentEvents) == 0 {
		msg := "No new token bursts detected yet. Watching SQLite DBs & transcript logs..."
		if colors {
			msg = theme.TextMuted.Sprint(true, msg)
		}
		eventLines = append(eventLines, padLine("  "+msg, innerWidth))
	} else {
		count := len(m.snapshot.RecentEvents)
		if count > availEvents {
			count = availEvents
		}
		for i := 0; i < count; i++ {
			ev := m.snapshot.RecentEvents[i]
			eventLines = append(eventLines, formatEventRow(ev, innerWidth, colors))
		}
	}

	// 7. Footer - exactly matches cardWidth
	bottomLine := border(" ╰" + strings.Repeat("─", cardWidth-2) + "╯")
	footerControls := theme.TextMuted.Sprint(colors, " • Controls: ") +
		theme.DarkStep12.Sprint(colors, "[Space]") + theme.TextMuted.Sprint(colors, " Pause/Resume  ·  ") +
		theme.DarkStep12.Sprint(colors, "[c]") + theme.TextMuted.Sprint(colors, " Reset Session  ·  ") +
		theme.DarkStep12.Sprint(colors, "[?]") + theme.TextMuted.Sprint(colors, " Help  ·  ") +
		theme.DarkStep12.Sprint(colors, "[q]") + theme.TextMuted.Sprint(colors, " Quit")

	var sb strings.Builder
	sb.WriteString(topLine + "\n")
	for _, gr := range gridRows {
		sb.WriteString(border(" │ ") + padLine(gr, innerWidth) + border(" │\n"))
	}
	sb.WriteString(divider("FLAME INTENSITY") + "\n")
	sb.WriteString(border(" │ ") + padLine(flameLine, innerWidth) + border(" │\n"))
	sb.WriteString(divider("ROLLING ACTIVITY (60s)") + "\n")
	sb.WriteString(border(" │ ") + padLine(sparkLine, innerWidth) + border(" │\n"))
	sb.WriteString(divider("RECENT TOKEN BURSTS") + "\n")
	sb.WriteString(border(" │ ") + padLine(tickerHeader, innerWidth) + border(" │\n"))
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
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[?]"), "Toggle this help screen"),
		fmt.Sprintf("    %-14s %s", theme.DarkStep12.Sprint(colors, "[q] / [Esc]"), "Quit live monitor"),
		"",
		theme.Secondary.SprintBold(colors, "  Real-Time Metrics & Indicators:"),
		"    Burn Velocity  Calculated over a rolling 60-second activity window",
		"    Flame Meter    Visual burst intensity; decays naturally when idle",
		"    Rolling 60s    Second-by-second activity sparkline across the past minute",
		"    Event Ticker   Completed turn deltas with model, project, and cost",
		"",
		theme.Secondary.SprintBold(colors, "  Headless & Status Bar Automation:"),
		"    thermal live --fresh           Start live session counters from 0 instead of today",
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

func formatRate(rate float64, isActivity bool, colors bool) string {
	if isActivity {
		if rate <= 0 {
			return theme.TextMuted.Sprint(colors, "0 step/min")
		}
		str := fmt.Sprintf("%s step/min", thermal.CompactNumber(int64(rate)))
		return theme.Primary.SprintBold(colors, str)
	}
	if rate <= 0 {
		return theme.TextMuted.Sprint(colors, "0 tok/min")
	}
	str := fmt.Sprintf("%s/min", thermal.CompactNumber(int64(rate)))
	return theme.Primary.SprintBold(colors, str)
}

func formatSpeed(speed float64, isActivity bool, colors bool) string {
	if isActivity {
		if speed <= 0 {
			return theme.TextMuted.Sprint(colors, "0 step/s")
		}
		return fmt.Sprintf("%s step/s", thermal.CompactNumber(int64(speed)))
	}
	if speed <= 0 {
		return theme.TextMuted.Sprint(colors, "0 tok/s")
	}
	return fmt.Sprintf("%s tok/s", thermal.CompactNumber(int64(speed)))
}

func formatCostHr(costHr float64, colors bool) string {
	if costHr <= 0 {
		return theme.TextMuted.Sprint(colors, "$0.00/hr")
	}
	if costHr < 0.01 {
		return "<$0.01/hr"
	}
	return fmt.Sprintf("~$%.2f/hr", costHr)
}

func formatVolume(tokens int64, turns int, colors bool) string {
	if tokens <= 0 && turns > 0 {
		str := fmt.Sprintf("%s step", thermal.CompactNumber(int64(turns)))
		return theme.Secondary.SprintBold(colors, str)
	}
	if tokens <= 0 {
		return theme.TextMuted.Sprint(colors, "0 tok")
	}
	str := fmt.Sprintf("%s tok", thermal.CompactNumber(tokens))
	return theme.Secondary.SprintBold(colors, str)
}

func formatDelta(tokens int64, turns int, colors bool) string {
	if tokens <= 0 && turns > 0 {
		str := fmt.Sprintf("+%s step", thermal.CompactNumber(int64(turns)))
		return theme.Accent.SprintBold(colors, str)
	}
	if tokens <= 0 {
		return theme.TextMuted.Sprint(colors, "+0 tok")
	}
	str := fmt.Sprintf("+%s tok", thermal.CompactNumber(tokens))
	return theme.Accent.SprintBold(colors, str)
}

func formatCost(c float64, colors bool) string {
	if c <= 0 {
		return theme.TextMuted.Sprint(colors, "$0.00")
	}
	if c < 0.01 {
		return "<$0.01"
	}
	if c >= 1000 {
		return fmt.Sprintf("~$%.0f", c)
	}
	return fmt.Sprintf("~$%.2f", c)
}

func formatTurnsAndHit(turns int, hit float64, colors bool) string {
	tStr := fmt.Sprintf("%d turns", turns)
	if hit > 0 {
		return fmt.Sprintf("%s (%.0f%% hit)", tStr, hit)
	}
	return tStr
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

func formatFlameLine(intensity float64, model string, innerWidth int, colors bool) string {
	barW := innerWidth / 4
	if barW > 32 {
		barW = 32
	}
	if barW < 16 {
		barW = 16
	}

	barStr, pct := thermal.FlameBar(intensity, barW)
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

	prefix := fmt.Sprintf("  Flame: %s %s  %s", barStyled, pctStr, state)
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
		formatted := thermal.CompactNumber(peak) + unitSec
		if colors {
			peakStr = "  Peak: " + theme.DarkStep12.Sprint(true, formatted)
		} else {
			peakStr = "  Peak: " + formatted
		}
	}

	availForSpark := innerWidth - len(prefix) - len(suffix) - ansi.StringWidth(peakStr) - 2
	if availForSpark > 60 {
		availForSpark = 60
	}
	if availForSpark < 20 {
		availForSpark = 20
	}

	spark := thermal.Sparkline(buckets, 59, availForSpark)
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

func calcColWidths(innerWidth int) (timeW, toolW, tokW, modW, projW, costW int) {
	timeW = 8
	toolW = 10
	tokW = 12
	costW = 9
	// Fixed spacing across six columns:
	// 2 (indent) + 8 (time) + 2 (gap) + 10 (tool) + 2 (gap) + 12 (tokens) + 2 (gap) + 2 (gap) + 9 (cost) + 2 (gap) = 51 cells
	remaining := innerWidth - 51
	if remaining < 18 {
		remaining = 18
	}
	modW = remaining * 55 / 100
	if modW < 12 {
		modW = 12
	}
	projW = remaining - modW
	if projW < 8 {
		projW = 8
	}
	return
}

func formatTickerHeader(innerWidth int, colors bool) string {
	timeW, toolW, tokW, modW, projW, costW := calcColWidths(innerWidth)

	hdr := "  " +
		padCell("TIME", timeW, false) + "  " +
		padCell("TOOL", toolW, false) + "  " +
		padCell("TOKENS", tokW, true) + "  " +
		padCell("MODEL", modW, false) + "  " +
		padCell("PROJECT", projW, false) + "  " +
		padCell("COST", costW, true)

	if colors {
		return theme.TextMuted.Sprint(true, hdr)
	}
	return hdr
}

func formatEventRow(ev thermal.LiveEvent, innerWidth int, colors bool) string {
	timeW, toolW, tokW, modW, projW, costW := calcColWidths(innerWidth)

	timeStr := ev.Timestamp.Format("15:04:05")
	toolStr := ev.Tool
	tokStr := fmt.Sprintf("+%s tok", thermal.CompactNumber(ev.Tokens))
	if ev.Tokens == 0 && ev.Turns > 0 {
		tokStr = fmt.Sprintf("+%s step", thermal.CompactNumber(int64(ev.Turns)))
	}
	modStr := loaders.CanonicalModelName(cleanModelName(ev.Model))
	if modStr == "" {
		modStr = "—"
	}
	projStr := ev.Project
	if projStr == "" {
		projStr = "—"
	}
	costStr := formatCost(ev.Cost, false)

	// Pad plain strings to exact visual cell widths FIRST
	timeCell := padCell(timeStr, timeW, false)
	toolCell := padCell(toolStr, toolW, false)
	tokCell := padCell(tokStr, tokW, true)
	modCell := padCell(modStr, modW, false)
	projCell := padCell(projStr, projW, false)
	costCell := padCell(costStr, costW, true)

	// Apply colors to the padded cells
	if colors {
		timeCell = theme.TextMuted.Sprint(true, timeCell)
		toolCell = theme.Secondary.Sprint(true, toolCell)
		tokCell = theme.Primary.SprintBold(true, tokCell)
		modCell = theme.DarkStep12.Sprint(true, modCell)
		projCell = theme.Accent.Sprint(true, projCell)
		costCell = theme.TextMuted.Sprint(true, costCell)
	}

	return "  " + timeCell + "  " + toolCell + "  " + tokCell + "  " + modCell + "  " + projCell + "  " + costCell
}
