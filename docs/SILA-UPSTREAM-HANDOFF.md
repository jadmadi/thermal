# Sila Work Handoff

All Sila core and harness-adoption work now belongs to the `sila-cli` registry at `/mnt/Jad/github/lab/sila/cli`, not this repository.

## Upstream owners

| Concern | Sila CLI goal |
|---|---|
| Command Code installation | `sila-install-command-code-target` |
| Command Code lifecycle hooks | `command-code-sila-lifecycle-hooks` |
| Command Code MCP bridge | `command-code-sila-mcp-bridge` |
| Core session bridge | `harness-session-memory-bridge` |
| Generic lifecycle adapters | `active-lifecycle-hook-bridge` |
| MCP continuity and nine verbs | `mcp-continuity-9verb-and-tier0-parity` |
| Trigger routing and specialist verbs | `trigger-contract-full-surface-routing` |
| Prime, cases, capability discovery | `prime-briefing-cases-and-capability-discovery` |
| Doctor and utilisation gate | `doctor-mcp-and-trigger-utilization-gate` |
| Persona, audit, and skill workflows | `sila-persona-audit-skill-workflow` |
| Global and workspace skills | `global-and-workspace-sila-skill-provisioning` |
| Plugin subsystem and MCP bridge | `sila-plugins-subsystem`, `sila-mcp-plugin-bridge` |
| Fact ledger integrity | `facts-ledger-integrity-campaign` |
| Goal memory and evidence | `memory-protocol-goal-contract` |
| Onboarding strictness | `onboard-check-contract` |
| Conformance vectors | `ccli-conformance-golden-vectors` |

The goals are general harness/plugin work, not Thermal-specific. The Sila CLI agent owns implementation. Thermal has no pending Sila goals.

## Thermal boundary

Thermal keeps only product work and its existing completed onboarding history. Do not recreate Sila core, adapter, hook, MCP, persona, skill, scorecard, or conformance goals here. If Thermal needs an exception, record it as a consumer integration note with a link to the upstream goal.

## Live baseline

- Sila CLI: 361 goals, 327 implemented and verified, 7 ready, 32 pending.
- Thermal after handoff: 114 goals, 0 pending, 518/518 tasks.
- `sila facts verify` remains red with 539 violations across 25,110 blocks; `facts-ledger-integrity-campaign` is the upstream owner.
