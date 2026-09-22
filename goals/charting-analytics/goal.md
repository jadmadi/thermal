# Goal: Shared chart analytics and static chart verbs

## Goal Description
Add the aggregation layer every chart reads from: tool and model mix with share, switches and Herfindahl concentration; distribution stats with median, p90, weekday profile and robust outliers; trend with least-squares slope and a 14-day month-end projection. Then expose them as the static verbs thermal trend, thermal mix and thermal stats with --metric, --by, --last, --grain and --json. Numbers get verified in text before any TUI exists.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: charting-dep-spike
- **Sequence**: 2
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/charting-analytics/facts.md`](facts.md)
- **Execution Plan**: [`goals/charting-analytics/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
