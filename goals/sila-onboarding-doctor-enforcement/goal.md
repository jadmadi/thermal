# Goal: Sila Onboarding Doctor Enforcement

**Priority**: P0
**Program**: Sila Utilisation Program
**Implementation surface**: Sila heal/onboard/doctor and Thermal CI

## Goal Description
Make Sila onboarding health fail-closed. `sila onboard check --strict` currently reports 11 pass and 1 warn while the agent still proceeds; the warning on Agent rules must become a failure for this repo, and doctor/agents/hook checks must run in CI.

## Problem and Evidence
- `sila heal --dry-run` still reports pending AGENTS wiring.
- `sila onboard check --strict` exits successfully with `Agent rules: warn`.
- `sila agents check` passes on hash sync without checking protocol completeness.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: sila-command-code-adapter, sila-session-lifecycle-hooks, sila-fact-ledger-integrity
- **Sequence**: 309
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `AGENTS.md`
- `.githooks/`
- `.github/workflows/ci.yml`
- `scripts/check.sh`
- `docs/SILA-UTILISATION-PROGRAM.md`
- `goals/sila-onboarding-doctor-enforcement/evidence.md`

## References
- [Facts and evidence](facts.md)
- [Execution plan](plan.md)
- [Program index](../../docs/SILA-UTILISATION-PROGRAM.md)

## Requirements
1. **Heal**: Run `sila heal` idempotently and commit the managed AGENTS change.
2. **Strict warn**: Make the Agent rules warning a hard failure in strict mode.
3. **Coverage check**: Extend agents check to required sections/verbs, not only hash.
4. **Doctor**: Run `sila doctor` and `sila hook status`; fail on unhealthy store or missing hooks.
5. **CI**: Add a Sila health job to the Thermal CI/pre-push path.
6. **Preflight integration**: Preflight refuses to proceed when heal is pending or onboarding is strict-red.
7. **Remediation**: Provide exact repair commands for every failure.
8. **Tests**: Simulate clean, stale, missing, corrupt, and permission-denied states.
9. **Docs**: Document the health matrix and repair runbook.

## Done Condition
1. Every task in `plan.md` is checked.
2. The implementation is committed with `(goals/sila-onboarding-doctor-enforcement/goal.md)` under a `feat`, `fix`, `refactor`, `perf`, or `chore` subject.
3. An evidence file exists with a standalone `Verdict: PASS` and command transcripts that still reproduce.
4. `sila facts verify` passes and `sila goals --json` reports this goal as verified with `unverified_count: 0`.
5. No product behavior outside this goal changed, and no bot co-author trailer was introduced.

## Evidence Contract
- `goals/sila-onboarding-doctor-enforcement/evidence.md` contains: baseline commands and outputs, changed files, positive tests, negative tests, conformance transcript, limitations, and a final standalone `Verdict: PASS`.
- The verdict is written only after every task is checked; a partial run may record evidence but must not contain a PASS verdict.
- `sila goals certify` or `sila goals audit` must consume the evidence file, and `sila facts verify` must confirm every referenced file still exists.
- No secret, prompt body, token, or private path may be written into evidence.

## Scope Boundary
- Do not mask warnings as passes.
- Do not auto-repair destructive issues without a backup.
- Do not make CI depend on a developer home directory.

## Handoff Contract
- Record a `sila memo` for each durable invariant created here.
- Record a `sila case` for each new architecture seam and a `sila lesson` for each non-trivial failure.
- Leave a `sila handoff` with open items when the goal is complete.
- Do not claim the goal until the implementation surface is ready; do not close it with only documentation.

## Open Questions
- Should strict mode be the default for all repos?
- Can Sila doctor run without network?
