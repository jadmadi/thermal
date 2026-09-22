# Goal: Shrink the unpriceable token line

## Goal Description
Shrink the UnattributedTokens line. Investigation found two causes: one Codex day whose rollout log names a model but whose day loses it, and activity-only tools whose step counts are counted as tokens. Fix the Codex mapping; decide the activity-count rule with the user before changing it.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 10
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/unpriceable-token-sources/facts.md`](facts.md)
- **Execution Plan**: [`goals/unpriceable-token-sources/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
