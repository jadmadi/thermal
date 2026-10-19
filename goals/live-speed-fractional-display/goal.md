# Task: Preserve fractional sub-1.0 speeds in live monitor (live-speed-fractional-display)

## Context
- Goal: Display fractional token and step speeds for positive sub-1.0 rates in `internal/tui/live.go` `formatSpeed`.
- Why: Truncating float speeds to `int64` turns rates like 40 step/min (0.67 step/s) into `0 step/s`, contradicting the rate metric and active flame state.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: None (Independent, Seq #93)
- Readiness: Ready to claim

## Files to Touch
- internal/tui/live.go
- internal/tui/live_test.go

## Scope & File Details
- Edit: `internal/tui/live.go` in `formatSpeed` (lines 586-603)
- Create: `internal/tui/live_test.go` (speed formatting unit tests)
- Read: `internal/tui/live.go:586-603` (existing speed formatting)

## Requirements
1. In `internal/tui/live.go:586-603`, when `turnSpeed` or `tokSpeed` is greater than 0 and less than 1.0, format with one decimal place (`fmt.Sprintf("%.1f step/s", turnSpeed)` or `tok/s`).
2. Exact zero speeds must continue to format cleanly as `0 step/s` or `0 tok/s` in muted text.
3. Rates between 1.0 and 9.9 with fractional components may retain one decimal place if column width permits (`colW >= 24`).

## Interfaces / signatures (fixed)
```go
func formatSpeed(tokSpeed, turnSpeed float64, isActivity bool, colW int, colors bool) string
```

## Constraints and non-goals
- Non-goal: Do NOT edit `docs/MIGRATION.md` (pure TUI visual format, no CLI flags or `--json` schema changes).
- No new external dependencies.
- Must honor `--no-color` mode and maintain visual cell padding.
- Do not change `--json` output fields.

## Example
- Input: `turnSpeed = 0.667`, `isActivity = true`, `colW = 24`, `colors = false`.
- Output: `0.7 step/s`.
- Negative boundary: `turnSpeed = 0.0` outputs `0 step/s` (no decimal places).

## Acceptance criteria
- AC-001: `go test -race -v -run TestLiveTUI_FormatSpeed_Fractional ./internal/tui`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that `formatSpeed(0, 0.667, true, 24, false)` returns `0.7 step/s`.
- AC-002: `go test -race -v -run TestLiveTUI_FormatSpeed_ZeroMuted ./internal/tui`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts that zero speed returns `0 step/s` without trailing decimal places (negative case).

## If blocked
Stop affected work; report evidence and the decision needed. Do not guess.

## Return
Return status, summary, changed files, case results, commands/exits, evidence, candidate, blocker.
