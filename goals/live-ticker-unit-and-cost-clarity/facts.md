# Facts: Separate token and step ticker columns and make live cost precision explicit

## Observed Evidence
- `internal/tui/live.go:889-911` (`formatTickerHeader`) selects `TOK/STEP` when both token and activity events exist. `internal/tui/live.go:913-951` (`formatEventRow`) prints `+X tok` or `+X step` in the same `tokW` cell depending on the event. There is no dedicated step column and no `—` for the non-applicable unit.
- `internal/tui/live.go:867-887` (`calcColWidths`) fixes six columns: time, tool, tokens, model, project, cost. It has no step column and no responsive drop order.
- `internal/tui/live.go:642-662` (`formatVolume`) renders hybrid volume as `X tok · Y step`; `internal/tui/live.go:699-739` (`formatTurnsAndHit`) renders `N turns (Z% hit)` from the same `TodayTurns` value. At 80 columns this produces the duplicated `12.0K step` / `11957 turns` pair.
- `internal/tui/live.go:686-697` (`formatCost`) returns `<$0.01` for every `0 < c < 0.01` and `$0.00` for `c <= 0`; `internal/tui/live.go:913-951` uses it for event rows without consulting `IsEstimated`. `$0.00` therefore means explicit zero, tiny-but-rounded, or unpriced depending on context.
- `internal/thermal/live.go:15-24` has only `IsEstimated` on `LiveEvent`; there is no `CostKind`, so the UI cannot distinguish recorded, estimated, and unpriced. `internal/thermal/live.go:248-252` computes cache hit; the abbreviation "hit" is purely presentational.
- `internal/thermal/mix.go:310-321` (`dayCost`) discards the pricer's `missing` list: `cost, _ := pricer.PriceDay(day)`. `internal/thermal/types.go:202-203` defines `PriceDay(day) (cost float64, missing []string)`, so the information needed to mark unpriced days is available but unused.
- `internal/render/leaderboard.go:210-222` already implements the desired convention: `—` when no cost, `$%.4f` below one cent, otherwise `$%.2f`, with `~` for estimates. The live view should reuse this convention rather than define a second one.
- Existing tests encode the old presentation: `internal/tui/live_test.go:484-541` asserts `TOK/STEP` and mixed `+... tok · +... step` strings; `internal/tui/live_test.go:590-695` asserts `hit` and the `6.6K (86% hit)` form.

## Interpretation
The ticker needs a schema change, not a string tweak: tokens and steps are different units and must be separate columns with explicit not-applicable cells. The duplicate count is caused by two helpers rendering the same `TodayTurns` field; remove one rendering path. Cost needs a status, not a magnitude heuristic: `CostKind` carries whether a value is recorded, estimated, or unpriced, and formatting can then use the leaderboard's established 4-decimal/`—` convention.

Outlier highlighting is a display concern: compute it from the visible event deltas in `internal/thermal` (or a pure helper there) so the TUI does not invent statistics, and exclude baseline rows. The default factor is 3x; keep it named for the later sort/rollup goal.

## Ownership and Constraints
- Event/snapshot semantics and median helper: `internal/thermal/live.go`, `internal/thermal/live_test.go`.
- Ticker/KPI rendering and width handling: `internal/tui/live.go`, `internal/tui/live_test.go`.
- Public JSON contract: `docs/MIGRATION.md`, `project.jsonc`.
- Preserve token-type disjointness, cache-hit math, baseline exclusion, and the 60-second bucket contract from the prerequisites.
- No new dependencies. Keep `--json` stdout clean and `--no-color` free of ANSI sequences.

## Verification Contract
```bash
go test -race ./internal/thermal ./internal/tui -run 'Live|Ticker|Cost|Outlier|Cache'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

Presentation cases:
- hybrid event set at 120x40 and 80x24: separate TOKENS/STEPS columns, `—` placeholders, no `TOK/STEP`.
- `formatLiveCost(0.0042, recorded) == "$0.0042"`, `formatLiveCost(0, unpriced) == "—"`, `formatLiveCost(0, recorded) == "$0.00"`.
- view text contains `Cache hit:` and does not contain the substring `hit` without `cache`.
- outlier set `[1, 1, 1, 4]` flags the 4 with `!`; `[1, 4]` does not flag (fewer than three rows).
