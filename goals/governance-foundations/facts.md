# Facts: Written Solo Authority Governance, Scope Defense & Decision Rules

## Architectural Invariants & Constraints

1. **Governance Model: Written Solo Authority**:
   - Following Nadia Eghbal's *Working in Public* taxonomy, Thermal is in the **Stadium** quadrant (high usage/potential, contributor absence factor = 1).
   - Attempting an elected steering committee or multi-vendor council with an absence factor of 1 is an anti-pattern (CNCF reviews reject "governance theatre").
   - Model must be **Written Solo Authority + Staged Trust Ladder + Succession Plan**, maintaining low overhead (< 30 min/month) while establishing full decision transparency.

2. **Scope Defense & Anti-Goals**:
   - The primary purpose of `GOVERNANCE.md` is to defend maintainer attention and architectural integrity against scope creep.
   - Core invariants from `AGENTS.md` must be elevated to constitutional project rules:
     - **In-Scope**: Zero-allocation terminal heatmaps, local-first read-only SQLite/JSONL scanning (`PRAGMA mmap_size`), bounded worker pools, canonical model normalization, disjoint token attribution, accurate cost estimation.
     - **Anti-Goals / Out-of-Scope**: Cloud-hosted telemetry phoning home, mandatory background daemons, bloated GUI web frameworks, non-local network requirements for core scans, features sacrificing sub-10ms invocation time.

3. **Decision Classes & Response Timers**:
   - Every decision class must carry an explicit clock to eliminate the "unbounded thread" stall failure mode:
     - **Routine Changes & Bug Fixes**: 1 maintainer approval. Lazy consensus with a **72-hour** review window (aligned with Node.js `GOVERNANCE.md` standard).
     - **New Agent Loaders**: Must meet `AGENTS.md` requirements (read-only SQLite, canonical models, disjoint tokens, test fixtures). Evaluated within 7 days.
     - **Breaking Changes & Architectural RFCs**: Require a written proposal issue with a **7-day** public comment period.
     - **Deadlock Resolution**: Technical justifications required for all vetoes ("justified veto" rule from Apache ASF). If consensus fails, Project Lead (@jadmadi) serves as final tiebreaker.
     - **Weekly Maintenance Commitment**: Project Lead publishes expected availability (2–4 hours/week) so delays are recognized as capacity, not abandonment.

## File & Interface Contracts

- `GOVERNANCE.md`: Public governance contract defining values, scope, roles, decision classes, clocks, and escalation.
- `README.md`: Add badge and link to `GOVERNANCE.md` in project overview.
- `AGENTS.md`: Cross-reference `GOVERNANCE.md` for architectural decision-making and scope defense.

