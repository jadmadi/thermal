# Goal: Support DeepSeek harness (dsh) in Thermal

## Goal Description
Add dsh loader engine, registry, tests, and documentation to thermal

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/dsh-loader/facts.md`](facts.md)
- **Execution Plan**: [`goals/dsh-loader/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
