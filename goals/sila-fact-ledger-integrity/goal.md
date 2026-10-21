# Goal: Sila Fact Ledger Integrity

**Priority**: P0
**Program**: Sila Utilisation Program
**Implementation surface**: Sila core knowledge store and fact ledger

## Goal Description
Investigate and remediate the hash-chain mismatches reported by sila facts verify. The goal must distinguish concurrent-write corruption from actual tampering, preserve the original store, and make verification fail closed before any goal can be certified.

## Problem and Evidence
- `sila facts verify` reports multiple `[tamper_mismatch]` facts with stored and computed hashes differing at recent blocks.
- The mismatches appeared after normal Sila writes in this session; the cause is not yet isolated.
- A red fact ledger invalidates evidence certification and any claim of 100 percent utilisation.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 299
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `.sila/sila.db`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `docs/SILA-COMMAND-CODE.md`
- `goals/sila-fact-ledger-integrity/evidence.md`
- `Sila upstream CLI repository: store and facts verification changes under cli/ in the external Sila checkout`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Reproduction**: Capture the complete facts verify output, affected IDs, block numbers, Sila version, and store path.
2. **Classification**: Determine whether the mismatch is concurrent-write corruption, interrupted chain repair, schema migration, or deliberate tampering.
3. **Backup**: Create a consistent store backup before any repair; never overwrite the only copy.
4. **Canonical chain**: Verify block ordering, canonical serialization, timestamp fields, and hash input.
5. **Concurrency**: Audit write locks and transactions; reproduce with parallel memo/lesson/feedback writes.
6. **Repair policy**: Provide quarantine, rechain, or restore options with explicit approval and a dry-run.
7. **Fail closed**: Make onboarding, evidence certification, and preflight fail when facts verify is red.
8. **Tests**: Cover clean chain, concurrent writes, interrupted write, deliberate tamper, migration, and long-chain verification.
9. **Docs**: Publish the integrity model, recovery runbook, and user-facing feedback guidance.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-fact-ledger-integrity/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-fact-ledger-integrity/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not silently rechain or delete facts.
- Do not expose private fact content in shared feedback.
- Do not certify any goal while the ledger is red.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Is the mismatch caused by concurrent Sila writes in this session?
- Should repair be automatic after backup or always human-approved?
- Does the hash chain include normalized timestamps?
