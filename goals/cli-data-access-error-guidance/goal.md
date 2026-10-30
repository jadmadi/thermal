# Goal: Turn every data-access failure into a named condition and a runnable next step

## Goal Description
Priority: P1 (recovery). Thermal's explicit `--db` errors are model-quality; its
default-path errors are dead ends. A user who follows the help page's flagship
example on a machine without OpenCode gets one line naming a path they never
chose, with no inventory, no available-tool list, and no alternative command.
Filesystem conditions are also collapsed: a symlink loop and a `mode 000`
directory produce the same "check permissions" sentence.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `thermal opencode weekly --chart` -> exit `1`, stdout empty:
  `thermal: database not found: <home>/.local/share/opencode/opencode.db`
- `thermal --tool codex` -> exit `1`:
  `thermal: data directory not found: <home>/.codex`
- `thermal nous` -> exit `1`; the deprecation warning is followed by
  `thermal: database not found: <home>/.hermes/state.db`
- `thermal claude --db <symlink-loop>` -> exit `1`:
  `thermal: cannot read <path>; check permissions`
- `thermal claude --db <mode-000-dir>` -> exit `1`: the same sentence.
- `thermal claude --db <regular-file>` -> exit `1`:
  `thermal: <path> is a file; Claude expects a directory`
  `  pass the data directory, not a single file`
- `thermal opencode --db <missing>` -> exit `1`:
  `thermal: --db path does not exist: <path>`
  `  check the path or omit --db`

The outcome of this goal is a **data-access error contract**:

1. Every failure names the condition (missing, unreadable, wrong shape, loop,
   not-a-directory), the tool, and the display path.
2. Every default-path failure offers the inventory, the all-tools alternative,
   and the exact `--db` override.
3. Every filesystem condition gets its own remediation, not a generic
   "permissions" sentence.
4. No stdout, no stack traces, exit `1`; `--verbose` adds the raw OS error.
5. The help page's examples either run or state the prerequisite.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: cli-first-run-source-inventory, cli-db-override-scope-and-errors
- **Sequence**: 256
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3
Related goals:
- `cli-empty-data-success-contract` owns the all-empty success state; this goal
  owns a *named* tool's missing or unreadable data.
- `cli-help-glossary-and-routing` owns unknown-word classification; this goal
  owns paths that were resolved but cannot be read.
- `jsonl-partial-read-observability` owns malformed records inside an otherwise
  readable file; this goal owns the file or directory itself.

## Files to Touch
- `internal/loaders/registry.go`
- `internal/loaders/pathcheck.go`
- `internal/loaders/pathcheck_test.go`
- `cmd/thermal/diagnostics.go`
- `cmd/thermal/main.go`
- `cmd/thermal/help.go`
- `cmd/thermal/db_override_test.go`
- `cmd/thermal/simulated_user_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-data-access-error-guidance/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-data-access-error-guidance/plan.md`](plan.md)
- Prior contracts:
  [`goals/cli-db-override-scope-and-errors/goal.md`](../cli-db-override-scope-and-errors/goal.md),
  [`goals/cli-first-run-source-inventory/goal.md`](../cli-first-run-source-inventory/goal.md)

## Requirements

1. **Typed error model.**
   Introduce one user-facing error carrying:
   ```go
   type DataAccessError struct {
       Tool        string // display name, e.g. "OpenCode"
       Path        string // absolute, for --verbose
       DisplayPath string // ~-shortened, for the message
       Condition   string // missing|unreadable|wrong-shape|loop|not-dir
       Message     string
       Fix         string
       Err         error  // raw OS/driver error; --verbose only
   }
   ```
   `Error()` returns `Message` plus `Fix`; it never includes `Err`.
   `Unwrap()` returns `Err`. Reuse the existing `DBPathError` fields where
   possible; do not add a parallel type if a rename is cleaner.

