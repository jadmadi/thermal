# Plan: Enforce one width, colour, and sanitisation policy across every renderer

## Execution Steps
- [x] **Phase 1 — Width resolver**: Add `ResolveWidth`/`WidthSource`, ignore
  `COLUMNS` on a non-TTY pipe, thread the source into `WidthRefusal`, and keep
  `--width` authoritative. Acceptance: `TestResolveWidthMatrix`,
  `TestNonTTYIgnoresColumns`, and `TestWidthRefusalSource` pass.
  - Pipe + `COLUMNS=40` -> 100, no refusal.
  - TTY + `COLUMNS=40` -> refusal naming the terminal.
  - `--width 40` -> refusal naming the flag.
  - JSON ignores width.
- [x] **Phase 2 — Legend pairs**: Introduce `LegendItem` and `WrapLegend`; move
  the streak and cost legends onto it. Acceptance: `TestLegendAtomic` at
  40/60/80/120; no dangling marker.
  - Continuation indent matches the legend column.
  - Exact marker/gloss strings are preserved.
- [x] **Phase 3 — Sources table and path fitting**: Add `FitPath`, responsive
  column drop (`Detail`, then `Alias`), short detail words, and a 40-column
  one-line form. Acceptance: `TestSourcesFit40/60/80/120` and
  `TestSourcesKeepsHead` pass.
  - `~/.claude/projects/…`, never `…ed/threads/threads.db`.
  - `State` is never truncated.
- [x] **Phase 4 — Sanitisation and colour policy**: Add `SanitizeText` and
  `ColorEnabled`; apply to card titles, tables, charts, audit paths, and the
  statusline; make `--plain` imply no colour. Acceptance:
  `TestSanitizeText`, `TestFrameWithNewlinePath`, `TestColorMatrix`, and
  `TestStatuslineNoColor` pass.
  - JSON keeps raw values.
  - No ANSI under `NO_COLOR`, `--no-color`, `--plain`, `TERM=dumb`, or a pipe.
- [x] **Phase 5 — Gallery, docs, gate, closure**: Update the output gallery,
  README width section, `project.jsonc`, and both gates; record evidence;
  commit, certify, reconcile.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | `COLUMNS=40 thermal daily` (pipe) | exit 0, no refusal |
| AC-002 | `COLUMNS=40 thermal daily` (PTY) | exit 2, terminal source |
| AC-003 | `--width 40 thermal daily` | exit 2, flag source |
| AC-004 | `COLUMNS=60` legend | complete pairs |
| AC-005 | `COLUMNS=80 sources` | head+tail paths, short detail |
| AC-006 | `COLUMNS=40 sources` | one line per source, <=40 |
| AC-007 | newline path | one line, closed frame |
| AC-008 | `NO_COLOR` / `--plain` statusline | zero ANSI |
| AC-009 | `--no-color` vs pipe | byte-identical |
| AC-010 | `TERM=dumb` (TTY) | zero ANSI |
| AC-011 | JSON width 40 vs 200 | identical except timestamp |
| AC-012 | `go test -race ...` | exit 0 |
| AC-013 | gallery + user gates | exit 0 |

## Rollback
- Keep `--width` authoritative even if the `COLUMNS` policy is reverted; the
  explicit flag must always work.
- If sanitisation changes copied paths, limit it to control characters only;
  never escape `/`, `~`, or spaces.
- If the responsive sources table is too complex, keep the 80-column table and
  fall back to one line per source below 60.
- Do not restore `colors := !opts.NoColor`; `NO_COLOR` is a hard requirement.

## Phase 6 — Attribution & Registry Closure
- [x] **Phase 6 — Attribution & Registry Closure**: (1) check every task
  above; (2) commit with
  `fix(cli): unify width, colour, and sanitisation across renderers
  (goals/cli-terminal-output-integrity/goal.md)`; (3) certify via
  `sila goals certify cli-terminal-output-integrity
  --evidence=goals/cli-terminal-output-integrity/evidence.md`; (4) run
  `sila goals` and commit the updated `goals/README.md` and certification
  receipts; verify `sila goals --json` reports `unverified_count: 0`.
