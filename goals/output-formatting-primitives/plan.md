# Plan: Unify output formatting semantics across every Thermal surface

## Execution Steps
- [x] **Task 1 — Define the shared helpers and tests**: Create the shared cost/window/tool/percent helpers with table-driven tests for recorded, estimated, unpriced, partial, sub-cent, zero, active-window, calendar-window, and tool-list cases. Acceptance: `go test -race ./internal/render -run 'TestFormat'` passes; no surface is migrated yet.
- [x] **Task 2 — Reserve the tilde and migrate trend/audit**: Replace `~` on non-money approximations in trend and audit; add golden assertions. Acceptance: `thermal trend --no-color` contains `≈`/`Mean` but not `~97.5M`; `thermal audit --no-color` contains `approx.` but not `~45.4K`.
- [x] **Task 3 — Migrate cost-bearing surfaces**: Move leaderboard/daily/projects/models/mix/stats/share/live cost rendering onto the shared formatter; make mixed totals explicit. Acceptance: each surface's render tests pass with recorded/estimated/unpriced fixtures, and no local duplicate cost formatter remains.
- [x] **Task 4 — Migrate windows and tool lists**: Apply active-window/calendar-window/project-day/model-day labels and the ` · ` tool joiner. Acceptance: daily default says `all time (N active days)`, `--last 7` says `last 7 days`, trend says `calendar days`, projects/models use `project-days`/`model-days`, and tool lists use ` · `.
- [x] **Task 5 — Update docs and metadata**: Add the six-part migration entry, README legend, and project metadata update. Acceptance: `git diff --check` passes; the migration entry has search pattern, detection signal, rationale, before/after, fix, and escape hatch.
- [x] **Task 6 — Run the full gate and commit**: Run focused tests, full race suite, build, and simulated gate; stage only this goal's files; commit with a feat/fix/refactor/perf/chore subject ending in `(goals/output-formatting-primitives/goal.md)`; run `sila goals`; do not push. Acceptance: gate output is clean and `git show --stat` lists only the allowed files.

## Dependencies and Handoff
The reconciliation goal consumes the cost/window helpers; the JSON goal consumes the same precision policy; the chart/live/audit/share goals consume the tool/cost/legend helpers. Do not change aggregation in this package.

## Verification
```bash
go test -race ./internal/render ./internal/thermal -run 'Format|Cost|Window|Tool|Percent'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
