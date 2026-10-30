# Evidence: Treat SIGINT as a normal stop for live and stream

Date: 2026-10-30
Binary: `/tmp/thermal-g5` (source build)
Goal: `cli-sigint-lifecycle-contract`

## Acceptance Matrix

| AC | Test | Result |
| --- | --- | --- |
| AC-001 | PTY `live`, one SIGINT | exit 0; no `failed`/`killed`/`interrupted` text; restore sequences present |
| AC-002 | second SIGINT during shutdown | watchdog exits 130 after restoring the screen; a signal after a completed stop leaves exit 0 |
| AC-003 | PTY `live --json --stream`, SIGINT | exit 0; last line is complete NDJSON |
| AC-004 | PTY `live`, SIGTERM | exit 143; no failure text; restore sequences present |
| AC-005 | injected renderer error | `isNormalLiveStop` returns false; failure path kept |
| AC-006 | cache after interruption | no lock/pid/tmp files |
| AC-007 | run after SIGINT | clean startup, same output |
| AC-008 | non-TTY text `live` | exit 2 with stderr guidance (unchanged) |
| AC-009 | `go test ./cmd/thermal/... ./internal/tui/...` | ok |
| AC-010 | `simulated_user_gate.sh` | 164 passed, 0 failed |
| AC-011 | `live --help` | documents `q or Ctrl+C` |
| AC-012 | MIGRATION | exit-code change documented |

## Observed PTY Results

```text
$ thermal live --offline --no-update-check --no-color
<PTY; one SIGINT>
exit 0
contains "failed": false
contains "killed": false
restore sequences: true (\x1b[?1049l, \x1b[?25h, \x1b[?2004l)
```

```text
$ thermal live --offline --no-update-check --no-color
<PTY; SIGTERM>
exit 143
contains "failed": false
restore sequences: true
```

```text
$ thermal live --json --stream --offline --no-update-check
<PTY; SIGINT>
exit 0
last line: a complete snapshot object
```

## Implementation Notes

- `runLive` now receives the root context and passes `tea.WithContext(ctx)`.
- `isNormalLiveStop` treats nil, `context.Canceled`, and the Bubble Tea
  `program was killed` / `program was interrupted` errors as normal stops.
- A second-signal watchdog restores the alternate screen, cursor, and bracketed
  paste before exiting `130` when shutdown is still in progress.
- SIGTERM is tracked separately and exits `143`.
- Real renderer errors still print `live monitor failed` and exit `1`.
- The stream loop already stopped on SIGINT at a line boundary and is
  unchanged.

## Verification Commands

```bash
go test ./cmd/thermal/... ./internal/tui/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `164 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-sigint-lifecycle-contract",
  "prime": {"ran": true},
  "recall": [
    {"query": "live monitor failed interrupted", "matches": 5},
    {"query": "SIGINT exit code", "matches": 5}
  ],
  "memos": [{"id": "fact_9b03b389"}],
  "lessons": [{"id": "fact_bd884bc0"}],
  "cases": [{"id": "fact_1b94ff76"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_02c53ec5cb0e"}
}
```

Verdict: PASS
