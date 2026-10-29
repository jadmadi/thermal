# Goal: Enforce SQLite query timeouts, value size limits, and read-only regression tests

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_sqlite-loader-resource-bounds_1793262093`


## Goal Description
Enforce resource bounds on SQLite loader connections (`SQLITE_LIMIT_LENGTH` capped at 16 MiB via `modernc.org/sqlite.Limit` and per-query `context.WithTimeout` deadlines via `QueryContext`/`QueryRowContext`) to remediate `BUG-R2-S2-A1-H3` (unbounded query CPU/memory/temp-disk DoS), and add explicit regression tests for `BUG-R2-S1-A2-H3`, `BUG-R2-S1-A2-H1`, and `BUG-R2-S1-A2-H2` verifying that hot-journal super-journal records and WAL close-time checkpoints fail closed without modifying or unlinking any files.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sqlite-schema-substitution-guard
- **Sequence**: 210
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

## Files to Touch
- Create: `internal/loaders/bounded.go`
- Edit: `internal/loaders/sqlite.go`
- Edit: `internal/loaders/muse.go`
- Edit: `internal/loaders/zcode.go`
- Edit: `internal/loaders/codex.go`
- Edit: `internal/loaders/hermes.go`
- Edit: `internal/loaders/agy.go`
- Edit: `internal/loaders/zed.go`
- Edit: `internal/loaders/sqlite_dsn_test.go`

## Requirements
1. **Bounded SQLite Store Helper (`internal/loaders/bounded.go`)**:
   - Define `const storeQueryBudget = 30 * time.Second` and `const storeValueLimit = 16 << 20` (16 MiB `SQLITE_LIMIT_LENGTH`).
   - Implement `openBoundedStore(dbPath string, mmapSize int64) (*sql.DB, error)` which opens the database via `sqliteReadOnlyDSN(dbPath, mmapSize)`, sets `db.SetMaxOpenConns(1)`, acquires the pooled connection with a timeout context, sets `sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_LENGTH, storeValueLimit)`, and returns `db`.
   - Implement `queryStore(db *sql.DB, query string, args ...any) (*boundedRows, error)` and `queryRowStore(db *sql.DB, dest []any, query string, args ...any) error` wrapping `db.QueryContext` and `db.QueryRowContext` with `storeQueryBudget` (and `defer cancel()` / `boundedRows.Close()`).
2. **Loader Adoption**: Update all 8 SQLite loaders (`sqlite.go`, `muse.go`, `zcode.go`, `codex.go`, `hermes.go`, `agy.go`, `zed.go`) to open connections via `openBoundedStore` and execute queries through `queryStore` / `queryRowStore`.
3. **Regression Tests (`internal/loaders/sqlite_dsn_test.go`)**:
   - Add `TestReadOnlyHotJournalAndSuperJournalNoUnlink`: create a database and a forged `<db>-journal` with a super-journal trailer pointing to a canary victim file in `t.TempDir()`, run a loader against it, and assert the canary victim file, `<db>-journal`, and `<db>` remain byte-identical.
   - Add `TestReadOnlyWALCloseNoCheckpoint`: create a WAL-mode database with uncheckpointed frames, run a loader against it, and assert `-wal` and `-shm` are not unlinked and `<db>` is not rewritten on `db.Close()`.
   - Add `TestBoundedStoreValueLimitAndTimeout`: verify that values exceeding `storeValueLimit` (16 MiB) or queries exceeding a short test deadline fail closed with an error rather than hanging or exhausting memory.

## Acceptance Criteria
- AC-001: `CGO_ENABLED=0 go test -race ./internal/loaders -run 'TestReadOnlyHotJournalAndSuperJournalNoUnlink|TestReadOnlyWALCloseNoCheckpoint|TestBoundedStoreValueLimitAndTimeout'` exits 0.
- AC-002: Forged `<db>-journal` super-journal records never unlink victim files; WAL stores never checkpoint or unlink `-wal` on close.
- AC-003: `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` exit 0.

## References
- **Shared Understanding & Fact Sheet**: [`goals/sqlite-loader-resource-bounds/facts.md`](facts.md)
- **Execution Plan**: [`goals/sqlite-loader-resource-bounds/plan.md`](plan.md)

## Memory Plan
- **Prime**: run `sila prime` (CLI) or `prime_context` (MCP) at task start and record the briefing timestamp.
- **Recall**: run `sila recall "<symptom or design question>"` before implementation; record each query and match count.
- **Writes**: record at least one `memo`, `lesson`, or `case` grounded to a file in this goal; use a `case` when the goal introduces a reusable mechanism.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "<summary>" --open="..."` at session end and record the handoff ID.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`, per `memory-protocol-goal-contract`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/sqlite-loader-resource-bounds/goal.md)` or `[goal: sqlite-loader-resource-bounds]` attribution under a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` prefix.
3. Goal is certified (`goals/sqlite-loader-resource-bounds/certification.json` with `verified: true` or a signed `audit-receipt.json`) and any markdown evidence contains an explicit standalone `Verdict: PASS` declaration.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0` for delivered goals, with a clean `git status`.
