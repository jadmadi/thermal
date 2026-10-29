# Plan: Name denominators and intervals and remove run-on or zero-noise report prose

## Execution Steps
- [ ] **Phase 1 — Stats basis**: Add the `Per active day` block, the `Mean / active day` label, and the `Basis: N active days` line; keep costed-day semantics for `--metric cost`; add `meanBasis` to JSON. Acceptance: `TestStatsBasisLabel` and the 4/10 fixture pass; the arithmetic is unchanged.
- [ ] **Phase 2 — Trend basis and projection**: Add `Mean / calendar day`, the basis line with active/zero counts, `Point estimate`, the `95% band (±1.96 residual SD)` label, and the method footnote; add `meanBasis`, `activeDays`, `bandMethod`, `residualStd` to JSON. Acceptance: `TestTrendBasisLabel` and `TestTrendProjectionLabels` pass.
- [ ] **Phase 3 — Outliers and histogram**: Replace the self-referential outlier sentence with a total count and shown count; omit zero-count bins and report the omitted count. Acceptance: `TestOutlierSentence` (0/5/27) and `TestHistogramOmitsEmptyBins` pass.
- [ ] **Phase 4 — Audit prose**: Split the score components, largest penalty, and biggest-win lines; strip/splice severity correctly; add the PASS/WARN/CRIT legend. Acceptance: `TestAuditBiggestWinLine` and `TestAuditLegend` pass; no `Biggest win: Critical:` sequence.
- [ ] **Phase 5 — Gate, docs, and closure**: Add the simulated-user-gate prose section, README analytics section, MIGRATION, project.jsonc, and close the goal.
  - Record the 4/10 fixture math and the prose captures in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(render): label analytic denominators and clean report prose (goals/cli-analytics-and-audit-prose-clarity/goal.md)`.
  - Certify with `sila goals certify cli-analytics-and-audit-prose-clarity --evidence=goals/cli-analytics-and-audit-prose-clarity/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | stats | `Mean / active day`, basis |
| AC-002 | trend | `Mean / calendar day`, basis |
| AC-003 | 4/10 fixture | stats 25, trend 10 |
| AC-004 | trend projection | point estimate, 95% band, footnote |
| AC-005 | stats outliers | `Outliers: 27 ...; top 5 shown` |
| AC-006 | histogram | no zero bins, omitted count |
| AC-007 | audit | separate lines, actionable biggest win |
| AC-008 | audit legend | PASS/WARN/CRIT |
| AC-009 | JSON bases | `meanBasis`, `activeDays` |
| AC-010 | JSON band | `bandMethod` |
| AC-011 | 80-col check | no long lines |
| AC-012 | `go test -race ...` | exit 0 |
| AC-013 | gate | exit 0 |

## Rollback
- The JSON additions are additive; keep them if the text is reverted.
- If the longer labels break the width budget, move the basis to a footnote rather than dropping it.
- Never restore the `all 5 top days above, plus` wording or the `Biggest win: Critical:` splice.

## Phase 6 — Attribution & Registry Closure
- [ ] **Phase 6 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-analytics-and-audit-prose-clarity/plan.md`; (2) commit with `(goals/cli-analytics-and-audit-prose-clarity/goal.md)` using a `fix(render):` subject; (3) certify via `sila goals certify cli-analytics-and-audit-prose-clarity --evidence=goals/cli-analytics-and-audit-prose-clarity/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
