# Plan: Make verification independent of host data and color settings

## Execution Steps
- [ ] **Task 1 — Isolate process state and color expectations**: Set and restore all relevant color variables per color test. Use temporary HOME and explicit GROK_HOME/DSH_HOME/HERMES_HOME for collection tests; freeze clocks and pricing inputs.

  Acceptance and verification: Focused and full tests pass with NO_COLOR set/unset and CLICOLOR_FORCE set/unset without scanning personal logs.

- [ ] **Task 2 — Make the shell gate fixture-first and deterministic**: Always provision known fixtures, isolate cache/tool-root overrides, use cached fixture pricing or offline mode, separate stdout/stderr, and use bounded command timeouts. Keep any real-data smoke run explicit and separate.

  Acceptance and verification: Gate produces identical semantic assertions with populated or empty host stores and does not depend on network access.

- [ ] **Task 3 — Add user-level semantic assertions**: Assert static/web/live token parity, unknown receipt evidence, partial coverage and database cache isolation after their owning goals land. Include real server start/stop, SSE cancellation, and deterministic 100ms live collector-count/pause/quit assertions; run sustained CPU measurements separately from timing-sensitive unit tests.

  Acceptance and verification: A controlled regression in each repaired contract must fail an assertion; syntactically valid JSON with wrong totals is insufficient.

- [ ] **Checkpoint — Review scope and release evidence**: Inspect only owned files; confirm acceptance criteria, source-data safety, required six-part migration entry, focused checks, full race tests, build and simulated gate. Record actual command outcomes, including any environment limits.
- [ ] **Attribution commit**: Stage only this goal's implementation and evidence; inspect `git diff --cached --name-only`; create an impact-first Conventional Commit ending in `(goals/hermetic-verification-gates/goal.md)`, then run `sila goals`. Do not push without explicit authorization. Planning commits must not close this goal.

## Dependencies and Handoff
Start after `receipt-accounting-coverage`, `live-metric-parity`, `devin-cache-source-isolation`, `localhost-request-boundary`, `live-polling-cpu-budget`. Read prerequisite contracts before editing shared code.

## Verification
`go test -race ./...; ./scripts/simulated_user_gate.sh`

Do not hide regressions by globally forcing colors or skipping assertions. Repair the two color tests early if needed; final integration assertions depend on the listed remediation goals.
