# Plan: Release Communication, Editorial Changelog & Automation Alignment

## Execution Steps

- [x] **Phase 1 — Configure `release-please-config.json` Sections**:
  - Add explicit `changelog-sections` mapping to suppress non-user-facing commit types (`chore`, `test`, `ci`, `docs`, `refactor`, `style`).
  - Keep `feat` ("Features"), `fix` ("Bug Fixes"), `perf` ("Performance Improvements"), and `revert` ("Reverts") visible.

- [x] **Phase 2 — Codify Commit Hygiene & Release PR Polish in `AGENTS.md`**:
  - Update Section 4.C in `AGENTS.md` with explicit commit title guidelines:
    - Subject MUST describe user-observable CLI/TUI behavior rather than internal file paths or packages.
    - Sila goal token `(goals/<slug>/goal.md)` MUST be appended at the end of the subject to satisfy Sila attribution without displacing the user-facing summary.
  - Document the Release PR Polish procedure: prior to merging automated release PRs (`chore(main): release x.y.z`), review and polish the changelog text according to `/changelog-writing` guidelines, ensuring breaking changes follow the 3-part form (What breaks, Who it hits, What to do) and link to `docs/MIGRATION.md`.

- [x] **Phase 3 — Verify Release Gate & Workflow**:
  - Update `scripts/simulated_user_gate.sh` to check `release-please-config.json` validity and ensure `CHANGELOG.md` conforms to Keep a Changelog standards.
  - Run `./scripts/simulated_user_gate.sh` and `go test -race ./...`.

- [x] **Phase 4 — Attribution Commit**:
  - Commit with subject: `feat(release): adapt changelog-writing standards and configure release-please sections (goals/release-communication-and-changelog/goal.md)`.
  - Run `sila audit_goals` to verify closed-loop attribution.
