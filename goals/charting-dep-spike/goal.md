# Goal: Charting dependency spike: ntcharts v2 against the stock charm stack

## Goal Description
Verify ntcharts v2.2.0 builds and renders against stock charm.land/bubbletea v2.0.9 and lipgloss v2.0.6 in a scratch module outside the repo, measure the binary size delta, cross-build windows, and pick the hand-rolled braille and block canvas fallback if anything fails. The ntcharts go.mod pins a Bubble Tea fork through a replace that does not travel, so this goal decides the stack before any chart code lands.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 1
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/charting-dep-spike/facts.md`](facts.md)
- **Execution Plan**: [`goals/charting-dep-spike/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
