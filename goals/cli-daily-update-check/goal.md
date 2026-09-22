# Goal: Daily Cached Background Update Check and Upgrade Notification

## Goal Description
Implement daily non-blocking background update check caching GitHub releases in ~/.cache/thermal/update.json and notifying users on interactive TTY when an upgrade is available.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-daily-update-check/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-daily-update-check/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
