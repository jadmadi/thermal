# Evidence: Reject SQLite view and virtual-table substitution across all loaders

Date: 2026-10-26
Project: thermal-streak
Goal: sqlite-schema-substitution-guard

## Changes

- `internal/loaders/sqlite_dsn.go`: the read-only DSN now appends
  `&_pragma=trusted_schema=0`, so unsafe schema-embedded objects and
  user-defined functions are refused on every pooled connection.
- `internal/loaders/sqlite.go`: added
  `rejectSubstitutableObjects(db, names...)`. It reads
  `SELECT type, rootpage FROM sqlite_master WHERE lower(name) = lower(?)`
  and returns `refusing store: %q is not an ordinary table` unless the object
  is `type = 'table'` with `rootpage > 0`. Missing optional tables are skipped.
- `hasColumn` now returns false unless the object passes `hasTable` and
  `rejectSubstitutableObjects`; a `VIEW` body is never compiled by
  `pragma_table_info`.
- Guards wired before the first query in all eight SQLite loaders:
  `LoadOpenCodeData` (`session_v2`, `session`, `message`),
  `LoadMiMoCodeData` (`session`, `message`),
  `LoadDevinData` (`sessions`, `message_nodes`, `prompt_history`),
  `LoadMuseData` (`sessions`),
  `LoadZCodeData` (`model_usage`, `turn_usage`, `session`),
  `loadCodexFromStateDB` (`threads`, `thread_spawn_edges`),
  `LoadHermesData` (`sessions`, `session`, `session_model_usage`),
  `loadAgyWorkspaces` (`conversation_summaries`), and
  `LoadZedData` (`threads`).
- `internal/loaders/sqlite_dsn_test.go`: added
  `TestRejectSubstitutableObjects` (ordinary table accepted, absent table
  accepted, `VIEW` and FTS5 virtual table rejected, `hasColumn` false on a
  deliberately broken view) and `TestLoadersRejectSubstitutedSchemaObjects`
  (every loader rejects both a `VIEW` and `CREATE VIRTUAL TABLE ... USING
  fts5(...)` at its expected table name; Agy skips a substituted summaries
  store).

## Verification

```text
$ CGO_ENABLED=1 go test -race ./internal/loaders -run 'TestRejectSubstitutableObjects|TestLoadersRejectSubstitutedSchemaObjects' -v
--- PASS: TestRejectSubstitutableObjects (0.02s)
--- PASS: TestLoadersRejectSubstitutedSchemaObjects (0.13s)
    --- PASS: .../opencode/view
    --- PASS: .../opencode/virtual_table
    --- PASS: .../mimocode/view
    --- PASS: .../mimocode/virtual_table
    --- PASS: .../devin/view
    --- PASS: .../devin/virtual_table
    --- PASS: .../muse/view
    --- PASS: .../muse/virtual_table
    --- PASS: .../zcode/view
    --- PASS: .../zcode/virtual_table
    --- PASS: .../codex/view
    --- PASS: .../codex/virtual_table
    --- PASS: .../hermes/view
    --- PASS: .../hermes/virtual_table
    --- PASS: .../zed/view
    --- PASS: .../zed/virtual_table
    --- PASS: .../agy/conversation_summaries
PASS

$ CGO_ENABLED=0 go test ./...
ok  all packages (including all existing loader unit tests)

$ CGO_ENABLED=1 go test -race ./...
ok  all packages

$ ./scripts/simulated_user_gate.sh
Simulated User Gate Summary: 77 passed, 0 failed (out of 77 checks)
RELEASE GATE PASSED
```

Contract note: the goal text lists `CGO_ENABLED=0 go test -race`, but the Go
toolchain rejects that combination (`-race requires cgo`). Race detection was
therefore run with `CGO_ENABLED=1`, and the CGO-free portability path was run
with `CGO_ENABLED=0 go test ./...`.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "sqlite-schema-substitution-guard",
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
