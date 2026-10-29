// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/jadmadi/thermal/internal/thermal"
)

func maxLineWidth(text string) int {
	max := 0
	for _, line := range strings.Split(text, "\n") {
		if w := ansi.StringWidth(line); w > max {
			max = w
		}
	}
	return max
}

func assertFitsWidth(t *testing.T, label, text string, width int) {
	t.Helper()
	if got := maxLineWidth(text); got > width {
		t.Fatalf("%s at %d columns has a %d-column line:\n%s", label, width, got, text)
	}
}

func TestContentWidthBounds(t *testing.T) {
	t.Setenv("COLUMNS", "30")
	if got := ContentWidth(); got != 40 {
		t.Fatalf("ContentWidth(30)=%d, want 40", got)
	}
	t.Setenv("COLUMNS", "60")
	if got := ContentWidth(); got != 60 {
		t.Fatalf("ContentWidth(60)=%d, want 60", got)
	}
	t.Setenv("COLUMNS", "160")
	if got := ContentWidth(); got != 96 {
		t.Fatalf("ContentWidth(160)=%d, want 96", got)
	}
}

func TestWidthMatrix(t *testing.T) {
	report := sampleReport()
	report.Rows[0].Reasoning = 1200
	if report.Rows[0].Models == nil {
		report.Rows[0].Models = map[string]thermal.ModelTokens{}
	}
	report.Rows[0].Models["a-very-long-model-name-that-needs-truncation"] = thermal.ModelTokens{Input: 1000, Output: 500}
	report.Totals.Cache = 500
	report.Totals.Input = 1200
	report.Totals.UnattributedTokens = 25
	report.Totals.MissingPricing = []string{"a-very-long-unpriced-model-name"}

	projects := sampleProjectReport()
	models := sampleModelReport()

	for _, width := range []int{60, 80, 100, 120, 160} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			cases := map[string]string{
				"report":             RenderReportAt(report, true, width),
				"report-breakdown":   RenderReportBreakdownAt(report, true, width),
				"projects":           RenderProjectsAt(projects, 0, true, width),
				"projects-breakdown": RenderProjectsBreakdownAt(projects, 0, true, width),
				"models":             RenderModelsAt(models, 0, true, width),
				"models-breakdown":   RenderModelsBreakdownAt(models, 0, true, width),
			}
			for label, out := range cases {
				assertFitsWidth(t, label, out, width)
				if strings.Contains(out, "\x1b[") {
					t.Fatalf("%s contains ANSI with noColor: %q", label, out)
				}
			}

			reportOut := cases["report"]
			for _, header := range []string{"Period", "Tokens", "Cost"} {
				if !strings.Contains(reportOut, header) {
					t.Fatalf("report at %d missing %q:\n%s", width, header, reportOut)
				}
			}
			if !strings.Contains(cases["projects"], "Project") || !strings.Contains(cases["models"], "Model") {
				t.Fatalf("project/model header missing at %d", width)
			}
		})
	}
}

func TestHelpAndMixWidthMatrix(t *testing.T) {
	for _, width := range []int{60, 80, 100, 120, 160} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			t.Setenv("COLUMNS", strconv.Itoa(width))
			assertFitsWidth(t, "help", RenderHelp(true), width)
			assertFitsWidth(t, "mix", RenderMix(sampleMixReport(), true), width)
		})
	}
}

func TestColumnDropOrder(t *testing.T) {
	report := sampleReport()
	report.Rows[0].Reasoning = 100
	projects := sampleProjectReport()
	models := sampleModelReport()

	report60 := RenderReportAt(report, true, 60)
	header60 := ""
	for _, line := range strings.Split(report60, "\n") {
		if strings.Contains(line, "Period") && strings.Contains(line, "Tokens") {
			header60 = line
			break
		}
	}
	if header60 == "" {
		t.Fatalf("daily at 60 has no table header:\n%s", report60)
	}
	for _, absent := range []string{"Input", "Output", "Reason", "Cache"} {
		if strings.Contains(header60, absent) {
			t.Errorf("daily at 60 should drop %q from the table:\n%s", absent, header60)
		}
	}
	for _, present := range []string{"Period", "Tokens", "Cost"} {
		if !strings.Contains(header60, present) {
			t.Errorf("daily at 60 missing %q:\n%s", present, header60)
		}
	}
	report120 := RenderReportAt(report, true, 120)
	for _, present := range []string{"Input", "Output", "Cache"} {
		if !strings.Contains(report120, present) {
			t.Errorf("daily at 120 should include %q:\n%s", present, report120)
		}
	}

	proj60 := RenderProjectsAt(projects, 0, true, 60)
	projHeader := ""
	for _, line := range strings.Split(proj60, "\n") {
		if strings.Contains(line, "Project") && strings.Contains(line, "Tokens") {
			projHeader = line
			break
		}
	}
	for _, absent := range []string{"Tools", "P-Days", "Last"} {
		if strings.Contains(projHeader, absent) {
			t.Errorf("projects at 60 should drop %q from the table:\n%s", absent, projHeader)
		}
	}
	model60 := RenderModelsAt(models, 0, true, 60)
	modelHeader := ""
	for _, line := range strings.Split(model60, "\n") {
		if strings.Contains(line, "Model") && strings.Contains(line, "Tokens") {
			modelHeader = line
			break
		}
	}
	for _, absent := range []string{"Tools", "M-Days", "Last"} {
		if strings.Contains(modelHeader, absent) {
			t.Errorf("models at 60 should drop %q from the table:\n%s", absent, modelHeader)
		}
	}
}

func TestLongNameTruncation(t *testing.T) {
	longName := strings.Repeat("very-long-project-name-", 12)
	rep := thermal.ProjectReport{
		Type:   "projects",
		Rows:   []thermal.ProjectRow{{Project: longName, Tokens: 1234567, Cost: 1.25}},
		Totals: thermal.ProjectRow{Tokens: 1234567, Cost: 1.25},
	}
	out := RenderProjectsAt(rep, 0, true, 60)
	if !strings.Contains(out, "…") {
		t.Fatalf("long project name was not truncated with an ellipsis:\n%s", out)
	}
	if !strings.Contains(out, "╰") {
		t.Fatalf("long project name broke the bottom border:\n%s", out)
	}
	assertFitsWidth(t, "long-name projects", out, 60)
}

func TestHeatmapAutoFit(t *testing.T) {
	t.Setenv("COLUMNS", "50")
	activity := map[string]thermal.DayActivity{thermal.LocalDay(time.Now()): {Tokens: 1000}}
	lines := RenderHeatmap(activity, 52, false)
	out := strings.Join(lines, "\n")
	if !strings.Contains(out, "auto-fit to 50 columns") {
		t.Fatalf("heatmap missing auto-fit caption:\n%s", out)
	}
	if got := maxLineWidth(out); got > 50 {
		t.Fatalf("heatmap line is %d columns at 50:\n%s", got, out)
	}
	t.Setenv("COLUMNS", "200")
	if strings.Contains(strings.Join(RenderHeatmap(activity, 10, false), "\n"), "auto-fit") {
		t.Fatal("heatmap should not auto-fit when the requested window fits")
	}
}
