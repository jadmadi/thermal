# Goal: Sila Mcp Surface Parity

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: Sila MCP server and Command Code MCP config

## Goal Description
Close the gap between Sila CLI capabilities and the MCP tools visible to Command Code. The goal is an explicit parity matrix, safe read/write tool classes, and tools for the surfaces that were invisible in this session.

## Problem and Evidence
- MCP exposed prime, recall, lesson, memo, handoff, and status, but not heal, persona, audit, skill, doc, case, facts, guide, or learn.
- The model cannot invoke CLI-only capabilities through MCP discovery.
- `sila --help` contained the full surface, but the agent did not run it.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-command-code-adapter
- **Sequence**: 303
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `.commandcode/mcp.json`
- `docs/SILA-COMMAND-CODE.md`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `goals/sila-mcp-surface-parity/evidence.md`
- `Sila upstream CLI repository: companion MCP changes under cli/ in the external Sila checkout`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Inventory**: Generate a command-to-MCP coverage matrix for every Sila command family.
2. **Add tools**: Expose heal/check, persona list/get, audit gap/squad, skill list/show/assess/create, doc search, case, facts, guide, learn, feedback, goals test/certify.
3. **Read/write classes**: Mark each tool read-only, dry-run, or mutating; require explicit confirmation for mutating calls.
4. **Schemas**: Provide JSON schemas and descriptions that name preconditions and output fields.
5. **Errors**: Return actionable errors for missing project, stale store, unsupported command, and denied mutation.
6. **Fallback bridge**: Add a documented command bridge for any capability that cannot be a native MCP tool.
7. **Capability negotiation**: Report server version and supported surface so Command Code can show missing tools.
8. **Tests**: Exercise every tool through an MCP client and snapshot names/schemas.
9. **Docs**: Publish the matrix and a one-line "what to call when" guide.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-mcp-surface-parity/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-mcp-surface-parity/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not expose dangerous mutations without explicit confirmation.
- Do not remove existing continuity tools.
- Do not change Sila store semantics.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should the adapter expose one generic `sila_command` tool or one tool per command?
- Which mutations require user confirmation every time?
