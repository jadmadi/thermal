# Facts: Align flame intensity gauge and sparkline bracket widths (live-meter-bracket-alignment)

## Architectural Invariants & Constraints
- Gauges must remain pasteable, width-aware, and aligned in monospace terminals.
- Brackets and bars must strictly adhere to the terminal width without wrapping or truncation artifacts.

## Problem Evidence & Root Causes
- Location: `internal/tui/live.go:693-746` (`formatFlameLine`) and `internal/tui/live.go:748-805` (`formatSparkLine`)
- Flame line layout:
  Prefix: `  Flame: [`
  Width: `barW := innerWidth / 4` (16 to 32)
  Suffix: `] 85%  BURNING 🔥  ·  gemini-3.8-flash`
- Spark line layout:
  Prefix: `  Spark: [`
  Width: `availForSpark := innerWidth - len(prefix) - len(suffix) - len(peakStr) - 2` (20 to 60)
  Suffix: `]  Peak: 4 step/s`
- Visual Disparity: The closing `]` on the flame meter sits around column 28-40, while the closing `]` on the sparkline stretches up to column 68-75. Visually, the two stacked gauge panels feel uncoordinated.
- Solution: Coordinate gauge bar widths or align bracket widths so the visual flow down the dashboard is balanced and consistent.

## File & Interface Contracts
- Edit: `internal/tui/live.go` in `formatFlameLine` and `formatSparkLine`
- Test: `internal/tui/live_test.go`
- Document: `docs/MIGRATION.md`
