# Facts: Shared chart analytics and static chart verbs

## Architectural Invariants & Constraints
- Aggregations consume the existing types only: `DailyRow`, `ProjectDay`, `ToolDays`, `ModelRow`, `ProjectRow`. No loader changes in this goal.
- Data is day granularity. Switching is measured across consecutive active days, never inside a session.
- Estimated cost is never presented as billed cost. Charts and tables label estimates.
- Deterministic ordering everywhere, because golden tests assert on output.
- `--json` carries the raw numbers; charts and tables are presentation only.

## File & Interface Contracts
- New files: `internal/thermal/mix.go`, `internal/thermal/stats.go`, `internal/thermal/trend.go`, each with `_test.go` siblings.
- New reserved CLI words: `trend`, `mix`, `stats`. Flags: `--metric tokens|cost`, `--by tool|model`, `--last N`, `--grain day|week|month`, `--json`.
- Concentration is the Herfindahl index over series shares, 1 means one series, near 0 means an even spread.
- Outliers use median plus two MAD, not the mean, so a single huge day does not hide the rest.
