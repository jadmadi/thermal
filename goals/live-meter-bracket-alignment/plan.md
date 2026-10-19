# Plan: Align flame intensity gauge and sparkline bracket widths (live-meter-bracket-alignment)

## Execution Steps
- [ ] Task 1 — Design coordinated gauge layout:
  Calculate a shared gauge inner width or coordinate bracket offsets between `formatFlameLine` and `formatSparkLine` in `internal/tui/live.go`.
- [ ] Task 2 — Implement aligned brackets:
  Ensure `Flame: [` and `Spark: [` start at identical indentation and scale proportionally.
- [ ] Task 3 — Add unit tests:
  In `internal/tui/live_test.go`, assert aligned bracket positions at multiple terminal widths.
- [ ] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-meter-bracket-alignment/goal.md)` and verify with `sila goals`.
