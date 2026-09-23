// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/thermal"
)

// DenseModel is the Bubble Tea model for the 9-box high-density FinOps grid view.
type DenseModel struct {
	payload  thermal.FinOpsGridPayload
	palette  Palette
	width    int
	height   int
	metric   Metric
	rng      Range
	quitting bool
}

// NewDense constructs an initialized DenseModel.
func NewDense(payload thermal.FinOpsGridPayload, colorful bool) DenseModel {
	return DenseModel{
		payload: payload,
		palette: newPalette(colorful),
		width:   80,
		height:  24,
		metric:  MetricTokens,
		rng:     Range30d,
	}
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
	case tea.KeyPressMsg:
		switch msg.Text {
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "m":
			if m.metric == MetricTokens {
				m.metric = MetricCost
			} else {
				m.metric = MetricTokens
			}
			return m, nil
		case "r":
			switch m.rng {
			case Range30d:
				m.rng = Range90d
			case Range90d:
				m.rng = Range1y
			case Range1y:
				m.rng = RangeAll
			default:
				m.rng = Range30d
			}
			return m, nil
		}
	}
	return m, nil
}

// View renders the 9-box FinOps grid.
func (m DenseModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	w := m.width
	if w <= 0 {
		w = 80
	}
	content := RenderDenseFinOps(m.payload, w, m.palette.Colorful)
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// RenderDenseFinOps renders the 9-box high-density FinOps grid view adaptively.
func RenderDenseFinOps(p thermal.FinOpsGridPayload, width int, colorful bool) string {
	if width <= 0 {
		width = 80
	}

	// 1. Top Period Selector Bar
	topBar := renderPeriodBar(width, "Today", colorful)

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

	// Box 3: By Activity (Yellow)
	b3 := renderActivityCard(p, boxWidth, targetHeight, colorful)

	// Box 4: By Model (Magenta)
	b4 := renderModelCard(p, boxWidth, targetHeight, colorful)

	// Box 5: MCP Servers (Purple)
	b5 := renderMCPCard(p, boxWidth, targetHeight, colorful)

	// Box 6: Core Tools (Cyan)
	b6 := renderCoreToolsCard(p, boxWidth, targetHeight, colorful)

	// Box 7: Shell Commands (Orange)
	b7 := renderShellCommandsCard(p, boxWidth, targetHeight, colorful)

	// Box 8: Skills & Agents (Lavender)
	b8 := renderSkillsAgentsCard(p, boxWidth, targetHeight, colorful)

	// Box 9: Workflow (Violet)
	b9 := renderWorkflowCard(p, boxWidth, targetHeight, colorful)

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

	pal := newPalette(colorful)
	footer := pal.Dim.Render("• Navigation: q=quit · m=metric · r=range · FinOps 9-Box Grid")

	return "\n " + topBar + "\n\n " + summaryCard + "\n\n" + gridOut + "\n\n  " + footer + "\n"
}

// renderPeriodBar renders the top filter tabs: [ Today ] 7 Days 30 Days ... | [p] All
func renderPeriodBar(width int, activePeriod string, colorful bool) string {
	var periods = []string{"Today", "7 Days", "30 Days", "This Month", "6 Months", "Lifetime"}
	var items []string

	for _, p := range periods {
		if p == activePeriod {
			if colorful {
				items = append(items, lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true).Render("[ "+p+" ]"))
			} else {
				items = append(items, "[ "+p+" ]")
			}
		} else {
			if colorful {
				items = append(items, lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render(p))
			} else {
				items = append(items, p)
			}
		}
	}

	left := strings.Join(items, "   ")
	right := "[p] All"
	if colorful {
		right = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("|  ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render("[p] All")
	} else {
		right = "|  [p] All"
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

	var titleLine string
	if colorful {
		titleLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true).Render("Thermal") + "  " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render("Today")
	} else {
		titleLine = "Thermal  Today"
	}

	costStr := fmtCost(p.Today.Cost)
	hitStr := fmt.Sprintf("%.1f%%", p.Today.CacheHitRate)
	callsStr := fmt.Sprintf("%d", p.Today.Calls)
	sessionsStr := fmt.Sprintf("%d", max(1, p.Today.Sessions))

	var row1 string
	if colorful {
		row1 = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true).Render(costStr) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("cost") + "   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true).Render(callsStr) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("calls") + "   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render("At least "+sessionsStr+" sessions") + "   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true).Render(hitStr) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("cache hit")
	} else {
		row1 = fmt.Sprintf("%s cost   %s calls   At least %s sessions   %s cache hit", costStr, callsStr, sessionsStr, hitStr)
	}

	inTok := thermal.CompactNumber(p.Today.InputTokens)
	outTok := thermal.CompactNumber(p.Today.OutputTokens)
	cacheTok := thermal.CompactNumber(p.Today.CachedTokens)
	writeTok := thermal.CompactNumber(p.Today.WriteTokens)

	var row2 string
	if colorful {
		row2 = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render(inTok) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("in") + "   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render(outTok) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("out") + "   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render(cacheTok) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("cached") + "   " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render(writeTok) + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render("written")
	} else {
		row2 = fmt.Sprintf("%s in   %s out   %s cached   %s written", inTok, outTok, cacheTok, writeTok)
	}

	borderColor := "#d97706" // Amber
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
		rightStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
		ruleStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
		borderCharStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
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
			var col string
			switch {
			case i < 2:
				col = "#38bdf8" // Cyan / Blue
			case i < 4:
				col = "#f97316" // Orange
			default:
				col = "#fbbf24" // Yellow
			}
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(string(r)))
		}
	}
	return b.String()
}

