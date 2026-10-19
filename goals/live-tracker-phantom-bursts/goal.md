# Goal: Fix live tracker idle baseline collapse (live-tracker-phantom-bursts)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Prevent `LiveTracker.Poll` from clobbering cumulative project token and turn baselines to single-day counts during idle (no-delta) poll ticks.

## Dependencies & Execution Order
- Mode: Independent (disjoint, parallelizable)
- Depends On: none
- Sequence: 90
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-tracker-phantom-bursts/facts.md`](facts.md)
- Execution Plan: [`goals/live-tracker-phantom-bursts/plan.md`](plan.md)

## Done Condition
1. In `internal/thermal/live.go:806-813`, when `deltaTok == 0 && deltaTurns == 0`, project baselines accumulate across all days matching the delta path instead of filtering on `pd.Day == lt.todayDay`.
2. Sequential polling with zero tool deltas preserves cumulative project baselines in `lt.prevProjTokens` and `lt.prevProjTurns`.
3. Subsequent tool increments (e.g. +1 turn) emit only the new activity delta in `RecentEvents` rather than re-emitting historical lifetime project totals.
4. Deterministic unit tests in `internal/thermal/live_test.go` verify sequential baseline stability across idle polls.
5. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
