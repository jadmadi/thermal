// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/theme"
	"github.com/jadmadi/thermal/internal/thermal"
)

var (
	densePrimary   = theme.Primary.LipGloss()
	denseSecondary = theme.Secondary.LipGloss()
	denseAccent    = theme.Accent.LipGloss()
	denseSuccess   = theme.Success.LipGloss()
	denseWarning   = theme.Warning.LipGloss()
	denseError     = theme.Error.LipGloss()
	denseCyan      = theme.Cyan.LipGloss()
	denseText      = theme.Text.LipGloss()
	denseMuted     = theme.TextMuted.LipGloss()
	denseBorder    = theme.Border.LipGloss()
)

// DenseModel is the Bubble Tea model for the 9-box high-density FinOps grid view.
type DenseModel struct {
	payload       thermal.FinOpsGridPayload
	palette       Palette
	width         int
	height        int
	metric        Metric
	rng           Range
	quitting      bool
	periodIdx     int
	periods       []string
	activeProject string
	projectList   []string
	scrollOffset  int
	showHelp      bool

	days     []thermal.DailyRow
	results  []thermal.ToolResult
	projects []thermal.ProjectDay
	yieldRep thermal.YieldReport
	pricer   thermal.Pricer
}

// NewDense constructs an initialized DenseModel.
func NewDense(payload thermal.FinOpsGridPayload, colorful bool) DenseModel {
	return NewDenseWithData(payload, nil, nil, nil, thermal.YieldReport{}, nil, colorful)
}

// NewDenseWithData constructs an initialized DenseModel with complete telemetry datasets for reactive recomputation.
func NewDenseWithData(payload thermal.FinOpsGridPayload, days []thermal.DailyRow, results []thermal.ToolResult, projects []thermal.ProjectDay, yieldRep thermal.YieldReport, pricer thermal.Pricer, colorful bool) DenseModel {
	var projList []string
	seen := make(map[string]bool)
	for _, p := range projects {
		if p.Project != "" && !seen[p.Project] {
			seen[p.Project] = true
			projList = append(projList, p.Project)
		}
	}
	sort.Strings(projList)

	return DenseModel{
		payload:     payload,
		palette:     newPalette(colorful),
		width:       80,
		height:      24,
		metric:      MetricTokens,
		rng:         Range30d,
		periodIdx:   0,
		periods:     []string{"Today", "7 Days", "30 Days", "This Month", "6 Months", "Lifetime"},
		projectList: projList,
		days:        days,
		results:     results,
		projects:    projects,
		yieldRep:    yieldRep,
		pricer:      pricer,
	}
}

func (m *DenseModel) recompute() {
	if len(m.days) == 0 && len(m.projects) == 0 {
		return
	}
	period := m.periods[m.periodIdx]
	m.payload = thermal.ComputeFinOpsGridForPeriod(m.days, m.results, m.projects, m.yieldRep, m.pricer, period, m.activeProject)
}

func (m *DenseModel) cycleProject() {
	if len(m.projectList) == 0 {
		return
	}
	if m.activeProject == "" {
		m.activeProject = m.projectList[0]
		return
	}
	for i, p := range m.projectList {
		if p == m.activeProject {
			if i+1 < len(m.projectList) {
				m.activeProject = m.projectList[i+1]
			} else {
				m.activeProject = "" // back to All
			}
			return
		}
	}
	m.activeProject = ""
}

func (m *DenseModel) cycleProjectPrev() {
	if len(m.projectList) == 0 {
		return
	}
	if m.activeProject == "" {
		m.activeProject = m.projectList[len(m.projectList)-1]
		return
	}
	for i, p := range m.projectList {
		if p == m.activeProject {
			if i-1 >= 0 {
				m.activeProject = m.projectList[i-1]
			} else {
				m.activeProject = "" // back to All
			}
			return
		}
	}
	m.activeProject = ""
}

// Init satisfies tea.Model.
func (m DenseModel) Init() tea.Cmd { return nil }

// Update handles window resizing and navigation keys.
func (m DenseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		if m.activeProject != "" {
			m.activeProject = ""
			m.recompute()
			return m, nil
		}
		m.quitting = true
		return m, tea.Quit
	case "?":
		m.showHelp = true
		return m, nil

	// Period navigation:
	case "tab", "right", "l", "]":
		m.periodIdx = (m.periodIdx + 1) % len(m.periods)
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "shift+tab", "left", "h", "[":
		m.periodIdx = (m.periodIdx - 1 + len(m.periods)) % len(m.periods)
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "1":
		m.periodIdx = 0
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "2":
		m.periodIdx = 1
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "3":
		m.periodIdx = 2
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "4":
		m.periodIdx = 3
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "5":
		m.periodIdx = 4
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "6":
		m.periodIdx = 5
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "t":
		m.periodIdx = 0
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "w":
		m.periodIdx = 1
		m.scrollOffset = 0
		m.recompute()
		return m, nil
	case "m":
		m.metric = m.metric.next()
		return m, nil
	case "r":
		m.rng = m.rng.next()
		return m, nil
	case "R":
		m.periodIdx = 0
		m.activeProject = ""
		m.scrollOffset = 0
		m.recompute()
		return m, nil

	// Project filter:
	case "p":
		m.cycleProject()
		m.recompute()
		return m, nil
	case "P":
		m.cycleProjectPrev()
		m.recompute()
		return m, nil
	case "x", "0":
		if m.activeProject != "" {
			m.activeProject = ""
			m.recompute()
		}
		return m, nil

	// Scrolling:
	case "down", "j":
		m.scrollOffset += 2
		return m, nil
	case "up", "k":
		m.scrollOffset -= 2
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return m, nil
	case "J":
		m.scrollOffset += 12
		return m, nil
	case "K":
		m.scrollOffset -= 12
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return m, nil
	case "pgdown", "pagedown", "ctrl+d", " ", "space":
		step := m.height / 2
		if step < 6 {
			step = 6
		}
		m.scrollOffset += step
		return m, nil
	case "pgup", "pageup", "ctrl+u":
		step := m.height / 2
		if step < 6 {
			step = 6
		}
		m.scrollOffset -= step
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return m, nil
	case "g", "home":
		m.scrollOffset = 0
		return m, nil
	case "G", "end":
		m.scrollOffset = 99999
		return m, nil
	}

	return m, nil
}

