# Thermal (`agy`) Multi-Agent Implementation Plan

This document is a comprehensive, production-grade engineering plan tailored for autonomous AI agents (or a human-agent team) to implement all 10 core architectural, reliability, and performance improvements across `thermal` (`github.com/jadmadi/thermal`).

---

## Orchestration & Concurrency Guidelines for AI Agents

To allow **multiple AI agents to work simultaneously without Git merge conflicts or overwritten files**, the entire improvement suite is divided into **strictly disjoint work packages across 3 execution phases**.

### Key Rules for Implementing Agents
1. **Strict File Ownership**: Each task explicitly lists **Owned Files**. An implementing agent assigned to a task **MUST NOT modify any file outside its owned files list**.
2. **Zero-Conflict File Extraction**: Where legacy files previously mixed multiple responsibilities (e.g., `jsonl.go` housing Codex, Agy, and command-code loaders), the tasks instruct agents to extract those loaders into clean, dedicated files (`codex.go`, `agy.go`, `commandcode.go`). This eliminates concurrent edit collisions.
3. **Verification Integrity**: Before marking any task complete, the agent **must compile the full repository** (`go build ./...`) and run the package-specific test command with the race detector enabled (`go test -v -race <pkg>`).
4. **Preserve Existing Comments & Docstrings**: Never strip unrelated comments or documentation during edits.

---

## Disjoint Task Matrix

| Phase | Task ID | Task Title | Owned Files (Strict Boundaries) | Concurrency Group | Exact Verification Command |
| :---: | :---: | :--- | :--- | :---: | :--- |
| **1** | **Task 1.1** | Domain Logic & Streak DST Fixes | `internal/thermal/streak.go`<br>`internal/thermal/streak_test.go` | **Phase 1 (Concurrent)** | `go test -v -race ./internal/thermal` |
| **1** | **Task 1.2** | Formatting & Time Utilities Tests | `internal/thermal/format_test.go`<br>`internal/thermal/time_test.go` | **Phase 1 (Concurrent)** | `go test -v -race ./internal/thermal` |
| **1** | **Task 1.3** | Concurrency-Safe Progress Bar | `internal/render/progress.go`<br>`internal/render/progress_test.go` | **Phase 1 (Concurrent)** | `go test -v -race ./internal/render` |
| **1** | **Task 1.4** | Heatmap & Dashboard Render Tests | `internal/render/heatmap_test.go`<br>`internal/render/dashboard_test.go`<br>`internal/render/leaderboard_test.go` | **Phase 1 (Concurrent)** | `go test -v -race ./internal/render` |
| **2** | **Task 2.1** | Codex Precision & Worker Pool | `internal/loaders/codex.go` *(NEW)*<br>`internal/loaders/codex_test.go` *(NEW)* | **Phase 2 (Concurrent)** | `go test -v -race ./internal/loaders -run TestCodex` |
| **2** | **Task 2.2** | Agy & CommandCode Parallel Scanning | `internal/loaders/agy.go` *(NEW)*<br>`internal/loaders/commandcode.go` *(NEW)*<br>`internal/loaders/agy_test.go` *(NEW)*<br>`internal/loaders/commandcode_test.go` *(NEW)* | **Phase 2 (Concurrent)** | `go test -v -race ./internal/loaders -run "TestAgy\|TestCommandCode"` |
| **2** | **Task 2.3** | SQLite & JSONL Loader Tests & Cleanup | `internal/loaders/jsonl.go`<br>`internal/loaders/sqlite.go`<br>`internal/loaders/sqlite_test.go` *(NEW)* | **Phase 2 (Concurrent)** | `go test -v -race ./internal/loaders` |
| **3** | **Task 3.1** | Upgrader Security Hardening | `cmd/thermal/upgrade.go`<br>`cmd/thermal/upgrade_test.go` *(NEW)* | **Phase 3 (Concurrent)** | `go test -v -race ./cmd/thermal -run TestUpgrade` |
| **3** | **Task 3.2** | CLI Diagnostics & Registry Abstraction | `cmd/thermal/main.go`<br>`internal/thermal/types.go`<br>`internal/loaders/registry.go`<br>`internal/loaders/registry_test.go` *(NEW)* | **Phase 3 (Concurrent)** | `go test -v -race ./cmd/thermal ./internal/loaders` |

---

## Detailed Task Specifications

---

### Phase 1: Foundation & Pure Domain Logic (Fully Concurrent)

