# Evidence: Turn every data-access failure into a named condition and a runnable next step

Date: 2026-10-30
Binary: `/tmp/thermal-g2` (source build)
Goal: `cli-data-access-error-guidance`

## Acceptance Matrix

| AC | Command | Result |
| --- | --- | --- |
| AC-001 | `thermal opencode weekly --chart` (no OpenCode) | exit 1; `OpenCode has no data at ~/...`; inventory, all-tools, `--db` lines |
| AC-002 | `thermal --tool codex` | exit 1; `Codex has no data at ~/.codex`; `thermal sources` hint |
| AC-003 | `thermal claude --db <regular-file>` | exit 1; `is a file; Claude expects a directory` |
| AC-004 | `thermal opencode --db <missing>` | exit 1; `--db path does not exist`; `check the path or omit --db` |
| AC-005 | `thermal claude --db <symlink-loop>` | exit 1; `too many levels of symbolic links`; `readlink -f` fix |
| AC-006 | `thermal claude --db <mode-000-dir>` | skipped: test process runs as root, permission bypass does not reproduce |
| AC-007 | `--verbose` | exactly one `thermal: verbose:` raw OS line; no stack trace |
| AC-008 | `--json` + missing tool | exit 1; stdout empty; human condition and next commands on stderr |
| AC-009 | valid `--db` with missing default | loads normally |
| AC-010 | help examples | `opencode weekly --chart` states the OpenCode prerequisite; all-tools example added |
| AC-011 | `go test ./internal/loaders/... ./cmd/thermal/...` | ok |
| AC-012 | `./scripts/simulated_user_gate.sh` | 161 passed, 0 failed |

## Observed Output

```text
$ thermal opencode weekly --chart
exit 1
thermal: OpenCode has no data at ~/.local/share/opencode/opencode.db
  inventory: thermal sources
  all tools: thermal weekly
  override:  thermal --tool opencode --db <path>
```

```text
$ thermal --tool codex
exit 1
thermal: Codex has no data at ~/.codex
  inventory: thermal sources
  all tools: thermal
  override:  thermal --tool codex --db <path>
```

```text
$ thermal claude --db <symlink-loop>
exit 1
thermal: too many levels of symbolic links at <path>
  resolve it with readlink -f <path> and pass the resolved path
```

```text
$ thermal claude --db <regular-file>
exit 1
thermal: <path> is a file; Claude expects a directory
  pass the data directory, not a single file
```

```text
$ thermal opencode --db <missing>
exit 1
thermal: --db path does not exist: <path>
  check the path or omit --db
```

```text
$ thermal opencode --verbose
exit 1; stdout empty
thermal: OpenCode has no data at ~/.local/share/opencode/opencode.db
  ...
thermal: verbose: stat <path>: no such file or directory
```

## Implementation Notes

- Extended `DBPathError` with `Condition`, `ToolID`, `DisplayPath`, and
  `Next`; `Error()` renders the next-step lines.
- Added `DataCondition`, `classifyStatError`, `MissingDefaultDataError`, and
  `EnrichDefaultDataError` in `internal/loaders/pathcheck.go`.
- Replaced the two bare `database not found` / `data directory not found`
  errors in `registry.go` with typed, enriched errors.
- `loadUsage` enriches the all-tools line with the requested report.
- `ValidateDBPath` now preserves precise explicit `--db` messages and
  classifies loops, permissions, and `ENOTDIR` separately.
- Help examples now state the OpenCode prerequisite and show the all-tools and
  `sources` alternatives.

## Verification Commands

```bash
go test ./internal/loaders/... ./cmd/thermal/...
./scripts/check.sh --pre-commit
./scripts/simulated_user_gate.sh
```

All passed. The gate reported `161 passed, 0 failed`.

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-data-access-error-guidance",
  "prime": {"ran": true},
  "recall": [
    {"query": "database not found", "matches": 5},
    {"query": "cannot read check permissions", "matches": 5}
  ],
  "memos": [{"id": "fact_e52a0d93"}],
  "lessons": [{"id": "fact_9ca4ad5b"}],
  "cases": [{"id": "fact_dd84d31c"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_dcd68ae6b817"}
}
```

Verdict: PASS
