# Facts: Enforce tilde prefix on mixed-cost totals and surface missing-price provenance on leaderboard and dashboard

## Observed Evidence
- `internal/render/format.go:formatTotalCost` currently prefixes `~` only when `stored == 0 && estimated > 0`. When `stored > 0 && estimated > 0`, it returns `$502.59` without `~`, making a total that is 50% estimated look like a purely recorded invoice amount.
- `internal/render/leaderboard.go:217` (`if r.Summary.Cost <= 0 && estimate && r.EstimatedCost > 0`) and `internal/render/leaderboard.go:357` (`if r.Summary.Cost > 0 { ... } else if estimate && r.EstimatedCost > 0`) drop `r.EstimatedCost` whenever `r.Summary.Cost > 0`, whereas `cmd/thermal/json.go:99-100` sums both `r.Summary.Cost` and `r.EstimatedCost`.
- `cmd/thermal/main.go:716` and `cmd/thermal/main.go:809` discard the `missing []string` slice returned by `pricer.PriceDay(d)`, so neither `thermal` (leaderboard) nor `thermal --tool <name>` (dashboard) discloses unpriced models or unattributed tokens.

## Architectural Invariants & Constraints
- Purely recorded totals (`stored > 0 && estimated == 0`) must render as `$X.XX` (no `~`).
- Any total or row with `estimated > 0` must render as `~$X.XX` (with `~`) and state the exact `recorded + ~estimated` split in the footer.
- `--no-estimate` must suppress all estimated amounts, `~` prefixes, and missing-pricing notes.

## File & Interface Contracts
- `internal/render/format.go`: `formatTotalCost(total, stored, estimated float64) string`
- `internal/thermal/types.go`: `ToolResult` (`MissingPricing []string`, `UnattributedTokens int64`)
- `internal/render/leaderboard.go`: `RenderLeaderboard`
- `internal/render/dashboard.go`: `RenderDashboard`

## Verification Contract
```bash
CGO_ENABLED=0 go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```
