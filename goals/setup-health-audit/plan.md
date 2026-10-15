# Plan: Read-Only Local Setup & Context Health Audit ('thermal audit')

## Execution Steps
- [x] Phase 1 — Config & MCP Audit Heuristics: Implement scanners for MCP configurations, rule file token bloat, and session tool utilization in internal/audit/audit.go.
- [x] Phase 2 — Diagnostic Reporter: Build diagnostic grading engine that correlates installed tools against actual 30-day session calls.
- [x] Phase 3 — Terminal Presentation & Tests: Implement golden terminal output and JSON reporting for 'thermal audit', with unit tests verifying heuristic thresholds.

## Remediation, 2026-10-14

- [x] Implement the cards and stable acceptance IDs in [the remediation contract](remediation-2026-10-14.md).
- [x] Attach candidate-bound results and receive independent review under `tasks/plan.md`.

Earlier checked steps describe historical work and do not certify these reopened cases.
