# Goal: Make web totals match canonical reports

## Goal Description
Reuse canonical token and cost aggregation in web telemetry and honor supported selection and estimation options.

Priority: **P1**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 63
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/server/server.go`
- `internal/server/server_test.go`
- `internal/thermal/telemetry.go`
- `cmd/thermal/main.go`
- `docs/MIGRATION.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Web token totals equal static totals for the same source set and window, including activity-only data.
2. Recorded cost, estimated cost and unknown pricing remain distinguishable.
3. Supported options work consistently and unsupported options fail rather than being silently ignored.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
No visual redesign, new telemetry store or cloud endpoint. Preserve established JSON keys unless a documented migration is necessary. Follow ui-craft if frontend rendering changes become necessary.