// helpScreen formats the interactive navigation cheatsheet modal.
func (m DenseModel) helpScreen() string {
	var b strings.Builder
	title := "Thermal FinOps 9-Box Grid — Navigation Keys"
	if m.palette.Colorful {
		b.WriteString(lipgloss.NewStyle().Foreground(densePrimary).Bold(true).Render(title))
	} else {
		b.WriteString(title)
	}
	b.WriteString("\n\n")

	rows := [][2]string{
		{"Period Selection", ""},
		{"  tab / shift+tab", "next and previous period"},
		{"  ← / → or h / l", "next and previous period"},
		{"  [ / ]", "next and previous period"},
		{"  1 – 6", "jump to period (1=Today, 2=7d, 3=30d, 4=Month, 5=6m, 6=All)"},
		{"  t / w", "jump: (t)oday, (w)eek"},
		{"", ""},
		{"Project Filter", ""},
		{"  p / P", "cycle next and previous project"},
		{"  x or 0", "clear project filter (reset to All)"},
		{"", ""},
		{"Scrolling & Movement", ""},
		{"  ↑ / ↓ or j / k", "scroll up and down by 2 lines"},
		{"  J / K", "jump up and down by card section (~12 lines)"},
		{"  pgup / pgdn", "scroll up and down by half screen"},
		{"  space / ctrl+d", "scroll down by half screen"},
		{"  ctrl+u", "scroll up by half screen"},
		{"  g / home", "scroll to top"},
		{"  G / end", "scroll to bottom"},
		{"", ""},
		{"General Controls", ""},
		{"  r / R", "r=cycle range, R=reset view & filters"},
		{"  ?", "close this help screen"},
		{"  esc / q", "close help / quit"},
	}

	for _, row := range rows {
		if row[0] == "" && row[1] == "" {
			b.WriteString("\n")
			continue
		}
		if row[1] == "" {
			if m.palette.Colorful {
				b.WriteString(lipgloss.NewStyle().Foreground(denseSecondary).Bold(true).Render(row[0]) + "\n")
			} else {
				b.WriteString(row[0] + "\n")
			}
			continue
		}
		keyPart := padTo(row[0], 22)
		desc := row[1]
		if m.palette.Colorful {
			b.WriteString(lipgloss.NewStyle().Foreground(denseText).Render(keyPart) + m.palette.Muted.Render(desc) + "\n")
		} else {
			b.WriteString(keyPart + desc + "\n")
		}
	}

	b.WriteString("\n")
	if m.palette.Colorful {
		b.WriteString(m.palette.Dim.Render("Press ? or Esc to return to the FinOps grid"))
	} else {
		b.WriteString("Press ? or Esc to return to the FinOps grid")
	}

	return b.String()
}

