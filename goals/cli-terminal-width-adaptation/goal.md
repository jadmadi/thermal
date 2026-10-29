# Goal: Make every surface fit the terminal width or refuse with a clear minimum-width message

## Goal Description
Priority: P1 (terminal citizenship). Thermal renders fixed-width box art regardless of the terminal. At an 80-column terminal — the common SSH and split-pane baseline — every report except `license` overflows, and the live TUI clips below 80 columns with no right border and truncated controls. `COLUMNS` is ignored, and the one width helper (`reportWidth()`) is used only for charts.

Measured on the current source build (2026-10-29) with `COLUMNS=60` and with a real 50-column PTY:

| Surface | Max line width |
| --- | --- |
| `thermal` leaderboard | 85 |
| `thermal --help` | 94 |
| `thermal daily` / `weekly` / `monthly` | 104 |
| `thermal projects` | 105 |
| `thermal models` | 106 |
| `thermal mix` | 101 |
| `thermal stats` | 113 |
| `thermal trend` | 119 |
| `thermal audit` | 98 |
| `thermal changelog` | 98 |
| `thermal license` | 77 |
| `thermal opencode` heatmap | 100 |

At a real 50-column PTY, the leaderboard still emits 85-character lines. In the live TUI, a 50x20 pane loses the right border and cuts `[q] Quit`; at 60, 61, and 70 columns the header and table cells are clipped mid-word (`LIVE (1.0s`, `TOOLTOK`, `S` for `STEPS`). The TUI is golden-tested only at 80x24 and 120x40, so the failure begins one column below the tested range.

The outcome of this goal is a **width contract**: every surface detects the available width, adapts by a documented column-priority order, and either fits or refuses with a one-line minimum-width message. No surface may emit a line wider than the terminal, wrap box art, or hide a control key.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 249
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3 (after the parser/help/statusline goals settle output shapes)
- **Related implemented goals (do not regress)**: `output-formatting-primitives` owns symbol/format semantics; `chart-scale-order-and-top-contract` owns chart scaling and row counts; `help-supported-tools-and-flags` owns help content. This goal owns only geometry and width-driven degradation.

## Files to Touch
- `internal/render/width.go`
- `internal/render/card.go`
- `internal/render/report.go`
- `internal/render/leaderboard.go`
- `internal/render/projects.go`
- `internal/render/models.go`
- `internal/render/analytics.go`
- `internal/render/audit.go`
- `internal/render/chart.go`
- `internal/render/help.go`
- `internal/tui/live.go`
- `internal/tui/*`
- `cmd/thermal/main.go`
- `cmd/thermal/width_test.go`
- `internal/render/*_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-terminal-width-adaptation/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-terminal-width-adaptation/plan.md`](plan.md)

## Requirements

1. **One width resolver.**
   - Resolution order: explicit `--width <N>`; `COLUMNS` if set to a positive integer; TTY ioctl on stdout; TTY ioctl on stdin; default `100`.
   - Clamp to `[40, 240]`; values outside the clamp are rejected with a usage error (exit `2`) rather than silently clamped.
   - Resolve once per invocation and pass the width into every renderer; no renderer may call `term.GetSize` directly.
   - `--width` affects text rendering only; `--json` ignores it.
   - Width must be resolved after flag parsing and before any rendering, including help.
   - `COLUMNS` must be honored even when stdout is a pipe (for reproducible captures).

2. **Static report contract.**
   - Every rendered line must satisfy `len(line) <= width` (after ANSI stripping). No wrapping by the terminal.
   - Each report declares a minimum width and a column drop order. Drop columns from lowest to highest priority:
     | Report | Keep first | Then | Minimum |
     | --- | --- | --- | --- |
     | Leaderboard | #, Tool, Tokens | Cost, Strk, Best, Days | 40 (compact list) |
     | Daily/weekly/monthly | Period, Tokens, Cost | Models, Input, Output, Reason, Cache | 56 |
     | Projects | Project, Tokens, Cost | P-Days, Tools, Last | 48 |
     | Models | Model, Tokens, Cost | M-Days, Tools, Last | 48 |
     | Stats | label + value | bars shrink by 2 cells per 10 columns | 40 |
     | Trend | Range, Mean, Slope | sparkline cells, then the card border | 48 |
     | Audit | STATUS, path, lines | path truncation from the left, advice wrap | 56 |
     | Changelog | version, date, bullets | bullet wrap | 48 |
     | Help | flag + description | description wrap, no box | 48 |
     | Heatmap | weeks | auto-reduce weeks to fit, minimum 4 | 40 |
   - When the width is below the minimum, print exactly:
     ```
     thermal: daily needs at least 56 columns (terminal is 50)
       widen the terminal, pipe to a file, or set --width 100
     ```
     Exit `2`; stdout empty; stderr only.
   - At or above the minimum, never print the refusal; always render a complete table.
   - Truncate long project/model names from the right with `…`; never truncate the numeric columns.
   - Box borders must be recomputed, never clipped; a card must end with its bottom-right corner.
   - `--no-color` output must satisfy the same width invariants.

