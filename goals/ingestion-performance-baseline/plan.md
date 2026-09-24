# Plan: Measure cold and warm CLI performance on representative fixtures

## Execution Steps
- [ ] **Task 1 — Define reproducible workload fixtures**: Generate small and representative large SQLite/JSONL datasets for Codex, Devin, OpenCode/MiMoCode and directory scanners, including append, rewrite and unchanged cases. Keep private logs out of fixtures.

  Acceptance and verification: Record fixture sizes, row/file counts, seed, toolchain, hardware and exact command; validate output totals before timing.

- [ ] **Task 2 — Measure the complete execution path**: Capture cold/warm/changed-input CLI timings, p50/p95, allocations, bytes read and stage costs where useful. Separate pricing network latency, rendering and ingestion; include --offline warm invocation.

  Acceptance and verification: Repeated unchanged runs establish a baseline; benchmarks cannot claim sub-10ms solely from an in-process helper.

- [ ] **Task 3 — Publish budgets and prioritize measured hotspots**: Document the target's measured scope, results and gaps; correct the nonexistent helper reference in AGENTS.md without weakening the governance target. Rank follow-up cache slices using measured cost.

  Acceptance and verification: Commit runnable benchmark commands and a baseline report; add a stable regression comparison, avoiding machine-independent absolute timing assertions.

- [ ] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/ingestion-performance-baseline/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
No prerequisites. Coordinate shared test/document files if another goal is active.

## Verification
`go test ./internal/loaders -run '^$' -bench . -benchmem; ./scripts/benchmark.sh`

Measurement only. Do not optimize all loaders at once or make unmeasured speed claims. scripts/benchmark.sh is a proposed deliverable, not an existing verified command.
