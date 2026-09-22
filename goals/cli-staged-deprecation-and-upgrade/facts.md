# Facts: CLI Staged Deprecation Warnings & Self-Updater Migration Notifications

## Architectural Invariants & Constraints

1. **Staged Path Deprecation Lifecycle**:
   - In accordance with the `version-migration-guide` standard, breaking changes should follow the *Staged Path* wherever feasible: surface deprecation warnings in version `N-1` before dropping syntax or changing behavior in version `N`.
   - Avoid cold breaks when deprecating flags, positional verbs, or tool aliases.

2. **Strict Output Stream Separation (Rule A.10 & 3.3)**:
   - Deprecation warnings MUST write exclusively to `os.Stderr`.
   - Deprecation warnings MUST NEVER leak into `os.Stdout` or corrupt machine-readable `--json` output (`jq` parseability must remain 100% clean).
   - Exit code must remain 0 for successful commands even when emitting deprecation warnings (warnings are non-fatal).

3. **Consistent Warning Structure & Migration Anchor**:
   - Deprecation notices must provide clear, actionable guidance:
     `thermal: warning: [item] is deprecated and will be removed in v[X.Y]; use [replacement] (see docs/MIGRATION.md#[anchor])`
   - Quoted exact strings allow scripts and search tools to identify deprecated call sites programmatically.

4. **Self-Updater (`thermal upgrade`) Notification**:
   - When `thermal upgrade` successfully replaces the binary across minor or major versions, it should print a helpful pointer to the migration guide:
     `  migration guide: https://thermal.jadmadi.net/migration or docs/MIGRATION.md`
   - Keeps users aware of breaking changes or new capabilities immediately after self-updating.

5. **Deterministic Testing Invariants**:
   - Unit tests must verify that deprecated flags trigger stderr warnings while returning exit code 0.
   - Test JSON output with deprecation warnings to verify stdout remains valid JSON.
   - Pass `./scripts/simulated_user_gate.sh` with zero failures.

## File & Interface Contracts

- `cmd/thermal/main.go`: Deprecation warning helper and wiring across deprecated CLI flags/aliases.
- `cmd/thermal/upgrade.go`: Post-upgrade notification linking to the migration guide.
- `cmd/thermal/args_test.go`: Unit tests for CLI flag parsing and stderr warning capture.
- `cmd/thermal/upgrade_test.go`: Unit tests for upgrade messaging.
- `scripts/simulated_user_gate.sh`: Release gate verification for deprecation warnings under `--json` and `--no-color`.
