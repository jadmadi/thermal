# Plan: Reconcile report totals, unclassified tokens, cost bases, windows, and share scope

## Execution Steps
- [ ] **Task 1 — Add basis and unclassified fields**: Extend report/stat/model/share types with classified/unclassified, cost-basis, active/costed-day, and scope/window fields; add JSON tags. Acceptance: `go test -race ./internal/thermal -run 'ReconcileTypes|Unclassified'` and a temporary JSON decode of each report show the fields.
- [ ] **Task 2 — Fix stats cost day accounting**: Track active and costed days separately; choose zero-cost bin vs costed label; update text/JSON. Acceptance: fixture with one zero-cost day and one unpriced day shows 96 active/80 costed or 96 active with a zero bin; means document their denominator.
- [ ] **Task 3 — Canonicalize cost estimates**: Route leaderboard/daily/projects/models/mix through one estimator or emit basis-specific reconciliation lines; add the fixture equality test. Acceptance: leaderboard estimated == daily estimated within 0.0001; projects total == daily total within 0.0001; models/mix are labeled estimate-only.
- [ ] **Task 4 — Fix models and stats unclassified output**: Add the unclassified row/line and JSON fields; label composition as classified-only. Acceptance: models text/JSON total = daily total with explicit classified+unclassified; stats composition includes the gap; no percentage denominator ambiguity.
- [ ] **Task 5 — Reconcile projects and share scope**: Add project unattributed attribution and canonical ShareSnapshot; make activeDays/volume/spend one scope. Acceptance: projects total reconciles to daily; share activeDays/volume/spend share a scope and JSON fields are non-null.
- [ ] **Task 6 — Add cross-surface regression**: Add a fixture-driven `TestCrossSurfaceTotals` covering all surfaces and bases; fail on silent totals. Acceptance: test passes and would fail if any surface drops unclassified/basis fields.
- [ ] **Task 7 — Docs, metadata, and attribution commit**: Update docs/MIGRATION, README, project.jsonc; run full gate; commit with the goal token. Acceptance: migration has six parts; full gate passes; commit subject ends in `(goals/output-report-totals-reconciliation/goal.md)`.

## Dependencies and Handoff
The JSON goal consumes the new basis/unclassified/share fields. The live goal consumes canonical model-cost semantics. The share/statusline goal consumes the ShareSnapshot. Do not implement chart/help/heatmap work here.

## Verification
```bash
go test -race ./internal/thermal ./internal/render -run 'Reconcile|Unclassified|StatsCost|ProjectCost|ModelCost|ShareScope'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
