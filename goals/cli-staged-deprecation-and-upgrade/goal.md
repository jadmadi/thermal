# Goal: CLI Staged Deprecation Warnings & Self-Updater Migration Notifications

## Goal Description
Implement runtime deprecation warnings and upgrade alerting in Thermal CLI. 1. Add non-fatal warning infrastructure in cmd/thermal/main.go emitting to os.Stderr (bypassing stdout and JSON) when deprecated commands, aliases, or flags are used, directing users to docs/MIGRATION.md during the N-1 deprecation window before removal in N. 2. Enhance thermal upgrade (cmd/thermal/upgrade.go) to detect version hops across breaking boundaries and print a notice pointing to the migration guide. 3. Add simulated user tests in scripts/simulated_user_gate.sh and unit tests in cmd/thermal/upgrade_test.go verifying deprecation warning behavior and migration link output.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: version-migration-guide
- **Sequence**: 8
- **Shape**: ship

## Files to Touch
- cmd/thermal/main.go
- cmd/thermal/upgrade.go
- cmd/thermal/upgrade_test.go
- cmd/thermal/args_test.go
- scripts/simulated_user_gate.sh

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-staged-deprecation-and-upgrade/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-staged-deprecation-and-upgrade/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
