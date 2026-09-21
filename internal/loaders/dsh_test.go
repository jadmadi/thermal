package loaders

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDshData_MultiSessionDir(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "storages", "session_projcache", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	// Session 1: DeepSeek model, 2 turns, active
	s1 := dshSessionWrapper{
		Version: 5,
		Record: dshRecord{
			Identity: dshIdentity{
				CreatedAt: int64(1789916748480), // 2026-09-20T15:05:48.480Z
				Cwd:       "/home/user/projects/repo1",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{
					Val: &dshSessionStatsVal{
						Turns:  2,
						Steps:  10,
						LLMMs:  15000,
						ToolMs: 5000,
					},
				},
				ModelSelection: &dshModelSelectionRow{
					Val: &dshModelSelectionVal{
						LastUsed: &dshLastUsed{
							Model: "DeepSeek-V4-Flash",
						},
					},
				},
				TokenUsage: &dshTokenUsageRow{
					Val: &dshTokenUsageVal{
						Totals: &dshTokenTotals{
							UncachedInputTokens: 1000,
							OutputTokens:        200,
							CacheReadTokens:     500,
							CacheWriteTokens:    100,
						},
					},
				},
			},
		},
	}
	data1, _ := json.Marshal(s1)
	if err := os.WriteFile(filepath.Join(sessionsDir, "session-1.json"), data1, 0o644); err != nil {
		t.Fatalf("failed to write session 1: %v", err)
	}

	// Session 2: Another session, different model, same day
	s2 := dshSessionWrapper{
		Version: 5,
		Record: dshRecord{
			Identity: dshIdentity{
				CreatedAt: float64(1789917000000), // same day 2026-09-20
				Cwd:       "/home/user/projects/repo2",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{
					Val: &dshSessionStatsVal{
						Turns:  3,
						Steps:  5,
						LLMMs:  25000,
						ToolMs: 10000,
					},
				},
				ModelSelection: &dshModelSelectionRow{
					Val: &dshModelSelectionVal{
						LastUsed: &dshLastUsed{
							Model: "atria-dawn-preview",
						},
					},
				},
				TokenUsage: &dshTokenUsageRow{
					Val: &dshTokenUsageVal{
						Totals: &dshTokenTotals{
							UncachedInputTokens: 2000,
							OutputTokens:        400,
							CacheReadTokens:     1000,
							CacheWriteTokens:    0,
						},
					},
				},
			},
		},
	}
	data2, _ := json.Marshal(s2)
	if err := os.WriteFile(filepath.Join(sessionsDir, "session-2.json"), data2, 0o644); err != nil {
		t.Fatalf("failed to write session 2: %v", err)
	}

	// Session 3: Inactive session (0 turns, 0 tokens) - should be skipped
	s3 := dshSessionWrapper{
		Version: 5,
		Record: dshRecord{
			Identity: dshIdentity{
				CreatedAt: int64(1789917500000),
				Cwd:       "/home/user/projects/repo1",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{
					Val: &dshSessionStatsVal{Turns: 0},
				},
			},
		},
	}
	data3, _ := json.Marshal(s3)
	if err := os.WriteFile(filepath.Join(sessionsDir, "session-3.json"), data3, 0o644); err != nil {
		t.Fatalf("failed to write session 3: %v", err)
	}

	summary, daily, projects, err := LoadDshData(tmpDir)
	if err != nil {
		t.Fatalf("LoadDshData failed: %v", err)
	}

	if summary.Tool != "dsh" {
		t.Errorf("expected Tool dsh, got %s", summary.Tool)
	}
	if summary.Sessions != 2 {
		t.Errorf("expected 2 active sessions, got %d", summary.Sessions)
	}

	expectedInput := int64(1000 + 2000)
	expectedOutput := int64(200 + 400)
	expectedCache := int64(600 + 1000)
	expectedLifetime := expectedInput + expectedOutput + expectedCache

	if summary.InputTokens != expectedInput {
		t.Errorf("InputTokens = %d, want %d", summary.InputTokens, expectedInput)
	}
	if summary.OutputTokens != expectedOutput {
		t.Errorf("OutputTokens = %d, want %d", summary.OutputTokens, expectedOutput)
	}
	if summary.CacheTokens != expectedCache {
		t.Errorf("CacheTokens = %d, want %d", summary.CacheTokens, expectedCache)
	}
	if summary.LifetimeTokens != expectedLifetime {
		t.Errorf("LifetimeTokens = %d, want %d", summary.LifetimeTokens, expectedLifetime)
	}
	if summary.LongestSessionMs != 35000 {
		t.Errorf("LongestSessionMs = %d, want 35000", summary.LongestSessionMs)
	}

	// Check canonical model breakdown (lowercased)
	if summary.ModelBreakdown["deepseek-v4-flash"] != 1 {
		t.Errorf("expected 1 session for deepseek-v4-flash, got %d", summary.ModelBreakdown["deepseek-v4-flash"])
	}
	if summary.ModelBreakdown["atria-dawn-preview"] != 1 {
		t.Errorf("expected 1 session for atria-dawn-preview, got %d", summary.ModelBreakdown["atria-dawn-preview"])
	}

	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	if daily[0].Turns != 5 {
		t.Errorf("Daily Turns = %d, want 5", daily[0].Turns)
	}
	if daily[0].Tokens != expectedLifetime {
		t.Errorf("Daily Tokens = %d, want %d", daily[0].Tokens, expectedLifetime)
	}

	if len(projects) != 2 {
		t.Errorf("expected 2 project rows, got %d", len(projects))
	}
}

