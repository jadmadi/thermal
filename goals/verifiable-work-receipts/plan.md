# Plan: Verifiable Work Receipts & Session Outcome Correlation ('thermal receipt')

## Execution Steps
- [x] Phase 1 — Outcome Parser: Parse session tool calls for command exit codes, test runners (go test, pytest, npm test, cargo test), and git commits in internal/thermal/receipt.go.
- [x] Phase 2 — Receipt Data Model: Define Work Receipt schema with evidence tiers, cost attribution, and verification state.
- [x] Phase 3 — CLI & Presentation Table: Implement 'thermal receipt' command with formatted terminal summary cards and JSON export.
