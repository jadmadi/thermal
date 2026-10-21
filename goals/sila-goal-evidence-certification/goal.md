# Goal: Sila Goal Evidence Certification

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: Sila goals certification, evidence templates, and Thermal gate

## Goal Description
Make goal completion evidence-based. Every delivered goal must have reproducible evidence, a standalone `Verdict: PASS`, a certification record, and a zero unverified count. The goal also remediates the current 88 unverified and 22 stale goals rather than accepting a green task checkbox count.

## Problem and Evidence
- Sila reports 114 implemented but only 4 verified, 22 stale, and 88 unverified.
- Previous goal commits closed on attribution but not certification.
- A 100% task count can coexist with unverified semantics.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-mandatory-preflight-gate, sila-nine-verb-memory-protocol, sila-fact-ledger-integrity
- **Sequence**: 307
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `goals/*/evidence.md`
- `goals/*/certification.json`
- `scripts/check.sh`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `goals/sila-goal-evidence-certification/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Evidence template**: Define commands, outputs, tree hashes, exit codes, environment, and limitations.
2. **Verdict rule**: Require a standalone `Verdict: PASS`; a bare checkbox or green test is insufficient.
3. **Certification**: Require `sila goals certify --evidence` or `sila goals audit` for every goal.
4. **Facts verification**: Run `sila facts verify` before certification and fail on missing evidence files.
5. **Stale/unverified remediation**: Classify the 110 unverified/stale goals into recertify, reopen, supersede, or archive.
6. **No false closure**: A docs-only commit, empty evidence, or skipped test cannot close a goal.
7. **Gate integration**: Add certification checks to the pre-commit/pre-push path or a dedicated Sila gate.
8. **Tests**: Exercise pass, missing verdict, missing file, stale evidence, and failed test.
9. **Docs**: Publish the evidence anatomy and remediation ledger.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-goal-evidence-certification/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-goal-evidence-certification/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not weaken Sila certification rules.
- Do not mass-mark goals verified without command evidence.
- Do not modify product behavior to make a test pass.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should all 88 unverified goals be reopened or audited in place?
- What is the minimum evidence for docs-only goals?
- Can certification run in CI without a live Sila store?
