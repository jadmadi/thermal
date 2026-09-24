# Plan: Require observed command results for verified receipts

## Execution Steps
- [x] **Task 1 — Represent missing command outcomes explicitly**: Use an optional exit status and correlate results to invocation IDs. Parse real nested Claude tool_result content. Keep unmatched, interrupted and unrelated results unknown; never execute recorded commands.

  Acceptance and verification: Fixtures cover two concurrent calls, reversed results, missing exit fields, no result at EOF, and result-only records.

- [x] **Task 2 — Classify only observed outcomes**: Remove optimistic EOF/default success. Give explicit nonzero exits priority; mixed pass/fail text with unknown exit cannot be verified. Do not count quoted command text or git tag as a successful test/commit.

  Acceptance and verification: Unknown evidence stays UNVERIFIED unless a separate agent claim exists; explicit pass and explicit failure remain distinguishable.

- [x] **Task 3 — Verify evidence semantics and document corrections**: Add regression fixtures and the six-part migration entry for corrected receipt classification; retain sanitized labels only.

  Acceptance and verification: Run `go test -race ./internal/thermal -run 'Receipt|Evidence|Outcome'`; probe unfinished command is no longer VERIFIED; run full gate before closing.

- [x] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/receipt-evidence-integrity/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
No prerequisites. Coordinate shared test/document files if another goal is active.

## Verification
`go test -race ./internal/thermal -run 'Receipt|Evidence|Outcome'`

Do not redesign spend aggregation or add additional harness adapters here. Those changes belong to receipt-accounting-coverage.
