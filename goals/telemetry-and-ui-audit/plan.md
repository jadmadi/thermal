# Plan: Telemetry Accuracy, Pricing Estimates, and UI Polish

## Phase 1: Tool Name & `--help` (Issues 2 & 3)
- Update `internal/loaders/registry.go` to set `Name: "DeepSeek (DSH)"`.
- Update `cmd/thermal/main.go`:
  - Add `dsh DeepSeek (DSH)` to `printUsage()` supported tools table.
  - Add `dsh` to `flag.StringVar(&opts.Tool, "tool", ...)` description.
- Update tests expecting tool name.

## Phase 2: Pricing Estimates on Leaderboard and Dashboard (Issue 1)
- In `cmd/thermal/main.go` and `internal/render/leaderboard.go`:
  - Pass estimated cost per tool to `RenderLeaderboard` or calculate via `pricer` when `estimate` is true.
  - When recorded cost is 0, if estimated cost > 0, display `~$X.XX` with a `~` prefix.
  - Update leaderboard footer to reflect estimated costs.
- In `cmd/thermal/main.go` and `internal/render/dashboard.go`:
  - In `RenderDashboard`, calculate estimated cost when `summary.Cost == 0` from daily model counts.
  - If estimated cost > 0, display `~$X.XX spent (est)` in the extra summary line.

## Phase 3: Telemetry Invariant in Analytics (Issue 4)
- In `internal/thermal/mix.go`: in `AggregateToolMix`, when `a.metric == "tokens"`, check `!isActivityOnly(day)`.
- In `internal/thermal/stats.go`: in `AggregateStats`, when `metric == "tokens"`, check `!isActivityOnly(day)`.
- In `internal/thermal/trend.go`: in `AggregateTrend`, when `metric == "tokens"`, check `!isActivityOnly(day)`.
- Add unit tests verifying activity-only rows do not appear in token mixes, stats, or trends.

## Phase 4: UI Redesign of `thermal stats` (Issue 5)
- In `internal/render/analytics.go`:
  - Replace raw `#` ASCII histogram with `chartBar` or sleek block glyphs (`█` / `·`).
  - Align range labels cleanly (`From – To`).
  - Add mini-bars to `Weekday profile` showing relative volume across days.
  - Deduplicate / streamline `Top days` and `Outliers`.
  - Ensure `--no-color` strips ANSI codes cleanly.

## Phase 5: Verification & Documentation Parity
- Run full test suite: `go test -v -race ./...`.
- Run `golangci-lint run --disable-all -E gofmt,ineffassign,misspell`.
- Run smoke tests on real local data: `thermal`, `thermal dsh`, `thermal mix`, `thermal stats`, `thermal --help`.
- Update documentation and HTML pages to reflect `DeepSeek (DSH)` and updated output.
- Close Sila goal with commit and prompt user before remote push.
