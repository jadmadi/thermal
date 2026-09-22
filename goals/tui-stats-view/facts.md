# Facts: TUI Stats view

## Architectural Invariants & Constraints
- Every number shown must match `thermal stats --json` for the same window and metric. Parity is tested.
- Outliers use median plus two MAD so a single outlier day does not distort the threshold.
- The projection is a range, drawn as a band from the 14-day trend, never a single confident point.
- Histograms switch to log scale when the maximum exceeds ten times the median, with the scale labeled.

## File & Interface Contracts
- New view file `internal/tui/views/stats.go`, reusing the stats and trend aggregations from the charting-analytics goal.
- Widgets: histogram with bin count, weekday bars, top-days list, outlier list, projection line.
- Ranges follow the shared range selector (30d, 90d, 1y, all), metric toggle switches tokens and cost.
