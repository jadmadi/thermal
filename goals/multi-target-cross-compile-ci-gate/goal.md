# Goal: Multi-Target Cross-Compilation Verification Gate in CI

## Goal Description
Add multi-target cross-compilation check (darwin/arm64, darwin/amd64, windows/amd64, linux/arm64) to GitHub Actions CI

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/multi-target-cross-compile-ci-gate/facts.md`](facts.md)
- **Execution Plan**: [`goals/multi-target-cross-compile-ci-gate/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
