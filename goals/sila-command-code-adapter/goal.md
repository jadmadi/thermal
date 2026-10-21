# Goal: Sila Command Code Adapter

**Priority**: P0
**Program**: Sila Utilisation Program
**Implementation surface**: Sila core CLI plus thermal AGENTS/mcp wiring

## Goal Description
Create a first-class Command Code adapter for Sila. The adapter must write the full working contract to AGENTS.md, register MCP, install the Command Code lifecycle hooks, and make the integration verifiable. It replaces the current split where the short contract lives in AGENTS.md and the full nine-verb protocol lives only in Cursor rules.

## Problem and Evidence
- `sila heal --dry-run` reports `Provision sila wiring in this repository` and lists `AGENTS.md` as affected.
- `sila install --help` supports `claude`, `cursor`, and `devin`; `command-code` is absent.
- `sila agents check` passes on a hash of the short managed block, so it cannot detect that the nine-verb MUST protocol is missing from AGENTS.md.
- Command Code loads AGENTS.md and MCP, not `.cursor/rules`; therefore the strongest Sila instructions never reach this agent.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 300
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `AGENTS.md`
- `.commandcode/mcp.json`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `docs/SILA-COMMAND-CODE.md`
- `scripts/sila-preflight.sh`
- `project.jsonc`
- `goals/sila-command-code-adapter/evidence.md`
- `Sila upstream CLI repository: companion changes under cli/ in the external Sila checkout`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Adapter target**: Add `command-code` to install/agents-sync targets, or add a documented adapter implementation with the same behavior.
2. **Full protocol**: Write both the seven-step working loop and the nine-verb memory protocol into the managed AGENTS.md block, with explicit MUST language.
3. **Managed boundaries**: Preserve user-authored AGENTS.md content and make writes idempotent through `sila:begin` and `sila:end` markers.
4. **MCP registration**: Install or validate the Sila MCP entry in `.commandcode/mcp.json` without overwriting unrelated servers.
5. **Lifecycle wiring hook**: Provide the exact Command Code session-start/session-end hook configuration, and document whether it lives in repo or user config.
6. **Coverage validation**: Extend `sila agents check` to compare required sections and verbs, not only a content hash.
7. **Dry run and rollback**: Make `sila heal --dry-run`, `sila install --dry-run`, and rollback show the complete diff and restore the prior AGENTS.md.
8. **Harness detection**: Treat Command Code as a first-class harness in `sila sources` and installation output.
9. **Hermetic tests**: Test against temporary HOME and repo roots; never mutate the real user config in tests.
10. **Documentation**: Add a Command Code section to the Sila docs and a Thermal runbook that names the adapter version and repair command.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-command-code-adapter/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-command-code-adapter/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not implement the lifecycle behavior itself; that belongs to `sila-session-lifecycle-hooks`.
- Do not change Thermal CLI output or aggregation.
- Do not add a second source of truth for the protocol; AGENTS.md is canonical for Command Code.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should the adapter live upstream in Sila core or as a repo-local generator?
- Should AGENTS.md contain the full persona/audit/skill trigger table, or link to a generated appendix?
- What is the compatibility contract for older Command Code versions?
