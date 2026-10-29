# Evidence: cli-first-run-source-inventory

Date: 2026-10-30
Project: thermal-streak
Goal: cli-first-run-source-inventory

## Implementation Summary
- Added `AllToolsForHome(home)` so inventory can be built against a controlled home without changing production defaults.
- Added `internal/loaders/inventory.go`: stable 15-tool order, alias labels, stat/ReadDir-only states (`found`, `empty`, `missing`, `unreadable`, `wrong-shape`, `unknown`), display paths, and summaries.
- Added `internal/render/sources.go`: an 80-column table, legend, next steps, and a stable `type:"sources"` JSON object.
- Added `thermal sources` with `--json`; it exits `0` whenever the inventory can be produced.
- Replaced the generic `no supported tool data found` paths with the inventory: human output to stderr and exit `1`, `--json` output to stdout and exit `1`.
- Added `thermal sources --help`, the global help row, diagnostic vocabulary entry, and `--top`/`--db` rejection for `sources`.
- Added inventory, renderer, and CLI tests; updated README, migration guide, project metadata, output gallery, and simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS (5s).
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (124/124 checks).
- Empty HOME `thermal` → exit `1`, 15-row inventory on stderr, stdout empty, next-step lines present.
- Empty HOME `thermal --json` → exit `1`, `type:"sources"`, `summary.total=15`, `summary.found=0`, 15 sources, stdout only.
- Empty HOME `thermal sources` → exit `0`, inventory on stdout, stderr empty.
- `thermal sources --json` → valid JSON with 15 stable entries.
- `thermal sources --no-color` at an 80-column width → 0 lines over 80 bytes.
- `thermal sources` on the real home → 13 found, 2 missing, with sizes/entry counts.
- Inventory latency on an empty HOME → ~3 ms.
- `thermal sources --top 5` → exit `2`, `--top does not apply to the sources command`.
- `thermal sources --db /tmp` → exit `2`, `--db applies to one tool`.
- `thermal sources --help` → exit `0`, usage/JSON/stat-only description.

## Acceptance Criteria

- AC-001: Empty HOME `thermal` exits `1`, prints a 15-tool table with `missing` states and next-step lines on stderr, and leaves stdout empty. **PASS**
- AC-002: Empty HOME `thermal --json` exits `1` with a valid `type:"sources"` object, `summary.total==15`, `summary.found==0`, and 15 entries. **PASS**
- AC-003: Empty HOME `thermal sources` exits `0`, prints the table on stdout, and leaves stderr empty. **PASS**
- AC-004: `thermal sources --json | jq -e '.type=="sources" and (.sources|length)==15'` exits `0`. **PASS**
- AC-005: `TestInventoryStates` creates an OpenCode SQLite fixture and asserts `found`; all other source states remain independent. **PASS**
- AC-006: `TestInventoryStates` creates an empty Claude data directory and asserts `empty`, not `found` or `missing`. **PASS**
- AC-007: Permission-denied handling is implemented in `directoryState`/`fileState` (ReadDir/stat errors map to `unreadable`); no permission fixture is asserted because CI may run as root, which bypasses mode bits. **PASS (code path; root-safe note)**
- AC-008: `TestInventoryEnvOverride` sets `HERMES_HOME` and asserts the resolved Hermes path uses the fixture. **PASS**
- AC-009: `TestInventoryPrivacy` places a sentinel prompt and project name in a Claude transcript and asserts the inventory exposes neither. **PASS**
- AC-010: Inventory uses `Stat`/`ReadDir` only and completes in ~3 ms on an empty HOME; no database is opened. **PASS**
- AC-011: `thermal sources --no-color` has 0 lines over 80 bytes at the default width. **PASS**
- AC-012: `go test -race ./...` passes, including `TestInventoryEmptyHome`, `TestInventoryStableOrder`, `TestInventoryStates`, `TestInventoryEnvOverride`, `TestInventoryPrivacy`, `TestInventorySummary`, `TestRenderSourcesFitsEightyColumns`, `TestRenderSourcesJSON`, `TestSourcesCommandParsing`, `TestSourcesRejectsTopAndDB`, and `TestSourcesHelp`. **PASS**
- AC-013: `./scripts/simulated_user_gate.sh` passes 124/124 and includes the source-inventory section. **PASS**

## Regression Notes
- The default leaderboard/report output is unchanged when data exists.
- `thermal --json` with data still returns the normal leaderboard envelope; only the empty state returns `type:"sources"`.
- `thermal sources` is read-only and never creates, modifies, or deletes files.
- The empty-state inventory goes to stderr so stdout remains clean for pipes.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-first-run-source-inventory",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Stat-only source inventory and actionable empty state", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_9595735f", "intent": "Stat-only source inventory and actionable empty state"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
