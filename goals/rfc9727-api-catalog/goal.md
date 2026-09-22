# Goal: Publish an API catalog for automated API discovery (RFC 9727)

## Goal Description
Publish /.well-known/api-catalog returning application/linkset+json with anchor URL and link relations (service-desc, service-doc, status) per RFC 9727 and RFC 9264.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/.well-known/api-catalog

## References
- **Shared Understanding & Fact Sheet**: [`goals/rfc9727-api-catalog/facts.md`](facts.md)
- **Execution Plan**: [`goals/rfc9727-api-catalog/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
