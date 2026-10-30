# Evidence: Document exit codes and give upgrade global-flag parity

Date: 2026-10-30
Binary: `/tmp/thermal-g8` (source build)
Goal: `cli-automation-exit-and-upgrade-parity`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | `upgrade --check --offline --json` | exit 1; `offline: release metadata is not cached`; no `unknown flag` |
| AC-002 | offline path with a nil HTTP client | returns before any request; dead API is never contacted |
| AC-003 | `upgrade --no-update-check --check` | parses as a no-op |
| AC-004 | `upgrade --width 100` | exit 2; `not applicable`; `--json` alternative |
| AC-005 | `upgrade --bogus` | exit 2; `unknown flag` |
| AC-006 | `upgrade --json` non-interactive | exit 3; one JSON object on stdout |
| AC-007 | global help | `Exit Codes:` block |
| AC-008 | upgrade help | `Exit codes: 0 ok ...` and `--offline` |
| AC-009 | version / missing data / bad tool / refusal | 0 / 1 / 2 / 3 |
| AC-010 | `docs/CLI-CONTRACT.md` | table and examples per code |
| AC-011 | `go test ./cmd/thermal/... ./internal/render/...` | ok |
| AC-012 | `simulated_user_gate.sh` | 167 passed, 0 failed |

## Observed Output

```text
$ thermal upgrade --check --offline --json
exit 1
{
  "action": "error",
  "current": "dev",
  "error": "offline: release metadata is not cached; run thermal upgrade --check without --offline",
  "mutated": false,
  "type": "upgrade"
}
```

```text
$ thermal upgrade --width 100
exit 2
thermal: --width is not applicable to upgrade
  upgrade output is fixed-width; use --json for structured output
```

```text
$ thermal version          -> exit 0
$ thermal opencode         -> exit 1
$ thermal --tool open      -> exit 2
$ thermal upgrade --json   -> exit 3
```

```text
$ thermal --help
  Exit Codes:
    0 / 1 / 2 / 3   ok · data/IO error · usage error · refused non-interactive mutation

$ thermal upgrade --help
  Exit codes: 0 ok · 1 data/IO error · 2 usage error · 3 refused. --offline
```

## Implementation Notes

- `upgradeOptions` gained `Offline`; `parseUpgradeArgs` accepts
  `--offline`/`--no-update-check` and rejects `--width`/`--fresh` with
  `not applicable` plus `--json`.
- `runUpgradeWithOptions` returns the offline error before any fetch.
- Added `cmd/thermal/exit.go` with `ExitOK`/`ExitFailure`/`ExitUsage`/
  `ExitRefused` and `printExitCodes`.
- Global help gained an `Exit Codes:` section; upgrade help gained the same
  contract plus the offline semantics.
- Added `docs/CLI-CONTRACT.md`, README exit table, and MIGRATION entry.
- Added upgrade flag, offline, and exit-code help tests.

## Verification Commands

```bash
go test ./cmd/thermal/... ./internal/render/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `167 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-automation-exit-and-upgrade-parity",
  "prime": {"ran": true},
  "recall": [
    {"query": "unknown flag for upgrade offline", "matches": 5},
    {"query": "exit code 3", "matches": 5}
  ],
  "memos": [{"id": "fact_06f2d881"}],
  "lessons": [{"id": "fact_352ded1d"}],
  "cases": [{"id": "fact_6e42b154"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_481e9814f372"}
}
```

Verdict: PASS
