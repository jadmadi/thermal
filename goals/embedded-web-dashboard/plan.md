# Plan: Embedded Local Web Dashboard for Agent Telemetry & Streak Visualization

## Execution Steps
- [x] Phase 1 — HTTP Server Core & Embed Pipeline: Implement localhost HTTP listener with go:embed in internal/server/server.go.
- [x] Phase 2 — Read-Only Telemetry APIs: Expose JSON endpoints querying Thermal's local SQLite/analytics cache.
- [x] Phase 3 — Web UI & Verification: Build responsive single-page dashboard and verify security headers, concurrent requests, and localhost binding.
