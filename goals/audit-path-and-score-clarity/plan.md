# Plan: Make audit findings addressable and score understandable

## Execution Steps
- [x] **Task 1 — Add path/scope to findings**: Extend audit finding structs/JSON and render a path/scope column; de-duplicate basenames. Acceptance: duplicate `AGENTS.md`/`CLAUDE.md` fixture renders two distinct paths.
- [x] **Task 2 — Align readiness rows and long names**: Use the shared status/table formatter; add truncation tests. Acceptance: goldens at 80/120 show aligned `PASS`/`WARN`/`CRIT` and no border breaks.
- [x] **Task 3 — Explain score and advice**: Render score components/biggest win and map every advice bullet to a path. Acceptance: fixture contains a component line and no advice without a path.
- [x] **Task 4 — JSON parity and docs**: Add path/scope/advice fields to JSON, update README/project.jsonc/migration if needed. Acceptance: text/JSON path sets match; docs have six parts if keys change.
- [x] **Task 5 — Full gate and attribution commit**: Run focused/full tests, build, gate; commit with the goal token. Acceptance: gate clean; commit ends in `(goals/audit-path-and-score-clarity/goal.md)`.

## Dependencies and Handoff
The formatting goal owns status/table/cost helpers; this goal only adds audit-specific path/score content. The gallery goal captures the corrected audit.

## Verification
```bash
go test -race ./internal/audit ./internal/render -run 'Audit|Path|Score'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