// View renders the 9-box FinOps grid with fixed header, scrollable body, and sticky status footer.
func (m DenseModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.showHelp {
		content := "\n" + indent(m.helpScreen(), "  ") + "\n"
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}

	w := m.width
	if w <= 0 {
		w = 80
	}
	period := "Today"
	if m.periodIdx >= 0 && m.periodIdx < len(m.periods) {
		period = m.periods[m.periodIdx]
	}

	topBar := renderPeriodBar(w, period, m.activeProject, m.palette.Colorful)
	body := renderDenseBody(m.payload, w, m.palette.Colorful)
	bodyLines := strings.Split(body, "\n")

	// Fixed header takes 3 lines: "\n " + topBar + "\n"
	headerLines := 3
	// Fixed footer takes 3 lines: "\n  " + footer + "\n"
	footerLines := 3

	avail := m.height - headerLines - footerLines
	if avail < 1 {
		avail = len(bodyLines)
	}

	maxOffset := len(bodyLines) - avail
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := m.scrollOffset
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	end := offset + avail
	if end > len(bodyLines) {
		end = len(bodyLines)
	}
	visibleBody := strings.Join(bodyLines[offset:end], "\n")

	var scrollText string
	if maxOffset > 0 {
		pct := int(math.Round(float64(offset) / float64(maxOffset) * 100))
		if offset == 0 {
			scrollText = " [Top]"
		} else if offset >= maxOffset {
			scrollText = " [Bot]"
		} else {
			scrollText = fmt.Sprintf(" [%d%%]", pct)
		}
	}

	footerText := fmt.Sprintf("• [Tab/←→/1-6] Period  ·  [p/P] Project  ·  [↑↓/j/k] Scroll%s  ·  [?] Help  ·  [q] Quit", scrollText)
	footer := m.palette.Dim.Render(footerText)

	var content string
	if m.height > 0 && len(bodyLines) > avail {
		content = "\n " + topBar + "\n\n" + visibleBody + "\n\n  " + footer + "\n"
	} else {
		content = "\n " + topBar + "\n\n" + body + "\n\n  " + footer + "\n"
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// RenderDenseFinOps renders the 9-box high-density FinOps grid view adaptively.
func RenderDenseFinOps(p thermal.FinOpsGridPayload, width int, colorful bool) string {
	period := p.Period
	if period == "" {
		period = "Today"
	}
	return RenderDenseFinOpsWithPeriod(p, width, period, p.ActiveProject, colorful)
}

// RenderDenseFinOpsWithPeriod renders the 9-box high-density FinOps grid view with an explicit period and project.
func RenderDenseFinOpsWithPeriod(p thermal.FinOpsGridPayload, width int, activePeriod string, activeProject string, colorful bool) string {
	if width <= 0 {
		width = 80
	}

	// 1. Top Period Selector Bar
	topBar := renderPeriodBar(width, activePeriod, activeProject, colorful)
	body := renderDenseBody(p, width, colorful)
	pal := newPalette(colorful)
	footer := pal.Dim.Render("• Tip: Run 'thermal dashboard stats' for interactive drill-down and live filtering.")

	return "\n " + topBar + "\n\n" + body + "\n\n  " + footer + "\n"
}

// renderDenseBody renders the headline summary card and the 9 data boxes.
func renderDenseBody(p thermal.FinOpsGridPayload, width int, colorful bool) string {
	// 2. Headline Summary Card (Amber Rounded Box)
	summaryCard := renderHeadlineSummary(p, width, colorful)

	// Determine column layout based on terminal width:
	// >= 135 columns: 3 columns x 3 rows (full 9-box grid)
	// 90..134 columns: 2 columns (4 rows of 2 + 1 bottom box)
	// < 90 columns: 1 column (vertical stack of cards)
	var cols int
	var boxWidth int
	switch {
	case width >= 135:
		cols = 3
		boxWidth = (width - 4) / 3
		if boxWidth > 54 {
			boxWidth = 54
		}
	case width >= 90:
		cols = 2
		boxWidth = (width - 2) / 2
		if boxWidth > 64 {
			boxWidth = 64
		}
	default:
		cols = 1
		boxWidth = width - 2
		if boxWidth < 30 {
			boxWidth = 30
		}
	}

	const targetHeight = 11

	// Box 1: Daily Activity (Blue)
	b1 := renderDailyActivityCard(p, boxWidth, targetHeight, colorful)

	// Box 2: By Project (Green)
	b2 := renderProjectCard(p, boxWidth, targetHeight, colorful)

	// Box 3: Assistant Mix (Yellow)
	b3 := renderAssistantMixCard(p, boxWidth, targetHeight, colorful)

	// Box 4: By Model (Magenta)
	b4 := renderModelCard(p, boxWidth, targetHeight, colorful)

	// Box 5: Prompt Cache & Savings (Cyan)
	b5 := renderCacheSavingsCard(p, boxWidth, targetHeight, colorful)

	// Box 6: Code Yield & Velocity (Mint)
	b6 := renderCodeYieldCard(p, boxWidth, targetHeight, colorful)

	// Box 7: Work Receipts & Tests (Indigo)
	b7 := renderWorkReceiptsCard(p, boxWidth, targetHeight, colorful)

	// Box 8: Subscription Replay (Orange)
	b8 := renderSubscriptionReplayCard(p, boxWidth, targetHeight, colorful)

	// Box 9: Token Composition (Violet)
	b9 := renderTokenCompositionCard(p, boxWidth, targetHeight, colorful)

	boxes := []string{b1, b2, b3, b4, b5, b6, b7, b8, b9}

	var gridOut string
	if cols == 3 {
		row1 := joinColumns([]string{boxes[0], boxes[1], boxes[2]}, boxWidth)
		row2 := joinColumns([]string{boxes[3], boxes[4], boxes[5]}, boxWidth)
		row3 := joinColumns([]string{boxes[6], boxes[7], boxes[8]}, boxWidth)
		gridOut = row1 + "\n\n" + row2 + "\n\n" + row3
	} else if cols == 2 {
		row1 := joinColumns([]string{boxes[0], boxes[1]}, boxWidth)
		row2 := joinColumns([]string{boxes[2], boxes[3]}, boxWidth)
		row3 := joinColumns([]string{boxes[4], boxes[5]}, boxWidth)
		row4 := joinColumns([]string{boxes[6], boxes[7]}, boxWidth)
		row5 := boxes[8]
		gridOut = row1 + "\n\n" + row2 + "\n\n" + row3 + "\n\n" + row4 + "\n\n" + row5
	} else {
		gridOut = strings.Join(boxes, "\n\n")
	}

	return " " + summaryCard + "\n\n" + gridOut
}

// renderPeriodBar renders the top filter tabs: [ Today ] 7 Days 30 Days ... | [p] All
func renderPeriodBar(width int, activePeriod string, activeProject string, colorful bool) string {
	var periods = []string{"Today", "7 Days", "30 Days", "This Month", "6 Months", "Lifetime"}
	var items []string

	for _, p := range periods {
		if p == activePeriod {
			if colorful {
				items = append(items, lipgloss.NewStyle().Foreground(densePrimary).Bold(true).Render("[ "+p+" ]"))
			} else {
				items = append(items, "[ "+p+" ]")
			}
		} else {
			if colorful {
				items = append(items, lipgloss.NewStyle().Foreground(denseMuted).Render(p))
			} else {
				items = append(items, p)
			}
		}
	}

	left := strings.Join(items, "   ")
	projLabel := "All"
	if activeProject != "" {
		projLabel = thermal.ProjectSlug(activeProject)
	}
	rightText := "[p] " + projLabel
	var right string
	if colorful {
		right = lipgloss.NewStyle().Foreground(denseBorder).Render("|  ") +
			lipgloss.NewStyle().Foreground(denseSecondary).Render(rightText)
	} else {
		right = "|  " + rightText
	}

	totalInner := width - 2
	leftLen := len([]rune(ansi.Strip(left)))
	rightLen := len([]rune(ansi.Strip(right)))
	gap := totalInner - leftLen - rightLen
	if gap < 2 {
		gap = 2
	}

	return left + strings.Repeat(" ", gap) + right
}

// renderHeadlineSummary renders the amber-bordered summary card
func renderHeadlineSummary(p thermal.FinOpsGridPayload, width int, colorful bool) string {
	innerWidth := width - 4
	if innerWidth < 50 {
		innerWidth = 50
	}

	period := p.Period
	if period == "" {
		period = "Today"
	}
	subTitle := period
	if p.ActiveProject != "" {
		subTitle += " · " + p.ActiveProject
	}

	var titleLine string
	if colorful {
		titleLine = lipgloss.NewStyle().Foreground(densePrimary).Bold(true).Render("Thermal") + "  " +
			lipgloss.NewStyle().Foreground(denseSecondary).Render("FinOps 9-Box Grid") + "  ·  " +
			lipgloss.NewStyle().Foreground(denseMuted).Render(subTitle)
	} else {
		titleLine = "Thermal  FinOps 9-Box Grid · " + subTitle
	}

	costStr := fmtCost(p.Today.Cost)
	hitStr := fmt.Sprintf("%.1f%%", p.Today.CacheHitRate)
	callsStr := fmt.Sprintf("%d", p.Today.Calls)
	sessionsStr := fmt.Sprintf("%d", max(1, p.Today.Sessions))

	var row1 string
	if colorful {
		row1 = lipgloss.NewStyle().Foreground(densePrimary).Bold(true).Render(costStr) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("cost") + "   " +
			lipgloss.NewStyle().Foreground(denseText).Bold(true).Render(callsStr) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("calls") + "   " +
			lipgloss.NewStyle().Foreground(denseMuted).Render("At least "+sessionsStr+" sessions") + "   " +
			lipgloss.NewStyle().Foreground(denseText).Bold(true).Render(hitStr) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("cache hit")
	} else {
		row1 = fmt.Sprintf("%s cost   %s calls   At least %s sessions   %s cache hit", costStr, callsStr, sessionsStr, hitStr)
	}

	inTok := thermal.CompactNumber(p.Today.InputTokens)
	outTok := thermal.CompactNumber(p.Today.OutputTokens)
	cacheTok := thermal.CompactNumber(p.Today.CachedTokens)
	writeTok := thermal.CompactNumber(p.Today.WriteTokens)

	var row2 string
	if colorful {
		row2 = lipgloss.NewStyle().Foreground(denseText).Render(inTok) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("in") + "   " +
			lipgloss.NewStyle().Foreground(denseText).Render(outTok) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("out") + "   " +
			lipgloss.NewStyle().Foreground(denseText).Render(cacheTok) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("cached") + "   " +
			lipgloss.NewStyle().Foreground(denseText).Render(writeTok) + " " +
			lipgloss.NewStyle().Foreground(denseBorder).Render("written")
	} else {
		row2 = fmt.Sprintf("%s in   %s out   %s cached   %s written", inTok, outTok, cacheTok, writeTok)
	}

	borderColor := theme.Primary.Hex // Blaze Amber
	return renderFramedBox([]string{titleLine, row1, row2}, innerWidth, borderColor, colorful)
}

// renderFramedBox wraps lines in a rounded border
func renderFramedBox(lines []string, innerWidth int, borderColor string, colorful bool) string {
	var borderStyle lipgloss.Style
	if colorful {
		borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
	}

	top := borderStyle.Render("╭" + strings.Repeat("─", innerWidth+2) + "╮")
	bottom := borderStyle.Render("╰" + strings.Repeat("─", innerWidth+2) + "╯")

	var b strings.Builder
	b.WriteString(top)
	b.WriteString("\n")

	for _, line := range lines {
		stripped := ansi.Strip(line)
		pad := innerWidth - len([]rune(stripped))
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderStyle.Render("│ ") + line + strings.Repeat(" ", pad) + borderStyle.Render(" │\n"))
	}

	b.WriteString(bottom)
	return b.String()
}

// renderFinOpsCard renders a rounded 9-box section card with tinted title and right-aligned header
func renderFinOpsCard(title string, titleColor string, rightHeader string, lines []string, width int, targetHeight int, colorful bool) string {
	if width < 30 {
		width = 30
	}
	innerWidth := width - 4

	var titleStyled, rightStyled, ruleStyled, borderCharStyle lipgloss.Style
	if colorful {
		titleStyled = lipgloss.NewStyle().Foreground(lipgloss.Color(titleColor)).Bold(true)
		rightStyled = lipgloss.NewStyle().Foreground(denseMuted)
		ruleStyled = lipgloss.NewStyle().Foreground(denseBorder)
		borderCharStyle = lipgloss.NewStyle().Foreground(denseBorder)
	}

	titleText := titleStyled.Render(title)
	rightText := rightStyled.Render(rightHeader)

	titleLen := len([]rune(title))
	rightLen := len([]rune(rightHeader))

	// Top line format: ╭─ Title ──...── rightHeader ─╮
	ruleSpaces := innerWidth - titleLen - rightLen - 2
	if rightHeader == "" {
		ruleSpaces = innerWidth - titleLen - 1
	}
	if ruleSpaces < 1 {
		ruleSpaces = 1
	}

	var topLine string
	if rightHeader != "" {
		topLine = borderCharStyle.Render("╭─ ") + titleText + " " + ruleStyled.Render(strings.Repeat("─", ruleSpaces)) + " " + rightText + borderCharStyle.Render(" ─╮")
	} else {
		topLine = borderCharStyle.Render("╭─ ") + titleText + " " + ruleStyled.Render(strings.Repeat("─", ruleSpaces)) + borderCharStyle.Render("─╮")
	}

	bottomLine := borderCharStyle.Render("╰" + strings.Repeat("─", width-2) + "╯")

	var b strings.Builder
	b.WriteString(topLine)
	b.WriteString("\n")

	for i := 0; i < targetHeight; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		stripped := ansi.Strip(line)
		pad := innerWidth - len([]rune(stripped))
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderCharStyle.Render("│ ") + line + strings.Repeat(" ", pad) + borderCharStyle.Render(" │\n"))
	}

	b.WriteString(bottomLine)
	return b.String()
}

