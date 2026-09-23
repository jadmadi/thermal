// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package render

import (
	"strings"
	"testing"

	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderYield(t *testing.T) {
	rep := thermal.YieldReport{
		Type: "yield",
		Tools: []thermal.YieldRow{
			{
				Name:         "OpenCode",
				Type:         "tool",
				Tokens:       100000,
				LinesAdded:   1000,
				LinesDeleted: 200,
				NetLines:     800,
				GrossLines:   1200,
				FilesTouched: 5,
				TokensPerNet: 125.0,
				Efficiency:   "HIGH",
				Status:       "MEASURED",
			},
		},
		Models: []thermal.YieldRow{
			{
				Name:         "claude-3-7-sonnet",
				Type:         "model",
				Tokens:       50000,
				LinesAdded:   600,
				LinesDeleted: 100,
				NetLines:     500,
				GrossLines:   700,
				FilesTouched: 3,
				TokensPerNet: 100.0,
				Efficiency:   "HIGH",
				Status:       "MEASURED",
			},
			{
				Name:         "unmeasured-model",
				Type:         "model",
				Tokens:       10000,
				LinesAdded:   0,
				LinesDeleted: 0,
				NetLines:     0,
				GrossLines:   0,
				FilesTouched: 0,
				TokensPerNet: 0,
				Efficiency:   "EXPLORATORY",
				Status:       "UNMEASURED",
			},
		},
		Totals: thermal.YieldRow{
			Name:         "Totals",
			Type:         "total",
			Tokens:       100000,
			LinesAdded:   1000,
			LinesDeleted: 200,
			NetLines:     800,
			GrossLines:   1200,
			FilesTouched: 5,
			TokensPerNet: 125.0,
			Efficiency:   "HIGH",
			Status:       "MEASURED",
		},
	}

	out := RenderYield(rep, 0, true)
	if !strings.Contains(out, "claude-3-7-sonnet") {
		t.Errorf("expected claude-3-7-sonnet in output, got:\n%s", out)
	}
	if !strings.Contains(out, "OpenCode") {
		t.Errorf("expected OpenCode in output, got:\n%s", out)
	}
	if !strings.Contains(out, "[HIGH]") {
		t.Errorf("expected [HIGH] badge in output, got:\n%s", out)
	}
	if !strings.Contains(out, "unmeasured") {
		t.Errorf("expected unmeasured badge in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Totals:") {
		t.Errorf("expected Totals: in output, got:\n%s", out)
	}

	// Verify no ANSI escape codes when noColor=true
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected zero ANSI escape codes when noColor=true, found some in:\n%s", out)
	}
}

func TestRenderYield_Empty(t *testing.T) {
	rep := thermal.YieldReport{Type: "yield"}
	out := RenderYield(rep, 0, true)
	if !strings.Contains(out, "No token activity") {
		t.Errorf("expected 'No token activity' for empty report, got:\n%s", out)
	}
}
