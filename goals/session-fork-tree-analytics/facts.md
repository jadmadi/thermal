# Facts: Session Fork & Branch Tree Yield Analytics

## Architectural Invariants & Constraints
- Source SQLite connections remain read-only (?mode=ro) with PRAGMA mmap_size.
- Preserve disjoint token calculations.
- Sub-10ms cached execution target preserved.

## File & Interface Contracts
- Relevant files:
  - `internal/loaders/codex.go`
  - `internal/loaders/codex_test.go`
  - `internal/thermal/yield.go`
  - `internal/thermal/yield_test.go`
  - `internal/render/yield.go`
