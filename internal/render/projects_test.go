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
	out := RenderProjects(sampleProjectReport(), 0, true)
	for _, want := range []string{
		"projects", "Project", "Tools", "Tokens", "alpha", "beta",
		"Claude,OpenCode", "Total", "$12.50", "2026-09-16",
		"estimated from pricing", "No pricing for: mystery-model",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestRenderProjects_Empty(t *testing.T) {
	out := RenderProjects(thermal.ProjectReport{Type: "projects"}, 0, true)
	if !strings.Contains(out, "No project activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}
}

func TestRenderProjects_CapsRows(t *testing.T) {
	rep := thermal.ProjectReport{Type: "projects"}
	for i := 0; i < 12; i++ {
		rep.Rows = append(rep.Rows, thermal.ProjectRow{
			Project: fmt.Sprintf("/repo/p%02d", i),
			Tools:   []string{"OpenCode"},
			Tokens:  int64(1000 - i),
			LastDay: "2026-09-16",
		})
	}

	// No limit prints every row.
	all := RenderProjects(rep, 0, true)
	if strings.Contains(all, "more") {
		t.Errorf("unlimited output should not summarise rows:\n%s", all)
	}
	if !strings.Contains(all, "p11") {
		t.Errorf("expected the last row to print:\n%s", all)
	}

	// A limit prints the head and summarises the rest.
	capped := RenderProjects(rep, 5, true)
	if !strings.Contains(capped, "and 7 more") {
		t.Errorf("expected the overflow line:\n%s", capped)
	}
	if strings.Contains(capped, "p05") {
		t.Errorf("rows past the limit should not print:\n%s", capped)
	}
}

func TestDisplayNames(t *testing.T) {
	rows := []thermal.ProjectRow{
		{Project: "/home/user/projects/waqftech/mahak-bench"},
		{Project: "/mnt/Jad/github/lab/sila"},
		{Project: "/home/user/work/other/mahak-bench"},
		{Project: "/srv/app"},
	}
	names := displayNames(rows)

	if names["/home/user/projects/waqftech/mahak-bench"] != "mahak-bench (waqftech)" {
		t.Errorf("colliding name should carry a parent hint, got %q",
			names["/home/user/projects/waqftech/mahak-bench"])
	}
	if names["/home/user/work/other/mahak-bench"] != "mahak-bench (other)" {
		t.Errorf("colliding name should carry a parent hint, got %q",
			names["/home/user/work/other/mahak-bench"])
	}
	if names["/srv/app"] != "app" {
		t.Errorf("unique name = %q, want app", names["/srv/app"])
	}
	if names["/mnt/Jad/github/lab/sila"] != "sila" {
		t.Errorf("unique name = %q, want sila", names["/mnt/Jad/github/lab/sila"])
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

func TestProjectLabelFallsBackToTail(t *testing.T) {
	names := displayNames([]thermal.ProjectRow{{Project: "/home/user/very/deep/path/to/project-name"}})
	got := names["/home/user/very/deep/path/to/project-name"]
	if got != "project-name" {
		t.Errorf("displayNames = %q, want the last segment", got)
	}
}
