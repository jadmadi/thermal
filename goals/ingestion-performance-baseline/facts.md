# Facts: Measure cold and warm CLI performance on representative fixtures

## Observed Evidence
GOVERNANCE.md section 2 promises sub-10ms cached invocation, while repository benchmarks cover model-name normalization and leaderboard rendering only. `rg -n LoadOrScanWithCache internal` found no implementation of the mechanism named in AGENTS.md. Codex scans each eligible rollout on every invocation (codex.go:156-185); OpenCode/MiMoCode issue full aggregations. No representative end-to-end warm latency was measured in this evaluation, so the target is unverified rather than declared failed.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/loaders/benchmark_test.go`
- `scripts/benchmark.sh`
- `docs/PERFORMANCE.md`
- `AGENTS.md`

Measurement only. Do not optimize all loaders at once or make unmeasured speed claims. scripts/benchmark.sh is a proposed deliverable, not an existing verified command.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test ./internal/loaders -run '^$' -bench . -benchmem; ./scripts/benchmark.sh`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.

## Live CPU follow-up
The user reported high CPU at `thermal live --interval 100ms`. The evaluation now includes a narrow synthetic PTY/NDJSON reproduction (96.2%/130.5% of one core at 100ms). Extend the benchmark matrix with sustained 100ms/1s idle and active polling, but do not block the independent P1 `live-polling-cpu-budget` fix on this broader baseline. These CPU measurements do not verify the separate sub-10ms invocation claim.
