# Plan: Cross-Platform Operating System Test Matrix in CI

## Execution Steps
- [ ] Phase 1: Matrix Configuration: Update `.github/workflows/ci.yml` with a runner matrix for `ubuntu-latest`, `macos-latest`, and `windows-latest`.
- [ ] Phase 2: OS-Specific Test Strategy: Run `go test -race ./...` on Linux and standard unit tests `go test ./...` on macOS and Windows runners.
- [ ] Phase 3: Gate Verification: Verify workflow syntax and run local validation to ensure actions succeed deterministically.
- [ ] Phase 4: Attribution Commit: Commit with `(goals/cross-platform-ci-matrix/goal.md)` using a `ci:`, `chore:`, or `feat:` subject, then run `sila goals`.