// renderHeatBar generates a multi-stop color gradient horizontal bar (Blue -> Orange -> Yellow)
func renderHeatBar(val, maxVal float64, barWidth int, colorful bool) string {
	if maxVal <= 0 || val <= 0 || barWidth <= 0 {
		return strings.Repeat(" ", barWidth)
	}
	ratio := val / maxVal
	if ratio > 1.0 {
		ratio = 1.0
	}
	totalCells := ratio * float64(barWidth)
	fullCells := int(totalCells)
	remainder := totalCells - float64(fullCells)

	fracChars := []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}
	fracIdx := int(remainder * 8.0)
	if fracIdx > 7 {
		fracIdx = 7
	}

	var b strings.Builder
	for i := 0; i < barWidth; i++ {
		var r rune
		if i < fullCells {
			r = '█'
		} else if i == fullCells && fracIdx > 0 {
			r = fracChars[fracIdx]
		} else {
			r = ' '
		}

		if r == ' ' {
			b.WriteRune(' ')
			continue
		}

		if !colorful {
			b.WriteRune(r)
		} else {
			var col color.Color
			switch {
			case i < 2:
				col = theme.DarkStep5.LipGloss()
			case i < 4:
				col = denseSecondary
			default:
				col = densePrimary
			}
			b.WriteString(lipgloss.NewStyle().Foreground(col).Render(string(r)))
		}
	}
	return b.String()
}

