# Plan: Enforce SQLite query timeouts, value size limits, and read-only regression tests

## Execution Steps
- [x] **Phase 1 — Bounded Store Primitives**: Create `internal/loaders/bounded.go` implementing `openBoundedStore`, `boundedRows`, `queryStore`, and `queryRowStore` with `storeQueryBudget` (30s) and `storeValueLimit` (16 MiB `SQLITE_LIMIT_LENGTH`).
- [x] **Phase 2 — Loader Adoption**: Switch `sqlite.go`, `muse.go`, `zcode.go`, `codex.go`, `hermes.go`, `agy.go`, and `zed.go` to `openBoundedStore`, `queryStore`, and `queryRowStore`.
- [x] **Phase 3 — Security Regression Suite & Evidence**: Add `TestReadOnlyHotJournalAndSuperJournalNoUnlink`, `TestReadOnlyWALCloseNoCheckpoint`, and `TestBoundedStoreValueLimitAndTimeout` in `internal/loaders/sqlite_dsn_test.go`; run `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh`; write `goals/sqlite-loader-resource-bounds/evidence.md` with `Verdict: PASS`.
- [x] **Phase 4 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/sqlite-loader-resource-bounds/plan.md`; (2) commit with `(goals/sqlite-loader-resource-bounds/goal.md)` using a `fix(loaders):` subject; (3) certify via `sila goals certify sqlite-loader-resource-bounds --evidence=goals/sqlite-loader-resource-bounds/evidence.md`; (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts.
