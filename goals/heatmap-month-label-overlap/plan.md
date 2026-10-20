# Plan: Eliminate overlapping heatmap month labels

## Execution Steps
- [ ] **Task 1 — Add collision metadata**: Track occupied label columns and clear/skip overlaps; preserve the six-space indent. Acceptance: unit test with a synthetic 52-week late-October window produces no `ONov` and contains a readable `Nov`.
- [ ] **Task 2 — Add fixed-date table tests**: Cover late-October, December/January, leap year, and 4/26/52/104-week cases at 80/120/160 columns. Acceptance: `go test -race ./internal/render -run TestHeatmapMonth` passes; no label exceeds width.
- [ ] **Task 3 — Golden single-tool views**: Update/add OpenCode and Agy golden snapshots; assert no partial month words. Acceptance: `go test -race ./internal/render -run TestDashboard` passes; header text is stable.
- [ ] **Task 4 — Docs and attribution commit**: Add a short migration note if the label policy changes; run full gate; commit with the goal token. Acceptance: full gate clean; commit ends in `(goals/heatmap-month-label-overlap/goal.md)`.

## Dependencies and Handoff
The gallery goal will capture the fixed header. Do not change heatmap data or chart formatting.

## Verification
```bash
go test -race ./internal/render -run 'Heatmap|Dashboard'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
/tmp/thermal-test opencode --no-color | grep -E 'Nov|ONov'
./scripts/simulated_user_gate.sh
```
