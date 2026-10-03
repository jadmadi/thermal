# Goal: Platform-Native Application Data Directory Resolution

## Goal Description
Add Windows APPDATA/LOCALAPPDATA and macOS Library/Application Support directory discovery to internal/loaders/registry.go

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/platform-native-data-directories/facts.md`](facts.md)
- **Execution Plan**: [`goals/platform-native-data-directories/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
