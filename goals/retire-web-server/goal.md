# Task: Retire embedded localhost web dashboard and serve API (retire-web-server)

## Chain of Custody (Isnad)
- **Registered By**: @antigravity
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: Pending Epistemic Mystery Audit

## Context
- Goal: Remove `thermal web` / `thermal serve` commands, associated CLI flags (`--port`, `--host`, `--open`, `--no-open`), and package `internal/server/`.
- Why: Thermal is a terminal-first utility. Embedding an HTTP server, Server-Sent Events hub, and a 1,685-line HTML/JS bundle directly conflicts with the anti-goals in GOVERNANCE.md ("Bloated Web/GUI Frameworks: Thermal is a terminal-native tool"). Users who want web visualizations or dashboards can consume the `--json` / `--stream` output via external tools or static site generators.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: `retire-dashboard-tui` (Seq #95)
- Readiness: Blocked on `retire-dashboard-tui`

## Files to Touch
- cmd/thermal/main.go
- internal/render/help.go
- docs/MIGRATION.md

## Files to Archive / Remove
- internal/server/server.go
- internal/server/server_test.go
- internal/server/assets/index.html
- internal/server/assets/

## Requirements
1. Remove `runServe` function and the `"web"` / `"serve"` positional cases in `cmd/thermal/main.go`.
2. Remove `--port`, `--host`, `--open`, and `--no-open` flags from `cmd/thermal/main.go`.
3. When invoked as `thermal web` or `thermal serve`, print a migration notice on stderr directing users to `thermal live --json --stream` for headless telemetry integration, and exit with code 1.
4. Remove `web` and `serve` from `internal/render/help.go`.
5. Remove `internal/server/` package entirely from the build.
6. Add a 6-part migration entry in `docs/MIGRATION.md`.

## Constraints and non-goals
- Do NOT touch `internal/thermal/live.go` or NDJSON streaming in `thermal live --json --stream`.
- Do NOT affect static report commands.

## Example
Input:
`thermal web`
Expected stderr output:
`thermal: embedded web server has been retired to preserve terminal-native focus.`
`  For headless telemetry and streaming: thermal live --json --stream`
Expected exit code: 1

## Acceptance Criteria
- AC-001: `./scripts/check.sh -q`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts binary compiles cleanly without `internal/server` package.
- AC-002: `/tmp/thermal-test web`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 1; asserts retired command produces migration message on stderr (negative case).

## Return
Return JSON: `{"status": "ready_for_review", "summary": "...", "changed_files": [...], "case_results": {"AC-001": true, "AC-002": true}}`
