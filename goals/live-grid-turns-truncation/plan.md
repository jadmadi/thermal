# Plan: Fix turns and cache-hit text truncation in narrow live monitor grid (live-grid-turns-truncation)

## Execution Steps
- [ ] Task 1 — Implement responsive turns formatting:
  Update `formatTurnsAndHit` in `internal/tui/live.go` to accept `availWidth int`. Implement progressive abbreviation: full count with percentage -> compact count with percentage -> plain count.
- [ ] Task 2 — Update call sites in `internal/tui/live.go`:
  Pass available width (`colW - len("Turns:    ")`) into `formatTurnsAndHit`.
- [ ] Task 3 — Add unit tests:
  In `internal/tui/live_test.go`, test width constraints with 18, 22, 24, and 30 cell widths to verify no ellipses are produced.
- [ ] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-grid-turns-truncation/goal.md)` and verify with `sila goals`.
