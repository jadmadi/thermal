# Plan: Public promises backed by evidence

## Context and files

P2. Cwd `/mnt/Jad/github/lab/thermal-streak`. Blocked on all goal.md prerequisites.
Why: users need to distinguish observed local activity from inferred productivity and savings.
Read the evaluation, `tasks/plan.md`, `docs/PERFORMANCE.md`, `.github/workflows/release.yml`,
`internal/thermal/yield.go:CalculateYield`, and all prerequisite remediation evidence.

## Card C1: Public claim ledger

Allowed: create `docs/PROMISES.md`; edit `README.md` and `project.jsonc`.
For each claim record exact public wording/location, actual behavior, evidence command/artifact,
source revision, coverage limits and outcome (`verified`, `partial`, `unverified`).
Cover local data safety, token/cost coverage, cross-surface totals, receipts, yield, replay,
performance, audit and distribution/provenance. Scope zero-allocation to measured helpers.
Describe yield as recorded diff volume per token, not correctness or business productivity.
Describe verified receipts as observed command outcomes, not proof of task completion or source
revision validity. State supported receipt sources separately from the full ingestion registry.
Do not claim actual release verification from a workflow definition. Leave provenance unverified
unless an actual artifact is independently verified; no release/publishing work in this card.
Preserve existing README structure and user edits. No redesign or product expansion.

- [ ] CLM-201: every strong README promise maps to concrete evidence or an explicit qualification;
  `docs/PROMISES.md` links all prerequisite receipts. No “all goals implemented” acceptance shortcut.
- [ ] CLM-202: metadata descriptions match qualified claims; `lastEdit` updated. Independent reviewer
  compares wording to behavior and signs off; this is a human review criterion, not a grep proof.

## Card C2: Semantic release gate

Depends on C1 and all prerequisite implementation cards. Allowed: `scripts/simulated_user_gate.sh`
and `docs/PROMISES.md`; use prerequisite Go regression tests instead of cloning helper logic.
Gate runs the named receipt/replay/audit regression tests and fails if any fail. Keep all existing
107 checks. Add explicit nonempty assertions for test selection so a renamed/nonexistent test
cannot yield false success. Reuse synthetic fixtures and preserve private data isolation.
Performance evidence runs separately; a small cache timing check is not an end-to-end 10ms guarantee.

- [ ] CLM-203: introduce a temporary known failing expectation in an isolated copy, prove the gate
  exits nonzero, restore it, and rerun green. Attach both commands/exits bound to their exact diffs.
- [ ] CLM-204: common final checks from tasks/plan.md exit 0 on the final candidate; return
  ready_for_review, not self-approved. No remote push, release, or goal certification in planning.

Stop and return blockers after two repeated failures; schema and reviewer in `tasks/plan.md`.
