// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"
)

func mockLivePollFunc(eventCount int) LivePollFunc {
	callCount := 0
	return func() ([]thermal.ToolResult, []thermal.ProjectDay, error) {
		callCount++
		results := []thermal.ToolResult{
			{
				Name: "opencode",
				Summary: thermal.Summary{
					LifetimeTokens: int64(100_000 + callCount*15_000),
					Sessions:       5 + callCount,
					Cost:           2.50 + float64(callCount)*0.15,
					ModelBreakdown: map[string]int64{
						"claude-3-5-sonnet": int64(100_000 + callCount*15_000),
					},
				},
				Daily: []thermal.DailyRow{
					{
						Day:    thermal.LocalDay(time.Now()),
						Tokens: int64(25_000 + callCount*15_000),
						Cost:   1.20,
						Input:  10_000,
						Output: 5_000,
						Cache:  10_000,
						Turns:  4 + callCount,
					},
				},
			},
		}

		projects := []thermal.ProjectDay{
			{
				Project: "/home/user/code/thermal",
				Tool:    "opencode",
				Day:     thermal.LocalDay(time.Now()),
				Tokens:  25_000,
			},
		}

		return results, projects, nil
	}
}

func TestLiveModel_StackedAtSixty(t *testing.T) {
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "", false)
	m.width = 60
	m.height = 20
	view := m.View().Content
	if !strings.Contains(view, "[q] Quit") {
		t.Fatalf("stacked frame missing [q] Quit:\n%s", view)
	}
	if !strings.Contains(view, "THERMAL LIVE") {
		t.Fatalf("stacked frame missing title:\n%s", view)
	}
	if !strings.Contains(view, "╰") {
		t.Fatalf("stacked frame missing bottom border:\n%s", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if ansi.StringWidth(line) > 60 {
			t.Fatalf("stacked frame line exceeds 60 columns (%d): %q", ansi.StringWidth(line), line)
		}
	}
}

func TestLiveModel_Dimensions(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"Compact (80x24)", 80, 24},
		{"Large (120x40)", 120, 40},
	}

	for _, sz := range sizes {
		t.Run(sz.name, func(t *testing.T) {
			m := NewLiveModel(mockLivePollFunc(1), time.Second, "", false)
			m.width = sz.width
			m.height = sz.height

			// Trigger one tick to generate an event
			updated, _ := m.Update(liveTickMsg(time.Now().Add(time.Second)))
			m = updated.(LiveModel)

			view := m.View().Content
			if !strings.Contains(view, "THERMAL LIVE") {
				t.Fatalf("expected view to contain 'THERMAL LIVE', got:\n%s", view)
			}
			if !strings.Contains(view, "BURN VELOCITY") {
				t.Fatalf("expected view to contain 'BURN VELOCITY', got:\n%s", view)
			}
			if !strings.Contains(view, "TODAY'S USAGE") {
				t.Fatalf("expected view to contain 'TODAY'S USAGE', got:\n%s", view)
			}
			if !strings.Contains(view, "LIVE SESSION") {
				t.Fatalf("expected view to contain 'LIVE SESSION', got:\n%s", view)
			}
			if !strings.Contains(view, "FLAME INTENSITY") {
				t.Fatalf("expected view to contain 'FLAME INTENSITY', got:\n%s", view)
			}
			if !strings.Contains(view, "ROLLING ACTIVITY") {
				t.Fatalf("expected view to contain 'ROLLING ACTIVITY', got:\n%s", view)
			}
			if !strings.Contains(view, "RECENT TOKEN BURSTS") {
				t.Fatalf("expected view to contain 'RECENT TOKEN BURSTS', got:\n%s", view)
			}
			if !strings.Contains(view, "Controls:") {
				t.Fatalf("expected view to contain controls footer, got:\n%s", view)
			}
		})
	}
}

