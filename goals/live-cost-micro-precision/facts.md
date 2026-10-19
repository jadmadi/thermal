# Facts: Support 4-decimal sub-cent precision in live event cost display (live-cost-micro-precision)

## Architectural Invariants & Constraints
- Financial figures must be accurate and never round up to deceptive thresholds.
- Fixed cell widths in the live events ticker (`costW = 9`) must be respected without wrapping or clipping.

## Problem Evidence & Root Causes
- Location: `internal/tui/live.go:659-670`
- Code:
  ```go
  func formatCost(c float64, colors bool) string {
      if c <= 0 {
          return theme.TextMuted.Sprint(colors, "$0.00")
      }
      if c < 0.01 {
          return "<$0.01"
      }
      if c >= 1000 {
          return fmt.Sprintf("~$%.0f", c)
      }
      return fmt.Sprintf("~$%.2f", c)
  }
  ```
- Symptom: In real-time token streaming, individual bursts typically consume 100 to 2,000 tokens on fast models (e.g. Gemini Flash, Claude Haiku, GPT-4o-mini). Each burst costs between $0.0001 and $0.005.
- Consequence: All micro-bursts display `<$0.01` or `$0.00`, obscuring real-time spend velocity across events.
- Solution: With `costW = 9`, values under $0.01 format cleanly as `~$0.0042` or `~$0.0005` (7-8 chars), fitting well within 9 cells.

## File & Interface Contracts
- Edit: `internal/tui/live.go` in `formatCost`
- Test: `internal/tui/live_test.go`
- Document: `docs/MIGRATION.md`
