# Goal: Preallocated Slice Capacity in Leaderboard Rendering

## Goal Description
In `internal/render/leaderboard.go`, `RenderLeaderboard` splits all discovered tool results into token-based warriors and activity-based hunters. Previously, slices `tokenResults` and `activityResults` were initialized with `make([]thermal.ToolResult, 0)`, incurring unnecessary memory growth and backing array copy operations as tools were appended.

This goal preallocates capacity `len(results)` for both slices (`make([]thermal.ToolResult, 0, len(results))`), ensuring zero reallocations during leaderboard partitioning.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- `internal/render/leaderboard.go`
- `internal/render/leaderboard_test.go`

## References
- **Shared Understanding & Fact Sheet**: [`goals/leaderboard-slice-preallocation/facts.md`](facts.md)
- **Execution Plan**: [`goals/leaderboard-slice-preallocation/plan.md`](plan.md)

## Done Condition
1. Leaderboard slices are preallocated with capacity based on total tools count.
2. Rendering performance is verified with zero extra allocations on slice expansion.
3. Unit tests pass cleanly: `go test -race ./internal/render/...`.
4. Simulated user gate passes: `./scripts/simulated_user_gate.sh`.
