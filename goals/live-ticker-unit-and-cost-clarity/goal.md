# Goal: Separate token and step ticker columns and make live cost precision explicit

## Goal Description
The live ticker's `TOK/STEP` column prints `+567.9K tok` on one row and `+10.5K step` on the next, so rows cannot be compared and it is unclear whether steps are included in token sums. In the KPI grid, `Turns: 12.0K turns (99% hit)` repeats the same 11,957 step count already shown in `Volume` and uses the ambiguous word "hit". The cost column collapses every positive sub-cent value to `<$0.01` and every zero to `$0.00`, conflating tiny spend, explicit free, and unpriced sources.

Replace the mixed column with dedicated `TOKENS` and `STEPS` columns, show the aggregate turn/step count once, label cache hit explicitly, use the established sub-cent/`—` cost convention, and flag median outliers with `!`. This is a presentation and event-status goal; token math must stay unchanged.

Priority: **P1**. Derived from the 2026-10-19 live-monitor feedback (bugs 5, 6, 8; improvements "Normalize units", "Cost precision", "Outlier highlight", "Rename").

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: live-baseline-separation
- **Sequence**: 102
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)
- Prerequisite: live-baseline-separation
- Cost precedent: `internal/render/leaderboard.go:210-222` (`$0.0000` sub-cent, `—` absent)

## Requirements
1. **Separate ticker columns**: Replace the `TOK/STEP` header with two columns, `TOKENS` and `STEPS`. Token rows print a token delta only in `TOKENS`; activity rows print a step delta only in `STEPS`; a row that has no value for a column prints `—`. Values are right-aligned and carry the `+` delta marker; the header, not every cell, carries the unit. Remove the `TOK/STEP` label entirely.
2. **Responsive column priority**: At wide widths (>=110 cells) show TIME, TOOL, TOKENS, STEPS, MODEL, PROJECT, COST. As width shrinks, drop PROJECT first and then MODEL before shrinking TOKENS/STEPS; at 80 columns the minimum set is TIME, TOOL, TOK, STEP, COST. Never merge token and step values into one cell. Add width tests at 80, 100, and 120 cells.
3. **Show the aggregate once**: In the KPI grid, Today shows `Volume: X tok` for token data, `Steps: N step` for activity-only data, and one hybrid line `Volume: X tok · N step`; do not also render `Turns: N turns`. Replace the old Turns line with `Cache hit: Z%` (or `Cache hit: —` when unavailable). Session shows `Burned: +X tok` or `Steps: +N`; it must not repeat the same count. Pick one canonical label per mode and use it consistently in the ticker and KPI grid.
4. **Cache-hit wording**: The word "hit" must never appear without "cache". Full form: `Cache hit: 99%`; compact form must retain `Cache` (e.g. `99% cache` or `Cache 99%`), never the old `99% hit`. Activity-only tools show `Cache hit: —`.
5. **Cost semantics**: Add `CostKind` to `LiveEvent` with values `recorded`, `estimated`, and `unpriced` (JSON `costKind`), and add `UnpricedTokens`/`UnpricedEvents` to `LiveSnapshot` when the pricing path reports missing models or `--no-estimate` suppresses estimation. The ticker and KPI cost formatter must follow the leaderboard convention:
   - `unpriced` -> `—`
   - recorded `$0` only when the source explicitly records zero -> `$0.00`
   - `0 < cost < 0.01` -> four decimals (`$0.0042`, `$0.0007`); values below `0.00005` -> `<$0.0001`
   - estimated costs keep the `~` prefix
   - never emit `<$0.01` in the live view
6. **Outlier flag**: Compute the median of visible, non-baseline token deltas and step deltas separately. With at least three positive rows, mark a row with `!` when its value is strictly greater than 3x the median; add a legend `! >3x median`. Baseline/carried-in rows are excluded from the median and never flagged. If the median is zero, do not flag. Use a named constant for the 3x factor so a later goal can expose it.
7. **Preserve data math**: Do not change loader ingestion, token-type disjointness, cache-hit calculation, or session/baseline semantics. `CostKind=unpriced` must not turn unpriced tokens into zero cost in `TodayCost`; it only changes display and adds the unpriced count.
8. **Rendering safety**: Keep ANSI width padding, `--no-color` output plain, and tests at 80x24 and 120x40. No new dependencies.

## Done Condition
1. At 120x40 the ticker header contains `TOKENS` and `STEPS` and no `TOK/STEP`; a hybrid view shows a token delta in TOKENS, a step delta in STEPS, and `—` in the non-applicable cell.
2. At 80x24 the header still contains separate `TOK` and `STEP` columns and every row fits without ellipsis.
3. The view contains `Cache hit:` and never a bare `hit`; it contains the step count once, not once in `Volume` and again in `Turns`.
4. `formatLiveCost` returns `$0.0042` for 0.0042, `—` for unpriced, `$0.00` for explicit zero, and `<$0.01` is absent from all live output.
5. A row at 4x the median is marked `!`, while baseline and sub-3-row sets are not flagged.
6. `thermal live --json` exposes `costKind` and unpriced counts without changing existing token totals; migration documentation and `project.jsonc` are synchronized.
7. Focused tests, full race suite, CLI build, and simulated user gate pass; the attribution commit ends in `(goals/live-ticker-unit-and-cost-clarity/goal.md)`.

## Scope Boundary
No one-second bucket/Peak changes (`live-second-bucket-accounting`), no flame ceiling policy (`live-flame-scale-contract`), and no rollup/sort toggle (`live-ticker-rollup-and-cost-sort`). This goal owns the ticker schema and cost semantics those goals reuse.
