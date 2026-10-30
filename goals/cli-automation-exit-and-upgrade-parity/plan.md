# Plan: Document exit codes and give upgrade global-flag parity

## Execution Steps
- [x] **Phase 1 — Exit constants and help block**: Add `exit.go` constants and
  `printExitCodes`; render the four-line block in global and upgrade help.
  Acceptance: `TestExitCodeHelp` and `TestExitConstants` pass.
- [x] **Phase 2 — Global flag disposition**: Add `FlagRule` and the matrix for
  upgrade; accept `--offline`/`--no-update-check`, reject `--width`/`--fresh`
  with `not applicable` + `--json`, keep `--bogus` unknown. Acceptance:
  `TestUpgradeGlobalFlags` and `TestKnownGlobalNeverUnknown` pass.
- [x] **Phase 3 — Offline cache path**: Add `Offline`, `Source`, and a cached
  release index; fail closed with a clear message; prove zero network on a dead
  API. Acceptance: `TestOfflineUsesCache`, `TestOfflineNoCache`,
  `TestOfflineNoNetwork`.
- [x] **Phase 4 — Exit-code matrix and docs**: Replace magic numbers with
  constants where safe, add the canonical table to `docs/CLI-CONTRACT.md`,
  README, MIGRATION, `project.jsonc`, and the gate. Acceptance:
  `TestExitCodeMatrix` and `simulated_user_gate.sh` pass.
- [x] **Phase 5 — Evidence and closure**: Record the flag matrix, offline
  behavior, and exit-code matrix; commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `upgrade --check --offline` | cache or offline error, no unknown |
| AC-002 | dead API + offline | zero HTTP requests |
| AC-003 | `upgrade --no-update-check` | exit 0, no-op |
| AC-004 | `upgrade --width 100` | exit 2, not applicable + `--json` |
| AC-005 | `upgrade --bogus` | exit 2, unknown flag |
| AC-006 | `upgrade --json` non-TTY | exit 3, one object |
| AC-007 | global help | exit-code block |
| AC-008 | upgrade help | block + `--offline` |
| AC-009 | version/data/usage/refusal | 0/1/2/3 |
| AC-010 | CLI contract doc | table matches tests |
| AC-011 | `go test -race ...` | exit 0 |
| AC-012 | simulated user gate | exit 0 |

## Rollback
- Keep `--offline` additive; if the cache path is unavailable, fail closed with
  the `--check` instruction rather than returning to "unknown flag".
- Keep `--width` rejection explicit; do not silently ignore it.
- Keep exit `3`; changing it would break existing wrappers.
- Do not add a fifth exit code for "no update available"; it is `0`.

## Phase 6 — Attribution & Registry Closure
- [x] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `fix(cli): document exit codes and accept global flags on upgrade
  (goals/cli-automation-exit-and-upgrade-parity/goal.md)`; (3) certify via
  `sila goals certify cli-automation-exit-and-upgrade-parity
  --evidence=goals/cli-automation-exit-and-upgrade-parity/evidence.md`;
  (4) run `sila goals` and commit the updated `goals/README.md` and
  certification receipts; verify `sila goals --json` reports
  `unverified_count: 0`.
