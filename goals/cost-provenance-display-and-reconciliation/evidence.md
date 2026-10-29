# Evidence: Enforce tilde prefix on mixed-cost totals and surface missing-price provenance on leaderboard and dashboard

Date: 2026-10-29
Project: thermal-streak
Goal: cost-provenance-display-and-reconciliation

## Changes

- `internal/thermal/format.go`:
  - `FormatTotalCost`: enforces tilde prefix `~` whenever `estimated > 0`, ensuring mixed recorded and estimated spend is never mistaken for purely billed invoice spend.
- `internal/thermal/types.go`:
  - `ToolResult`: added `MissingPricing []string` and `UnattributedTokens int64` fields to pass unpriced provenance from scan to renderers.
- `cmd/thermal/main.go`:
  - Computed `estCost`, `toolMissing`, and `toolUnattr` across all uncosted days (`d.Cost == 0 && d.Tokens > 0`) regardless of whether `data.Summary.Cost > 0`.
  - Passed `dashMissing` and `dashUnattr` to `render.RenderDashboard`.
- `internal/render/dashboard.go`:
  - `RenderDashboard`: updated signature to accept `extraArgs ...any` for estimated cost, missing models, and unattributed tokens.
  - When both recorded and estimated cost are present, displays `~$XX.XX spent ($YY.YY recorded + ~$ZZ.ZZ est)`.
  - Renders unattributed-token and `"No pricing for: ..."` disclosures in dashboard footer when present.
- `internal/render/leaderboard.go`:
  - Displays `formatEstimatedCost(r.Summary.Cost + r.EstimatedCost)` on table rows when a tool has both recorded and estimated spend.
  - Sums both recorded and estimated spend into footer totals unconditionally when estimation is enabled.
  - Renders unattributed-token and `"No pricing for: ..."` disclosures in leaderboard footer.
- `internal/render/dashboard_test.go` and `internal/render/leaderboard_test.go`:
  - Added unit tests `TestRenderDashboard_MixedCostAndDisclosures` and `TestRenderLeaderboard_MixedCostAndDisclosures`.

## Verification

```text
$ CGO_ENABLED=1 go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...
ok  	github.com/jadmadi/thermal/internal/render	1.043s
ok  	github.com/jadmadi/thermal/internal/thermal	1.016s
ok  	github.com/jadmadi/thermal/cmd/thermal	24.138s

$ ./scripts/simulated_user_gate.sh
Simulated User Gate Summary: 77 passed, 0 failed (out of 77 checks)
✔ RELEASE GATE PASSED: All simulated user commands, formatting, semantics, and language verified.
```

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cost-provenance-display-and-reconciliation",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "cost provenance mixed totals and missing price disclosures", "matches": 3}],
  "memos": [{"id": "fact_61f4ed95", "text": "FormatTotalCost and leaderboard rows prefix with ~ whenever estimated cost is non-zero, ensuring mixed spend is never mistaken for purely billed invoice spend."}],
  "lessons": [{"id": "fact_151772b7", "symptom": "Leaderboard totals disagreed with monthly reports because tools with both recorded and uncosted days had uncosted days excluded from estimated spend."}],
  "cases": [{"id": "fact_e3009ea1", "intent": "Reconcile mixed recorded and estimated cost across leaderboard and report surfaces"}],
  "facts_verify": {"exit_code": 2},
  "handoff": {"id": "hoff_438fed9b584f"},
  "waivers": [{"verb": "facts", "reason": "Pre-existing grounding drift reporting missing files from earlier relocated projects."}]
}
```
