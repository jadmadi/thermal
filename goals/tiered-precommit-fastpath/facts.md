# Facts: Tiered Pre-Commit Fast-Path: Scoped Formatting, Diff Sentry & Package Tests

## Architectural Invariants & Constraints
- **Sub-3s Latency**: Fast-path pre-commit checks MUST complete in under 3 seconds on single-package or docs-only modifications.
- **Fail-Closed Security**: Diff sentry MUST fail closed on detected secrets, private keys, merge conflict markers, or temporary debug lines.
- **AGENTS.md Compliance**: `gofmt` MUST be scoped to touched files; never fail on unrelated pre-existing drift across historical test files.
- **Zero Push Compromise**: Full pre-push / CI verification MUST run full race detector (`go test -race ./...`), full package coverage (`go test -cover ./...`), static analysis (`go vet`), binary compilation, and simulated user release gate.
- **Hook Autonomy**: Git hooks must be optional, non-destructive, and easily toggled via `--install-hooks` or `core.hooksPath`.

## File & Interface Contracts
- `scripts/check.sh`: Main entrypoint supporting `--pre-commit`, `--quick`, `-q`, `--full`, `--pre-push`, `-f`, and `--install-hooks`.
- `.githooks/pre-commit`: Git hook delegating to `./scripts/check.sh --pre-commit`.
- `.githooks/pre-push`: Git hook delegating to `./scripts/check.sh --pre-push`.
- `scripts/check_test.sh`: Automated test suite exercising fast-path pre-commit, diff sentry, package isolation, docs bypass, and full coverage modes.
- `CONTRIBUTING.md` & `AGENTS.md`: Updated contributor guidelines documenting tiered gate usage.
