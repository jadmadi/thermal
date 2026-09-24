# Plan: Keep cached Devin totals isolated by source database

## Execution Steps
- [ ] **Task 1 — Key snapshots by canonical database identity**: Bind cache data to the resolved source path and a replacement-sensitive identity; reject legacy snapshots without that identity. Keep cache versioning explicit.

  Acceptance and verification: Two databases with identical row/count probes never reuse totals; symlink aliases resolve consistently; replacing a file cannot reuse old data.

- [ ] **Task 2 — Invalidate on relevant non-append changes**: Design and test a WAL-aware freshness key covering updates/deletes, session visibility/model/project changes and prompt-history activity. Use delta scans only when append-only assumptions have been established; otherwise rescan.

  Acceptance and verification: Update an existing row without changing MAX(row_id), swap hidden sessions preserving the count, change a model, and modify prompt-only activity; each result equals a cold scan.

- [ ] **Task 3 — Publish cache files safely and verify parity**: Use atomic cache replacement and private permissions, handle corrupt/truncated snapshots as misses, and prevent concurrent readers from seeing partial writes.

  Acceptance and verification: Run Devin tests with -race, simultaneous cache writers and cold-versus-warm comparisons; record probe changing from 100/100 to 100/900.

- [ ] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/devin-cache-source-isolation/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
No prerequisites. Coordinate shared test/document files if another goal is active.

## Verification
`go test -race ./internal/loaders -run 'Devin|Cache'`

No general cache framework or changes to other tools. Add migration documentation only if user-visible flags/schema/aggregation semantics change beyond eliminating stale cache reuse.
