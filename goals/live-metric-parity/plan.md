# Plan: Make live burn metrics reflect recorded tokens and turns

## Execution Steps
- [ ] **Task 1 — Use canonical metric inputs**: Reuse the canonical treatment established in web-telemetry-parity; derive token burn only from token-bearing rows and cache hit rate from CacheRead/(Input+CacheRead+CacheWrite). Keep unavailable breakdowns explicit.

  Acceptance and verification: Activity-only input generates no token burn; cache-write-only prompts have zero cache-hit rate; today's estimates follow no-estimate.

- [ ] **Task 2 — Track turns and source changes accurately**: Use recorded turn counts, not session count. Define first appearance, source disappearance, counter reset and midnight transitions so historical data is not emitted as new spend.

  Acceptance and verification: One additional turn in an existing session increments turns; baseline initialization and reset do not emit lifetime spikes.

- [ ] **Task 3 — Verify live snapshots and deltas**: Add fixed-clock sequences covering mixed tools, pricing, midnight and reset. Document corrected metrics in MIGRATION.md.

  Acceptance and verification: Run live unit/CLI tests under -race and compare the initial day's totals against static reports.

- [ ] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/live-metric-parity/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
Start after `web-telemetry-parity` and `live-polling-cpu-budget`. Preserve the bounded refresh/cancellation contract and CPU budget. Read prerequisite contracts before editing shared code.

## Verification
`go test -race ./internal/thermal ./cmd/thermal -run 'Live|Sim'`

No animation or TUI layout work. Dependency serializes shared CLI wiring and establishes the common metric contract.
