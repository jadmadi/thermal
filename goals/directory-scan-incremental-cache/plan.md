# Plan: Reuse unchanged directory transcripts across invocations

## Execution Steps
- [x] **Phase 1 — Measure and design the cache key**: Re-produce the large-fixture warm baseline (`./scripts/benchmark.sh --out /tmp/bench-dir-cache --stream-seconds 5 --skip-stream` style is acceptable for iteration; the full run is the closing evidence). Define the snapshot identity: resolved directory root, file set, per-file size + mtime (monotonic), and a content fallback when size/mtime are ambiguous (same-second rewrites). Follow the `DevinCache` pattern in `internal/loaders/cache.go`.
  Acceptance and verification: a written identity definition plus unit tests for key stability (same tree → same key) and change detection (append/rewrite/truncate/delete → new key).
- [x] **Phase 2 — Implement for Claude, then CodeWhale/Droid/command-code**: Serialize per-file scan results (per-day rows, project days, model maps, warnings) under `~/.cache/thermal/<tool>/`; reuse when the directory signature matches; fall back to a full scan on any doubt. Keep bounded worker pools; never write into the source tree.
  Acceptance and verification: parity tests compare a cached second scan against a full scan (deep equality of rows and totals) and a mutation matrix (append, rewrite, truncate, delete, corrupt cache JSON, missing cache dir, read-only cache dir) for each loader.
- [x] **Phase 3 — Re-measure and update evidence**: Run the full `./scripts/benchmark.sh`; update `docs/PERFORMANCE.md` with the new numbers and verdict (PASS if ≤10ms at 1,000 files, otherwise re-measured MISSED with the delta). Add the migration entry only if an observable contract (cache location/behavior) changes. Update the AGENTS.md loader-quirk note.
  Acceptance and verification: `./scripts/benchmark.sh` exit 0 with exact-fixture checks; docs updated; no user-store mutation.
- [x] **Phase 4 — Common final checks and attribution commit**: `env -u NO_COLOR go test -race ./...`, `go vet ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, `./scripts/simulated_user_gate.sh`; commit with a `perf:` or `fix:` subject ending in `(goals/directory-scan-incremental-cache/goal.md)`, then `sila goals`.
  Acceptance and verification: all four commands exit 0; commit attributed; no remote push.

## Verification
- `go test -race ./internal/loaders`
- `./scripts/benchmark.sh`
- Common final checks from `tasks/plan.md`

## Risks
- Same-mtime rewrites: if a file changes twice within the filesystem timestamp resolution, the cache must detect it (size or content hash fallback) or fall back to a full scan.
- Memory growth: serialize to disk rather than holding all file results in memory; reuse must not increase peak RSS materially.