func TestLiveModel_KeyFlow(t *testing.T) {
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false)
	m.width = 80
	m.height = 24

	// 1. Initial State: not paused, not quitting, not showHelp
	if m.paused || m.quitting || m.showHelp {
		t.Fatalf("unexpected initial state: paused=%v, quitting=%v, showHelp=%v", m.paused, m.quitting, m.showHelp)
	}

	// 2. Space toggles pause
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = updated.(LiveModel)
	if !m.paused {
		t.Fatalf("expected m.paused = true after pressing space")
	}

	view := m.View().Content
	if !strings.Contains(view, "PAUSED") {
		t.Fatalf("expected view to indicate PAUSED, got:\n%s", view)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = updated.(LiveModel)
	if m.paused {
		t.Fatalf("expected m.paused = false after second space")
	}

	// 3. '?' toggles help
	updated, _ = m.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	m = updated.(LiveModel)
	if !m.showHelp {
		t.Fatalf("expected m.showHelp = true after pressing ?")
	}

	helpView := m.View().Content
	if !strings.Contains(helpView, "CONTROLS & MONITOR GUIDE") {
		t.Fatalf("expected help view, got:\n%s", helpView)
	}

	// Esc dismisses help
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(LiveModel)
	if m.showHelp {
		t.Fatalf("expected m.showHelp = false after Esc")
	}
	if m.quitting {
		t.Fatalf("Esc during help should not quit the model")
	}

	// 4. 'c' resets session stats
	// Advance tick first to generate session tokens
	updated, _ = m.Update(liveTickMsg(time.Now().Add(time.Second)))
	m = updated.(LiveModel)
	if m.snapshot.SessionTokens == 0 {
		t.Logf("note: session tokens after tick = %d", m.snapshot.SessionTokens)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	m = updated.(LiveModel)
	if m.snapshot.SessionTokens != 0 {
		t.Fatalf("expected session tokens to be 0 after reset, got %d", m.snapshot.SessionTokens)
	}

	// 5. 'q' quits
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	m = updated.(LiveModel)
	if !m.quitting {
		t.Fatalf("expected m.quitting = true after pressing q")
	}
	if cmd == nil {
		t.Fatalf("expected non-nil tea.Quit cmd after pressing q")
	}
}

func TestLiveModel_BorderAlignmentAcrossWidths(t *testing.T) {
	widths := []int{80, 100, 120, 140, 160, 200, 240}

	for _, w := range widths {
		t.Run(fmt.Sprintf("Width_%d", w), func(t *testing.T) {
			m := NewLiveModel(mockLivePollFunc(1), time.Second, "Agy", true)
			m.width = w
			m.height = 30

			// Trigger tick
			updated, _ := m.Update(liveTickMsg(time.Now().Add(time.Second)))
			m = updated.(LiveModel)

			// Inject a model name and event with empty project
			m.snapshot.ActiveModel = "Gemini 3.8 Flash (High)"
			m.snapshot.RecentEvents = append(m.snapshot.RecentEvents, thermal.LiveEvent{
				Timestamp: time.Now(),
				Tool:      "Agy",
				Tokens:    2,
				Model:     m.snapshot.ActiveModel,
				Project:   "", // empty project to test em-dash placeholder and cost alignment
				Cost:      0.00,
			})

			view := m.View().Content
			if calcColWidths(w-8).showModel && !strings.Contains(view, "gemini-3.8-flash") {
				t.Errorf("expected view to contain canonical model gemini-3.8-flash at width %d", w)
			}
			lines := strings.Split(view, "\n")

			var expectedWidth int
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "•") {
					continue
				}
				// Border lines start with ╭, │, ├, or ╰
				if strings.HasPrefix(trimmed, "╭") || strings.HasPrefix(trimmed, "│") ||
					strings.HasPrefix(trimmed, "├") || strings.HasPrefix(trimmed, "╰") {
					sw := ansi.StringWidth(line)
					if expectedWidth == 0 {
						expectedWidth = sw
					} else if sw != expectedWidth {
						t.Errorf("width %d: line width mismatch: got %d, want %d.\nLine: %s", w, sw, expectedWidth, line)
					}
				}
			}
		})
	}
}

func TestLiveModel_AsyncCollectionAndCoalescing(t *testing.T) {
	callCount := 0
	pollFn := func() ([]thermal.ToolResult, []thermal.ProjectDay, error) {
		callCount++
		return []thermal.ToolResult{
			{
				Name: "opencode",
				Summary: thermal.Summary{
					LifetimeTokens: int64(10_000 * callCount),
				},
				Daily: []thermal.DailyRow{
					{
						Day:    thermal.LocalDay(time.Now()),
						Tokens: int64(10_000 * callCount),
					},
				},
			},
		}, nil, nil
	}

	m := NewLiveModel(pollFn, 100*time.Millisecond, "", false)
	// Initial call happened in NewLiveModel
	if callCount != 1 {
		t.Fatalf("expected 1 call during NewLiveModel, got %d", callCount)
	}
	if m.inFlight {
		t.Fatalf("expected inFlight=false initially")
	}

	// First tick should schedule tickCmd and collectCmd
	updated, cmd := m.Update(liveTickMsg(time.Now()))
	m = updated.(LiveModel)
	if !m.inFlight {
		t.Fatalf("expected inFlight=true after first tick")
	}
	if cmd == nil {
		t.Fatalf("expected batch cmd with tick and collector")
	}

	// Second tick before collection completes should NOT spawn a duplicate collectCmd
	updated, _ = m.Update(liveTickMsg(time.Now().Add(100 * time.Millisecond)))
	m2 := updated.(LiveModel)
	if m2.pollSeq != m.pollSeq {
		t.Fatalf("expected pollSeq to remain %d while inFlight, got %d", m.pollSeq, m2.pollSeq)
	}

	// Deliver completed message
	dataMsg := liveCollectedMsg{
		seq: m.pollSeq,
		results: []thermal.ToolResult{
			{
				Name: "opencode",
				Summary: thermal.Summary{
					LifetimeTokens: 50_000,
				},
				Daily: []thermal.DailyRow{
					{
						Day:    thermal.LocalDay(time.Now()),
						Tokens: 50_000,
					},
				},
			},
		},
		timestamp: time.Now(),
	}
	updated, _ = m.Update(dataMsg)
	m = updated.(LiveModel)
	if m.inFlight {
		t.Fatalf("expected inFlight=false after data delivery")
	}
	if m.snapshot.TodayTokens != 50_000 {
		t.Fatalf("expected today tokens 50000, got %d", m.snapshot.TodayTokens)
	}
}

