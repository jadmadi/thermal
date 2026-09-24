# Goal: Make verification independent of host data and color settings

## Goal Description
Use isolated fixtures and explicit color environments, and assert metric semantics in the release gate.

Priority: **P2**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Dependent (sequenced)
- **Depends On**: receipt-accounting-coverage, live-metric-parity, devin-cache-source-isolation, localhost-request-boundary
- **Sequence**: 67
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `scripts/simulated_user_gate.sh`
- `cmd/thermal/args_test.go`
- `cmd/thermal/simulated_user_test.go`
- `internal/render/help_test.go`
- `internal/server/server_test.go`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Verification results are independent of personal tool stores, ambient color variables and live pricing.
2. JSON stdout and diagnostic stderr are tested separately with bounded execution.
3. The gate detects the repaired semantic failures and can run offline on a fresh fixture home.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
Do not hide regressions by globally forcing colors or skipping assertions. Repair the two color tests early if needed; final integration assertions depend on the listed remediation goals.
