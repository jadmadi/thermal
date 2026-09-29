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

func TestRenderYieldMarkdown(t *testing.T) {
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
	}

	md := RenderYieldMarkdown(rep, 0)
	if !strings.Contains(md, "### ⚡ Token Yield & Code Delta") {
		t.Errorf("missing header in markdown yield:\n%s", md)
	}
	if !strings.Contains(md, "#### Models") || !strings.Contains(md, "#### Tools") {
		t.Errorf("missing sections in markdown yield:\n%s", md)
	}
	if !strings.Contains(md, "| 1 | `claude-3-7-sonnet` | 50.0K | +600 | -100 | +500 | 100 tok/ln | `[HIGH]` |") {
		t.Errorf("missing model row in markdown yield:\n%s", md)
	}
	if !strings.Contains(md, "| 2 | `unmeasured-model` | 10.0K | +0 | -0 | 0 | unmeasured | `[EXPLORATORY]` |") {
		t.Errorf("missing unmeasured model row in markdown yield:\n%s", md)
	}
	if !strings.Contains(md, "<details>") || !strings.Contains(md, "</details>") {
		t.Errorf("missing collapsible details in markdown yield:\n%s", md)
	}

	// Test empty markdown
	emptyMd := RenderYieldMarkdown(thermal.YieldReport{}, 0)
	if !strings.Contains(emptyMd, "*No token activity or code deltas found") {
		t.Errorf("expected empty markdown notice, got:\n%s", emptyMd)
	}
}

func TestRenderYield_WithLineage(t *testing.T) {
	rep := thermal.YieldReport{
		Type: "yield",
		Tools: []thermal.YieldRow{
			{
				Name:           "Codex",
				Type:           "tool",
				Tokens:         100000,
				LinesAdded:     1200,
				LinesDeleted:   200,
				NetLines:       1000,
				GrossLines:     1400,
				FilesTouched:   5,
				TokensPerNet:   100.0,
				Efficiency:     "HIGH",
				Status:         "MEASURED",
				MainlineTokens: 80000,
				ForkTokens:     20000,
			},
		},
		Totals: thermal.YieldRow{
			Name:           "Totals",
			Type:           "total",
			Tokens:         100000,
			LinesAdded:     1200,
			LinesDeleted:   200,
			NetLines:       1000,
			GrossLines:     1400,
			FilesTouched:   5,
			TokensPerNet:   100.0,
			Efficiency:     "HIGH",
			Status:         "MEASURED",
			MainlineTokens: 80000,
			ForkTokens:     20000,
		},
		Lineage: &thermal.YieldBranchSummary{
			MainlineTokens: 80000,
			ForkTokens:     20000,
			RootSessions:   8,
			ForkSessions:   2,
			TotalSessions:  10,
			ForkRate:       20.0,
			MainlineYield:  80.0,
		},
	}

	// 1. Terminal text render
	txt := RenderYield(rep, 0, true)
	if !strings.Contains(txt, "Session Lineage & Branch Churn:") {
		t.Errorf("missing Lineage header in text yield:\n%s", txt)
	}
	if !strings.Contains(txt, "Mainline:    80.0K tok (8 sessions) · 80.0% · yield: 80 tok/ln") {
		t.Errorf("missing Mainline metrics in text yield:\n%s", txt)
	}
	if !strings.Contains(txt, "Fork/Branch: 20.0K tok (2 sessions) · 20.0% exploratory churn") {
		t.Errorf("missing Fork metrics in text yield:\n%s", txt)
	}

	// 2. Markdown render
	md := RenderYieldMarkdown(rep, 0)
	if !strings.Contains(md, "#### 🌿 Session Branching & Exploratory Lineage") {
		t.Errorf("missing lineage section in markdown yield:\n%s", md)
	}
	if !strings.Contains(md, "- **Mainline Work**: `80.0K tok` across 8 sessions (80.0%) · Mainline Yield: `80 tok/ln`") {
		t.Errorf("missing mainline markdown metrics:\n%s", md)
	}
	if !strings.Contains(md, "- **Fork / Prototyping**: `20.0K tok` across 2 sessions (20.0% exploratory churn)") {
		t.Errorf("missing fork markdown metrics:\n%s", md)
	}
}
