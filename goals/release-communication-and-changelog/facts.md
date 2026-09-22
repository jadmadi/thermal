# Facts: Release Communication, Editorial Changelog & Automation Alignment

## Architectural Invariants & Constraints

1. **Sila Closed-Loop Goal Attribution Contract (`fact_0bd07325`)**:
   - A Sila goal transitions to `implemented` only when a commit with a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` subject reaches `main` carrying the literal token `(goals/<slug>/goal.md)` and modifying at least one tracked file outside the goal directory.
   - Sila ignores commits with prefixes `test(...)` or `docs(...)` (`fact_6ddeb3c4`).
   - Consequently, the goal token `(goals/<slug>/goal.md)` MUST remain present in the commit subject to ensure automated goal closure.

2. **Release-Please Generation Invariant**:
   - `release-please-action` (v5) reads commit subjects since the last tag to calculate SemVer and populate `CHANGELOG.md`.
   - By default, without `changelog-sections` configuration, `release-please` treats all unclassified commits as features or miscellaneous, exposing internal scaffolding (e.g. `goals/.../goal.md`, test release gates, sila tasks).

3. **Reconciliation Mechanism**:
   - Rather than stripping the Sila goal token from commits (which breaks Sila), we:
     1. Configure `release-please-config.json` to hide development chores, tests, refactors, and CI churn from the public changelog.
     2. Enforce impact-first phrasing in commit subjects: `feat(<scope>): <user-observable-effect> (goals/<slug>/goal.md)`.
     3. Establish the Release PR Polish step: when `release-please` opens a release PR (`chore(main): release x.y.z`), apply `/changelog-writing` principles to polish the release description and strip raw goal references before merging.
     4. Enforce the 3-part breaking change standard (What breaks, Who it hits, What to do) linking directly to `docs/MIGRATION.md`.

## File & Interface Contracts

- `release-please-config.json`: Release-please package and changelog section configuration.
- `AGENTS.md`: Sections 4.C (Automated Release Pipeline) and 3 (Terminal Rendering & CLI) defining commit subject rules and release notes hygiene.
- `.github/workflows/release-please.yml`: GitHub Actions workflow orchestrating release-please and GoReleaser.
- `scripts/simulated_user_gate.sh`: Release gate verification script.
