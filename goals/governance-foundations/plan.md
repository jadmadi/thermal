# Plan: Written Solo Authority Governance, Scope Defense & Decision Rules

## Execution Tasks

1. [x] **Draft `GOVERNANCE.md` (Constitutional Document)**:
   - Section 1: Scope, Philosophy & Anti-Goals (performance invariants, local-first, zero daemons).
   - Section 2: Roles (Project Lead: Jad Madi `@jadmadi`, Maintainers, Contributors).
   - Section 3: Decision Classes & Mechanisms:
     - Routine fixes (72h lazy consensus, 1 approval).
     - New agent loaders (strict checklist against `AGENTS.md`).
     - Architectural & breaking changes (7-day RFC window).
     - License & dual-licensing amendments (Project Lead authority).
   - Section 4: Conflict Resolution & Justified Vetoes (technical rationale required, Lead tiebreaker).
   - Section 5: Maintainer Availability & Maintenance SLA (stated 2–4 hours/week).
2. [x] **Update `README.md` & Navigation**:
   - Add Governance badge (`Governance: Solo Authority`) pointing to `GOVERNANCE.md`.
   - Add Governance section in table of contents and documentation index.
3. [x] **Cross-Reference in `AGENTS.md`**:
   - Link architectural invariants to the formal scope boundaries in `GOVERNANCE.md`.
4. [x] **Verification & Validation**:
   - Verify all links in `README.md`, `GOVERNANCE.md`, and `AGENTS.md` resolve cleanly.
   - Run `./scripts/simulated_user_gate.sh` to guarantee no docs or build regressions.
- [x] **Phase 4 — Attribution Commit**: Commit with `(goals/governance-foundations/goal.md)` using a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` subject (docs/test commits are ignored), then run `sila goals`.