#### Task 1.1: Domain Logic & Streak DST Fixes
- **Assigned Agent Role**: Domain Logic & Algorithm Specialist
- **Owned Files**: `internal/thermal/streak.go`, `[NEW] internal/thermal/streak_test.go`
- **Objective**: Fix bug where streak duration checks (`t.Sub(prev) == 24*time.Hour`) and day subtraction (`cursor.Add(-24 * time.Hour)`) fail across Daylight Saving Time (DST) transitions. Also fix quartile threshold generation so duplicate low values don't skip color levels `2` and `3`.

##### Detailed Implementation Instructions for `internal/thermal/streak.go`
1. Locate `ComputeStreaks(days map[string]bool) (current int, longest int)`.
2. Replace the 24-hour subtraction check inside the `sorted` loop with calendar date equality (`prev.AddDate(0, 0, 1).Equal(t)`):
```go
	longest = 0
	run := 0
	var prev time.Time
	for _, d := range sorted {
		t, _ := time.Parse("2006-01-02", d)
		if !prev.IsZero() && prev.AddDate(0, 0, 1).Equal(t) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
		prev = t
	}
```
3. Replace the `cursor = cursor.Add(-24 * time.Hour)` decrement inside the current streak loop with calendar date arithmetic (`cursor = cursor.AddDate(0, 0, -1)`):
```go
	current = 0
	if len(sorted) > 0 {
		cursor, _ := time.Parse("2006-01-02", sorted[len(sorted)-1])
		for {
			if !days[LocalDay(cursor)] {
				break
			}
			current++
			cursor = cursor.AddDate(0, 0, -1)
		}
	}
```
4. Locate `ActivityThresholds(activity map[string]DayActivity) [3]int64`. Modify the return logic to ensure strictly ascending distinct thresholds when `nonzero` values allow it:
```go
	at := func(q float64) int64 {
		idx := int(math.Floor(float64(len(nonzero)-1) * q))
		return nonzero[idx]
	}
	t1, t2, t3 := at(0.25), at(0.5), at(0.75)
	if t2 <= t1 && t1 < nonzero[len(nonzero)-1] {
		t2 = t1 + 1
	}
	if t3 <= t2 && t2 < nonzero[len(nonzero)-1] {
		t3 = t2 + 1
	}
	return [3]int64{t1, t2, t3}
```

##### Detailed Instructions for `internal/thermal/streak_test.go`
Create `internal/thermal/streak_test.go` with exhaustive tests:
```go
package thermal

import (
	"testing"
)

func TestComputeStreaks_Empty(t *testing.T) {
	cur, long := ComputeStreaks(map[string]bool{})
	if cur != 0 || long != 0 {
		t.Fatalf("expected (0,0), got (%d,%d)", cur, long)
	}
}

func TestComputeStreaks_SingleDay(t *testing.T) {
	cur, long := ComputeStreaks(map[string]bool{"2026-03-15": true})
	if cur != 1 || long != 1 {
		t.Fatalf("expected (1,1), got (%d,%d)", cur, long)
	}
}

func TestComputeStreaks_ConsecutiveAcrossDST(t *testing.T) {
	// Test days spanning US/Europe spring forward and fall back transitions
	days := map[string]bool{
		"2026-03-07": true,
		"2026-03-08": true, // Spring forward
		"2026-03-09": true,
		"2026-10-31": true,
		"2026-11-01": true, // Fall back
		"2026-11-02": true,
	}
	cur, long := ComputeStreaks(days)
	if long != 3 {
		t.Fatalf("expected longest streak 3 across DST transition, got %d", long)
	}
	_ = cur
}

func TestActivityThresholds_DistinctEnforcement(t *testing.T) {
	// 4 days with 10 tokens, 1 day with 1000 tokens
	act := map[string]DayActivity{
		"2026-01-01": {Tokens: 10},
		"2026-01-02": {Tokens: 10},
		"2026-01-03": {Tokens: 10},
		"2026-01-04": {Tokens: 10},
		"2026-01-05": {Tokens: 1000},
	}
	thresh := ActivityThresholds(act)
	if thresh[1] <= thresh[0] || thresh[2] <= thresh[1] {
		t.Fatalf("expected strictly ascending thresholds, got %v", thresh)
	}
}

func TestActivityLevel(t *testing.T) {
	thresh := [3]int64{100, 500, 1000}
	tests := []struct {
		tokens int64
		want   int
	}{
		{0, 1},
		{50, 1},
		{100, 1},
		{101, 2},
		{500, 2},
		{501, 3},
		{1000, 3},
		{1001, 4},
	}
	for _, tc := range tests {
		got := ActivityLevel(DayActivity{Tokens: tc.tokens}, thresh)
		if got != tc.want {
			t.Errorf("ActivityLevel(%d, %v) = %d, want %d", tc.tokens, thresh, got, tc.want)
		}
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./internal/thermal` -> All tests MUST PASS with 0 race conditions.

