# Plan: Repository Alignment, Code Ownership, PR Lifecycle & Security Policy

## Execution Tasks

1. [x] **Implement `.github/CODEOWNERS`**:
   - Map root `*` to `@jadmadi`.
   - Explicitly map subsystem boundaries: `/cmd/`, `/internal/loaders/`, `/internal/thermal/`, `/internal/pricing/`, `/internal/render/`, `/build.sh`, `/.github/`.
2. [x] **Create `MAINTAINERS.md`**:
   - Document Project Lead: Jad Madi (`@jadmadi`, independent).
   - Define structure for active maintainers, affiliations, and emeritus members.
   - Cross-link to `GOVERNANCE.md` for maintainer progression criteria.
3. [x] **Author `SECURITY.md` (Coordinated Vulnerability Disclosure)**:
   - Define supported versions matrix (active release branch).
   - Specify private reporting email (`contact@jadmadi.net`) and GitHub Security Advisory route.
   - Document response timeline (48h triage, 7d evaluation, 30d patch SLA).
   - Define security boundaries and non-vulnerabilities (e.g. read-only local database inspection).
4. [x] **Add PR & Issue Lifecycle Templates**:
   - Create `.github/PULL_REQUEST_TEMPLATE.md` with CLA sign-off and testing release gate checklist.
   - Create `.github/ISSUE_TEMPLATE/bug_report.yml` structured form.
   - Create `.github/ISSUE_TEMPLATE/loader_proposal.yml` verifying SQLite/JSONL performance invariants.
5. [x] **Integrate Automated Vulnerability Scanning in CI**:
   - Add `golang.org/x/vuln/cmd/govulncheck@latest` step to `.github/workflows/ci.yml`.
6. [x] **Verification & Validation**:
   - Verify GitHub Actions syntax via dry-run / linter.
   - Run `go test -race ./...` and `./scripts/simulated_user_gate.sh`.
- [x] **Phase 4 — Attribution Commit**: Commit with `(goals/repo-alignment-and-security/goal.md)` using a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` subject (docs/test commits are ignored), then run `sila goals`.