// Box 1: Daily Activity
func renderDailyActivityCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxCost float64
	for _, d := range p.DailyHistory {
		if d.Cost > maxCost {
			maxCost = d.Cost
		}
	}
	if maxCost <= 0 {
		maxCost = 1.0
	}

	var lines []string
	for _, d := range p.DailyHistory {
		bar := renderHeatBar(d.Cost, maxCost, 4, colorful)
		dateStr := d.Date
		costStr := fmtCost(d.Cost)
		callsStr := fmt.Sprintf("%d", d.Calls)

		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(denseSecondary).Render(costStr)
		} else {
			costStyled = costStr
		}

		// Layout: Bar Date ... Cost Calls
		line := fmt.Sprintf("%s %s %10s %5s", bar, dateStr, costStyled, callsStr)
		lines = append(lines, line)
	}

	for len(lines) < targetHeight-1 {
		lines = append(lines, "")
	}

	// Footer: Showing 1-10 of X days scanned · newest first
	footerText := fmt.Sprintf("Showing 1-%d of %d days scanned · newest first", min(10, p.TotalDaysScanned), max(1, p.TotalDaysScanned))
	if colorful {
		lines = append(lines, lipgloss.NewStyle().Foreground(denseMuted).Render(footerText))
	} else {
		lines = append(lines, footerText)
	}

	return renderFinOpsCard("Daily Activity", theme.Secondary.Hex, "cost calls", lines, width, targetHeight, colorful)
}

// Box 2: By Project
func renderProjectCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxCost float64
	for _, pr := range p.ProjectBreakdown {
		if pr.Cost > maxCost {
			maxCost = pr.Cost
		}
	}
	if maxCost <= 0 {
		maxCost = 1.0
	}

	var lines []string
	for _, pr := range p.ProjectBreakdown {
		bar := renderHeatBar(pr.Cost, maxCost, 4, colorful)
		name := pr.Name
		if ansi.StringWidth(name) > 10 {
			name = ansi.Truncate(name, 10, "…")
		}
		costStr := fmtCost(pr.Cost)
		avgStr := fmtCost(pr.AvgCost)
		sessStr := fmt.Sprintf("%d", pr.Sessions)
		ohStr := fmtOverhead(pr.Overhead)

		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(denseSecondary).Render(costStr)
		} else {
			costStyled = costStr
		}

		line := fmt.Sprintf("%s %-10s %7s %7s %2s %6s", bar, name, costStyled, avgStr, sessStr, ohStr)
		lines = append(lines, line)
	}

	return renderFinOpsCard("By Project", theme.Success.Hex, "cost avg/s session overhead", lines, width, targetHeight, colorful)
}

