# Facts: TUI Projects view with drill-down

## Architectural Invariants & Constraints
- Project identity comes from `thermal.ProjectKey` only. Symlinks resolve first, subdirectories fold into the git root, and one directory counts once.
- Totals for a window must equal `thermal projects --json` for the same window. The parity check is a test.
- Model attribution is partial: Devin and the activity-only tools record no models, so model lines state their coverage instead of showing misleading shares.
- Estimated cost stays labeled. Stored cost wins per day, exactly as in the static reports.

## File & Interface Contracts
- New view files under `internal/tui/views/`: `projects.go` (table plus bars), `project_detail.go` (tool mix, model mix, weekly trend, heatmap).
- Sort keys: tokens, cost, days, recent. Filters: range and tool.
- Reuse `thermal.AggregateProjects` and `thermal.AggregateModels`; add no aggregation logic inside the view layer.
