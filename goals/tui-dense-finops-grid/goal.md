# Goal: Multi-Dimensional FinOps TUI Grid, Activity Taxonomy & Sub-Tool Decomposition

## Goal Description
Implement high-density 9-box FinOps grid view in the Bubble Tea TUI ('thermal stats --dense'). Features activity taxonomy classification (Coding, Debugging, Testing, Exploration), sub-tool shell command decomposition from bash logs, MCP server overhead tracking, and real-time cache efficiency metrics.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/tui/dense.go
- internal/tui/dense_test.go
- internal/thermal/taxonomy.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/tui-dense-finops-grid/facts.md`](facts.md)
- **Execution Plan**: [`goals/tui-dense-finops-grid/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
