# Plan: Sila Onboarding Doctor Enforcement

## Execution Steps
- [ ] **Task 1 — Baseline and reproduction**: Capture the 11-pass/1-warn baseline and trace why `--strict` exits clean. Acceptance: Evidence contains command transcripts and a standalone `Verdict: PASS` only after the gap is reproduced.
- [ ] **Task 2 — Contract and failure modes**: Write the interface contract, state machine, and explicit failure modes for the goal before editing implementation files. Acceptance: A reviewer can answer what happens on missing Sila, stale wiring, missing MCP, timeout, duplicate event, and partial write.
- [ ] **Task 3 — Core implementation**: Make the warning fail and add the CI health job + preflight refusal. Acceptance: The declared paths change; unrelated product code and unrelated goals are untouched.
- [ ] **Task 4 — Positive-path tests**: Add table-driven unit tests and at least one integration test for the expected path. Acceptance: Tests fail before the implementation and pass after it; no skips hide missing behavior.
- [ ] **Task 5 — Negative-path and recovery tests**: Exercise absent tools, stale state, duplicate calls, timeouts, and rollback paths. Acceptance: Failures are explicit, non-zero where appropriate, and leave no partial or corrupted state.
- [ ] **Task 6 — Command Code integration**: Wire the behavior into a real or hermetic Command Code session, not only a Go unit test. Acceptance: The session artifact proves the surface was actually invoked and not merely documented.
- [ ] **Task 7 — Docs and runbook**: Update the repo runbook, AGENTS.md or docs surface, and `project.jsonc` metadata. Acceptance: A new agent can follow the runbook without reading this transcript.
- [ ] **Task 8 — Evidence and certification**: Run `sila facts verify`, write the evidence file, certify with `sila goals certify`, and confirm `unverified_count` is zero for this goal. Acceptance: Certification exists and evidence is grounded in files that still exist.
- [ ] **Task 9 — Attribution and handoff**: Check every plan task, commit with the goal token under a non-docs subject, reconcile `sila goals`, and record `sila handoff` and `sila sync`. Acceptance: Commit and handoff exist; no bot co-author trailer is introduced.

## Verification
- `go test -race ./...`
- `./scripts/simulated_user_gate.sh`
- `./scripts/output_gallery_gate.sh`
- `sila goals test sila-onboarding-doctor-enforcement`
- `sila goals certify sila-onboarding-doctor-enforcement --evidence=goals/sila-onboarding-doctor-enforcement/evidence.md`

## Dependencies and Handoff
- Prerequisites: sila-command-code-adapter, sila-session-lifecycle-hooks, sila-fact-ledger-integrity.
- Do not start implementation until every prerequisite goal is implemented and certified.
- Do not mark this goal implemented until every task above is checked and an evidence file contains `Verdict: PASS`.
- Commit with `feat: ... (goals/sila-onboarding-doctor-enforcement/goal.md)`; docs-only commits do not close Sila goals.
- Run `sila goals test sila-onboarding-doctor-enforcement` and `sila goals` after the commit; confirm the claim clears.
