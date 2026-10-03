# Goal: Windows SQLite URI and Path Normalization

## Goal Description
Add shared SQLite connection URI helper in internal/loaders to clean Windows drive letters and convert backslashes to forward slashes

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/windows-sqlite-uri-path-normalization/facts.md`](facts.md)
- **Execution Plan**: [`goals/windows-sqlite-uri-path-normalization/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
