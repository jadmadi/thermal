# Goal: Enforce one width, colour, and sanitisation policy across every renderer

## Goal Description
Priority: P1 (terminal trust). Thermal has several good rendering rules, but
they are not applied uniformly. A non-TTY pipeline that exports `COLUMNS=40`
gets a refusal that tells it to pipe to a file; the cost legend breaks between
a marker and its meaning; the sources table truncates the one part of a path a
user needs; the statusline ignores `NO_COLOR` and still emits truecolor under
`--plain`; and a newline in a path tears the report frame.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `COLUMNS=40 thermal daily` -> exit `2`, stdout 0 B:
  `thermal: daily needs at least 56 columns (terminal is 40)`
  `  widen the terminal, pipe to a file, or set --width 100`
- `COLUMNS=60 thermal` -> the legend wraps as:
  `Legend: Strk = current streak · Best = longest streak ·`
  `Days = active days`
  `Cost: $ recorded · ~$ estimated · ≥$ partial known · —`
  `unpriced · ≈ approximate · $0.00 recorded zero`
- `COLUMNS=80 thermal sources` -> rows read
  `…/opencode/opencode.db`, `expected dire…`, `…ed/threads/threads.db`.
- `thermal claude --db "<dir with a newline>"` -> the top border wraps:
  `╭─ Claude Activity ─── <path>/new`
  `line/projects ─╮`
- `NO_COLOR=1 thermal statusline` -> 10 ANSI sequences, byte-identical to the
  default.
- `thermal statusline --plain` -> 10 ANSI sequences; only the emoji is removed.
- `thermal statusline --no-color` -> 0 ANSI sequences.

The outcome of this goal is one **terminal output integrity contract**:

1. Width resolves from `--width`, then the real terminal when stdout is a TTY,
   then `COLUMNS` only when stdout is a TTY, then the 100-column default.
2. Legends wrap between marker/gloss pairs, never inside one.
3. The sources table fits the width and elides the middle of a path, keeping
   the home prefix and the meaningful tail.
4. Every untrusted string is control-character sanitised before rendering.
5. One colour policy covers reports, legends, warnings, and the statusline;
   `--plain` implies no colour.
6. JSON is untouched by width and colour.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-terminal-width-adaptation, cli-leaderboard-ranking-and-legend, cli-nontty-and-statusline-output-contract
- **Sequence**: 258
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3
Related goals:
- `cli-empty-data-success-contract` owns when the sources table is printed;
  this goal owns how it renders.
- `statusline-and-share-scope-contract` owns the statusline content labels;
  this goal owns its colour policy.
- `output-gallery-regression-contract` owns the golden gallery; this goal adds
  the width/colour/sanitisation cases to it.

## Files to Touch
- `internal/render/wrap.go`
- `internal/render/wrap_test.go`
- `internal/render/leaderboard.go`
- `internal/render/leaderboard_test.go`
- `internal/render/sources.go`
- `internal/render/sources_test.go`
- `internal/render/statusline.go`
- `internal/render/card.go`
- `internal/render/report.go`
- `internal/render/analytics.go`
- `internal/render/models.go`
- `internal/render/projects.go`
- `cmd/thermal/main.go`
- `cmd/thermal/statusline_test.go`
- `cmd/thermal/width_test.go`
- `scripts/simulated_user_gate.sh`
- `scripts/output_gallery_gate.sh`
- `docs/OUTPUT-GALLERY.md`
- `README.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-terminal-output-integrity/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-terminal-output-integrity/plan.md`](plan.md)
- Prior contracts:
  [`goals/cli-terminal-width-adaptation/goal.md`](../cli-terminal-width-adaptation/goal.md),
  [`goals/cli-leaderboard-ranking-and-legend/goal.md`](../cli-leaderboard-ranking-and-legend/goal.md),
  [`goals/cli-nontty-and-statusline-output-contract/goal.md`](../cli-nontty-and-statusline-output-contract/goal.md)

## Requirements

1. **Width resolution.**
   - Precedence: `--width` (40-240) > terminal size when stdout is a TTY >
     `COLUMNS` when stdout is a TTY > 100.
   - When stdout is **not** a TTY and `--width` is absent, ignore `COLUMNS`.
     Use 100. This is the fix for the CI refusal.
   - `--width 40` on a non-TTY still refuses below the command minimum,
     because the user explicitly asked for 40. The message must say
     `--width 40` and the required minimum.
   - `COLUMNS=40` on a TTY still refuses with the existing two-line message.
   - `--width` wins over `COLUMNS` and over the real terminal size.
   - Width resolves once per invocation; no renderer calls `os.Getenv`
     directly.
   - JSON ignores width entirely.