3. **Live TUI contract.**
   - Preserve the existing 80x24 and 120x40 golden frames byte-for-byte (except for intentional label changes).
   - 80–240 columns: current layout, width-aware cards.
   - 60–79 columns: a stacked layout with one KPI column, shortened labels, no sparkline/flame bar, and the controls line wrapped or abbreviated so `[q] Quit` remains visible. No clipping, no missing right border.
   - 40–59 columns: print a single message instead of a frame:
     `Thermal Live needs at least 60 columns (terminal is 50); resize or use thermal live --json`
     Exit `2`; do not enter the alternate screen.
   - Below 40: same refusal with the exact minimum.
   - The controls line must always include `q Quit` and must never be the first line clipped.
   - A resize (`tea.WindowSizeMsg`) must re-layout without a blank frame; test 120→60→120.
   - Height: below 20 rows, show the refusal or a compact 12-row layout; never clip the controls.

4. **Heatmap contract.**
   - The heatmap's natural width is 100 columns at 52 weeks. At narrower widths, reduce the number of weeks automatically (`weeks = clamp((width - fixed - 1) / 2, 4, opts.Weeks)`) and add a caption `weeks 4–52, auto-fit to 50 columns`.
   - Never wrap the heatmap rows; never print a partial month label row.
   - `--weeks` remains the user's maximum; auto-fit may only reduce it, never increase it.
   - At widths below 40, print the minimum-width refusal.

5. **Help and changelog.**
   - `--help` must fit 80 columns by default: wrap flag descriptions with a hanging indent, no box wider than the width, and a one-column gutter.
   - Changelog bullets wrap with a two-space continuation indent; version cards shrink; no text is clipped.
   - `--width 120` may use the wider layout; `--width 60` must remain readable.
   - The global help must state the minimum supported width and that `--width`/`COLUMNS` are honored.

6. **Tests and golden files.**
   - A table test for every command × widths `{40, 50, 56, 60, 80, 100, 120, 160}` asserting `maxLine <= width` and the expected refusal below the minimum.
   - Golden files for the full text at 80 and 120 for the leaderboard, daily, projects, models, stats, trend, audit, help, changelog, and the opencode heatmap.
   - PTY tests for `live` at 50x20, 60x20, 80x24, 100x30, 120x40, 160x40; assert no clipped border and `q Quit` present.
   - `COLUMNS` and `--width` precedence tests.
   - A property test: stripping ANSI and measuring every line always respects the width.
   - A fuzz test over long project/model names (256 chars) and wide unicode (CJK) to ensure width is measured in display cells, not bytes.
   - Verify that a 256-char project name is truncated with `…` and does not break the border.

7. **Docs and migration.**
   - README documents the width contract, minimums, `--width`, and `COLUMNS`.
   - MIGRATION six-part entry: reports may drop columns below 80; `live` refuses below 60; detection and fix commands; escape hatch (`--width 100` or `--json`).
   - `project.jsonc` updated.
   - `scripts/simulated_user_gate.sh`: width matrix for 50/80/120 and the live refusal.

## Non-Goals
- Changing aggregation, sorting, totals, or token math.
- Changing chart scaling semantics (owned by `chart-scale-order-and-top-contract`).
- Adding mouse support or a scrollable TUI.
- Adding a responsive web layout.
- Changing color palettes or glyph choices.

