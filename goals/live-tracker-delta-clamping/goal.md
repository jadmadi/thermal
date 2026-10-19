# Goal: Clamp project activity deltas to tool-level totals (live-tracker-delta-clamping)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Ensure individual project token and turn deltas in `LiveTracker.Poll` cannot exceed tool-level `deltaTok` and `deltaTurns`, maintaining mathematical consistency between tool volume and project breakdowns.

## Dependencies & Execution Order
- Mode: Dependent (requires prerequisite goals)
- Depends On: live-tracker-phantom-bursts
- Sequence: 92
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-tracker-delta-clamping/facts.md`](facts.md)
- Execution Plan: [`goals/live-tracker-delta-clamping/plan.md`](plan.md)

## Done Condition
1. In `internal/thermal/live.go:654-670`, `pDTurns` and `pDTok` are explicitly bounded: when `turnSeen` is true, `pDTurns` cannot exceed `deltaTurns`; when `tokSeen` is true, `pDTok` cannot exceed `deltaTok`.
2. Total project turn delta (`totalProjectTurnDelta`) and total project token delta (`totalProjectTokDelta`) cannot exceed tool `deltaTurns` and `deltaTok`.
3. Unit tests in `internal/thermal/live_test.go` assert that anomalous or mismatched project rows never emit a project burst larger than the tool-level activity.
4. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