2. **Condition coverage.**
   | Condition | Trigger | Message | Fix |
   | --- | --- | --- | --- |
   | `missing` (default) | `os.Stat` ENOENT on a registry path | `OpenCode has no data at ~/.local/share/opencode/opencode.db` | 3 lines: `thermal sources`, `thermal weekly` (all tools), `thermal --tool opencode --db <path>` |
   | `missing` (`--db`) | ENOENT on an explicit path | `--db path does not exist: <path>` | `check the path or omit --db` |
   | `unreadable` | EACCES/EPERM | `permission denied reading ~/.claude (mode 000)` | `ls -ld <path>` and `chmod u+rX <path>` |
   | `wrong-shape` | file where dir expected, or dir where SQLite file expected | `<path> is a file; Claude expects a directory` | `pass the data directory, not a single file` |
   | `not-dir` | ENOTDIR in a path component | `<path> is not a directory` | `check the path components or omit --db` |
   | `loop` | ELOOP | `too many levels of symbolic links at <path>` | `readlink -f <path>` to resolve, then pass the resolved path |
   | `other` | any other error | `cannot read <path>: <short reason>` | `thermal --verbose` for the driver error |

   The exact wording may change, but every row must exist and be tested.

3. **Default-path enrichment.**
   - A missing default path must print:
     ```
     thermal: OpenCode has no data at ~/.local/share/opencode/opencode.db
       inventory: thermal sources
       all tools: thermal weekly
       override:  thermal --tool opencode --db <path>
     ```
   - `thermal weekly` in the second line must be the report the user asked for
     when known (`daily`, `weekly`, `monthly`, `projects`, `models`, `stats`,
     `trend`); otherwise use `thermal`.
   - Do not print a 15-tool list here; `sources` owns that.
   - If another tool has data, add one line:
     `Claude and codewhale have data (thermal sources).`
     This requires only the stat-based inventory, never a load.
   - If no tool has data, the message must defer to the no-data contract:
     `no tool has data yet; run thermal for the first-run guide.`

4. **Explicit `--db` behavior.**
   - Keep the existing precise messages; add the condition-specific fixes.
   - Never fall back to a default path after a bad `--db`.
   - A valid `--db` path must still load even when the default path is missing.
   - A `--db` directory for a file-shaped tool reports `wrong-shape`, not
     `missing`.
   - `--db` without `--tool` keeps its existing usage error (exit `2`).

5. **Channel and exit contract.**
   - All data-access errors: stdout empty, stderr message + fix, exit `1`.
   - `--verbose` appends one line: `thermal: verbose: <raw error>`. No stack
     trace, no driver dump.
   - `--json` keeps stdout empty and writes the condition, next commands, and
     fix to stderr, matching the existing JSON-error channel contract. Do not
     print the human fix on stdout.
   - `--no-color`/`NO_COLOR` produce no ANSI; color is never the only signal.
   - SIGINT during a slow scan remains a normal stop (owned by
     `cli-sigint-lifecycle-contract`); do not report it as a data error.

6. **Help examples.**
   - `thermal opencode weekly --chart` is the flagship `[tool] [report]`
     example. Either keep it with an explicit prerequisite line
     (`requires OpenCode data; run thermal sources`) or replace it with
     `thermal weekly --chart` and move the scoped form under "more examples".
   - Every example in `thermal --help` and each command page must either run
     on a clean machine or state the prerequisite in the same block.
   - Add `thermal sources` to the "Diagnostics & Utilities" examples.

7. **Tests.**
   - Table-driven tests for each condition with a fixture filesystem:
     missing default, missing `--db`, permission denied (skip as root),
     file-for-directory, directory-for-file, `ENOTDIR`, symlink loop.
   - Assert stdout empty, exit `1`, message contains the tool, display path,
     and a runnable command.
   - Assert `--verbose` adds exactly one raw-error line.
   - Assert `--json` keeps stdout valid and does not include the human fix.
   - Assert a valid `--db` still works when the default is missing.
   - Assert the help example either runs or names its prerequisite.
   - Gate: `simulated_user_gate.sh` asserts the missing-default and loop cases.

8. **Documentation.**
   - README "Usage" and "Argument validation" sections show the enriched
     message for a missing tool.
   - MIGRATION entry only if the message text changes in a way scripts parse;
     otherwise a changelog note is enough. Human error text is not a stable
     contract unless documented.
   - `project.jsonc` updated if the help examples change.

## Non-Goals
- Removing or renaming the `database not found` wording for explicit `--db`
  paths; it is already precise.
