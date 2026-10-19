# Task: Preserve cumulative project baselines across midnight transitions (live-tracker-midnight-drift)

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_live-tracker-midnight-drift_1792417587`


## Context
- Goal: Prevent midnight day boundary transitions in `LiveTracker.Poll` from resetting cumulative project baselines to current-day only counts.
- Why: Day boundary crossing wipes `lt.prevProjTokens` and `lt.prevProjTurns` to `pd.Day == currentDay`; subsequent turns after midnight re-emit pre-midnight lifetime project totals as new bursts.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: `live-tracker-phantom-bursts` (Seq #91)
- Readiness: Blocked on `live-tracker-phantom-bursts`

## Files to Touch
- internal/thermal/live.go
- internal/thermal/live_test.go
- docs/MIGRATION.md

## Scope & File Details
- Edit: `internal/thermal/live.go` in `LiveTracker.Poll` (lines 153-171)
- Create: `internal/thermal/live_test.go` (midnight transition tests)
- Edit: `docs/MIGRATION.md` (aggregation bugfix entry)
- Read: `internal/thermal/live.go:153-171` (existing day boundary logic)

## Requirements
1. In `internal/thermal/live.go:153-171`, when a midnight boundary is crossed (`currentDay != lt.todayDay`), `lt.prevProjTokens` and `lt.prevProjTurns` must retain cumulative counts across all days rather than filtering on `pd.Day == currentDay`.
2. The active project and active model pointers may reset on midnight, but project token and turn baselines must remain cumulative.
3. Activity occurring immediately after midnight must emit only the incremental post-midnight activity in `RecentEvents`.

## Interfaces / signatures (fixed)
```go
func (lt *LiveTracker) Poll(results []ToolResult, projDays []ProjectDay, now time.Time) (LiveSnapshot, []LiveEvent)

type ProjectDay struct {
    Project string `json:"project"`
    Day     string `json:"day"`
    Tool    string `json:"tool"`
    Tokens  int64  `json:"tokens"`
    Turns   int    `json:"turns,omitempty"`
}
```

## Constraints and non-goals
- No new dependencies.
- Sub-10ms poll time; zero allocations on day boundaries.
- Do not change `--json` schema keys.
- Do not alter `todayTokens` and `todayTurns` resetting at midnight.

## Example
- T1 (23:59:59): Day 2026-10-19 has 1,000 turns in project `lab`. Cumulative baseline is 1,000.
- T2 (00:00:01): Day 2026-10-20 begins. Cumulative baseline remains 1,000.
- T3 (00:00:05): +2 turns logged on Day 2026-10-20. Total is 1,002. `RecentEvents` emits exactly 1 event: `LiveEvent{Turns: 2, Project: "lab"}`.
- Boundary failure prevented: Day transition does not collapse baseline to 0, which would have emitted a phantom +1,002 turns.

## Acceptance criteria
- AC-001: `go test -race -v -run TestLiveTracker_MidnightTransitionBaseline ./internal/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that crossing midnight preserves lifetime project baselines and emits only post-midnight turn deltas.
- AC-002: `go test -race -v -run TestLiveTracker_MidnightTransitionNoPhantomReplay ./internal/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that crossing midnight with no new activity produces zero `LiveEvent` entries (negative case) and does not reset `lt.prevProjTurns` to zero.

## If blocked
Stop affected work; report evidence and the decision needed. Do not guess.

## Return
Return status, summary, changed files, case results, commands/exits, evidence, candidate, blocker.
