# Plan: Fix live tracker idle baseline collapse (live-tracker-phantom-bursts)

## Execution Steps
- [x] Task 1 — Accumulate cumulative project baselines in idle branch:
  In `internal/thermal/live.go:806-813`, remove the `pd.Day == lt.todayDay` filter and accumulate `pd.Tokens` and `pd.Turns` across all days into `lt.prevProjTokens` and `lt.prevProjTurns`.
- [x] Task 2 — Add regression unit test in `internal/thermal/live_test.go`:
  Simulate a sequence: initial poll with multi-day project data -> idle poll with identical data -> incremental poll with +1 turn. Assert that the incremental poll produces a single `LiveEvent` with +1 turn and does not re-emit historical project totals.
- [x] Task 3 — Run test suite:
  Execute `go test -race ./internal/thermal/...`.
- [x] Task 4 — Attribution commit:
  Commit with Conventional Commit ending in `(goals/live-tracker-phantom-bursts/goal.md)` and verify with `sila goals`.
