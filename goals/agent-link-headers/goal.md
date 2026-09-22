# Goal: Include Link response headers for agent discovery (RFC 8288)

## Goal Description
Add Link response headers (or HTML link relations) to the Thermal documentation site pointing agents to API catalog, skills index, and documentation per RFC 8288 and RFC 9727.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/index.html
- docs/pages/_headers

## References
- **Shared Understanding & Fact Sheet**: [`goals/agent-link-headers/facts.md`](facts.md)
- **Execution Plan**: [`goals/agent-link-headers/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