---

#### Task 1.2: Formatting & Time Utilities Tests
- **Assigned Agent Role**: Unit Test & Utilities Engineer
- **Owned Files**: `[NEW] internal/thermal/format_test.go`, `[NEW] internal/thermal/time_test.go`
- **Objective**: Create 100% test coverage for all string formatting (`CompactNumber`, padding, path shortening) and time utility functions (`UnixDay`, `LocalDay`, `StartOfToday`).

##### Detailed Instructions for `internal/thermal/format_test.go`
Create `internal/thermal/format_test.go`:
```go
package thermal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompactNumber(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1540, "1.5K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{44000000, "44.0M"},
		{1000000000, "1.0B"},
		{1850000000, "1.9B"},
	}
	for _, tc := range tests {
		if got := CompactNumber(tc.input); got != tc.want {
			t.Errorf("CompactNumber(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestPadding(t *testing.T) {
	if got := PadRight("abc", 5); got != "abc  " {
		t.Errorf("PadRight = %q, want %q", got, "abc  ")
	}
	if got := PadRight("abcde", 3); got != "abcde" {
		t.Errorf("PadRight truncation = %q, want %q", got, "abcde")
	}
	if got := PadLeft("abc", 5); got != "  abc" {
		t.Errorf("PadLeft = %q, want %q", got, "  abc")
	}
}

func TestFormatPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	sub := filepath.Join(home, ".local", "share", "opencode")
	got := FormatPath(sub)
	if got != "~/.local/share/opencode" && got != "~\\.local\\share\\opencode" {
		t.Errorf("FormatPath(%q) = %q", sub, got)
	}
}
```

##### Detailed Instructions for `internal/thermal/time_test.go`
Create `internal/thermal/time_test.go`:
```go
package thermal

import (
	"testing"
	"time"
)

func TestUnixDay(t *testing.T) {
	// 1710504000 is 2024-03-15 12:00:00 UTC
	ts := int64(1710504000)
	got := UnixDay(ts)
	want := time.Unix(ts, 0).Format("2006-01-02")
	if got != want {
		t.Errorf("UnixDay(%d) = %q, want %q", ts, got, want)
	}
}

func TestStartOfToday(t *testing.T) {
	today := StartOfToday()
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 || today.Nanosecond() != 0 {
		t.Errorf("StartOfToday not at midnight: %v", today)
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./internal/thermal` -> All tests MUST PASS.

---

#### Task 1.3: Concurrency-Safe Progress Bar
- **Assigned Agent Role**: Concurrency & CLI Systems Specialist
- **Owned Files**: `internal/render/progress.go`, `[NEW] internal/render/progress_test.go`
- **Objective**: Fix race condition on `p.started` and double-close panic on `p.done` when `Done()` or `Start()` are called repeatedly across error recovery or concurrent goroutines.

##### Detailed Implementation Instructions for `internal/render/progress.go`
1. Open `internal/render/progress.go`. Add `"sync"` to imports and replace `started bool` with `started atomic.Bool` and `stopOnce sync.Once`:
```go
type Progress struct {
	total    int64
	current  atomic.Int64
	label    string
	done     chan struct{}
	started  atomic.Bool
	stopOnce sync.Once
}
```
2. Update `Start()` to use atomic compare-and-swap so the goroutine is spawned exactly once when on a terminal:
```go
func (p *Progress) Start() {
	if !stderrIsTerminal() || !p.started.CompareAndSwap(false, true) {
		return
	}
	go p.render()
}
```
3. Update `Done()` to check `p.started.Load()` and wrap channel closing in `p.stopOnce.Do`:
```go
func (p *Progress) Done() {
	if !p.started.Load() {
		return
	}
	p.stopOnce.Do(func() {
		close(p.done)
		time.Sleep(60 * time.Millisecond)
		fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", 72)+"\r")
	})
}
```

##### Detailed Instructions for `internal/render/progress_test.go`
Create `internal/render/progress_test.go` to test concurrent operations and double calls without panics or race conditions:
```go
package render

import (
	"sync"
	"testing"
)

func TestProgress_ConcurrencySafe(t *testing.T) {
	p := NewProgress("Test", 1000)
	// Even if not TTY, calling Increment, Start, and Done concurrently must not race or panic
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Start()
			p.Increment(100)
			p.Done()
			p.Done() // Double call must not panic
		}()
	}
	wg.Wait()
	if p.current.Load() != 1000 {
		t.Errorf("expected current 1000, got %d", p.current.Load())
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./internal/render` -> All tests MUST PASS with zero race warnings.

---

