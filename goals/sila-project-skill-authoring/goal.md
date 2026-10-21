# Goal: Sila Project Skill Authoring

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: Thermal-specific project skills; Sila skill governor is upstream
**Upstream ownership**: Sila core changes belong to `global-and-workspace-sila-skill-provisioning` and `trigger-contract-full-surface-routing` in `sila-cli`.
**Legacy surface note**: Project skills, Sila skill governor, Command Code discovery

## Goal Description
Create and govern Thermal-specific Sila skills for the repeated workflows this repo actually needs: Sila preflight, goal execution, release verification, and output-gallery maintenance. The goal includes token-cost assessment, preview/apply/rollback, and Command Code discovery.

## Problem and Evidence
- `sila skills list` reports no project skills.
- The same Sila loop and Thermal gate sequence was manually rediscovered each session.
- Skill authoring exists in Sila but is invisible in AGENTS and MCP.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 302
- **Shape**: ship
- **Tier**: roadmap

## Upstream Ownership
This goal is a Thermal adoption layer. It does not implement Sila core changes. Sila CLI owns `global-and-workspace-sila-skill-provisioning` and `trigger-contract-full-surface-routing` in `sila-cli`.

## Files to Touch
- `.agents/skills/`
- `docs/SILA-COMMAND-CODE.md`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `AGENTS.md`
- `goals/sila-project-skill-authoring/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Skill selection**: Define the criteria for promoting repeated instructions into a skill.
2. **Initial skills**: Create `thermal-sila-preflight`, `thermal-goal-execution`, and `thermal-release-verification`.
3. **Content**: Each skill names triggers, commands, evidence, failure paths, and links to AGENTS; it must not duplicate the entire AGENTS file.
4. **Token budget**: Assess metadata versus full-load cost; avoid skills that cost more than the work they save.
5. **Preview/apply**: Use `sila skill preview`, `apply`, `remove`, and `assess` with dry-run defaults.
6. **Conflicts**: Detect duplicate skill names and harness collisions; provide rollback.
7. **Discovery**: Verify Command Code can see and load the skills without a manual copy step.
8. **Tests**: Cover creation, assessment, apply, load, collision, rollback, and removal.
9. **Docs**: Publish a skill inventory and when-not-to-create rule.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-project-skill-authoring/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-project-skill-authoring/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not implement Sila core CLI, MCP, store, hook-bridge, persona, skill-governor, or goal-engine changes here; consume the upstream Sila CLI goals.
- Do not create a skill for a one-off task.
- Do not duplicate AGENTS wholesale.
- Do not auto-apply skills without preview.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should the preflight be a skill, a hook, or both?
- Which Command Code skill directories are authoritative?