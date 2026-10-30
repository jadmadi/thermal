# Evidence: Make no-data a successful, schema-stable state

Date: 2026-10-30
Binary: `/tmp/thermal-g1` (source build, `CGO_ENABLED=0`)
Goal: `cli-empty-data-success-contract`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | `HOME=<empty> thermal` | exit 0; 7-line onboarding on stdout; stderr empty |
| AC-002 | `HOME=<empty> thermal --json` | `type:"leaderboard"`, `state:"no-data"`, `results:[]`, `ranking:{key,order}`; exit 0 |
| AC-003 | `daily/weekly/monthly/projects/models/mix/stats/trend --json` | each returns its own `type` with `state:"no-data"`; exit 0 |
| AC-004 | `HOME=<empty> thermal --require-data` | exit 1; same stdout; stderr `thermal: no data found (--require-data)` |
| AC-005 | `HOME=<empty> thermal sources` / `--json` | exit 0; `type:"sources"` only for `sources` |
| AC-006 | `HOME=<empty> thermal statusline` | exit 0; stdout and stderr empty |
| AC-007 | `HOME=<empty> thermal statusline --json` | `state:"no-data"`, zero totals; exit 0 |
| AC-008 | `HOME=<empty> thermal statusline --verbose` | exit 0; stdout empty; one stderr line |
| AC-009 | `HOME=<empty> thermal share` | exit 1; stdout empty; short stderr message |
| AC-010 | empty `~/.claude` fixture | `Claude is installed but has no sessions yet.` |
| AC-011 | sentinel prompt fixture | no session content in onboarding or JSON |
| AC-012 | `--json --width 40` vs `--width 200` | byte-identical after removing `generatedAt` |
| AC-013 | `go test ./cmd/thermal/... ./internal/thermal/...` | ok |
| AC-014 | `./scripts/simulated_user_gate.sh` | 159 passed, 0 failed |
| AC-015 | `docs/MIGRATION.md` | six-part entry added under v0.22.0 |

## Observed Output

```text
$ HOME=<empty> thermal
  No AI coding sessions found yet.
  Run one prompt in OpenCode, Claude, Codex, or any supported tool, then
  re-run thermal.

  Next: thermal sources   inspect all 15 source paths and states
        thermal --help    supported tools and flags
exit 0
```

```text
$ HOME=<empty> thermal daily --json
{
  "costBasis": "total",
  "data": [],
  "generatedAt": "...",
  "next": {"help": "thermal --help", "inventory": "thermal sources --json"},
  "state": "no-data",
  "totals": {"activeDays": 0, "cost": 0, "tokens": 0, ...},
  "type": "daily"
}
exit 0
```

```text
$ HOME=<empty> thermal --require-data
exit 1
stdout: onboarding block
stderr: thermal: no data found (--require-data)
```

## Implementation Notes

- Added `--require-data` to `thermal.Options` and the flag parser.
- Added `cmd/thermal/nodata.go` with the onboarding renderer, `noDataNext`,
  `markNoDataPayload`, `dailyHasActivity`, and `loadUsageForReport`.
- Every report keeps its requested `type`; `state`/`next` are additive.
- `stats` and `trend` now initialise their collection fields to empty slices so
  no-data payloads are `[]`, not `null`.
- A source that exists but has no activity no longer becomes a zero-value
  leaderboard row.
- `share` remains a deliberate failure with a short message instead of the
  15-row inventory.
- `thermal sources --json` remains the only `type:"sources"` producer.

## Verification Commands

```bash
go test ./...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All three passed. `simulated_user_gate.sh` reported `159 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-empty-data-success-contract",
  "prime": {"ran": true},
  "recall": [
    {"query": "no supported tool data found", "matches": 5},
    {"query": "empty json sources payload", "matches": 5}
  ],
  "memos": [{"id": "fact_5924cf13"}],
  "lessons": [{"id": "fact_023ba3a7"}],
  "cases": [{"id": "fact_426dcbb5"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_cb118c8b569d"}
}
```

Verdict: PASS
