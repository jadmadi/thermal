# Plan: TUI Projects view with drill-down

## Execution Steps
- [ ] **Task 1 — Ranked table**: Project table with token and cost bars, sortable columns, and filters for range and tool.
- [ ] **Task 2 — Drill-down detail**: Enter opens a project detail with tool mix, model mix, weekly trend line, and a project heatmap.
- [ ] **Task 3 — Parity**: Totals for the same window must equal `thermal projects --json`, checked by a test rather than by eye.
- [ ] **Task 4 — Golden view tests**: Table and detail views at 80x24 and 120x40, color forced off.
- [ ] **Task 5 — Attribution commit**: Commit with `Goal-Ref: tui-projects-view` using a `feat:` subject, then run `sila goals`.
