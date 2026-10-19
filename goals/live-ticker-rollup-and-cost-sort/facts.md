# Facts: Roll up live ticker bursts and add cost-first decision sorting

## Observed Evidence
- `internal/thermal/live.go:767-785`, `:801-816`, and `:819-834` prepend each new event to `lt.events`; there is no grouping or burst count. `lt.maxEvents` is 30 (`:91`). The ticker therefore shows near-identical rows for a project/model that emits repeated deltas.
- `internal/tui/live.go:424-427` renders `RecentEvents` in slice order, which is newest-first because of the prepends. There is no sort key or grouping pass.
- `internal/tui/live.go:176-201` handles only `q`/`ctrl+c`, `esc`, space, `c`, and `?`. There is no `[s]` key.
- `internal/tui/live.go:430-436` renders a fixed footer with Pause/Reset/Help/Quit; it has no sort-mode indicator. `renderHelp` at `:481-499` lists the same keys.
- `internal/tui/live.go:913-951` formats one row per `LiveEvent`; there is no burst count, group total, or top-spender line.
- The ticker width budget is computed in `calcColWidths` (`:867-887`) and used for six columns. Grouping adds a count/label and a possible top-spender line, so the layout must remain width-aware.
- The prerequisite `live-ticker-unit-and-cost-clarity` introduces separate TOKENS/STEPS columns, `CostKind`, and the outlier helper. This goal must consume those, not reimplement them.
- `internal/thermal/live_test.go` has no grouping tests; `internal/tui/live_test.go:109-183` exercises the existing key flow and must be extended for `[s]`.

## Interpretation
The ticker is an event log, not a decision summary. Grouping by `(Tool, Model, Project)` over a bounded 60-second window preserves recency while collapsing repeated bursts. A burst count plus summed values gives the user the `thermal-streak 3 bursts 1.7M ~$0.01` view from the feedback. Sorting by cost is the decision view; the top-spender line makes the answer visible even when the ticker is truncated.

Grouping belongs in `internal/thermal` as a pure function so the TUI does not invent aggregation and the behavior is unit-testable without Bubble Tea. The tracker's JSON event stream should remain ungrouped for machine consumers; grouping is a presentation concern.

Cost ordering must not put unpriced groups first. A partial sum must be marked (`≥$X` or `—`) so an unpriced member does not look free.

## Ownership and Constraints
- Grouping helper/tests: `internal/thermal/live.go`, `internal/thermal/live_test.go`.
- Ticker rendering, key handling, footer/help, top-spender line: `internal/tui/live.go`, `internal/tui/live_test.go`.
- Feature metadata: `project.jsonc`.
- No JSON schema or tracker-event changes; `--json`/`--stream` are unaffected.
- Preserve the prerequisite columns, cost statuses, outlier helper, baseline exclusion, and color/width guarantees.
- No new dependencies; no unbounded work over the 30-event cap.

## Verification Contract
```bash
go test -race ./internal/thermal ./internal/tui -run 'Live|Group|Rollup|Sort|Spender'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

Cases:
- Three events same tool/model/project at t, t+10s, t+20s -> one group with count 3, latest timestamp t+20s, summed tokens/steps/cost.
- Same key at t and t+61s -> two groups (outside the 60s window).
- Two models or two projects -> separate groups.
- `[s]` twice returns to time order; cost order is descending; unpriced group last.
- Baseline event present -> excluded from groups and medians.
- Cost mode shows `Top spenders (60s):`; time mode does not; all-unpriced shows `—`.
