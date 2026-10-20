# Plan: Make statusline and share-card scope explicit and URLs usable

## Execution Steps
- [x] **Task 1 — Label statusline scope**: Add `all tools`/`today`/`top` labels and tests for global/per-tool. Acceptance: statusline tests at plain/nerd show unambiguous scope; values unchanged.
- [x] **Task 2 — Canonical share snapshot**: Add scope/window/activeDays/cost fields; derive text/JSON from one snapshot. Acceptance: share JSON has no null displayed fields; activeDays/volume/spend share one scope.
- [x] **Task 3 — URL display policy**: Add shortened default display and full URL via JSON/explicit flag; test length/truncation. Acceptance: default output is usable; full token recoverable; docs state limits.
- [x] **Task 4 — Responsive/golden tests**: Test statusline/share at 80/120 and no-color. Acceptance: no wrapped values/broken card; golden snapshots pass.
- [x] **Task 5 — Docs and attribution commit**: Update migration/README/project.jsonc; run full gate; commit with the goal token. Acceptance: gate clean; commit ends in `(goals/statusline-and-share-scope-contract/goal.md)`.

## Dependencies and Handoff
The reconciliation goal owns canonical share totals; the JSON goal owns the share envelope; this goal owns statusline/share presentation and URL usability. Do not change streak/encoding math.

## Verification
```bash
go test -race ./internal/render ./internal/share ./cmd/thermal -run 'Statusline|Share|Scope|URL'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
