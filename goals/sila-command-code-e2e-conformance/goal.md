# Goal: Sila Command Code E2E Conformance

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: Thermal consumer conformance; Sila core conformance is upstream
**Upstream ownership**: Sila core changes belong to `harness-session-memory-bridge`, `mcp-continuity-9verb-and-tier0-parity`, `trigger-contract-full-surface-routing`, and the Sila CLI conformance goals.
**Legacy surface note**: Hermetic Command Code + Sila conformance harness

## Goal Description
Prove the whole utilisation program in a hermetic Command Code session. The harness must exercise the adapter, hooks, preflight, MCP, nine verbs, persona/audit selection, skill authoring, goal evidence, claim hygiene, onboarding, scorecard, and feedback, and fail on any silent skip.

## Problem and Evidence
- Every prior goal has a unit-level contract but no end-to-end proof that the agent actually invokes Sila.
- The original failure was behavioral, so unit tests alone cannot close it.
- No harness models a missing Sila binary, stale wiring, or skipped preflight.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-mandatory-preflight-gate, sila-session-lifecycle-hooks, sila-project-skill-authoring, sila-utilization-scorecard
- **Sequence**: 304
- **Shape**: ship
- **Tier**: roadmap

## Upstream Ownership
This goal is a Thermal adoption layer. It does not implement Sila core changes. Sila CLI owns `harness-session-memory-bridge`, `mcp-continuity-9verb-and-tier0-parity`, `trigger-contract-full-surface-routing`, and the Sila CLI conformance goals.

## Files to Touch
- `scripts/sila-command-code-e2e.sh`
- `scripts/sila-preflight.sh`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `goals/sila-command-code-e2e-conformance/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Harness**: Create a temporary HOME/repo and a fake or real Command Code session driver.
2. **Happy path**: Run preflight, prime, recall, goal claim, memo/lesson/case, persona selection, skill load, evidence, certify, scorecard, feedback, handoff, sync.
3. **Negative paths**: Missing Sila, stale heal, MCP unavailable, no ready goal, claim collision, missing evidence, bot trailer, hook timeout.
4. **Assertions**: Fail on skipped required trigger, duplicate side effect, missing artifact, or non-zero unexpected exit.
5. **Golden transcript**: Compare normalized command/event sequence and scorecard fields.
6. **Isolation**: No mutation of the user real HOME, Sila store, git remotes, or credentials.
7. **CI**: Run in CI and as a pre-push release gate.
8. **Docs**: Explain how to run and extend the harness.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-command-code-e2e-conformance/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-command-code-e2e-conformance/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not implement Sila core CLI, MCP, store, hook-bridge, persona, skill-governor, or goal-engine changes here; consume the upstream Sila CLI goals.
- Do not weaken assertions to make the harness pass.
- Do not touch real user config or remote refs.
- Do not replace the simulated user gate; add this as a Sila-specific gate.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should the harness use a real Sila database or a fixture?
- Can Command Code be driven headlessly in CI?