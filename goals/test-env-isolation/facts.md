# Facts: Make transcript verification independent of ambient tool-root overrides

## Observed Evidence
Poisoned-environment run (2026-10-15, revision `911144c`):

```bash
env CODEX_HOME=/tmp/empty-tool-home GROK_HOME=/tmp/empty-tool-home DSH_HOME=/tmp/empty-tool-home \
    HERMES_HOME=/tmp/empty-tool-home ANTIGRAVITY_APP_DATA_DIR=/tmp/empty-tool-home \
    THERMAL_CONFIG=/tmp/empty-tool-home/none.config go test ./... -count=1
```

Failures:

| Package | Test | Ambient variable |
| :--- | :--- | :--- |
| `internal/thermal` | `TestScanSessionReceipts_Codex` | `CODEX_HOME` — `ScanSessionReceipts` at `internal/thermal/receipt.go:770` prefers the env override over the test's temporary home (`expected 1 codex receipt, got 0`). |
| `internal/loaders` | `TestLoadUserConfig_INI`, `TestLoadUserConfig_StructuredJSON`, `TestLoadUserConfig_FlatJSON`, `TestAllTools_ConfigOverrides` | `THERMAL_CONFIG` — `internal/loaders/config.go:21` reads the override before the test's fixture path. |

With only the tool-root overrides set (no `THERMAL_CONFIG`), the loaders package passes; with only `THERMAL_CONFIG` set, the thermal package passes. Both variables must be isolated per test.

Also observed while wiring the promise-claim-evidence gate suites: the release gate exports fixture overrides (`CODEX_HOME`, `GROK_HOME`, `DSH_HOME`, `HERMES_HOME`, `OPENCODE_HOME`, `AGY_HOME`, `ZED_HOME`, `NO_COLOR`, …) for its CLI checks, and running `go test` inside that environment reproduced the `TestScanSessionReceipts_Codex` failure. The gate now runs Go suites in a neutral `env -u …` environment as a workaround; the tests themselves remain ambient-sensitive for any developer or CI shell that sets these variables.

## Constraints
- Test-only change: production discovery and override semantics stay as they are.
- No weakened assertions: the failing tests must still verify the same behavior after isolation.
- Use `t.Setenv` (which also restores after the test and forbids `t.Parallel` where incompatible); for tests that must see the override, set it explicitly to a fixture path.
- Keep the fix cross-platform (Windows `USERPROFILE`/`APPDATA`/`LOCALAPPDATA` variants included where relevant).

## File & Interface Contracts
- `internal/thermal/receipt_test.go`: `TestScanSessionReceipts_Codex` (and audit the file for other env-reading scans).
- `internal/loaders/config_test.go` (or wherever `TestLoadUserConfig_*`/`TestAllTools_ConfigOverrides` live): `THERMAL_CONFIG` isolation.
- `internal/audit/*_test.go`: audit reads `state_5.sqlite` under the home passed in; verify it stays ambient-insensitive (the poisoned run shows it passing today).

## Verification Contract
- Poisoned-environment suite: the exact `env …` command above exits 0 for every package.
- `go test -race ./internal/thermal ./internal/loaders` exits 0 in a normal shell.
- Common final checks from `tasks/plan.md`.
