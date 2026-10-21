# Goal: Sila Goal Claim Lifecycle Hygiene

**Priority**: P2
**Program**: Sila Utilisation Program
**Implementation surface**: Sila goal claims, commit attribution, and agent identity

## Goal Description
Make goal claims and attribution honest. The goal removes forced claims, resolves dependency collisions properly, requires session focus, enforces commit subject/goal tokens, and prevents bot co-author trailers in agent-generated commits.

## Problem and Evidence
- Earlier work used `sila goals claim --force` because of path collisions.
- The repo accrued 55 `CommandCodeBot` co-author trailers; they were removed by rewriting history.
- Goal closure depended on commit tokens but not on plan completion or evidence.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-mandatory-preflight-gate, sila-goal-evidence-certification
- **Sequence**: 308
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `AGENTS.md`
- `scripts/check.sh`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `goals/sila-goal-claim-lifecycle-hygiene/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Claim discipline**: Forbid `--force` except with an explicit user-approved recovery reason.
2. **Dependency-first**: Resolve collisions by dependency or release, not override.
3. **Session focus**: Require `sila goals focus <slug> --session=<id>` and surface the active goal in the preflight artifact.
4. **Stale release**: Use `sila goals release --stale` and `sila goals prune --claims` instead of manual deletion.
5. **Commit contract**: Require a feat/fix/refactor/perf/chore subject with the exact goal token at the end.
6. **Plan completion**: Require every plan task checked before commit.
7. **Attribution hygiene**: Reject bot co-author trailers and non-user author identities in agent commits.
8. **Tests**: Cover collision, force rejection, stale claim, missing token, unchecked task, and bot trailer.
9. **Docs**: Publish claim recovery and commit examples.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-goal-claim-lifecycle-hygiene/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-goal-claim-lifecycle-hygiene/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not bypass existing Sila claims by deleting the store.
- Do not rewrite published history without explicit consent.
- Do not add a bot identity as a workaround.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should the claim lock be per path or per goal?
- Should Sila reject bot co-author trailers itself?
- Should goal tokens be structured trailers instead of subject suffixes?
