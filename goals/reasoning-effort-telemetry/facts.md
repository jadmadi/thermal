# Facts: Reasoning Effort Breakdown & Cognitive Intensity Telemetry

## Architectural Invariants & Constraints
- Reasoning tokens must remain strictly disjoint from Input, Output, CacheRead, and CacheWrite.
- Source SQLite connections remain read-only (?mode=ro) with PRAGMA mmap_size.
- Diagnostics ride Summary.Warnings to stderr only; stdout and --json remain clean.

## File & Interface Contracts
- Relevant files:
  - `internal/loaders/codex.go`
  - `internal/loaders/codex_test.go`
  - `internal/thermal/types.go`
  - `internal/thermal/aggregator.go`
  - `internal/render/models.go`
  - `internal/render/mix.go`
