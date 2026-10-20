# Plan: Fix live ticker units, truncation, sparkline glyphs, and model-cost semantics

## Execution Steps
- [x] **Task 1 — Fix rate/speed/burned formatting**: Rewrite formatters to emit full units and allocate width for units; add compact fallback that drops step/token secondary values before truncating. Acceptance: golden frame has no `step…`/`tok…`; `+60 step`; rate/speed units complete.
- [x] **Task 2 — Unify sparkline and cache glyphs**: Choose one zero-cell glyph and one cache precision; update render/tests. Acceptance: populated and all-zero sparklines follow one policy; live/daily cache values match.
- [x] **Task 3 — Label model-cost basis**: Add/consume `costKind`/`basis` for todayModels and render `Est.`/token share; update JSON/text. Acceptance: model cost is never presented as the reconciled Today total; `costKind` exists in JSON.
- [x] **Task 4 — 80/120 golden tests**: Add hybrid/long-name fixtures at both sizes; assert no clipped units and preserved controls. Acceptance: `go test -race ./internal/tui -run TestLiveTUI` passes.
- [x] **Task 5 — Docs and attribution commit**: Update migration/README/project.jsonc; run full gate; commit with the goal token. Acceptance: gate clean; commit ends in `(goals/live-ticker-unit-and-rollup-contract/goal.md)`.

## Dependencies and Handoff
The formatting goal owns shared cost/precision helpers; the reconciliation goal owns canonical model-cost basis; this goal owns live presentation. The gallery goal captures the corrected frames.

## Verification
```bash
go test -race ./internal/tui ./internal/thermal -run 'Live|Ticker|Speed|Rate|Spark|CostBasis'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
