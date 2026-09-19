# Plan: TUI shell and Overview view

## Execution Steps
- [x] **Task 1 — App shell** (`internal/tui/app.go`): Bubble Tea v2 program with tab routing, key map (1-5, tab, j/k, s, t, r, /, enter, esc, ?, q), quit, and resize handling.
- [x] **Task 2 — Data adapter** (`internal/tui/data.go`): Load once through `loaders.LoadToolData` for every tool with data, keep day rows and project rows in memory, and aggregate per view and range on demand.
- [x] **Task 3 — Overview view**: KPI cards (tokens, cost, current streak, active tools, projects), ranked tool table with inline bars, per-tool sparklines, and a 12-week share strip.
- [x] **Task 4 — Command and guard**: `thermal dashboard` launches the TUI. A non-TTY stdout prints a pointer to the static commands instead, and `--json` never opens the TUI.
- [x] **Task 5 — Tests**: Golden `View()` snapshots at 80x24 and 120x40 with color forced off, plus a key-flow test for tab and range switching.
- [x] **Task 6 — Attribution commit**: Commit with `Goal-Ref: tui-shell-overview` using a `feat:` subject, then run `sila goals`.
