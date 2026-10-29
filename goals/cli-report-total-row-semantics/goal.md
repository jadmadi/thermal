# Goal: Make report total rows contain only additive, dimension-consistent values

## Goal Description
Priority: P2 (output semantics). The projects and models reports reuse the row struct as the totals struct, so the total row inherits fields that are not additive. `sumProjects` sums `ActiveDays` across projects and unions the `Tools` list; the renderer then prints that union in the Tools cell. The result is a total row that looks like a row but means something different in two of its seven columns.

Observed on the current source build (2026-10-29):

- `thermal projects` total row:
  `│ Total    Agy · CMD +10    23.3B   ~$503.59   505   2026-10-29 │`
- `thermal models` total row:
  `│ Total    CMD · Claude +9   22.0B   ~$614.41   190   2026-10-29 │`
- The leaderboard reports 96 active calendar days; the projects total says `505` P-Days and the models total says `190` M-Days. Both are sums of per-row active days, so a calendar day can be counted once per project or model.
- `internal/thermal/projects.go:197-229` (`sumProjects`) sums `ActiveDays`, unions `Tools`, takes the min `FirstDay`, and the max `LastDay`.
- The models aggregation follows the same pattern.
- Period reports (`daily`, `weekly`, `monthly`) already deduplicate active days (`internal/thermal/report.go:322`), so their totals are correct.
- The projects footer (`Total = $242.09 recorded + ~$261.50 estimated`) is correct and should be preserved.

The outcome of this goal is a **typed total-row contract**: every cell is explicitly `sum`, `distinct`, `min`, `max`, `count`, `recomputed`, or `blank`; the renderer prints only values with the right semantics; JSON exposes both the deduplicated calendar days and the per-row day sum when both are useful; and tests prove that additive columns equal the sum of rows.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 252
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3 (coordinate with `output-report-totals-reconciliation` and the width goal)
- **Related implemented goals (do not regress)**: `output-report-totals-reconciliation` owns numeric reconciliation across reports; this goal owns the *semantics of the total row itself*. `report-period-totals` owns period rows/totals; this goal verifies them and does not change their aggregation.

## Files to Touch
- `internal/thermal/types.go`
- `internal/thermal/projects.go`
- `internal/thermal/models.go`
- `internal/render/projects.go`
- `internal/render/models.go`
- `internal/render/totals.go`
- `internal/thermal/projects_test.go`
- `internal/thermal/models_test.go`
- `internal/render/projects_test.go`
- `internal/render/models_test.go`
- `cmd/thermal/simulated_user_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-report-total-row-semantics/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-report-total-row-semantics/plan.md`](plan.md)

## Requirements

1. **Cell semantics table (frozen).**
   | Report | Column | Total semantics |
   | --- | --- | --- |
   | projects | Project | blank |
   | projects | Tools | count (`N tools`) or blank |
   | projects | Tokens | sum |
   | projects | Cost | sum of recorded + estimated (footer) |
   | projects | P-Days | distinct calendar active days |
   | projects | Last | max (latest day) |
   | models | Model | blank |
   | models | Tools | count (`N tools`) or blank |
   | models | Tokens | sum |
   | models | Cost | sum (estimated) |
   | models | M-Days | distinct calendar active days |
   | models | Last | max |
   | daily/weekly/monthly | Period | `Total` label |
   | daily/weekly/monthly | numeric columns | sum; active days distinct |
   | mix | shares | recomputed from the total volume |
   | leaderboard | — | no total row |
   - Blank cells render as spaces, not `—` (the em dash is reserved for unpriced cost).
   - The `Tools` count uses the union size, not the number of row entries.

