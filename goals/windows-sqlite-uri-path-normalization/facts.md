# Facts: Windows SQLite URI and Path Normalization

## Architectural Invariants & Constraints
- SQLite read-only mode (`mode=ro`) and memory-mapping (`PRAGMA mmap_size=268435456`) must remain intact across all platforms.
- Windows file paths containing drive letters (e.g. `C:\path\to\db.sqlite`) and spaces must format cleanly as SQLite DSNs.
- Backslashes must convert to forward slashes with `filepath.ToSlash` for SQLite URI consistency.
- Zero extra allocations on the hot query path.

## File & Interface Contracts
- Relevant files:
  - `internal/loaders/sqlite.go`: OpenCode, MiMoCode, and Devin loaders.
  - `internal/loaders/hermes.go`: Hermes state database loader.
  - `internal/loaders/agy.go`: Agy brain SQLite loader.
  - `internal/loaders/codex.go`: Codex state database loader.
  - `internal/loaders/zcode.go`: ZCode SQLite loader.
  - `internal/loaders/muse.go`: Muse SQLite loader.

