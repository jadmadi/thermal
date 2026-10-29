# Goal: Reject SQLite view and virtual-table substitution across all loaders

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_sqlite-schema-substitution-guard_1793262091`


## Goal Description
Enforce `PRAGMA trusted_schema = OFF` and `rejectSubstitutableObjects` (verifying `sqlite_master` rows have `type = 'table'` and `rootpage > 0`) before any query or `pragma_table_info` probe across all 8 SQLite loaders (`sqlite.go`, `muse.go`, `zcode.go`, `codex.go`, `hermes.go`, `agy.go`, and `zed.go`). This closes the OpenVuln schema-object substitution primitive (`BUG-R2-S2-A1-H1`) and cuts the delivery path to all FTS5/`fts5vocab`/aggregate view sinks (`BUG-R2-S2-A1-H4`, `BUG-R2-S2-A2-H1`, `BUG-R2-S2-A2-H3`, `BUG-R2-S2-A2-H2`, `BUG-R2-S2-A1-H2`).

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 209
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

## Files to Touch
- Edit: `internal/loaders/sqlite_dsn.go`
- Edit: `internal/loaders/sqlite.go`
- Edit: `internal/loaders/muse.go`
- Edit: `internal/loaders/zcode.go`
- Edit: `internal/loaders/codex.go`
- Edit: `internal/loaders/hermes.go`
- Edit: `internal/loaders/agy.go`
- Edit: `internal/loaders/zed.go`
- Edit: `internal/loaders/sqlite_dsn_test.go`

## Requirements
1. **Disable trusted schema in DSN**: Update `sqliteReadOnlyDSN` in `internal/loaders/sqlite_dsn.go` to include `&_pragma=trusted_schema=0` (or execute `PRAGMA trusted_schema = OFF;` on open) so SQLite refuses unsafe virtual tables and user-defined functions inside schemas.
2. **Implement `rejectSubstitutableObjects`**: In `internal/loaders/sqlite.go`, define `func rejectSubstitutableObjects(db *sql.DB, names ...string) error`. For each name, query `SELECT type, rootpage FROM sqlite_master WHERE lower(name) = lower(?)`. If `sql.ErrNoRows`, continue (missing optional tables are handled by each loader). If `typ != "table" || !rootpage.Valid || rootpage.Int64 <= 0`, return `fmt.Errorf("refusing store: %q is not an ordinary table", n)`.
3. **Harden `hasColumn`**: In `internal/loaders/sqlite.go`, make `hasColumn(db *sql.DB, table, column string) bool` return `false` immediately if `!hasTable(db, table)` or `rejectSubstitutableObjects(db, table) != nil`, preventing `pragma_table_info(?)` from compiling or expanding a `VIEW` body.
4. **Guard all 8 SQLite loaders before their first query**:
   - `LoadOpenCodeData` (`internal/loaders/sqlite.go`): `rejectSubstitutableObjects(db, "session_v2", "session", "message")`
   - `LoadMiMoCodeData` (`internal/loaders/sqlite.go`): `rejectSubstitutableObjects(db, "session", "message")`
   - `LoadDevinData` (`internal/loaders/sqlite.go`): `rejectSubstitutableObjects(db, "sessions", "message_nodes", "prompt_history")`
   - `LoadMuseData` (`internal/loaders/muse.go`): `rejectSubstitutableObjects(db, "sessions")`
   - `LoadZCodeData` (`internal/loaders/zcode.go`): `rejectSubstitutableObjects(db, "model_usage", "turn_usage", "session")`
   - `loadCodexFromStateDB` (`internal/loaders/codex.go`): `rejectSubstitutableObjects(db, "threads", "thread_spawn_edges")`
   - `LoadHermesData` (`internal/loaders/hermes.go`): `rejectSubstitutableObjects(db, "sessions", "session", "session_model_usage")`
   - `loadAgyWorkspaces` (`internal/loaders/agy.go`): `rejectSubstitutableObjects(db, "conversation_summaries")`
   - `LoadZedData` (`internal/loaders/zed.go`): `rejectSubstitutableObjects(db, "threads")`
5. **Unit tests**: Add table-driven tests in `internal/loaders/sqlite_dsn_test.go` verifying that each loader rejects a crafted SQLite store where an expected table name is replaced by a `CREATE VIEW` or `CREATE VIRTUAL TABLE ... USING fts5(...)`, while normal B-tree tables (`rootpage > 0`) continue to load cleanly.

## Acceptance Criteria
- AC-001: `CGO_ENABLED=0 go test -race ./internal/loaders -run TestRejectSubstitutableObjects` exits 0 and verifies that `VIEW` and `VIRTUAL TABLE` objects at any loader table name are rejected with `"is not an ordinary table"` before executing view SQL.
- AC-002: `hasColumn` on a `VIEW` returns `false` without compiling the view's `SELECT` body.
- AC-003: All existing loader unit tests (`CGO_ENABLED=0 go test -race ./internal/loaders/...`) pass with zero regressions.

## References
- **Shared Understanding & Fact Sheet**: [`goals/sqlite-schema-substitution-guard/facts.md`](facts.md)
- **Execution Plan**: [`goals/sqlite-schema-substitution-guard/plan.md`](plan.md)

## Memory Plan
- **Prime**: run `sila prime` (CLI) or `prime_context` (MCP) at task start and record the briefing timestamp.
- **Recall**: run `sila recall "<symptom or design question>"` before implementation; record each query and match count.
- **Writes**: record at least one `memo`, `lesson`, or `case` grounded to a file in this goal; use a `case` when the goal introduces a reusable mechanism.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "<summary>" --open="..."` at session end and record the handoff ID.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`, per `memory-protocol-goal-contract`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/sqlite-schema-substitution-guard/goal.md)` or `[goal: sqlite-schema-substitution-guard]` attribution under a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` prefix.
3. Goal is certified (`goals/sqlite-schema-substitution-guard/certification.json` with `verified: true` or a signed `audit-receipt.json`) and any markdown evidence contains an explicit standalone `Verdict: PASS` declaration.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0` for delivered goals, with a clean `git status`.