#### Task 1.4: Heatmap & Dashboard Render Tests
- **Assigned Agent Role**: UI/Terminal Rendering Engineer
- **Owned Files**: `[NEW] internal/render/heatmap_test.go`, `[NEW] internal/render/dashboard_test.go`, `[NEW] internal/render/leaderboard_test.go`
- **Objective**: Create unit tests verifying the layout lines, ANSI color codes, and medal/streak rendering across `heatmap.go`, `dashboard.go`, and `leaderboard.go`.

##### Detailed Instructions for `internal/render/heatmap_test.go`
Create `internal/render/heatmap_test.go`:
```go
package render

import (
	"strings"
	"testing"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderHeatmap_Dimensions(t *testing.T) {
	act := map[string]thermal.DayActivity{
		"2026-07-15": {Tokens: 1500, Turns: 5},
	}
	lines := RenderHeatmap(act, 52, false)
	// Must have 1 header line + 7 weekday lines + 1 legend line = 9 total lines
	if len(lines) != 9 {
		t.Fatalf("expected 9 lines in heatmap output, got %d", len(lines))
	}
	if !strings.Contains(lines[8], "Less") || !strings.Contains(lines[8], "More") {
		t.Errorf("legend line missing Less/More: %s", lines[8])
	}
}
```

##### Detailed Instructions for `internal/render/leaderboard_test.go` and `dashboard_test.go`
Create `internal/render/leaderboard_test.go` and `internal/render/dashboard_test.go` verifying that `RenderLeaderboard` and `RenderDashboard` correctly format token numbers and fire emojis (`🔥` for streaks >= 7):
```go
package render

import (
	"strings"
	"testing"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestRenderLeaderboard_Output(t *testing.T) {
	results := []thermal.ToolResult{
		{
			Tool: thermal.ToolOpenCode, Name: "OpenCode", CurrentStreak: 10, LongestStreak: 15,
			ActiveDays: 30, TotalActivity: 1500000, Summary: thermal.Summary{Cost: 12.50},
		},
	}
	out := RenderLeaderboard(results, 52, true)
	if !strings.Contains(out, "Token Warriors") {
		t.Errorf("expected Token Warriors section in output")
	}
	if !strings.Contains(out, "OpenCode") {
		t.Errorf("expected OpenCode in output")
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./internal/render` -> All tests MUST PASS.

---

### Phase 2: Disjoint Loaders Refactoring & Testing (Fully Concurrent)

#### Task 2.1: Codex Loader Precision Reconciliation & Worker Pool
- **Assigned Agent Role**: Backend Loaders & Concurrency Engineer
- **Owned Files**: `internal/loaders/codex.go` *(NEW)*, `[NEW] internal/loaders/codex_test.go`
- **Objective**: Extract all Codex loading logic out of `jsonl.go` into a new, independent file `internal/loaders/codex.go`. Fix integer truncation rounding errors when scaling token breakdown, and introduce a bounded worker pool (`sync.WaitGroup` with a channel of 8 concurrent workers) to scan rollout JSONL files in parallel.

##### Detailed Implementation Instructions for `internal/loaders/codex.go`
1. Create `internal/loaders/codex.go`. Move `LoadCodexData`, `loadCodexFromStateDB`, `tokenBreakdown`, and `readLastTokenBreakdown` out of `jsonl.go` into `codex.go`.
2. *Note: To ensure zero conflicts between Task 2.1, 2.2, and 2.3, **Task 2.1 is permitted to delete ONLY the Codex functions from `jsonl.go`**, or Task 2.3 will do the final `jsonl.go` cleanup after Phase 2.*
3. In `loadCodexFromStateDB` inside `codex.go`, replace the sequential `for _, t := range threads` rollout scan with a bounded worker pool:
```go
	// Worker pool for parallel rollout file scanning
	breakdowns := make([]*tokenBreakdown, len(threads))
	var wg sync.WaitGroup
	workerLimit := make(chan struct{}, 8) // max 8 concurrent workers

	for i, t := range threads {
		if t.rolloutPath == "" || t.tokensUsed == 0 {
			continue
		}
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			workerLimit <- struct{}{}
			defer func() { <-workerLimit }()
			breakdowns[idx] = readLastTokenBreakdown(path)
		}(i, t.rolloutPath)
	}
	wg.Wait()

	for _, b := range breakdowns {
		if b == nil {
			continue
		}
		summary.InputTokens += b.input
		summary.OutputTokens += b.output
		summary.ReasoningTokens += b.reasoning
		summary.CacheTokens += b.cache
	}
```
4. In `loadCodexFromStateDB`, update the ratio adjustment so exact equality `Input + Output + Reasoning + Cache == LifetimeTokens` is guaranteed without 1-token truncation gaps:
```go
	breakdownTotal := summary.InputTokens + summary.OutputTokens +
		summary.ReasoningTokens + summary.CacheTokens
	if breakdownTotal > 0 && breakdownTotal != summary.LifetimeTokens {
		ratio := float64(summary.LifetimeTokens) / float64(breakdownTotal)
		summary.InputTokens = int64(float64(summary.InputTokens) * ratio)
		summary.OutputTokens = int64(float64(summary.OutputTokens) * ratio)
		summary.ReasoningTokens = int64(float64(summary.ReasoningTokens) * ratio)
		summary.CacheTokens = int64(float64(summary.CacheTokens) * ratio)
		
		// Reconcile rounding remainder directly onto InputTokens
		remainder := summary.LifetimeTokens - (summary.InputTokens + summary.OutputTokens + summary.ReasoningTokens + summary.CacheTokens)
		summary.InputTokens += remainder
	} else if breakdownTotal == 0 {
		summary.InputTokens = summary.LifetimeTokens
	}
```

