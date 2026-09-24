# Facts: Avoid rescanning unchanged Codex rollout transcripts

## Observed Evidence
`internal/loaders/codex.go:156-185` invokes readLastTokenBreakdown for every nonempty rollout on every load. That function reads the file from its beginning at line 349 onward. Bounded concurrency exists, but persistent per-file reuse does not. The cache path must also preserve diff-line telemetry, warnings, model identity and authoritative state DB reconciliation.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/loaders/codex.go`
- `internal/loaders/codex_test.go`
- `internal/loaders/codexcache.go`
- `internal/loaders/benchmark_test.go`
- `docs/PERFORMANCE.md`

Do not change token attribution policy or cache other tools here. Any observable aggregation/schema change requires the six-part migration entry. This is a candidate targeted optimization; baseline measurement must confirm its benefit before implementation.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/loaders -run 'Codex|Cache'; ./scripts/benchmark.sh`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
