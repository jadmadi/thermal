# Goal: Sila Utilization Scorecard

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: Thermal session measurement; Sila doctor/utilization gate is upstream
**Upstream ownership**: Sila core changes belong to `doctor-mcp-and-trigger-utilization-gate`, `memory-adoption-observability`, and `cross-agent-benchmark-scorecard` in `sila-cli`.
**Legacy surface note**: Session ledger, CLI report, and Command Code scorecard

## Goal Description
Define and measure "100% Sila utilisation" without turning it into command-count theater. The scorecard compares expected trigger moments to actual verb/tool calls, records skip reasons, and reports coverage by Sila family.

## Problem and Evidence
- The current gap was invisible because no session measured expected versus actual Sila usage.
- A command count would reward empty calls, not correct timing.
- No report shows which Sila families were never used.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-mandatory-preflight-gate, sila-session-lifecycle-hooks, sila-project-skill-authoring
- **Sequence**: 303
- **Shape**: ship
- **Tier**: roadmap

## Upstream Ownership
This goal is a Thermal adoption layer. It does not implement Sila core changes. Sila CLI owns `doctor-mcp-and-trigger-utilization-gate`, `memory-adoption-observability`, and `cross-agent-benchmark-scorecard` in `sila-cli`.

## Files to Touch
- `scripts/sila-scorecard.sh`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `docs/SILA-COMMAND-CODE.md`
- `goals/sila-utilization-scorecard/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Definition**: Define utilisation as coverage of required trigger moments plus evidence quality, not raw command count.
2. **Session ledger**: Record timestamp, verb/tool, trigger, outcome, skip reason, goal, and evidence reference.
3. **Expected set**: Compute expected verbs from task type, goal state, code changes, debugging, and completion.
4. **Coverage matrix**: Report continuity, knowledge, roadmap, personas/audits, skills, quality gates, setup, and feedback.
5. **Score**: Produce a bounded score with required failures, optional misses, and performative-call penalties.
6. **Reports**: Emit text, JSON, and a one-line status for CI/preflight.
7. **Thresholds**: Fail when required moments are missing; warn on optional misses.
8. **Tests**: Fixture sessions for full, partial, skipped, and performative usage.
9. **Docs**: Explain why 100% is not "run every command".

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-utilization-scorecard/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-utilization-scorecard/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not implement Sila core CLI, MCP, store, hook-bridge, persona, skill-governor, or goal-engine changes here; consume the upstream Sila CLI goals.
- Do not reward optional commands that add no value.
- Do not store prompt contents or secrets.
- Do not make the score a vanity metric.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Which task classes require personas/audits?
- What threshold is acceptable for optional verbs?
- Should the score live in Sila core or the repo?