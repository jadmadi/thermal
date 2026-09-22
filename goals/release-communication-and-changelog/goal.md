# Goal: Release Communication, Editorial Changelog & Automation Alignment

## Goal Description
Adapt the changelog-writing discipline to Thermal while preserving automated release-please and goreleaser pipelines. Configure release-please-config.json to filter internal noise, reconcile Sila commit attribution with Release-Please to prevent leaking internal goal tokens into public changelogs, codify impact-first and 3-part breaking change standards in AGENTS.md, and add verification checks to simulated user gate.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: purge-legacy-mit-releases, cli-staged-deprecation-and-upgrade
- **Sequence**: 9
- **Shape**: ship

## Files to Touch
- release-please-config.json
- AGENTS.md
- .github/workflows/release-please.yml
- scripts/simulated_user_gate.sh

## References
- **Shared Understanding & Fact Sheet**: [`goals/release-communication-and-changelog/facts.md`](facts.md)
- **Execution Plan**: [`goals/release-communication-and-changelog/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
