# Plan: Clamp project activity deltas to tool-level totals (live-tracker-delta-clamping)

## Execution Steps
- [x] Task 1 — Enforce delta ceilings:
  In `internal/thermal/live.go:654-670`, clamp `pDTok` to `deltaTok` and `pDTurns` to `deltaTurns` unconditionally, whether or not the baseline was previously seen.
- [x] Task 2 — Add ceiling enforcement unit tests:
  In `internal/thermal/live_test.go`, test that if an anomaly introduces a large project jump exceeding the tool delta, `pDTok` and `pDTurns` are bounded by `deltaTok` and `deltaTurns`.
- [x] Task 3 — Run test suite:
  Execute `go test -race ./internal/thermal/...`.
- [x] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-tracker-delta-clamping/goal.md)` and verify with `sila goals`.
