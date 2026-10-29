# Plan: Upgrade modernc.org/sqlite to remediate FTS5 and aggregate CVEs

## Execution Steps
- [x] **Phase 1 — Dependency Upgrade**: Run `go get modernc.org/sqlite@v1.61.0 && go mod tidy` in `/mnt/Jad/github/lab/thermal-streak` and verify `CGO_ENABLED=0 go build ./cmd/thermal` succeeds cleanly.
- [x] **Phase 2 — Engine Version & CVE Regression Test**: Add `TestSQLiteEngineVersionAndCVEMitigations` in `internal/loaders/sqlite_dsn_test.go` asserting `sqlite_version() >= 3.53.2` and verifying that a crafted FTS5 shadow table (`fts_idx_data` continuation leaf with `szLeaf < 4`) fails gracefully without process crash.
- [x] **Phase 3 — Verification & Evidence**: Run `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh`, and write `goals/sqlite-engine-security-upgrade/evidence.md` with an explicit standalone `Verdict: PASS` declaration and `sila/memory-writes/v1` block.
- [x] **Phase 4 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/sqlite-engine-security-upgrade/plan.md`; (2) commit with `(goals/sqlite-engine-security-upgrade/goal.md)` using a `fix(deps):` subject; (3) certify via `sila goals certify sqlite-engine-security-upgrade --evidence=goals/sqlite-engine-security-upgrade/evidence.md`; (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts.
