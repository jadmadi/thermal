# Task: Fix live tracker idle baseline collapse (live-tracker-phantom-bursts)

## Context
- Goal: Prevent `LiveTracker.Poll` from clobbering cumulative project baselines to single-day slices on idle ticks.
- Why: Idle ticks overwrite cumulative project baselines with today-only counts; subsequent single-step updates emit entire lifetime project totals as phantom bursts.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: None (Independent, Seq #90)
- Readiness: Ready to claim

## Files to Touch
- internal/thermal/live.go
- internal/thermal/live_test.go
- docs/MIGRATION.md

## Scope & File Details
- Edit: `internal/thermal/live.go` in `LiveTracker.Poll` (lines 806-813)
- Create: `internal/thermal/live_test.go` (new regression cases)
- Edit: `docs/MIGRATION.md` (aggregation bugfix entry)
- Read: `internal/thermal/live.go:616-635` (canonical cumulative project accumulation)

## Requirements
1. In `internal/thermal/live.go:806-813`, when `deltaTok == 0 && deltaTurns == 0`, project baselines in `lt.prevProjTokens` and `lt.prevProjTurns` must accumulate across all days matching the delta path, rather than filtering on `pd.Day == lt.todayDay`.
2. Sequential polling with zero tool deltas must preserve cumulative project baselines.
3. Subsequent tool increments (e.g. +1 turn) must emit only the incremental delta in `RecentEvents` rather than replaying historical lifetime project totals.
4. Snapshot invariant: the sum of turn deltas across `RecentEvents` within an observation window must strictly equal the session turn delta (`lt.sessionTurns`).

## Interfaces / signatures (fixed)
```go
func (lt *LiveTracker) Poll(results []ToolResult, projDays []ProjectDay, now time.Time) (LiveSnapshot, []LiveEvent)

type LiveEvent struct {
    Timestamp   time.Time `json:"timestamp"`
    Tool        string    `json:"tool"`
    Tokens      int64     `json:"tokens"`
    Turns       int       `json:"turns,omitempty"`
    Model       string    `json:"model,omitempty"`
    Project     string    `json:"project,omitempty"`
    Cost        float64   `json:"cost"`
    IsEstimated bool      `json:"isEstimated,omitempty"`
}

type LiveSnapshot struct {
    TodayTokens       int64            `json:"todayTokens"`
    TodayTurns        int              `json:"todayTurns"`
    SessionTokens     int64            `json:"sessionTokens"`
    SessionTurns      int              `json:"sessionTurns"`
    RecentEvents      []LiveEvent      `json:"recentEvents"`
    ToolTotals        map[string]int64 `json:"toolTotals"`
}
```

## Constraints and non-goals
- No new external dependencies.
- Maintain sub-10ms poll time; preserve zero-allocation execution on idle ticks.
- Do not change `--json` schema keys or types.
- Do not alter tool-level delta calculations in lines 524-546.

## Example
- Initial poll: Tool `Agy` has 125,600 turns in project `lab` across historical days. Baseline initialized to 125,600.
- Poll 2 (idle): No new turns (`deltaTurns == 0`). Baseline remains 125,600. `RecentEvents` emits 0 events.
- Poll 3 (+1 turn): Tool `Agy` reports 125,601 turns (`deltaTurns == 1`). Emits exactly 1 event: `LiveEvent{Tool: "Agy", Turns: 1, Project: "lab"}`.
- Boundary failure prevented: Historical 125.6K turns are not re-emitted on Poll 3.

## Acceptance criteria
- AC-001: `go test -race -v -run TestLiveTracker_IdlePollBaselinePreservation ./internal/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that sequential idle polls followed by a +1 turn update emit exactly one `LiveEvent` with `Turns: 1`, and `sum(RecentEvents.Turns) == SessionTurns`.
- AC-002: `go test -race -v -run TestLiveTracker_IdlePollNoDuplicateBursts ./internal/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that calling `Poll` with zero delta produces empty `[]LiveEvent` (negative case) and does not clobber `prevProjTurns` with single-day totals.

## If blocked
Stop affected work; report evidence and the decision needed. Do not guess.

## Return
Return status, summary, changed files, case results, commands/exits, evidence, candidate, blocker.
