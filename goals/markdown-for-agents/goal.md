# Goal: Return HTML responses as markdown when agents request it

## Goal Description
Enable Markdown for Agents so HTTP requests with Accept: text/markdown return a markdown version of HTML responses with Content-Type: text/markdown.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/index.md
- docs/pages/_worker.js

## References
- **Shared Understanding & Fact Sheet**: [`goals/markdown-for-agents/facts.md`](facts.md)
- **Execution Plan**: [`goals/markdown-for-agents/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
