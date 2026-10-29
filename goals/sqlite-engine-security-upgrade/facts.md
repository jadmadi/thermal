# Facts: Upgrade modernc.org/sqlite to remediate FTS5 and aggregate CVEs

## Observed Evidence
- `go.mod:12` currently pins `modernc.org/sqlite v1.37.0`, which embeds transpiled SQLite `3.49.1` (`2025-02-18`).
- SQLite `3.49.1` is vulnerable to:
  - `CVE-2025-6965` (`BUG-R2-S2-A1-H2`): `findOrCreateAggInfoColumn` truncates `Expr.iAgg` to `int16` when a query references `> 32,767` aggregate columns, fixed in SQLite `3.50.2` (`modernc.org/sqlite >= v1.39.0`).
  - `CVE-2026-11822` and `CVE-2026-11824` (`BUG-R2-S2-A1-H4`, `BUG-R2-S2-A2-H1`, `BUG-R2-S2-A2-H3`, `BUG-R2-S2-A2-H2`): `fts5LeafSeek` OOB read and `fts5ChunkIterate` continuation-page `szLeaf < 4` negative-chunk integer underflow / heap buffer overflow write, fixed upstream in SQLite `3.53.2`.
- `go list -m -versions modernc.org/sqlite` shows `v1.61.0` is available in the module proxy.

## Architectural Invariants & Constraints
- Pure-Go portability (`CGO_ENABLED=0`) across `linux`, `darwin`, and `windows` must remain intact.
- Do not modify CLI flags or JSON output schemas.
- `sqliteReadOnlyDSN` in `internal/loaders/sqlite_dsn.go` must continue to work identically with the upgraded driver.

## File & Interface Contracts
- `go.mod` & `go.sum`: Upgrade `modernc.org/sqlite` to `v1.61.0`.
- `internal/loaders/sqlite_dsn_test.go`: Add `TestSQLiteEngineVersionAndCVEMitigations(t *testing.T)` verifying `SELECT sqlite_version()` is at least `3.53.2` and testing corrupted FTS5 segment rejection.

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/loaders -run TestSQLiteEngineVersionAndCVEMitigations
CGO_ENABLED=0 go test -race ./...
./scripts/simulated_user_gate.sh
```
