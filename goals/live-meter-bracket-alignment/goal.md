# Goal: Align flame intensity gauge and sparkline bracket widths (live-meter-bracket-alignment)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Harmonize the visual alignment and bracket widths between the flame intensity meter and the 60-second rolling sparkline in `internal/tui/live.go`.

## Dependencies & Execution Order
- Mode: Independent (disjoint, parallelizable)
- Depends On: none
- Sequence: 95
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-meter-bracket-alignment/facts.md`](facts.md)
- Execution Plan: [`goals/live-meter-bracket-alignment/plan.md`](plan.md)

## Done Condition
1. The flame intensity gauge and the 60-second rolling activity sparkline share consistent visual bracket formatting and coordinated column widths across standard terminal sizes.
2. Trailing brackets and spacing avoid visual raggedness in both colorful and no-color terminal modes.
3. Unit tests in `internal/tui/live_test.go` verify bracket formatting across 80, 100, and 120 column widths.
4. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
