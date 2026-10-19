# Task: Retire token yield efficiency metric and diff parsing (retire-token-yield)

## Chain of Custody (Isnad)
- **Registered By**: @antigravity
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: Pending Epistemic Mystery Audit

## Context
- Goal: Remove `thermal yield` command, git diff line parsing, and token yield efficiency badge calculations in `internal/thermal/yield.go` and `internal/render/yield.go`.
- Why: Net lines of code added is a misleading proxy for AI code efficiency (it penalizes refactoring, debugging, and deletions) and couples git diff parsing to token telemetry.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: `retire-work-receipts` (Seq #98)
- Readiness: Blocked on `retire-work-receipts`

## Files to Touch
- cmd/thermal/main.go
- internal/render/help.go
- docs/MIGRATION.md

## Files to Archive / Remove
- internal/thermal/yield.go
- internal/thermal/yield_test.go
- internal/render/yield.go
- internal/render/yield_test.go

## Requirements
1. Remove `runYieldReport` function and `"yield"` positional case in `cmd/thermal/main.go`.
2. When invoked as `thermal yield`, print a migration message on stderr directing users to `thermal stats` and `thermal live` for token telemetry, and exit with code 1.
3. Remove `yield` from `FinOps & Analytics` section in `internal/render/help.go`.
4. Remove `yield.go` and `yield_test.go` from `internal/thermal/` and `internal/render/`.
5. Add a 6-part migration entry in `docs/MIGRATION.md`.

## Constraints and non-goals
- Do NOT alter git-root project attribution in `internal/thermal/project.go`.
- Do NOT touch `thermal stats` or `thermal live`.

## Example
Input:
`thermal yield`
Expected stderr output:
`thermal: token yield efficiency reporting has been retired.`
`  For token burn rates and velocity: thermal live, thermal stats`
Expected exit code: 1

## Acceptance Criteria
- AC-001: `go test -race -v ./internal/thermal ./internal/render`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts test suites pass without yield files.
- AC-002: `/tmp/thermal-test yield`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 1; asserts retired command produces migration message on stderr (negative case).

## Return
Return JSON: `{"status": "ready_for_review", "summary": "...", "changed_files": [...], "case_results": {"AC-001": true, "AC-002": true}}`
