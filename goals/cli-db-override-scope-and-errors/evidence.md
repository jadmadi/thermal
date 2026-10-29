# Evidence: cli-db-override-scope-and-errors

Date: 2026-10-29
Project: thermal-streak
Goal: cli-db-override-scope-and-errors

## Implementation Summary
- `--db` now requires a single tool. Aggregate contexts (`daily`, `projects`, `live`, `statusline`, etc. with `--tool all/auto` or no tool) exit `2` before any data read.
- Added `internal/loaders/pathcheck.go`: `ToolPathShape`, `ExpectedDBFile`, `ValidateDBPath`, `DBPathError`, and `translateDBError`.
- SQLite-file tools accept a database file or a directory containing the expected filename; directory tools accept a directory; Muse accepts either a session-index file or a directory.
- Validation resolves symlinks, checks existence/readability/regular-file/empty-file, and verifies the `SQLite format 3\0` header for file-shaped tools.
- `LoadToolData` validates a user override before loading and wraps raw loader errors in `DBPathError`; the raw driver error remains available through `Unwrap` and is printed only with `--verbose`.
- Added a config-override warning when a configured directory lacks its expected database file; it is printed with `--verbose`.
- Updated help text, README, migration guide, project metadata, and the simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./...` → PASS.
- `./scripts/check.sh --pre-commit` → PASS (7s; gofmt, package tests, binary build).
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (104/104 checks).
- `thermal --db /nonexistent daily` → exit `2`, `--db applies to one tool; add --tool <name>`; stdout empty.
- `thermal --tool opencode --db /nonexistent daily` → exit `1`, `--db path does not exist: /nonexistent`.
- `thermal --tool opencode --db /tmp daily` → exit `1`, `/tmp is a directory; OpenCode expects opencode.db inside it`; no `disk I/O error`.
- `thermal --tool opencode --db /etc/passwd daily` → exit `1`, `/etc/passwd is not a SQLite database for OpenCode`; no `(26)`.
- `thermal --tool opencode --db /etc/passwd daily --verbose` → adds `thermal: verbose: missing SQLite header`.
- `thermal --tool claude --db /etc/passwd daily` → exit `1`, `is a file; Claude expects a directory`.
- `thermal --tool claude --db /tmp daily` → exit `0` (empty directory is a valid zero-data source).
- `thermal --tool all --db /tmp daily` → exit `2` scope error.
- `thermal --db /tmp opencode daily` → exit `1` directory message (positional tool accepted).
- `thermal --tool opencode --db <real opencode.db> daily` → exit `0`.

## Acceptance Criteria

- AC-001: `thermal --db /nonexistent daily` exits `2`, stdout empty, stderr contains `--db applies to one tool`, and no `Thermal · daily usage` header. **PASS**
- AC-002: `thermal --tool opencode --db /nonexistent daily` exits `1`, contains `--db path does not exist`, and prints no report. **PASS**
- AC-003: `thermal --tool opencode --db /tmp daily` exits `1`, contains `is a directory` and `opencode.db`, and does not contain `disk I/O error`. **PASS**
- AC-004: `thermal --tool opencode --db /etc/passwd daily` exits `1`, contains `is not a SQLite database`, and does not contain `(26)`. **PASS**
- AC-005: `TestValidateDBPathSQLiteFile`, `TestValidateDBPathDirectoryWithDB`, and `TestLoadToolDataDirectoryOverrideUsesExpectedFile` pass; a real OpenCode database path exits `0` in the CLI matrix. **PASS**
- AC-006: `thermal --tool claude --db /tmp daily` exits `0`; `thermal --tool claude --db /etc/passwd daily` exits `1` with `expects a directory`; tested by `TestValidateDBPathDirectoryTool` and `TestValidateDBPathFileForDirectoryTool`. **PASS**
- AC-007: `thermal --tool opencode --db /tmp daily` exits `1` with the expected filename, not a panic; `TestValidateDBPathDirectoryForFileTool` covers the same path. **PASS**
- AC-008: `--verbose` adds `thermal: verbose:` with the raw driver/header detail; without `--verbose`, the raw text is absent. **PASS**
- AC-009: `TestLoadToolDataNoFallback` proves the loader is not called with the default path after a bad `--db`; the CLI renders no report and exits `1`. **PASS**
- AC-010: `TestDBOverrideRequiresOneTool` rejects no-tool/all/auto contexts with exit `2`; `TestDBOverrideAllowsSingleTool` accepts positional and `--tool` forms. **PASS**
- AC-011: `go test -race ./...` passes, including `TestToolPathShape`, `TestValidateDBPath*`, `TestTranslateDBError`, `TestLoadToolDataNoFallback`, `TestLoadToolDataTranslatesLoaderError`, `TestLoadToolDataDirectoryOverrideUsesExpectedFile`, `TestConfigOverrideMissingExpectedFileWarns`, and `TestDBOverride*`. **PASS**
- AC-012: `./scripts/simulated_user_gate.sh` passes 104/104 and includes the `--db` scope/path section. **PASS**

## Regression Notes
- Default paths and read-only SQLite opening are unchanged.
- `thermal --tool claude --db <empty dir>` remains a valid zero-data source.
- The config-override fallback is now documented with a `--verbose` warning rather than silent.
- `--db` is still rejected for aggregate views by design; omit it or add `--tool`.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-db-override-scope-and-errors",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Scope --db to one tool and validate path shapes", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_c369dd3e", "intent": "Scope --db to one tool and validate path shapes"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