##### Detailed Instructions for `internal/loaders/codex_test.go`
Create `internal/loaders/codex_test.go` creating a temporary directory with a synthetic `state_5.sqlite` database and a `.jsonl` rollout file:
```go
package loaders

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	_ "modernc.org/sqlite"
)

func TestLoadCodexData_ReconciliationAndRollouts(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state_5.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE threads (
			id TEXT, tokens_used INTEGER, model TEXT, source TEXT,
			reasoning_effort TEXT, agent_role TEXT, created_at INTEGER,
			updated_at INTEGER, rollout_path TEXT, archived INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}

	rolloutPath := filepath.Join(dir, "rollout.jsonl")
	rolloutContent := `{"type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":60,"output_tokens":30,"reasoning_output_tokens":0,"cached_input_tokens":0}}}}` + "\n"
	if err := os.WriteFile(rolloutPath, []byte(rolloutContent), 0644); err != nil {
		t.Fatalf("failed writing rollout: %v", err)
	}

	// Insert row with 100 tokens used (while rollout has 90 sum -> requires exact scaling and remainder check)
	_, err = db.Exec(`INSERT INTO threads VALUES ('t1', 100, 'gpt-4o', 'cli', '', '', 1710504000, 1710504060, ?, 0)`, rolloutPath)
	if err != nil {
		t.Fatalf("failed inserting thread: %v", err)
	}
	db.Close()

	sum, daily, err := LoadCodexData(dir)
	if err != nil {
		t.Fatalf("LoadCodexData error: %v", err)
	}
	if sum.LifetimeTokens != 100 {
		t.Errorf("expected lifetime 100, got %d", sum.LifetimeTokens)
	}
	totalParts := sum.InputTokens + sum.OutputTokens + sum.ReasoningTokens + sum.CacheTokens
	if totalParts != sum.LifetimeTokens {
		t.Errorf("breakdown sum %d != lifetime %d", totalParts, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}
```

##### Verification & Done Criteria
- Remove `LoadCodexData`, `loadCodexFromStateDB`, `tokenBreakdown`, and `readLastTokenBreakdown` from `jsonl.go` (leaving them only in `codex.go`).
- Run `go test -v -race ./internal/loaders -run TestLoadCodex` -> MUST PASS.

---

#### Task 2.2: Agy & CommandCode Parallel Scanning & Robust Model Extraction
- **Assigned Agent Role**: Transcripts & File Systems Specialist
- **Owned Files**: `internal/loaders/agy.go` *(NEW)*, `internal/loaders/commandcode.go` *(NEW)*, `[NEW] internal/loaders/agy_test.go`, `[NEW] internal/loaders/commandcode_test.go`
- **Objective**: Extract `LoadAgyData` and `LoadCommandCodeData` out of `jsonl.go` into new independent files `agy.go` and `commandcode.go`. Introduce bounded parallel worker pools (8 workers) when walking across multiple project directories or scanning brain session `overview.txt`/`transcript.jsonl` files. Replace fragile `strings.Index` loops with clean `strings.IndexAny(model, ".\r\n\"")` model extraction.

