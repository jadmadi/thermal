# Facts: Require observed command results for verified receipts

## Observed Evidence
`internal/thermal/receipt.go:123-145` defaults an unknown exit to zero and checks pass text before failure text. `parseClaudeReceipt` and `parseAgyReceipt` turn trailing pending commands into exit-zero evidence (lines 607-610 and 753-756). The probe returned VERIFIED for a Claude Bash call with no result; both empty output and `1 passed, 2 failed` became `go test (exit 0)`.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/thermal/receipt.go`
- `internal/thermal/receipt_test.go`
- `docs/MIGRATION.md`

Do not redesign spend aggregation or add additional harness adapters here. Those changes belong to receipt-accounting-coverage.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/thermal -run 'Receipt|Evidence|Outcome'`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
