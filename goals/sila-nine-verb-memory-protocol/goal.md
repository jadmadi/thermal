# Goal: Sila Nine Verb Memory Protocol

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: AGENTS.md, Sila skills, and session evidence

## Goal Description
Turn the nine-verb Sila memory protocol into explicit, testable trigger moments. The goal is not to call every verb in every session; it is to make the required verb unambiguous at the moment it matters and to capture the evidence.

## Problem and Evidence
- The full MUST protocol is in Cursor rules and never reached Command Code.
- AGENTS.md lacks doc, case, facts, guide, and learn triggers.
- This session re-derived prior lessons instead of recalling them.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-mandatory-preflight-gate, sila-mcp-surface-parity
- **Sequence**: 304
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `AGENTS.md`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `docs/SILA-COMMAND-CODE.md`
- `scripts/sila-preflight.sh`
- `goals/sila-nine-verb-memory-protocol/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Trigger table**: Define when each verb is required, optional, or forbidden, with examples.
2. **Recall**: Require recall before guessing, web search, or workaround for an error, package, or unfamiliar API.
3. **Doc**: Require doc search before integrating an external API or changing a dependency.
4. **Memo**: Require a memo immediately after a durable architectural or policy decision.
5. **Lesson**: Require a lesson after a non-trivial fix, including cause and fix.
6. **Case**: Require a case when introducing an architectural seam or pattern.
7. **Facts**: Require facts verify before declaring a goal complete.
8. **Prime and gate**: Require prime at session start and gate before commit.
9. **Handoff**: Require handoff and sync at session end.
10. **Quality controls**: Reject duplicate, empty, or performative entries; link entries to goals and files.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-nine-verb-memory-protocol/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-nine-verb-memory-protocol/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not turn optional verbs into ceremonial calls.
- Do not record secrets or prompt contents in the store.
- Do not change the meaning of existing Sila fact/lesson records.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Which verbs are required for read-only questions?
- How many recall hits must be reviewed before proceeding?
- Should feedback be part of the nine verbs or a tenth loop?
