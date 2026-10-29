# Goal: Replace the empty-state dead end with a per-source inventory and next steps

## Goal Description
Priority: P1 (activation). On a machine with no supported agent data, the default command prints one line to stderr and exits `1`. The tool knows exactly which fifteen sources it probed, where each one lives, and whether it is missing, empty, or unreadable, but it tells the user none of that. The per-tool path is better (`database not found: /path`), which makes the aggregate message feel deliberately unhelpful.

Observed on the current source build (2026-10-29) with `HOME=$(mktemp -d)`:

- `thermal` → `thermal: no supported tool data found`, exit `1`; stdout empty.
- `thermal --json` → the same stderr line, stdout empty, exit `1`.
- `thermal --verbose` → no additional diagnostics.
- `thermal stats` and `thermal live --json` → the same generic line.
- `thermal statusline` → the same generic line, which is especially noisy in a shell prompt (see `cli-nontty-and-statusline-output-contract`).
- `thermal opencode` → `thermal: database not found: /tmp/…/.local/share/opencode/opencode.db` (specific, but still no remedy).
- `thermal --tool claude daily` → `thermal: data directory not found: /tmp/…/.claude`.
- `thermal audit` still works because it audits instruction files in the CWD, not agent data.

The outcome of this goal is a **source inventory**: a table of every supported tool with its canonical name, alias, expected file or directory, state (`found`, `empty`, `missing`, `unreadable`, `wrong-shape`), and a concrete next action. It appears in the empty-state path (stderr, exit `1` for compatibility) and as a first-class `thermal sources` command (stdout, exit `0`), with a stable JSON schema and a strict no-content privacy guarantee.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 247
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 2
- **Related goals**: `cli-db-override-scope-and-errors` should reuse `ToolPathShape` and the per-tool path metadata; `cli-nontty-and-statusline-output-contract` owns the statusline behavior when this inventory is empty. `setup-health-audit` and `agent-readiness-audit` cover instruction-file/MCP health, not data-source discovery; do not duplicate their scoring.

## Files to Touch
- `internal/loaders/inventory.go`
- `internal/loaders/inventory_test.go`
- `internal/render/sources.go`
- `cmd/thermal/main.go`
- `cmd/thermal/sources_test.go`
- `internal/render/help.go`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`
- `scripts/simulated_user_gate.sh`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-first-run-source-inventory/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-first-run-source-inventory/plan.md`](plan.md)

## Requirements

1. **Inventory model.**
   - Enumerate every tool in `loaders.AllTools()` in the stable `allToolOrder` display order.
   - For each tool emit: canonical display name, CLI alias (if different), tool id, expected path(s), path shape (`file`, `directory`, `either`), state, and an optional detail (e.g. `expected opencode.db inside this directory`).
   - States:
     - `found` — path exists and is readable; for a file, non-zero size; for a directory, at least one entry.
     - `empty` — path exists and is readable but contains no usable entries (empty directory or zero-byte file).
     - `missing` — path does not exist.
     - `unreadable` — path exists but `stat` or `open` returns permission denied.
     - `wrong-shape` — a file where a directory is expected or vice versa.
     - `unknown` — the registry has no path for the tool (should not happen; test it).
   - Never read file contents, session rows, prompts, or database tables during inventory. `stat` and a directory `ReadDir` (names only) are allowed.
   - Resolve `$HOME` and `~` consistently with the loaders; display `~` in human output and the absolute path in JSON.
   - Include the environment overrides that affect paths (`HERMES_HOME`, `DSH_HOME`, and any other documented env var) in the resolved path.

2. **Empty-state integration.**
   - Bare `thermal` with zero sources prints the inventory to **stderr** and exits `1` (compatibility), followed by:
     ```
     No agent data found. Run one prompt in any supported tool, then re-run thermal.
       thermal sources        inspect every source path and state
       thermal --help         list supported tools and flags
     ```
   - `stats`, `projects`, `models`, `mix`, `trend`, `share`, `live --json`, and `statusline` use the same inventory on their empty-data path, with a command-specific hint (`thermal stats --help`, `thermal live --json`, etc.).
   - `--verbose` adds the resolved path and the underlying `stat` error; it must not add file contents.
   - `--no-color` and `NO_COLOR` produce a plain table with no ANSI.
   - The inventory itself must not trigger a pricing fetch, update check, or cache write.
   - If exactly one source is found but empty, say so (`OpenCode: empty — no sessions yet`) rather than the generic message.
   - If a source is unreadable, name the path and the permission remedy; do not silently skip it.

