# Facts: Contributor Onboarding Path, Preflight Gate & Retention Loop

## Architectural Invariants & Constraints

1. **Rubric & Assessment Standards (`oss-contributor-onboarding`)**:
   - Evaluated against the 4 blocking checks (B1: Invited, B2: Setup runs cold, B3: Work available, B4: Submissions get answered) and 8 scored checks (S1: Out-of-scope, S2: Pinned tools, S3: Local gates match CI, S4: Useful fork CI, S5: Complete first issues, S6: Defined claiming, S7: Pre-announced legal gate, S8: Credit & next step).
   - Baseline scan of Thermal scored **FAIL** (2/4 blocking passed, 6/16 scored points). The primary blockers are the absence of an open issue stock (0 open issues on GitHub), absence of a published response SLA, absence of unified local preflight parity with CI, and missing out-of-scope/claiming documentation.

2. **Harmonization with Ready, Unimplemented Goals (Zero Contradictions)**:
   - **`governance-foundations` (Sequence #1)**:
     - Defines authoritative *Scope & Anti-Goals* (zero-alloc, local-first, read-only vs no web GUIs, no telemetry, no daemons) and *Decision Classes* (routine fixes 72h lazy consensus, 7d loaders/RFCs).
     - *Harmonization*: `CONTRIBUTING.md` must link to `GOVERNANCE.md`'s Scope & Anti-Goals rather than inventing distinct definitions. The response target must be set to 72 hours for PR reviews (aligning with the 72h lazy consensus window and Jad's 2–4 hours/week maintainer budget).
   - **`repo-alignment-and-security` (Sequence #2)**:
     - Owns `.github/CODEOWNERS`, `SECURITY.md`, `MAINTAINERS.md`, `.github/PULL_REQUEST_TEMPLATE.md`, and issue templates (`bug_report.yml`, `loader_proposal.yml`).
     - *Harmonization*: Contributor onboarding will NOT duplicate or modify PR/issue templates or security policies. It leverages the PR template's pre-flight checklist and points contributors to the structured loader proposal form.
   - **`community-covenant-neutrality` (Sequence #4)**:
     - Owns the Thermal Community Reciprocity Covenant in `DUAL-LICENSE.md` and `CONTRIBUTING.md`.
     - *Harmonization*: Onboarding additions in `CONTRIBUTING.md` (claiming rules, SLA, local preflight, out-of-scope summary) reside in dedicated sections that do not displace or alter the CLA or Covenant sections.
   - **`trust-ladder-and-succession` (Sequence #3)**:
     - Owns the staged trust ladder (Contributor -> Reviewer -> Maintainer -> Key Holder).
     - *Harmonization*: Contributor onboarding feeds the top of the funnel (first and second merged PR), preparing contributors for promotion along the trust ladder.

3. **Local Preflight Gate Parity (`scripts/check.sh`)**:
   - A contributor who cannot reproduce a red CI check locally is blocked by definition.
   - CI (`.github/workflows/ci.yml`) enforces:
     1. `Signed-off-by` trailer verification on PR commits.
     2. `gofmt -l .` clean formatting.
     3. `go vet ./...` clean static analysis.
     4. `go test -v -race ./...` passing unit tests and race detection.
     5. `go build -v ./cmd/thermal` compilation check.
     6. `./scripts/simulated_user_gate.sh` 46-step end-to-end simulated user release gate.
   - `scripts/check.sh` must run all of these locally in order, failing fast with actionable guidance.

4. **Good First Issues Seed Stock (7-Part Anatomy)**:
   - Per Steinmacher et al. and Kent C. Dodds, a good first issue is a recruiting vehicle requiring:
     1. Observed behavior.
     2. Expected behavior.
     3. Exact file location and test path.
     4. Required technical skills/APIs.
     5. Suggested approach / implementation outline.
     6. Verification command (`./scripts/check.sh`).
     7. Named mentor (`@jadmadi`).
   - 3 initial seed candidates:
     - *Candidate 1*: Loader for Cline / Roo Code JSON transcripts.
     - *Candidate 2*: Benchmark fixture suite for Muse & Grok loaders.
     - *Candidate 3*: Step-by-step developer guide for implementing and registering new AI tool loaders.

5. **Review Playbook (Two-Clock Rule & Retention)**:
   - Clock 1: Acknowledgment within 24 hours ("Thanks for opening this! Review scheduled by <date>").
   - Clock 2: Technical review within 72 hours.
   - Decline Fast & Kindly: Cite scope rules, thank contributor, provide alternative issues.
   - Post-Merge Retention: Credit in release notes, invite to specific next issue within 24h of merge.

## File & Interface Contracts

- `scripts/check.sh`: Executable preflight script returning code 0 only when all CI-parity checks pass.
- `CONTRIBUTING.md`: Add sections for Out-of-Scope, Response Targets, Claiming Conventions, and Local Preflight.
- `docs/GOOD_FIRST_ISSUES.md`: Ready-to-file templates for the first 3 starter issues.
- `docs/CONTRIBUTOR_PLAYBOOK.md`: Maintainer review runbook containing the Two-Clock rule, reply templates, and retention nudges.
- `README.md`: Link to contributing guide and local check script.
