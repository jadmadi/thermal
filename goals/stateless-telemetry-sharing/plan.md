# Plan: Stateless Zero-Database URL Sharing Engine for Streaks & Telemetry

## Execution Steps
- [x] Phase 1 — Sharing Data Contract: Define TelemetryShareSnapshot schema with strict field allowlist in internal/share/share.go.
- [x] Phase 2 — Flate Compression & Base64URL Codec: Implement EncodeShareToken and DecodeShareToken with bounded checksum verification.
- [x] Phase 3 — CLI Integration & Tests: Add 'thermal share' command and table-driven unit tests verifying compression ratio, roundtrips, and tamper rejection.
