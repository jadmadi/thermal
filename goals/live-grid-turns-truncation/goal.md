# Task: Fix turns text truncation and align gauge brackets in live monitor grid (live-grid-turns-truncation)

## Context
- Goal: Prevent `Turns: X turns (Y% hit)` metric from clipping with ellipsis on standard 80-column terminals and harmonize gauge bracket widths.
- Why: 30-character turn strings overflow 24-character grid cells (`6642 turns (86% h…`), and unaligned flame and sparkline brackets create visual raggedness.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: None (Independent, Seq #94)
- Readiness: Ready to claim

## Files to Touch
- internal/tui/live.go
- internal/tui/live_test.go

## Scope & File Details
- Edit: `internal/tui/live.go` in `formatTurnsAndHit` (lines 672-678, lines 332) and gauge bracket rendering (lines 693-705, 748-760)
- Create: `internal/tui/live_test.go` (responsive turn format and bracket tests)
- Read: `internal/tui/live.go:322-346` (3-box metric KPI grid layout)

## Requirements
1. In `internal/tui/live.go`, `formatTurnsAndHit` must accept available cell width and apply progressive abbreviation:
   Full: `6642 turns (86% hit)` (if text fits within available width)
   Compact: `6.6K (86% hit)` (if full overflows but compact fits)
   Minimal: `6642 turns` (if compact also overflows)
2. On 80-column terminal layouts (`colW = 24`), turns and cache hit must never clip with an ellipsis (`…`).
3. Gauge bracket widths in `formatFlameLine` and `formatSparkLine` must align consistently so opening and closing brackets do not show visual raggedness across terminal sizes.

## Interfaces / signatures (fixed)
```go
func formatTurnsAndHit(turns int, hit float64, availWidth int, colors bool) string
func formatFlameLine(intensity float64, model string, innerWidth int, colors bool) string
func formatSparkLine(buckets [60]int64, peak int64, innerWidth int, isActivity bool, colors bool) string
```

## Constraints and non-goals
- Non-goal: Do NOT edit `docs/MIGRATION.md` (pure TUI visual format, no CLI flags or `--json` schema changes).
- No new external dependencies.
- Monospace alignment and ANSI cell padding must be preserved.
- Do not alter turn count or cache hit rate calculations in `internal/thermal`.

## Example
- Input: `turns = 6642`, `hit = 86.0%`, `availWidth = 14`.
- Output: `6.6K (86% hit)` (14 chars) -> fits within `colW = 24` without `…`.
- Boundary failure prevented: `6642 turns (86% h…` is never produced on 80-column screens.

## Acceptance criteria
- AC-001: `go test -race -v -run TestLiveTUI_FormatTurnsAndHit_Responsive ./internal/tui`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that `colW = 24` formats without ellipsis truncation.
- AC-002: `go test -race -v -run TestLiveTUI_FormatTurnsAndHit_NoEllipsisOverflow ./internal/tui`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that string width never exceeds `colW - len("Turns:    ")` for any turn count from 1 to 10,000,000 (negative case).

## If blocked
Stop affected work; report evidence and the decision needed. Do not guess.

## Return
Return status, summary, changed files, case results, commands/exits, evidence, candidate, blocker.
