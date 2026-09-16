package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func sampleModelReport() thermal.ModelReport {
	return thermal.ModelReport{
		Type: "models",
		Rows: []thermal.ModelRow{
			{
				Model:    "accounts/fireworks/models/deepseek-v4.1-flash",
				Tools:    []string{"OpenCode", "MiMoCode"},
				Tokens:   5_000_000,
				Cost:     12.5,
				Days:     12,
				LastDay:  "2026-09-17",
				FirstDay: "2026-09-01",
			},
			{
				Model:   "gpt-6-astra",
				Tools:   []string{"Codex"},
				Tokens:  900_000,
				Days:    3,
				LastDay: "2026-09-16",
			},
		},
		Totals: thermal.ModelRow{
			Tokens:         5_900_000,
			Cost:           12.5,
			Days:           15,
			LastDay:        "2026-09-17",
			Tools:          []string{"Codex", "MiMoCode", "OpenCode"},
			MissingPricing: []string{"codex-auto-review"},
		},
	}
}

func TestRenderModels_Table(t *testing.T) {
	out := RenderModels(sampleModelReport(), 0, true)
	for _, want := range []string{
		"models", "Model", "Tools", "Tokens", "deepseek-v4.1-flash",
		"OpenCode,MiMoCode", "$12.50", "estimated from models.dev",
		"No pricing for: codex-auto-review",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestRenderModels_EmptyAndLimit(t *testing.T) {
	if out := RenderModels(thermal.ModelReport{Type: "models"}, 0, true); !strings.Contains(out, "No model activity") {
		t.Errorf("expected empty notice:\n%s", out)
	}

	rep := sampleModelReport()
	rep.Rows = append(rep.Rows, thermal.ModelRow{Model: "muse-spark", Tools: []string{"OpenCode"}, Tokens: 10, Days: 1})
	out := RenderModels(rep, 2, true)
	if !strings.Contains(out, "and 1 more") {
		t.Errorf("expected the overflow line:\n%s", out)
	}
	if strings.Contains(out, "muse-spark") {
		t.Errorf("rows past the limit should not print:\n%s", out)
	}
}

func TestRenderProjectsBreakdown(t *testing.T) {
	rep := thermal.ProjectReport{
		Type: "projects",
		Rows: []thermal.ProjectRow{
			{
				Project:    "/home/user/projects/alpha",
				Tools:      []string{"Claude", "OpenCode"},
				ToolTokens: map[string]int64{"Claude": 750_000, "OpenCode": 250_000},
				Tokens:     1_000_000,
				Models: map[string]thermal.ModelTokens{
					"claude-opus":    {Input: 750_000},
					"deepseek-flash": {Input: 250_000},
				},
				ActiveDays: 2,
				LastDay:    "2026-09-16",
			},
		},
	}
	out := RenderProjectsBreakdown(rep, 0, true)
	for _, want := range []string{"alpha", "Claude 750.0K", "OpenCode 250.0K", "claude-opus 750.0K", "deepseek-flash 250.0K"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in breakdown:\n%s", want, out)
		}
	}

	// Without the flag the detail stays out.
	plain := RenderProjects(rep, 0, true)
	if strings.Contains(plain, "tools") || strings.Contains(plain, "750.0K") {
		t.Errorf("plain table should not carry the breakdown:\n%s", plain)
	}
}
