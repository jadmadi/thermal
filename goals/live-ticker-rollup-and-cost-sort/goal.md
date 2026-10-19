# Goal: Roll up live ticker bursts and add cost-first decision sorting

## Goal Description
The live ticker renders one row per poll delta, so a project/model that emits several near-identical bursts produces a stack of rows like `thermal-streak ... +567.9K tok` and `thermal-streak ... +10.5K tok` instead of a single decision summary. There is also no way to reorder the ticker by spend; the only order is newest-first (`internal/tui/live.go:424-427`), and the key handler owns only pause/reset/help/quit (`:176-201`).

Roll up visible bursts by `(tool, model, project)` within a 60-second window into one row with a burst count and summed values, and add an `[s]` toggle between time and cost ordering. In cost mode, show a `Top spenders (60s)` line by project. This is the view the feedback calls "the decision view"; it must preserve the normalized token/step columns and outlier flags.

Priority: **P1**. Derived from the 2026-10-19 live-monitor feedback (improvements "Project/model rollup" and "Cost-first sorting option").

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: live-ticker-unit-and-cost-clarity
- **Sequence**: 104
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)
- Prerequisite: live-ticker-unit-and-cost-clarity

## Requirements
1. **Rollup contract**: Add a pure grouping helper in `internal/thermal` that takes visible `LiveEvent`s and a 60-second window and returns groups keyed by `(Tool, Model, Project)`. Only events within 60 seconds of the newest visible event are grouped; a single event renders as a normal row. Baseline events are excluded. Each group carries: event count, latest timestamp, summed tokens, summed turns/steps, summed recorded/estimated cost, whether any member is unpriced, and whether any member is an outlier.
2. **Rollup row**: Render a group with `N bursts` (or `Nx` at narrow widths) and the summed token/step/cost values using the dedicated columns from the prerequisite. Example: `thermal-streak 3 bursts 1.7M 12 ~$0.01`. Do not collapse different models or projects into one row. If any member is unpriced, show `≥$X` for the known portion or `—` when all are unpriced; never present a partial sum as complete.
3. **Sort toggle**: Add `[s]` to the live key handler, toggling between `time` (default, latest group first) and `cost` (highest known group cost first). Cost sort places priced groups before unpriced/`—` groups and breaks ties by total tokens then latest timestamp. Show the active mode in the footer (`[s] Sort: time|cost`) and in the help screen. The key must not conflict with pause/reset/help/quit.
4. **Top spenders**: In cost mode, render a `Top spenders (60s): project ~$0.01 · project2 ~$0.00` line above the ticker, aggregating group costs by project and showing at most three projects. If no group has a known cost, render `Top spenders (60s): —`. In time mode, show the current time-ordered ticker without this line.
5. **Outlier interaction**: Recompute the median outlier flag over the displayed group totals (tokens and steps separately) using the prerequisite helper, with baseline excluded. A group is flagged if its total exceeds 3x the median group total; individual flags are preserved in the group metadata if needed. The `! >3x median` legend remains.
6. **Responsive behavior**: At 80 columns, the rollup count and both numeric columns must remain visible; use `Nx`/`3b` and drop PROJECT/MODEL per the prerequisite drop order. At 120 columns, show the full `N bursts` label. Never wrap the ticker; truncate model/project before numeric data.
7. **No data changes**: Grouping/sorting is view-level only. The tracker's event stream, JSON snapshot, token/step sums, and cost statuses remain unchanged. `--json` and `--json --stream` are unaffected by `[s]`.
8. **Safety**: No new dependencies; preserve ANSI width padding, `--no-color` plain output, and the 80x24/120x40 test requirement.

## Done Condition
1. Three events for the same tool/model/project within 60 seconds render as one row with `3 bursts`, summed tokens/steps, and summed known cost; events outside the window remain separate.
2. Different models or projects never merge; a single event renders without a burst label.
3. Pressing `[s]` switches between time and cost order; cost order is descending by known cost with unpriced groups last; the footer/help reflect the mode.
4. Cost mode shows `Top spenders (60s)` by project; time mode does not; an all-unpriced set shows `—`.
5. The prerequisite normalized columns, `—` placeholders, cache-hit wording, and outlier flags remain intact at 80x24 and 120x40.
6. `thermal live --json` output is byte-for-byte unaffected by the view toggle/grouping (apart from timestamps), and no new CLI flag is required.
7. Focused tests, full race suite, CLI build, and simulated user gate pass; `project.jsonc` is synchronized; the attribution commit ends in `(goals/live-ticker-rollup-and-cost-sort/goal.md)`.

## Scope Boundary
No tracker metric, baseline, bucket, flame, cost-status, or ticker-column changes; those are owned by the prerequisites. This goal adds grouping and ordering only, plus the `[s]` key and top-spender summary.
