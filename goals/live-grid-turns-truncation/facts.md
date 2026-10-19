# Facts: Fix turns and cache-hit text truncation in narrow live monitor grid (live-grid-turns-truncation)

## Architectural Invariants & Constraints
- Live monitor TUI must look polished on standard 80x24 terminal windows without clipped words or unwanted ellipses in primary KPI cells.
- Zero-allocation string formatting where possible.

## Problem Evidence & Root Causes
- Location: `internal/tui/live.go:672-678` and `internal/tui/live.go:332`
- Code:
  ```go
  func formatTurnsAndHit(turns int, hit float64, colors bool) string {
      tStr := fmt.Sprintf("%d turns", turns)
      if hit > 0 {
          return fmt.Sprintf("%s (%.0f%% hit)", tStr, hit)
      }
      return tStr
  }
  ```
- Grid cell width: On an 80-column terminal, `cardWidth = 76`, `innerWidth = 72`.
  `colW = (72 - 4) / 3 = 22` to `24`.
  Row text: `Turns:    6642 turns (86% hit)` is 30 characters wide.
  `padCell(text, 24, false)` clips with ellipsis: `Turns:    6642 turns (86% h…`.
- Fix: Add width-aware fallback in `formatTurnsAndHit`:
  1. Try `fmt.Sprintf("%d turns (%.0f%% hit)", turns, hit)`
  2. If it exceeds available width, try `fmt.Sprintf("%s (%.0f%% hit)", thermal.CompactNumber(int64(turns)), hit)`
  3. If still too wide, fall back to `fmt.Sprintf("%d turns", turns)`.

## File & Interface Contracts
- Edit: `internal/tui/live.go` in `formatTurnsAndHit`
- Test: `internal/tui/live_test.go`
- Document: `docs/MIGRATION.md`