2. **Legend integrity.**
   - Represent the legend as ordered pairs:
     ```go
     type LegendItem struct {
         Marker string // "Strk", "Best", "$", "~$", "≥$", "—", "≈"
         Gloss  string // "current streak", "recorded", ...
     }
     ```
   - Wrap between items, never inside an item. Continuation lines indent to
     the legend's first column.
   - The cost legend must keep each marker adjacent to its gloss:
     `$ recorded`, `~$ estimated`, `≥$ partial known`, `— unpriced`,
     `≈ approximate`, `$0.00 recorded zero`.
   - At 60 columns, the output must be two or more complete pairs per line,
     not a dangling marker.
   - No colour is required to understand a marker.

3. **Sources table.**
   - Columns: `Tool`, `Alias`, `Path`, `State`, `Detail`.
   - Fit the width by dropping `Detail`, then `Alias`; never truncate `State`.
   - At 80 columns, use a shorter detail vocabulary: `file`/`dir` instead of
     `expected file`/`expected directory`.
   - Path elision keeps the head and tail:
     `~/.claude/projects/…` (never `…ed/threads/threads.db`).
   - If the path fits, print it whole. If not, elide the middle, not the start
     or end.
   - At 40 columns, print one line per source:
     `Zed  ~/.zed/threads/threads.db  missing`.
   - The table must not exceed the resolved width at 40, 60, 80, or 120.
   - `sources --json` keeps full absolute paths and is width-independent.

4. **Control-character sanitisation.**
   - Add `SanitizeText(s string) string`:
     - `\n` -> `\n` literal (`\\n`), `\r` -> `\\r`, `\t` -> `\\t`.
     - ESC (0x1b) -> `\\x1b`; other C0/C1 controls -> `\\xNN`.
     - Valid UTF-8 is preserved; invalid bytes -> `�`.
   - Apply to every string that originates outside the renderer: tool names,
     display paths, project names, model names, card titles, chart labels,
     audit paths, and statusline scope.
   - Never apply it to ANSI sequences the renderer itself emits.
   - A path containing a newline must render on one line; the frame must stay
     closed.
   - JSON: keep the original string or the escaped form? Recommendation:
     JSON keeps the original value; only terminal renderers sanitise.
   - Add a test with `new\nline`, `tab\there`, `esc\x1b[31m`, and an invalid
     UTF-8 byte.

5. **Colour policy.**
   - One helper:
     ```go
     type ColorOptions struct {
         NoColor bool
         Plain   bool
     }

     func ColorEnabled(opts ColorOptions, w io.Writer) bool
     ```
     false when `--no-color`, `NO_COLOR` is non-empty, `TERM=dumb`, or `w` is
     not a TTY.
   - `--plain` implies no colour and no emoji/glyph, on every surface.
   - Apply to the statusline, the `prompt`/`status` aliases, leaderboard,
     reports, analytics, charts, audit, warnings, and help.
   - `CLICOLOR_FORCE=1` may force colour only when the stream is a TTY;
     document the exception. A pipe stays clean.
   - `--no-color` is byte-identical to a non-TTY pipe with no colour.
   - Colour is never the sole encoding; the marker/word stays.

6. **Refusal messages.**
   - `WidthRefusal(command, min, width)` must state the source:
     `thermal: daily needs at least 56 columns (--width 40)` or
     `(COLUMNS=40)` or `(terminal is 40)`.
   - When stdout is not a TTY, do not say "pipe to a file"; say
     `stdout is not a terminal: omit COLUMNS or set --width 100`.
   - The refusal is two lines, exit `2`, stdout empty.

7. **Tests and gallery.**
   - Width matrix: 40, 60, 80, 120 for `thermal`, `daily`, `weekly`,
     `projects`, `models`, `stats`, `trend`, `audit`, `sources`, `help`,
     `statusline`.
   - Assert every line's display width <= the resolved width, using
     `ansi.StringWidth`.
   - Legend pair test: no line ends with a marker and no line starts with a
     gloss that belongs to the previous line.
   - Sources path test: head and tail present; `expected dire…` absent.
   - Control-character test: one line per record; frame closes.
   - Colour matrix: default, `NO_COLOR`, `--no-color`, `--plain`,
     `TERM=dumb`, `CLICOLOR_FORCE=1` × TTY/pipe.
   - Update `docs/OUTPUT-GALLERY.md` and run
     `./scripts/output_gallery_gate.sh`.

