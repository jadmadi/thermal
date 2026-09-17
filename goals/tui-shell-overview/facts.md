# Facts: TUI shell and Overview view

## Architectural Invariants & Constraints
- `thermal dashboard` is the only TUI entry point. Plain `thermal` keeps printing the leaderboard so pipes, scripts, and JSON stay stable.
- A non-TTY stdout gets a short message pointing at the static commands, never a broken render.
- Every view total must equal the matching static command for the same window. Parity is tested, not eyeballed.
- Warm start to the first frame stays under 300ms on this machine. The Devin cold scan shows a loading frame.
- Layout must survive resize and widths from 80 to 200 columns.
- No color-only encoding. Legends and labels carry the meaning; `--no-color` and dumb terminals stay readable.

## File & Interface Contracts
- New package `internal/tui`: `app.go` (model, update, view), `data.go` (load once, aggregate on demand), `views/` for per-view models, `theme.go` for the palette, `keys.go` for the key map.
- Key map: 1-5 tabs, tab and shift-tab cycle, j/k move, s cycle sort, t toggle tokens and cost, r cycle range (30d, 90d, 1y, all), / filter, enter drill-down, esc back, ? help, q quit.
- Theme reuses the existing palette: 38;5;40 green for accent, 38;5;245 for muted text, 1;38;5;255 for emphasis.
