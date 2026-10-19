# Task: Retire subscription capacity replay simulation (retire-workload-replay)

## Chain of Custody (Isnad)
- **Registered By**: @antigravity
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: Pending Epistemic Mystery Audit

## Context
- Goal: Remove `thermal replay` command, `--against` and `--compare` CLI flags, and associated simulation packages in `internal/thermal/replay.go` and `internal/render/replay.go`.
- Why: Upstream AI subscription quotas (ChatGPT Plus/Pro, Claude Pro, Antigravity Pro) and rate cards change unpredictably. Hardcoding rolling-window heuristics inside a local CLI creates constant maintenance drag, inaccurate `DEGRADED`/`FAIL` capacity verdicts, and code bloat.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: `retire-web-server` (Seq #96)
- Readiness: Blocked on `retire-web-server`

## Files to Touch
- cmd/thermal/main.go
- internal/render/help.go
- docs/MIGRATION.md

## Files to Archive / Remove
- internal/thermal/replay.go
- internal/thermal/replay_test.go
- internal/render/replay.go
- internal/render/replay_test.go

## Requirements
1. Remove `runReplayReport` function and `"replay"` positional case in `cmd/thermal/main.go`.
2. Remove `--against` and `--compare` flags from `cmd/thermal/main.go`.
3. When invoked as `thermal replay`, print a migration message on stderr directing users to standard period reports (`thermal monthly`, `thermal models`) and pricing lookups, and exit with code 1.
4. Remove `replay` from `FinOps & Analytics` section in `internal/render/help.go`.
5. Remove `replay.go` and `replay_test.go` from `internal/thermal/` and `internal/render/`.
6. Add a 6-part migration entry in `docs/MIGRATION.md`.

## Constraints and non-goals
- Do NOT touch `internal/pricing/` models catalog or pricing cache.
- Do NOT alter any period reports (`daily`, `weekly`, `monthly`, `models`).

## Example
Input:
`thermal replay`
Expected stderr output:
`thermal: subscription replay simulation has been retired.`
`  For monthly volume and model pricing: thermal monthly, thermal models`
Expected exit code: 1

## Acceptance Criteria
- AC-001: `go test -race -v ./internal/thermal ./internal/render`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts test suites pass without replay files.
- AC-002: `/tmp/thermal-test replay`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 1; asserts retired command produces migration message on stderr (negative case).

## Return
Return JSON: `{"status": "ready_for_review", "summary": "...", "changed_files": [...], "case_results": {"AC-001": true, "AC-002": true}}`