func TestLiveModel_SlowCollectorDoesNotBlockQuitOrPause(t *testing.T) {
	started := make(chan struct{})
	unblock := make(chan struct{})

	slowPoll := func() ([]thermal.ToolResult, []thermal.ProjectDay, error) {
		close(started)
		<-unblock
		return nil, nil, nil
	}

	m := NewLiveModel(nil, 100*time.Millisecond, "", false)
	m.pollFn = slowPoll

	// Start collection
	updated, cmd := m.Update(liveTickMsg(time.Now()))
	m = updated.(LiveModel)
	if !m.inFlight {
		t.Fatalf("expected inFlight=true")
	}

	// Run collector in background to simulate slow collection
	go func() {
		if cmd != nil {
			msg := cmd()
			if batch, ok := msg.(tea.BatchMsg); ok {
				for _, c := range batch {
					if c != nil {
						go c()
					}
				}
			}
		}
	}()

	<-started

	// While collector is blocked in background, quitting should be immediate (<50ms)
	t0 := time.Now()
	updated, quitCmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	elapsed := time.Since(t0)
	m = updated.(LiveModel)

	if !m.quitting {
		t.Fatalf("expected quitting=true immediately")
	}
	if quitCmd == nil {
		t.Fatalf("expected non-nil quitCmd")
	}
	if elapsed > 50*time.Millisecond {
		t.Fatalf("quit took too long (%v), should be <50ms", elapsed)
	}

	close(unblock)
}

func TestLiveModel_StaleMessageDiscard(t *testing.T) {
	m := NewLiveModel(nil, 100*time.Millisecond, "", false)
	m.appliedSeq = 5

	staleMsg := liveCollectedMsg{
		seq: 3,
		results: []thermal.ToolResult{
			{
				Name:    "opencode",
				Summary: thermal.Summary{LifetimeTokens: 999_999},
			},
		},
		timestamp: time.Now(),
	}

	updated, _ := m.Update(staleMsg)
	m = updated.(LiveModel)
	if m.snapshot.TodayTokens == 999_999 {
		t.Fatalf("expected stale message (seq 3 < 5) to be discarded")
	}
}

func TestLiveModel_PauseStopsCollection(t *testing.T) {
	called := 0
	pollFn := func() ([]thermal.ToolResult, []thermal.ProjectDay, error) {
		called++
		return nil, nil, nil
	}

	m := NewLiveModel(nil, 100*time.Millisecond, "", false)
	m.pollFn = pollFn

	// Pause
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	m = updated.(LiveModel)
	if !m.paused {
		t.Fatalf("expected paused=true")
	}

	// Tick while paused
	updated, cmd := m.Update(liveTickMsg(time.Now()))
	m = updated.(LiveModel)
	if m.inFlight {
		t.Fatalf("expected inFlight=false when paused")
	}
	if called != 0 {
		t.Fatalf("expected pollFn not to be called while paused")
	}
	// Cmd should only be tickLive, not a batch with collector
	if cmd != nil {
		msg := cmd()
		if _, isBatch := msg.(tea.BatchMsg); isBatch {
			t.Fatalf("expected simple tick cmd, not batch with collector when paused")
		}
	}
}

func TestLiveModel_LaunchRelativeSessionAndActivityUnits(t *testing.T) {
	// 1. Default NewLiveModel records today's usage as carried-in, not as session activity.
	mSeeded := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false)
	if mSeeded.snapshot.SessionTokens != 0 {
		t.Errorf("expected sessionTokens 0 at launch, got %d", mSeeded.snapshot.SessionTokens)
	}
	if mSeeded.snapshot.CarriedInTokens != 40_000 {
		t.Errorf("expected carriedInTokens 40_000, got %d", mSeeded.snapshot.CarriedInTokens)
	}
	if mSeeded.snapshot.CarriedInTurns != 5 {
		t.Errorf("expected carriedInTurns 5, got %d", mSeeded.snapshot.CarriedInTurns)
	}
	if !strings.Contains(mSeeded.View().Content, "Carried in: 40.0K tok") {
		t.Errorf("expected carried-in line, got:\n%s", mSeeded.View().Content)
	}

	// 2. Activity tool (e.g. Agy) displays step units across all KPI widgets
	actPoll := func() ([]thermal.ToolResult, []thermal.ProjectDay, error) {
		return []thermal.ToolResult{
			{
				Tool: thermal.ToolAgy,
				Name: "Agy",
				Summary: thermal.Summary{
					LifetimeTokens: 0,
					Sessions:       12,
				},
				Daily: []thermal.DailyRow{
					{
						Day:   thermal.LocalDay(time.Now()),
						Turns: 150,
					},
				},
			},
		}, nil, nil
	}

	mAct := NewLiveModel(actPoll, time.Second, "Agy", false)
	mAct.width = 100
	mAct.height = 30

	if mAct.snapshot.TodayTokens != 0 {
		t.Errorf("expected todayTokens 0 for activity tool, got %d", mAct.snapshot.TodayTokens)
	}
	if mAct.snapshot.TodayTurns != 150 {
		t.Errorf("expected todayTurns 150, got %d", mAct.snapshot.TodayTurns)
	}
	if mAct.snapshot.SessionTurns != 0 {
		t.Errorf("expected sessionTurns 0 at launch, got %d", mAct.snapshot.SessionTurns)
	}
	if mAct.snapshot.CarriedInTurns != 150 {
		t.Errorf("expected carriedInTurns 150, got %d", mAct.snapshot.CarriedInTurns)
	}

	view := mAct.View().Content
	if !strings.Contains(view, "150 step") {
		t.Errorf("expected view to contain '150 step' for Volume, got:\n%s", view)
	}
	if !strings.Contains(view, "Carried in: 150 step") {
		t.Errorf("expected view to contain carried-in steps, got:\n%s", view)
	}
	if !strings.Contains(view, "Since launch: 0 tok") {
		t.Errorf("expected explicit empty session state, got:\n%s", view)
	}
	if strings.Contains(view, "+150 step") {
		t.Errorf("expected no seeded session delta, got:\n%s", view)
	}
	if !strings.Contains(view, "step/m") {
		t.Errorf("expected view to contain 'step/m' for Rate, got:\n%s", view)
	}
	if !strings.Contains(view, "step/s") {
		t.Errorf("expected view to contain 'step/s' for Speed, got:\n%s", view)
	}
}

