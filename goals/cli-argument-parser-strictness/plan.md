# Plan: Reject malformed and contradictory arguments instead of silently dropping them

## Execution Steps
- [x] **Phase 1 — Flag metadata and parser skeleton**: Create `cmd/thermal/args.go` with the `flagSpec` table for every value flag, `parseInvocation`, and `usageError`. Do not wire it into dispatch yet. Acceptance: unit tests parse valid, missing, empty, invalid, and `=` forms correctly.
  - Enumerate flags from `flag.*Var` and the manual scanner; eliminate duplicates.
  - Define value kinds, defaults, ranges, enums, and placeholders.
  - Implement `--` termination and interspersed positionals.
- [x] **Phase 2 — Strict parsing and error voice**: Replace the `flag.*Var` block and the manual scanner with `parseInvocation`. Add canonical errors for missing, empty, invalid, out-of-range, unknown, and unexpected positional. Acceptance: `TestArgumentMatrix`, `TestPositionalMatrix`, and `TestErrorVoice` pass; the manual is not printed for value errors.
  - Preserve `-h`, `-v`, `-d`, `-f` behavior or delegate to the help/version goals.
  - Ensure `--last --json` does not consume `--json`.
  - Ensure unknown flags fail in every position.
- [x] **Phase 3 — Conflict, range, and semantic validation**: Add positional-vs-`--tool` conflict detection, reversed date ranges, `--last >= 1`, `--top >= 0`, duration/ceiling positivity, and the existing per-command flag rules. Acceptance: `TestConflictMatrix` passes; every case exits `2` with a ≤4-line message.
  - Keep `validateReportFlags` semantics except the documented changes.
  - Move report-only flag errors into the new error voice.
  - Ensure `--verbose` adds detail without changing the first line or exit code.
- [x] **Phase 4 — Tests, fuzz, gate, docs, and closure**: Add the 100k fuzz test, golden stderr, simulated-user-gate section, README/MIGRATION/project.jsonc updates, and close the goal.
  - Record the full AC matrix in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(cli): reject malformed and contradictory arguments (goals/cli-argument-parser-strictness/goal.md)`.
  - Certify with `sila goals certify cli-argument-parser-strictness --evidence=goals/cli-argument-parser-strictness/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `daily --last abc` | exit 2, empty stdout, message names `--last` |
| AC-002 | six invalid values | each exit 2, no report |
| AC-003 | `--weeks abc` | exit 2, ≤4 lines, no manual |
| AC-004 | `opencode --tool codex daily` | exit 2, conflict |
| AC-005 | `--tool=` / `--tool ""` | exit 2, non-empty |
| AC-006 | reversed dates | exit 2, swap suggestion |
| AC-007 | `daily extra third` | exit 2, unexpected argument |
| AC-008 | `daily --metirc cost` | exit 2, names `--metirc` |
| AC-009 | `daily --last --json` | exit 2, missing value |
| AC-010 | error voice grep | no stdlib/flag jargon |
| AC-011 | `go test -race ./cmd/thermal/...` | exit 0, fuzz 100k |
| AC-012 | `./scripts/simulated_user_gate.sh` | exit 0 |

## Rollback
- Keep the old parser behind a build-tag-free fallback only if a production regression is found; the fallback must still reject invalid values rather than silently defaulting.
- Revert only the error-wording changes if they break golden tests, not the strictness.
- Never restore silent numeric fallback; that is the defect this goal exists to remove.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-argument-parser-strictness/plan.md`; (2) commit with `(goals/cli-argument-parser-strictness/goal.md)` using a `fix(cli):` subject; (3) certify via `sila goals certify cli-argument-parser-strictness --evidence=goals/cli-argument-parser-strictness/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
