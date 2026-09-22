# Goal: Defensive Timestamp Parsing and Malformed Date Guards in Streak Computation

## Goal Description
`ComputeStreaks` calculates current active streaks and longest historical streaks across daily activity timestamps. Previously, `time.Parse("2006-01-02", ...)` discarded errors using `_`. When encountering corrupt or non-standard date entries from unformatted loaders, `time.Parse` produced a zero `time.Time`, resetting the ongoing streak run or causing unexpected loop iterations.

This goal introduces defensive error checking on `time.Parse`, cleanly skipping unparseable date strings and ensuring streak math is resilient against corrupted or non-conforming day keys.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- `internal/thermal/streak.go`
- `internal/thermal/streak_test.go`

## References
- **Shared Understanding & Fact Sheet**: [`goals/streak-parsing-defensive-guards/facts.md`](facts.md)
- **Execution Plan**: [`goals/streak-parsing-defensive-guards/plan.md`](plan.md)

## Done Condition
1. `ComputeStreaks` validates `err == nil` on all date parsing operations.
2. Malformed or unparseable day strings are safely bypassed without breaking streak runs.
3. Unit tests pass cleanly with race detection: `go test -race ./internal/thermal/...`.
4. Simulated user gate passes: `./scripts/simulated_user_gate.sh`.
