# Goal: Keep the live monitor responsive without saturating CPU at 100ms

## Goal Description
Fix the user-reported high CPU usage from `thermal live --interval 100ms`. Separate display updates from expensive source collection, reuse unchanged snapshots, and keep refresh work bounded and cancellable. Fast display cadence must remain usable without repeatedly rescanning full histories.

Priority: **P1**, first in this evaluation's recommended execution order. Planning only; no CPU fix has been implemented.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 60
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)
- [Evaluation](../../docs/evaluations/2026-09-24-thermal.md)

## Done Condition
1. After warmup on the documented unchanged 1,000-file/30,000-record fixture, `live --interval 100ms` and its NDJSON mode average at most 10% of one CPU core over 30 seconds, with before/after results on the same machine. Measure TTY and streaming separately; preserve exact metric outputs and document workload limits.
2. At most one collection is in flight. TUI updates do not synchronously run full loaders; pause/quit are handled within 250ms under an injected slow collector, pending work is cancelled or safely discarded, and missed ticks cannot create a backlog. Unchanged history is reused; active changes appear within the documented collection latency budget (target <=2s after warmup on the fixture).
3. Cover 100ms and 1s, unchanged and changing sources, slow/erroring loaders, pause/resume and cancellation. Keep 100ms usable rather than only raising the default. Document any changed `--interval` meaning or new validation with six-part MIGRATION.md entries; complete focused tests, full race suite, build and simulated gate.

## Scope Boundary
Use the existing in-process collector and loader/cache path. No watcher service, background daemon, new database, blanket cache TTL that hides changes, or per-tick goroutine fan-out. No general TUI redesign. The 10% CPU budget is a proposed acceptance target, not a measured current result or promise for arbitrary datasets. This goal can start before the broader performance baseline; coordinate shared CLI changes with the other ready goals.
