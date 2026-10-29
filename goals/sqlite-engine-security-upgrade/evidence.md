# Evidence: Upgrade modernc.org/sqlite to remediate FTS5 and aggregate CVEs

Date: 2026-10-26
Project: thermal-streak
Goal: sqlite-engine-security-upgrade

## Changes

- `go.mod` / `go.sum`: upgraded `modernc.org/sqlite` from `v1.37.0` to
  `v1.61.0` via `go get modernc.org/sqlite@v1.61.0 && go mod tidy`.
  Transitive upgrades: `modernc.org/libc v1.79.0`, `modernc.org/memory
  v1.12.1`, `modernc.org/mathutil v1.7.1`, `golang.org/x/sys v0.49.0`.
- Verified the embedded engine reports SQLite `3.54.1`, which is at or above
  the required `3.53.2` fix level.
- `internal/loaders/sqlite_dsn_test.go`: added
  `TestSQLiteEngineVersionAndCVEMitigations`. It asserts
  `sqlite_version() >= 3.53.2` and corrupts every FTS5 `%_data` segment page
  by forcing the big-endian `szLeaf` header field at bytes 2..3 to `1`, then
  runs a `MATCH` query under `recover()` and requires a clean error rather than
  a panic or heap overflow.

## Verification

```text
$ CGO_ENABLED=0 go build ./cmd/thermal
(exit 0)

$ CGO_ENABLED=1 go test -race ./internal/loaders -run TestSQLiteEngineVersionAndCVEMitigations -v
--- PASS: TestSQLiteEngineVersionAndCVEMitigations (0.02s)
    sqlite_dsn_test.go:289: malformed FTS5 leaf rejected with: database disk image is malformed: fts5: corruption found reading blob 137438953473 from table "ft" (267)
PASS

$ CGO_ENABLED=0 go test ./...
ok  github.com/jadmadi/thermal/cmd/thermal
ok  github.com/jadmadi/thermal/internal/audit
ok  github.com/jadmadi/thermal/internal/changelog
ok  github.com/jadmadi/thermal/internal/loaders
ok  github.com/jadmadi/thermal/internal/pricing
ok  github.com/jadmadi/thermal/internal/render
ok  github.com/jadmadi/thermal/internal/share
ok  github.com/jadmadi/thermal/internal/theme
ok  github.com/jadmadi/thermal/internal/thermal
ok  github.com/jadmadi/thermal/internal/tui
ok  github.com/jadmadi/thermal/scripts/benchmark

$ CGO_ENABLED=1 go test -race ./...
ok  github.com/jadmadi/thermal/cmd/thermal
ok  github.com/jadmadi/thermal/internal/audit
ok  github.com/jadmadi/thermal/internal/changelog
ok  github.com/jadmadi/thermal/internal/loaders
ok  github.com/jadmadi/thermal/internal/pricing
ok  github.com/jadmadi/thermal/internal/render
ok  github.com/jadmadi/thermal/internal/share
ok  github.com/jadmadi/thermal/internal/theme
ok  github.com/jadmadi/thermal/internal/thermal
ok  github.com/jadmadi/thermal/internal/tui
ok  github.com/jadmadi/thermal/scripts/benchmark

$ ./scripts/simulated_user_gate.sh
Simulated User Gate Summary: 77 passed, 0 failed (out of 77 checks)
RELEASE GATE PASSED
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
  "goal_slug": "sqlite-engine-security-upgrade",
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
