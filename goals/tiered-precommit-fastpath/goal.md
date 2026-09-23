# Goal: Tiered Pre-Commit Fast-Path: Scoped Formatting, Diff Sentry & Package Tests

## Goal Description
Introduce a high-performance, tiered gate verification architecture in Thermal:
1. **Fast-Path Pre-Commit Mode (`--pre-commit` / `--quick` / `-q`)**:
   - Executes under 3 seconds for single-package and docs-only changes.
   - Scopes code formatting (`gofmt -l`) strictly to modified Go files, preventing noise from pre-existing repository drift in accordance with AGENTS.md.
   - Runs **Diff Sentry**: audits the active diff for leaked secrets / private keys, unresolved merge conflict markers, forbidden debug statements / temporary panics, and accidental large binary additions.
   - Runs unit tests **ONLY** for modified Go packages through git diff package isolation. Bypasses tests entirely for docs-only / assets-only diffs.
   - Skips heavy operations (full 10-package suite, statement coverage matrix, Go report card, and 87-case simulated user release gate) to preserve local commit momentum.
2. **Comprehensive Pre-Push / Pipeline Gate Mode (`--full` / `--pre-push`)**:
   - Executes full verification: CLA / DCO sign-off, formatting, diff sentry, Go report card & static analysis (`go vet ./...`), full race detector across all packages (`go test -race ./...`), full per-package statement coverage table (`go test -cover ./...`), binary compilation, simulated user release gate, and vulnerability scanning.
3. **Automated Git Hook Integration**:
   - Provides ready-to-use hooks (`.githooks/pre-commit` and `.githooks/pre-push`).
   - Supports automated installation via `./scripts/check.sh --install-hooks`.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 1
- **Shape**: ship
- **Tier**: roadmap

## References
- **Shared Understanding & Fact Sheet**: [`goals/tiered-precommit-fastpath/facts.md`](facts.md)
- **Execution Plan**: [`goals/tiered-precommit-fastpath/plan.md`](plan.md)

## Done Condition
1. `./scripts/check.sh --pre-commit` executes in under 3 seconds on clean or single-package changes.
2. Scoped formatting checks only modified Go files and offers actionable remediation.
3. Diff Sentry detects and halts on secret tokens, merge conflict markers, debug prints, and binary additions.
4. Git diff package isolation identifies changed packages and tests only those; skips tests on docs-only changes.
5. Full package coverage table and Go report card / simulated user gate remain cleanly reserved for `--pre-push` / `--full`.
6. `.githooks/pre-commit` and `.githooks/pre-push` are functional and installable via `--install-hooks`.
7. Comprehensive automated tests in `scripts/check_test.sh` verify all fast-path, sentry, and isolation behaviors.
8. Zero architectural regressions; all existing tests pass with `-race`.
