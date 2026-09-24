# Goal: Avoid rescanning unchanged Codex rollout transcripts

## Goal Description
Add bounded source-aware rollout caching with exact parity against full scans and measured warm-run gains.

Priority: **P2**. Created from the [2026-09-24 evaluation](../../docs/evaluations/2026-09-24-thermal.md). Planning only; implementation is pending.

## Dependencies & Execution Order
- **Mode**: Dependent (sequenced)
- **Depends On**: ingestion-performance-baseline
- **Sequence**: 69
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/loaders/codex.go`
- `internal/loaders/codex_test.go`
- `internal/loaders/codexcache.go`
- `internal/loaders/benchmark_test.go`
- `docs/PERFORMANCE.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)

## Done Condition
1. Unchanged Codex transcripts are reused without rescanning bodies and all cached results match full scans.
2. Append/rewrite/corrupt-cache cases remain correct under bounded concurrency.
3. Benchmark evidence shows improvement against the prerequisite baseline without claiming an unmeasured universal latency.
4. Focused verification, full race suite, CLI build and simulated user gate pass; include required migration documentation and an implementation-attributed commit before closure.

## Scope Boundary
Do not change token attribution policy or cache other tools here. Any observable aggregation/schema change requires the six-part migration entry. This is a candidate targeted optimization; baseline measurement must confirm its benefit before implementation.
