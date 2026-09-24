# Facts: Make verification independent of host data and color settings

## Observed Evidence
The 2026-09-24 race run failed TestEmitDeprecationWarning and TestRenderHelp_WithColors under inherited NO_COLOR; both passed with NO_COLOR unset. `internal/server/server_test.go` calls real collection without fixture HOME in route/concurrency tests. `scripts/simulated_user_gate.sh:33-91` uses host stores if present; run_check merges stderr/stdout and mostly validates exit/format/JSON parseability. Its 95/95 success did not detect the reproduced metric and evidence errors.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `scripts/simulated_user_gate.sh`
- `cmd/thermal/args_test.go`
- `cmd/thermal/simulated_user_test.go`
- `internal/render/help_test.go`
- `internal/server/server_test.go`

Do not hide regressions by globally forcing colors or skipping assertions. Repair the two color tests early if needed; final integration assertions depend on the listed remediation goals.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./...; ./scripts/simulated_user_gate.sh`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
