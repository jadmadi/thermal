# Facts: Preserve recorded usage and disclose receipt coverage

## Observed Evidence
`internal/thermal/receipt.go:770-780` multiplies Agy steps by 500; one activity-only fixture produced 500 tokens. Claude usage is re-summed without message-ID deduplication and priced as Unclassified at lines 564-565 and 622-629. `cmd/thermal/main.go:1630-1647` falls back only when the entire receipt list is empty, so any supported transcript suppresses aggregate-only sources. Fallback rows are labeled CLAIMED without evidence of a claim. The scanner recognizes raw names rather than ResolveTool aliases.

Evaluated working tree: HEAD `88c24d5` plus pre-existing uncommitted changes, 2026-09-24. Re-read these locations before implementing; line numbers describe this snapshot. Existing goal completion is commit attribution, not evidence that these edge cases pass.

## Ownership and Constraints
- `internal/thermal/receipt.go`
- `internal/thermal/receipt_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`

Do not implement every unsupported transcript schema in this goal. Disclose unsupported evidence and preserve available usage. Stream existing receipt readers with bounded memory and warnings if their parsing path is changed.

Preserve unrelated work. Source ingestion stays read-only; normalized token categories remain disjoint; no prompts in output/cache; diagnostics stay off JSON stdout. Keep aggregation in internal/thermal, bounded worker pools, and no background service. Load the repository Go/UI skills when the implementation scope needs them.

## Verification Contract
`go test -race ./internal/thermal ./cmd/thermal -run 'Receipt|Sim'`

Then `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh`. Use an isolated synthetic home, tool-root overrides and cached pricing. New benchmark paths are deliverables, not commands already present at evaluation time.
