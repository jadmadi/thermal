# Plan: Avoid rescanning unchanged Codex rollout transcripts

## Execution Steps
- [x] **Task 1 — Cache unchanged rollout summaries**: Use resolved source identity, precise modtime/size and a versioned parser schema; atomically cache only normalized usage/diff metrics and scan diagnostics. Never cache prompts or tool output bodies.

  Acceptance and verification: Unchanged second load opens no rollout bodies; deleting or replacing a file removes stale contributions; read-only source handles remain unchanged.

- [x] **Task 2 — Handle append and rewrite safely**: Maintain sufficient cumulative parser state for append deltas, including incomplete final lines and cumulative token_count semantics. Detect shrink/rewrite and fall back to full scan; reapply current DB tokens/model/project mapping.

  Acceptance and verification: Appending usage or diff events matches a cold parse; truncation, partial writes, DB-only changes and scan errors cannot double count or silently omit data.

- [x] **Task 3 — Prove parity and measured improvement**: Keep the worker limit, add race/corruption/concurrent-cache tests, verify source DB mmap configuration, and update the baseline with warm and delta timings.

  Acceptance and verification: All summary/daily/project/model/diff totals and warnings match uncached results; representative warm invocations improve by measured amounts with unchanged output.

- [x] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [x] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/codex-incremental-rollout-cache/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
Start after `ingestion-performance-baseline`. Read prerequisite contracts before editing shared code.

## Verification
`go test -race ./internal/loaders -run 'Codex|Cache'; ./scripts/benchmark.sh`

Do not change token attribution policy or cache other tools here. Any observable aggregation/schema change requires the six-part migration entry. This is a candidate targeted optimization; baseline measurement must confirm its benefit before implementation.
