# Goal: Eliminate overlapping heatmap month labels

## Goal Description
Priority: P1 (visible rendering bug). Single-tool heatmaps print `ONov` because `RenderHeatmap` writes adjacent month abbreviations into a shared character slice without clearing or collision checks. When the first Sunday falls late in October, `Oct` and `Nov` overlap and leave only `O` before `Nov`. The same class of bug can affect any adjacent month pair and any 52-week window.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 204
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/render/heatmap.go`
- `internal/render/heatmap_test.go`
- `internal/render/dashboard.go`
- `internal/render/dashboard_test.go`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Collision-free placement**: Track occupied columns for month labels. A new label may overwrite only its own three columns; it must clear stale characters from an overlapping previous label or skip the new label if it cannot fit with at least one space. Never leave a partial old label such as `O` before `Nov`.
2. **Year boundaries**: Handle December/January and multi-year 52-week windows without merging labels or printing a stray year character. If a year marker is desired, print it as a separate, deliberate token; do not let it collide with a month abbreviation.
3. **Widths**: Preserve the existing six-space heatmap header indent and cell alignment. Test widths 80, 100, 120, 160 and weeks 4, 26, 52, 104. Labels must not exceed the heatmap width or push the right border.
4. **No data changes**: Do not change activity thresholds, heat cells, `Less □░▒▓█ More`, streak numbers, or session counts. This is a label-placement fix only.
5. **Regression test**: Add a fixed-date test for a first Sunday in late October (the `ONov` case) asserting the line contains `Nov` and does not contain `ONov`, plus a table of expected month sequences for year boundaries and leap years. Add one golden single-tool output snapshot with no color.
6. **Docs**: If the header policy changes (e.g., year marker or skipped month), add a short migration/docs note; otherwise no user-facing schema change is expected.

## Done Condition
1. `thermal opencode --no-color` and `thermal agy --no-color` no longer contain `ONov`; month labels have at least one space between adjacent months.
2. Fixed-date tests cover late-October first Sundays, December/January, leap years, and 52/104-week widths; no label clips the border.
3. Heatmap cells, activity totals, and single-tool summary values are unchanged against the pre-fix fixture.
4. Full race suite, build, and simulated gate pass; the attribution commit ends in `(goals/heatmap-month-label-overlap/goal.md)`.

## Scope Boundary
Month-label placement only. Do not touch heatmap thresholds, activity data, chart/table formatting, help, or cost logic.
