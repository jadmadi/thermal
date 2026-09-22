# Goal: TUI Mix and Models views

## Goal Description
Stacked share over time by tool and by model, a switching panel with dominant tool, switches per month, concentration and active tool count, and a Models view with ranked bars and a share strip. Shares must sum to 100 percent and concentration must match the static mix output.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: tui-shell-overview
- **Sequence**: 5
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/tui-mix-models-views/facts.md`](facts.md)
- **Execution Plan**: [`goals/tui-mix-models-views/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