// Box 3: Assistant Mix
func renderAssistantMixCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	if len(p.ToolMix) == 0 {
		msg := "No assistant activity recorded"
		if colorful {
			msg = lipgloss.NewStyle().Foreground(denseMuted).Render(msg)
		}
		return renderFinOpsCard("Assistant Mix", theme.Warning.Hex, "share tokens cost streak", []string{msg}, width, targetHeight, colorful)
	}

	var maxTok int64
	for _, t := range p.ToolMix {
		if t.Tokens > maxTok {
			maxTok = t.Tokens
		}
	}
	if maxTok <= 0 {
		maxTok = 1
	}

	var lines []string
	for _, t := range p.ToolMix {
		bar := renderHeatBar(float64(t.Tokens), float64(maxTok), 4, colorful)
		name := t.Name
		if len([]rune(name)) > 11 {
			name = string([]rune(name)[:11])
		}
		shareStr := fmt.Sprintf("%5.1f%%", t.Share)
		tokStr := thermal.CompactNumber(t.Tokens)
		costStr := fmtCost(t.Cost)

		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(denseSecondary).Render(costStr)
		} else {
			costStyled = costStr
		}

		streakStr := "-"
		if t.Streak > 0 {
			streakStr = fmt.Sprintf("%dd", t.Streak)
		}

		line := fmt.Sprintf("%s %-11s %6s %7s %7s %4s", bar, name, shareStr, tokStr, costStyled, streakStr)
		lines = append(lines, line)
	}

	return renderFinOpsCard("Assistant Mix", theme.Warning.Hex, "share tokens cost streak", lines, width, targetHeight, colorful)
}

// Box 4: By Model
func renderModelCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxCost float64
	for _, m := range p.ModelBreakdown {
		if m.Cost > maxCost {
			maxCost = m.Cost
		}
	}
	if maxCost <= 0 {
		maxCost = 1.0
	}

	var lines []string
	for _, m := range p.ModelBreakdown {
		bar := renderHeatBar(m.Cost, maxCost, 4, colorful)
		name := m.Name
		if len([]rune(name)) > 9 {
			name = string([]rune(name)[:9])
		}
		costStr := fmtCost(m.Cost)
		if m.IsEstimated {
			costStr = "~" + costStr
		}
		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(denseSecondary).Render(costStr)
		} else {
			costStyled = costStr
		}

		cacheStr := fmt.Sprintf("%.1f%%", m.CachePct)
		callsStr := fmt.Sprintf("%3d", m.Calls)

		line := fmt.Sprintf("%s %-9s %7s %5s %4s %6s %5s", bar, name, costStyled, cacheStr, callsStr, m.OneShot, m.TokPerS)
		lines = append(lines, line)
	}

	for len(lines) < targetHeight-2 {
		lines = append(lines, "")
	}

	// Footers
	f1 := "~ estimated cost (priced from estimated tokens)"
	f2 := "~ Effective Tok/s: generated tokens ÷ time the ..."
	if colorful {
		lines = append(lines, lipgloss.NewStyle().Foreground(denseMuted).Render(f1))
		lines = append(lines, lipgloss.NewStyle().Foreground(denseMuted).Render(f2))
	} else {
		lines = append(lines, f1, f2)
	}

	return renderFinOpsCard("By Model", theme.Primary.Hex, "cost cache calls 1-shot Tok/s", lines, width, targetHeight, colorful)
}

// Box 5: Prompt Cache & Savings
func renderCacheSavingsCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var lines []string

	hitRateStr := fmt.Sprintf("%.1f%%", p.Cache.HitRate)
	savedStr := fmt.Sprintf("~$%.2f", p.Cache.EstimatedSavings)
	if colorful {
		hitRateStr = lipgloss.NewStyle().Foreground(denseCyan).Bold(true).Render(hitRateStr)
		savedStr = lipgloss.NewStyle().Foreground(denseSuccess).Bold(true).Render(savedStr)
	}

	lines = append(lines, fmt.Sprintf("Hit Rate:      %s", hitRateStr))
	lines = append(lines, fmt.Sprintf("Cache Read:    %s tok", thermal.CompactNumber(p.Cache.CacheReadTokens)))
	lines = append(lines, fmt.Sprintf("Cache Write:   %s tok", thermal.CompactNumber(p.Cache.CacheWriteTokens)))
	lines = append(lines, fmt.Sprintf("Uncached In:   %s tok", thermal.CompactNumber(p.Cache.InputTokens)))
	lines = append(lines, fmt.Sprintf("Prompt Total:  %s tok", thermal.CompactNumber(p.Cache.CacheReadTokens+p.Cache.InputTokens)))
	lines = append(lines, fmt.Sprintf("Est. Savings:  %s saved", savedStr))

	effic := "MODERATE"
	switch {
	case p.Cache.HitRate >= 80:
		effic = "EXCELLENT"
	case p.Cache.HitRate >= 50:
		effic = "HEALTHY"
	case p.Cache.HitRate == 0:
		effic = "NONE"
	}
	if colorful {
		switch effic {
		case "EXCELLENT":
			effic = lipgloss.NewStyle().Foreground(denseSuccess).Bold(true).Render(effic)
		case "HEALTHY":
			effic = lipgloss.NewStyle().Foreground(denseCyan).Render(effic)
		default:
			effic = lipgloss.NewStyle().Foreground(denseMuted).Render(effic)
		}
	}
	lines = append(lines, fmt.Sprintf("Efficiency:    %s", effic))

	for len(lines) < targetHeight-1 {
		lines = append(lines, "")
	}

	footer := "~ prompt cache saves ~$2.20/M read tok"
	if colorful {
		footer = lipgloss.NewStyle().Foreground(denseMuted).Render(footer)
	}
	lines = append(lines, footer)

	return renderFinOpsCard("Prompt Cache & Savings", theme.Cyan.Hex, "hit% saved", lines, width, targetHeight, colorful)
}

