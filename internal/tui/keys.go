package tui

import "charm.land/bubbletea/v2"

// keyMap is the dashboard's whole key surface. It stays a struct rather than a
// map so the help view and the update switch cannot drift apart: adding a key
// means adding a field and one case in Update.
type keyMap struct {
	Quit      []string
	Help      []string
	Tabs      []string
	NextTab   []string
	PrevTab   []string
	Down      []string
	Up        []string
	TopBottom []string
	Sort      []string
	Metric    []string
	Range     []string
	Filter    []string
	FilterCyc []string
	Drill     []string
	Back      []string
}

func defaultKeys() keyMap {
	return keyMap{
		Quit:      []string{"q", "ctrl+c"},
		Help:      []string{"?"},
		Tabs:      []string{"1", "2", "3", "4", "5"},
		NextTab:   []string{"tab"},
		PrevTab:   []string{"shift+tab"},
		Down:      []string{"j", "down"},
		Up:        []string{"k", "up"},
		TopBottom: []string{"g", "G"},
		Sort:      []string{"s"},
		Metric:    []string{"t"},
		Range:     []string{"r"},
		Filter:    []string{"/"},
		FilterCyc: []string{"f"},
		Drill:     []string{"enter"},
		Back:      []string{"esc"},
	}
}

// helpRows drives the help overlay. Order is the order a reader needs them:
// leave, then move, then change what you are looking at.
func (k keyMap) helpRows() [][2]string {
	return [][2]string{
		{"q", "quit"},
		{"1-5", "jump to a view"},
		{"tab / shift+tab", "next and previous view"},
		{"j / k", "move down and up"},
		{"g / G", "first and last row"},
		{"s", "cycle sort"},
		{"f", "cycle the tool filter (Projects)"},
		{"t", "toggle tokens and cost"},
		{"r", "cycle range (30d, 90d, 1y, all)"},
		{"esc", "back"},
		{"?", "close this help"},
	}
}

// matches reports whether a key press is one of the strings in a binding.
func matches(binding []string, pressed string) bool {
	for _, b := range binding {
		if b == pressed {
			return true
		}
	}
	return false
}

// isPress is the one place the key message type is named, so the rest of the
// package does not import bubbletea for anything but the Model contract.
func isPress(msg tea.Msg) (string, bool) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		return k.String(), true
	}
	return "", false
}
