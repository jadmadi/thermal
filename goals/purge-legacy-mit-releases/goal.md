# Goal: Retire Legacy MIT Releases and Reset Release Line to AGPL-3.0

## Goal Description
Audit, purge, and archive legacy pre-AGPL-3.0 releases, assets, and git tags from GitHub to ensure public releases strictly align with the new AGPL-3.0 Community and Commercial Enterprise dual-licensing track. Provide a verified GitHub CLI (`gh release delete`, `git push origin --delete`) runbook and migration script so external binary distributions exclusively distribute AGPL-3.0 binaries starting with the next release.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: community-covenant-neutrality
- **Sequence**: 4
- **Shape**: ship

## Files to Touch
- scripts/retire_legacy_releases.sh
- docs/RELEASE.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/purge-legacy-mit-releases/facts.md`](facts.md)
- **Execution Plan**: [`goals/purge-legacy-mit-releases/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
