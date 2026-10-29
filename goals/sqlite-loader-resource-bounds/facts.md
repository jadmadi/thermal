# Facts: Enforce SQLite query timeouts, value size limits, and read-only regression tests

## Observed Evidence
- `BUG-R2-S2-A1-H3`: None of the SQLite loaders in `internal/loaders/` pass a `context.Context` with a deadline (`QueryContext` / `QueryRowContext`) or configure `sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_LENGTH, ...)`. The embedded SQLite build allows up to 1 GiB (`MAX_LENGTH = 1000000000`) per value and compiles out `PRAGMA hard_heap_limit` (`SQLITE_DEFAULT_MEMSTATUS=0`).
- `BUG-R2-S1-A2-H3`, `BUG-R2-S1-A2-H1`, `BUG-R2-S1-A2-H2`: While `e30bd4d` fixed the missing `file:` prefix via `sqliteReadOnlyDSN`, `internal/loaders/sqlite_dsn_test.go` only tests basic `SELECT id FROM test` and does not include regression tests for hot-journal super-journal deletion refusal (SQLite code 776 `SQLITE_READONLY_ROLLBACK`) or close-time WAL checkpoint refusal.

## Architectural Invariants & Constraints
- Setting `db.SetMaxOpenConns(1)` and applying `sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_LENGTH, 16<<20)` on the single pooled connection guarantees that per-connection SQLite limits persist across all subsequent queries on `db`.
- Every `context.WithTimeout` must have its `cancel()` called when `QueryRowContext` returns or when `boundedRows.Close()` is invoked, preventing context/goroutine leaks (Stack-Specific Standard #2).

## File & Interface Contracts
- `internal/loaders/bounded.go`:
  - `openBoundedStore(dbPath string, mmapSize int64) (*sql.DB, error)`
  - `queryStore(db *sql.DB, query string, args ...any) (*boundedRows, error)`
  - `queryRowStore(db *sql.DB, dest []any, query string, args ...any) error`
- `internal/loaders/sqlite_dsn_test.go`:
  - `TestReadOnlyHotJournalAndSuperJournalNoUnlink`
  - `TestReadOnlyWALCloseNoCheckpoint`
  - `TestBoundedStoreValueLimitAndTimeout`

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/loaders -run 'TestReadOnlyHotJournalAndSuperJournalNoUnlink|TestReadOnlyWALCloseNoCheckpoint|TestBoundedStoreValueLimitAndTimeout'
CGO_ENABLED=0 go test -race ./...
./scripts/simulated_user_gate.sh
```
