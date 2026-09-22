# Goal: Version Migration Guide Architecture & docs/MIGRATION.md

## Goal Description
Establish Thermal's formal migration guide (docs/MIGRATION.md) conforming strictly to the version-migration-guide standard. Retroactively inventory and document past breaking changes (dropping cc alias in v0.7.0, strict report flag validation in v0.7.0, disjoint token model in v0.11.0, and replay schema additions in v0.12.0) using the 6-part Entry Anatomy (stable anchors, affected-if detection signals with literal error strings and search regex, 3-sentence rationale, before/after snippets, fix, and escape hatch). Document machine-readable JSON schema contracts for script consumers, and codify the migration documentation governance gate in AGENTS.md.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 7
- **Shape**: ship

## Files to Touch
- docs/MIGRATION.md
- docs/pages/index.html
- AGENTS.md
- README.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/version-migration-guide/facts.md`](facts.md)
- **Execution Plan**: [`goals/version-migration-guide/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
