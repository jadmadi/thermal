# Plan: Never swallow -h/--help and add command-scoped help

## Execution Steps
- [ ] **Phase 1 — Inventory and command registry**: Enumerate every dispatch case, `isReportWord` value, flag, alias, and JSON envelope currently implemented. Create `cmd/thermal/help.go` with the `commandSpec` table and a parity test that fails when a dispatch case or report word has no spec. Acceptance: `go test ./cmd/thermal -run TestCommandSpecParity` passes; adding a fake verb without a spec fails.
  - Inventory inputs: `cmd/thermal/main.go` dispatch switch, `isReportWord`, `validateReportFlags`, `flag.*Var` registrations, `internal/render/help.go`.
  - Include aliases: `prompt`, `status`, `mimo`, `oc`, `cmd`, `whale`, `zc`, `ccode`, `dsh`, `deepseek`, `hermes`, `nous`.
  - Include retired verbs and their retirement messages.
- [ ] **Phase 2 — Global interception and precedence**: Add a raw-argv help scan before `flag.Parse()` and before dispatch. Implement the precedence table (help > version > validation > execution). Resolve the context (global, verb, tool, tool+verb, help-about-help, unknown). Acceptance: `TestHelpInterceptionMatrix` covers all negative cases in `goal.md` and passes.
  - Replace the `len(os.Args)==2` special case with the general resolver.
  - Ensure `--json`, `--no-color`, `--verbose`, and positional order do not change the outcome.
  - Ensure `--help` after `--` is still honoured (document if not).
- [ ] **Phase 3 — Scoped renderer and side-effect isolation**: Implement `RenderCommandHelp` with the page anatomy, width-aware wrapping, and color rules. Short-circuit before `loadUsage`, `pricing.Load`, `runUpgrade`, `runLive`, and cache writes. Acceptance: `TestHelpHasNoSideEffects` passes with a temp HOME and a read-only cache; `thermal upgrade --help` leaves a disposable copy's md5 unchanged.
  - Add a `helpOnly` early return in `main()` before the upgrade/live/statusline cases.
  - Guard `runUpgrade` so it cannot be called from a help path.
  - Add PTY byte assertions for the absence of alternate-screen sequences.
- [ ] **Phase 4 — Tests, gate, docs, and closure**: Add golden pages at 80/120 columns, the simulated-user-gate help section, README/MIGRATION/project.jsonc updates, and close the goal.
  - AC commands from `goal.md` are executed and recorded in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `feat(cli): intercept help globally and add command-scoped help pages (goals/cli-help-interception-and-scoped-help/goal.md)`.
  - Certify with `sila goals certify cli-help-interception-and-scoped-help --evidence=goals/cli-help-interception-and-scoped-help/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | loop over 14 verbs with `--help` | all exit 0, all contain `Usage:` |
| AC-002 | `thermal daily --help` | flags present, no date row |
| AC-003 | `cmp <(thermal daily --help) <(thermal help daily)` | exit 0 |
| AC-004 | `thermal live --help` in PTY | no `\x1b[?1049h`, no frame |
| AC-005 | disposable `upgrade --help` | md5 unchanged, no `.old`/`.new` |
| AC-006 | empty HOME help | exit 0, no data error |
| AC-007 | `thermal help nope` | exit 1, suggestion |
| AC-008 | `thermal help help` | exit 0, both forms documented |
| AC-009 | `--json` + `--help` | help, not JSON |
| AC-010 | `NO_COLOR=1`, 80-col check | 0 ANSI, no long lines |
| AC-011 | `go test -race ./cmd/thermal/... ./internal/render/...` | exit 0 |
| AC-012 | `./scripts/simulated_user_gate.sh` | exit 0 |

## Rollback
- Revert the interception block and `cmd/thermal/help.go`; keep the global `RenderHelp` path untouched.
- If a scoped page causes a data-pipeline regression, restore the previous `daily --help` behavior only behind a temporary `--legacy-help` escape hatch, document it in MIGRATION, and remove it in the next minor.
- Do not revert the side-effect fixes for `upgrade --help` and `live --help`; those are safety regressions.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-help-interception-and-scoped-help/plan.md`; (2) commit with `(goals/cli-help-interception-and-scoped-help/goal.md)` using a `feat(cli):` subject; (3) certify via `sila goals certify cli-help-interception-and-scoped-help --evidence=goals/cli-help-interception-and-scoped-help/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
