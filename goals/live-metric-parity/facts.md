# Facts: Make live burn metrics reflect recorded tokens and turns

## Observed Evidence
`internal/thermal/live.go:146-165` sums all daily Tokens and uses combined cache read+write as the hit-rate numerator. The probe turned seven activity steps into todayTokens=7. Lines 175 and 240-252 derive completed turns from Summary.Sessions, so a new turn inside an existing session cannot increment the turn delta. Today's cost only adds recorded costs despite lifetime estimated-cost deltas.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`

No animation or TUI layout work. Dependency serializes shared CLI wiring and establishes the common metric contract.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/thermal ./cmd/thermal -run 'Live|Sim'`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
