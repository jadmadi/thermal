# Facts: Keep cached Devin totals isolated by source database

## Observed Evidence
`internal/loaders/cache.go:17-40` has one ~/.cache/thermal/devin.json with no source identity. `internal/loaders/sqlite.go:535-545` returns it when MAX(row_id) and visible session count match. Two fixture databases each had one session and row_id=1 but 100 versus 900 tokens; sequential loads returned 100 then 100. Existing-row edits and count-neutral hidden-session changes also are not represented by these keys (code inspection).

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/loaders/cache.go`
- `internal/loaders/sqlite.go`
- `internal/loaders/sqlite_test.go`

No general cache framework or changes to other tools. Add migration documentation only if user-visible flags/schema/aggregation semantics change beyond eliminating stale cache reuse.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/loaders -run 'Devin|Cache'`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
