# Facts: Stateless Zero-Database URL Sharing Engine for Streaks & Telemetry

## Architectural Invariants & Constraints
- Zero Database / Stateless Architecture: State is serialized entirely into the URL fragment. No database, server-side persistence, or central telemetry tracking.
- Sanitized Privacy Envelope: Payload schema must strictly exclude file paths, git repositories, user prompts, completions, and environment paths. Only aggregated telemetry facts and streak days are encoded.
- Integrity & Tamper Defense: Token carries a bounded SHA-256 checksum and bounds-checked decompression to reject corrupted or malformed payloads.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/share/share.go`
  - `internal/share/share_test.go`
