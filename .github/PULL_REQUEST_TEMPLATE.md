## Description

<!-- Briefly describe what this pull request does, why it is needed, and any context. -->

Fixes / Closes: #<!-- Issue number or goal slug -->

## Type of Change

- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New tool loader (adds support for a new AI coding assistant)
- [ ] Performance improvement or allocation reduction
- [ ] Documentation update or typo fix
- [ ] Other (please describe):

## Contributor & Quality Checklist

Before submitting, please ensure you have completed the following:

- [ ] **CLA / DCO Sign-off**: Every commit in this PR has a `Signed-off-by: Name <email>` trailer (`git commit -s`), accepting the Thermal Contributor License Agreement documented in `CONTRIBUTING.md`.
- [ ] **Tests Passing**: Verified locally via `go test -v -race ./...`.
- [ ] **Simulated User Gate**: Verified locally via `./scripts/simulated_user_gate.sh`.
- [ ] **Code Formatting**: Code is formatted cleanly with `gofmt`.
- [ ] **Loader Invariants (if touching `internal/loaders/`)**:
  - [ ] SQLite connections open in strictly read-only mode (`?mode=ro`).
  - [ ] SQLite connections enable memory mapping (`PRAGMA mmap_size=268435456`).
  - [ ] Concurrency is bounded via worker pools / semaphores (no unbounded goroutines).
  - [ ] Token types are disjoint (`Input`, `Output`, `Reasoning`, `CacheRead`, `CacheWrite`).
  - [ ] Models pass through `modelName()` canonicalization.
  - [ ] Unit tests added in `internal/loaders/<tool>_test.go` with mock data.
- [ ] **Governance & Scope**:
  - [ ] Change adheres to the architectural scope and anti-goals in `GOVERNANCE.md` (no cloud telemetry, no background daemons, offline-capable).
  - [ ] Breaking changes have a pre-approved RFC issue.
