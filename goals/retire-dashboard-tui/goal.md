# Task: Retire interactive dashboard TUI and focus on thermal live (retire-dashboard-tui)

## Chain of Custody (Isnad)
- **Registered By**: @antigravity
- **Spec Audited By**: Pending Spec Audit
- **Implemented By**: Pending Implementation
- **Mystery Audited By**: Pending Epistemic Mystery Audit

## Context
- Goal: Remove the historical multi-tab terminal dashboard (`thermal dashboard`) and its 7,417 lines in `internal/tui/`, retaining `thermal live` as the sole interactive terminal monitor.
- Why: Every tab in `thermal dashboard` (Overview, Projects, Mix, Models, Stats) duplicates an existing fast, scriptable static CLI command (`thermal`, `projects`, `mix`, `models`, `stats`). Maintaining a parallel 5-tab TUI adds enormous complexity without unique utility, whereas `thermal live` provides real-time streaming velocity and sparkline metrics that static commands cannot replicate.
- Repository/cwd: `/mnt/Jad/github/lab/thermal-streak`
- Prerequisite Task IDs: None (Independent, Seq #95)
- Readiness: Ready to claim

## Files to Touch
- cmd/thermal/main.go
- internal/render/help.go
- docs/MIGRATION.md

## Files to Archive / Remove
- internal/tui/app.go
- internal/tui/app_test.go
- internal/tui/dense.go
- internal/tui/dense_test.go
- internal/tui/view_overview.go
- internal/tui/view_projects.go
- internal/tui/view_mix.go
- internal/tui/view_models.go
- internal/tui/view_stats.go
- internal/tui/data.go
- internal/tui/data_mix.go
- internal/tui/data_projects.go
- internal/tui/data_stats.go
- internal/tui/keys.go
- internal/tui/keys_test.go
- internal/tui/projects_test.go
- internal/tui/mix_test.go
- internal/tui/stats_test.go
- internal/tui/viewport_test.go
- internal/tui/realdataproj_test.go
- internal/tui/realdatamix_test.go
- internal/tui/realdata_test.go
- internal/tui/viewprobe_test.go

## Files Preserved Intact
- internal/tui/live.go (Live monitor Bubble Tea model, view, update, ticks)
- internal/tui/live_test.go (Live monitor test suite)
- internal/tui/theme.go (Shared color palettes needed by live monitor)

## Requirements
1. Remove `runDashboard` function and the `"dashboard"` positional case in `cmd/thermal/main.go`.
2. When invoked as `thermal dashboard`, output a clear migration notice on stderr directing users to `thermal live` for interactive monitoring or static commands (`thermal`, `thermal projects`, `thermal stats`) for historical queries, and exit with code 1.
3. Remove `dashboard` from `Dashboards & Views` in `internal/render/help.go`.
4. Archive or remove all historical dashboard TUI view, data adapter, and dense grid files from `internal/tui/`.
5. Ensure `internal/tui/live.go` compiles cleanly and all `TestLiveTUI_` test cases pass with `-race`.
6. Add a 6-part migration entry in `docs/MIGRATION.md` documenting the removal and redirection.

## Constraints and non-goals
- Do NOT touch `internal/tui/live.go` or `internal/tui/live_test.go`. Thermal live remains the premier terminal experience.
- Do NOT alter any static report command (`thermal`, `projects`, `models`, `mix`, `stats`).
- Maintain sub-10ms cached execution ceiling across all static commands.

## Example
Input:
`thermal dashboard`
Expected stderr output:
`thermal: the interactive dashboard has been retired in favor of 'thermal live' and static report commands.`
`  Interactive monitoring: thermal live`
`  Historical breakdown:   thermal, thermal projects, thermal stats`
Expected exit code: 1

## Acceptance Criteria
- AC-001: `go test -race -v ./internal/tui`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts `internal/tui` compiles cleanly without historical dashboard files and all live monitor tests pass.
- AC-002: `go test -race ./cmd/thermal`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 0; asserts main CLI compiles without `runDashboard`.
- AC-003: `/tmp/thermal-test dashboard`, cwd `/mnt/Jad/github/lab/thermal-streak`, expected exit 1; asserts retired command produces migration message on stderr (negative case).

## Return
Return JSON: `{"status": "ready_for_review", "summary": "...", "changed_files": [...], "case_results": {"AC-001": true, "AC-002": true, "AC-003": true}}`
