# Goal: Written Solo Authority Governance, Scope Defense & Decision Rules

## Goal Description
Establish formal GOVERNANCE.md with Solo Authority model, Scope & Anti-Goals, decision classes with bounded clocks, and justified veto rules.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 1
- **Shape**: ship

## Files to Touch
- GOVERNANCE.md
- README.md
- AGENTS.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/governance-foundations/facts.md`](facts.md)
- **Execution Plan**: [`goals/governance-foundations/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
