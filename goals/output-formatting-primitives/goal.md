# Goal: Unify output formatting semantics across every Thermal surface

## Goal Description
Priority: P0 (foundation). The gallery shows one symbol doing three jobs: `~` means estimated cost in the leaderboard, approximate volume in trend (`~97.5M`), approximate context tax in audit (`~45.4K`), and a mixed total in daily/share (`~$20596.53`). Cost markers are also inconsistent: the leaderboard prefixes estimates with `~`, while daily/projects/models rows print `$...` and explain the mixture only in a footer. Tool lists use `OpenCode,Devin` in projects/models but `·` elsewhere. Windows say `Window: 96 days` for active days, while trend counts 390 calendar days. Define one formatting contract and shared helpers, then migrate every renderer to it without changing aggregation.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 200
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/render/format.go`
- `internal/render/report.go`
- `internal/render/leaderboard.go`
- `internal/render/projects.go`
- `internal/render/models.go`
- `internal/render/analytics.go`
- `internal/render/statusline.go`
- `internal/render/audit.go`
- `internal/share/share.go`
- `internal/render/report_test.go`
- `internal/render/leaderboard_test.go`
- `internal/render/projects_test.go`
- `internal/render/models_test.go`
- `internal/render/analytics_test.go`
- `internal/render/statusline_test.go`
- `internal/render/audit_test.go`
- `internal/thermal/format.go`
- `internal/thermal/format_test.go`
- `docs/MIGRATION.md`
- `README.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Cost formatter**: Add one shared `formatCost` family for text and one JSON rounding helper. Required forms: recorded zero `$0.00`; recorded >= 0.01 `$X.XX`; recorded sub-cent `$0.0000`; estimated `~$X.XX` / `~$0.0000`; absent or unpriced `—`; partial-known live group `≥$X.XX`; mixed recorded+estimated totals must explicitly state both components or carry a legend, never a bare `~` on a total whose recorded portion is non-zero.
2. **Tilde contract**: Reserve `~` for estimated money only. Replace non-cost approximations with `≈` or an explicit word: trend volume mean becomes `≈97.5M` or `Mean 97.5M`; audit context tax becomes `approx. 45.4K tokens/turn`; daily/share mixed totals show `$X.XX (recorded + est.)` or a two-part line. Add a table-driven test that rejects `~` in non-cost formatters.
3. **Window formatter**: Add `formatWindow`/`windowLabel` that distinguishes active-period counts from calendar spans. Default daily/weekly/monthly use `all time (96 active days)`, `all time (17 active weeks)`, `all time (6 active months)`; explicit `--last` uses `last 7 days`; trend uses `390 calendar days (2025-09-26 to 2026-10-20)`; projects/models use `project-days`/`model-days` for summed day columns. No bare `Window: 96 days`.
4. **Tool-list join**: Add `joinTools` and use ` · ` consistently. Projects/models must render `OpenCode · Devin +8`, not `OpenCode,Devin +8`; the ticker may keep `TOOL×N` but must not introduce comma-only lists. Preserve width truncation.
5. **Precision contract**: One decimal for cache-hit percentages and share percentages; two decimals for ordinary dollar amounts; four decimals for sub-cent dollars; compact numbers for token/step counts; `$0.00` (not `$0.0000`) for zero cost boundaries; percentages in mix stay integers only if the legend says they are rounded. Add golden assertions for each precision class.
6. **Shared implementation**: Move duplicate `formatCost`/cost-prefix logic out of `internal/render/report.go` and `internal/render/leaderboard.go` into the shared helper file. Do not duplicate a second copy in tests. Keep `thermal.CompactNumber` as the count formatter; add only the missing money/percent/window/tool helpers.
7. **Surface migration**: Migrate leaderboard, daily/weekly/monthly, projects, models, mix, stats, trend, audit, statusline, share, and live cost/unit captions to the shared helpers. JSON serialization is owned by json-output-contract-hardening; this goal only defines the text policy and may add pure helpers the JSON goal reuses.
8. **Regression tests**: Table-driven tests for cost kinds, sub-cent thresholds, mixed totals, window labels, tool joiners, cache precision, and `$0.0000` boundaries. Golden snippets must assert no `~97.5M`, no `OpenCode,Devin`, no bare `Window: N days`, and no `$0.0000` for a zero boundary.
9. **Docs**: Add a six-part `docs/MIGRATION.md` entry for the formatting semantics. Add a short README legend covering `$`, `~$`, `—`, `≥$`, `≈`, and active-window labels. Update `project.jsonc` `lastEdit` and any affected feature wording.

## Done Condition
1. All text surfaces use the shared formatter; a grep/AST test shows no local `formatCost` copy remains in the migrated files.
2. `thermal trend` uses `≈` or an explicit mean label, never `~` on a volume; `thermal audit` uses `approx.` for token tax; daily/share mixed totals state recorded + estimated or carry an explicit legend.
3. `thermal daily` default says `all time (N active days)`; `--last 7` says `last 7 days`; trend says `calendar days`; projects/models day columns say `project-days`/`model-days`.
4. Projects/models tool lists use ` · `; cache precision is one decimal; zero-cost boundaries print `$0.00`; sub-cent rows still print four decimals.
5. Focused render/format tests, full `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh` pass; `docs/MIGRATION.md` and `project.jsonc` are synchronized; the attribution commit ends in `(goals/output-formatting-primitives/goal.md)`.

## Scope Boundary
Formatting policy and shared helpers only. Do not change token/cost aggregation, JSON field names, report row sets, chart scaling, heatmap label placement, or live bucket semantics. The reconciliation goal owns totals; the JSON goal owns serialization shape.
