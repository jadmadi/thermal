# Plan: Tiered Pre-Commit Fast-Path: Scoped Formatting, Diff Sentry & Package Tests

## Execution Steps
- [x] **Phase 1**: Enhance `scripts/check.sh` with CLI flag parsing (`--pre-commit`, `--quick`, `-q`, `--full`, `--pre-push`, `-f`, `--install-hooks`, `--help`).
- [x] **Phase 2**: Implement scoped formatting (`gofmt -l` restricted to changed Go files) respecting `AGENTS.md` drift rule.
- [x] **Phase 3**: Implement Diff Sentry scanner checking for secrets, merge conflict markers, debug prints, and forbidden binary files.
- [x] **Phase 4**: Implement git diff package isolation for unit tests (determining modified Go packages, testing only those, bypassing on docs/assets diffs).
- [x] **Phase 5**: Implement full mode package coverage table (`go test -cover ./...` with formatted summary table and threshold check) and Go report card reserved for pre-push / CI.
- [x] **Phase 6**: Create `.githooks/pre-commit` and `.githooks/pre-push` with installer support.
- [x] **Phase 7**: Author `scripts/check_test.sh` to thoroughly test all modes, sentry detections, isolation logic, and hook installation.
- [x] **Phase 8**: Update `CONTRIBUTING.md` and `AGENTS.md` documentation, run full verification suite, and commit with goal reference.
