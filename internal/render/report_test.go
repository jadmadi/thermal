package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sampleReport() thermal.Report {
	return thermal.Report{
		Type: "weekly",
		Tool: "OpenCode",
		Rows: []thermal.PeriodRow{
			{
				Period:     "2026-09-13",
				Input:      1200,
				Output:     3400,
				Cache:      500,
				Tokens:     5100,
				Turns:      12,
				ActiveDays: 3,
				Cost:       2.5,
				Models: map[string]thermal.ModelTokens{
					"deepseek-v4-flash": {Input: 1200, Output: 3400, CacheRead: 500},
				},
			},
		},
		Totals: thermal.PeriodRow{
			Input:      1200,
			Output:     3400,
			Cache:      500,
			Tokens:     5100,
			Turns:      12,
			ActiveDays: 3,
			Cost:       2.5,
		},
	}
}

func TestRenderReport_Table(t *testing.T) {
	out := RenderReport(sampleReport(), true)
	for _, want := range []string{"Thermal", "weekly usage", "OpenCode", "2026-09-13", "Total", "$2.50", "deepseek-v4-flash"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Reason") {
		t.Errorf("reason column should be hidden when no reasoning tokens exist:\n%s", out)
	}
}

func TestRenderReport_ReasoningColumn(t *testing.T) {
	rep := sampleReport()
	rep.Rows[0].Reasoning = 42
	rep.Rows[0].Tokens += 42
	out := RenderReport(rep, true)
	if !strings.Contains(out, "Reason") {
		t.Errorf("expected reason column:\n%s", out)
	}
}

func TestRenderReport_Empty(t *testing.T) {
	rep := thermal.Report{Type: "daily"}
	out := RenderReport(rep, true)
	if !strings.Contains(out, "No activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}
}

func TestRenderReport_EstimatedAndMissing(t *testing.T) {
	rep := sampleReport()
	rep.Totals.EstimatedCost = 1.25
	rep.Totals.MissingPricing = []string{"mystery-model"}
	out := RenderReport(rep, true)
	if !strings.Contains(out, "estimated from pricing") {
		t.Errorf("expected estimated footnote:\n%s", out)
	}
	if !strings.Contains(out, "No pricing for: mystery-model") {
		t.Errorf("expected missing pricing note:\n%s", out)
	}
}

func TestRenderReportBreakdown(t *testing.T) {
	out := RenderReportBreakdown(sampleReport(), true)
	if !strings.Contains(out, "└─") {
		t.Errorf("expected breakdown row:\n%s", out)
	}
}

func TestModelCell(t *testing.T) {
	models := map[string]thermal.ModelTokens{
		"aaa": {Input: 300},
		"bbb": {Input: 200},
		"ccc": {Input: 100},
	}
	cell := modelCell(models, 8)
	if len(cell) > 8 {
		t.Errorf("modelCell width = %d, want <= 8: %q", len(cell), cell)
	}
	if !strings.Contains(cell, "+") {
		t.Errorf("expected a +N suffix when truncated, got %q", cell)
	}
	if got := modelCell(nil, 8); got != "—" {
		t.Errorf("modelCell(nil) = %q, want em dash", got)
	}
}
