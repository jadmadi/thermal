# Facts: ChatGPT Rolling 5-Hour Burst Capacity Replay

## Architectural Invariants & Constraints
- No cloud telemetry or external network calls; all burst calculations use local timestamps and models.dev catalog pricing.
- Parity across CLI table, TUI, and --json representations.
- Sub-10ms invocation maintained via pure in-memory rolling window evaluation.

## File & Interface Contracts
- Relevant files:
  - `internal/thermal/replay.go`
  - `internal/thermal/replay_test.go`
  - `internal/render/replay.go`
  - `internal/render/replay_test.go`
  - `docs/MIGRATION.md`
