// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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
			if !strings.Contains(view, "gemini-3.8-flash") {
				t.Errorf("expected view to contain canonical model gemini-3.8-flash")
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

func TestLiveModel_SeedTodayAndActivityUnits(t *testing.T) {
	// 1. Default NewLiveModel seeds today's usage into session counters
	mSeeded := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false)
	if mSeeded.snapshot.SessionTokens != 40_000 {
		t.Errorf("expected sessionTokens seeded with 40_000, got %d", mSeeded.snapshot.SessionTokens)
	}

	// 2. NewLiveModel with seedToday=false (fresh mode) starts from 0
	mFresh := NewLiveModel(mockLivePollFunc(1), time.Second, "opencode", false, false)
	if mFresh.snapshot.SessionTokens != 0 {
		t.Errorf("expected fresh sessionTokens 0, got %d", mFresh.snapshot.SessionTokens)
	}

	// 3. Activity tool (e.g. Agy) displays step units across all KPI widgets
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
	if mAct.snapshot.SessionTurns != 150 {
		t.Errorf("expected sessionTurns 150, got %d", mAct.snapshot.SessionTurns)
	}

	view := mAct.View().Content
	if !strings.Contains(view, "150 step") {
		t.Errorf("expected view to contain '150 step' for Volume, got:\n%s", view)
	}
	if !strings.Contains(view, "+150 step") {
		t.Errorf("expected view to contain '+150 step' for Burned, got:\n%s", view)
	}
	if !strings.Contains(view, "step/min") {
		t.Errorf("expected view to contain 'step/min' for Rate, got:\n%s", view)
	}
	if !strings.Contains(view, "step/s") {
		t.Errorf("expected view to contain 'step/s' for Speed, got:\n%s", view)
	}
}
