# Goal: Sila Session Lifecycle Hooks

**Priority**: P0
**Program**: Sila Utilisation Program
**Implementation surface**: Command Code host lifecycle config; Sila core bridge is upstream
**Upstream ownership**: Sila core changes belong to `harness-session-memory-bridge`, `active-lifecycle-hook-bridge`, and `project-onboarding` in `sila-cli`.
**Legacy surface note**: Command Code lifecycle config plus Sila hook bridge

## Goal Description
Install and verify session lifecycle hooks so each Command Code session automatically runs the Sila start and close loop. The hook layer is the enforcement point that turns the AGENTS prose into a real preflight and handoff.

## Problem and Evidence
- `.githooks/` contains only pre-commit and pre-push.
- `.commandcode/config.json` has no `hooks` entry.
- `sila hook` supports session-start/session-end events, but nothing invokes them for Command Code.
- The skipped start loop is the direct cause of the memory and persona gaps.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 301
- **Shape**: ship
- **Tier**: roadmap

## Upstream Ownership
This goal is a Thermal adoption layer. It does not implement Sila core changes. Sila CLI owns `harness-session-memory-bridge`, `active-lifecycle-hook-bridge`, and `project-onboarding` in `sila-cli`.

## Files to Touch
- `.commandcode/config.json`
- `scripts/sila-preflight.sh`
- `scripts/sila-session-end.sh`
- `docs/SILA-COMMAND-CODE.md`
- `AGENTS.md`
- `goals/sila-session-lifecycle-hooks/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Event contract**: Define session-start, user-prompt, pre-tool-use, stop, and session-end payloads and exit semantics.
2. **Start sequence**: Run heal check, prime, goals-ready, and active-claim lookup before the first edit.
3. **Close sequence**: Run handoff with summary/open items and sync once per session, without duplicate writes.
4. **Fail closed**: A missing Sila binary, stale heal, or failed prime must be reported and block implementation rather than pass silently.
5. **Idempotency**: Repeated start/stop events must not duplicate handoffs, syncs, or claim mutations.
6. **Timeouts**: Every hook has a bounded timeout and a visible failure path; no hook can hang the agent.
7. **No recursion**: Sila hooks must not invoke another Sila hook or spawn an unbounded sync loop.
8. **State stamp**: Write a session preflight artifact with command, exit code, duration, and Sila version.
9. **Offline/no-TTY**: Degrade explicitly, not silently, when network or TTY is absent.
10. **Tests and docs**: Cover normal, duplicate, missing, stale, timeout, and crash-recovery paths.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-session-lifecycle-hooks/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-session-lifecycle-hooks/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not implement Sila core CLI, MCP, store, hook-bridge, persona, skill-governor, or goal-engine changes here; consume the upstream Sila CLI goals.
- Do not implement the adapter rule generator here.
- Do not change Sila store schema unless the hook contract requires a new event field.
- Do not auto-sync on every tool call; only at defined lifecycle points.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Which Command Code hook event is closest to session start?
- Should heal run every session or only when stale?
- Should handoff be automatic or require a summary argument?