##### Detailed Implementation Instructions for `internal/loaders/agy.go` and `commandcode.go`
1. Create `internal/loaders/agy.go`. Move `LoadAgyData` from `jsonl.go` into `agy.go` and remove it from `jsonl.go`.
2. In `agy.go`, parallelize the per-session `overview.txt` and `transcript.jsonl` scanning using a bounded worker pool (`sync.WaitGroup`, channel concurrency limit of 8):
```go
	type sessionResult struct {
		steps      int
		firstTs    time.Time
		lastTs     time.Time
		dayCounts  map[string]int
		modelCounts map[string]int64
	}
	results := make(chan sessionResult, len(entries))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for _, entry := range entries {
		if !entry.IsDir() { continue }
		sessionDir := filepath.Join(brainDir, entry.Name())
		wg.Add(1)
		go func(sDir string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			// scan overview.txt and transcript.jsonl here and send to results channel
		}(sessionDir)
	}
```
3. In `agy.go` transcript parsing, replace the multi-delimiter loop with clean `strings.IndexAny`:
```go
	if toIdx := strings.Index(rest, " to "); toIdx >= 0 {
		model := strings.TrimSpace(rest[toIdx+4:])
		if idx := strings.IndexAny(model, ".\r\n\""); idx >= 0 {
			model = strings.TrimSpace(model[:idx])
		}
		if model != "" && model != "None" {
			res.modelCounts[model]++
		}
	}
```
4. Create `internal/loaders/commandcode.go`. Move `LoadCommandCodeData` from `jsonl.go` into `commandcode.go` and remove it from `jsonl.go`.
5. In `commandcode.go`, similarly wrap the `for _, path := range sessionFiles` loop with a bounded worker pool (`sync.WaitGroup` with 8 workers) and merge the `byDay`, `modelCounts`, and duration statistics cleanly under a mutex (`sync.Mutex`) or via result channels.

##### Detailed Instructions for `agy_test.go` and `commandcode_test.go`
Create `internal/loaders/agy_test.go` and `internal/loaders/commandcode_test.go` using `t.TempDir()` with mock `overview.txt`, `transcript.jsonl`, and `.commandcode/projects/dir/uuid.jsonl` files:
```go
package loaders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgyData_MockSession(t *testing.T) {
	dir := t.TempDir()
	brainDir := filepath.Join(dir, "brain", "sess-1", ".system_generated", "logs")
	if err := os.MkdirAll(brainDir, 0755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	overview := `{"created_at":"2026-07-18T10:00:00Z","source":"SYSTEM","type":"STEP","content":"test step"}` + "\n"
	if err := os.WriteFile(filepath.Join(brainDir, "overview.txt"), []byte(overview), 0644); err != nil {
		t.Fatalf("write overview error: %v", err)
	}
	transcript := `{"type":"SYSTEM","content":"Model Selection changed from None to Gemini 3.1 Pro (High)."}` + "\n"
	if err := os.WriteFile(filepath.Join(brainDir, "transcript.jsonl"), []byte(transcript), 0644); err != nil {
		t.Fatalf("write transcript error: %v", err)
	}

	sum, daily, err := LoadAgyData(dir)
	if err != nil {
		t.Fatalf("LoadAgyData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 1 {
		t.Errorf("expected 1 session and 1 token (step), got %d and %d", sum.Sessions, sum.LifetimeTokens)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
	if sum.ModelBreakdown["Gemini 3.1 Pro (High)"] != 1 {
		t.Errorf("expected model breakdown for Gemini 3.1 Pro (High), got %v", sum.ModelBreakdown)
	}
}
```

##### Verification & Done Criteria
- Remove `LoadAgyData` and `LoadCommandCodeData` from `jsonl.go`.
- Run `go test -v -race ./internal/loaders -run "TestLoadAgy\|TestLoadCommandCode"` -> MUST PASS.

---

#### Task 2.3: SQLite & JSONL Loader Tests & Cleanup
- **Assigned Agent Role**: Database Loaders & QA Specialist
- **Owned Files**: `internal/loaders/jsonl.go`, `internal/loaders/sqlite.go`, `[NEW] internal/loaders/sqlite_test.go`
- **Objective**: Clean up `jsonl.go` so it retains only `loadJsonlData` (or general helper methods). Create comprehensive unit tests (`sqlite_test.go`) for `LoadOpenCodeData`, `LoadMiMoCodeData`, and `LoadDevinData` using in-memory SQLite fixtures via `modernc.org/sqlite`.

