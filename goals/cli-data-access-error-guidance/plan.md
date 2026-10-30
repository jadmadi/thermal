# Plan: Turn every data-access failure into a named condition and a runnable next step

## Execution Steps
- [x] **Phase 1 — Typed condition model**: Add `DataCondition`,
  `DataAccessError`, and `classifyDataAccess` in `internal/loaders/pathcheck.go`;
  reuse the tool display-name table. Acceptance:
  `TestDataConditionClassification` covers ENOENT, EACCES, ELOOP, ENOTDIR, and
  file/dir mismatches.
- [x] **Phase 2 — Registry and explicit `--db` adoption**: Replace the two
  `fmt.Errorf` sites in `registry.go`, extend `pathcheck.go` branches, and keep
  the existing precise `--db` messages. Acceptance:
  `TestMissingDefaultEnrichment`, `TestDBPathConditions`, and
  `TestValidDBWithMissingDefault` pass.
  - Default path: display path, sources, all-tools report, `--db`.
  - Explicit path: existing message and fix, plus condition-specific fix.
  - Valid `--db` never falls back.
- [x] **Phase 3 — Rendering and channels**: Update `printLoadError` for the
  typed error, add the single `--verbose` raw line, and assert stdout empty.
  Acceptance: `TestLoadErrorChannels` and `TestVerboseRawLine` pass.
  - No stack traces, no driver dump.
  - `--json` keeps stdout valid and excludes the human fix.
- [x] **Phase 4 — Help examples and docs**: Add the prerequisite line or swap
  the flagship example; update README and `project.jsonc`; add the gate cases.
  Acceptance: `TestHelpExamplesRunnable` and `simulated_user_gate.sh` pass.
- [x] **Phase 5 — Evidence and closure**: Record the condition matrix, raw
  outputs, and channel results in `evidence.md`; commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | missing OpenCode `weekly --chart` | enriched error, exit 1 |
| AC-002 | missing Codex | data directory + sources hint |
| AC-003 | file for Claude dir | `expects a directory` |
| AC-004 | missing `--db` | precise path error |
| AC-005 | symlink loop | `too many levels of symbolic links` |
| AC-006 | mode 000 dir | `permission denied`, chmod fix |
| AC-007 | `--verbose` | one raw line, no stack |
| AC-008 | `--json` missing tool | valid JSON, human fix not on stdout |
| AC-009 | valid `--db`, missing default | loads, exit 0 |
| AC-010 | help examples | runnable or prerequisite stated |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | simulated user gate | exit 0 |

## Rollback
- Keep the typed error even if the extra lines are too verbose; shorten the
  fix block rather than reverting to a bare path.
- Do not change the explicit `--db` wording without a MIGRATION note if a test
  or doc depends on it.
- If the inventory probe is slow on a network HOME, omit the
  "another tool has data" line and keep the rest.
- Never auto-create a missing directory to turn an error into a success.

## Phase 6 — Attribution & Registry Closure
- [x] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `fix(cli): name data-access conditions and offer runnable next steps
  (goals/cli-data-access-error-guidance/goal.md)`; (3) certify via
  `sila goals certify cli-data-access-error-guidance
  --evidence=goals/cli-data-access-error-guidance/evidence.md`;
  (4) run `sila goals` and commit the updated `goals/README.md` and
  certification receipts; verify `sila goals --json` reports
  `unverified_count: 0`.
