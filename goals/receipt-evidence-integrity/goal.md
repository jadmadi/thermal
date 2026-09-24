# Goal: Require observed command results for verified receipts

## Goal Description
Prevent missing, mixed or mismatched command outcomes from becoming verified work receipts.

Priority: **P1**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 61
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/receipt.go`
- `internal/thermal/receipt_test.go`
- `docs/MIGRATION.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. No missing or mismatched result can produce Tier 1 Verified.
2. Failures cannot be overridden by an incidental pass substring; actual invocation/result pairs are counted once.
3. Existing confirmed outcomes stay classified correctly and corrected CLI semantics have migration documentation.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
Do not redesign spend aggregation or add additional harness adapters here. Those changes belong to receipt-accounting-coverage.
