# Goal: Measure cold and warm CLI performance on representative fixtures

## Goal Description
Establish reproducible end-to-end latency and allocation evidence for the documented cached invocation target.

Priority: **P2**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 68
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/loaders/benchmark_test.go`
- `scripts/benchmark.sh`
- `docs/PERFORMANCE.md`
- `AGENTS.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. The cached invocation claim has reproducible end-to-end evidence or a clearly documented gap.
2. Performance fixtures preserve correctness and include unchanged and changed source behavior.
3. Follow-up optimization priorities follow measured costs; no new daemon or telemetry store is introduced.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
Measurement only. Do not optimize all loaders at once or make unmeasured speed claims. scripts/benchmark.sh is a proposed deliverable, not an existing verified command.
