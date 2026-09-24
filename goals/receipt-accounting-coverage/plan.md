# Plan: Preserve recorded usage and disclose receipt coverage

## Execution Steps
- [x] **Task 1 — Preserve source usage and cost fidelity**: Remove step-to-token conversion. Deduplicate Claude usage by source message identity, preserve disjoint token categories and per-model estimates, retain native recorded cost, and expose unavailable/unpriced coverage rather than zero implying free.

  Acceptance and verification: Activity-only sessions contribute no token spend; duplicate messages count once; multi-model fixture estimates equal canonical pricing inputs.

- [x] **Task 2 — Disclose partial receipt coverage**: Normalize aliases through the existing registry before scanning. Merge uncovered tools as explicitly aggregate-only UNVERIFIED coverage, without pretending daily rows are observed session outcomes or double counting covered sources. Define session-first-day versus daily-spend window semantics explicitly.

  Acceptance and verification: A fixture containing Claude plus an aggregate-only tool accounts for both; `ccode receipt` matches `claude receipt`; missing transcripts cannot inflate verification rates.

- [x] **Task 3 — Pin receipt totals and coverage in CLI tests**: Add mixed-source, missing transcript, date-window and unavailable-pricing tests; document changed aggregation/schema semantics using MIGRATION.md.

  Acceptance and verification: Run focused receipt tests and CLI simulations; test all sources absent, partial coverage, duplicate records and multi-day sessions.

- [x] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [x] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/receipt-accounting-coverage/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
Start after `receipt-evidence-integrity`. Read prerequisite contracts before editing shared code.

## Verification
`go test -race ./internal/thermal ./cmd/thermal -run 'Receipt|Sim'`

Do not implement every unsupported transcript schema in this goal. Disclose unsupported evidence and preserve available usage. Stream existing receipt readers with bounded memory and warnings if their parsing path is changed.
