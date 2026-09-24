# Goal: Keep cached Devin totals isolated by source database

## Goal Description
Key Devin snapshots by canonical source identity and invalidate on relevant source changes.

Priority: **P1**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 65
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/loaders/cache.go`
- `internal/loaders/sqlite.go`
- `internal/loaders/sqlite_test.go`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Cached data cannot cross database identities or survive incompatible replacement.
2. Source changes affecting results invalidate cached state without writing user databases.
3. Corrupt/concurrent cache access preserves correct cold-scan results and remains best-effort.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
No general cache framework or changes to other tools. Add migration documentation only if user-visible flags/schema/aggregation semantics change beyond eliminating stale cache reuse.
