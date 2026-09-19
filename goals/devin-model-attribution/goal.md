# Goal: Devin model attribution and honest unestimated reporting

## Goal Description
Emit Devin per-model daily tokens from sessions.model so the estimator can price 19.2B tokens it currently ignores, and report which tools carry no price instead of treating them as free

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 8
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/devin-model-attribution/facts.md`](facts.md)
- **Execution Plan**: [`goals/devin-model-attribution/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
