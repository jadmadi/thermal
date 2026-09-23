# Goal: Stateless Zero-Database URL Sharing Engine for Streaks & Telemetry

## Goal Description
Implement stateless URL hash sharing for Thermal streaks, summary badges, and telemetry cards using compressed URL-safe tokens (v1.<checksum>.<deflate-b64>). Provides private, serverless, and database-free sharing with zero user tracking.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/share/share.go
- internal/share/share_test.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/stateless-telemetry-sharing/facts.md`](facts.md)
- **Execution Plan**: [`goals/stateless-telemetry-sharing/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
