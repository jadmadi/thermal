# Goal: Chart flags, docs, and 0.7.0 release

## Goal Description
Add --chart to projects, models and the period reports, print concrete bar rows under the tables, document the chart and TUI surfaces in the README with a screencast, record the TUI dependency boundary and chart conventions in AGENTS.md, run the full verification checklist, and cut release 0.7.0.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: tui-projects-view, tui-mix-models-views, tui-stats-view
- **Sequence**: 7
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/charting-flags-docs-release/facts.md`](facts.md)
- **Execution Plan**: [`goals/charting-flags-docs-release/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
