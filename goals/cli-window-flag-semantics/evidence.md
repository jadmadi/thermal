# Evidence: Make window and repeated-flag semantics explicit

Date: 2026-10-30
Binary: `/tmp/thermal-g3` (source build)
Goal: `cli-window-flag-semantics`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | `weekly --last 7` | exit 0; `last 7 weeks` |
| AC-002 | `weekly --last 14d` | exit 0; `last 14 days`; same totals as the 14-day window |
| AC-003 | `monthly --last 3m` vs `monthly --last 3` | both `last 3 months`; identical totals |
| AC-004 | `daily --last` | exit 2; unit-aware message and example |
| AC-005 | `weekly --last 7x` | exit 2; `d, w, or m` in the message |
| AC-006 | `--tool claude --tool codewhale` | exit 0; one stderr warning naming both values |
| AC-007 | `--json --json` | exit 0; no warning |
| AC-008 | `weekly -- --offline` | exit 2; `was passed after --` explanation, no tool list |
| AC-009 | usage `example:` lines | shell-safe; no `<path>` in `flagMetas` |
| AC-010 | `--db` metadata example | uses `$HOME/.local/share/opencode/opencode.db` |
| AC-011 | `go test ./cmd/thermal/... ./internal/thermal/...` | ok |
| AC-012 | `./scripts/simulated_user_gate.sh` | 164 passed, 0 failed |

## Observed Output

```text
$ thermal weekly --last 7
Window: [last 7 weeks · all tools]

$ thermal weekly --last 14d
Window: [last 14 days · all tools]

$ thermal monthly --last 3m
Window: [last 3 months · all tools]

$ thermal daily --last 7w
Window: [last 7 weeks · all tools]
```

```text
$ thermal weekly --last 7x
exit 2
thermal: --last wants a whole number, optionally followed by d, w, or m, got "7x"
  example: thermal weekly --last 4
```

```text
$ thermal --tool claude --tool codewhale
exit 0; stderr:
thermal: warning: --tool given 2 times; using "codewhale" (last), ignoring "claude"
```

```text
$ thermal weekly -- --offline
exit 2
thermal: "--offline" was passed after --, so it is a positional, not a flag
  put flags before --, or drop -- to let thermal parse them
```

## Implementation Notes

- `--last` now accepts a bare count or a `d`/`w`/`m` suffix.
- A bare count keeps the report's grain; a suffix resolves to an explicit
  `Since` date and a `WindowLabel`, so `weekly --last 14d` is 14 days.
- `--last` with a suffix cannot combine with `--since`/`--until`.
- Repeated singleton flags are detected before dispatch and warn on stderr
  with both the discarded and kept values. Boolean repeats stay silent.
- `--` now explains flag-like positional tokens instead of reporting an
  unknown tool.
- The `--db` usage example uses a real path; flag signatures keep `<path>`.

## Verification Commands

```bash
go test ./cmd/thermal/... ./internal/thermal/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `164 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-window-flag-semantics",
  "prime": {"ran": true},
  "recall": [
    {"query": "--last weeks months", "matches": 5},
    {"query": "repeated flag last wins", "matches": 5}
  ],
  "memos": [{"id": "fact_850c1150"}],
  "lessons": [{"id": "fact_f3ef06aa"}],
  "cases": [{"id": "fact_9580a810"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_00100e8050ca"}
}
```

Verdict: PASS
