# Goal: Community Reciprocity Covenant, Scope Demarcation & Fork Guidelines

## Goal Description
Establish Thermal Community Reciprocity Covenant in DUAL-LICENSE.md and CONTRIBUTING.md, demarcate commercial vs community scopes, define fair-use fork guidelines, and expose covenant in thermal license CLI.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: governance-foundations
- **Sequence**: 3
- **Shape**: ship

## Files to Touch
- DUAL-LICENSE.md
- CONTRIBUTING.md
- cmd/thermal/main.go
- cmd/thermal/args_test.go
- scripts/simulated_user_gate.sh

## References
- **Shared Understanding & Fact Sheet**: [`goals/community-covenant-neutrality/facts.md`](facts.md)
- **Execution Plan**: [`goals/community-covenant-neutrality/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