func TestLiveModel_HybridTokenAndActivity(t *testing.T) {
	hybridPoll := func() ([]thermal.ToolResult, []thermal.ProjectDay, error) {
		return []thermal.ToolResult{
			{
				Tool: thermal.ToolCodex,
				Name: "Codex",
				Summary: thermal.Summary{
					LifetimeTokens: 6_200_000,
					Sessions:       2,
				},
				Daily: []thermal.DailyRow{
					{
						Day:    thermal.LocalDay(time.Now()),
						Tokens: 6_200_000,
						Input:  5_000_000,
						Output: 1_200_000,
						Turns:  2,
					},
				},
			},
			{
				Tool: thermal.ToolAgy,
				Name: "Agy",
				Summary: thermal.Summary{
					LifetimeTokens: 0,
					Sessions:       10,
					ModelBreakdown: map[string]int64{
						"gemini-3.8-flash": 4_000,
					},
				},
				Daily: []thermal.DailyRow{
					{
						Day:   thermal.LocalDay(time.Now()),
						Turns: 4_000,
					},
				},
			},
		}, nil, nil
	}

	m := NewLiveModel(hybridPoll, time.Second, "", false)
	m.width = 120
	m.height = 30

	view := m.View().Content
	if !strings.Contains(view, "6.2M tok · 4.0K step") {
		t.Errorf("expected view to contain hybrid volume '6.2M tok · 4.0K step', got:\n%s", view)
	}
	if !strings.Contains(view, "Since launch: 0 tok") {
		t.Errorf("expected view to contain explicit empty session state, got:\n%s", view)
	}
	if strings.Contains(view, "+6.2M tok · +4.0K step") {
		t.Errorf("expected no seeded hybrid session delta, got:\n%s", view)
	}
	if !strings.Contains(view, "Carried in: 6.2M tok · 4.0K step") {
		t.Errorf("expected carried-in hybrid volume, got:\n%s", view)
	}
	if !strings.Contains(view, "RECENT TOKEN BURSTS") {
		t.Errorf("expected an empty token ticker, got:\n%s", view)
	}
	if !strings.Contains(view, "TOKENS") {
		t.Errorf("expected token ticker header, got:\n%s", view)
	}
}

func TestLiveModel_FeedStateIndependentOfActivity(t *testing.T) {
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false)
	m.width = 100
	m.height = 30

	view := m.View().Content
	if !strings.Contains(view, "● LIVE") {
		t.Errorf("expected healthy feed indicator, got:\n%s", view)
	}
	if !strings.Contains(view, "IDLE") {
		t.Errorf("expected activity state IDLE alongside LIVE, got:\n%s", view)
	}

	m.lastPollAt = time.Now().Add(-10 * time.Second)
	view = m.View().Content
	if !strings.Contains(view, "● STALE") {
		t.Errorf("expected stale feed indicator, got:\n%s", view)
	}

	m.pollErr = fmt.Errorf("collector unavailable")
	view = m.View().Content
	if !strings.Contains(view, "● OFFLINE") {
		t.Errorf("expected offline feed indicator, got:\n%s", view)
	}
}

func TestLiveModel_IdleGlyphs(t *testing.T) {
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false)
	m.width = 100
	m.height = 30

	view := m.View().Content
	if !strings.Contains(view, "no activity (60s)") {
		t.Errorf("expected idle sparkline caption, got:\n%s", view)
	}
	if !strings.Contains(view, "Flame: [·") {
		t.Errorf("expected dotted idle flame, got:\n%s", view)
	}
	if !strings.Contains(view, "Spark: [·") {
		t.Errorf("expected dotted idle sparkline, got:\n%s", view)
	}
}

func TestLiveModel_LastEventAge(t *testing.T) {
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false)
	m.width = 100
	m.height = 30
	m.lastEventAt = time.Now().Add(-2 * time.Minute)

	view := m.View().Content
	if !strings.Contains(view, "Last event: 2m ago") {
		t.Errorf("expected last-event age, got:\n%s", view)
	}
}

