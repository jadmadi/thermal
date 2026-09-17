# Goal: TUI Projects view with drill-down

## Goal Description
Ranked project table with token and cost bars, sortable columns, and filters for range and tool. Enter opens a project detail with tool mix, model mix, weekly trend and a project heatmap. Totals must equal thermal projects --json for the same window.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: tui-shell-overview
- **Sequence**: 4
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/tui-projects-view/facts.md`](facts.md)
- **Execution Plan**: [`goals/tui-projects-view/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
