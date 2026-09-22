# Plan: Staged Trust Ladder, Succession Protocol & Absence Factor Monitoring

## Execution Tasks

1. [x] **Amend `GOVERNANCE.md` with Staged Trust Ladder**:
   - Detail the 5 stages: Contributor, Triager, Reviewer, Maintainer, Key Holder.
   - Enforce criteria: 3+ bug reproductions for Triager; 10+ substantive reviews for Reviewer; 3+ months sustained contributions for Maintainer.
   - Separate Key Holder rights: at least 6 months as Maintainer + 1 supervised release cycle before receiving release credentials.
2. [x] **Define Emeritus, Inactivity & Hiatus Invariants**:
   - 6-month inactivity window triggers automated notice and transition to Emeritus.
   - Document credential rotation SLA within 30 days of role change.
   - Define reinstatement mechanism (simple majority vote of remaining maintainers).
3. [x] **Draft Emergency Succession & Escrow Runbook in `GOVERNANCE.md`**:
   - Specify 90-day unreachable trigger for Project Lead.
   - Define interim stewardship procedure and archive fallback.
4. [x] **Develop Governance Metrics Check Script (`scripts/check_governance_metrics.sh`)**:
   - Calculate CHAOSS contributor absence factor across 12-month commit history.
   - Check credential inventory concentration and flag single-point-of-failure rows.
   - Provide summary diagnostics: current absence factor, target threshold (>= 2), and active maintainer count.
5. [x] **Integrate into Release Gate**:
   - Hook `scripts/check_governance_metrics.sh` into `./scripts/simulated_user_gate.sh`.
6. [x] **Verification & Validation**:
   - Execute `scripts/check_governance_metrics.sh` and verify accurate reporting.
   - Run `go test -race ./...` and `./scripts/simulated_user_gate.sh`.
- [x] **Phase 4 — Attribution Commit**: Commit with `(goals/trust-ladder-and-succession/goal.md)` using a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` subject (docs/test commits are ignored), then run `sila goals`.

