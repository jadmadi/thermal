# Plan: Preallocated Slice Capacity in Leaderboard Rendering

## Execution Steps
- [x] **Task 1 — Update Slice Allocation**: In `internal/render/leaderboard.go`, allocate `tokenResults` and `activityResults` with `len(results)` capacity.
- [x] **Task 2 — Unit Test Verification**: Run `go test -race ./internal/render/...` to verify leaderboard table generation and sorting.
- [x] **Task 3 — Release Gate**: Run `./scripts/simulated_user_gate.sh` to ensure full CLI parity.
