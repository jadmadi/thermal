# Facts: Staged Trust Ladder, Succession Protocol & Absence Factor Monitoring

## Architectural Invariants & Constraints

1. **Staged Trust Ladder (Post-xz Mitigations)**:
   - The 2024 `xz-utils` (CVE-2024-3094) incident demonstrated that granting commit rights and release signing keys simultaneously to relieve maintainer pressure is a critical vulnerability.
   - Thermal must enforce strict separation between code merge rights and release credentials:
     - **Stage 1 — Contributor**: Submits pull requests, adheres to CLA sign-off and testing gates.
     - **Stage 2 — Triager**: Granted issue triage, label assignment, and PR reproduction rights after 3+ verified bug reproductions.
     - **Stage 3 — Reviewer**: Granted review authority for specific areas (e.g. loaders or render) after 10+ substantive reviews.
     - **Stage 4 — Maintainer**: Granted merge rights to `main` after $\ge 3$ months of sustained contribution and unanimous current maintainer consensus.
     - **Stage 5 — Key Holder**: Granted GoReleaser publishing rights and domain administrative access **only** after $\ge 6$ months as Maintainer and completion of at least one supervised release cycle.

2. **Inactivity, Emeritus & Hiatus Policies**:
   - Inactivity defined: No merged PRs, reviews, or issue responses for 6 consecutive months.
   - Procedural Emeritus transition: Automatic move to Emeritus status (preserving credit in `MAINTAINERS.md`, removing merge/admin rights).
   - Reinstatement: Simple majority consensus or Project Lead confirmation.
   - Planned Hiatus: Maintainers can post a planned hiatus window without triggering emeritus de-provisioning.

3. **Emergency Succession & Credential Escrow**:
   - Thermal's absence factor is currently 1.
   - If Project Lead is incapacitated or unreachable for $>90$ days:
     - Pre-designated trusted contact / escrow protocol defined.
     - Public archive trigger: If no successor is appointed or available, repository enters structured read-only archive state rather than abandoned limbo.

4. **Automated Absence Factor & Concentration Tracking**:
   - Add `scripts/check_governance_metrics.sh` computing the CHAOSS contributor absence factor (smallest number of contributors responsible for 50% of contributions over the last 12 months) and checking for single-point-of-failure credential rows.
   - Embed governance metric reporting into `./scripts/simulated_user_gate.sh` under `--verbose` or diagnostic mode.

## File & Interface Contracts

- `GOVERNANCE.md`: Document Sections on Staged Trust Ladder, Emeritus Policy, Hiatus, and Emergency Succession.
- `MAINTAINERS.md`: Add Emeritus section and credential inventory table.
- `scripts/check_governance_metrics.sh`: Script calculating CHAOSS absence factor and credential concentration.
- `scripts/simulated_user_gate.sh`: Release gate verification for governance metrics sanity.

