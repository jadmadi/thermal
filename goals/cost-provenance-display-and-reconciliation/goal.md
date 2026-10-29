# Goal: Enforce tilde prefix on mixed-cost totals and surface missing-price provenance on leaderboard and dashboard

## Chain of Custody (Isnad)
- **Registered By**: Pending Registration
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: @mystery-auditor | Report: `receipt_cost-provenance-display-and-reconciliation_1793262681`


## Goal Description
Ensure cost provenance is unambiguous across all Thermal CLI surfaces (`thermal` leaderboard, `thermal --tool <name>` dashboard, `daily`/`weekly`/`monthly` reports, and `projects` ranking) so that any total containing an estimate carries the `~` prefix (`~$`), tools mixing recorded and estimated days reconcile identically between the leaderboard and period reports, and missing-price / unattributed-token disclosures appear on the leaderboard and single-tool dashboard.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: pricing-catalog-and-partial-model-provenance
- **Sequence**: 240
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required

## Files to Touch
- Edit: `cmd/thermal/main.go`
- Edit: `internal/thermal/types.go`
- Edit: `internal/render/format.go`
- Edit: `internal/render/leaderboard.go`
- Edit: `internal/render/dashboard.go`
- Edit: `internal/render/report_test.go`

## Requirements
1. **Mixed-Cost Tilde Prefix (`internal/render/format.go`)**:
   - Update `formatTotalCost(total, stored, estimated float64) string` so that whenever `estimated > 0`, the formatted total carries the `~` prefix (`formatEstimatedCost(total)`), preventing a mixed recorded + estimated total from looking like a purely billed invoice amount.
2. **Mixed-Tool Cost Reconciliation (`cmd/thermal/main.go`, `internal/render/leaderboard.go`, `internal/render/dashboard.go`)**:
   - In `cmd/thermal/main.go` (single-tool dashboard path), compute `estCost` across all uncosted days (`d.Cost == 0 && d.Tokens > 0`) regardless of whether `data.Summary.Cost > 0`, matching the leaderboard and report aggregators.
   - In `internal/render/leaderboard.go` and `internal/render/dashboard.go`, when a tool has both `Summary.Cost > 0` and `EstimatedCost > 0` (with estimation enabled), display `formatEstimatedCost(r.Summary.Cost + r.EstimatedCost)` on the row and include `r.EstimatedCost` in the footer's estimated sum so leaderboard totals (`recorded + estimated`) reconcile with `thermal monthly` and `thermal projects`.
3. **Leaderboard & Dashboard Missing-Price Disclosures (`cmd/thermal/main.go`, `internal/thermal/types.go`, `internal/render/leaderboard.go`, `internal/render/dashboard.go`)**:
   - Populate `MissingPricing []string` and `UnattributedTokens int64` on `thermal.ToolResult` in `cmd/thermal/main.go` from `pricer.PriceDay(d)` and unattributed days (`d.Cost == 0 && len(d.Models) == 0 && d.Tokens > 0`).
   - Render the unattributed-token note and `"No pricing for: ..."` disclosure in `RenderLeaderboard` and `RenderDashboard` when estimation is enabled.

## Acceptance Criteria
- AC-001: `CGO_ENABLED=0 go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...` exits 0.
- AC-002: `formatTotalCost(502.59, 253.03, 249.56)` returns `"~$502.59"`, while `formatTotalCost(253.03, 253.03, 0)` returns `"$253.03"`.
- AC-003: Leaderboard recorded + estimated cost totals match `thermal monthly` and `thermal projects` totals, and `./scripts/simulated_user_gate.sh` exits 0.

## References
- **Shared Understanding & Fact Sheet**: [`goals/cost-provenance-display-and-reconciliation/facts.md`](facts.md)
- **Execution Plan**: [`goals/cost-provenance-display-and-reconciliation/plan.md`](plan.md)

## Memory Plan
- **Prime**: run `sila prime` (CLI) or `prime_context` (MCP) at task start and record the briefing timestamp.
- **Recall**: run `sila recall "<symptom or design question>"` before implementation; record each query and match count.
- **Writes**: record at least one `memo`, `lesson`, or `case` grounded to a file in this goal; use a `case` when the goal introduces a reusable mechanism.
- **Facts**: run `sila facts verify` before certification and record the exit code and violation delta.
- **Handoff**: run `sila handoff "<summary>" --open="..."` at session end and record the handoff ID.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`, per `memory-protocol-goal-contract`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`); unchecked tasks block the `implemented` transition.
2. Implementation is committed with `(goals/cost-provenance-display-and-reconciliation/goal.md)` or `[goal: cost-provenance-display-and-reconciliation]` attribution under a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` prefix.
3. Goal is certified (`goals/cost-provenance-display-and-reconciliation/certification.json` with `verified: true` or a signed `audit-receipt.json`) and any markdown evidence contains an explicit standalone `Verdict: PASS` declaration.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0` for delivered goals, with a clean `git status`.
