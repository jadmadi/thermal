# Plan: Make every surface fit the terminal width or refuse with a clear minimum-width message

## Execution Steps
- [x] **Phase 1 — Width resolver and geometry primitives**: Add `internal/render/width.go` with `ResolveWidth`, `DisplayWidth`, `TruncateToWidth`, and the layout table. Add `--width` and `Options.Width`. Acceptance: `TestWidthPrecedence`, `TestDisplayWidthUnicode`, and `TestWidthClamp` pass.
  - Precedence: `--width` > `COLUMNS` > stdout TTY > stdin TTY > 100.
  - Clamp/reject `[40,240]`.
  - Resolve once after parsing.
- [x] **Phase 2 — Static reports**: Thread width through the leaderboard, period reports, projects, models, mix, stats, trend, audit, changelog, help, and heatmap. Implement the frozen column-priority table, truncation, border recomputation, and refusal. Acceptance: `TestWidthMatrix`, `TestColumnDropOrder`, `TestLongNameTruncation`, and golden captures at 80/120 pass.
  - No line > width after ANSI stripping.
  - JSON unaffected.
  - Heatmap auto-fit with caption.
  - Help wrapping with hanging indent.
- [x] **Phase 3 — Live TUI breakpoints**: Add a breakpoint layout function: 80+ current, 60–79 stacked, <60 refusal; handle resize without a blank frame; keep `q Quit` visible. Acceptance: `TestLiveWidthRefusal`, `TestLiveStackedLayout`, `TestLiveResize`, and the existing 80x24/120x40 goldens pass.
  - Do not enter the alternate screen when refusing.
  - 50x20 refusal message exact.
  - 60x20 stacked frame has a complete border.
- [x] **Phase 4 — Gate, docs, and closure**: Add the simulated-user-gate width matrix, README/MIGRATION/project.jsonc, and close the goal. Record the matrix and golden diffs in `evidence.md`.
  - `CGO_ENABLED=0 go test -race ./...` and `./scripts/simulated_user_gate.sh` pass.
  - Commit with `fix(render): make every surface fit the terminal width (goals/cli-terminal-width-adaptation/goal.md)`.
  - Certify with `sila goals certify cli-terminal-width-adaptation --evidence=goals/cli-terminal-width-adaptation/evidence.md`.

## Acceptance Command Ledger
| AC | Command | Expected |
| --- | --- | --- |
| AC-001 | 80-col matrix | all lines ≤80 |
| AC-002 | 50-col matrix | fit or refusal, no clipping |
| AC-003 | 120-col | wider layout, ≤120 |
| AC-004 | `COLUMNS=60` vs `--width 60` | identical |
| AC-005 | `COLUMNS=60 --width 100` | 100 wins |
| AC-006 | `--width 39`/`241` | exit 2 |
| AC-007 | JSON at 40/160 | identical |
| AC-008 | 256-char project | ellipsis, border intact |
| AC-009 | 50-col heatmap | auto-fit, caption |
| AC-010 | live 60x20 / 50x20 | stack / refusal |
| AC-011 | 80x24/120x40 goldens | unchanged |
| AC-012 | `go test -race ...` | exit 0 |
| AC-013 | gate | exit 0 |

## Rollback
- The width parameter is additive; if a renderer regresses, pass `100` for that renderer and keep the resolver for the others.
- Do not revert the refusal behavior for `live` below 60; a clipped frame hides the quit key.
- Keep `--width` and `COLUMNS` support even if a layout is simplified; they are the test seam.

## Phase 5 — Attribution & Registry Closure
- [x] **Phase 5 — Attribution & Registry Closure**: (1) check all completed tasks in `goals/cli-terminal-width-adaptation/plan.md`; (2) commit with `(goals/cli-terminal-width-adaptation/goal.md)` using a `fix(render):` subject; (3) certify via `sila goals certify cli-terminal-width-adaptation --evidence=goals/cli-terminal-width-adaptation/evidence.md` (with `Verdict: PASS`); (4) reconcile via `sila goals` and commit the updated `goals/README.md` and certification receipts; verify `sila goals --json` reports `unverified_count: 0` for this goal.
