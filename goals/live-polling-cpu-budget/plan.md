# Plan: Keep 100ms live monitoring within a CPU budget

## Execution Steps
- [x] **Task 1 — Reproduce and define the budget**: Measure the exact TTY command and NDJSON mode at 100ms and 1s using unchanged and changing synthetic stores. Separate initial scan, metadata checks, parsing, aggregation, pricing and rendering. Record source size and CPU/wall time; use the goal's <=10% one-core idle target after warmup and <=2s fixture freshness target.

  Verification: durable before/after results on the same machine, at least 30 seconds of steady-state measurement, and a counted fake collector proving how many expensive polls happen.
  - Baseline (Before, 30s steady-state on 1,000-file/30,000-record fixture):
    - NDJSON 1s: 12.9% of one core (3.88s CPU / 30.0s)
    - NDJSON 100ms: 130.4% of one core (39.11s CPU / 30.0s)
    - PTY 120x40 1s: 12.6% of one core (3.77s CPU / 30.0s)
    - PTY 120x40 100ms: 95.9% of one core (28.76s CPU / 30.0s)

- [x] **Task 2 — Bound collection independently of display cadence**: Schedule collection outside synchronous TUI Update; permit at most one refresh in flight, coalesce ticks and discard stale completions. Reuse normalized unchanged-source results with conservative invalidation and budget cheap change checks. Keep the 100ms display cadence and cheap decay updates independent from expensive full scans; use the same collector policy in NDJSON mode.

  Verification: unchanged ticks do not call the full loader repeatedly; a slow collector does not block pause/quit or create queued goroutines; changed fixtures update without missing or double-counting usage. No network fetch or full pricing recomputation on every unchanged tick.

- [x] **Task 3 — Verify lifecycle and resource behavior**: Add deterministic fake-clock/poll-count tests for startup, slow/erroring sources, interval validation, source changes, pause/resume, midnight and shutdown. Measure TTY and streaming CPU at both intervals and report CPU, freshness and output-parity tradeoffs. Add required migration documentation if interval semantics or validation changes.

  Verification: `go test -race ./internal/tui ./cmd/thermal -run Live`, existing two-size view/key-flow tests where affected, the 30-second CPU budget, and bounded cancellation under a slow collector.
  - Acceptance (After, 30s steady-state on identical fixture and machine):
    - NDJSON 1s: 0.2% of one core (0.07s CPU / 30.0s) — 64x reduction
    - NDJSON 100ms: 0.9% of one core (0.28s CPU / 30.0s) — 145x reduction
    - PTY 120x40 1s: 0.7% of one core (0.20s CPU / 30.0s) — 18x reduction
    - PTY 120x40 100ms: 1.9% of one core (0.57s CPU / 30.0s) — 50x reduction

- [x] **Checkpoint — Release evidence**: Verify exact metrics against the pre-change fixture, complete `go test -race ./...`, CLI build and simulated gate, inspect owned files, and record actual resource measurements rather than only a timer unit test.
- [ ] **Attribution commit**: Stage only this implementation and evidence, inspect staged paths, and commit an impact-first Conventional Commit ending in `(goals/live-polling-cpu-budget/goal.md)`. Run `sila goals`; do not push without authorization. Planning commits do not close the goal.

## Dependencies and boundaries
Ready immediately. Coordinate shared main.go/MIGRATION.md edits with receipt and web goals; those goals are not prerequisites for reducing live CPU. No new daemon, process-global watcher, or unbounded background worker. Increasing the interval alone does not satisfy this goal.
