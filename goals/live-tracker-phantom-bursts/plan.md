# Plan: Fix live tracker phantom bursts and fractional speed display

## Execution Steps
- [ ] Task 1 — Preserve cumulative project baselines in `internal/thermal/live.go`:
  In the idle / no-delta branch (lines 806-813), accumulate `pd.Tokens` and `pd.Turns` across all days matching the delta path instead of filtering by `pd.Day == lt.todayDay`.
  Also preserve cumulative baselines across day boundaries (lines 159-165).
- [ ] Task 2 — Clamp project deltas to tool-level deltas in `internal/thermal/live.go`:
  Clamp `pDTurns` to `deltaTurns` and `pDTok` to `deltaTok` when `turnSeen` and `tokSeen` are true in lines 654-670.
- [ ] Task 3 — Support fractional speeds in `internal/tui/live.go`:
  Update `formatSpeed` so positive speeds below 1.0 (or below 10.0) format as `0.X step/s` or `0.X tok/s` rather than truncating to `0 step/s`.
- [ ] Task 4 — Add comprehensive test cases:
  In `internal/thermal/live_test.go`, test sequential polls: initial poll -> idle poll (delta == 0) -> incremental poll (+1 turn) verifying that only +1 turn is emitted and historical project baselines are not re-emitted.
  In `internal/tui/live_test.go`, test `formatSpeed` with `0.667` speed.
- [ ] Task 5 — Full test suite and race verification:
  Run `go test -race ./...` and verify clean build of CLI binary.
- [ ] Task 6 — Attribution commit:
  Commit with Conventional Commit referencing `(goals/live-tracker-phantom-bursts/goal.md)` and verify with `sila goals`.
