# Goal: Publish an ARD (Agentic Resource Discovery) manifest for site capabilities

## Goal Description
Serve /.well-known/ai-catalog.json at origin root with Content-Type: application/json and Access-Control-Allow-Origin: * conforming to ARD spec with URNs, displayNames, and representativeQueries.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/.well-known/ai-catalog.json

## References
- **Shared Understanding & Fact Sheet**: [`goals/ard-capability-manifest/facts.md`](facts.md)
- **Execution Plan**: [`goals/ard-capability-manifest/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
