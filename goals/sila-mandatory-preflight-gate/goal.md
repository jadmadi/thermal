# Goal: Sila Mandatory Preflight Gate

**Priority**: P0
**Program**: Sila Utilisation Program
**Implementation surface**: Thermal agent enforcement layer; Sila core is upstream
**Upstream ownership**: Sila core changes belong to `harness-session-memory-bridge`, `trigger-contract-full-surface-routing`, `facts-ledger-integrity-campaign`, and `doctor-mcp-and-trigger-utilization-gate` in `sila-cli`.
**Legacy surface note**: Command Code preflight gate and /goal bootstrap

## Goal Description
Make Sila preflight a hard gate before any code edit. The gate must produce a signed session artifact and be integrated into the `/goal` launch path, so an agent cannot treat Sila as optional without an explicit recorded override.

## Problem and Evidence
- The `/goal` bootstrap asked for implementation and live evidence but did not require Sila startup.
- AGENTS.md says `sila salam` first and `sila heal`/`prime` at task start, but no gate enforced it.
- The agent edited files before any `salam`, `heal`, or `prime` call.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 300
- **Shape**: ship
- **Tier**: roadmap

## Upstream Ownership
This goal is a Thermal adoption layer. It does not implement Sila core changes. Sila CLI owns `harness-session-memory-bridge`, `trigger-contract-full-surface-routing`, `facts-ledger-integrity-campaign`, and `doctor-mcp-and-trigger-utilization-gate` in `sila-cli`.

## Files to Touch
- `scripts/sila-preflight.sh`
- `AGENTS.md`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `.commandcode/config.json`
- `goals/sila-mandatory-preflight-gate/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Preflight command**: Add a script or Sila command that runs salam, heal check, prime, goals-ready, and recall for the stated symptom.
2. **Blocking semantics**: Block file edits and implementation claims until the preflight artifact exists and is fresh for the session.
3. **Evidence artifact**: Record command, version, exit code, timestamp, active goal, and whether a recall hit was reviewed.
4. **Goal launch integration**: Modify the `/goal` launch contract or wrapper to invoke the gate before exploration and before any edit.
5. **Read-only exception**: Allow questions and inspection without a write stamp, but still run salam and prime where possible.
6. **Skip/override**: Require an explicit reason and user-visible warning for any skipped check; never silently bypass.
7. **Staleness**: A preflight older than the session or a changed goal invalidates the stamp.
8. **Tests**: Simulate no Sila, stale heal, no goals, recall hit, no MCP, and a clean start.
9. **Docs**: Document the gate in AGENTS.md, the program runbook, and the `/goal` help text.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-mandatory-preflight-gate/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-mandatory-preflight-gate/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not implement Sila core CLI, MCP, store, hook-bridge, persona, skill-governor, or goal-engine changes here; consume the upstream Sila CLI goals.
- Do not weaken the gate for convenience.
- Do not make the gate depend on network access.
- Do not duplicate the goal-claim rules owned by the claim-hygiene goal.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Can Command Code block a tool call with a pre-tool-use hook?
- Should the gate run before goal claim or after?
- What is the exact freshness window?