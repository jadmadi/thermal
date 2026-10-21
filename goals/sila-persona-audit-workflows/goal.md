# Goal: Sila Persona Audit Workflows

**Priority**: P1
**Program**: Sila Utilisation Program
**Implementation surface**: Sila persona/audit CLI and Command Code docs

## Goal Description
Define when Command Code should use Sila personas, hires, reports, and audits, and when it should not. The goal adds selection rules, prompt-only/register-goals workflows, and report capture so persona work is evidence rather than ceremony.

## Problem and Evidence
- `sila persona` exposes a 31-persona registry; no AGENTS trigger mentions it.
- `sila audit` has web, ui, squad, and gap modes; no rule says which mode fits which task.
- A code implementation task did not need a persona audit, but a release/promise audit might.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-command-code-adapter, sila-mcp-surface-parity
- **Sequence**: 305
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `docs/SILA-COMMAND-CODE.md`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `AGENTS.md`
- `goals/sila-persona-audit-workflows/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Inventory**: List persona groups, specialist IDs, audit modes, and report storage.
2. **Selection matrix**: Map task types to persona/audit mode, or to "no audit".
3. **Prompt-only first**: Default to `--prompt-only` or dry-run before executing a multi-persona audit.
4. **Scope input**: Require a URL, file, or explicit context; reject empty audits.
5. **Reports**: Capture report IDs, verdicts, and linked goals; do not leave reports unread.
6. **Goal registration**: Use `--register-goals` only after human review and deduplication.
7. **Cost limits**: Bound pages, personas, tokens, and runtime.
8. **Tests**: Exercise selection, empty-input rejection, prompt-only, report capture, and duplicate goal prevention.
9. **Docs**: Add a "when not to use" section to prevent audit theater.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-persona-audit-workflows/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-persona-audit-workflows/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not auto-run audits on every task.
- Do not let an audit create goals without review.
- Do not change persona definitions.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Which persona is the default release auditor for Thermal?
- Should `sila audit gap` be part of every roadmap review?
