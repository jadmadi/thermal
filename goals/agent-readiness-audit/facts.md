# Facts: Repository Agent Readiness and AGENTS.md Audit

## Architectural Invariants & Constraints
- Strictly non-destructive and read-only; never create or alter repository files.
- Maintain sub-10ms invocation target for audit scans.
- Zero cloud telemetry or external network calls.

## File & Interface Contracts
- Relevant files:
  - `internal/audit/audit.go`
  - `internal/audit/audit_test.go`
  - `internal/render/audit.go`
