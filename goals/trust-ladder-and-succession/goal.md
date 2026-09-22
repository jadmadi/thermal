# Goal: Staged Trust Ladder, Succession Protocol & Absence Factor Monitoring

## Goal Description
Implement staged trust ladder (contributor to key holder) separating merge rights from release keys, hiatus/emeritus policies, emergency succession runbook, and automated absence factor monitoring.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: governance-foundations, repo-alignment-and-security
- **Sequence**: 5
- **Shape**: ship

## Files to Touch
- GOVERNANCE.md
- MAINTAINERS.md
- scripts/check_governance_metrics.sh
- scripts/simulated_user_gate.sh

## References
- **Shared Understanding & Fact Sheet**: [`goals/trust-ladder-and-succession/facts.md`](facts.md)
- **Execution Plan**: [`goals/trust-ladder-and-succession/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
