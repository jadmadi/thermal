# Plan: Multi-Target Cross-Compilation Verification Gate in CI

## Execution Steps
- [ ] Phase 1: Workflow Step: Add a multi-target build verification step to `.github/workflows/ci.yml`.
- [ ] Phase 2: Compilation Matrix: Cross-compile `cmd/thermal` for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64` with `CGO_ENABLED=0 -trimpath`.
- [ ] Phase 3: Gate Validation: Verify zero errors or warnings during cross-compilation passes.
- [ ] Phase 4: Attribution Commit: Commit with `(goals/multi-target-cross-compile-ci-gate/goal.md)` using a `ci:` or `chore:` subject, then run `sila goals`.

