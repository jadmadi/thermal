# Goal: Reasoning Effort Breakdown & Cognitive Intensity Telemetry

## Goal Description
Project reasoning effort levels (low, medium, high) from Codex (and reasoning-capable models like o1, o3-mini, Claude Thinking, DeepSeek-R1) onto thermal models and thermal mix, measuring cognitive intensity and reasoning token proportions across sessions.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 71
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- internal/loaders/codex.go
- internal/loaders/codex_test.go
- internal/thermal/types.go
- internal/thermal/aggregator.go
- internal/render/models.go
- internal/render/mix.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/reasoning-effort-telemetry/facts.md`](facts.md)
- **Execution Plan**: [`goals/reasoning-effort-telemetry/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
