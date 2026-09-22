# Plan: Contributor Onboarding Path, Preflight Gate & Retention Loop

## Execution Tasks

1. [x] **Build Unified Local Preflight Script (`scripts/check.sh`)**:
   - Create executable `scripts/check.sh` mirroring the full `.github/workflows/ci.yml` pipeline:
     1. Sign-off trailer verification against base branch.
     2. Code formatting verification (`gofmt -l .`).
     3. Go vet static analysis (`go vet ./...`).
     4. Unit test suite with race detector (`go test -v -race ./...`).
     5. Build verification (`go build -v ./cmd/thermal`).
     6. Simulated user testing gate (`./scripts/simulated_user_gate.sh`).
   - Add conditional `govulncheck` execution if installed on host.
   - Verify script returns exit code 0 when clean and non-zero with helpful remediation instructions on failure.

2. [x] **Stock 3 Good First Issues (`docs/GOOD_FIRST_ISSUES.md`)**:
   - Draft 3 starter issues strictly adhering to the 7-part anatomy:
     - Issue 1: *feat(loaders): Add Cline / Roo Code agent loader* (location: `internal/loaders/cline.go`, test: `internal/loaders/cline_test.go`, skills: Go JSON parsing, disjoint token mapping).
     - Issue 2: *test(loaders): Add benchmark fixture suite for Muse and Grok loaders* (location: `internal/loaders/bench_test.go`, skills: Go testing/benchmarking).
     - Issue 3: *docs(contributing): Guide for adding new agent loaders* (location: `docs/ADDING_A_LOADER.md`, skills: technical documentation, architecture understanding).
   - Each issue must include expected behavior, file pointers, testing commands, and named mentor (`@jadmadi`).

3. [x] **Enhance `CONTRIBUTING.md`**:
   - Add **Out of Scope & Anti-Goals** section referencing `GOVERNANCE.md` (no GUI frameworks, no cloud telemetry, no mutable SQLite operations, no fuzzy pricing).
   - Add **Finding Work & Issue Claiming** section:
     - State claiming convention: Comment on an issue to request assignment.
     - State stale claim timeout: Claims expire after 7 days without an open PR or status update.
     - Link to good first issues.
   - Add **Response Targets & Review Expectations**:
     - State the Two-Clock rule: Maintainer initial acknowledgment within 24–48 hours, deep technical review within 72 hours (aligned with 2–4 hours/week maintainer budget and 72h lazy consensus in `GOVERNANCE.md`).
   - Add **Local Pre-Flight Verification**:
     - Document `./scripts/check.sh` as the single required command before opening a pull request.

4. [x] **Codify Review Playbook & Contributor Retention (`docs/CONTRIBUTOR_PLAYBOOK.md`)**:
   - Document maintainer procedures:
     - Two-clock communication workflow.
     - Review order: Scope -> Correctness & Tests -> Architecture -> Machine-formattable style.
     - Decline fast and kindly response template.
     - Contributor credit at merge (release notes and author attribution).
     - Second-contribution invitation template (nudging the drive-by contributor to become a recurring contributor).

5. [x] **Cross-Reference in `README.md`**:
   - Update Contributing section in `README.md` to reference `scripts/check.sh` and the enhanced `CONTRIBUTING.md`.

6. [x] **Verification & Validation**:
   - Execute `./scripts/check.sh` locally to ensure exit code 0.
   - Run `./scripts/simulated_user_gate.sh` to verify zero regressions.
   - Audit the contributor path with `contributor-path-audit.sh` to confirm increased score and resolved blocking checks.

- [x] **Phase 4 — Attribution Commit**: Commit with `(goals/contributor-onboarding/goal.md)` using a `feat:`, `fix:`, `refactor:`, `perf:`, or `chore:` subject, then run `sila goals`.
