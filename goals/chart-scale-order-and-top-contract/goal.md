# Goal: Fix chart scaling, ordering, row counts, and no-op --chart flags

## Goal Description
Priority: P1 (chart correctness). The projects chart uses the first table row's token count as its maximum. When the table is sorted by cost, the first row (tree.waqf.app, 3.9B) is not the largest token row (mahak-bench, 7.1B); both bars fill 48 cells and the caption falsely says “largest row, 3.9B.” The models chart repeats the same first-row scaling pattern. Charts also show 12 rows while tables honor `--top 10`, and `--chart` is accepted on mix/stats/trend but produces no chart.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-formatting-primitives
- **Sequence**: 205
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/render/chart.go`
- `internal/render/chart_test.go`
- `internal/render/chart_width_test.go`
- `internal/render/projects.go`
- `internal/render/models.go`
- `internal/render/analytics.go`
- `internal/render/help.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Scale from charted rows**: `RenderProjectChart` and `RenderModelChart` must compute `maxVal` as the maximum token value across the rows they actually render, not `rep.Rows[0].Tokens`. The caption must use that same max. Bars larger than the first row must not clamp to full unless they are the max.
2. **Metric/order contract**: Either sort chart rows independently by the bar metric (tokens) or keep table order and label it explicitly (`rows in table order; bars = tokens`). The current cost-sorted table plus token-scaled bars must not imply that bar length is cost. Add a legend/caption test for both projects and models.
3. **Row-count contract**: Align the chart row cap with `--top` or explicitly state the chart's own 12-row cap. The table and chart must not report contradictory remainders (`and 106 more` vs `and 104 more`). Define one `chartRows` policy and test it with `--top 5`, 10, 12, and 0.
4. **`--chart` applicability**: Decide per command. For projects/models/daily/weekly/monthly, preserve the working chart. For mix/stats/trend, either implement a concrete chart (e.g., mix shares over time, stats histogram is already present, trend sparkline is already present) or reject `--chart` with a clear validation error. Do not silently accept a no-op. Update `validateReportFlags` and help to match.
5. **Caption/legend**: Every chart must name the scale (`bar = share of largest token row, 7.1B`) and the metric (`tokens` vs `cost`), and state the remainder. Charts must not rely on color to convey meaning.
6. **Width safety**: Preserve 80/120/160-column behavior; truncate labels before bars; no line wraps. Add tests for narrow widths with long project/model names.
7. **No data changes**: Chart fixes are view-only; do not change report sorting/aggregation or project/model totals. The reconciliation goal owns totals.
8. **Docs/migration**: Add a six-part migration entry if `--chart` becomes invalid on analytics commands; update help/README/project.jsonc.

## Done Condition
1. With a cost-sorted fixture where the first row is not the max-token row, the project chart's largest bar belongs to the max-token row and the caption names that value; models chart uses the same row-set max.
2. Chart rows and table rows agree on `--top`/remainder, or the chart explicitly labels its own cap.
3. `--chart` is either visibly implemented or explicitly rejected on mix/stats/trend; help and validation agree.
4. Charts remain readable at 80/120/160 columns, are monochrome-safe, and do not wrap.
5. Full race suite, build, and simulated gate pass; docs/MIGRATION/README/project.jsonc updated; the attribution commit ends in `(goals/chart-scale-order-and-top-contract/goal.md)`.

## Scope Boundary
Chart rendering/scaling/row counts/flag applicability only. Do not change report totals, table sorting, heatmap labels, live ticker, or JSON. Formatting helpers come from the prerequisite.
