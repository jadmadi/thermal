# Goal: Fix turns and cache-hit text truncation in narrow live monitor grid (live-grid-turns-truncation)

## Chain of Custody (Isnad)
- Registered By: Jad Madi (via Antigravity)
- Spec Audited By: Pending Spec Audit
- Implemented By: Pending Implementation

## Goal Description
Prevent `Turns: X turns (Y% hit)` metric from clipping with ellipsis (`6642 turns (86% h…`) on standard 80-column terminals through width-aware responsive abbreviation in `internal/tui/live.go`.

## Dependencies & Execution Order
- Mode: Independent (disjoint, parallelizable)
- Depends On: none
- Sequence: 94
- Shape: fix
- Tier: immediate

## Files to Touch
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`

## References
- Shared Understanding & Fact Sheet: [`goals/live-grid-turns-truncation/facts.md`](facts.md)
- Execution Plan: [`goals/live-grid-turns-truncation/plan.md`](plan.md)

## Done Condition
1. In `internal/tui/live.go`, `formatTurnsAndHit` accepts column width (`colW`) and renders text that fits without ellipsis truncation on 80-column terminal layouts (`colW = 24`).
2. When full text exceeds available width, it compactly formats turns (e.g. `6.6K (86% hit)` or `6642 turns`) to stay within cell boundaries.
3. Unit tests in `internal/tui/live_test.go` verify that at `colW = 24`, no ellipsis or truncation artifacts appear.
4. All automated unit tests pass cleanly with `-race` (`go test -race ./...`).
