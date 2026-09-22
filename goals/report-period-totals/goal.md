# Goal: Period report totals and table format

## Goal Description
Weekly and monthly totals must equal the displayed rows for the requested window, empty periods hidden, table well formatted

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/report-period-totals/facts.md`](facts.md)
- **Execution Plan**: [`goals/report-period-totals/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
