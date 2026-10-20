# Facts: Fix live ticker units, truncation, sparkline glyphs, and model-cost semantics

## Observed Evidence
- `docs/OUTPUT-GALLERY.md:1325` `Rate: 3.4M/m · 23 step/m`; `:1326` `Speed: 56.7K · 0.383 step…`; `:1325` `Burned: +1.7M tok (+60 s)`.
- `docs/OUTPUT-GALLERY.md:1327` live `Cache hit: 96%` vs `:261` daily `Cache: 96.1% hit`.
- `docs/OUTPUT-GALLERY.md:1334` sparkline leading spaces; `formatSparkLine` uses `·` only when every bucket is zero.
- `docs/OUTPUT-GALLERY.md:1379` `todayCost: 1.9774259520000008` vs `:1426` `todayModels[0].cost: 3.9548519040000008`; current JSON shows the model cost is estimated and the project/today cost is recorded.
- `internal/tui/live.go` `formatRate`/`formatSpeed`/`formatSparkLine` and `padCell` are the formatting paths; `internal/thermal/live.go` owns `TodayProjects`/`TodayModels` and `costKind`.

## Interpretation
The live view uses abbreviations that are clear inside the code but not at a glance: `m` can mean minute or million, `s` can mean seconds or steps, and a truncated unit destroys the value. The model cost is not wrong as an estimate, but it is a different basis from Today's recorded cost and must be labeled.

## Ownership and Constraints
`internal/tui/live.go` owns text; `internal/thermal/live.go` owns rollups/costKind; tests own 80/120 golden frames. Preserve launch-relative session, carried-in, one-second buckets, flame, and rollup behavior. No new dependencies.

## Verification Contract
```bash
go test -race ./internal/tui ./internal/thermal -run 'Live|Ticker|Speed|Rate|Spark|CostBasis'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

## Open Questions
- At 80 columns, should the secondary step speed be dropped or abbreviated? The goal prefers dropping it before truncating a unit.
- Should Top model show estimated cost, token share, or both? The goal accepts either if the basis is explicit and it cannot be mistaken for Today's total.
