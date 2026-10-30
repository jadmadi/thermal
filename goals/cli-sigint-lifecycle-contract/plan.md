# Plan: Treat SIGINT as a normal stop for live and stream

## Execution Steps
- [x] **Phase 1 — Signal context and normal-stop classifier**: Add
  `sigint.go` with `notifyLiveSignals`, `isNormalStop`, and the second-signal
  handler; pass a context into the TUI. Acceptance: `TestIsNormalStop`,
  `TestLiveSIGINT`, and `TestLiveSecondSIGINT` pass.
  - First signal: cancel, restore, exit 0.
  - Second signal: restore, exit 130.
  - Real TUI errors still exit 1.
- [x] **Phase 2 — Stream shutdown**: Check `ctx.Done()` around each write,
  always end on a line boundary, and keep SIGPIPE behavior. Acceptance:
  `TestStreamSIGINTLineBoundary` and `TestStreamBrokenPipe` pass.
- [x] **Phase 3 — Scan cancellation and cache safety**: Thread the context
  into directory scans; do not persist a delta when canceled; clean stale temp
  files. Acceptance: `TestInterruptedScanNoCache` and
  `TestNoStaleTempAfterSignal` pass.
- [x] **Phase 4 — PTY gate, docs, evidence**: Add the PTY case to
  `simulated_user_gate.sh`, document `q or Ctrl+C; exit 0` in help and README,
  add MIGRATION, update `project.jsonc`, record evidence, commit, certify,
  reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | PTY `live`, one SIGINT | exit 0, no failure text, restore |
| AC-002 | PTY `live`, two SIGINTs | exit 130, restore |
| AC-003 | PTY `--stream`, SIGINT | exit 0, valid last line |
| AC-004 | SIGTERM | exit 143, no failure |
| AC-005 | injected TUI error | exit 1, one failure line |
| AC-006 | cache scan after SIGINT | no lock/pid/tmp |
| AC-007 | run after SIGINT | byte-identical to clean |
| AC-008 | non-TTY `live` | exit 2, guidance |
| AC-009 | `go test -race ...` | exit 0 |
| AC-010 | simulated user gate | exit 0 |
| AC-011 | `live --help` | `Ctrl+C` and exit 0 documented |
| AC-012 | MIGRATION | six parts, detection command |

## Rollback
- Keep the normal-stop classifier even if the exit code is debated; the
  failure text must not return.
- If exit `0` for SIGINT breaks a wrapper, document `--require-exit` in a
  follow-up; do not reintroduce `live monitor failed` for an interrupt.
- Keep the second-signal forced exit `130`; it is the escape hatch if cleanup
  hangs.
- Do not make scans uncancellable to simplify the cache path.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `fix(cli): treat SIGINT as a normal live/stream stop
  (goals/cli-sigint-lifecycle-contract/goal.md)`; (3) certify via
  `sila goals certify cli-sigint-lifecycle-contract
  --evidence=goals/cli-sigint-lifecycle-contract/evidence.md`; (4) run
  `sila goals` and commit the updated `goals/README.md` and certification
  receipts; verify `sila goals --json` reports `unverified_count: 0`.
