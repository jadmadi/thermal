# Plan: Make help's tool registry and flag reference exhaustive and accurate

## Execution Steps
- [x] **Task 1 — Establish the tool source of truth**: Expose a registry list of tool display names/aliases and assert the help list equals it. Acceptance: `TestHelpToolRegistry` lists 15 tools including Zed; adding a fake registry entry fails until help updates.
- [x] **Task 2 — Fix flag reference**: Add per-command sort values, `--format` policy, `--start-of-week <day>`, and all value placeholders; align with `validateReportFlags`. Acceptance: a table-driven help test covers every flag and allowed value.
- [x] **Task 3 — Reflect chart applicability**: After the chart goal, update the `--chart` text to the final behavior; add a test that the help text matches the validation path. Acceptance: `--chart` help and behavior agree.
- [x] **Task 4 — Golden/width tests**: Add 80/120 golden help tests and retired-word assertions. Acceptance: `go test -race ./internal/render -run TestHelp` passes; no retired words.
- [x] **Task 5 — Docs and attribution commit**: Update README/project.jsonc, add a migration entry if needed, run full gate, commit with the goal token. Acceptance: gate clean; commit ends in `(goals/help-supported-tools-and-flags/goal.md)`.

## Dependencies and Handoff
The chart goal owns `--chart` behavior; this goal only documents the final result. The formatting goal owns the legend, not the help command list.

## Verification
```bash
go test -race ./internal/render -run 'Help|Tool|Flag'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
/tmp/thermal-test --help | grep -E 'zed|--start-of-week|--sort'
./scripts/simulated_user_gate.sh
```
