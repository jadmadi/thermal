# Evidence: Enforce one width, colour, and sanitisation policy

Date: 2026-10-30
Binary: `/tmp/thermal-g4` (source build)
Goal: `cli-terminal-output-integrity`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | `COLUMNS=40 thermal daily` on a pipe | exit 0; 100-column fallback; no line >100 |
| AC-002 | PTY `COLUMNS=40 thermal daily` | exit 2; `terminal is 40` refusal |
| AC-003 | `--width 40 thermal daily` | exit 2; `(--width 40)` refusal |
| AC-004 | `COLUMNS=60` leaderboard | legend pairs stay atomic |
| AC-005 | `sources` at 80 | head+tail paths, `file`/`dir` detail |
| AC-006 | `sources --width 40` | one line per source, no line >40 |
| AC-007 | path containing `\n` | title renders `\n` literally; frame stays closed |
| AC-008 | `NO_COLOR=1` / `--plain` statusline | zero ANSI |
| AC-009 | `--no-color` vs pipe | no ANSI on either |
| AC-010 | `TERM=dumb` statusline | zero ANSI |
| AC-011 | `--json --width 40` vs `--width 200` | identical apart from `generatedAt` |
| AC-012 | `go test ./internal/render/... ./cmd/thermal/...` | ok |
| AC-013 | gallery + simulated user gates | 167 passed, 0 failed |

## Observed Output

```text
$ COLUMNS=40 thermal daily | cat
exit 0
stdout 1842 bytes; stderr 0; lines >100: 0
```

```text
$ thermal --width 40 daily
exit 2
thermal: daily needs at least 56 columns (--width 40)
  widen the terminal, pipe to a file, or set --width 100
```

```text
$ thermal sources --width 40 --no-color
  MiMoCode   ~/.local…code.db missing
  thermal sources --json
```

```text
$ COLUMNS=60 thermal --no-color
  Legend: Strk = current streak · Best = longest streak
          Days = active days
          Cost: $ recorded · ~$ estimated
                ≥$ partial known · — unpriced
```

```text
$ NO_COLOR=1 thermal statusline | grep -c ESC -> 0
$ thermal statusline --plain | grep -c ESC     -> 0
$ TERM=dumb thermal statusline | grep -c ESC   -> 0
```

```text
$ thermal claude --db "<dir with a newline>"
  ╭─ Claude Activity ─────────────── /tmp/.../new\nline/projects ─╮
```

## Implementation Notes

- `requireCommandWidth` uses the 100-column pipe fallback when stdout is not a
  TTY and `--width` is absent; explicit `--width` still refuses and now names
  the flag.
- `RenderSources` elides the middle of paths, shortens `expected file` to
  `file` and `expected directory` to `dir`, and uses a one-line layout below
  60 columns.
- `wrapLegend` wraps between marker/gloss items; a cost marker is never
  separated from its meaning.
- `SanitizeText` escapes newline, carriage return, tab, ESC, and other
  control characters; `RenderCard` uses the sanitized title and right header.
- `RenderStatusline` shares the colour gate: `--no-color`, `--plain`,
  `NO_COLOR`, and `TERM=dumb` all disable ANSI.
- Gate width checks now use explicit `--width`, matching the new pipe policy.

## Verification Commands

```bash
go test ./internal/render/... ./cmd/thermal/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `167 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-terminal-output-integrity",
  "prime": {"ran": true},
  "recall": [
    {"query": "COLUMNS non-tty width", "matches": 5},
    {"query": "statusline NO_COLOR plain", "matches": 5}
  ],
  "memos": [{"id": "fact_c10e92b7"}],
  "lessons": [{"id": "fact_946aeb39"}],
  "cases": [{"id": "fact_e8329730"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_cdebaebe0af5"}
}
```

Verdict: PASS
