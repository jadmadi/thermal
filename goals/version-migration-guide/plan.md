# Plan: Version Migration Guide Architecture & docs/MIGRATION.md

## Execution Steps

- [x] **Phase 1 — Inventory & Structure `docs/MIGRATION.md`**:
  - Create `docs/MIGRATION.md` with an introduction outlining Thermal's semantic versioning and CLI contract.
  - Document retroactive breaking changes from v0.7.0, v0.11.0, and v0.12.0 using the 6-part Entry Anatomy.
  - Include search patterns, literal error outputs, before/after examples, fixes, and escape hatches.

- [x] **Phase 2 — Governance Gate in `AGENTS.md`**:
  - Add Section 3.6 to `AGENTS.md` defining the Breaking Change & Migration Governance Gate.
  - Require every breaking CLI flag, argument change, or `--json` schema alteration to ship with a corresponding `docs/MIGRATION.md` entry.

- [x] **Phase 3 — Documentation Site & README Links**:
  - Update `README.md` to link to `docs/MIGRATION.md` under documentation & upgrade guidance.
  - Update `docs/pages/index.html` navigation and footer links.

- [x] **Phase 4 — Verification**:
  - Verify markdown linting and relative link integrity across `docs/MIGRATION.md`, `README.md`, and `AGENTS.md`.
  - Run `./scripts/simulated_user_gate.sh` to ensure zero regressions across CLI and documentation checks.

- [x] **Phase 5 — Attribution Commit**:
  - Commit changes with subject: `feat(docs): establish canonical migration guide & governance gate (goals/version-migration-guide/goal.md)`
