# Plan: Token Yield & Code Output Delta Telemetry ('thermal yield')

## Execution Steps
- [x] Phase 1 — Session Diff Extraction: Extract patch and line change statistics from agent loader transcripts in internal/loaders/.
- [x] Phase 2 — Yield Computation Engine: Implement Token Yield ratios, efficiency metrics, and model comparisons in internal/thermal/yield.go.
- [x] Phase 3 — CLI Integration & Render Table: Expose 'thermal yield' CLI command with tabular and JSON outputs, guarded by automated tests.
