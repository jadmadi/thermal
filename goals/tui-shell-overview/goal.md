# Goal: TUI shell and Overview view

## Goal Description
Add the thermal dashboard command: a Bubble Tea v2 application shell with tab routing, a key map, range selection (30d, 90d, 1y, all) and a metric toggle, plus the Overview view with KPI cards, a ranked tool table with inline bars, per-tool sparklines and a 12-week share strip. The data adapter loads through the existing loaders once and aggregates per view and range on demand. stdout that is not a TTY prints a pointer to the static commands instead.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: charting-analytics
- **Sequence**: 3
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/tui-shell-overview/facts.md`](facts.md)
- **Execution Plan**: [`goals/tui-shell-overview/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
