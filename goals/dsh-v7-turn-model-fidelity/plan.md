# Plan: Read DSH v7 turn counts and session models

## Execution Steps
- [x] **Phase 1 — Red tests**: Add the version 7 fixtures and the failing
  tests before touching the loader. Acceptance:
  `go test ./internal/loaders/ -run 'TestLoadDshData_(TurnBoundary|Model)'
  -count=1` fails with 0 turns and an empty model.
  - `TestLoadDshData_TurnBoundaryFallback`
  - `TestLoadDshData_TurnBoundaryWinsOverSessionStats`
  - `TestLoadDshData_TurnFallbackWithoutSources`
  - `TestLoadDshData_ModelFromSessionLog`
  - `TestLoadDshData_ModelFromRequestHeader`
  - `TestLoadDshData_ModelSelectionSkipsLog`
  - `TestLoadDshData_ModelLogCorruptWarns`
  - `TestLoadDshData_ModelLogMissingSilent`
  - `TestLoadDshData_ModelLogWithoutRequestSilent`
  - `TestLoadDshData_ZeroTokenSessionSkipsLog`
- [x] **Phase 2 — Turn counts**: Add `turnBoundary` to `dshRows` and take
  the larger of `sessionStats.val.turns` and
  `turnBoundary.val.lastTurn`. Acceptance:
  `TestLoadDshData_TurnBoundaryFallback` and
  `TestLoadDshData_TurnBoundaryWinsOverSessionStats` pass; every earlier
  DSH test stays green.
- [x] **Phase 3 — Model fallback**: Index
  `<dataDir>/sessions/*/<id>/session*.jsonl.zstd`, head-decode the first
  `request/context` or `request/header` event for model-less active
  sessions, canonicalise through `modelName()`, and warn only on read or
  decode failure. Acceptance:
  `TestLoadDshData_ModelFromSessionLog`,
  `TestLoadDshData_ModelFromRequestHeader`,
  `TestLoadDshData_ModelSelectionSkipsLog`,
  `TestLoadDshData_ModelLogCorruptWarns`,
  `TestLoadDshData_ModelLogMissingSilent`,
  `TestLoadDshData_ModelLogWithoutRequestSilent`, and
  `TestLoadDshData_ZeroTokenSessionSkipsLog` pass.
- [x] **Phase 4 — Frozen snapshot parity**: Rebuild `/tmp/thermal-test`
  and run it against `/tmp/dshprobe`. Acceptance: 30 turns,
  `deepseek-v4.1-flash` and `deepseek-flash` present, unattributed tokens
  limited to `atria-dawn-preview`, cost about $2.69, and the token total
  still 79,381,433.
- [x] **Phase 5 — Docs**: Update `README.md` DSH source line and
  `AGENTS.md` section 2.D, and refresh `project.jsonc` `lastEdit`.
  Acceptance: both lines describe version 7 fields and the zstd fallback;
  no capability text changes.
- [x] **Phase 6 — Gates**: `gofmt -l` on touched files, `go vet ./...`,
  `go test -race ./internal/loaders/...`, `./scripts/check.sh
  --pre-commit`, and a live `thermal daily --tool=dsh --verbose`.
  Acceptance: all exit `0`; no new warnings.
- [x] **Phase 7 — Evidence and closure**: Write `evidence.md` with the
  snapshot receipts and a standalone `Verdict: PASS`; commit with
  `fix(loaders): ... (goals/dsh-v7-turn-model-fidelity/goal.md)`;
  certify; run `sila goals` and confirm `unverified_count: 0`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `go test -race ./internal/loaders/...` | exit 0, five named tests present |
| AC-002 | `DSH_HOME=/tmp/dshprobe thermal daily --tool=dsh --json` | 30 turns |
| AC-003 | same | models contain both DeepSeek ids; unattributed is the atria remainder |
| AC-004 | same | cost about 2.69 |
| AC-005 | `thermal daily --tool=dsh --verbose` | exit 0, no warnings |
| AC-006 | `gofmt -l internal/loaders/dsh.go internal/loaders/dsh_test.go` | no output; `go vet ./...` exit 0 |
| AC-007 | `./scripts/check.sh --pre-commit` | exit 0 |
| AC-008 | `sila goals --json` | implemented and verified, clean tree |

## Test Fixture Notes
- Use `t.TempDir()` with
  `<tmp>/storages/session_projcache/sessions/<id>.json` for projections
  and `<tmp>/sessions/<slug>/<id>/session.v4.jsonl.zstd` for logs.
- Compress log fixtures with `zstd.NewWriter`, the same helper pattern as
  `internal/loaders/zed_test.go`.
- A fixed createdAt of `1793846091097` keeps every fixture on
  2026-11-05 in the local timezone.
- A corrupt fixture is the bytes `not a zstd frame`.
- The "model set" isolation fixture writes corrupt bytes under the session
  id and asserts zero warnings, which proves the log was not opened.

## Rollback
Revert the single loader commit. The projections and logs are read-only,
so no data migration or cleanup is required.
