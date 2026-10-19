# Task: Retire transcript session command scraping and receipts (retire-work-receipts)

## Chain of Custody (Isnad)
- **Registered By**: @antigravity
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: Pending Epistemic Mystery Audit

## Context
- Goal: Remove `thermal receipt` command, transcript command scraping, and confidence classification engines in `internal/thermal/receipt.go` and `internal/render/receipt.go`.
- Why: Scraping shell commands inside AI tool call blocks across fifteen different tool formats is fragile, slow, and duplicates standard test runners and CI gates without providing verifiable token accounting.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: `retire-workload-replay` (Seq #97)
- Readiness: Blocked on `retire-workload-replay`

## Files to Touch
- cmd/thermal/main.go
- internal/render/help.go
- docs/MIGRATION.md

## Files to Archive / Remove
- internal/thermal/receipt.go
- internal/thermal/receipt_test.go
- internal/render/receipt.go
- internal/render/receipt_test.go

## Requirements
1. Remove `runReceiptReport` function and `"receipt"` positional case in `cmd/thermal/main.go`.
2. When invoked as `thermal receipt`, print a migration message on stderr directing users to native test runners and CI verification, and exit with code 1.
3. Remove `receipt` from `FinOps & Analytics` section in `internal/render/help.go`.
4. Remove `receipt.go` and `receipt_test.go` from `internal/thermal/` and `internal/render/`.
5. Add a 6-part migration entry in `docs/MIGRATION.md`.

## Constraints and non-goals
- Do NOT touch transcript reading in `internal/loaders/` needed for token and turn attribution.
- Preserve all session metrics and tool attribution.

## Example
Input:
`thermal receipt`
Expected stderr output:
`thermal: transcript command receipt auditing has been retired.`
`  Verify test outcomes using native test runners: go test, pytest, cargo test`
Expected exit code: 1

## Acceptance Criteria
- AC-001: `go test -race -v ./internal/thermal ./internal/render`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts test suites pass without receipt files.
- AC-002: `/tmp/thermal-test receipt`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 1; asserts retired command produces migration message on stderr (negative case).

## Return
Return JSON: `{"status": "ready_for_review", "summary": "...", "changed_files": [...], "case_results": {"AC-001": true, "AC-002": true}}`
