# Plan: Make thermal upgrade previewable, scriptable, and non-destructive by default

## Execution Steps
- [ ] **Phase 1 — Option model and parser**: Introduce `upgradeOptions`, `parseUpgradeArgs`, and the documented flag set. Reject unknown flags with exit `2`. Wire `main.go` to pass args and writers. Acceptance: `TestUpgradeUnknownFlag` and `TestUpgradeFlagMatrix` pass; `upgrade --force-now` exits `2` before any HTTP call.
  - Flags: `--check`, `--dry-run`, `--json`, `--yes`, `--force`.
  - Reject positionals and short flags not in the set.
  - Keep `--help` delegated to the help goal, but defend with an early return.
- [ ] **Phase 2 — Preview state machine**: Implement `--check` (metadata only), `--dry-run` (metadata + resolved plan), and the no-mutation guarantees. Buffer JSON and write once. Acceptance: `TestUpgradeCheckNoMutation`, `TestUpgradeDryRunNoMutation`, and `TestUpgradeJSONContract` pass; md5 and inode unchanged.
  - Inject `Client`, `APIURL`, and `Executable`.
  - Ensure `.old`/`.new` are never created in preview modes.
  - Define `--check --json` and `--dry-run --json` schemas exactly as in `facts.md`.
- [ ] **Phase 3 — Confirmation and mutation safety**: Add TTY detection, the `[y/N]` prompt, the non-TTY `--yes` requirement, atomic swap preservation, and failure rollback. Acceptance: `TestUpgradeNonTTYRefusal`, `TestUpgradeDecline`, `TestUpgradeFailureRollback`, `TestUpgradeReadOnlyTarget`, and `TestUpgradeTruncatedArchive` pass.
  - Decline exits `3`; no download begins.
  - All download/extract/chmod work completes before the first rename.
  - On any error, assert the target md5 is unchanged and no partial artifacts remain.
  - `--force` prints and confirms the version transition.
- [ ] **Phase 4 — Help, docs, gate, and closure**: Add the upgrade help page, README/MIGRATION/project.jsonc updates, the simulated-user-gate safety section, and close the goal.
  - Run every AC in `goal.md` and record command, exit code, stdout/stderr, and md5 before/after in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(cli): make upgrade previewable and non-destructive by default (goals/cli-upgrade-nondestructive-safety-flags/goal.md)`.
  - Certify with `sila goals certify cli-upgrade-nondestructive-safety-flags --evidence=goals/cli-upgrade-nondestructive-safety-flags/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `upgrade --help` on a copy | exit 0, flags listed, md5 unchanged |
| AC-002 | `upgrade --check --json` | valid JSON, `action:"check"`, no download |
| AC-003 | `upgrade --dry-run --json` | `action:"dry-run"`, `mutated:false` |
| AC-004 | `upgrade --json --yes` (stub current) | `action:"up-to-date"` |
| AC-005 | `upgrade --json --yes` (stub newer) | `action:"upgraded"`, `mutated:true` |
| AC-006 | `upgrade --force-now` | exit 2, unknown flag named |
| AC-007 | `upgrade </dev/null` | exit 3, refusal, no mutation |
| AC-008 | PTY `n` / `y` | `3` / proceed |
| AC-009 | stub 404 | exit 1, original intact |
| AC-010 | truncated archive | exit 1, original intact |
| AC-011 | `go test -race ./cmd/thermal/...` | exit 0 |
| AC-012 | `./scripts/simulated_user_gate.sh` | exit 0 |

## Rollback
- The preview paths are additive; reverting them restores the old destructive default, which is unacceptable. If a regression is found, disable mutation by default (`--check` behavior) and keep `--yes` as the only mutating path until fixed.
- Keep the injected seams even if the UI changes; they are required for tests.
- Never roll back the non-TTY refusal or the unknown-flag rejection.

## Phase 5 — Attribution & Registry Closure
- [ ] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-upgrade-nondestructive-safety-flags/plan.md`; (2) commit with `(goals/cli-upgrade-nondestructive-safety-flags/goal.md)` using a `fix(cli):` subject; (3) certify via `sila goals certify cli-upgrade-nondestructive-safety-flags --evidence=goals/cli-upgrade-nondestructive-safety-flags/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
