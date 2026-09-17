# Goal: TUI Stats view

## Goal Description
Histogram of daily tokens or cost with a log toggle, weekday profile, top days, robust outlier list, and a month-end projection line with a band. Values must match thermal stats --json.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: tui-shell-overview
- **Sequence**: 6
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/tui-stats-view/facts.md`](facts.md)
- **Execution Plan**: [`goals/tui-stats-view/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
