# Goal: Fix live tracker phantom bursts and fractional speed display

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Prevent live tracker from clobbering project cumulative baselines on idle ticks and preserve fractional sub-1.0 step and token speeds in terminal live monitor.

## Dependencies & Execution Order
- Mode: Independent (disjoint, parallelizable)
- Depends On: none
- Sequence: 90
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-tracker-phantom-bursts/facts.md`](facts.md)
- Execution Plan: [`goals/live-tracker-phantom-bursts/plan.md`](plan.md)

## Done Condition
1. `LiveTracker.Poll` maintains cumulative project turn and token baselines across all days on idle (no-delta) ticks and day boundaries instead of collapsing to single-day counts.
2. Project deltas (`pDTurns` and `pDTok`) clamp to tool-level `deltaTurns` and `deltaTok` when baselines exist, preventing project deltas from exceeding overall tool activity.
3. `formatSpeed` displays fractional speeds (e.g., `0.7 step/s` or `0.7 tok/s`) for positive sub-1.0 speeds instead of truncating via `int64` to `0 step/s` when burn rate is positive.
4. Unit tests in `internal/thermal/live_test.go` and `internal/tui/live_test.go` verify baseline preservation on idle polls and fractional speed formatting.
5. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
