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
}

// New builds a dashboard model over already-loaded data.
func New(adapter Adapter, colorful bool) Model {
	return Model{
		adapter: adapter,
		palette: newPalette(colorful),
		rng:     Range30d,
		metric:  MetricTokens,
		sort:    SortTokens,
		width:   80,
		height:  24,
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

	if m.showHe {
		b.WriteString(indent(m.helpScreen(), "  "))
	} else if m.tab == 0 {
		b.WriteString(indent(renderOverview(m.overview(), m.innerWidth(), m.palette), "  "))
	} else {
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
	return fmt.Sprintf("r range %s  ·  t metric %s  ·  s sort %s", m.rng, m.metric, m.sort)
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
