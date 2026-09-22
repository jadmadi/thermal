# Goal: Publish an agent skills discovery index

## Goal Description
Publish a skills discovery index at /.well-known/agent-skills/index.json conforming to Agent Skills Discovery RFC v0.2.0 with $schema, skills array, names, urls, and sha256 digests.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/.well-known/agent-skills/index.json

## References
- **Shared Understanding & Fact Sheet**: [`goals/agent-skills-discovery/facts.md`](facts.md)
- **Execution Plan**: [`goals/agent-skills-discovery/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