2. **Projects total aggregation.**
   - `sumProjects` must compute the set of distinct calendar days across all project rows and set the total's `ActiveDays` to its size.
   - Keep the per-row `ActiveDays` (P-Days) unchanged; it remains per-project.
   - Add `ProjectDays` to the report totals as the sum of per-row `ActiveDays`, for callers who need the workload sum.
   - Add `ToolCount` as `len(union(Tools))`.
   - Keep `FirstDay` as min and `LastDay` as max; neither is printed in the total row except `Last`.
   - The footer must state: `P-Days are per project; the total uses distinct calendar days (N project-days across M projects).` Use the actual numbers.

3. **Models total aggregation.**
   - Same as projects with `ModelDays` and `ToolCount`; `ActiveDays` is the distinct calendar-day count.
   - Cost remains the estimated sum; the existing footer explaining that recorded cost is not attributable to a model stays.
   - `MissingPricing` and `UnattributedTokens` totals remain sums/unions as appropriate.
   - Footer: `M-Days are per model; the total uses distinct calendar days (N model-days across M models).`

4. **Period and mix totals.**
   - `daily`, `weekly`, `monthly`: every numeric cell equals the sum of the displayed rows (within integer/rounding tolerance); `ActiveDays` equals the number of distinct days in the window.
   - `mix`: percentages are recomputed from the total volume and sum to 100% within rounding; the `Total` volume equals the sum of row volumes.
   - `trend`/`stats` have no total row; do not add one.
   - `leaderboard` has no total row; do not add one.

5. **Renderer contract.**
   - Replace ad-hoc total-row construction with a shared `TotalCell` helper that carries the semantic (`sum`, `distinct`, `count`, `max`, `blank`, `recomputed`).
   - The renderer must not call `toolsCell(totals.Tools, …)` for a total row.
   - The total row aligns exactly with the header and body; blank cells keep their width.
   - At narrow widths, the width goal's drop order applies; a dropped column is dropped from the header, body, and total together.
   - The total row is visually distinct (existing color/bold) but not color-only.

6. **JSON contract (additive).**
   - Add to `ProjectReport.Totals` and `ModelReport.Totals`:
     ```json
     {"activeDays": 96, "projectDays": 505, "toolCount": 13}
     {"activeDays": 96, "modelDays": 190, "toolCount": 10}
     ```
   - Keep the existing `activeDays` key but change its meaning to distinct calendar days; this is a semantic change and must be documented in MIGRATION.
   - Keep the per-row `activeDays` meaning unchanged (per-project/per-model days).
   - If `tools` remains in totals, it is the union (documented); the renderer uses `toolCount`.
   - Do not rename `totalTokens`, `estimatedCost`, or other existing keys.

7. **Tests.**
   - A fixture with two projects active on the same day: each row `activeDays=1`, total `activeDays=1`, `projectDays=2`, `toolCount=2`.
   - The model equivalent.
   - Table parsing: sum every additive numeric column and compare to the total row.
   - Distinct day count: compare to the union of the fixture days.
   - Cost arithmetic: recorded + estimated = shown.
   - Blank-cell rendering: no `—` in the total row's Project/Model cell.
   - `--top` interaction: totals remain full-window, not the capped rows (coordinated with the `--top` goal).
   - JSON snapshot for totals.
   - Mix percentage sum and daily/weekly/monthly additive columns.

8. **Documentation and migration.**
   - README report section explains the total row and P-Days/M-Days.
   - MIGRATION six-part entry: total `activeDays` now counts distinct calendar days; `projectDays`/`modelDays` added; the Tools cell is a count; detection (`thermal projects --json | jq .totals.activeDays`), fix, escape hatch (use `projectDays`).
   - `project.jsonc` updated.
   - `scripts/simulated_user_gate.sh`: assert AC-001, AC-002, AC-004, and AC-005.

## Non-Goals
- Changing per-row P-Days/M-Days semantics.
- Changing cost calculations or pricing.
- Adding a total row to the leaderboard, stats, trend, or audit.
- Changing the width/layout engine.
- Renaming existing JSON keys other than documenting the `activeDays` semantic change.

