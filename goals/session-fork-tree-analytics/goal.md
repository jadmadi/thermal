# Goal: Session Fork & Branch Tree Yield Analytics

## Goal Description
Ingest parent-child thread relationships and session forks from Codex (and branching agents) into thermal yield, distinguishing exploratory prototyping tokens from mainline landed code.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 75
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- internal/loaders/codex.go
- internal/loaders/codex_test.go
- internal/thermal/yield.go
- internal/thermal/yield_test.go
- internal/render/yield.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/session-fork-tree-analytics/facts.md`](facts.md)
- **Execution Plan**: [`goals/session-fork-tree-analytics/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
