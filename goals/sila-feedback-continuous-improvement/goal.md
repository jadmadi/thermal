# Goal: Sila Feedback Continuous Improvement

**Priority**: P2
**Program**: Sila Utilisation Program
**Implementation surface**: Sila feedback ledger and upstream issue workflow

## Goal Description
Turn Sila feedback into a closed loop. The goal starts with `fb_heal-contract-protocol`, defines triage and upstream escalation, links feedback to goals, and verifies resolution rather than leaving an open ledger entry.

## Problem and Evidence
- `fb_heal-contract-protocol` is open and not linked to the new program goals.
- There is no rule for local feedback versus anonymous sharing versus upstream issue.
- No metric shows whether feedback was resolved.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-utilization-scorecard
- **Sequence**: 311
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `docs/SILA-UTILISATION-PROGRAM.md`
- `docs/SILA-COMMAND-CODE.md`
- `goals/sila-feedback-continuous-improvement/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Triage**: Classify issue, tip, idea; severity; affected surface; owner.
2. **Linkage**: Link feedback IDs to goals and evidence; reject orphan feedback.
3. **Share policy**: Define when to keep local, share anonymously, or file upstream.
4. **Dedup**: Search existing feedback and issues before creating a new entry.
5. **Resolution**: Close only after a linked goal or explicit wont-fix decision.
6. **Metrics**: Report open age, resolution, duplicates, and upstream status.
7. **Tests**: Cover create, duplicate, share, submit, resolve, reopen.
8. **Docs**: Publish the feedback runbook.
9. **Privacy**: Never include secrets, prompts, or private paths in shared feedback.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-feedback-continuous-improvement/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-feedback-continuous-improvement/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not file public issues without user consent.
- Do not share anonymously by default.
- Do not resolve feedback without evidence.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should `fb_heal-contract-protocol` be filed upstream?
- Who owns anonymous sharing?
