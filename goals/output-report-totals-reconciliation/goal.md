# Goal: Reconcile report totals, unclassified tokens, cost bases, windows, and share scope

## Goal Description
Priority: P0 (numeric trust). The gallery makes a reader ask why the same data yields different totals: daily says 38.0B tokens / $20596.53, models says 36.7B / $20606.08, projects says 38.0B / $20599.42, mix says $20606.08, stats says $20596.53, and share says 38.0B / $20596.53 with 145 active days while daily says 96. The 1.3B token gap is unclassified volume; the cost differences come from different pricing/attribution paths; stats cost reports 80 active days because zero/unpriced cost days are filtered; share mixes scopes. This goal makes every aggregate either equal or explicitly declare its basis.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-formatting-primitives
- **Sequence**: 201
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/stats.go`
- `internal/thermal/stats_test.go`
- `internal/thermal/models.go`
- `internal/thermal/models_test.go`
- `internal/thermal/projects.go`
- `internal/thermal/projects_test.go`
- `internal/thermal/report.go`
- `internal/thermal/report_test.go`
- `internal/thermal/types.go`
- `internal/share/share.go`
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `internal/render/analytics.go`
- `internal/render/report.go`
- `internal/render/projects.go`
- `internal/render/models.go`
- `internal/render/leaderboard.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`
- `README.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Unclassified-token contract**: Every token total must expose classified + unclassified = total. Add `UnattributedTokens`/`UnclassifiedTokens` to the models report totals and stats composition, and ensure daily/weekly/monthly/projects already expose it. Models text must show an explicit `Unclassified: X tok` line (or a row) and its `--json` total must include `unattributedTokens`; stats text must add an `Unclassified` composition row and label composition percentages as `of classified tokens`. The default fixture's 1.3B gap must be visible on every token surface.
2. **Stats cost day semantics**: Split `ActiveDays` (days with token/activity volume) from `CostedDays` (days with a non-zero cost) in `StatsReport`. `thermal stats --metric cost` must print `Costed days` and disclose the excluded zero/unpriced days, or include a zero-cost histogram bin with `Active days` remaining 96. The mean/median/percentile denominators must be documented in the footer. Tests must cover a zero-cost day and an unpriced-model day.
3. **Canonical cost basis**: Add a documented cost basis enum/fields (`recorded`, `estimated`, `total`) and make leaderboard/daily/projects/models/mix/stats/share consume the same estimator or declare a different basis. The fixture must satisfy: leaderboard estimated sum == daily estimated cost; projects total cost == daily total cost; mix total == models estimate; models is explicitly `Est.` and never compared as recorded+estimated. If exact equality is impossible, print a reconciliation line that names the basis and the excluded/attributed portion.
4. **Models total**: The models table total must not silently become 36.7B when daily says 38.0B. Show `Classified 36.7B + Unclassified 1.3B = 38.0B` (or a dedicated row). The models cost total must be labeled `Est. from list prices` and must not imply it equals daily recorded+estimated cost.
5. **Project cost attribution**: Reconcile `ProjectRow.Cost` with `DailyRow.Cost` for the same window; if project model maps cannot attribute a token/model, put it in a project `unattributed` bucket and disclose it. The projects footer must reconcile to daily's recorded+estimated total or state why not.
6. **Stats composition**: Add a `ClassifiedTotal` and `Unclassified` field to `TokenComposition`; percentages must use the correct denominator and the footer must say whether they are of classified or all tokens. Cache hit rate remains `cacheRead/(uncachedInput+cacheRead+cacheWrite)` and must be labeled with that formula or a plain-language equivalent.
7. **Window/scope labels**: Use the formatting goal's window helper. Daily/weekly/monthly default labels are active periods (`all time (96 active days)`), not `Window: 96 days`; explicit `--last` stays `last 7 days`; trend stays calendar days; projects/models use project-days/model-days. Share must expose `scope`/window and one consistent set of totals.
8. **Share snapshot contract**: Add a canonical `ShareSnapshot` with `scope`, `window`, `activeDays`, `totalTokens`, `recordedCost`, `estimatedCost`, `totalCost`, and `coverage`. Either all fields are all-time or all are the same window; mixed scopes are forbidden. `share` text/JSON must derive from this snapshot, not from a mix of daily and streak fields.
9. **Cross-surface test**: Add one fixture-driven test that decodes daily, weekly, monthly, projects, models, stats, mix, leaderboard, and share reports and asserts the permitted equalities/inequalities. The test must fail if a new surface reintroduces a silent total without a basis.
10. **Docs and migration**: Document the cost-basis and unclassified-token rules in `docs/MIGRATION.md`, README, and `project.jsonc`; add a concise legend for `Classified`, `Unclassified`, `Recorded`, `Estimated`, and `Total`.

## Done Condition
1. On the fixture, daily/weekly/monthly tokens equal 38.0B; models shows classified 36.7B + unclassified 1.3B = 38.0B; stats composition includes 1.3B unclassified; no surface hides the gap.
2. `thermal stats --metric cost` reports 96 active days with a zero-cost bin or explicitly reports `Costed days 80` and the 16 excluded days; mean/median denominators are documented.
3. Leaderboard estimated sum equals daily estimated cost (within a documented rounding epsilon); projects total equals daily recorded+estimated; models/mix are labeled estimate-only and reconcile to their own basis.
4. `thermal share` uses one scope for active days, volume, and spend; daily/share active-day integers agree or the difference is explained.
5. Projects/models charts/rows still render, JSON remains valid, and the full race suite, build, and simulated gate pass; docs/MIGRATION and project.jsonc are updated; the attribution commit ends in `(goals/output-report-totals-reconciliation/goal.md)`.

## Scope Boundary
Numerical reconciliation and disclosure only. No chart layout, heatmap labels, help text, JSON float formatting, live ticker units, audit paths, or share URL layout; those are owned by sibling goals. Formatting primitives come from the prerequisite.
