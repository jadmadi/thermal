# Facts: Embedded Local Web Dashboard for Agent Telemetry & Streak Visualization

## Architectural Invariants & Constraints
- Single Binary via go:embed: Web assets must be compiled and embedded directly into the Go binary. No Node.js runtime or external daemon required for end users.
- Localhost Security Boundary: The server binds to 127.0.0.1 by default with strict Content-Security-Policy headers and read-only local telemetry APIs.
- Clean-Room UX Design: Interface inspired by modern telemetry dashboards, displaying streak flame counters, multi-tool usage matrix, and activity timelines without replicating proprietary code or styles.

## File & Interface Contracts
- Relevant files:
  - `cmd/thermal/main.go`
  - `internal/server/server.go`
  - `internal/server/server_test.go`
