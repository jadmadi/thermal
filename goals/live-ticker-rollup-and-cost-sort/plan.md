# Plan: Roll up live ticker bursts and add cost-first decision sorting

## Execution Steps
- [ ] **Task 1 — Add the pure grouping helper**: Define a `LiveEventGroup` (key, count, latest time, summed tokens/steps/cost, unpriced flag, outlier flag) and `GroupLiveEvents(events, window, now)` in `internal/thermal/live.go`. Filter baseline events, group only events within the 60-second window, and preserve deterministic ordering for equal keys.
  Acceptance and verification: `go test -race -run TestLiveGroupEvents ./internal/thermal`; same-key triples collapse, >60s events split, different model/project keys stay separate, baseline is excluded.

- [ ] **Task 2 — Render rollup rows**: Update the ticker to render groups with `N bursts` (or `Nx` at 80 columns), summed tokens/steps, and cost using the prerequisite columns/status. Single events render without a burst label. Show `≥$X` when a group has some unpriced members and `—` when all are unpriced.
  Acceptance and verification: `go test -race -run 'TestLiveTUI_RollupRow|TestLiveTUI_RollupResponsive' ./internal/tui` at 80x24 and 120x40; no merged model/project groups; no partial cost presented as complete.

- [ ] **Task 3 — Add the `[s]` sort toggle**: Add the key to `Update`, track `sortMode` in `LiveModel`, and order groups by latest time or known cost. Cost mode puts priced groups first, then unpriced, with token/time tie-breaks. Update the footer and help screen to show `Sort: time|cost` and `[s]`.
  Acceptance and verification: `go test -race -run 'TestLiveModel_SortToggle|TestLiveModel_KeyFlow' ./internal/tui`; pressing `s` twice restores time order; the key does not trigger pause/reset/help/quit.

- [ ] **Task 4 — Add the top-spender line**: In cost mode, aggregate displayed group costs by project, sort descending, take up to three, and render `Top spenders (60s): project ~$0.01 · ...` above the ticker. Render `—` when no known cost exists. Do not render it in time mode.
  Acceptance and verification: `go test -race -run TestLiveTUI_TopSpenders ./internal/tui`; priced projects sort correctly; unpriced-only data shows `—`; time mode omits the line.

- [ ] **Task 5 — Recompute outliers after grouping**: Use the prerequisite outlier helper on group totals (tokens and steps separately), excluding baseline. Flag groups over 3x the median and keep the `! >3x median` legend visible in both sort modes.
  Acceptance and verification: `go test -race -run 'TestLiveGroupOutliers|TestLiveTUI_RollupOutlierMarker' ./internal/thermal ./internal/tui`; a grouped total over threshold is flagged; fewer than three groups and median zero do not flag.

- [ ] **Task 6 — Width, no-color, and JSON invariance**: Verify 80x24 and 120x40 layouts with color off, including burst count, both numeric columns, sort indicator, and top-spender line. Assert `thermal live --json` is unchanged by the view toggle/grouping.
  Acceptance and verification: `go test -race -run 'TestLiveTUI_Golden|TestLiveJSON_UnchangedByGrouping' ./internal/tui ./cmd/thermal`; no ANSI under `--no-color`; no wrapped lines.

- [ ] **Task 7 — Metadata and attribution commit**: Update `project.jsonc` to describe burst rollup and `[s]` cost sorting. Run the full gate and create an impact-first commit ending in `(goals/live-ticker-rollup-and-cost-sort/goal.md)`. Run `sila goals` and do not push.

## Dependencies and Handoff
Start after `live-ticker-unit-and-cost-clarity` closes. Reuse its `LiveEvent` cost status, normalized columns, and outlier helper. Do not alter the tracker's event stream or JSON schema; grouping is a TUI projection over the existing 30-event cap.

## Verification
```bash
go test -race ./internal/thermal ./internal/tui -run 'Live|Group|Rollup|Sort|Spender'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