func TestLiveModel_TodayRollups(t *testing.T) {
	m := NewLiveModel(nil, time.Second, "", false)
	m.width = 120
	m.height = 40
	m.snapshot.TodayProjects = []thermal.LiveRollup{
		{Name: "thermal-streak", Cost: 4.43, CostKind: thermal.LiveCostEstimated},
		{Name: "scripts", Cost: 0.46, CostKind: thermal.LiveCostEstimated},
	}
	m.snapshot.TodayModels = []thermal.LiveRollup{
		{Name: "deepseek-v4.1-flash", Cost: 4.43, CostKind: thermal.LiveCostEstimated},
	}

	view := m.View().Content
	if !strings.Contains(view, "Today by project: thermal-streak ~$4.43 · scripts ~$0.46") {
		t.Errorf("expected project cost rollup, got:\n%s", view)
	}
	if !strings.Contains(view, "Top model (est. cost): deepseek-v4.1-flash ~$4.43") {
		t.Errorf("expected top model rollup, got:\n%s", view)
	}
}

func TestLiveModel_NoDuplicateSessionPanel(t *testing.T) {
	m := NewLiveModel(nil, time.Second, "", false)
	m.width = 80
	m.height = 24
	m.snapshot.TodayTokens = 1_000
	m.snapshot.TodayTurns = 10
	m.snapshot.TodayCost = 0.50
	m.snapshot.SessionTokens = 1_000
	m.snapshot.SessionTurns = 10
	m.snapshot.SessionCost = 0.50

	view := m.View().Content
	if !strings.Contains(view, "= Today's usage") {
		t.Errorf("expected session panel to defer to Today, got:\n%s", view)
	}
	if strings.Contains(view, "Burned:") {
		t.Errorf("expected no duplicated session values, got:\n%s", view)
	}
}

func TestLiveModel_CompactViewport(t *testing.T) {
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "", false)
	m.width = 80
	m.height = 14

	view := m.View().Content
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) > 14 {
		t.Errorf("expected at most 14 lines in compact viewport, got %d:\n%s", len(lines), view)
	}
	if !strings.Contains(view, "THERMAL LIVE") {
		t.Errorf("expected top line in compact viewport, got:\n%s", view)
	}
	if !strings.Contains(view, "Controls:") {
		t.Errorf("expected footer controls in compact viewport, got:\n%s", view)
	}
	if strings.Contains(view, "ROLLING ACTIVITY") {
		t.Errorf("expected sparkline to be omitted at height 14, got:\n%s", view)
	}
}

func TestLiveTUI_FormatSpeed_Fractional(t *testing.T) {
	got := formatSpeed(0, 0.667, true, 24, false)
	want := "0.667 step/s"
	if got != want {
		t.Errorf("formatSpeed(0, 0.667, true, 24, false) = %q, want %q", got, want)
	}

	gotTok := formatSpeed(0.5, 0, false, 24, false)
	wantTok := "0.500 tok/s"
	if gotTok != wantTok {
		t.Errorf("formatSpeed(0.5, 0, false, 24, false) = %q, want %q", gotTok, wantTok)
	}
}

func TestLiveTUI_UnitsAndNoTruncation(t *testing.T) {
	rate := formatRate(3_400_000, 23, false, 80, false)
	if !strings.Contains(rate, "3.4M tok/m · 23 step/m") {
		t.Errorf("rate units = %q", rate)
	}
	speed := formatSpeed(56_700, 0.383, false, 80, false)
	if !strings.Contains(speed, "56.7K tok/s · 0.383 step/s") {
		t.Errorf("speed units = %q", speed)
	}
	narrow := formatSpeed(56_700, 0.383, false, 22, false)
	if strings.Contains(narrow, "step") || strings.Contains(narrow, "…") || !strings.Contains(narrow, "tok/s") {
		t.Errorf("narrow speed truncated a unit: %q", narrow)
	}
	delta := formatDelta(1_700_000, 60, 80, false)
	if !strings.Contains(delta, "+1.7M tok") || !strings.Contains(delta, "+60 step") || strings.Contains(delta, "(+60 s)") {
		t.Errorf("burned units = %q", delta)
	}
	volume := formatVolume(1_700_000, 60, 80, false)
	if !strings.Contains(volume, "step") {
		t.Errorf("volume units = %q", volume)
	}
}

func TestLiveTUI_HybridUnitsAtWidths(t *testing.T) {
	for _, width := range []int{80, 100, 120} {
		m := NewLiveModel(nil, time.Second, "", false)
		m.width = width
		m.height = 30
		m.snapshot.BurnTokensPerMin = 3_400_000
		m.snapshot.BurnTurnsPerMin = 23
		m.snapshot.BurnTokensPerSec = 56_700
		m.snapshot.BurnTurnsPerSec = 0.383
		m.snapshot.TodayTokens = 1_700_000
		m.snapshot.TodayTurns = 60
		m.snapshot.TodayCost = 1.20
		m.snapshot.TodayCacheHit = 96.1
		m.snapshot.SessionTokens = 1_600_000
		m.snapshot.SessionTurns = 60
		m.snapshot.SessionCost = 1.10
		m.snapshot.ActiveModel = "very-long-model-name-that-must-truncate"

		view := m.View().Content
		for _, unit := range []string{"tok/m", "step/m", "tok/s", "step/s"} {
			if !strings.Contains(view, unit) {
				t.Errorf("width %d: missing unit %q:\n%s", width, unit, view)
			}
		}
		if strings.Contains(view, "step…") || strings.Contains(view, "tok…") {
			t.Errorf("width %d: unit was truncated:\n%s", width, view)
		}
		if !strings.Contains(view, "Cache hit: 96.1%") {
			t.Errorf("width %d: cache precision drift:\n%s", width, view)
		}
		if width >= 100 && !strings.Contains(view, "+60 step") {
			t.Errorf("width %d: burned steps missing unit:\n%s", width, view)
		}
		if !strings.Contains(view, "Controls:") || !strings.Contains(view, "[s]") {
			t.Errorf("width %d: controls changed:\n%s", width, view)
		}
	}
}

