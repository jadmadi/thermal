# Goal: Upgrade modernc.org/sqlite to remediate FTS5 and aggregate CVEs

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_sqlite-engine-security-upgrade_1793262089`


## Goal Description
Upgrade `modernc.org/sqlite` from `v1.37.0` (which embeds SQLite 3.49.1) to `v1.61.0` (which embeds SQLite >= 3.53.2) in `go.mod` and `go.sum` to remediate five OpenVuln engine memory-corruption findings: `BUG-R2-S2-A1-H4`, `BUG-R2-S2-A2-H1`, `BUG-R2-S2-A2-H3`, `BUG-R2-S2-A2-H2` (CVE-2026-11822 `fts5LeafSeek` OOB read and CVE-2026-11824 `fts5ChunkIterate` heap buffer overflow write), and `BUG-R2-S2-A1-H2` (CVE-2025-6965 `AggInfo` 16-bit signed column index truncation).

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 208
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

## Files to Touch
- Edit: `go.mod`
- Edit: `go.sum`
- Edit: `internal/loaders/sqlite_dsn_test.go`

## Requirements
1. Upgrade `modernc.org/sqlite` to `v1.61.0` (and its transitive `modernc.org/libc`, `modernc.org/memory`, `modernc.org/mathutil` dependencies via `go get modernc.org/sqlite@v1.61.0 && go mod tidy`) while keeping `CGO_ENABLED=0`.
2. Add a deterministic unit test `TestSQLiteEngineVersionAndCVEMitigations` in `internal/loaders/sqlite_dsn_test.go` that queries `SELECT sqlite_version()` and asserts the embedded SQLite version is `>= 3.53.2` (lexicographically or parsed semver `major=3, minor>=53, patch>=2`), and verifies that a malformed FTS5 continuation-page `szLeaf < 4` query returns a clean SQLite error (`database disk image is malformed`) rather than crashing the process.
3. Preserve all existing loader behavior, DSN formatting (`sqliteReadOnlyDSN`), and sub-10ms cached execution targets.

## Acceptance Criteria
- AC-001: `CGO_ENABLED=0 go test -race ./internal/loaders -run TestSQLiteEngineVersionAndCVEMitigations` (cwd `/mnt/Jad/github/lab/thermal-streak`) exits 0 and confirms `sqlite_version() >= 3.53.2`.
- AC-002: A crafted FTS5 table with `szLeaf = 1` in `fts_idx_data` queried via `MATCH` returns an error without crashing or overflowing heap memory.
- AC-003: `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` exit 0.

## References
- **Shared Understanding & Fact Sheet**: [`goals/sqlite-engine-security-upgrade/facts.md`](facts.md)
- **Execution Plan**: [`goals/sqlite-engine-security-upgrade/plan.md`](plan.md)

## Memory Plan
- **Prime**: run `sila prime` (CLI) or `prime_context` (MCP) at task start and record the briefing timestamp.
- **Recall**: run `sila recall "<symptom or design question>"` before implementation; record each query and match count.
- **Writes**: record at least one `memo`, `lesson`, or `case` grounded to a file in this goal; use a `case` when the goal introduces a reusable mechanism.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "<summary>" --open="..."` at session end and record the handoff ID.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`, per `memory-protocol-goal-contract`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/sqlite-engine-security-upgrade/goal.md)` or `[goal: sqlite-engine-security-upgrade]` attribution under a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` prefix.
3. Goal is certified (`goals/sqlite-engine-security-upgrade/certification.json` with `verified: true` or a signed `audit-receipt.json`) and any markdown evidence contains an explicit standalone `Verdict: PASS` declaration.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0` for delivered goals, with a clean `git status`.
