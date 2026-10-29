# Plan: Define machine-clean non-TTY stdout, exit codes, and silent statusline degradation

## Execution Steps
- [ ] **Phase 1 — Refactor `loadUsage` to return errors**: Change `loadUsage` from `os.Exit` on empty data to `(usageSet, error)` with a typed `ErrNoData`; update all callers to preserve their current behavior except statusline. Acceptance: `TestLoadUsageNoDataError` and all existing command tests pass.
  - Keep the bare `thermal` empty-state behavior delegated to the first-run goal.
  - Add `ErrNoData` with the inventory summary attached.
  - No behavior change for `daily`/`stats`/`share` yet.
- [ ] **Phase 2 — `live` channel and exit contract**: Move the non-TTY guidance to stderr, exit `2`, and verify `--json`/`--stream` remain valid and pipe-safe. Add SIGPIPE handling. Acceptance: `TestLiveNonTTYContract`, `TestLiveJSONPipe`, `TestStreamBrokenPipe`, and `TestLiveTTYUnchanged` pass.
  - No stdout guidance.
  - `--stream` implies NDJSON or is rejected with a clear message; choose and test.
  - PTY harness proves the TUI still launches.
- [ ] **Phase 3 — `statusline` no-data and JSON state**: Add `State` to `StatuslineData`, render empty text/no-data JSON, degrade silently on missing data and cache errors, and bound the no-data cache. Acceptance: `TestStatuslineNoData`, `TestStatuslineJSONState`, `TestStatuslineCacheDegradation`, and `TestStatuslineRecovery` pass.
  - Exit `0`, no stdout/stderr in text no-data mode.
  - `--json` emits the no-data object with `state`.
  - `--verbose` adds one stderr line.
  - Keep `prompt`/`status` aliases identical.
- [ ] **Phase 4 — Channel matrix, gate, docs, and closure**: Implement the full channel matrix test, the simulated-user-gate section, README prompt integration, MIGRATION, project.jsonc, and close the goal.
  - Record the matrix and latency in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(cli): clean non-TTY output and quiet statusline degradation (goals/cli-nontty-and-statusline-output-contract/goal.md)`.
  - Certify with `sila goals certify cli-nontty-and-statusline-output-contract --evidence=goals/cli-nontty-and-statusline-output-contract/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `live \| cat` | exit 2, stdout empty, stderr guidance |
| AC-002 | `live --json \| jq` | exit 0, valid JSON |
| AC-003 | `live --json --stream \| head -1` | exit 0, valid NDJSON, no error |
| AC-004 | empty HOME `statusline` | exit 0, both channels empty |
| AC-005 | empty HOME `statusline --json` | `state:"no-data"` |
| AC-006 | empty HOME `statusline --plain` | exit 0, empty |
| AC-007 | normal `statusline` | one line; JSON `state:"ok"` |
| AC-008 | empty HOME `share` | exit 1, stdout empty |
| AC-009 | channel matrix | all rows pass |
| AC-010 | ANSI checks | zero ANSI |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | gate | exit 0 |

## Rollback
- If `live` exit `2` breaks an internal script, that script must add `--json`; do not restore exit `0` with stdout guidance.
- If the statusline no-data JSON breaks a consumer, keep the text mode silent and gate the JSON state behind `--json` only.
- Keep the `loadUsage` error return even if a caller chooses to exit `1`; the refactor is required for quiet degradation.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-nontty-and-statusline-output-contract/plan.md`; (2) commit with `(goals/cli-nontty-and-statusline-output-contract/goal.md)` using a `fix(cli):` subject; (3) certify via `sila goals certify cli-nontty-and-statusline-output-contract --evidence=goals/cli-nontty-and-statusline-output-contract/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
