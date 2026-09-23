# Goal: Embedded Local Web Dashboard for Agent Telemetry & Streak Visualization

## Goal Description
Implement a local-first web dashboard served via 'thermal serve' on localhost. Embeds frontend assets directly into the Go binary via go:embed, offering local-first visualization for streaks, multi-tool token consumption, and session timelines with zero telemetry tracking.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/server/server.go
- internal/server/server_test.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/embedded-web-dashboard/facts.md`](facts.md)
- **Execution Plan**: [`goals/embedded-web-dashboard/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
