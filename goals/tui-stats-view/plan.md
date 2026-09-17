# Plan: TUI Stats view

## Execution Steps
- [ ] **Task 1 — Histogram**: Distribution of daily tokens or cost with a bin count and a log-scale toggle for long tails.
- [ ] **Task 2 — Rhythm**: Weekday profile bars and a top-days list with dates.
- [ ] **Task 3 — Outliers**: Robust outlier list with scores, matching the static stats output.
- [ ] **Task 4 — Projection**: Month-end projection line with a band drawn from the trend aggregation.
- [ ] **Task 5 — Parity and tests**: Values match `thermal stats --json`, golden view tests at two sizes with color off.
- [ ] **Task 6 — Attribution commit**: Commit with `Goal-Ref: tui-stats-view` using a `feat:` subject, then run `sila goals`.
