# Facts: Reuse unchanged directory transcripts across invocations

## Observed Evidence
- `docs/PERFORMANCE.md` (run at `e2cdc72`): warm p50 **7.8ms** on the five-source fixture vs warm p50 **42.4ms** at 1,000 Claude files (n=20, p95 52.9ms) — the 42.4ms figure exceeds the 10ms governance budget.
- The gap is attributed to directory-backed sources that glob and parse every `.jsonl`/`.json` file per invocation; SQLite-backed sources (Devin, Codex) already reuse disk snapshots via `internal/loaders/cache.go` (`devinCachePath`) and `internal/loaders/codexcache.go` (`codexRolloutCachePath`).
- Section 7.1 of PERFORMANCE.md names this the largest measured hotspot and asks for "a fast directory modification check or an incremental scan cache for directory loaders, with the same read-only and exactness guarantees as the SQLite snapshot caches".
- The live collector re-scans directory sources on its ~200ms budget at the 100ms cadence (measured 1.32% CPU at 1,000 files); any reuse mechanism should remain usable there while preserving measured freshness (133ms at 100ms, 909ms at 1s).

## Constraints
- Read-only ingestion: never write to tool stores; caches live under `~/.cache/thermal/<tool>/` like existing snapshots.
- Exact results: reuse must be a cache, not an approximation; any signature ambiguity falls back to a full scan.
- Bounded concurrency (existing worker pools) and bounded memory; no unbounded goroutines.
- No new dependencies, no daemon, no watcher.
- The 10ms target remains the governance invariant to measure against; a per-machine MISSED verdict is disclosed, not hidden.

## File & Interface Contracts
- `internal/loaders/cache.go`: shared snapshot identity/serialization helpers (follow `DevinCache`: canonical path, size, mtime, max ids).
- `internal/loaders/claude.go`, `codewhale.go`, `droid.go`, `commandcode.go`: scan loops that currently re-read files.
- `scripts/benchmark/fixtures.go`: large fixture (1,000 files × 30 records) drives the verdict.
- `internal/loaders/benchmark_test.go`: per-loader microbenchmarks for warm/cold/delta paths.

## Verification Contract
- `go test -race ./internal/loaders` with parity tests (full scan vs cached scan, mutation matrix).
- `./scripts/benchmark.sh` exit 0 with updated warm verdicts.
- Then the common final checks from `tasks/plan.md`.
