# Facts: Zero-Leak Markdown PR Contribution Receipts

## Architectural Invariants & Constraints
- Zero prompt leakage: output must never disclose prompt text, conversation turns, or repository secrets.
- Strict Markdown compatibility: valid GitHub Flavored Markdown tables and badges.
- --json output remains pristine and unaffected.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/render/receipt.go`
  - `internal/render/yield.go`
  - `internal/render/receipt_test.go`
  - `internal/render/yield_test.go`
  - `docs/MIGRATION.md`
