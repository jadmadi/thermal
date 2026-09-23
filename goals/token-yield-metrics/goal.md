# Goal: Token Yield & Code Output Delta Telemetry ('thermal yield')

## Goal Description
Implement code output delta tracking and token yield telemetry ('thermal yield'). Correlates agent token spend with tangible engineering output by parsing git diff stats (+lines, -lines) from session transcripts, calculating Token Yield (tokens consumed per net line shipped), and identifying high-efficiency versus verbose model behavior.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/thermal/yield.go
- internal/thermal/yield_test.go
- internal/render/yield.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/token-yield-metrics/facts.md`](facts.md)
- **Execution Plan**: [`goals/token-yield-metrics/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
