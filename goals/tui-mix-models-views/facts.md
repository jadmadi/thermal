# Facts: TUI Mix and Models views

## Architectural Invariants & Constraints
- Shares are computed over the tokens or cost inside the selected range, and must sum to 100 percent including the other bucket.
- Concentration and switch counts must match the static `mix` output. Same aggregation code, no duplicates in the view layer.
- Model names are canonical lowercase keys. `GLM-5.3-Flash` and `glm-5.3-flash` are one series.
- No color-only encoding: each series also gets a legend row and, in no-color mode, a distinct glyph pattern.

## File & Interface Contracts
- New view files under `internal/tui/views/`: `mix.go` (stacked share, switching panel), `models.go` (ranked bars, share strip).
- Reuse `thermal.AggregateModels` and the new mix aggregation from the charting-analytics goal.
- Series cap: top four plus an other bucket, so narrow terminals stay readable.