## Acceptance Criteria
- AC-001: For each command in `thermal`, `daily`, `weekly`, `monthly`, `projects`, `models`, `stats`, `mix`, `trend`, `audit`, `changelog`, `--help`, and `opencode`, at `COLUMNS=80`, every line is ≤80 columns after ANSI stripping.
- AC-002: At `COLUMNS=50`, each command either renders ≤50 columns or exits `2` with the minimum-width message; no command emits a wrapped or clipped line.
- AC-003: At `COLUMNS=120`, the leaderboard and daily report render a wider layout and still satisfy `maxLine <= 120`.
- AC-004: `thermal --width 60 daily` and `COLUMNS=60 thermal daily` produce identical output.
- AC-005: `COLUMNS=60 thermal --width 100 daily` uses 100 (explicit flag wins).
- AC-006: `thermal --width 39 daily` and `--width 241` exit `2` with the range message.
- AC-007: `thermal --json daily` is byte-identical under `COLUMNS=40` and `COLUMNS=160` (JSON unaffected).
- AC-008: A 256-character project name is truncated with `…`; the table border remains complete and `maxLine <= width`.
- AC-009: The opencode heatmap at 50 columns auto-fits to 4–12 weeks, prints the auto-fit caption, and has no line >50.
- AC-010: In a 60x20 PTY, `thermal live` renders a stacked layout, includes `[q] Quit`, and has no clipped border; in a 50x20 PTY it prints the minimum-width message and does not enter the alternate screen.
- AC-011: The 80x24 and 120x40 live golden frames are unchanged (or the intentional diff is documented in evidence).
- AC-012: `go test -race ./internal/render/... ./internal/tui/... ./cmd/thermal/...` exits `0`, including `TestWidthMatrix`, `TestColumnDropOrder`, `TestLongNameTruncation`, `TestLiveWidthRefusal`, and `TestWidthPrecedence`.
- AC-013: `./scripts/simulated_user_gate.sh` exits `0` and includes the width matrix.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| `COLUMNS=40` leaderboard | compact list ≤40 or refusal below 40 |
| `COLUMNS=50` daily | refusal (min 56) or ≤50 |
| `COLUMNS=60` projects | drop P-Days/Tools, ≤60 |
| `COLUMNS=60` stats | shrink bars, ≤60 |
| `COLUMNS=60` help | wrapped descriptions, ≤60 |
| `COLUMNS=60` heatmap | auto-fit weeks, caption, ≤60 |
| `COLUMNS=50` live | refusal, no alternate screen |
| `COLUMNS=60` live | stacked layout, `q Quit` visible |
| `--width 39` | exit 2, range message |
| `--width abc` | exit 2, parser message |
| `--width` with `--json` | ignored, JSON unchanged |
| Long project name | truncated with `…`, border intact |
| CJK project name | display-cell width, not byte length |
| Resize 120→60 | re-layout, no blank frame |

## Metrics & Regression Guards
- Width invariant: 100% of golden lines satisfy `displayWidth(line) <= width`.
- Refusal exactness: below-minimum messages name the command, the minimum, and the actual width.
- Column drop order is frozen in a table test; changing it requires updating the doc.
- Golden diff: 80x24 and 120x40 live frames must match the existing snapshots.
- Gate: simulated user gate asserts AC-001, AC-002, AC-003, AC-004, AC-009, and AC-010.

## Risks & Mitigations
- **Risk:** Dropping columns changes the information users rely on. **Mitigation:** documented priority order, keep Tokens/Cost/Period longest, and prefer truncation over dropping for names.
- **Risk:** Width measured in bytes breaks unicode alignment. **Mitigation:** use a display-width function (`runewidth` logic implemented in-repo, no new dependency) and test CJK/emoji.
- **Risk:** `COLUMNS` may be exported by a parent shell with a stale value. **Mitigation:** explicit `--width` wins; document the precedence.
- **Risk:** A stacked TUI layout doubles the rendering code. **Mitigation:** one layout function with breakpoints, golden tests at each breakpoint.
- **Risk:** The refusal exit `2` breaks scripts that pipe live. **Mitigation:** `--json` remains the documented script path; MIGRATION entry.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "terminal width fixed box art"` and `sila recall "live tui clipping narrow"`.
- **Writes**: record a `case` for the "single width resolver + column priority contract" mechanism, grounded to `internal/render/width.go`; record a `lesson` for byte-vs-cell width bugs if encountered.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "width contract shipped" --open="unicode width edge cases"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-terminal-width-adaptation/goal.md)` under a `fix(render):` or `feat(render):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the width matrix and golden diffs.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
