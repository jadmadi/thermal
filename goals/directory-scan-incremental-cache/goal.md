# Goal: Reuse unchanged directory transcripts across invocations

## Goal Description
Close the 1,000-file warm-latency gap measured in [`docs/PERFORMANCE.md`](../../docs/PERFORMANCE.md): directory-backed sources (Claude, CodeWhale, Droid, command-code, and any future directory scanner) are re-walked and re-parsed on every invocation, so the warm p50 at 1,000 Claude files is **42.4ms** against the 10ms governance target, while SQLite-backed sources with delta caches stay far below it.

Add a bounded, disk-backed incremental scan cache for directory loaders with exact parity against full scans.

Priority: **P1** (largest measured gap against a governance invariant). Follow-up from the [2026-10-14 promise-vs-delivery evaluation](../../docs/evaluations/2026-10-14-promise-vs-delivery.md) and the ingestion-performance-baseline remediation.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 107
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/loaders/cache.go` (shared cache primitives, e.g. the Devin snapshot pattern)
- `internal/loaders/claude.go`
- `internal/loaders/codewhale.go`
- `internal/loaders/droid.go`
- `internal/loaders/commandcode.go`
- `internal/loaders/*_test.go` and `internal/loaders/benchmark_test.go`
- `scripts/benchmark/fixtures.go` (large-fixture verdict may flip from MISSED to PASS)
- `docs/PERFORMANCE.md`
- `docs/MIGRATION.md` (only if an observable contract changes)
- `AGENTS.md` (section 2.A.4 loader-quirk note: directory scanner caching now implemented)

## References
- [Evidence and constraints](facts.md)
- [Execution plan](plan.md)
- [Performance baseline §7.1](../../docs/PERFORMANCE.md)
- [Ingestion performance remediation result](../ingestion-performance-baseline/remediation-result.json)

## Done Condition
1. Unchanged directory transcripts are reused without re-reading file bodies, and cached results are **byte-identical** to full scans (token totals, per-day rows, project rows, model maps, warnings).
2. Append, rewrite, truncate, delete, and corrupt-cache cases remain correct; caches invalidate on directory identity (resolved-path canonicalization), file set, per-file size/mtime, and contents when ambiguous.
3. `./scripts/benchmark.sh` shows the large-fixture warm p50 within the 10ms budget on this machine, or the report re-measures honestly and updates the disclosed verdict; the small fixture's PASS verdict does not regress.
4. Read-only ingestion, bounded worker pools, cache files under `~/.cache/thermal/` only, no user-store mutation, and no daemon.
5. Focused verification, full race suite, CLI build and simulated user gate pass; migration documentation when the observable contract changes; implementation-attributed commit before closure.

## Scope Boundary
Do not change token attribution policy or cache other tools outside directory scanners in the first pass. Do not add a daemon, watcher, or background process; the cache must be a pure function of on-disk state. A missed target stays a disclosed gap, never a forced PASS.
