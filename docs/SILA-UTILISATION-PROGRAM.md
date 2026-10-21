# Sila Utilisation Program

**Status**: planning only. No implementation is claimed.

## Scope correction

The first draft registered fourteen goals in `thermal-streak`. That was the wrong registry boundary. Most of those goals describe Sila core behavior and already have richer goals in the Sila CLI project (`sila-cli`, `/mnt/Jad/github/lab/sila/cli`). They were removed from the Thermal registry so this repository tracks only Thermal adoption work.

**Sila CLI owns** the adapter, MCP surface, trigger contract, persona/audit routing, skill provisioning, fact ledger, goal engine, onboarding/doctor, scorecard, and feedback mechanics.

**Thermal owns** the consumer integration: mandatory preflight, Command Code host hooks, Thermal-specific project skills, a Thermal scorecard, and a consumer conformance harness.

## Governing thought

100% Sila utilisation does not mean running every Sila command in every session. It means every required Sila moment has a trigger, an invocation, an outcome, and evidence, and every optional surface is available without ceremony. The core capability gap is upstream in Sila CLI; the Thermal gap is enforcement and adoption.

## Definition of 100% utilisation

A Thermal session reaches 100% when all of the following are true:

1. **Integrity**: upstream `facts-ledger-integrity-campaign` is green, or its failure is quarantined with a recorded recovery decision.
2. **Lifecycle**: session-start runs heal/prime/goals-ready, and session-end runs handoff/sync.
3. **Preflight**: no implementation edit occurs before a fresh Thermal preflight artifact exists.
4. **Memory**: recall is consulted before guessing; memo, lesson, and case are recorded at their trigger moments; facts verify runs before completion.
5. **Goals**: one claim, one session, dependency-correct sequencing, complete plan tasks, exact goal attribution, and zero unverified delivered goals.
6. **Review**: persona/audit use is selected only when a task needs it, and reports are read and linked.
7. **Skills**: repeated Thermal workflows are captured as project skills with measured cost and rollback.
8. **Health**: upstream heal/agents/onboard/doctor gates are green and the Thermal preflight refuses to proceed when they are not.
9. **Feedback**: feedback is triaged, linked, and resolved; public sharing is explicit.
10. **Scorecard**: expected versus actual trigger coverage is recorded, with skip reasons and evidence.
11. **Non-ceremony**: a skipped optional verb with a reason is better than an empty invocation.

## Thermal adoption goals (registered here)

| Seq | Goal | Surface | Depends on |
|---:|---|---|---|
| 300 | `sila-mandatory-preflight-gate` | Thermal preflight and `/goal` enforcement | none |
| 301 | `sila-session-lifecycle-hooks` | Command Code host lifecycle config | none |
| 302 | `sila-project-skill-authoring` | Thermal-specific project skills | none |
| 303 | `sila-utilization-scorecard` | Thermal session measurement | preflight, hooks, skills |
| 304 | `sila-command-code-e2e-conformance` | Thermal consumer conformance | preflight, hooks, skills, scorecard |

These five goals must not change Sila core CLI, MCP, store, hook-bridge, persona, skill-governor, goal-engine, or fact-ledger code. They consume upstream releases.

## Upstream Sila CLI ownership

These goals already exist in `sila-cli`; they are not duplicated here.

| Upstream goal | Owns |
|---|---|
| `facts-ledger-integrity-campaign` | Fact-chain audit, classification, and safe recovery |
| `harness-session-memory-bridge` | Core session-start/session-end hook bridge |
| `active-lifecycle-hook-bridge` | Harness lifecycle adapters |
| `mcp-continuity-9verb-and-tier0-parity` | MCP continuity surface and tier-0 parity |
| `mcp-specialist-and-temporal-tool-parity` | Persona, standup, domain-learning, and blast-radius MCP tools |
| `trigger-contract-full-surface-routing` | Nine-verb and specialist command routing |
| `prime-briefing-cases-and-capability-discovery` | Prime briefing cases and capability discovery |
| `workflow-router-full-surface-taxonomy` | Workflow router coverage |
| `doctor-mcp-and-trigger-utilization-gate` | Doctor/MCP utilisation diagnostics |
| `global-and-workspace-sila-skill-provisioning` | Global and workspace Sila skill provisioning |
| `memory-protocol-goal-contract` | Memory protocol in goal templates and certification |
| `memory-verb-cli-consistency` | Memo/lesson/case/note CLI consistency |
| `onboard-check-contract` | Onboarding strictness and repair contract |
| `project-onboarding` | Harness onboarding and supported hooks |
| `agents-sync-atomic-project-writes` | Atomic rule generation and size ceilings |
| `agents-check-revision-visibility` | Rule drift and revision visibility |
| `cross-agent-benchmark-scorecard` | Cross-agent empirical scorecard |

**Known upstream gap**: there is no exact `command-code-adapter` goal in `sila-cli` yet. If the adapter remains unowned, register it in `sila-cli`, not in Thermal.

## Execution order

1. **Upstream first**: repair `facts-ledger-integrity-campaign`, then `harness-session-memory-bridge`, `mcp-continuity-9verb-and-tier0-parity`, `trigger-contract-full-surface-routing`, `doctor-mcp-and-trigger-utilization-gate`, and `global-and-workspace-sila-skill-provisioning`.
2. **Thermal enforcement**: `sila-mandatory-preflight-gate` and `sila-session-lifecycle-hooks`.
3. **Thermal practice**: `sila-project-skill-authoring`.
4. **Thermal proof**: `sila-utilization-scorecard` and `sila-command-code-e2e-conformance`.

## Baseline evidence

- Sila CLI: **361 goals**, 327 implemented and verified, 7 ready, 32 pending; the CLI working tree already has modified/untracked goals for trigger, MCP, doctor, skill, and fact-ledger work.
- Thermal after scope correction: **119 goals**, 5 pending, 3 ready, 2 blocked, 518/563 tasks completed.
- `sila heal --dry-run`: AGENTS.md wiring still pending.
- `sila onboard check --strict`: 11 pass, 1 warn on Agent rules.
- `sila agents check`: hash-sync pass, not full protocol coverage.
- `sila skills list`: no project skills.
- `sila persona` exposes a 31-persona registry; `sila audit` exposes web, ui, squad, and gap.
- `sila facts verify`: 539 violations across 25,110 blocks, exit 1; first divergence at block 12 (`fact_6a93ccbe`).
- Feedback: `fb_heal-contract-protocol` (adapter/preflight gap), `fb_verify-integrity-store` (ledger), and a correction tip.

## Non-goals

- Do not duplicate Sila CLI goals in the Thermal registry.
- Do not run every Sila command as a checkbox.
- Do not create audit or persona theater.
- Do not weaken evidence or certification gates.
- Do not make public feedback or remote pushes without explicit consent.
