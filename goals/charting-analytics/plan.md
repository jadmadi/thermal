# Plan: Shared chart analytics and static chart verbs

## Execution Steps
- [x] **Task 1 — Mix aggregation** (`internal/thermal/mix.go`): Period by series share for tool and model, switches per month (dominant tool changes between consecutive active days), Herfindahl concentration, and active tool count. Unit tests for flat, single-series, and empty windows.
- [x] **Task 2 — Stats aggregation** (`internal/thermal/stats.go`): Median, p90, max, weekday profile, robust outliers (median plus two MAD), and top days. Unit tests including an all-zero window.
- [x] **Task 3 — Trend aggregation** (`internal/thermal/trend.go`): Least-squares slope over the window and a 14-day month-end projection with a band. Unit tests with a known line and a flat series.
- [x] **Task 4 — Static verbs** (`cmd/thermal`): `thermal trend`, `thermal mix`, `thermal stats` with `--metric tokens|cost`, `--by tool|model`, `--last`, `--grain day|week|month` and `--json`. Golden tests, plus verification against hand calculations on local data.
- [x] **Task 5 — Attribution commit**: Commit with `Goal-Ref: charting-analytics` using a `feat:` subject, then run `sila goals`.
