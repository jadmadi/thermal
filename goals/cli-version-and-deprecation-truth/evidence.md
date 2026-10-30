# Evidence: Make version output and deprecation warnings factually true

Date: 2026-10-30
Binary: `/tmp/thermal-g7` and `/tmp/thermal-stamped`
Goal: `cli-version-and-deprecation-truth`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | unstamped `thermal version` | `dev (built from source; version not stamped)`, `commit: unknown`, `built: unknown` |
| AC-002 | `version --json` | `stamped:false`, `semver:""`, `commit:"unknown"`, `date:"unknown"` |
| AC-003 | stamped test build | `v9.9.9`, `stamped:true`, real commit/date |
| AC-004 | `version --help` | states the source-build fallback |
| AC-005 | `thermal -v` | one `verbose diagnostics enabled` stderr line |
| AC-006 | `thermal -v --json` | valid JSON stdout; announcement stderr only |
| AC-007 | `thermal --license` | no `v0.14.0`, no `docs/MIGRATION.md`, absolute URL |
| AC-008 | `thermal nous` | no expired version; names `hermes` |
| AC-009 | deprecation table test | `IsPast` rejects `v0.14.0` for `v0.21.0` |
| AC-010 | source grep | no `will be removed in v0.14.0` |
| AC-011 | `go test ./...` | ok |
| AC-012 | `simulated_user_gate.sh` | 164 passed, 0 failed |

## Observed Output

```text
$ thermal version
thermal dev (built from source; version not stamped)
  commit: unknown
  built:  unknown

$ thermal version --json
{"version":"dev (built from source; version not stamped)",
 "stamped":false,"semver":"","commit":"unknown","date":"unknown",...}
```

```text
$ thermal -v
thermal: verbose diagnostics enabled (stderr only)
<normal leaderboard on stdout>
```

```text
$ thermal --license
stderr:
thermal: warning: --license is deprecated; use thermal license
  details: https://jadmadi.net/project/thermal/migration#deprecated-license-flag
```

```text
$ thermal nous
stderr:
thermal: warning: nous is deprecated; use hermes
  details: https://jadmadi.net/project/thermal/migration#deprecated-tool-nous
thermal: Nous Hermes has no data at ~/.hermes/state.db
  ...
```

```text
$ go build -ldflags "...Version=v9.9.9 -X ...Commit=abc123 -X ...Date=..." ...
$ /tmp/thermal-stamped version --json
{"version":"v9.9.9","stamped":true,"semver":"v9.9.9","commit":"abc123",
 "date":"2026-01-01T00:00:00Z",...}
```

## Implementation Notes

- Added `version.Info`, `Current`, `IsPast`, and a numeric SemVer comparison.
- `String()` now says `dev (built from source; version not stamped)`.
- Text output always prints `commit` and `built`; JSON adds `stamped` and
  `semver`.
- `-v`/`--verbose` announces itself once on stderr.
- Deprecation warnings no longer carry a version when none is true; they carry
  the replacement and an absolute migration URL.
- The early `--license` path was fixed alongside the normal dispatch path.
- Help text now states the source-build fallback.

## Verification Commands

```bash
go test ./...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `164 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-version-and-deprecation-truth",
  "prime": {"ran": true},
  "recall": [
    {"query": "deprecated will be removed in v0.14.0", "matches": 5},
    {"query": "version dev built from source", "matches": 5}
  ],
  "memos": [{"id": "fact_2a97f27a"}],
  "lessons": [{"id": "fact_39ff4aeb"}],
  "cases": [{"id": "fact_263847df"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_1278cc1696de"}
}
```

Verdict: PASS
