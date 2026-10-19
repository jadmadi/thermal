# Plan: Preserve fractional sub-1.0 speeds in live monitor (live-speed-fractional-display)

## Execution Steps
- [x] Task 1 — Update `formatSpeed` to support fractional precision:
  In `internal/tui/live.go:586-603`, when `turnSpeed > 0 && turnSpeed < 1.0`, format as `fmt.Sprintf("%.1f step/s", turnSpeed)`. Apply the same fractional precision to `tokSpeed` when `tokSpeed > 0 && tokSpeed < 1.0`.
- [x] Task 2 — Add speed formatting unit tests:
  In `internal/tui/live_test.go`, test `formatSpeed` with `0.667` asserting `0.7 step/s`, `0.0` asserting `0 step/s`, and `4.2` asserting `4 step/s` (or `4.2 step/s`).
- [x] Task 3 — Run test suite:
  Execute `go test -race ./internal/tui/...`.
- [x] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-speed-fractional-display/goal.md)` and verify with `sila goals`.
