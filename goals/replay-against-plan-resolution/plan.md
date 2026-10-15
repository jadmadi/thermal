# Plan: Resolve plan identifiers passed to replay --against

## Execution Steps
- [ ] **Step 1 — Exact plan lookup**: Add `LookupPlanExact` to `internal/pricing/subscription.go` (case-insensitive exact id or exact name; no prefix matching) with unit tests covering: plan id, plan name, case variants, a model id that shares a plan prefix (`claude`), and an unknown key.
  Acceptance and verification: `go test ./internal/pricing`.
- [ ] **Step 2 — Resolve in the CLI**: In `runReplayReport`, when `opts.Against` matches a plan exactly, simulate it with `SimulatePlan` (same path as `--compare`); otherwise keep `SimulateModel`. Preserve output ordering and JSON keys.
  Acceptance and verification: parity test asserting `--against claude-pro` and `--compare claude-pro` produce identical plan rows; negative test for `--against claude`.
- [ ] **Step 3 — Simulated user coverage**: Extend the replay simulated test so a plan target asserts a subscription row (type "sub", non-missing pricingStatus) and an unknown model target keeps the PAYG row.
  Acceptance and verification: `go test ./cmd/thermal -run ReplaySimulation`.
- [ ] **Step 4 — Documentation**: Add the six-part migration entry (`--against` plan resolution) under the current unreleased section and update the README replay note; keep the PROMISES.md replay claim accurate.
  Acceptance and verification: entry has heading/anchor, detection signal, rationale, before/after, fix, escape hatch; README wording matches behavior.
- [ ] **Step 5 — Common final checks and attribution commit**: `env -u NO_COLOR go test -race ./...`, `go vet ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, `./scripts/simulated_user_gate.sh`; commit with a `fix:` subject ending in `(goals/replay-against-plan-resolution/goal.md)`, then `sila goals`.
  Acceptance and verification: four commands exit 0; commit attributed.

## Verification
- `go test -race ./internal/pricing ./internal/thermal ./cmd/thermal -run 'Replay|Plan'`
- `go test ./cmd/thermal -run ReplaySimulation`
- Common final checks from `tasks/plan.md`
