// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbletea/v2"
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
	pal := newPalette(colorful)

	if width <= 0 {
		width = 80
	}

	header := pal.Emphasis.Render("Thermal") + " " + pal.Dim.Render("·") + " " + pal.Emphasis.Render("FinOps 9-Box Grid") +
		" " + pal.Muted.Render(fmt.Sprintf("(tokens & spend allocation, %d active days)", p.ActiveDays))

	// Determine column layout based on terminal width:
	// >= 140 columns: 3 columns x 3 rows (full 9-box grid)
	// 90..139 columns: 2 columns (4 rows of 2 + 1 bottom box)
	// < 90 columns: 1 column (vertical stack of cards)
	var cols int
	var boxWidth int
	switch {
	case width >= 140:
		cols = 3
		boxWidth = (width - 4) / 3
		if boxWidth > 48 {
			boxWidth = 48
		}
	case width >= 90:
		cols = 2
		boxWidth = (width - 2) / 2
		if boxWidth > 58 {
			boxWidth = 58
		}
	default:
		cols = 1
		boxWidth = width - 2
		if boxWidth < 30 {
			boxWidth = 30
		}
	}

	// 1. Box 1: Executive KPI Summary
	box1 := renderCard("Executive KPIs", []string{
		fmt.Sprintf("Volume: %s tokens", thermal.CompactNumber(p.TotalTokens)),
		fmt.Sprintf("Spend:  $%.2f (avg $%.2f/d)", p.TotalCost, p.AvgDailyCost),
		fmt.Sprintf("Active: %d days (%s avg/d)", p.ActiveDays, thermal.CompactNumber(p.AvgDailyTokens)),
		fmt.Sprintf("Rating: [%s]", p.SpendEfficiency),
	}, boxWidth, pal)

	// 2. Box 2: Activity Taxonomy
	var taxLines []string
	for _, t := range p.Taxonomy {
		barLen := int(t.Percent / 12.5)
		if barLen > 8 {
			barLen = 8
		}
		bar := strings.Repeat("■", barLen) + strings.Repeat("·", 8-barLen)
		taxLines = append(taxLines, fmt.Sprintf("%-11s %4.1f%% [%s]", t.Category, t.Percent, bar))
	}
	for len(taxLines) < 4 {
		taxLines = append(taxLines, "")
	}
	box2 := renderCard("Activity Taxonomy", taxLines, boxWidth, pal)

	// 3. Box 3: Cache Performance & Savings
	box3 := renderCard("Cache & FinOps Savings", []string{
		fmt.Sprintf("Cache Read:  %s tok", thermal.CompactNumber(p.Cache.CacheReadTokens)),
		fmt.Sprintf("Hit Rate:    %.1f%%", p.Cache.HitRate),
		fmt.Sprintf("Net Savings: ~$%.2f", p.Cache.EstimatedSavings),
		fmt.Sprintf("Input Vol:   %s tok", thermal.CompactNumber(p.Cache.InputTokens)),
	}, boxWidth, pal)

	// 4. Box 4: Top Models Matrix
	var modelLines []string
	for i, m := range p.TopModels {
		if i >= 4 {
			break
		}
		modelLines = append(modelLines, fmt.Sprintf("%d. %-18s %s", i+1, truncateRunes(m.Name, 18), thermal.CompactNumber(m.Tokens)))
	}
	for len(modelLines) < 4 {
		modelLines = append(modelLines, "-")
	}
	box4 := renderCard("Top Models Spend", modelLines, boxWidth, pal)

	// 5. Box 5: Sub-Tool Decomposition
	var toolLines []string
	for i, st := range p.SubTools {
		if i >= 4 {
			break
		}
		toolLines = append(toolLines, fmt.Sprintf("%d. %-10s %3d calls (%4.1f%%)", i+1, st.Name, st.Calls, st.Share))
	}
	for len(toolLines) < 4 {
		toolLines = append(toolLines, "-")
	}
	box5 := renderCard("Sub-Tool Shell Calls", toolLines, boxWidth, pal)

	// 6. Box 6: MCP Overhead Telemetry
	box6 := renderCard("MCP Protocol Overhead", []string{
		fmt.Sprintf("Active Servers:  %d", p.MCP.ServersActive),
		fmt.Sprintf("Tool Calls:      %d", p.MCP.ServerCalls),
		fmt.Sprintf("Payload Tokens:  %s", thermal.CompactNumber(p.MCP.OverheadTokens)),
		fmt.Sprintf("Protocol Tax:    ~4.5%%"),
	}, boxWidth, pal)

	// 7. Box 7: Project Allocation
	var projLines []string
	for i, pr := range p.TopProjects {
		if i >= 4 {
			break
		}
		projLines = append(projLines, fmt.Sprintf("%d. %-18s %s", i+1, truncateRunes(pr.Name, 18), thermal.CompactNumber(pr.Tokens)))
	}
	for len(projLines) < 4 {
		projLines = append(projLines, "-")
	}
	box7 := renderCard("Project Allocation", projLines, boxWidth, pal)

	// 8. Box 8: Engineering Velocity & Yield
	netStr := fmt.Sprintf("+%d", p.Yield.NetLines)
	if p.Yield.NetLines < 0 {
		netStr = fmt.Sprintf("%d", p.Yield.NetLines)
	}
	box8 := renderCard("Velocity & Code Yield", []string{
		fmt.Sprintf("Lines: +%s / -%s", thermal.CompactNumber(p.Yield.LinesAdded), thermal.CompactNumber(p.Yield.LinesDeleted)),
		fmt.Sprintf("Net Shipped: %s lines", netStr),
		fmt.Sprintf("Token Yield: %s tok/ln", thermal.CompactNumber(int64(p.Yield.TokensPerNet))),
		fmt.Sprintf("Efficiency:  [%s]", p.Yield.Efficiency),
	}, boxWidth, pal)

	// 9. Box 9: Month-End Forecast
	box9 := renderCard("Month-End Forecast", []string{
		fmt.Sprintf("Run-Rate:  %s tok/mo", thermal.CompactNumber(p.MonthEndRunRateTokens)),
		fmt.Sprintf("Cost Run:  $%.2f / mo", p.MonthEndRunRateCost),
		fmt.Sprintf("Capacity:  [%s]", p.CapacityVerdict),
		fmt.Sprintf("Horizon:   30 calendar days"),
	}, boxWidth, pal)

	boxes := []string{box1, box2, box3, box4, box5, box6, box7, box8, box9}

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

	footer := pal.Dim.Render("• Navigation: q=quit · m=metric · r=range · 9-box responsive FinOps layout")

	return "\n  " + header + "\n\n" + gridOut + "\n\n  " + footer + "\n"
}

func renderCard(title string, lines []string, width int, pal Palette) string {
	if width < 20 {
		width = 20
	}
	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	titleTrunc := truncateRunes(title, innerWidth-2)
	topRule := innerWidth - len([]rune(titleTrunc)) - 1
	if topRule < 1 {
		topRule = 1
	}

	topBorder := "┌─ " + pal.Emphasis.Render(titleTrunc) + " " + pal.Dim.Render(strings.Repeat("─", topRule)) + "┐"
	bottomBorder := pal.Dim.Render("└" + strings.Repeat("─", width-2) + "┘")

	var b strings.Builder
	b.WriteString(topBorder)
	b.WriteString("\n")

	for _, line := range lines {
		lineTrunc := truncateRunes(line, innerWidth)
		pad := innerWidth - len([]rune(ansi.Strip(lineTrunc)))
		if pad < 0 {
			pad = 0
		}
		b.WriteString(pal.Dim.Render("│ ") + lineTrunc + strings.Repeat(" ", pad) + pal.Dim.Render(" │"))
		b.WriteString("\n")
	}

	b.WriteString(bottomBorder)
	return b.String()
}
