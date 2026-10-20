# Goal: Turn the output gallery into a reproducible consistency gate

## Goal Description
Priority: P2 (regression infrastructure). The current gallery is already useful and caught the stale 9-box hint, but it is not reproducible or trustworthy as a consistency oracle: captures drift between sections (daily 38.0B vs monthly 38.1B, live 225M), the “report flag without report” validation case actually runs `thermal daily --since` and exits 0 with a duplicate daily report, the top-level JSON table hides row aliases/floats, and there are no 80/120 or color captures. Turn the artifact into a fixed-fixture generator plus a consistency gate.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-formatting-primitives, output-report-totals-reconciliation, json-output-contract-hardening, help-supported-tools-and-flags, heatmap-month-label-overlap, chart-scale-order-and-top-contract, live-ticker-unit-and-rollup-contract, audit-path-and-score-clarity, statusline-and-share-scope-contract
- **Sequence**: 209
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `docs/OUTPUT-GALLERY.md`
- `scripts/output_gallery.sh`
- `scripts/output_gallery_gate.sh`
- `scripts/simulated_user_gate.sh`
- `docs/PROMISES.md`
- `docs/MIGRATION.md`
- `README.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Fix the validation case**: Replace the `thermal daily --since 2026-01-01` capture in the “report flag without report” section with `thermal` at the root plus `--since 2026-01-01` (which should exit 1). Remove the duplicate full daily output. Add a generator assertion for the expected exit code and stderr.
2. **Fixed fixture/single snapshot**: Generate the gallery from a hermetic fixture (reuse the simulated-gate fixture or a dedicated `scripts/output_gallery_fixture.sh`) or pass a fixed `--db`/home and freeze time. Every section must come from the same snapshot; record a capture timestamp and fixture hash in the header. No mixed live data (daily 38.0B vs monthly 38.1B).
3. **Consistency matrix**: Add a required table with one row per surface (leaderboard, daily, weekly, monthly, projects, models, mix, stats token, stats cost, trend token, trend cost, share, live) and columns for tokens, recorded cost, estimated cost, total cost, active/costed days, window/scope, and cost basis. The generator/gate must fail if a surface lacks a basis or if an allowed equality breaks.
4. **Row-level JSON schemas**: Add schema excerpts for leaderboard results, daily rows, project rows, model rows, live events/rollups, and share snapshot. Include the canonical keys, types, precision, alias policy, and unclassified/costKind fields. The top-level envelope table alone is insufficient.
5. **Width/color coverage**: Capture text views at 80 and 120 columns (and the live TUI at 80x24/120x40), with a note that ANSI color is excluded. Add a colored raw sample or a separate `--color` appendix; at minimum, test the `--no-color` and terminal-width contracts.
6. **Layout/readability**: Keep a compact first screen (headings, headers, first rows) and move full 116-row projects/models tables into appendices. Add TOC entries for every variant (daily/projects/models/mix/stats/trend), not only daily. Truncate the share URL in the body and keep the full URL in an appendix/JSON.
7. **Reproducible gate**: Add a script and a CI-friendly command that regenerates the gallery and fails on: stale retired commands, mixed fixture hashes, mismatched totals, missing unclassified/basis fields, JSON aliases/floats, help tool-count mismatch, `ONov`/chart scale regressions, live unit truncation, and empty validation cases. The gate may reuse `scripts/simulated_user_gate.sh` checks.
8. **Golden/contract tests**: Add or reference golden tests for the gallery sections that encode the fixes; the gate must fail if the generated markdown differs from a checked-in baseline in a semantic field. Do not make the markdown a byte-for-byte golden if totals/date are dynamic; compare normalized values and schema.
9. **Docs/metadata**: Update docs/PROMISES, README, project.jsonc, and AGENTS.md output-gallery references. Add a six-part migration entry only if user-visible CLI contracts change; otherwise document the gallery as a test artifact.
10. **Attribution**: Use a `chore(docs): ... (goals/output-gallery-regression-contract/goal.md)` subject (docs-only commits are ignored by Sila; `chore` is accepted). Run `sila goals`; do not push.

## Done Condition
1. The gallery is generated from one fixed snapshot; the header records fixture/time; no section uses a different live total.
2. The validation case runs the actual no-report command and exits 1; there is no duplicate daily block.
3. The consistency matrix, row-level schemas, 80/120 captures, compact/appendix layout, and TOC for every variant exist; the share URL is truncated in the body with the full value recoverable.
4. The gate fails on stale commands, mixed totals, missing basis/unclassified fields, aliases/floats, help/chart/heatmap/live regressions, and empty validation cases; the checked-in normalized baseline passes.
5. Full race suite, build, simulated gate, and the new gallery gate pass; docs/metadata updated; the attribution commit ends in `(goals/output-gallery-regression-contract/goal.md)`.

## Scope Boundary
Gallery/generator/gate and docs only. Do not change product behavior; if the gallery finds a new product bug, create a new goal. The product fixes are owned by the prerequisite goals.
