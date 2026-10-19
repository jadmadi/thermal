# Goal: Preserve cumulative project baselines across midnight transitions (live-tracker-midnight-drift)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Prevent midnight day boundary transition in `LiveTracker.Poll` from resetting cumulative project baselines to only `pd.Day == currentDay` values.

## Dependencies & Execution Order
- Mode: Dependent (requires prerequisite goals)
- Depends On: live-tracker-phantom-bursts
- Sequence: 91
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-tracker-midnight-drift/facts.md`](facts.md)
- Execution Plan: [`goals/live-tracker-midnight-drift/plan.md`](plan.md)

## Done Condition
1. In `internal/thermal/live.go:159-165`, when midnight boundary is crossed (`currentDay != lt.todayDay`), project baselines accumulate all days (`pd.Tokens` and `pd.Turns`) instead of resetting to only `pd.Day == currentDay`.
2. Crossing midnight during an active monitoring session preserves historical cumulative baselines so activity after midnight does not trigger burst inflation.
3. Unit tests in `internal/thermal/live_test.go` verify project baseline stability across simulated midnight transitions.
4. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
