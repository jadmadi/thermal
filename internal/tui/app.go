// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbletea/v2"

	"github.com/jadmadi/thermal/internal/thermal"
)

// Tab is one dashboard view. Only Overview is implemented; the remaining four
// are named so the shell, the key map, and the help screen already agree on
// the shape the later goals fill in.
type Tab struct {
	Name string
	Key  string
}

// Tabs is the tab bar order, left to right.
var Tabs = []Tab{
	{Name: "Overview", Key: "1"},
	{Name: "Projects", Key: "2"},
	{Name: "Mix", Key: "3"},
	{Name: "Models", Key: "4"},
	{Name: "Stats", Key: "5"},
}

// Model is the dashboard. It owns the loaded adapter, the current tab, and the
// three controls every view reads: range, metric, and sort.
type Model struct {
	adapter Adapter
	palette Palette

	tab    int
	rng    Range
	metric Metric
	sort   SortKey

	width  int
	height int
	showHe bool
	status string

	// overview is cached per (range, metric, sort) so a keystroke that does
	// not change any of the three re-renders without re-aggregating.
	cached      Overview
	cacheKey    string
	cacheFilled bool

	// Mix and Models view state.
	mixBy       string // "tool" or "model"
	mixSel      int
	statsLog    bool // histogram scale override
	statsDense  bool // 9-box dense FinOps grid toggle
	statsOffset int  // vertical scroll offset for dense grid
	modelsSort  ModelSort
	modelsSel   int
	modelsTools []ToolShare

	// Projects view state.
	projSort   ProjectSort
	projFilter string
	projSel    int
	projOffset int
	drill      string // non-empty means the project detail is open
	detail     ProjectDetail
}

// New builds a dashboard model over already-loaded data.
func New(adapter Adapter, colorful bool) Model {
	return NewWithTab(adapter, colorful, tabOverview)
}

// NewWithTab builds a dashboard model over already-loaded data, starting on the specified tab.
func NewWithTab(adapter Adapter, colorful bool, initialTab int) Model {
	if initialTab < 0 || initialTab >= len(Tabs) {
		initialTab = 0
	}
	return Model{
		adapter:    adapter,
		palette:    newPalette(colorful),
		tab:        initialTab,
		rng:        Range30d,
		metric:     MetricTokens,
		sort:       SortTokens,
		projSort:   ProjectSortTokens,
		mixBy:      "tool",
		modelsSort: ModelSortTokens,
		statsDense: initialTab == tabStats,
		width:      80,
		height:     24,
	}
}

// Init satisfies tea.Model. There is no startup command: loading happened
// before the program started so the first frame is already complete.
func (m Model) Init() tea.Cmd { return nil }

func (m Model) overview() Overview {
	key := fmt.Sprintf("%s|%s|%s", m.rng, m.metric, m.sort)
	if !m.cacheFilled || m.cacheKey != key {
		m.cached = m.adapter.BuildOverview(m.rng, m.metric, m.sort)
		m.cacheKey = key
		m.cacheFilled = true
	}
	return m.cached
}

// Update handles one message. Every branch either changes state or sets a
// status line, so a key never silently does nothing.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	keys := defaultKeys()

	// Views that own keys consume them before the shell does.
	if !m.showHe {
		switch m.tab {
		case tabProjects:
			if handled, model, cmd := m.updateProjects(pressed); handled {
				return model, cmd
			}
		case tabMix:
			if handled, model, cmd := (&m).updateMix(pressed); handled {
				return model, cmd
			}
		case tabModels:
			if handled, model, cmd := (&m).updateModels(pressed); handled {
				return model, cmd
			}
		case tabStats:
			if handled, model, cmd := (&m).updateStats(pressed); handled {
				return model, cmd
			}
		}
	}

	switch {
	case matches(keys.Quit, pressed):
		return m, tea.Quit

	case matches(keys.Help, pressed):
		m.showHe = !m.showHe
		return m, nil

	case pressed >= "1" && pressed <= "5":
		idx := int(pressed[0] - '1')
		if idx < len(Tabs) {
			m.tab = idx
			m.status = ""
		}
		return m, nil

	case matches(keys.NextTab, pressed):
		m.tab = (m.tab + 1) % len(Tabs)
		return m, nil

	case matches(keys.PrevTab, pressed):
		m.tab = (m.tab - 1 + len(Tabs)) % len(Tabs)
		return m, nil

	case matches(keys.Range, pressed):
		m.rng = m.rng.next()
		m.status = "range " + string(m.rng)
		m.projOffset = 0
		return m, nil

	case matches(keys.Metric, pressed):
		m.metric = m.metric.next()
		m.status = "metric " + string(m.metric)
		return m, nil

	case matches(keys.Sort, pressed):
		m.sort = m.sort.next()
		m.status = "sort " + string(m.sort)
		return m, nil

	case matches(keys.Filter, pressed):
		m.status = "filter arrives with the Projects view"
		return m, nil

	case matches(keys.Drill, pressed):
		m.status = "drill-down arrives with the Projects view"
		return m, nil

	case matches(keys.Back, pressed):
		if m.showHe {
			m.showHe = false
			return m, nil
		}
		m.status = ""
		return m, nil
	}
	return m, nil
}