##### Detailed Implementation Instructions
1. Verify `jsonl.go` only contains `loadJsonlData` (and any general JSON utilities) after Task 2.1 and 2.2 extracted their loaders.
2. In `sqlite.go`, verify that all SQL queries close rows (`defer rows.Close()`, `defer db.Close()`) and handle errors cleanly.
3. Create `internal/loaders/sqlite_test.go` creating temporary SQLite database files or memory fixtures (`sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))`) with synthetic tables for `session`/`message` (OpenCode/MiMo) and `sessions`/`message_nodes` (Devin):
```go
package loaders

import (
	"database/sql"
	"path/filepath"
	"testing"
	_ "modernc.org/sqlite"
)

func TestLoadOpenCodeData_PreAggregatedSchema(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql open error: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE session (
			id TEXT, time_created INTEGER, time_updated INTEGER,
			tokens_input INTEGER, tokens_output INTEGER, tokens_reasoning INTEGER,
			tokens_cache_read INTEGER, tokens_cache_write INTEGER,
			cost REAL, summary_additions INTEGER, summary_deletions INTEGER,
			summary_files INTEGER, agent TEXT
		);
		INSERT INTO session VALUES ('s1', 1710504000000, 1710504060000, 100, 50, 10, 5, 5, 0.05, 20, 5, 2, 'code');
	`)
	if err != nil {
		t.Fatalf("exec error: %v", err)
	}
	db.Close()

	sum, daily, err := LoadOpenCodeData(dbPath)
	if err != nil {
		t.Fatalf("LoadOpenCodeData error: %v", err)
	}
	if sum.Sessions != 1 || sum.LifetimeTokens != 170 {
		t.Errorf("expected sessions=1, lifetime=170, got %d, %d", sum.Sessions, sum.LifetimeTokens)
	}
	if sum.Cost != 0.05 {
		t.Errorf("expected cost=0.05, got %f", sum.Cost)
	}
	if len(daily) != 1 {
		t.Errorf("expected 1 daily row, got %d", len(daily))
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./internal/loaders` -> All loader tests (`Codex`, `Agy`, `CommandCode`, `OpenCode`, `Devin`) MUST PASS.

---

### Phase 3: CLI & Upgrader Hardening (Fully Concurrent)

#### Task 3.1: Upgrader Security Hardening & HTTP Client Timeout
- **Assigned Agent Role**: Security & Release Systems Engineer
- **Owned Files**: `cmd/thermal/upgrade.go`, `[NEW] cmd/thermal/upgrade_test.go`
- **Objective**: Configure explicit timeouts for all HTTP requests, prevent decompression bomb attacks (`io.LimitReader`), and remove dangerous fallback matching in binary asset selection.

##### Detailed Implementation Instructions for `cmd/thermal/upgrade.go`
1. Define a shared HTTP client with a 30-second timeout at package level:
```go
var httpClient = &http.Client{Timeout: 30 * time.Second}
```
2. Replace `http.Get(url)` calls (`runUpgrade`, `downloadFile`) with `http.NewRequestWithContext` using a 15-second or 30-second context timeout:
```go
func downloadFile(url, dest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	...
}
```
3. In `findAsset(assets []asset) (string, string, error)`, remove the risky OS-only fallback (`for _, a := range assets ... if strings.Contains(name, "_"+goos+"_")`) so `thermal` fails safely rather than installing a wrong binary architecture if exact CPU architecture match is missing:
```go
func findAsset(assets []asset) (string, string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
	}
	archName, ok := archMap[goarch]
	if !ok {
		archName = goarch
	}

	for _, a := range assets {
		name := strings.ToLower(a.Name)
		osMatch := strings.Contains(name, "_"+goos+"_") || strings.Contains(name, "-"+goos+"-")
		archMatch := strings.Contains(name, "_"+archName+".") || strings.Contains(name, "_"+archName+"_")

		if osMatch && archMatch {
			return a.Name, a.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("no binary found for %s/%s", goos, goarch)
}
```
4. In `extractBinary`, wrap `tr` with `io.LimitReader(tr, maxBinarySize)` (`250 MB` ceiling) before passing to `io.Copy`:
```go
	const maxBinarySize = 250 * 1024 * 1024 // 250 MB
	...
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, io.LimitReader(tr, maxBinarySize))
			return err
```

##### Detailed Instructions for `cmd/thermal/upgrade_test.go`
Create `cmd/thermal/upgrade_test.go` verifying asset selection logic across operating systems and CPU architectures:
```go
package main

import (
	"runtime"
	"testing"
)

func TestFindAsset_ExactMatching(t *testing.T) {
	assets := []asset{
		{Name: "thermal_0.3.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64"},
		{Name: "thermal_0.3.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_arm64"},
	}

	// This test verifies that whatever the current OS/Arch is, if it's in our mock assets, it matches exactly
	name, url, err := findAsset(assets)
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		if err != nil || name != "thermal_0.3.0_linux_amd64.tar.gz" {
			t.Errorf("failed matching linux_amd64: %v", err)
		}
		_ = url
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		if err != nil || name != "thermal_0.3.0_darwin_arm64.tar.gz" {
			t.Errorf("failed matching darwin_arm64: %v", err)
		}
	} else {
		// If on windows or arm linux, must return clean error without picking wrong architecture
		if err == nil {
			t.Errorf("expected error when no matching architecture exists, got %q", name)
		}
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./cmd/thermal -run TestFindAsset` -> MUST PASS.

---

#### Task 3.2: CLI Diagnostics & Registry Abstraction
- **Assigned Agent Role**: Core CLI & Architecture Lead
- **Owned Files**: `cmd/thermal/main.go`, `internal/thermal/types.go`, `internal/loaders/registry.go`, `[NEW] internal/loaders/registry_test.go`
- **Objective**: Add `--verbose` (`-v` / `-d`) flag support for non-fatal database loading diagnostics during `--tool all` leaderboard generation. Add `signal.NotifyContext` for clean `Ctrl+C` handling. Refactor `ToolInfo` registration to support modular extensibility.

##### Detailed Implementation Instructions
1. Open `internal/thermal/types.go`. Add `Verbose bool` to `Options`:
```go
type Options struct {
	Tool    string
	DBPath  string
	Weeks   int
	JSON    bool
	NoColor bool
	Verbose bool
}
```
2. Open `cmd/thermal/main.go`. In `parseArgs()`, parse `--verbose`, `-v`, and `-d`:
```go
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose warning diagnostics on stderr")
	flag.BoolVar(&opts.Verbose, "v", false, "Enable verbose warning diagnostics on stderr (shorthand)")
	flag.BoolVar(&opts.Verbose, "d", false, "Enable verbose warning diagnostics on stderr (shorthand)")
```
*Note: Make sure `-v` doesn't conflict with version subcommand checks (`case "version", "--version"`).*
3. In `main()`, set up signal handling context:
```go
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	_ = ctx
	opts := parseArgs()
```
4. In `main.go` inside the `opts.Tool == "all"` loop, print warning diagnostics to stderr if `opts.Verbose` is enabled and `loaders.LoadToolData` fails:
```go
			summary, daily, dataPath, err := loaders.LoadToolData(t, info, "")
			if err != nil {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "thermal: warning: failed loading %s: %v\n", info.Name, err)
				}
				continue
			}
