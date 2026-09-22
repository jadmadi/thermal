# Facts: Robust SQLite Rows Iteration Safety and rows.Err Checking

## Architectural Facts & Grounding

1. **`database/sql` Iteration Semantics**:
   In Go's standard library `database/sql`, `rows.Next()` returns a boolean indicating if a next row is available. If an error occurs during iteration (e.g. disk I/O failure, corrupt database page, or broken connection), `rows.Next()` returns `false`. Calling `rows.Err()` is the only mechanism provided to distinguish between clean EOF and premature termination.

2. **Parity with JSONL Diagnostics**:
   In `jsonl-scan-diagnostics`, Thermal established that unchecked `scanner.Err()` was an architectural risk. The exact same risk applies to database row iteration: silent truncation results in undercounted token metrics, broken streaks, and inaccurate attribution.

3. **Affected Loaders**:
   - `internal/loaders/codex.go`: `loadCodexFromStateDB` iterates thread records.
   - `internal/loaders/hermes.go`: `LoadHermesData` iterates session records.
   - `internal/loaders/muse.go`: `LoadMuseData` iterates session prompt counts.
   - (Note: `internal/loaders/sqlite.go` already checks `rows.Err()` in `foldDayModelProjectRows` and `deltaRows`).
