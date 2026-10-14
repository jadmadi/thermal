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

## Current evidence, 2026-10-14

The earlier observations above are historical. Current source revision is
`302501b17f96a26a13b7b5cbaf4e7654199f7d5c`. See the
[evaluation](../../docs/evaluations/2026-10-14-promise-vs-delivery.md) and
[current contract](remediation-2026-10-14.md). Reopening preserves prior commits and evidence;
it does not imply that the earlier implementation never existed.
