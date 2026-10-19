# Plan: Support 4-decimal sub-cent precision in live event cost display (live-cost-micro-precision)

## Execution Steps
- [ ] Task 1 — Implement 4-decimal precision for micro-costs:
  In `internal/tui/live.go:659-670`, when `c > 0 && c < 0.01`, format as `fmt.Sprintf("~$%.4f", c)`. Keep exact `$0.00` for `c <= 0`.
- [ ] Task 2 — Add cost formatting unit tests:
  In `internal/tui/live_test.go`, test `formatCost` with `0.0035`, `0.0004`, `0.0`, and `1.25`.
- [ ] Task 3 — Run test suite:
  Execute `go test -race ./internal/tui/...`.
- [ ] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-cost-micro-precision/goal.md)` and verify with `sila goals`.
