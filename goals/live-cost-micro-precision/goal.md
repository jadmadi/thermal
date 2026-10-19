# Goal: Support 4-decimal sub-cent precision in live event cost display (live-cost-micro-precision)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Format sub-cent event costs in `internal/tui/live.go` `formatCost` with 4 decimal places (`~$0.0042`) instead of collapsing small fractional token burst costs to `<$0.01` or `$0.00`.

## Dependencies & Execution Order
- Mode: Independent (disjoint, parallelizable)
- Depends On: none
- Sequence: 96
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-cost-micro-precision/facts.md`](facts.md)
- Execution Plan: [`goals/live-cost-micro-precision/plan.md`](plan.md)

## Done Condition
1. In `internal/tui/live.go:659-670`, `formatCost` renders micro-costs between $0.0001 and $0.01 with 4-decimal precision (e.g. `~$0.0035`) instead of collapsing to `<$0.01`.
2. Exact zero costs continue to format as `$0.00` in muted color.
3. Unit tests in `internal/tui/live_test.go` verify 4-decimal formatting for values such as `0.0035` and `0.0004`.
4. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
