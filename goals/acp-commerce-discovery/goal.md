# Goal: Publish ACP discovery metadata so agents can discover your commerce API

## Goal Description
Serve /.well-known/acp.json at origin root declaring protocol.name 'acp', version, api_base_url, transports, and capabilities.services for Agentic Commerce Protocol per agenticcommerce.dev.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/.well-known/acp.json

## References
- **Shared Understanding & Fact Sheet**: [`goals/acp-commerce-discovery/facts.md`](facts.md)
- **Execution Plan**: [`goals/acp-commerce-discovery/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