3. **`thermal sources` command.**
   - `thermal sources` prints the inventory to stdout and exits `0` whenever the inventory can be produced, even if every source is missing.
   - `thermal sources --json` prints one JSON object:
     ```json
     {
       "type": "sources",
       "generatedAt": "2026-10-29T19:00:00Z",
       "home": "/home/user",
       "summary": {"total": 15, "found": 2, "empty": 1, "missing": 11, "unreadable": 1, "wrongShape": 0},
       "sources": [
         {"tool": "opencode", "name": "OpenCode", "alias": "oc", "path": "/home/user/.local/share/opencode/opencode.db", "displayPath": "~/.local/share/opencode/opencode.db", "shape": "file", "state": "found", "detail": "", "modified": "2026-10-29T18:00:00Z"}
       ]
     }
     ```
   - `thermal sources --json` must be valid with every source missing, and must not include stack traces or raw driver text.
   - `thermal sources --verbose` adds `stat` errors and the config/env override source; JSON gains a `"detail"` string but no raw errors by default.
   - `sources` is not a data scan; it must complete in <100 ms on an empty HOME and must not open any database.
   - `thermal sources <tool>` filters to one tool (optional but recommended); `thermal sources --tool opencode` is equivalent.

4. **Bare `thermal --json` empty state.**
   - Emit the same `type:"sources"` object on **stdout** with `summary.found == 0`, and exit `1`.
   - This is a new documented empty-state contract; it must not be confused with a successful leaderboard payload.
   - `thermal --json` with data keeps its existing leaderboard envelope unchanged.
   - `thermal sources --json` always exits `0`, even when empty; the difference is intentional and documented.

5. **Privacy and safety invariants.**
   - No file contents, prompt text, session titles, project names, tokens, costs, or database rows in the inventory.
   - Directory listings are used only to decide `empty` vs `found`; entry names are not printed.
   - Symlinks are resolved for display; if resolution fails, show the given path and `unreadable`.
   - No network, no pricing catalog load, no cache write, no update check.
   - The inventory respects the same environment variables as the loaders; a test asserts the resolved path changes when `HERMES_HOME`/`DSH_HOME` change.

6. **Output and width.**
   - Human table columns: `Tool`, `Alias`, `Path`, `State`, `Detail`. Fit 80 columns by shortening the path from the left (`…/.local/share/opencode/opencode.db`) and dropping the Detail column first.
   - State is a word, not a color-only signal; color is optional and never the sole encoding.
   - A legend explains `found`, `empty`, `missing`, `unreadable`, `wrong-shape`.
   - The table is sorted by the stable tool order, not alphabetically, so it matches the leaderboard.
   - No line exceeds the detected width; fall back to one line per tool at narrow widths.

7. **Tests and docs.**
   - Fixture HOME matrix: empty, one file tool found, one directory tool found, empty directory, zero-byte file, permission-denied directory (skip as root), symlink, config override, `HERMES_HOME`/`DSH_HOME` override.
   - Assert all 15 tools are present exactly once and in stable order.
   - Privacy test: place a sentinel string in a session file and assert it never appears in text or JSON.
   - Performance test: `Inventory` on an empty HOME completes in <100 ms and performs no `Open` beyond `stat`/`ReadDir`.
   - README first-run section with the exact table; MIGRATION entry for the new `thermal --json` empty payload and the `sources` command; project.jsonc feature entry.
   - Gate: empty HOME `thermal`, `thermal sources`, `thermal sources --json | jq -e .`, one-tool fixture.

## Non-Goals
- Adding a health score or MCP/instruction-file audit (owned by `audit` and `setup-health-audit`).
- Adding a background daemon or watcher.
- Reading session contents to count prompts in the inventory; the loaders already do that when a command requests data.
- Changing default path discovery (`platform-native-data-directories`).
- Adding telemetry for activation events.

