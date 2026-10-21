# Sila Utilisation Program

**Status**: planning only. No implementation is claimed.

## Governing thought

100% Sila utilisation does not mean running every Sila command in every session. It means every required Sila moment has a trigger, an invocation, an outcome, and evidence, and every optional surface is available without ceremony. The current gap is a Command Code integration gap, not a Sila capability gap.

## Definition of 100% utilisation

A session reaches 100% when all of the following are true:

1. **Integrity**: `sila facts verify` is green, or the failure is quarantined with a recorded recovery decision.
2. **Lifecycle**: session-start runs heal/prime/goals-ready, and session-end runs handoff/sync.
3. **Preflight**: no implementation edit occurs before a fresh preflight artifact exists.
4. **Memory**: recall is consulted before guessing; memo, lesson, and case are recorded at their trigger moments; facts verify runs before completion.
5. **Goals**: one claim, one session, dependency-correct sequencing, complete plan tasks, exact goal attribution, and zero unverified delivered goals.
6. **Review**: persona/audit use is selected only when a task needs it, and reports are read and linked.
7. **Skills**: repeated workflows are captured as project skills with measured cost and rollback.
8. **Health**: heal, agents check, onboard check --strict, doctor, and hook status are green.
9. **Feedback**: feedback is triaged, linked, and resolved; public sharing is explicit.
10. **Scorecard**: expected versus actual trigger coverage is recorded, with skip reasons and evidence.
11. **Non-ceremony**: a skipped optional verb with a reason is better than an empty invocation.

## Goal graph

| Seq | Goal | Surface | Depends on |
|---:|---|---|---|
| 299 | `sila-fact-ledger-integrity` | Sila knowledge store | none |
| 300 | `sila-command-code-adapter` | Sila core plus AGENTS/MCP | none |
| 301 | `sila-session-lifecycle-hooks` | Command Code + Sila hook bridge | adapter |
| 302 | `sila-mandatory-preflight-gate` | preflight and `/goal` | adapter, hooks, fact ledger |
| 303 | `sila-mcp-surface-parity` | Sila MCP | adapter |
| 304 | `sila-nine-verb-memory-protocol` | AGENTS, skills, evidence | preflight, MCP |
| 305 | `sila-persona-audit-workflows` | personas/audits | adapter, MCP |
| 306 | `sila-project-skill-authoring` | skills | adapter, MCP, personas |
| 307 | `sila-goal-evidence-certification` | goals and evidence | preflight, memory |
| 308 | `sila-goal-claim-lifecycle-hygiene` | claims and attribution | preflight, certification |
| 309 | `sila-onboarding-doctor-enforcement` | health and CI | adapter, hooks, fact ledger |
| 310 | `sila-utilization-scorecard` | session measurement | memory, personas, skills, evidence, claims, health |
| 311 | `sila-feedback-continuous-improvement` | feedback loop | scorecard |
| 312 | `sila-command-code-e2e-conformance` | hermetic conformance | all of the above |

## Execution waves

1. **Integrity and foundation**: fact ledger integrity, adapter, and MCP parity.
2. **Enforcement**: lifecycle hooks, mandatory preflight, onboarding/doctor.
3. **Practice**: nine verbs, personas/audits, project skills.
4. **Proof**: goal evidence/certification and claim hygiene.
5. **Measurement and closure**: scorecard, feedback, end-to-end conformance.

## Coverage matrix

| Sila family | Goal owner | Required trigger |
|---|---|---|
| Fact ledger integrity | fact ledger integrity | every verification |
| Continuity: salam, heal, prime, handoff, sync | adapter, hooks, preflight | session start/end |
| Knowledge: recall, doc, memo, lesson, case, facts | nine-verb protocol | debugging, integration, decisions, fixes, completion |
| Roadmap: goals, claims, evidence, certification | evidence, claim hygiene | every goal |
| Personas, audits, reports | persona/audit workflows | scoped review only |
| Skill authoring | project skills | repeated workflow |
| Quality gates: gate, doctor, onboard | onboarding/doctor | preflight and CI |
| Feedback | feedback loop | issue/tip/idea |
| Measurement | scorecard | every session |

## Baseline evidence

- `sila salam`: heal pending; no ready goals at the start of the gap analysis.
- `sila heal --dry-run`: AGENTS.md would be updated.
- `sila onboard check --strict`: 11 pass, 1 warn on Agent rules.
- `sila agents check`: hash-sync pass, not full protocol coverage.
- `sila skills list`: no project skills.
- `sila persona` exposes a 31-persona registry; `sila audit` exposes web, ui, squad, and gap.
- `sila feedback` contains `fb_heal-contract-protocol`.
- `sila facts verify` reports 539 active hash-chain violations across 25,110 blocks and exits 1; the first divergent block is 12 (`fact_6a93ccbe`), and the full transcript belongs in the fact-ledger goal evidence.
- 114 goals, 518/518 tasks, 4 verified, 22 stale, 88 unverified.

## Non-goals

- Do not run every Sila command as a checkbox.
- Do not create audit or persona theater.
- Do not weaken test or evidence gates to declare success.
- Do not make public feedback or remote pushes without explicit consent.
