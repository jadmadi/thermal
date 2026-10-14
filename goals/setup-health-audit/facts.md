# Facts: Read-Only Local Setup & Context Health Audit ('thermal audit')

## Architectural Invariants & Constraints
- Strict Read-Only Non-Destructive Invariant: 'thermal audit' is strictly diagnostic. It reads configuration files and session logs but NEVER writes, edits, or deletes user configs (no destructive --apply mutation).
- Zero Third-Party Code: Written entirely from scratch in Go. Unencumbered dual-licensing under AGPLv3 and Commercial Enterprise.
- Bounded Scan Overhead: Config file parsing is bounded to known local agent configuration directories, with file size caps (max 1MB per config file) to prevent memory spikes.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/audit/audit.go`
  - `internal/audit/audit_test.go`
  - `internal/render/audit.go`

## Current evidence, 2026-10-14

The earlier observations above are historical. Current source revision is
`302501b17f96a26a13b7b5cbaf4e7654199f7d5c`. See the
[evaluation](../../docs/evaluations/2026-10-14-promise-vs-delivery.md) and
[current contract](remediation-2026-10-14.md). Reopening preserves prior commits and evidence;
it does not imply that the earlier implementation never existed.
