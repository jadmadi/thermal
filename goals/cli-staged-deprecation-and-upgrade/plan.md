# Plan: CLI Staged Deprecation Warnings & Self-Updater Migration Notifications

## Execution Steps

- [x] **Phase 1 — Deprecation Warning Engine in CLI**:
  - Implement a thread-safe `emitDeprecationWarning(item, replacement, removalVersion, anchor string)` in `cmd/thermal/main.go`.
  - Format output clearly to `os.Stderr` without ANSI escapes if `--no-color` is set or stderr is not a TTY.
  - Wire example deprecation checks for any upcoming deprecated syntax or aliases.

- [x] **Phase 2 — Upgrade Command Enhancement**:
  - Update `runUpgrade()` in `cmd/thermal/upgrade.go` to display a pointer to `docs/MIGRATION.md` and `https://thermal.jadmadi.net/migration` upon successful upgrade.
  - Update `cmd/thermal/upgrade_test.go` to verify post-upgrade messaging.

- [x] **Phase 3 — Stderr vs Stdout / JSON Parity Tests**:
  - Add unit tests in `cmd/thermal/args_test.go` confirming stderr emission while stdout contains clean, valid JSON parseable by `json.Unmarshal`.
  - Ensure exit codes are strictly preserved.

- [x] **Phase 4 — Release Gate Integration**:
  - Add deprecation warning simulation check to `scripts/simulated_user_gate.sh`.
  - Confirm all simulated user test checks pass cleanly (exit 0).

- [x] **Phase 5 — Attribution Commit**:
  - Commit changes with subject: `feat(cli): add staged deprecation warnings & upgrade migration alerts (goals/cli-staged-deprecation-and-upgrade/goal.md)`
