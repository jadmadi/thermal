# Facts: Repository Alignment, Code Ownership, PR Lifecycle & Security Policy

## Architectural Invariants & Constraints

1. **Alignment (Zero Contradictions)**:
   - Written roles must exactly match repository permissions.
   - Every file tree in Thermal (`cmd/`, `internal/loaders/`, `internal/thermal/`, `internal/pricing/`, `internal/render/`, `build.sh`, `.github/`) must be explicitly mapped in `.github/CODEOWNERS` to `@jadmadi`.
   - `MAINTAINERS.md` must list current active maintainers, their organizational affiliations (preventing unstated single-vendor control), contact channels, and an explicit Emeritus section.

2. **Security & Vulnerability Management (RFC 9116 & GitHub Security Advisory Standards)**:
   - Thermal processes local SQLite databases, JSON logs, and CLI parameters. Security boundaries focus on command injection, file path traversal during project key normalization, and untrusted JSONL line handling.
   - `SECURITY.md` must establish:
     - Supported versions policy (latest minor SemVer release, e.g. v0.12.x).
     - Private vulnerability reporting channel (`contact@jadmadi.net` or GitHub Private Vulnerability Reporting).
     - Response Timelines:
       - **48 hours**: Initial acknowledgment and triage.
       - **7 days**: Severity assessment and reproduction verification.
       - **30 days**: Patch development, validation, and coordinated CVE disclosure.
     - Embargo policy: No public issues or PRs for unpatched vulnerabilities.
   - Continuous Security Verification:
     - Add `govulncheck` to `.github/workflows/ci.yml` to automatically verify Go dependencies against the Go vulnerability database.

3. **PR Lifecycle & Issue Templates**:
   - `.github/PULL_REQUEST_TEMPLATE.md` must enforce:
     - [ ] Sign-off verification (`git commit -s` CLA requirement).
     - [ ] Automated gate compliance (`go test -race ./...` and `simulated_user_gate.sh`).
     - [ ] Loader adherence (read-only SQLite URI `?mode=ro`, `PRAGMA mmap_size`, canonical models).
   - `.github/ISSUE_TEMPLATE/` must provide structured YAML forms for Bug Reports and New Loader Proposals.

## File & Interface Contracts

- `.github/CODEOWNERS`: GitHub automated review assignment mapping all paths to `@jadmadi`.
- `MAINTAINERS.md`: Public roster of active and emeritus maintainers with affiliations.
- `SECURITY.md`: Vulnerability reporting process, response timelines, and security invariants.
- `.github/PULL_REQUEST_TEMPLATE.md`: Standard PR checklist.
- `.github/ISSUE_TEMPLATE/bug_report.yml`: Issue template for bugs.
- `.github/ISSUE_TEMPLATE/loader_proposal.yml`: Issue template for new agent loaders.
- `.github/workflows/ci.yml`: Add `govulncheck` step.

