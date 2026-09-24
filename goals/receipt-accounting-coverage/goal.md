# Goal: Preserve recorded usage and disclose receipt coverage

## Goal Description
Remove invented activity tokens and prevent partial transcript coverage from silently dropping other tools.

Priority: **P1**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Dependent (sequenced)
- **Depends On**: receipt-evidence-integrity
- **Sequence**: 62
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/receipt.go`
- `internal/thermal/receipt_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Receipt token totals never synthesize usage from actions; pricing uses recorded disjoint categories.
2. Partial source coverage is explicit and cannot silently exclude spend or invent claimed/verified work.
3. Aliases, date scope and aggregate-only fallback behavior are deterministic, tested and migration-documented.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
Do not implement every unsupported transcript schema in this goal. Disclose unsupported evidence and preserve available usage. Stream existing receipt readers with bounded memory and warnings if their parsing path is changed.
