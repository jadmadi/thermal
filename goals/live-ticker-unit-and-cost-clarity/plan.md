# Plan: Separate token and step ticker columns and make live cost precision explicit

## Execution Steps
- [x] **Task 1 — Add cost status and unpriced accounting**: Extend `LiveEvent` with `CostKind` (`recorded`/`estimated`/`unpriced`) and `LiveSnapshot` with unpriced token/event counts. Preserve recorded costs, estimates under the existing no-estimate rules, and the pricer's `missing` signal from `PriceDay`.
  Acceptance and verification: `go test -race -run 'TestLiveTracker_CostKind|TestLiveTracker_UnpricedTokens' ./internal/thermal`; recorded zero, estimated 0.0042, and missing-pricing cases produce distinct kinds without changing `TodayCost`.

- [x] **Task 2 — Replace TOK/STEP with two columns**: Rework `calcColWidths`, `formatTickerHeader`, and `formatEventRow` to render separate `TOKENS` and `STEPS` columns with right-aligned deltas and `—` placeholders. Implement the responsive drop order (PROJECT, then MODEL) while preserving both numeric columns.
  Acceptance and verification: `go test -race -run 'TestLiveTUI_TickerColumns|TestLiveTUI_TickerResponsive' ./internal/tui` at 80, 100, and 120 cells; no output contains `TOK/STEP`, and each row's visual width equals the inner width.

- [x] **Task 3 — Remove the duplicated aggregate**: Replace the `Turns: N turns (Z% hit)` line with `Cache hit: Z%` and ensure the step/turn count is rendered once (Today volume line or activity line). Use a single canonical label per mode and update the help/legend.
  Acceptance and verification: `go test -race -run 'TestLiveTUI_NoDuplicateTurns|TestLiveTUI_CacheHitLabel' ./internal/tui`; the view contains one occurrence of the step count and no bare `hit`.

- [x] **Task 4 — Adopt the cost convention**: Add a live cost formatter that mirrors `internal/render/leaderboard.go:210-222`: `—` for unpriced, `$0.00` only for recorded zero, four decimals below one cent, two decimals above, `~` for estimates, and `<$0.0001` for positive values below the four-decimal floor. Use it in the ticker and KPI grid.
  Acceptance and verification: `go test -race -run TestLiveTUI_CostPrecision ./internal/tui`; table cases include 0, 0.00001, 0.0042, 0.01, 1.25, estimated, and unpriced; `<$0.01` is absent from the rendered view.

- [x] **Task 5 — Add median outlier flags**: Add a pure helper in `internal/thermal` that returns median and outlier flags for a slice of positive deltas. In the ticker, compute token and step medians separately over visible non-baseline rows, flag `value > 3*median` when at least three positive rows exist, and add the `! >3x median` legend.
  Acceptance and verification: `go test -race -run 'TestLiveOutlierFlags|TestLiveTUI_OutlierMarker' ./internal/thermal ./internal/tui`; baseline rows are excluded; median zero and fewer-than-three cases do not flag.

- [x] **Task 6 — Width and no-color verification**: Add 80x24 and 120x40 view checks with color forced off, including hybrid and activity-only data. Verify borders, column alignment, and that no ANSI escape sequences appear under `--no-color`.
  Acceptance and verification: `go test -race -run TestLiveTUI_Golden ./internal/tui`; inspect the rendered frames for aligned headers and `—` cells.

- [x] **Task 7 — Migration, metadata, and attribution commit**: Document the new `costKind`/unpriced JSON fields and the removed `TOK/STEP`/`hit` presentation in `docs/MIGRATION.md`; update `project.jsonc`. Run the full gate and create an impact-first commit ending in `(goals/live-ticker-unit-and-cost-clarity/goal.md)`. Run `sila goals` and do not push.

## Dependencies and Handoff
Start after `live-baseline-separation` closes so `IsBaseline` and carried-in rows exist. Do not change token/bucket math; consume the prerequisite's event fields. The later rollup/sort goal will reuse the new column schema and outlier helper.

## Verification
```bash
go test -race ./internal/thermal ./internal/tui -run 'Live|Ticker|Cost|Outlier|Cache'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
