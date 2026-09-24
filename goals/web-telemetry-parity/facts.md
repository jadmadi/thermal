# Facts: Make web totals match canonical reports

## Observed Evidence
`internal/server/server.go:272-421` separately sums summaries and daily rows rather than using the activity-only exclusion in `internal/thermal/stats.go:104`. The probe returned web total 151 versus stats total 150 for the same Claude+one-step Agy fixture. Cost estimation is conditional on the entire tool summary having zero cost, unlike per-day pricing. `runServe` passes only Offline; accepted tool/window/no-estimate options cannot affect collection.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/server/server.go`
- `internal/server/server_test.go`
- `internal/thermal/telemetry.go`
- `cmd/thermal/main.go`
- `docs/MIGRATION.md`

No visual redesign, new telemetry store or cloud endpoint. Preserve established JSON keys unless a documented migration is necessary. Follow ui-craft if frontend rendering changes become necessary.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/server ./internal/thermal`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
