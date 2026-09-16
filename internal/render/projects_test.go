package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sampleProjectReport() thermal.ProjectReport {
	return thermal.ProjectReport{
		Type: "projects",
		Rows: []thermal.ProjectRow{
			{
				Project:    "/home/user/projects/alpha",
				Tools:      []string{"Claude", "OpenCode"},
				Tokens:     1_500_000,
				Cost:       12.5,
				ActiveDays: 4,
				LastDay:    "2026-09-16",
			},
			{
				Project:    "/home/user/projects/beta",
				Tools:      []string{"Devin"},
				Tokens:     900,
				ActiveDays: 1,
				LastDay:    "2026-09-10",
			},
		},
		Totals: thermal.ProjectRow{
			Tokens:         1_500_900,
			Cost:           12.5,
			ActiveDays:     5,
			LastDay:        "2026-09-16",
			EstimatedCost:  2.5,
			MissingPricing: []string{"mystery-model"},
		},
	}
}

func TestRenderProjects_Table(t *testing.T) {
	out := RenderProjects(sampleProjectReport(), true)
	for _, want := range []string{
		"projects", "Project", "Tools", "Tokens", "/home/user/projects/alpha",
		"Claude,OpenCode", "Total", "$12.50", "2026-09-16",
		"estimated from pricing", "No pricing for: mystery-model",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestRenderProjects_Empty(t *testing.T) {
	out := RenderProjects(thermal.ProjectReport{Type: "projects"}, true)
	if !strings.Contains(out, "No project activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}
}

func TestRenderProjects_CapsRows(t *testing.T) {
	rep := thermal.ProjectReport{Type: "projects"}
	for i := 0; i < defaultTop+3; i++ {
		rep.Rows = append(rep.Rows, thermal.ProjectRow{
			Project: fmt.Sprintf("/repo/p%02d", i),
			Tools:   []string{"OpenCode"},
			Tokens:  int64(1000 - i),
			LastDay: "2026-09-16",
		})
	}
	out := RenderProjects(rep, true)
	if !strings.Contains(out, "and 3 more") {
		t.Errorf("expected the overflow line:\n%s", out)
	}
	if strings.Contains(out, "/repo/p22") {
		t.Errorf("rows past the cap should not print:\n%s", out)
	}
}

func TestToolsCell(t *testing.T) {
	all := []string{"Claude", "Codex", "Devin", "Grok", "OpenCode"}
	cell := toolsCell(all, 16)
	if runeLen(cell) > 16 {
		t.Errorf("toolsCell width = %d, want <= 16: %q", runeLen(cell), cell)
	}
	if !strings.Contains(cell, "+") {
		t.Errorf("expected a +N suffix when truncated, got %q", cell)
	}
	if got := toolsCell(nil, 16); got != "—" {
		t.Errorf("toolsCell(nil) = %q, want an em dash", got)
	}
	if got := toolsCell([]string{"Devin"}, 16); got != "Devin" {
		t.Errorf("toolsCell(single) = %q", got)
	}
}

func TestProjectLabel(t *testing.T) {
	got := projectLabel("/home/user/very/deep/path/to/project-name", 20)
	if runeLen(got) > 20 {
		t.Errorf("projectLabel width = %d, want <= 20: %q", runeLen(got), got)
	}
	if !strings.HasSuffix(got, "project-name") {
		t.Errorf("projectLabel = %q, want the tail preserved", got)
	}
}