// Box 1: Daily Activity (Blue)
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
			costStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Render(costStr)
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
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render(footerText))
	} else {
		lines = append(lines, footerText)
	}

	return renderFinOpsCard("Daily Activity", "#60a5fa", "cost calls", lines, width, targetHeight, colorful)
}

// Box 2: By Project (Green)
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
		if len([]rune(name)) > 10 {
			name = ".../" + string([]rune(name)[len([]rune(name))-6:])
		}
		costStr := fmtCost(pr.Cost)
		avgStr := fmtCost(pr.AvgCost)
		sessStr := fmt.Sprintf("%d", pr.Sessions)
		ohStr := fmtOverhead(pr.Overhead)

		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Render(costStr)
		} else {
			costStyled = costStr
		}

		line := fmt.Sprintf("%s %-10s %7s %7s %2s %6s", bar, name, costStyled, avgStr, sessStr, ohStr)
		lines = append(lines, line)
	}

	return renderFinOpsCard("By Project", "#4ade80", "cost avg/s session overhead", lines, width, targetHeight, colorful)
}

// Box 3: By Activity (Yellow)
func renderActivityCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxCost float64
	for _, a := range p.Taxonomy {
		if a.Cost > maxCost {
			maxCost = a.Cost
		}
	}
	if maxCost <= 0 {
		maxCost = 1.0
	}

	// Distinct colors matching reference screenshot
	catColors := map[thermal.ActivityCategory]string{
		thermal.ActivityCoding:        "#60a5fa", // Blue
		thermal.ActivityConversation:  "#9ca3af", // Light Gray
		thermal.ActivityExploration:   "#22d3ee", // Cyan
		thermal.ActivityDelegation:    "#7dd3fc", // Light Blue
		thermal.ActivityFeatureDev:    "#6ee7b7", // Mint Green
		thermal.ActivityTesting:       "#f472b6", // Magenta
		thermal.ActivityBuildDeploy:   "#34d399", // Emerald
		thermal.ActivityBrainstorming: "#c084fc", // Orchid
		thermal.ActivityDebugging:     "#f87171", // Coral Red
	}

	var lines []string
	for _, a := range p.Taxonomy {
		bar := renderHeatBar(a.Cost, maxCost, 4, colorful)
		catName := string(a.Category)
		var catStyled string
		if colorful {
			if col, ok := catColors[a.Category]; ok {
				catStyled = lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(fmt.Sprintf("%-13s", catName))
			} else {
				catStyled = fmt.Sprintf("%-13s", catName)
			}
		} else {
			catStyled = fmt.Sprintf("%-13s", catName)
		}

		costStr := fmtCost(a.Cost)
		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Render(costStr)
		} else {
			costStyled = costStr
		}

		turnsStr := fmt.Sprintf("%3d", a.Turns)
		oneShotStr := a.OneShot
		if oneShotStr == "" {
			oneShotStr = "-"
		}

		line := fmt.Sprintf("%s %s %7s %4s %6s", bar, catStyled, costStyled, turnsStr, oneShotStr)
		lines = append(lines, line)
	}

	return renderFinOpsCard("By Activity", "#facc15", "cost turns 1-shot", lines, width, targetHeight, colorful)
}

