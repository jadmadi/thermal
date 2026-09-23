# Facts: Declared Model Identity, Router Alias Normalization & Canonical Catalog

## Architectural Invariants & Constraints
- Clean-Room Implementation: Implemented entirely from scratch in idiomatic Go without copying source code, tables, or ASTs from third-party repositories. Retains 100% sole copyright ownership by Jad Madi.
- Declared Exact Resolution: No fuzzy string matching or regex guessing. Resolution priority: Exact ID -> Canonical Name -> Declared Harness-Scoped Alias -> Unresolved with explicit fallback.
- Zero Allocations on Hot Paths: Parsing and alias lookup must use pre-indexed hash maps or frozen lookups with zero heap allocations per event.

## File & Interface Contracts
- Relevant files:
  - `internal/loaders/modelname.go`
  - `internal/loaders/modelname_test.go`
  - `internal/pricing/models.go`
