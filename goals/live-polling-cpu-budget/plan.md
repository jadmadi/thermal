# Plan: Keep 100ms live monitoring within a CPU budget

## Execution Steps
- [ ] **Task 1 — Reproduce and define the budget**: Measure the exact TTY command and NDJSON mode at 100ms and 1s using unchanged and changing synthetic stores. Separate initial scan, metadata checks, parsing, aggregation, pricing and rendering. Record source size and CPU/wall time; use the goal's <=10% one-core idle target after warmup and <=2s fixture freshness target.

  Verification: durable before/after results on the same machine, at least 30 seconds of steady-state measurement, and a counted fake collector proving how many expensive polls happen.

- [ ] **Task 2 — Bound collection independently of display cadence**: Schedule collection outside synchronous TUI Update; permit at most one refresh in flight, coalesce ticks and discard stale completions. Reuse normalized unchanged-source results with conservative invalidation and budget cheap change checks. Keep the 100ms display cadence and cheap decay updates independent from expensive full scans; use the same collector policy in NDJSON mode.

  Verification: unchanged ticks do not call the full loader repeatedly; a slow collector does not block pause/quit or create queued goroutines; changed fixtures update without missing or double-counting usage. No network fetch or full pricing recomputation on every unchanged tick.

- [ ] **Task 3 — Verify lifecycle and resource behavior**: Add deterministic fake-clock/poll-count tests for startup, slow/erroring sources, interval validation, source changes, pause/resume, midnight and shutdown. Measure TTY and streaming CPU at both intervals and report CPU, freshness and output-parity tradeoffs. Add required migration documentation if interval semantics or validation changes.

  Verification: `go test -race ./internal/tui ./cmd/thermal -run Live`, existing two-size view/key-flow tests where affected, the 30-second CPU budget, and bounded cancellation under a slow collector.

- [ ] **Checkpoint — Release evidence**: Verify exact metrics against the pre-change fixture, complete `go test -race ./...`, CLI build and simulated gate, inspect owned files, and record actual resource measurements rather than only a timer unit test.
- [ ] **Attribution commit**: Stage only this implementation and evidence, inspect staged paths, and commit an impact-first Conventional Commit ending in `(goals/live-polling-cpu-budget/goal.md)`. Run `sila goals`; do not push without authorization. Planning commits do not close the goal.

## Dependencies and boundaries
Ready immediately. Coordinate shared main.go/MIGRATION.md edits with receipt and web goals; those goals are not prerequisites for reducing live CPU. No new daemon, process-global watcher, or unbounded background worker. Increasing the interval alone does not satisfy this goal.
