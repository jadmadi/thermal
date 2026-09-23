# Facts: Verifiable Work Receipts & Session Outcome Correlation ('thermal receipt')

## Architectural Invariants & Constraints
- Clean-Room Implementation: 100% original Go code. 100% sole copyright by Jad Madi. Fully dual-licensable.
- Factual Evidence Hierarchy: Verifications are classified strictly by observable fact: Tier 1 = Observed zero-exit test/linter runs or git commits; Tier 2 = Agent-claimed completion; Tier 3 = Unverified/interrupted sessions.
- Privacy Preservation: Output contains only task outcome summaries, exit codes, and token/cost tallies; no sensitive prompt or error trace details leave the local machine.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/thermal/receipt.go`
  - `internal/thermal/receipt_test.go`
  - `internal/render/receipt.go`
