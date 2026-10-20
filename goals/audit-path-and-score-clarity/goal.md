# Goal: Make audit findings addressable and score understandable

## Goal Description
Priority: P2 (actionability). The audit table lists `CLAUDE.md` twice (204 lines/~6.8K tok and 1 line/~2 tok) and `AGENTS.md` twice (204 and 494 lines) with no path or scope, so a user cannot tell which file to fix. The readiness row `PASS Tool Databases` has different spacing from other PASS rows. The score `58/100 · Grade D` has no explanation of what drives it or how to improve it.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: output-formatting-primitives
- **Sequence**: 207
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/audit/audit.go`
- `internal/audit/audit_test.go`
- `internal/render/audit.go`
- `internal/render/audit_test.go`
- `cmd/thermal/main.go`
- `docs/MIGRATION.md`
- `README.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)


## Requirements
1. **Addressable rule files**: Add a `path` or `scope` field to every rule-file finding and render it. Duplicate basenames must be distinguishable (e.g., repo-root `AGENTS.md` vs `~/.claude/CLAUDE.md`). The table should show `AGENTS.md (repo root)` / `CLAUDE.md (~/.claude)` or a separate Path/Scope column.
2. **Actionable advice mapping**: Every advice bullet must name the same path/scope as the finding it references; no two `AGENTS.md` warnings may collapse into one ambiguous bullet. If the audit scans multiple roots, include the root in each recommendation.
3. **Readiness alignment**: Use the shared table/status formatter for `PASS`/`WARN`/`CRIT`; fix `PASS Tool Databases` spacing and any wrapped continuation line so the box borders and columns align. Keep plain-text `--no-color` output.
4. **Score/grade explanation**: Add a short `Score: 58/100 (Grade D)` explanation: the score components (rule tax, MCP tax, readiness checks, penalties) and the largest contributor(s). Include a one-line `biggest win` recommendation mapped to a path. Do not change the score formula unless needed; if changed, document it.
5. **Long MCP names**: Ensure long MCP server names do not break the border; truncate with an ellipsis and keep the client/scope column. Add a test with a name longer than the column.
6. **JSON parity**: If audit JSON already contains findings, add `path`/`scope`/`advice` fields consistently and test that text and JSON reference the same paths. Do not leak file contents.
7. **Tests/docs**: Add a fixture with duplicate basenames, long MCP names, and a known score; assert paths, alignment, score explanation, and advice. Update README/project.jsonc if the audit schema changes; add a migration entry if JSON keys are added/renamed.

## Done Condition
1. Audit text lists distinct paths/scopes for duplicate `AGENTS.md`/`CLAUDE.md` findings and actionable advice matches them.
2. `PASS Tool Databases` and all readiness rows align with the same status/table formatter; long MCP names do not break borders.
3. Score/grade includes a component/contribution explanation and at least one mapped improvement; the score itself is stable or explicitly migrated.
4. JSON findings include path/scope/advice where applicable; text and JSON agree.
5. Focused audit/render tests, full race suite, build, and simulated gate pass; docs/metadata updated; the attribution commit ends in `(goals/audit-path-and-score-clarity/goal.md)`.

## Scope Boundary
Audit presentation/actionability only. Do not change what the audit checks, add new network calls, or alter read-only guarantees. Do not touch report totals/formatting beyond shared helpers.
