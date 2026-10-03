# Goal: Cross-Platform Operating System Test Matrix in CI

## Goal Description
Add macOS and Windows test runners to GitHub Actions CI to verify platform behavior and unit tests on every pull request

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/cross-platform-ci-matrix/facts.md`](facts.md)
- **Execution Plan**: [`goals/cross-platform-ci-matrix/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