// Box 6: Code Yield & Velocity
func renderCodeYieldCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var lines []string

	badge := p.Yield.Efficiency
	if badge == "" {
		badge = "[BALANCED]"
	}
	var badgeStyled string
	if colorful {
		switch {
		case strings.Contains(badge, "HIGH"):
			badgeStyled = lipgloss.NewStyle().Foreground(denseSuccess).Bold(true).Render(badge)
		case strings.Contains(badge, "BALANCED"):
			badgeStyled = lipgloss.NewStyle().Foreground(denseCyan).Bold(true).Render(badge)
		case strings.Contains(badge, "VERBOSE"):
			badgeStyled = lipgloss.NewStyle().Foreground(denseWarning).Bold(true).Render(badge)
		default:
			badgeStyled = lipgloss.NewStyle().Foreground(denseSecondary).Bold(true).Render(badge)
		}
	} else {
		badgeStyled = badge
	}

	lines = append(lines, fmt.Sprintf("Efficiency:   %s", badgeStyled))
	lines = append(lines, fmt.Sprintf("Net Lines:    %+d lines", p.Yield.NetLines))
	lines = append(lines, fmt.Sprintf("Diff Scope:   +%d / -%d lines", p.Yield.LinesAdded, p.Yield.LinesDeleted))
	if p.Yield.TokensPerNet > 0 {
		lines = append(lines, fmt.Sprintf("Token Yield:  %.0f tok/net line", p.Yield.TokensPerNet))
	} else {
		lines = append(lines, "Token Yield:  -")
	}
	if p.Yield.TokensPerGross > 0 {
		lines = append(lines, fmt.Sprintf("Gross Yield:  %.0f tok/gross line", p.Yield.TokensPerGross))
	} else {
		lines = append(lines, "Gross Yield:  -")
	}
	lines = append(lines, fmt.Sprintf("Total Burn:   %s tokens", thermal.CompactNumber(p.Yield.Tokens)))

	status := p.Yield.Status
	if status == "" {
		status = "MEASURED"
	}
	lines = append(lines, fmt.Sprintf("Status:       %s", status))

	for len(lines) < targetHeight-1 {
		lines = append(lines, "")
	}

	footer := "• tokens burned per net line added"
	if colorful {
		footer = lipgloss.NewStyle().Foreground(denseMuted).Render(footer)
	}
	lines = append(lines, footer)

	return renderFinOpsCard("Code Yield & Velocity", theme.Success.Hex, "yield lines net", lines, width, targetHeight, colorful)
}

// Box 7: Work Receipts & Tests
func renderWorkReceiptsCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var lines []string

	rateStr := fmt.Sprintf("%.1f%% verified", p.Receipt.VerifiedRate)
	if colorful {
		if p.Receipt.VerifiedRate >= 80 {
			rateStr = lipgloss.NewStyle().Foreground(denseSuccess).Bold(true).Render(rateStr)
		} else if p.Receipt.VerifiedRate >= 50 {
			rateStr = lipgloss.NewStyle().Foreground(denseCyan).Bold(true).Render(rateStr)
		} else {
			rateStr = lipgloss.NewStyle().Foreground(denseWarning).Render(rateStr)
		}
	}

	lines = append(lines, fmt.Sprintf("Verification: %s", rateStr))

	t1Str := fmt.Sprintf("%d sessions", p.Receipt.VerifiedCount)
	t2Str := fmt.Sprintf("%d sessions", p.Receipt.ClaimedCount)
	t3Str := fmt.Sprintf("%d sessions", p.Receipt.FailedCount)
	if colorful {
		t1Str = lipgloss.NewStyle().Foreground(denseSuccess).Render(t1Str)
		t2Str = lipgloss.NewStyle().Foreground(denseWarning).Render(t2Str)
		if p.Receipt.FailedCount > 0 {
			t3Str = lipgloss.NewStyle().Foreground(denseError).Render(t3Str)
		} else {
			t3Str = lipgloss.NewStyle().Foreground(denseMuted).Render(t3Str)
		}
	}

	lines = append(lines, fmt.Sprintf("Tier 1:       %s", t1Str))
	lines = append(lines, fmt.Sprintf("Tier 2:       %s", t2Str))
	lines = append(lines, fmt.Sprintf("Tier 3:       %s", t3Str))
	lines = append(lines, fmt.Sprintf("Verified Tok: %s tokens", thermal.CompactNumber(p.Receipt.TokensVerified)))

	spendEff := p.Receipt.SpendEfficiency
	if spendEff == "" {
		spendEff = "EXPLORATORY"
	}
	lines = append(lines, fmt.Sprintf("Spend Effic:  %s", spendEff))

	tools := p.Receipt.TestTools
	if tools == "" {
		tools = "None detected"
	}
	if len([]rune(tools)) > 22 {
		tools = string([]rune(tools)[:19]) + "..."
	}
	lines = append(lines, fmt.Sprintf("Test Tools:   %s", tools))

	for len(lines) < targetHeight-1 {
		lines = append(lines, "")
	}

	footer := "• zero-exit test/linter verification"
	if colorful {
		footer = lipgloss.NewStyle().Foreground(denseMuted).Render(footer)
	}
	lines = append(lines, footer)

	return renderFinOpsCard("Work Receipts & Tests", theme.Accent.Hex, "verified rate", lines, width, targetHeight, colorful)
}

