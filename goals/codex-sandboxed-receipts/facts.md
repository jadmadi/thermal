# Facts: Codex Sandboxed Execution Verification Receipts

## Architectural Invariants & Constraints
- Zero prompt or secret leakage: receipts store only command signatures, verification tiers, and exit status.
- All rollout scans must honor JSONL 32MB line ceiling and bounded worker pools.
- Strict read-only ingestion without modifying user session transcripts.

## File & Interface Contracts
- Relevant files:
  - `internal/loaders/codex.go`
  - `internal/loaders/codex_test.go`
  - `internal/thermal/receipt.go`
  - `internal/thermal/receipt_test.go`
  - `internal/render/receipt.go`
