# Task: Clamp project activity deltas to tool-level totals (live-tracker-delta-clamping)

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_live-tracker-delta-clamping_1792417591`


## Context
- Goal: Ensure individual project token and turn deltas in `LiveTracker.Poll` cannot exceed tool-level `deltaTok` and `deltaTurns`.
- Why: Baseline desynchronization or race conditions can cause project diffs to exceed tool-level deltas; without clamping, massive project numbers leak into `RecentEvents`.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: `live-tracker-phantom-bursts` (Seq #92)
- Readiness: Blocked on `live-tracker-phantom-bursts`

## Files to Touch
- internal/thermal/live.go
- internal/thermal/live_test.go
- docs/MIGRATION.md

## Scope & File Details
- Edit: `internal/thermal/live.go` in `LiveTracker.Poll` (lines 648-678)
- Create: `internal/thermal/live_test.go` (delta ceiling enforcement tests)
- Edit: `docs/MIGRATION.md` (aggregation safety documentation)
- Read: `internal/thermal/live.go:648-678` (existing project delta calculations)

## Requirements
1. In `internal/thermal/live.go:648-678`, project activity deltas `pDTurns` and `pDTok` must be clamped to tool-level `deltaTurns` and `deltaTok` unconditionally, whether or not the project baseline was previously seen (`tokSeen` and `turnSeen`).
2. Total project turn delta (`totalProjectTurnDelta`) and total project token delta (`totalProjectTokDelta`) cannot exceed tool `deltaTurns` and `deltaTok`.
3. Snapshot invariant: the sum of turn deltas across all project `LiveEvent` entries in a single poll must never exceed the tool's `deltaTurns`.

## Interfaces / signatures (fixed)
```go
func (lt *LiveTracker) Poll(results []ToolResult, projDays []ProjectDay, now time.Time) (LiveSnapshot, []LiveEvent)

type projectDeltaInfo struct {
    project    string
    slug       string
    deltaTok   int64
    deltaTurns int
}
```

## Constraints and non-goals
- No new external dependencies.
- Sub-10ms poll execution; zero allocation on the polling fast path.
- Do not alter `--json` output structure.
- Do not suppress tool-level activity when project data is missing.

## Example
- Tool `Agy` reports +2 turns (`deltaTurns = 2`).
- Project telemetry produces an anomalous diff of 10,000 turns in project `lab`.
- Clamping enforces `pDTurns = min(10000, 2) = 2`.
- `RecentEvents` emits `LiveEvent{Turns: 2, Project: "lab"}`.
- Boundary failure prevented: An anomalous project delta of 10,000 is rejected and clamped to 2.

## Acceptance criteria
- AC-001: `go test -race -v -run TestLiveTracker_ProjectDeltaClamping ./internal/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that anomalous project increments are clamped to `deltaTurns` and `deltaTok`.
- AC-002: `go test -race -v -run TestLiveTracker_ProjectDeltaExceedsToolRejected ./internal/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that when a tool reports `deltaTurns = 0`, no project events are emitted even if raw project data changes (negative case).

## If blocked
Stop affected work; report evidence and the decision needed. Do not guess.

## Return
Return status, summary, changed files, case results, commands/exits, evidence, candidate, blocker.