- Making a missing tool auto-install or create directories.
- Reading session contents to decide whether a tool "has data".
- Changing the inventory table (owned by `cli-terminal-output-integrity`).
- Changing exit codes `2` (usage) or `3` (upgrade refusal).

## Acceptance Criteria
- AC-001: `HOME=<fixture-without-opencode> thermal opencode weekly --chart`
  exits `1`, stdout empty, stderr contains `OpenCode`, the display path,
  `thermal sources`, `thermal weekly`, and `--db`.
- AC-002: `thermal --tool codex` on the same HOME names `data directory` and
  offers `thermal sources`.
- AC-003: `thermal --tool claude --db <regular-file>` exits `1` with
  `expects a directory` and the existing fix.
- AC-004: `thermal --tool opencode --db <missing>` exits `1` with
  `--db path does not exist` and `check the path or omit --db`.
- AC-005: `thermal --tool claude --db <symlink-loop>` exits `1` with
  `too many levels of symbolic links`, not `check permissions`.
- AC-006: `thermal --tool claude --db <mode-000-dir>` exits `1` with
  `permission denied`, the mode, and a `chmod`/`ls -ld` fix (skip as root).
- AC-007: `--verbose` adds exactly one `thermal: verbose:` line and no stack
  trace.
- AC-008: `--json` with a missing tool exits `1` with stdout empty; the human
  condition and next commands go to stderr, and `--verbose` adds one raw line.
  This matches the existing JSON-error channel contract.
- AC-009: A valid `--db` path loads even when the tool's default path is
  missing.
- AC-010: Every help example either runs on a clean fixture or states its
  prerequisite in the same block.
- AC-011: `go test -race ./internal/loaders/... ./cmd/thermal/...` exits `0`,
  including `TestDataAccessConditions`, `TestMissingDefaultEnrichment`,
  `TestDBPathConditions`, and `TestHelpExamplesRunnable`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  missing-default and loop cases.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Missing default path | enriched message, exit 1 |
| Missing `--db` path | existing precise message, exit 1 |
| Permission denied | `permission denied`, mode, chmod fix |
| Symlink loop | `too many levels of symbolic links`, readlink fix |
| File for directory tool | `expects a directory`, existing fix |
| Directory for file tool | `expects a SQLite file` |
| `ENOTDIR` component | `is not a directory`, path-component fix |
| `--db` without `--tool` | exit 2, existing usage error |
| Valid `--db`, missing default | loads, exit 0 |
| `--json` + missing tool | valid JSON stdout, exit 1 |
| `--verbose` | one raw-error line, stderr |
| Root user + mode 000 | skip test with recorded note |

## Metrics & Regression Guards
- Condition coverage: one test per row in the condition table.
- No generic `check permissions` for a loop or `ENOTDIR`; a grep test fails if
  the loop fixture produces that sentence.
- Channel purity: stdout empty for every human error; JSON parses.
- Path display: `~` shortening is used in human output; absolute paths stay in
  `--verbose` and JSON.
- Gate: simulated user gate asserts AC-001, AC-003, and AC-005.

## Risks & Mitigations
- **Risk:** More verbose errors become noisy on every invocation.
  **Mitigation:** the enrichment is 3 lines and only on failure; `sources`
  remains the full inventory.
- **Risk:** The inventory check on the error path is slow.
  **Mitigation:** stat-only, bounded, and cached per invocation; no database
  open.
- **Risk:** Error text becomes a de facto API. **Mitigation:** document that
  only exit codes and JSON are stable; MIGRATION for any JSON change.
- **Risk:** OS error mapping differs on Windows.
  **Mitigation:** branch on `errors.Is` with `syscall` guards in build-tagged
  files; test the portable subset and skip POSIX-only cases on Windows.
- **Risk:** Enrichment duplicates the no-data onboarding.
  **Mitigation:** one `dataAccessFix(tool, path, condition)` helper; the
  all-empty case delegates to the no-data contract.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "database not found"` and
  `sila recall "cannot read check permissions"`.
- **Writes**: record a `case` for the "typed data-access error with condition
  and next commands" mechanism; record a `lesson` for the collapsed
  permission/loop diagnosis.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "data-access error contract shipped"
  --open="Windows errno mapping"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-data-access-error-guidance/goal.md)` under a `fix(cli):` or
   `feat(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the condition table and channel results.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
