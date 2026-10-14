# Goal: ChatGPT Rolling 5-Hour Burst Capacity Replay

## Goal Description
Enhance thermal replay to model dynamic 5-hour rolling message and token burst limits for Codex sessions evaluated against ChatGPT Plus and ChatGPT Pro subscription tiers, providing concrete capacity verdicts and upgrade rationale.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: workload-replay
- **Sequence**: 102
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- internal/thermal/replay.go
- internal/thermal/replay_test.go
- internal/render/replay.go
- internal/render/replay_test.go
- docs/MIGRATION.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/chatgpt-rolling-burst-replay/facts.md`](facts.md)
- **Execution Plan**: [`goals/chatgpt-rolling-burst-replay/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.

## Reopened acceptance, 2026-10-14

Historical implementation and completed tasks remain preserved. New acceptance is pending.
The [current remediation contract](remediation-2026-10-14.md) adds binding acceptance criteria;
complete it as well as preserved prior criteria before claiming this goal delivered.
See the [evaluation](../../docs/evaluations/2026-10-14-promise-vs-delivery.md).
