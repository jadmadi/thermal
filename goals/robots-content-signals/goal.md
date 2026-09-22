# Goal: Declare AI content usage preferences with Content Signals in robots.txt

## Goal Description
Add Content-Signal directives (ai-train, search, ai-input) to robots.txt to declare AI content usage preferences per contentsignals.org and draft-romm-aipref-contentsignals.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/robots.txt

## References
- **Shared Understanding & Fact Sheet**: [`goals/robots-content-signals/facts.md`](facts.md)
- **Execution Plan**: [`goals/robots-content-signals/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
