# Goal: Support WebMCP to expose site tools to AI agents via the browser

## Goal Description
Implement WebMCP API by invoking navigator.modelContext.provideContext() on the Thermal documentation website to expose tools (e.g., getHeatmap, inspectTelemetry, getSupportedTools) to browser AI agents.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/index.html

## References
- **Shared Understanding & Fact Sheet**: [`goals/webmcp-browser-tools/facts.md`](facts.md)
- **Execution Plan**: [`goals/webmcp-browser-tools/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
