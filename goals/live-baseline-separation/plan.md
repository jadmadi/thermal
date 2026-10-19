# Plan: Separate live baseline carry-in from real-time activity and session counters

## Execution Steps
- [x] **Task 1 — Add the carried-in contract**: Extend `LiveSnapshot` with `CarriedInTokens`, `CarriedInTurns`, `CarriedInCost`, and `CarriedInAt`; extend `LiveEvent` with `IsBaseline`. Populate them in the first-poll branch from the canonical today totals. Do not add them to session or rolling state.
  Acceptance and verification: `go test -race -run TestLiveTracker_CarriedInContract ./internal/thermal`; a 1M-token first poll yields carried-in 1M, session 0, burn 0, flame 0, and zero new events.

- [x] **Task 2 — Remove session seeding**: Delete the `seedToday` field, `SetSeedToday`, and the first-poll session/flame/event seeding block. Update `NewLiveModel` and `cmd/thermal/main.go` call sites. Keep `--fresh` accepted for live compatibility but make it a no-op for session semantics; update the live help text and migration entry.
  Acceptance and verification: `go test -race -run TestLiveTracker_LaunchRelativeSession ./internal/thermal`; first poll 1M, second poll +25K -> session 25K, Today 1.025M, carried-in 1M, exactly one +25K event. `thermal live --json --fresh` still emits a valid snapshot with session 0.

- [x] **Task 3 — Add carried-in and empty-session rendering**: Add a `Carried in:` line below the KPI grid when any carried-in value is non-zero, and make the new-since-launch portion visible through the launch-relative session panel or an annotation. Render session zero as `Since launch: no new activity`. Suppress duplicate Today/Session panels when all three displayed values match, and adjust the fixed-line/height budget so the line survives at 80x24.
  Acceptance and verification: `go test -race -run 'TestLiveModel_CarriedIn|TestLiveModel_NoDuplicateSession|TestLiveModel_CompactViewport' ./internal/tui`; 80x24 and 120x40 views contain `Carried in:`, show the new-vs-carried split, do not contain a synthetic `+531.5M` row, and do not contain two identical Today/Session value sets.

- [x] **Task 4 — Update JSON and stream**: Ensure `thermal live --json` exposes carried-in fields and zero session counters at startup. Ensure `--json --stream` emits the initial snapshot and no baseline event line. Add an `isBaseline` marker for any baseline display row.
  Acceptance and verification: `go test -race -run TestLiveJSON_CarriedIn ./cmd/thermal`; parse the first stream line and assert `carriedInTokens > 0`, `sessionTokens == 0`, and no baseline `event` line.

- [x] **Task 5 — Midnight and reset regressions**: On local-day rollover, recompute carried-in from the new day's totals (normally zero), keep session continuous, and never replay the previous day. `ResetSession` must not recreate the baseline or warm the flame.
  Acceptance and verification: `go test -race -run 'TestLiveTracker_MidnightCarriedIn|TestLiveTracker_ResetKeepsCarriedIn' ./internal/thermal`.

- [x] **Task 6 — Migration and project metadata**: Add a six-part `docs/MIGRATION.md` entry covering the changed `--fresh` live behavior and new JSON fields. Update `project.jsonc` (`lastEdit`, feature wording) without changing unrelated version fields.
  Acceptance and verification: inspect the migration entry for search pattern, detection signal, rationale, before/after, fix, and escape hatch; `git diff --check`.

- [x] **Task 7 — Full verification and attribution commit**: Run the focused tests, full race suite, CLI build, and simulated user gate. Stage only this goal's files, inspect `git diff --cached --name-only`, and create an impact-first Conventional Commit ending in `(goals/live-baseline-separation/goal.md)`. Run `sila goals` and do not push.

## Dependencies and Handoff
No new-goal prerequisites. The closed `live-metric-parity` contract must be preserved: token types stay disjoint, activity-only counts never become tokens, and cache-hit math stays `CacheRead/(Input+CacheRead+CacheWrite)`. Downstream goals consume `CarriedIn*` and `IsBaseline`.

## Verification
```bash
go test -race ./internal/thermal ./internal/tui ./cmd/thermal -run 'Live|Baseline|Sim'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
