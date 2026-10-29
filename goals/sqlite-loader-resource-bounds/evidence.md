# Evidence: Enforce SQLite query timeouts, value size limits, and read-only regression tests

Date: 2026-10-26
Project: thermal-streak
Goal: sqlite-loader-resource-bounds

## Changes

- `internal/loaders/bounded.go` (new):
  - `storeQueryBudget = 30 * time.Second` and
    `storeValueLimit = 16 << 20` (16 MiB `SQLITE_LIMIT_LENGTH`).
  - `openBoundedStore(dbPath, mmapSize)` opens through
    `sqliteReadOnlyDSN`, pins `SetMaxOpenConns(1)` / `SetMaxIdleConns(1)`,
    acquires the pooled `*sql.Conn`, applies
    `sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_LENGTH, storeValueLimit)`, and
    returns the connection to the pool.
  - `queryStore` / `queryStoreWithBudget` wrap `db.QueryContext` and return
    `boundedRows`, whose `Close` cancels the query context.
  - `queryRowStore` wraps `db.QueryRowContext` with a deferred cancel.
  - `rowIterator` lets `foldDayModelProjectRows` accept both bounded and
    plain rows.
- Adopted `openBoundedStore` / `queryStore` / `queryRowStore` in all SQLite
  loaders: `sqlite.go` (OpenCode, MiMoCode, Devin, message-level fallback),
  `muse.go`, `zcode.go`, `codex.go`, `hermes.go`, `agy.go`, `zed.go`.
  Redundant per-`db` `PRAGMA mmap_size`/`cache_size` Exec calls were removed
  because the DSN sets both pragmas on every pooled connection.
- A count query in the Devin path was moved ahead of the main row scan so a
  single-connection store never deadlocks on a nested query.
- `internal/loaders/sqlite_dsn_test.go`: added
  `TestReadOnlyHotJournalAndSuperJournalNoUnlink` (forged journal with a
  valid `-mj…9` super-journal trailer pointing at a canary file; loader must
  fail closed and leave the database, journal, and canary byte-identical),
  `TestReadOnlyWALCloseNoCheckpoint` (WAL database with 64 uncheckpointed
  4 KiB frames; read-only loader reads all rows, and close must not rewrite
  the database or WAL or unlink `-wal`/`-shm`), and
  `TestBoundedStoreValueLimitAndTimeout` (a 16 MiB + 1 KiB value fails with
  `SQLITE_TOOBIG`; a recursive query under a 1 ms budget fails with
  `context deadline exceeded`).

## Verification

```text
$ CGO_ENABLED=1 go test -race ./internal/loaders -run 'TestReadOnlyHotJournalAndSuperJournalNoUnlink|TestReadOnlyWALCloseNoCheckpoint|TestBoundedStoreValueLimitAndTimeout' -v
--- PASS: TestReadOnlyHotJournalAndSuperJournalNoUnlink (0.00s)
--- PASS: TestReadOnlyWALCloseNoCheckpoint (0.03s)
--- PASS: TestBoundedStoreValueLimitAndTimeout (0.08s)
PASS

$ CGO_ENABLED=0 go test ./internal/loaders
ok  github.com/jadmadi/thermal/internal/loaders

$ CGO_ENABLED=1 go test -race ./...
ok  all packages

$ ./scripts/simulated_user_gate.sh
Simulated User Gate Summary: 77 passed, 0 failed (out of 77 checks)
RELEASE GATE PASSED
```

Observed failure classes (independently reproduced with the same modernc
`Limit` call):

```text
limit err: <nil>
value err: string or blob too big (18)
timeout err: context deadline exceeded
```

Contract note: the goal text lists `CGO_ENABLED=0 go test -race`, but the Go
toolchain rejects that combination (`-race requires cgo`). Race detection was
therefore run with `CGO_ENABLED=1`, and the CGO-free portability path was run
with `CGO_ENABLED=0 go test ./...` and `CGO_ENABLED=0 go build ./cmd/thermal`.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "sqlite-loader-resource-bounds",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "sqlite read-only DSN schema substitution guard", "matches": 5}],
  "memos": [{"id": "fact_b307b8b2", "text": "Keep modernc.org/sqlite at or above v1.61.0 (embedded SQLite >= 3.53.2) and pin a sqlite_version() regression test so FTS5 and aggregate CVEs cannot silently regress."}],
  "lessons": [{"id": "fact_1b89398a", "symptom": "SQLite read-only opens could compile attacker-controlled VIEW bodies via pragma_table_info and had no value or time bounds"}],
  "cases": [{"id": "fact_de2a21fc", "intent": "SQLite loaders must reject schema substitution and bound query resources"}],
  "facts_verify": {"exit_code": 2},
  "handoff": {"id": "hoff_cd63055e31a4"},
  "waivers": [{"verb": "facts", "reason": "The tamper-evident ledger is intact (26061/26061 blocks verified); exit 2 is pre-existing grounding drift reporting 279 missing files from earlier relocated projects, unrelated to this goal."}]
}
```