func TestLoadDshData_MonolithFallback(t *testing.T) {
	tmpDir := t.TempDir()
	storagesDir := filepath.Join(tmpDir, "storages")
	if err := os.MkdirAll(storagesDir, 0o755); err != nil {
		t.Fatalf("failed to create storages dir: %v", err)
	}

	mono := dshMonolith{}
	mono.Tables.Sessions = map[string]dshRecord{
		"session-mono-1": {
			Identity: dshIdentity{
				CreatedAt: "2026-09-18T10:00:00Z",
				Cwd:       "/home/user/work/projectA",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{
					Val: &dshSessionStatsVal{Turns: 1, LLMMs: 5000},
				},
				ModelSelection: &dshModelSelectionRow{
					Val: &dshModelSelectionVal{
						LastUsed: &dshLastUsed{Model: "deepseek-v4-flash"},
					},
				},
				TokenUsage: &dshTokenUsageRow{
					Val: &dshTokenUsageVal{
						Totals: &dshTokenTotals{
							UncachedInputTokens: 500,
							OutputTokens:        100,
							CacheReadTokens:     200,
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(mono)
	if err := os.WriteFile(filepath.Join(storagesDir, "session_projcache.json"), data, 0o644); err != nil {
		t.Fatalf("failed to write monolith: %v", err)
	}

	summary, daily, projects, err := LoadDshData(tmpDir)
	if err != nil {
		t.Fatalf("LoadDshData failed on monolith: %v", err)
	}

	if summary.Sessions != 1 {
		t.Errorf("expected 1 session from monolith, got %d", summary.Sessions)
	}
	if summary.LifetimeTokens != 800 {
		t.Errorf("LifetimeTokens = %d, want 800", summary.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Fatalf("expected 1 daily row, got %d", len(daily))
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project row, got %d", len(projects))
	}
}

func TestLoadDshData_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "storages", "session_projcache", "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	// Individual file with ID session-shared
	s1 := dshSessionWrapper{
		Version: 5,
		Record: dshRecord{
			Identity: dshIdentity{
				CreatedAt: int64(1789916748480),
				Cwd:       "/home/user/p",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{Val: &dshSessionStatsVal{Turns: 1}},
				TokenUsage: &dshTokenUsageRow{
					Val: &dshTokenUsageVal{
						Totals: &dshTokenTotals{UncachedInputTokens: 100},
					},
				},
			},
		},
	}
	d1, _ := json.Marshal(s1)
	_ = os.WriteFile(filepath.Join(sessionsDir, "session-shared.json"), d1, 0o644)

	// Monolith containing BOTH session-shared and session-new
	mono := dshMonolith{}
	mono.Tables.Sessions = map[string]dshRecord{
		"session-shared": {
			Identity: dshIdentity{
				CreatedAt: int64(1789916748480),
				Cwd:       "/home/user/p",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{Val: &dshSessionStatsVal{Turns: 1}},
				TokenUsage: &dshTokenUsageRow{
					Val: &dshTokenUsageVal{
						Totals: &dshTokenTotals{UncachedInputTokens: 100},
					},
				},
			},
		},
		"session-new": {
			Identity: dshIdentity{
				CreatedAt: int64(1789916748480),
				Cwd:       "/home/user/p",
			},
			Rows: dshRows{
				SessionStats: &dshSessionStatsRow{Val: &dshSessionStatsVal{Turns: 2}},
				TokenUsage: &dshTokenUsageRow{
					Val: &dshTokenUsageVal{
						Totals: &dshTokenTotals{UncachedInputTokens: 200},
					},
				},
			},
		},
	}
	d2, _ := json.Marshal(mono)
	_ = os.WriteFile(filepath.Join(tmpDir, "storages", "session_projcache.json"), d2, 0o644)

	summary, _, _, err := LoadDshData(tmpDir)
	if err != nil {
		t.Fatalf("LoadDshData failed: %v", err)
	}

	// Should have exactly 2 sessions (session-shared not duplicated)
	if summary.Sessions != 2 {
		t.Errorf("Sessions = %d, want 2 (deduplicated)", summary.Sessions)
	}
	if summary.InputTokens != 300 {
		t.Errorf("InputTokens = %d, want 300", summary.InputTokens)
	}
}

func TestLoadDshData_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	// Case A: Nil modelSelection and nil sessionStats, but positive tokens
	sA := map[string]any{
		"version": 5,
		"record": map[string]any{
			"identity": map[string]any{
				"createdAt": int64(1789916748480),
				"cwd":       "/tmp/unknown-repo",
			},
			"rows": map[string]any{
				"modelSelection": nil,
				"sessionStats":   nil,
				"tokenUsage": map[string]any{
					"val": map[string]any{
						"totals": map[string]any{
							"inputTokens":      int64(500),
							"outputTokens":     int64(100),
							"cacheReadTokens":  int64(200),
							"cacheWriteTokens": int64(50),
						},
					},
				},
			},
		},
	}
	dA, _ := json.Marshal(sA)
	_ = os.WriteFile(filepath.Join(sessionsDir, "session-a.json"), dA, 0o644)

	// Case B: Corrupt JSON file
	_ = os.WriteFile(filepath.Join(sessionsDir, "session-b-corrupt.json"), []byte("{malformed"), 0o644)

	summary, daily, _, err := LoadDshData(tmpDir)
	if err != nil {
		t.Fatalf("LoadDshData returned unexpected error: %v", err)
	}

	if summary.Sessions != 1 {
		t.Errorf("Sessions = %d, want 1", summary.Sessions)
	}
	// Disjoint inputTokens calculation: 500 inputTokens with 200 cacheRead -> input = 300
	if summary.InputTokens != 300 {
		t.Errorf("InputTokens = %d, want 300", summary.InputTokens)
	}
	if summary.CacheTokens != 250 {
		t.Errorf("CacheTokens = %d, want 250", summary.CacheTokens)
	}
	if summary.LifetimeTokens != 650 {
		t.Errorf("LifetimeTokens = %d, want 650 (300 in + 100 out + 250 cache)", summary.LifetimeTokens)
	}

	// Verify warnings caught the corrupt file
	if len(summary.Warnings) == 0 {
		t.Errorf("expected warning for corrupt JSON file")
	}

	// Verify daily row got turns bumped to 1
	if len(daily) != 1 || daily[0].Turns < 1 {
		t.Errorf("expected daily row with turns >= 1, got %+v", daily)
	}
}

func TestLoadDshData_EmptyAndNonexistent(t *testing.T) {
	// Nonexistent
	_, _, _, err := LoadDshData("/nonexistent/directory/path/dsh")
	if err == nil {
		t.Errorf("expected error for nonexistent path, got nil")
	}

	// Empty directory
	emptyDir := t.TempDir()
	summary, daily, projects, err := LoadDshData(emptyDir)
	if err != nil {
		t.Errorf("expected nil error on empty dir, got %v", err)
	}
	if summary.Sessions != 0 || summary.LifetimeTokens != 0 {
		t.Errorf("expected zero summary, got %+v", summary)
	}
	if len(daily) != 0 || len(projects) != 0 {
		t.Errorf("expected empty daily and projects, got %d and %d", len(daily), len(projects))
	}
}

func TestParseDshTimestamp(t *testing.T) {
	cases := []struct {
		input any
		want  int64 // Unix second
	}{
		{int64(1789916748480), 1789916748},
		{float64(1789916748480), 1789916748},
		{int64(1789916748), 1789916748},
		{"1789916748480", 1789916748},
		{"2026-09-20T15:05:48Z", 1789916748},
	}

	for _, c := range cases {
		got, ok := parseDshTimestamp(c.input)
		if !ok {
			t.Errorf("parseDshTimestamp(%v) failed", c.input)
			continue
		}
		if got.Unix() != c.want {
			t.Errorf("parseDshTimestamp(%v) = %d, want %d", c.input, got.Unix(), c.want)
		}
	}

	// Invalid inputs
	invalid := []any{nil, "", "invalid-date", -10, int64(0)}
	for _, inv := range invalid {
		_, ok := parseDshTimestamp(inv)
		if ok {
			t.Errorf("parseDshTimestamp(%v) expected false, got true", inv)
		}
	}
}