// View renders the frame: tab bar, status line, then the active view.
func (m Model) View() tea.View {
	var b strings.Builder
	b.WriteString("  " + m.tabBar() + "\n")
	if m.status != "" {
		b.WriteString("  " + m.palette.Muted.Render(m.status) + "\n")
	}
	b.WriteString("\n")

	switch {
	case m.showHe:
		b.WriteString(indent(m.helpScreen(), "  "))
	case m.tab == tabProjects && m.drill != "":
		b.WriteString(indent(renderProjectDetail(m.detail, m.innerWidth(), m.palette), "  "))
	case m.tab == tabProjects:
		b.WriteString(indent(renderProjects(m.adapter.BuildProjects(m.rng, m.projSort, m.projFilter), m.innerWidth(), m.height, m.projSel, m.palette), "  "))
	case m.tab == tabMix:
		b.WriteString(indent(renderMix(m.buildMixView(), m.innerWidth(), m.height, m.palette), "  "))
	case m.tab == tabModels:
		b.WriteString(indent(renderModels(m.buildModelsView(), m.innerWidth(), m.height, m.modelsSel, m.palette), "  "))
	case m.tab == tabStats:
		if m.statsDense {
			gridContent := RenderDenseFinOps(m.adapter.BuildFinOpsGrid(), m.innerWidth(), m.palette.Colorful)
			lines := strings.Split(gridContent, "\n")
			avail := m.height - 5
			if avail < 1 {
				avail = 1
			}
			offset := m.statsOffset
			maxOffset := len(lines) - avail
			if maxOffset < 0 {
				maxOffset = 0
			}
			if offset > maxOffset {
				offset = maxOffset
			}
			if offset < 0 {
				offset = 0
			}
			end := offset + avail
			if end > len(lines) {
				end = len(lines)
			}
			b.WriteString(indent(strings.Join(lines[offset:end], "\n"), "  "))
		} else {
			sv := m.adapter.BuildStats(m.rng, m.metric)
			if m.statsLog {
				sv.UseLog = true
				if sv.LogReason == "" {
					sv.LogReason = "log scale: set by hand"
				}
			}
			b.WriteString(indent(renderStats(sv, m.innerWidth(), m.height, m.palette), "  "))
		}
	case m.tab == 0:
		b.WriteString(indent(renderOverview(m.overview(), m.innerWidth(), m.height, m.palette), "  "))
	default:
		b.WriteString(indent(m.placeholder(), "  "))
	}

	b.WriteString("\n\n")
	b.WriteString("  " + m.palette.Dim.Render(fmt.Sprintf("%s  ·  %s", m.hint(), "? help")))

	// The dashboard owns the whole screen. In v2 alt screen is a property of
	// the view, so it has to be set on every frame.
	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// innerWidth is the width available to a view. The frame indents every line by
// two cells, and the table leaves the same two cells free at the right so a
// full row never touches the edge.
func (m Model) innerWidth() int {
	if m.width <= 8 {
		return m.width
	}
	return m.width - 4
}

func (m Model) tabBar() string {
	var parts []string
	for i, t := range Tabs {
		label := fmt.Sprintf("%s %s", t.Key, t.Name)
		if i == m.tab {
			parts = append(parts, m.palette.Emphasis.Render("["+label+"]"))
			continue
		}
		parts = append(parts, m.palette.Muted.Render(" "+label+" "))
	}
	return strings.Join(parts, " ")
}

func (m Model) hint() string {
	switch m.tab {
	case tabProjects:
		if m.drill != "" {
			return fmt.Sprintf("esc back  ·  r range %s", m.rng)
		}
		filter := m.projFilter
		if filter == "" {
			filter = "all"
		}
		return fmt.Sprintf("r range %s  ·  s sort %s  ·  f tool %s  ·  enter detail",
			m.rng, m.projSort, filter)
	case tabMix:
		return fmt.Sprintf("r range %s  ·  t metric %s  ·  v by %s  ·  s sort %s",
			m.rng, m.metric, m.mixBy, m.sort)
	case tabModels:
		return fmt.Sprintf("r range %s  ·  t metric %s  ·  s sort %s", m.rng, m.metric, m.modelsSort)
	case tabStats:
		scale := "auto"
		if m.statsLog {
			scale = "log"
		}
		if m.statsDense {
			return fmt.Sprintf("r range %s  ·  t metric %s  ·  d standard view", m.rng, m.metric)
		}
		return fmt.Sprintf("r range %s  ·  t metric %s  ·  l scale %s", m.rng, m.metric, scale)
	}
	return fmt.Sprintf("r range %s  ·  t metric %s  ·  s sort %s", m.rng, m.metric, m.sort)
}

// Tab indexes, in the order Tabs declares them.
const (
	tabOverview = 0
	tabProjects = 1
	tabMix      = 2
	tabModels   = 3
	tabStats    = 4
)

// buildMixView assembles the Mix tab for the current range, metric and subject.
func (m Model) buildMixView() MixView {
	if m.mixBy == "model" {
		return m.adapter.BuildModelMix(m.rng, m.metric)
	}
	return m.adapter.BuildMix(m.rng, m.metric)
}

// buildModelsView assembles the Models tab and stashes the tool split so the
// view can print it without a second aggregation pass.
func (m *Model) buildModelsView() ModelsView {
	out := m.adapter.BuildModels(m.rng, m.metric, m.modelsSort)
	out.Tools = m.adapter.toolSharesFor(m.rng)
	return out
}

// updateStats handles the keys the Stats view owns.
func (m *Model) updateStats(pressed string) (bool, tea.Model, tea.Cmd) {
	switch pressed {
	case "l":
		m.statsLog = !m.statsLog
		if m.statsLog {
			m.status = "distribution: log scale"
		} else {
			m.status = "distribution: linear scale"
		}
		return true, m, nil
	case "d":
		m.statsDense = !m.statsDense
		m.statsOffset = 0
		if m.statsDense {
			m.status = "stats: dense 9-box FinOps grid"
		} else {
			m.status = "stats: distribution view"
		}
		return true, m, nil
	case "j", "down":
		if m.statsDense {
			m.statsOffset += 2
			return true, m, nil
		}
	case "k", "up":
		if m.statsDense && m.statsOffset > 0 {
			m.statsOffset -= 2
			if m.statsOffset < 0 {
				m.statsOffset = 0
			}
			return true, m, nil
		}
	case "J":
		if m.statsDense {
			m.statsOffset += 12
			return true, m, nil
		}
	case "K":
		if m.statsDense && m.statsOffset > 0 {
			m.statsOffset -= 12
			if m.statsOffset < 0 {
				m.statsOffset = 0
			}
			return true, m, nil
		}
	case "pgdown", "pagedown", "ctrl+d", " ", "space":
		if m.statsDense {
			step := m.height / 2
			if step < 6 {
				step = 6
			}
			m.statsOffset += step
			return true, m, nil
		}
	case "pgup", "pageup", "ctrl+u":
		if m.statsDense && m.statsOffset > 0 {
			step := m.height / 2
			if step < 6 {
				step = 6
			}
			m.statsOffset -= step
			if m.statsOffset < 0 {
				m.statsOffset = 0
			}
			return true, m, nil
		}
	case "g", "home":
		if m.statsDense {
			m.statsOffset = 0
			return true, m, nil
		}
	case "G", "end":
		if m.statsDense {
			m.statsOffset = 99999
			return true, m, nil
		}
	case "s", "?":
		return false, m, nil
	}
	return false, m, nil
}

// updateMix handles the keys the Mix view owns, with the same address trick as
// updateModels.
func (m *Model) updateMix(pressed string) (bool, tea.Model, tea.Cmd) {
	switch pressed {
	case "v":
		if m.mixBy == "model" {
			m.mixBy = "tool"
		} else {
			m.mixBy = "model"
		}
		m.status = "mix by " + m.mixBy
		return true, m, nil
	case "j", "down":
		m.mixSel++
		return true, m, nil
	case "k", "up":
		if m.mixSel > 0 {
			m.mixSel--
		}
		return true, m, nil
	case "g":
		m.mixSel = 0
		return true, m, nil
	case "s", "?":
		return false, m, nil
	}
	return false, m, nil
}

// updateModels handles the keys the Models view owns. The body takes the
// address of the value it was given, so the mutation lands on the copy Update
// returns rather than on the caller's stack value.
func (m *Model) updateModels(pressed string) (bool, tea.Model, tea.Cmd) {
	rows := len(m.adapter.BuildModels(m.rng, m.metric, m.modelsSort).Rows)
	switch pressed {
	case "j", "down":
		if rows > 0 {
			m.modelsSel = (m.modelsSel + 1) % rows
		}
		return true, m, nil
	case "k", "up":
		if rows > 0 {
			m.modelsSel = (m.modelsSel - 1 + rows) % rows
		}
		return true, m, nil
	case "g":
		m.modelsSel = 0
		return true, m, nil
	case "G":
		if rows > 0 {
			m.modelsSel = rows - 1
		}
		return true, m, nil
	case "s":
		m.modelsSort = m.modelsSort.next()
		m.modelsSel = 0
		m.status = "sort " + string(m.modelsSort)
		return true, m, nil
	case "?":
		return false, m, nil
	}
	return false, m, nil
}

// updateProjects handles the keys the Projects view owns. It reports whether it
// consumed the key so the shell can fall through to its own bindings.
func (m Model) updateProjects(pressed string) (bool, tea.Model, tea.Cmd) {
	if m.drill != "" {
		if pressed == "esc" || pressed == "q" {
			m.drill = ""
			m.status = ""
			return true, m, nil
		}
		return false, m, nil
	}

	data := m.adapter.BuildProjects(m.rng, m.projSort, m.projFilter)
	rows := len(data.Rows)

	switch pressed {
	case "j", "down":
		if rows > 0 {
			m.projSel = (m.projSel + 1) % rows
		}
		return true, m, nil
	case "k", "up":
		if rows > 0 {
			m.projSel = (m.projSel - 1 + rows) % rows
		}
		return true, m, nil
	case "g":
		m.projSel = 0
		return true, m, nil
	case "G":
		if rows > 0 {
			m.projSel = rows - 1
		}
		return true, m, nil
	case "s":
		m.projSort = m.projSort.next()
		m.projSel = 0
		m.status = "sort " + string(m.projSort)
		return true, m, nil
	case "f":
		m.projFilter = m.adapter.nextToolFilter(m.rng, m.projFilter)
		m.projSel = 0
		if m.projFilter == "" {
			m.status = "tool filter cleared"
		} else {
			m.status = "tool filter " + m.projFilter
		}
		return true, m, nil
	case "enter":
		if rows == 0 || m.projSel >= rows {
			return true, m, nil
		}
		name := data.Rows[m.projSel].Project
		m.drill = name
		m.detail = m.adapter.BuildProjectDetail(name, m.rng)
		m.status = ""
		return true, m, nil
	case "?":
		return false, m, nil
	}
	return false, m, nil
}

// nextToolFilter cycles through no filter, then each tool that has project rows
// in the window, then back to no filter.
func (a Adapter) nextToolFilter(r Range, current string) string {
	options := a.toolFilterOptions(r)
	if len(options) == 0 {
		return ""
	}
	if current == "" {
		return options[0]
	}
	for i, name := range options {
		if name == current {
			if i+1 < len(options) {
				return options[i+1]
			}
			return ""
		}
	}
	return ""
}

func (m Model) placeholder() string {
	name := Tabs[m.tab].Name
	return m.palette.Emphasis.Render(name) + "\n\n" +
		m.palette.Muted.Render("This view is not built yet. The shell, range, and metric controls") + "\n" +
		m.palette.Muted.Render("are wired, and the view lands in a later goal.") + "\n\n" +
		m.palette.Muted.Render("Press 1 for the Overview.")
}

func (m Model) helpScreen() string {
	var b strings.Builder
	b.WriteString(m.palette.Emphasis.Render("Keys"))
	b.WriteString("\n\n")
	for _, row := range defaultKeys().helpRows() {
		b.WriteString(padTo(row[0], 18) + m.palette.Muted.Render(row[1]) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l == "" {
			continue
		}
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

// LoadNote explains an empty dashboard. It is printed by the command rather
// than the model so a headless run reports it too.
func LoadNote(a Adapter) string {
	if a.Latest == "" {
		return "no supported tool data found"
	}
	if len(a.Skipped) > 0 {
		return "loaded with skipped tools: " + strings.Join(a.Skipped, ", ")
	}
	return ""
}

// Pricer is re-exported so callers do not need the thermal package to build a
// dashboard.
type Pricer = thermal.Pricer
