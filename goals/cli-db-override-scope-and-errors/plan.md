# Plan: Make --db scope explicit and turn raw SQLite failures into actionable errors

## Execution Steps
- [x] **Phase 1 — Scope gate**: Detect `opts.DBPath != ""` and count the tools in scope before loading. Reject aggregate/all-tools contexts with exit `2` and the scope message. Acceptance: `TestDBScopeMatrix` passes; `--db /nonexistent daily` prints no report.
  - Count scope from `opts.Tool` and the positional resolution.
  - Allow single-tool `thermal <tool>`, `<tool> <report>`, `live --tool`, `statusline --tool`, `share --tool`.
  - Reject `--tool all`, `--tool auto`, and no tool selection.
- [x] **Phase 2 — Path validation**: Add `internal/loaders/pathcheck.go` with `ToolPathShape`, `ValidateDBPath`, and `DBPathError`. Validate existence, readability, file/directory shape, regular-file type, non-empty, and SQLite header for SQLite tools. Acceptance: `TestDBPathValidation` covers every tool shape and the failure matrix.
  - Probe only the exact expected filename inside a directory.
  - Show both given and resolved paths when symlinks differ.
  - Keep directory tools permissive for empty directories.
- [x] **Phase 3 — Error translation and no fallback**: Map driver errors to the canonical messages, keep raw text behind `--verbose`, and ensure the aggregate path never passes `""` when `--db` is set. Audit the config-override fallback and add a warning. Acceptance: `TestDBErrorTranslation` and `TestDBNeverFallsBack` pass; no `(10)`/`(26)` in non-verbose output.
  - Use `errors.As` to detect `DBPathError` and driver errors.
  - Add a sentinel-row fixture to prove the default DB is not read.
  - Add the config-override warning test.
- [x] **Phase 4 — Help, gate, docs, and closure**: Update `--db` help, README troubleshooting, MIGRATION, project.jsonc, and the simulated-user-gate `--db` section. Record the AC matrix and close the goal.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(cli): scope --db to one tool and translate path errors (goals/cli-db-override-scope-and-errors/goal.md)`.
  - Certify with `sila goals certify cli-db-override-scope-and-errors --evidence=goals/cli-db-override-scope-and-errors/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `--db /nonexistent daily` | exit 2, scope error, no report |
| AC-002 | `--tool opencode --db /nonexistent` | exit 1, path does not exist |
| AC-003 | `--tool opencode --db /tmp` | exit 1, directory message |
| AC-004 | `--tool opencode --db /etc/passwd` | exit 1, not SQLite |
| AC-005 | valid OpenCode fixture | exit 0, fixture rows |
| AC-006 | valid Claude dir | exit 0 |
| AC-007 | OpenCode with Claude dir | exit 1, expected filename |
| AC-008 | `--verbose` bad path | raw driver text on stderr |
| AC-009 | sentinel default HOME | no report, sentinel absent |
| AC-010 | `--db` with all/no tool | exit 2 |
| AC-011 | `go test -race ./cmd/thermal/... ./internal/loaders/...` | exit 0 |
| AC-012 | `./scripts/simulated_user_gate.sh` | exit 0 |

## Rollback
- The scope gate is the only behavior change that can break existing scripts; keep the validation and error translation even if the gate is temporarily relaxed.
- If a supported SQLite variant fails the header pre-flight, exempt that variant by tool name and document it; do not remove the pre-flight globally.
- Never restore the silent aggregate fallback; if necessary, return an explicit warning and exit `1`.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-db-override-scope-and-errors/plan.md`; (2) commit with `(goals/cli-db-override-scope-and-errors/goal.md)` using a `fix(cli):` subject; (3) certify via `sila goals certify cli-db-override-scope-and-errors --evidence=goals/cli-db-override-scope-and-errors/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