```
5. Open `internal/loaders/registry.go`. Refactor `LoadToolData` to avoid hardcoded `if/else` checks for `dataPath` (`if t == thermal.ToolAgy ...`). Instead, add a `DataSubdir` or `PathResolver` field directly on `ToolInfo`:
```go
type ToolInfo struct {
	DBPath     string
	DataDir    string
	Name       string
	DataSubdir string // e.g. "history.jsonl", "brain", "sessions", "projects"
	Loader     func(string) (thermal.Summary, []thermal.DailyRow, error)
}
```
And in `LoadToolData`:
```go
		dataPath := filepath.Join(dir, info.DataSubdir)
		if info.DataSubdir == "" {
			dataPath = filepath.Join(dir, "history.jsonl")
		}
		return s, d, dataPath, err
```
6. Create `internal/loaders/registry_test.go` verifying alias resolution and registry structure:
```go
package loaders

import (
	"testing"
	"github.com/jadmadi/thermal/internal/thermal"
)

func TestResolveTool_Aliases(t *testing.T) {
	tests := []struct {
		alias string
		want  thermal.Tool
		ok    bool
	}{
		{"mimo", thermal.ToolMiMoCode, true},
		{"oc", thermal.ToolOpenCode, true},
		{"cmd", thermal.ToolCommandCode, true},
		{"whale", thermal.ToolCodewhale, true},
		{"nonexistent", "", false},
	}
	for _, tc := range tests {
		got, ok := ResolveTool(tc.alias)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ResolveTool(%q) = (%q, %v), want (%q, %v)", tc.alias, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAllTools_Config(t *testing.T) {
	tools := AllTools()
	if len(tools) == 0 {
		t.Fatalf("expected tools in registry")
	}
	for id, info := range tools {
		if info.Name == "" || info.Loader == nil {
			t.Errorf("tool %s has incomplete ToolInfo configuration", id)
		}
	}
}
```

##### Verification & Done Criteria
- Run `go test -v -race ./cmd/thermal ./internal/loaders` -> MUST PASS.
- Verify full repository build (`go build -o /dev/null ./cmd/thermal`) succeeds cleanly without warnings.

---

## Final Complete System Verification

Once all 3 Phases and 9 Tasks are completed by the implementing AI agents, run the final verification suite from the repository root:

```bash
# 1. Run all unit and mock integration tests across all packages with race detector enabled
go test -v -race -cover ./...

# 2. Verify compilation across primary targets
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /dev/null ./cmd/thermal
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /dev/null ./cmd/thermal

# 3. Verify that running `thermal --help` displays clean options including `--verbose`
go run ./cmd/thermal --help
```

If all tests pass with zero race warnings and clean build outputs, the implementation is complete and verified!
