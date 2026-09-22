# Goal: AGPL-3.0 Relicensing, Commercial Dual-Licensing & Contributor Gating

## Goal Description
Relicense Thermal to GNU AGPLv3, establish commercial dual-licensing track, implement thermal license CLI command, fix author attribution typo to Jad Madi, and add contributor CLA gating.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship

## Files to Touch
- LICENSE
- DUAL-LICENSE.md
- NOTICES.md
- CONTRIBUTING.md
- README.md
- AGENTS.md
- cmd/thermal/main.go
- cmd/thermal/args_test.go
- scripts/simulated_user_gate.sh

## References
- **Shared Understanding & Fact Sheet**: [`goals/agpl-dual-licensing/facts.md`](facts.md)
- **Execution Plan**: [`goals/agpl-dual-licensing/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
