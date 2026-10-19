# Plan: Preserve cumulative project baselines across midnight transitions (live-tracker-midnight-drift)

## Execution Steps
- [x] Task 1 — Preserve cumulative baselines across midnight:
  In `internal/thermal/live.go:159-165`, do not wipe or filter baselines by `currentDay`. Accumulate lifetime tokens and turns across all `projDays` for `lt.prevProjTokens` and `lt.prevProjTurns`.
- [x] Task 2 — Add midnight boundary regression unit test:
  In `internal/thermal/live_test.go`, write a test that polls right before midnight, advances time past midnight with new tool activity, and asserts that only the new day's increment is emitted as a `LiveEvent`.
- [x] Task 3 — Run test suite:
  Execute `go test -race ./internal/thermal/...`.
- [x] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-tracker-midnight-drift/goal.md)` and verify with `sila goals`.
