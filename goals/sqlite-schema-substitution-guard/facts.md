# Facts: Reject SQLite view and virtual-table substitution across all loaders

## Observed Evidence
- In SQLite, referencing a name in a `FROM` clause resolves views and virtual tables transparently unless the application checks `sqlite_master`.
- In `internal/loaders/sqlite.go:279`, `hasColumn` runs `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, which in SQLite resolves `VIEW` column names by compiling the view's `SELECT` statement (`sqlite3ViewGetColumnNames`).
- Even when a loader checks `hasTable` for one table (e.g. `session_v2` in `sqlite.go:41` or `model_usage` in `zcode.go:31`), secondary tables queried in the same loader (`session`, `message`, `turn_usage`, `sessions`, `message_nodes`) are not verified to be ordinary B-tree tables (`type = 'table' AND rootpage > 0`).
- `internal/loaders/zed.go:112` checks `type='table'` for `threads`, whereas `muse.go:35`, `sqlite.go:455`, `sqlite.go:631`, `codex.go:71`, and `agy.go:114` query table names directly without `rootpage > 0` verification.

## Architectural Invariants & Constraints
- Ordinary SQLite B-tree tables always have `type = 'table'` and `rootpage > 0` in `sqlite_master` (`VIRTUAL TABLE` rows have `type = 'table'` and `rootpage = 0` or `NULL`; views have `type = 'view'`).
- Checking `sqlite_master` takes microseconds on an already-opened connection and preserves Thermal's `<10ms` cached execution ceiling.
- Absent optional tables (such as `session_v2` on legacy OpenCode stores, or `prompt_history` on older Devin stores) must not error when absent (`sql.ErrNoRows`); only present objects that are not ordinary B-tree tables must be rejected.

## File & Interface Contracts
- `internal/loaders/sqlite.go`:
  - `func rejectSubstitutableObjects(db *sql.DB, names ...string) error`
  - Harden `hasColumn` to verify `rejectSubstitutableObjects(db, table) == nil` before calling `pragma_table_info`.
- Call `rejectSubstitutableObjects` at the top of `LoadOpenCodeData`, `LoadMiMoCodeData`, `LoadDevinData`, `LoadMuseData`, `LoadZCodeData`, `loadCodexFromStateDB`, `LoadHermesData`, `loadAgyWorkspaces`, and `LoadZedData`.

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/loaders -run 'TestRejectSubstitutableObjects|TestSqliteReadOnlyDSN'
CGO_ENABLED=0 go test -race ./...
./scripts/simulated_user_gate.sh
```