## Acceptance Criteria
- AC-001: `thermal projects --no-color` total row's Tools cell matches `^[0-9]+ tools$` or is blank; it contains no `·` and no `+N`.
- AC-002: In a two-projects-same-day fixture, `thermal projects --json | jq -e '.totals.activeDays==1 and .totals.projectDays==2 and .totals.toolCount==2'` exits `0`.
- AC-003: `thermal models --no-color` total row's Tools cell is a count or blank; a two-models-same-day fixture yields `activeDays==1`, `modelDays==2`, `toolCount==2`.
- AC-004: For `daily`, `weekly`, `monthly`, and `projects`, every additive numeric column equals the sum of the displayed rows (integer exact; cost within $0.01).
- AC-005: For `mix`, the total volume equals the sum of row volumes and the total percentages sum to 100 ± 1.
- AC-006: The projects footer contains `per project`, `distinct calendar days`, and the project-day count; the models footer contains the model equivalent.
- AC-007: Total-row blank cells are spaces, not `—`.
- AC-008: `thermal projects --json` contains `totals.projectDays` and `totals.toolCount`; `thermal models --json` contains `totals.modelDays` and `totals.toolCount`.
- AC-009: `thermal projects --top 3 --json | jq '.totals'` equals the uncapped totals (totals are full-window).
- AC-010: `go test -race ./internal/thermal/... ./internal/render/... ./cmd/thermal/...` exits `0`, including `TestProjectTotalsDistinctDays`, `TestModelTotalsDistinctDays`, `TestTotalRowAdditivity`, `TestTotalRowBlankCells`, and `TestMixTotals`.
- AC-011: `./scripts/simulated_user_gate.sh` exits `0` and includes the total-row section.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Two projects same day | total activeDays 1, projectDays 2 |
| Two models same day | total activeDays 1, modelDays 2 |
| One project | activeDays == projectDays |
| No projects | no total row / zero report message |
| `--top 3` | totals unchanged |
| Blank Project cell | spaces, not `—` |
| Tools union | count only in the total row |
| Last dates | max, not sum |
| Cost | recorded + estimated = shown |
| Mix rounding | 100 ± 1 |
| Daily total | sum of rows, distinct active days |
| Narrow width | dropped columns dropped from all three sections |

## Metrics & Regression Guards
- Additivity test: 100% of additive numeric columns equal the row sum.
- Distinct-day test: total activeDays equals the union of fixture days, not the row sum.
- Blank-cell grep: total row contains no `—` in name/tool cells.
- JSON contract: `projectDays`/`modelDays`/`toolCount` locked by a snapshot.
- Gate: simulated user gate asserts AC-001, AC-002, AC-004, and AC-005.

## Risks & Mitigations
- **Risk:** Changing `totals.activeDays` from 505 to 96 breaks consumers expecting the sum. **Mitigation:** add `projectDays`/`modelDays`, document the semantic change in MIGRATION, and call it out in the changelog.
- **Risk:** Computing distinct days requires retaining a day set in the aggregation, increasing memory. **Mitigation:** the report already iterates rows with day data; use a `map[string]struct{}` sized by active days, bounded by the report window.
- **Risk:** The total row becomes mostly blank at narrow widths. **Mitigation:** the width goal drops columns consistently; the count and distinct-day columns are prioritized.
- **Risk:** The footer becomes long. **Mitigation:** one compact line, wrapped by the width renderer.
- **Risk:** The tools count differs from the union list shown elsewhere. **Mitigation:** document `toolCount` as the union size and test it against the row unions.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "projects total P-Days sum"` and `sila recall "total row tool union"`.
- **Writes**: record a `case` for the "typed total-cell semantics" mechanism, grounded to `internal/render/totals.go`; record a `lesson` for the non-additive total row.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "total-row semantics shipped" --open="JSON activeDays migration notes"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-report-total-row-semantics/goal.md)` under a `fix(render):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the overlap fixtures and additivity results.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
