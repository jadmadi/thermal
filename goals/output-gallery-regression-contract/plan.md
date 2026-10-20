# Plan: Turn the output gallery into a reproducible consistency gate

## Execution Steps
- [x] **Task 1 — Fix and freeze the generator**: Commit a generator that uses a fixed fixture/time, fixes the validation command, and records fixture hash/timestamp. Acceptance: regenerating twice produces identical totals/schemas; the validation case exits 1.
- [x] **Task 2 — Add consistency matrix and row schemas**: Generate the surface matrix and row-level JSON schema excerpts; assert allowed equalities/bases. Acceptance: gate fails on a deliberately missing unclassified field/alias.
- [x] **Task 3 — Add width/color and layout**: Capture 80/120 static and live frames, compact+appendix layout, all variant TOC links, truncated share URL. Acceptance: gallery renders within the expected width; full share URL recoverable.
- [x] **Task 4 — Add the gallery gate**: Implement `scripts/output_gallery_gate.sh` with normalized checks for stale commands, totals, schemas, help/chart/heatmap/live regressions. Acceptance: gate passes on the fixed fixture and fails on a seeded regression.
- [x] **Task 5 — Docs, metadata, and attribution commit**: Update docs/PROMISES/README/project.jsonc/AGENTS; run full gate; commit with `chore(docs)` and the goal token. Acceptance: gate clean; commit ends in `(goals/output-gallery-regression-contract/goal.md)`.

## Dependencies and Handoff
All product fixes are prerequisites; this goal must not change product behavior. If the normalized gate finds a new discrepancy, open a new goal rather than weakening the assertion.

## Verification
```bash
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
./scripts/output_gallery_gate.sh
```
