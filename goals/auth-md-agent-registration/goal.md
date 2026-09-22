# Goal: Publish Auth.md metadata for agent registration

## Goal Description
Serve /auth.md at site root with agent registration instructions, link to oauth metadata, and agent_auth block per workos/auth.md specification.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- docs/pages/auth.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/auth-md-agent-registration/facts.md`](facts.md)
- **Execution Plan**: [`goals/auth-md-agent-registration/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