// Box 4: By Model (Magenta)
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
			costStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Render(costStr)
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
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render(f1))
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Render(f2))
	} else {
		lines = append(lines, f1, f2)
	}

	return renderFinOpsCard("By Model", "#e879f9", "cost cache calls 1-shot Tok/s", lines, width, targetHeight, colorful)
}

// Box 5: MCP Servers (Purple)
func renderMCPCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var lines []string
	if p.MCP.ServerCalls == 0 && p.MCP.ServersActive == 0 {
		if colorful {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af")).Render("No MCP usage"))
		} else {
			lines = append(lines, "No MCP usage")
		}
	} else {
		lines = append(lines, fmt.Sprintf("Active Servers:  %d", p.MCP.ServersActive))
		lines = append(lines, fmt.Sprintf("Tool Calls:      %d", p.MCP.ServerCalls))
		lines = append(lines, fmt.Sprintf("Payload Tokens:  %s", thermal.CompactNumber(p.MCP.OverheadTokens)))
	}

	return renderFinOpsCard("MCP Servers", "#c084fc", "", lines, width, targetHeight, colorful)
}

// Box 6: Core Tools (Cyan)
func renderCoreToolsCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxCalls int
	for _, ct := range p.CoreTools {
		if ct.Calls > maxCalls {
			maxCalls = ct.Calls
		}
	}
	if maxCalls <= 0 {
		maxCalls = 1
	}

	var lines []string
	for _, ct := range p.CoreTools {
		bar := renderHeatBar(float64(ct.Calls), float64(maxCalls), 4, colorful)
		name := ct.Name
		if len([]rune(name)) > 24 {
			name = string([]rune(name)[:24])
		}
		callsStr := fmt.Sprintf("%4d", ct.Calls)
		line := fmt.Sprintf("%s %-24s %5s", bar, name, callsStr)
		lines = append(lines, line)
	}

	return renderFinOpsCard("Core Tools", "#38bdf8", "calls", lines, width, targetHeight, colorful)
}

// Box 7: Shell Commands (Orange)
func renderShellCommandsCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxCalls int
	for _, st := range p.SubTools {
		if st.Calls > maxCalls {
			maxCalls = st.Calls
		}
	}
	if maxCalls <= 0 {
		maxCalls = 1
	}

	var lines []string
	for _, st := range p.SubTools {
		bar := renderHeatBar(float64(st.Calls), float64(maxCalls), 4, colorful)
		callsStr := fmt.Sprintf("%4d", st.Calls)
		line := fmt.Sprintf("%s %-16s %5s", bar, st.Name, callsStr)
		lines = append(lines, line)
	}

	return renderFinOpsCard("Shell Commands", "#fb923c", "calls", lines, width, targetHeight, colorful)
}

// Box 8: Skills & Agents (Lavender)
func renderSkillsAgentsCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	var maxUses int
	for _, sa := range p.SkillsAgents {
		if sa.Uses > maxUses {
			maxUses = sa.Uses
		}
	}
	if maxUses <= 0 {
		maxUses = 1
	}

	var lines []string
	for _, sa := range p.SkillsAgents {
		bar := renderHeatBar(float64(sa.Uses), float64(maxUses), 4, colorful)
		costStr := fmtCost(sa.Cost)
		var costStyled string
		if colorful {
			costStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Render(costStr)
		} else {
			costStyled = costStr
		}
		line := fmt.Sprintf("%s %-20s %2d %7s", bar, sa.Name, sa.Uses, costStyled)
		lines = append(lines, line)
	}

	return renderFinOpsCard("Skills & Agents", "#a78bfa", "uses  cost", lines, width, targetHeight, colorful)
}

// Box 9: Workflow (Violet)
func renderWorkflowCard(p thermal.FinOpsGridPayload, width int, targetHeight int, colorful bool) string {
	lines := []string{
		fmt.Sprintf("Corrections  %s", p.Workflow.Corrections),
		fmt.Sprintf("First edit   %s", p.Workflow.FirstEdit),
		fmt.Sprintf("Rework       %s", p.Workflow.Rework),
		fmt.Sprintf("Coverage     %s", p.Workflow.Coverage),
	}

	return renderFinOpsCard("Workflow", "#c084fc", "", lines, width, targetHeight, colorful)
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