## Non-Goals
- Changing report content, totals, or column meanings.
- Changing the statusline's content or scope labels.
- Adding a pager or a new TUI.
- Changing JSON schemas or paths.
- Making `COLUMNS` an error on a TTY.
- Supporting arbitrary control characters as visible glyphs.

## Acceptance Criteria
- AC-001: `COLUMNS=40 thermal daily` on a pipe exits `0` and produces no line
  wider than 100 (the non-TTY default), or exits `0` with a table that fits
  56+ columns; it must not refuse because of `COLUMNS`.
- AC-002: `COLUMNS=40 thermal daily` on a PTY exits `2` with `(terminal is 40)`
  or `(COLUMNS=40)` and the `--width 100` fix.
- AC-003: `--width 40 thermal daily` exits `2` naming `--width 40`.
- AC-004: At `COLUMNS=60`, no legend line ends with `· —` or starts with
  `unpriced`; every marker is adjacent to its gloss.
- AC-005: At `COLUMNS=80`, `thermal sources` paths keep the home prefix and
  the tail; no cell reads `expected dire…`.
- AC-006: At `COLUMNS=40`, `thermal sources` prints one line per source and no
  line exceeds 40 columns.
- AC-007: A fixture path containing `\n` renders the card on one line and the
  frame closes.
- AC-008: `NO_COLOR=1 thermal statusline`, `thermal prompt`, and `--plain`
  all emit zero ANSI.
- AC-009: `--no-color` is byte-identical to a piped default for every report.
- AC-010: `TERM=dumb` emits zero ANSI on a TTY.
- AC-011: `thermal --json --width 40` is byte-identical to
  `thermal --json --width 200`.
- AC-012: `go test -race ./internal/render/... ./cmd/thermal/...` exits `0`,
  including `TestWidthMatrix`, `TestLegendAtomic`, `TestSourcesFit`,
  `TestSanitizeText`, and `TestColorMatrix`.
- AC-013: `./scripts/output_gallery_gate.sh` and
  `./scripts/simulated_user_gate.sh` exit `0`.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| `COLUMNS=40`, pipe, no `--width` | ignore COLUMNS, use 100 |
| `COLUMNS=40`, TTY | refusal names the terminal |
| `--width 40`, any stream | refusal names `--width 40` |
| `COLUMNS=60`, legend | complete pairs, indent continuation |
| `COLUMNS=40`, sources | one line per source |
| `COLUMNS=80`, sources | elide middle, detail `file`/`dir` |
| newline in path | one line, closed frame |
| ESC in path | visible `\\x1b`, no live sequence |
| `NO_COLOR=1` | zero ANSI |
| `--plain` | zero ANSI, no glyphs |
| `TERM=dumb` | zero ANSI |
| `CLICOLOR_FORCE=1`, pipe | zero ANSI |
| JSON + width | byte-identical |

## Metrics & Regression Guards
- Width: 100% of lines in the width matrix <= resolved width.
- Legend: zero dangling markers across 40/60/80/120.
- Sources: zero truncated tails at 80; zero lines >40 at 40.
- Sanitisation: one line per record for `\n`, `\r`, `\t`, ESC, and invalid
  UTF-8.
- Colour: zero ANSI under every disable condition; `--no-color` equals pipe.
- Gate: gallery and simulated-user gates assert AC-001, AC-004, AC-005,
  AC-008, and AC-012.

## Risks & Mitigations
- **Risk:** Ignoring `COLUMNS` on a pipe changes CI snapshots.
  **Mitigation:** they should use `--width`; document the precedence in
  MIGRATION and README.
- **Risk:** Sanitising paths changes copied output.
  **Mitigation:** only control characters are escaped; normal paths are
  unchanged. JSON keeps the original.
- **Risk:** `--plain` losing colour breaks a themed prompt.
  **Mitigation:** `--plain` is documented as no glyphs and no ANSI; users who
  want colour keep the default.
- **Risk:** A responsive sources table becomes complex.
  **Mitigation:** a simple drop order (Detail, Alias) plus middle elision; test
  four widths, not every width.
- **Risk:** The legend pair model changes wording.
  **Mitigation:** keep the exact strings and only change the wrap unit.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "COLUMNS non-tty width"` and
  `sila recall "statusline NO_COLOR plain"`.
- **Writes**: record a `case` for the "single width/colour/sanitisation
  policy" mechanism; record a `lesson` for the COLUMNS precedence bug.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "terminal output integrity shipped"
  --open="light-theme visual check"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-terminal-output-integrity/goal.md)` under a `fix(cli):` or
   `feat(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the width, legend, sources, sanitisation, and colour matrices.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