// Box 8: Subscription Replay
func renderSubscriptionReplayCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var lines []string

	verdict := p.Replay.CapacityVerdict
	if verdict == "" {
		verdict = "PASS"
	}
	var verdictStyled string
	if colorful {
		switch verdict {
		case "PASS":
			verdictStyled = lipgloss.NewStyle().Foreground(denseSuccess).Bold(true).Render(verdict)
		case "DEGRADED":
			verdictStyled = lipgloss.NewStyle().Foreground(denseWarning).Bold(true).Render(verdict)
		default:
			verdictStyled = lipgloss.NewStyle().Foreground(denseError).Bold(true).Render(verdict)
		}
	} else {
		verdictStyled = verdict
	}

	lines = append(lines, fmt.Sprintf("Capacity:     %s", verdictStyled))

	if len(p.Replay.Plans) == 0 {
		lines = append(lines, "No replay plans configured")
	} else {
		for _, pl := range p.Replay.Plans {
			name := pl.Name
			if len([]rune(name)) > 19 {
				name = string([]rune(name)[:19])
			}

			planVerdict := pl.CapacityVerdict
			var pvStyled string
			if colorful {
				switch planVerdict {
				case "PASS":
					pvStyled = lipgloss.NewStyle().Foreground(denseSuccess).Render("PASS")
				case "DEGRADED":
					pvStyled = lipgloss.NewStyle().Foreground(denseWarning).Render("DEGR")
				default:
					pvStyled = lipgloss.NewStyle().Foreground(denseError).Render("FAIL")
				}
			} else {
				if planVerdict == "DEGRADED" {
					pvStyled = "DEGR"
				} else {
					pvStyled = planVerdict
				}
			}

			deltaSign := "+"
			if pl.CostDelta < 0 {
				deltaSign = "-"
			}
			deltaStr := fmt.Sprintf("%s$%.0f/mo", deltaSign, math.Abs(pl.CostDelta))

			line := fmt.Sprintf("%-19s %-4s %9s", name, pvStyled, deltaStr)
			lines = append(lines, line)
		}
	}

	for len(lines) < targetHeight-1 {
		lines = append(lines, "")
	}

	footer := "• 30d simulation against commercial caps"
	if colorful {
		footer = lipgloss.NewStyle().Foreground(denseMuted).Render(footer)
	}
	lines = append(lines, footer)

	return renderFinOpsCard("Subscription Replay", theme.Secondary.Hex, "verdict delta", lines, width, targetHeight, colorful)
}

// Box 9: Token Composition
func renderTokenCompositionCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	type compRow struct {
		name   string
		tokens int64
	}
	rows := []compRow{
		{"Cache Read", p.Composition.CacheReadTokens},
		{"Uncached In", p.Composition.InputTokens},
		{"Output", p.Composition.OutputTokens},
		{"Reasoning", p.Composition.ReasoningTokens},
		{"Cache Write", p.Composition.CacheWriteTokens},
	}

	var maxTok int64
	for _, r := range rows {
		if r.tokens > maxTok {
			maxTok = r.tokens
		}
	}
	if maxTok <= 0 {
		maxTok = 1
	}

	var lines []string
	for _, r := range rows {
		bar := renderHeatBar(float64(r.tokens), float64(maxTok), 4, colorful)
		tokStr := thermal.CompactNumber(r.tokens)
		pct := 0.0
		if p.TotalTokens > 0 {
			pct = (float64(r.tokens) / float64(p.TotalTokens)) * 100.0
		}
		pctStr := fmt.Sprintf("%5.1f%%", pct)
		if colorful {
			pctStr = lipgloss.NewStyle().Foreground(denseSecondary).Render(pctStr)
		}
		line := fmt.Sprintf("%s %-12s %8s %6s", bar, r.name, tokStr, pctStr)
		lines = append(lines, line)
	}

	reasRatio := fmt.Sprintf("%.1f%% of output", p.Composition.ReasoningRatio)
	lines = append(lines, fmt.Sprintf("Reasoning/Out: %s", reasRatio))
	lines = append(lines, fmt.Sprintf("Total Tokens:  %s tokens", thermal.CompactNumber(p.TotalTokens)))

	for len(lines) < targetHeight-1 {
		lines = append(lines, "")
	}

	footer := "• disjoint token taxonomy (zero overlap)"
	if colorful {
		footer = lipgloss.NewStyle().Foreground(denseMuted).Render(footer)
	}
	lines = append(lines, footer)

	return renderFinOpsCard("Token Composition", theme.Primary.Hex, "tokens share", lines, width, targetHeight, colorful)
}

func fmtCost(c float64) string {
	if c <= 0 {
		return "$0.00"
	}
	if c < 0.01 {
		return fmt.Sprintf("$%.3f", c)
	}
	if c >= 1000 {
		return fmt.Sprintf("$%.0f", c)
	}
	return fmt.Sprintf("$%.2f", c)
}

func fmtOverhead(oh int64) string {
	if oh <= 0 {
		return "-"
	}
	return thermal.CompactNumber(oh)
}