func TestLiveTUI_ModelCostBasisLabel(t *testing.T) {
	lines := formatTodayRollups(nil, []thermal.LiveRollup{
		{Name: "model-a", Cost: 3.954, CostKind: thermal.LiveCostEstimated},
	})
	if len(lines) != 1 || !strings.Contains(lines[0], "Top model (est. cost): model-a ~$3.95") {
		t.Errorf("model rollup missing estimate label: %v", lines)
	}
}

func TestLiveModel_GroupedBurstsAndSort(t *testing.T) {
	now := time.Now()
	m := NewLiveModel(nil, time.Second, "", false)
	m.width = 120
	m.height = 40
	m.lastPollAt = now
	m.lastEventAt = now
	m.snapshot.RecentEvents = []thermal.LiveEvent{
		{Timestamp: now, Tool: "CMD", Model: "model", Project: "thermal", Tokens: 1_000_000, Cost: 0.10, CostKind: thermal.LiveCostRecorded},
		{Timestamp: now.Add(-2 * time.Second), Tool: "CMD", Model: "model", Project: "thermal", Tokens: 500_000, Cost: 0.10, CostKind: thermal.LiveCostRecorded},
		{Timestamp: now.Add(-3 * time.Second), Tool: "CMD", Model: "model", Project: "thermal", Tokens: 200_000, Cost: 0.10, CostKind: thermal.LiveCostRecorded},
		{Timestamp: now.Add(-4 * time.Second), Tool: "Agy", Project: "scripts", Turns: 10, CostKind: thermal.LiveCostUnpriced},
	}

	view := m.View().Content
	if !strings.Contains(view, "CMD×3") {
		t.Errorf("expected grouped burst count, got:\n%s", view)
	}
	if !strings.Contains(view, "+1.7M") || !strings.Contains(view, "$0.30") {
		t.Errorf("expected summed tokens and cost, got:\n%s", view)
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	m = updated.(LiveModel)
	if m.sortMode != "cost" {
		t.Fatalf("expected cost sort, got %q", m.sortMode)
	}
	view = m.View().Content
	if !strings.Contains(view, "Sort: cost") {
		t.Errorf("expected sort mode in footer, got:\n%s", view)
	}
	if !strings.Contains(view, "Top spenders (60s):") {
		t.Errorf("expected top spenders line, got:\n%s", view)
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	m = updated.(LiveModel)
	if m.sortMode != "time" {
		t.Fatalf("expected time sort after toggle, got %q", m.sortMode)
	}
}

func TestLiveTUI_GroupCostAndSort(t *testing.T) {
	now := time.Now()
	group := thermal.LiveEventGroup{
		Tool: "CMD", Model: "model", Project: "thermal",
		Count: 3, Latest: now, Tokens: 1_700_000, Cost: 4.43,
		CostKind: thermal.LiveCostEstimated, AnyUnpriced: true,
	}
	row := formatGroupRow(group, false, false, now, 116, false)
	if !strings.Contains(row, "CMD×3") {
		t.Errorf("expected burst count, got %q", row)
	}
	if !strings.Contains(row, "≥$4.43") {
		t.Errorf("expected partial-known cost marker, got %q", row)
	}
	group.Cost = 0
	group.CostKind = thermal.LiveCostUnpriced
	row = formatGroupRow(group, false, false, now, 116, false)
	if !strings.Contains(row, "—") {
		t.Errorf("expected unpriced placeholder, got %q", row)
	}

	first := thermal.LiveEventGroup{Tool: "A", Latest: now, Cost: 0.10, Tokens: 10}
	second := thermal.LiveEventGroup{Tool: "B", Latest: now.Add(-time.Minute), Cost: 5.00, Tokens: 20}
	timeGroups := []thermal.LiveEventGroup{second, first}
	sortLiveGroups(timeGroups, "time")
	if timeGroups[0].Tool != "A" {
		t.Errorf("expected newest first in time mode, got %+v", timeGroups)
	}
	costGroups := []thermal.LiveEventGroup{first, second}
	sortLiveGroups(costGroups, "cost")
	if costGroups[0].Tool != "B" {
		t.Errorf("expected highest cost first in cost mode, got %+v", costGroups)
	}

	line := formatTopSpenders([]thermal.LiveEventGroup{
		{Project: "thermal", Cost: 4.43, CostKind: thermal.LiveCostEstimated},
		{Project: "scripts", Cost: 0.46, CostKind: thermal.LiveCostEstimated},
		{Project: "other", Cost: 0.10, CostKind: thermal.LiveCostRecorded},
		{Project: "free", Cost: 0, CostKind: thermal.LiveCostUnpriced},
	}, false)
	if !strings.Contains(line, "thermal ~$4.43") || !strings.Contains(line, "scripts ~$0.46") {
		t.Errorf("unexpected top spenders line: %q", line)
	}
}

func TestLiveTUI_FlameCeilingCaption(t *testing.T) {
	m := NewLiveModel(nil, time.Second, "", false)
	m.width = 100
	m.height = 30
	m.snapshot.FlameIntensity = 3_400_000.0 / 9_000_000.0
	m.snapshot.FlameMetric = "tokens"
	m.snapshot.FlameCeiling = 9_000_000

	view := m.View().Content
	if !strings.Contains(view, "38% of 9.0M tok/m") {
		t.Errorf("expected token ceiling caption, got:\n%s", view)
	}
	if !strings.Contains(view, "WARM") {
		t.Errorf("expected WARM state, got:\n%s", view)
	}

	m.snapshot.FlameIntensity = 23.0 / 60.0
	m.snapshot.FlameMetric = "steps"
	m.snapshot.FlameCeiling = 60
	view = m.View().Content
	if !strings.Contains(view, "38% of 60 step/m") {
		t.Errorf("expected step ceiling caption, got:\n%s", view)
	}
}

func TestLiveTUI_BucketLabels(t *testing.T) {
	m := NewLiveModel(nil, time.Second, "", false)
	m.width = 100
	m.height = 30
	m.snapshot.PeakTokensPerSec = 1_000
	m.snapshot.RollingTokens[0] = 1_000

	view := m.View().Content
	if !strings.Contains(view, "60s / 1s buckets") {
		t.Errorf("expected one-second window label, got:\n%s", view)
	}
	if !strings.Contains(view, "Peak: 1.0K tok/s (1s bucket)") {
		t.Errorf("expected one-second peak label, got:\n%s", view)
	}
}

func TestLiveTUI_RateSpeedParity(t *testing.T) {
	for _, turnsPerMin := range []float64{1, 7.8, 8, 9.9, 23, 60} {
		rateText := strings.TrimSuffix(formatRate(0, turnsPerMin, true, 24, false), " step/m")
		speedText := strings.TrimSuffix(formatSpeed(0, turnsPerMin/60, true, 24, false), " step/s")
		var rate, speed float64
		if _, err := fmt.Sscanf(rateText, "%f", &rate); err != nil {
			t.Fatalf("parse rate %q: %v", rateText, err)
		}
		if _, err := fmt.Sscanf(speedText, "%f", &speed); err != nil {
			t.Fatalf("parse speed %q: %v", speedText, err)
		}
		if math.Abs(speed*60-rate) > 0.05 {
			t.Errorf("turns/min %v: rate %q and speed %q disagree", turnsPerMin, rateText, speedText)
		}
	}
}

func TestLiveTUI_FormatSpeed_ZeroMuted(t *testing.T) {
	gotAct := formatSpeed(0, 0.0, true, 24, false)
	if gotAct != "0 step/s" {
		t.Errorf("expected '0 step/s', got %q", gotAct)
	}

	gotTok := formatSpeed(0.0, 0.0, false, 24, false)
	if gotTok != "0 tok/s" {
		t.Errorf("expected '0 tok/s', got %q", gotTok)
	}
}

func TestLiveTUI_FormatCacheHit_Responsive(t *testing.T) {
	formatted := formatCacheHit(86.0, 24, false)
	if formatted != "Cache hit: 86.0%" {
		t.Errorf("formatCacheHit(86, 24, false) = %q", formatted)
	}

	line := padCell(formatted, 24, false)
	if strings.Contains(line, "…") {
		t.Errorf("padCell clipped cache hit in colW=24: %q", line)
	}
	if ansi.StringWidth(line) != 24 {
		t.Errorf("expected cell width 24, got %d for %q", ansi.StringWidth(line), line)
	}

	// Verify inside the full LiveModel view at 80x24.
	m := NewLiveModel(mockLivePollFunc(1), time.Second, "", false)
	m.width = 80
	m.height = 24
	m.snapshot.TodayTurns = 6642
	m.snapshot.TodayCacheHit = 86.0

	view := m.View().Content
	if !strings.Contains(view, "Cache hit: 86.0%") {
		t.Errorf("expected cache-hit label in view, got:\n%s", view)
	}
	if strings.Contains(view, "Turns:") {
		t.Errorf("expected redundant Turns line removed, got:\n%s", view)
	}
}

func TestLiveTUI_FormatCacheHit_NoAmbiguousHit(t *testing.T) {
	for _, width := range []int{8, 10, 12, 14, 24} {
		formatted := formatCacheHit(86.0, width, false)
		lower := strings.ToLower(formatted)
		if strings.Contains(lower, "hit") && !strings.Contains(lower, "cache") {
			t.Errorf("width %d: bare hit in %q", width, formatted)
		}
		if ansi.StringWidth(formatted) > width && !strings.Contains(formatted, "…") {
			t.Errorf("width %d: %q overflows", width, formatted)
		}
	}
	if got := formatCacheHit(0, 24, false); got != "Cache hit: —" {
		t.Errorf("expected cache-hit placeholder, got %q", got)
	}
}

func TestLiveTUI_TickerColumnsAndCost(t *testing.T) {
	now := time.Now()
	ev := thermal.LiveEvent{
		Timestamp: now,
		Tool:      "CMD",
		Tokens:    1_000,
		Turns:     2,
		Model:     "model",
		Project:   "project",
		Cost:      0.0042,
		CostKind:  thermal.LiveCostEstimated,
	}
	row := formatEventRow(ev, false, false, now, 116, false)
	if !strings.Contains(row, "+1.0K") || !strings.Contains(row, "+2") {
		t.Errorf("expected separate token/step deltas, got %q", row)
	}
	if strings.Contains(row, "tok") || strings.Contains(row, "step") {
		t.Errorf("numeric cells must not carry units, got %q", row)
	}
	if !strings.Contains(row, "~$0.0042") {
		t.Errorf("expected sub-cent estimate, got %q", row)
	}

	header := formatTickerHeader(116, false)
	if !strings.Contains(header, "TOKENS") || !strings.Contains(header, "STEPS") {
		t.Errorf("expected separate token and step headers, got %q", header)
	}
	if strings.Contains(header, "TOK/STEP") {
		t.Errorf("mixed unit header still present: %q", header)
	}
	narrow := formatTickerHeader(76, false)
	if !strings.Contains(narrow, "TOK") || !strings.Contains(narrow, "STEP") {
		t.Errorf("expected narrow separate headers, got %q", narrow)
	}
}

func TestLiveTUI_TickerWidths(t *testing.T) {
	now := time.Now()
	ev := thermal.LiveEvent{
		Timestamp: now,
		Tool:      "CMD",
		Tokens:    651_400_000,
		Turns:     13_302,
		Model:     "deepseek-v4.1-flash",
		Project:   "thermal-streak",
		Cost:      4.43,
		CostKind:  thermal.LiveCostEstimated,
	}
	for _, inner := range []int{70, 72, 76, 85, 86, 92, 96, 112, 116} {
		header := formatTickerHeader(inner, false)
		if w := ansi.StringWidth(header); w > inner {
			t.Errorf("inner %d: header width %d overflows: %q", inner, w, header)
		}
		row := formatEventRow(ev, true, false, now, inner, false)
		if w := ansi.StringWidth(row); w > inner {
			t.Errorf("inner %d: row width %d overflows: %q", inner, w, row)
		}
	}
}

func TestLiveTUI_TickerCostAndAge(t *testing.T) {
	now := time.Now()
	unpriced := formatEventRow(thermal.LiveEvent{Timestamp: now, Tool: "Agy", Turns: 3, CostKind: thermal.LiveCostUnpriced}, false, false, now, 116, false)
	if !strings.Contains(unpriced, "—") {
		t.Errorf("expected unpriced placeholder, got %q", unpriced)
	}
	zero := formatEventRow(thermal.LiveEvent{Timestamp: now, Tool: "CMD", Tokens: 1, Cost: 0, CostKind: thermal.LiveCostRecorded}, false, false, now, 116, false)
	if !strings.Contains(zero, "$0.00") {
		t.Errorf("expected recorded zero, got %q", zero)
	}
	stale := formatEventRow(thermal.LiveEvent{Timestamp: now.Add(-2 * time.Minute), Tool: "CMD", Tokens: 400}, true, false, now, 116, false)
	if !strings.Contains(stale, "+400!") {
		t.Errorf("expected outlier marker, got %q", stale)
	}
	if !strings.Contains(stale, "2m*") {
		t.Errorf("expected stale age marker, got %q", stale)
	}
}

func TestLiveTUI_GaugeBracketWidthAlignment(t *testing.T) {
	widths := []int{80, 100, 120, 160}

	for _, w := range widths {
		t.Run(fmt.Sprintf("Width_%d", w), func(t *testing.T) {
			innerWidth := w - 4
			flameLine := formatFlameLine(0.5, "tokens", thermal.DefaultFlameCeilingTokens, "claude-3-5-sonnet", innerWidth, false)
			sparkLine := formatSparkLine([60]int64{}, 1000, innerWidth, false, false)

			flameOpen := strings.Index(flameLine, "[")
			sparkOpen := strings.Index(sparkLine, "[")
			if ansi.StringWidth(flameLine[:flameOpen]) != ansi.StringWidth(sparkLine[:sparkOpen]) {
				t.Errorf("width %d: flame open pos [%d] != spark open pos [%d]", w, ansi.StringWidth(flameLine[:flameOpen]), ansi.StringWidth(sparkLine[:sparkOpen]))
			}

			flameClose := strings.Index(flameLine, "]")
			sparkClose := strings.Index(sparkLine, "]")

			flameInner := flameLine[flameOpen+1 : flameClose]
			sparkInner := sparkLine[sparkOpen+1 : sparkClose]

			flameGaugeW := ansi.StringWidth(flameInner)
			sparkGaugeW := ansi.StringWidth(sparkInner)
			if flameGaugeW != sparkGaugeW {
				t.Errorf("width %d: flame bracket visual width %d != spark bracket visual width %d", w, flameGaugeW, sparkGaugeW)
			}
		})
	}
}

func TestLiveHelpUsesSharedCostLegend(t *testing.T) {
	m := NewLiveModel(nil, time.Second, "", false)
	if !strings.Contains(m.renderHelp(), render.CostLegendText()) {
		t.Fatalf("live help does not use the shared cost legend:\n%s", m.renderHelp())
	}
}
