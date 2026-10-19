# Facts: Preserve fractional sub-1.0 speeds in live monitor (live-speed-fractional-display)

## Architectural Invariants & Constraints
- TUI card widths and visual alignments must be preserved (column width checks via `ansi.StringWidth`).
- Formatting must remain allocation-efficient and color-safe.

## Problem Evidence & Root Causes
- Location: `internal/tui/live.go:586-603`
- Code:
  ```go
  func formatSpeed(tokSpeed, turnSpeed float64, isActivity bool, colW int, colors bool) string {
      if isActivity {
          if turnSpeed <= 0 {
              return theme.TextMuted.Sprint(colors, "0 step/s")
          }
          return fmt.Sprintf("%s step/s", thermal.CompactNumber(int64(turnSpeed)))
      }
      if tokSpeed <= 0 && turnSpeed <= 0 {
          return theme.TextMuted.Sprint(colors, "0 tok/s")
      }
      if tokSpeed > 0 && turnSpeed > 0 && colW >= 24 {
          return fmt.Sprintf("%s · %s step/s", thermal.CompactNumber(int64(tokSpeed)), thermal.CompactNumber(int64(turnSpeed)))
      }
      if tokSpeed <= 0 {
          return theme.TextMuted.Sprint(colors, "0 tok/s")
      }
      return fmt.Sprintf("%s tok/s", thermal.CompactNumber(int64(tokSpeed)))
  }
  ```
- Defect: Speeds are passed as `float64` (e.g. `turnPerSec := turnPerMin / 60.0`), but `formatSpeed` casts them to `int64` via `int64(turnSpeed)` before formatting.
- Symptom: At `Rate: 40 step/min`, `turnSpeed = 0.667 step/s`. `int64(0.667) == 0`. The live monitor prints `0 step/s` alongside `Rate: 40 step/min` and an 85% burning flame.
- Fix: When `speed > 0 && speed < 1.0`, format as `fmt.Sprintf("%.1f step/s", turnSpeed)` (or `tok/s`).

## File & Interface Contracts
- Edit: `internal/tui/live.go` in `formatSpeed`.
- Test: `internal/tui/live_test.go`.
- Document: `docs/MIGRATION.md`.
