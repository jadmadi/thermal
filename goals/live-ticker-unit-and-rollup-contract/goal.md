# Goal: Fix live ticker units, truncation, sparkline glyphs, and model-cost semantics

## Goal Description
Priority: P1 (live readability/trust). The 100x30 frame shows `Rate: 3.4M/m · 23 step/m`, `Speed: 56.7K · 0.383 step…`, and `Burned: +1.7M tok (+60 s)`. The token unit disappears from rate/speed, the speed cell truncates mid-unit, and `s` is ambiguous between seconds and steps. The sparkline uses spaces for zero cells but dots for the all-zero state. Live JSON shows `todayCost: 1.977` but `todayModels[0].cost: 3.954` because model cost is estimated while today/project cost is recorded; the UI must label that basis, not present two competing totals.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-formatting-primitives, output-report-totals-reconciliation
- **Sequence**: 206
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Canonical live units**: Rate must render tokens as `tok/m` and steps as `step/m`; Speed must render `tok/s` and `step/s`. Never emit a bare `K`/`M` value without a unit in the hybrid line. Example target: `Rate: 3.4M tok/m · 23 step/m`; `Speed: 56.7K tok/s · 0.383 step/s`.
2. **No mid-unit truncation**: `padCell` may truncate model/project labels, but never a unit suffix. If the speed cell is too narrow, drop the secondary step speed first, then use a compact form such as `56.7K/s` with a legend; never render `step…` or `tok…`.
3. **Burned/steps wording**: Replace `Burned: +1.7M tok (+60 s)` with `Burned: +1.7M tok (+60 step)` (or a dedicated Steps column), and use `step` consistently in carried-in, session, and ticker contexts. `s` may mean seconds only in duration contexts.
4. **Sparkline empty-state glyph**: Use one glyph for absent buckets in a populated sparkline (prefer `·`) and the same glyph for the all-zero state, or document a distinct shape scale. The leading spaces in the frame must not look like missing data. Add a test with leading/trailing zeros.
5. **Cache precision**: Use the shared formatter: one decimal for cache hit (`96.1%`), not `96%` in live while daily shows `96.1%`. If compact width requires rounding, show `96%` only with a tooltip/legend? At minimum make the same data value render at the same precision across surfaces.
6. **Model-cost basis**: `todayModels` is an estimate from list prices; `todayCost`/`todayProjects` are recorded/blended. Label the top-model line `Est. model cost` or show model token share instead of a dollar total that can exceed Today's cost. Add `costKind`/`basis` to the model rollup JSON and text; never present the model estimate as a reconciled total.
7. **Layout preservation**: Keep the baseline/carried-in line, group rollup, `[s]` sort, `[Space]` pause, `[c]` reset, `[?]` help, and `[q]` quit working. Do not reintroduce seeded session values or change one-second bucket/flame semantics.
8. **Golden tests**: Add 80x24 and 120x40 frames with a hybrid token+step poll and a long model/project name; assert full units, no `step…`/`tok…`, `+60 step`, one sparkline glyph policy, and `Cache hit: 96.1%`. Add a JSON test for model-cost basis.
9. **Docs**: Add a six-part migration entry for unit/label changes if any JSON keys change; update README/gallery legend and project.jsonc.

## Done Condition
1. Live frame at 80/100/120 contains complete `tok/m`, `step/m`, `tok/s`, and `step/s` units; no unit is truncated with an ellipsis.
2. `Burned`/session steps use `step`; `s` is reserved for seconds/durations; the ticker and carried-in lines agree.
3. Sparkline zero cells and the all-zero state use one documented glyph policy; no leading-space ambiguity.
4. Cache hit uses the shared precision; model cost is labeled estimated/basis-specific and never appears as a reconciled total.
5. Golden 80/120 tests, full race suite, build, and simulated gate pass; docs/MIGRATION/README/project.jsonc updated; the attribution commit ends in `(goals/live-ticker-unit-and-rollup-contract/goal.md)`.

## Scope Boundary
Live text units/layout and model-cost labeling only. Do not change bucket math, flame ceiling, baseline accounting, ticker grouping/sort, or JSON envelope precision beyond the model-cost fields. The formatting/reconciliation goals own shared helpers/totals.
