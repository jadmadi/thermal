# Goal: Preserve fractional sub-1.0 speeds in live monitor (live-speed-fractional-display)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Display fractional token and step speeds for positive sub-1.0 rates in `internal/tui/live.go` `formatSpeed` instead of integer-truncating to `0 step/s` or `0 tok/s`.

## Dependencies & Execution Order
- Mode: Independent (disjoint, parallelizable)
- Depends On: none
- Sequence: 93
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-speed-fractional-display/facts.md`](facts.md)
- Execution Plan: [`goals/live-speed-fractional-display/plan.md`](plan.md)

## Done Condition
1. In `internal/tui/live.go:586-603`, when `turnSpeed` or `tokSpeed` is greater than 0 but less than 1.0 (or less than 10.0), speed formats with one decimal place (e.g. `0.7 step/s` or `0.7 tok/s`) instead of truncating via `int64` to `0 step/s` or `0 tok/s`.
2. Exact zero speeds continue to format cleanly as `0 step/s` or `0 tok/s` in muted text.
3. Unit tests in `internal/tui/live_test.go` verify that rates such as 40 step/min (0.667 step/s) format as `0.7 step/s`.
4. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
