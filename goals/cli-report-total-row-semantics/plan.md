# Plan: Make report total rows contain only additive, dimension-consistent values

## Execution Steps
- [x] **Phase 1 — Totals data model**: Add `projectDays`/`modelDays` and `toolCount` to the report totals, and retain the distinct-day set in the project/model aggregation. Acceptance: `TestProjectTotalsDistinctDays`, `TestModelTotalsDistinctDays`, and `TestTotalsJSON` pass.
  - Distinct calendar days for totals; per-row days unchanged.
  - `toolCount` from the union size.
  - Totals independent of `--top`.
- [x] **Phase 2 — Total-cell renderer**: Add `internal/render/totals.go` with typed cells; replace the ad-hoc projects/models total rows. Acceptance: `TestTotalRowAdditivity`, `TestTotalRowBlankCells`, and golden rows pass.
  - Project/Model cells blank, not `—`.
  - Tools cell `N tools`.
  - Day cells distinct.
  - Last cell max.
- [x] **Phase 3 — Period and mix verification**: Add additivity tests for daily/weekly/monthly and recomputation tests for mix. Fix any discrepancy found. Acceptance: `TestPeriodTotalsAdditive` and `TestMixTotals` pass.
  - Numeric columns equal the row sum.
  - Active days distinct.
  - Mix percentages sum to 100 ± 1.
- [x] **Phase 4 — Footers, docs, gate, and closure**: Add the P-Days/M-Days footnote, README/MIGRATION/project.jsonc, the simulated-user-gate section, and close the goal.
  - Record the overlap fixtures and additivity results in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(render): make report totals additive and dimension-consistent (goals/cli-report-total-row-semantics/goal.md)`.
  - Certify with `sila goals certify cli-report-total-row-semantics --evidence=goals/cli-report-total-row-semantics/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | projects total Tools | `N tools` or blank, no `·` |
| AC-002 | overlap project fixture | activeDays 1, projectDays 2, toolCount 2 |
| AC-003 | overlap model fixture | activeDays 1, modelDays 2, toolCount 2 |
| AC-004 | additive columns | row sums match |
| AC-005 | mix totals | volume sum, 100 ± 1 |
| AC-006 | footers | per-project/per-model note |
| AC-007 | blank cells | spaces, not `—` |
| AC-008 | JSON totals | projectDays/modelDays/toolCount |
| AC-009 | `--top 3` totals | unchanged |
| AC-010 | `go test -race ...` | exit 0 |
| AC-011 | gate | exit 0 |

## Rollback
- The JSON additions are additive; keep them even if the text renderer is reverted.
- If the distinct-day computation is too invasive, keep `activeDays` as the sum and add `distinctActiveDays`; do not print the sum in the total row without a label.
- Never restore the tool union in a total row.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-report-total-row-semantics/plan.md`; (2) commit with `(goals/cli-report-total-row-semantics/goal.md)` using a `fix(render):` subject; (3) certify via `sila goals certify cli-report-total-row-semantics --evidence=goals/cli-report-total-row-semantics/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