## Acceptance Criteria
- AC-001: `HOME=$(mktemp -d) thermal` exits `1`, prints a table with all 15 tool names and at least 10 `missing` states on stderr, and prints the two next-step lines; stdout is empty.
- AC-002: `HOME=$(mktemp -d) thermal --json` exits `1`, stdout is valid JSON with `type:"sources"`, `summary.total==15`, `summary.found==0`, and `sources` length 15.
- AC-003: `HOME=$(mktemp -d) thermal sources` exits `0`, prints the same table to stdout, and stderr is empty.
- AC-004: `HOME=$(mktemp -d) thermal sources --json | jq -e '.type=="sources" and (.sources|length)==15'` exits `0`.
- AC-005: A fixture HOME with one OpenCode DB shows `OpenCode` as `found`, all others `missing`, and `summary.found==1`.
- AC-006: A fixture HOME with an empty `~/.claude` directory shows `Claude` as `empty`, not `found` or `missing`.
- AC-007: A fixture with a permission-denied directory shows `unreadable` and the path; running as root skips this case with a recorded note.
- AC-008: `HERMES_HOME=/tmp/hermes-fixture thermal sources --json` shows the fixture path for Nous Hermes.
- AC-009: A sentinel string placed in a fake transcript and a fake SQLite row never appears in `thermal sources` text or JSON.
- AC-010: `HOME=$(mktemp -d) thermal sources` completes in <100 ms (measure with `time` or a Go benchmark) and opens no database.
- AC-011: At `COLUMNS=80`, `thermal sources | awk 'length>80'` prints no lines.
- AC-012: `go test -race ./cmd/thermal/... ./internal/loaders/... ./internal/render/...` exits `0`, including `TestInventoryEmptyHome`, `TestInventoryOneTool`, `TestInventoryStates`, `TestInventoryPrivacy`, `TestSourcesJSON`, and `TestInventoryStableOrder`.
- AC-013: `./scripts/simulated_user_gate.sh` exits `0` and includes the source-inventory section.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Empty HOME, bare `thermal` | inventory on stderr, exit 1, no stdout |
| Empty HOME, `thermal --json` | sources JSON on stdout, exit 1 |
| Empty HOME, `thermal sources` | inventory on stdout, exit 0 |
| One tool found | that tool `found`, others `missing` |
| Empty directory | `empty`, not `found` |
| Zero-byte DB | `empty`, not `found` |
| Permission denied | `unreadable` + path + remedy |
| File where directory expected | `wrong-shape` + expected shape |
| Symlink loop | `unreadable` + given path |
| `HERMES_HOME` set | resolved path reflects it |
| Sentinel session content | never printed |
| Corrupt DB | inventory still `found` (stat only); the report command owns the open error |

## Metrics & Regression Guards
- All 15 tools present exactly once in stable order; a test fails when a tool is added without an inventory entry.
- Zero database opens during inventory; a test uses a counting wrapper or asserts only `stat`/`ReadDir`.
- Latency <100 ms on an empty HOME and <200 ms with 15 fixtures.
- Privacy grep: sentinel absent from text and JSON.
- Gate: simulated user gate asserts AC-001, AC-003, AC-004, AC-005, and AC-011.

## Risks & Mitigations
- **Risk:** Printing every path feels noisy on first run. **Mitigation:** compact 15-row table, `~` shortening, and a summary line; `--verbose` adds detail.
- **Risk:** The inventory duplicates loader path logic and drifts. **Mitigation:** reuse the registry `ToolInfo` and `ToolPathShape`; one test per tool.
- **Risk:** `stat` on a network home is slow. **Mitigation:** bound the inventory with a context/timeout (e.g. 100 ms per path, 500 ms total) and mark slow paths `unreadable` with a timeout detail.
- **Risk:** A new `thermal --json` empty payload breaks scripts expecting empty stdout. **Mitigation:** document in MIGRATION and keep exit `1`; scripts should use `thermal sources --json` for an exit-0 inventory.
- **Risk:** Permission checks behave differently as root. **Mitigation:** skip the permission case when `os.Geteuid()==0` and record the skip.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "no supported tool data found"` and `sila recall "first run empty state"`.
- **Writes**: record a `case` for the "stat-only source inventory with state taxonomy" mechanism, grounded to `internal/loaders/inventory.go`; record a `lesson` for the activation dead end.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "source inventory shipped" --open="activation funnel metrics"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-first-run-source-inventory/goal.md)` under a `feat(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the empty-home, one-tool, and privacy evidence.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
