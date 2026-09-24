# Goal: Make live burn metrics reflect recorded tokens and turns

## Goal Description
Exclude activity-only counts from token burn and use actual turn and cache-read data for live metrics.

Priority: **P1**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Dependent (sequenced)
- **Depends On**: web-telemetry-parity
- **Sequence**: 64
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Live snapshots and burn deltas exclude activity-only counts and agree with canonical daily totals.
2. Completed turn counts reflect observed turns; counter resets and new sources do not create historical bursts.
3. Cache and estimated-cost semantics are correct and covered by deterministic tests.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
No animation or TUI layout work. Dependency serializes shared CLI wiring and establishes the common metric contract.
