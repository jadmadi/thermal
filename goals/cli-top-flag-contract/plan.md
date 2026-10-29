# Plan: Give --top one documented contract and make every remediation hint valid

## Execution Steps
- [ ] **Phase 1 — Applicability table and validation**: Replace the scattered `--top` checks with one `topApplies` table and one canonical error. Track whether `--top` was explicitly set for the changelog default. Acceptance: `TestTopApplicabilityMatrix` passes for all 14 commands; exit `2` for rejecting commands.
  - Message lists all seven accepting commands.
  - `--top -1` and `--top abc` are handled by the parser goal and re-asserted here.
  - `--top 0` = all; absent = default (3 for changelog).
- [ ] **Phase 2 — Row-count policy**: Add `EffectiveRows`, `ChartRows`, and the footer helpers. Apply the cap to the leaderboard, daily, weekly, monthly, projects, models, changelog, charts, and JSON `data`. Acceptance: `TestTopRowCounts`, `TestTopJSON`, `TestTopBreakdown`, and `TestTopOrderInteraction` pass.
  - Cap after sorting/ordering.
  - Totals unchanged.
  - Chart rows ≤ table rows.
  - Breakdown only for capped rows.
- [ ] **Phase 3 — Footer and changelog affordance**: Replace the chart footer literal with `ChartFooter`; unify table and chart counts; add the changelog `Showing N of M · thermal changelog --top M for all` header. Acceptance: `TestTopFooterMatrix` and `TestChangelogTop` pass; the old `raise --top to see them in the table` string is absent.
- [ ] **Phase 4 — Help, gate, docs, and closure**: Update the global and per-command help, README, MIGRATION, project.jsonc, and the simulated-user-gate section. Record the matrices in `evidence.md` and close the goal.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(cli): unify the --top contract and chart remediation (goals/cli-top-flag-contract/goal.md)`.
  - Certify with `sila goals certify cli-top-flag-contract --evidence=goals/cli-top-flag-contract/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `thermal --top 5` | exit 0, 5 rows |
| AC-002 | `daily --top 5` | 5 rows, totals unchanged |
| AC-003 | weekly/monthly/projects/models | caps work |
| AC-004 | changelog top 5/0 | 5/all releases, header |
| AC-005 | stats top 5 | exit 2, accurate list |
| AC-006 | daily --chart footer | valid advice only |
| AC-007 | daily --chart --top 5 | 5 rows both, footers agree |
| AC-008 | `--top 5 --json` | data length 5, totals full |
| AC-009 | `--top -1` | exit 2 |
| AC-010 | help | applicability list |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | gate | exit 0 |

## Rollback
- If enabling `--top` on a report causes a regression, keep the cap for projects/models/changelog and change the chart footer to omit `--top`; do not restore the invalid advice.
- Keep the applicability table and the canonical validation message even if a command is temporarily excluded.
- Do not change totals to reflect capped rows; that would be a data regression.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-top-flag-contract/plan.md`; (2) commit with `(goals/cli-top-flag-contract/goal.md)` using a `fix(cli):` subject; (3) certify via `sila goals certify cli-top-flag-contract --evidence=goals/cli-top-flag-contract/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
