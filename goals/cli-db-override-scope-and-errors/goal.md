# Goal: Make --db scope explicit and turn raw SQLite failures into actionable errors

## Goal Description
Priority: P0 (scope and trust). `--db` is documented as "Override database/data directory path", but the aggregate loaders ignore it. In all-tools mode (`thermal`, `daily`, `stats`, `projects`, `models`, `mix`, `trend`, `share`, `statusline`), `loadUsage` calls `loaders.LoadToolData(t, info, "")` — an empty override — so every tool reads its default location regardless of `--db`. A user who believes they scoped the read to one directory gets every other tool's data from `$HOME` with exit `0`.

Observed on the current source build (2026-10-29):

- `thermal --db /nonexistent daily` renders the full machine-wide report (`23.3B tok`, 96 active days) and exits `0`.
- `thermal --db /tmp daily` and `thermal --db /etc/passwd daily` do the same.
- `thermal opencode --db /nonexistent` correctly exits `1` with `thermal: database not found: /nonexistent`.
- `thermal opencode --db /tmp` exits `1` with `thermal: cannot acquire sqlite connection: disk I/O error (10)`.
- `thermal opencode --db /etc/passwd` exits `1` with `thermal: cannot acquire sqlite connection: file is not a database (26)`.
- The help line gives no scope, no example, and no indication that a tool must be selected.

The outcome of this goal is a **scope-honest override**: `--db` either applies to exactly one selected tool with a validated path, or it is rejected before any data is read; it never silently falls back to the default database; and every unusable path produces an actionable, tool-specific message while the raw driver text stays behind `--verbose`.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 246
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 2 (after `cli-argument-parser-strictness` lands the `--db` value type and empty-value rejection)
- **Related goals**: `cli-argument-parser-strictness` owns missing/empty `--db` values; this goal owns path semantics, scope, and error translation. `platform-native-data-directories` owns default path discovery. `cli-first-run-source-inventory` should reuse the same per-tool path metadata.

## Files to Touch
- `cmd/thermal/main.go`
- `internal/loaders/registry.go`
- `internal/loaders/pathcheck.go`
- `internal/loaders/bounded.go`
- `cmd/thermal/diagnostics.go`
- `cmd/thermal/db_override_test.go`
- `internal/loaders/pathcheck_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`
- `internal/render/help.go`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-db-override-scope-and-errors/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-db-override-scope-and-errors/plan.md`](plan.md)

## Requirements

1. **Scope contract.**
   - `--db` requires exactly one tool in scope: `--tool <name>`, a positional `<tool>`, or a command that is inherently per-tool (`thermal opencode`, `thermal opencode weekly`).
   - `--db` with `--tool all`, `--tool auto`, or no tool selection exits `2` with:
     `thermal: --db applies to one tool; add --tool <name>` and `example: thermal --tool opencode --db ~/.local/share/opencode/opencode.db daily`.
   - The rejection happens before any database open; assert with a populated fixture HOME that no report header is printed and no default data is read.
   - If the implementation chooses the alternative (apply the path to the one matching tool and warn), it must print a warning at normal verbosity that other tools still use default locations, and the warning must not go to stdout or JSON. The recommended contract is rejection.
   - `--db` remains valid for `live --tool <name>`, `statusline --tool <name>`, and `share --tool <name>` when a single tool is selected.

2. **Path validation before opening.**
   - The path must exist; missing → exit `1`, `thermal: --db path does not exist: <path>` + `check the path or omit --db`.
   - The path must be readable by the current user; permission denied → exit `1` with the path and `check file permissions`.
   - A file must be a regular file (not a directory, socket, FIFO, or device).
   - For SQLite-file tools, pre-flight the first 16 bytes for `SQLite format 3\0`; a mismatch → exit `1`, `thermal: <path> is not a SQLite database for <Tool>` + `omit --db or point it at <expected filename>`.
   - A directory for a SQLite-file tool is valid only if it contains the expected filename; otherwise exit `1`, `thermal: <path> is a directory; <Tool> expects <filename> inside it`.
   - For directory-scanning tools, a file path is rejected with the expected directory shape; a directory path is accepted even if it is empty (empty is a valid zero-data source, handled by the empty-state goal).
   - Resolve symlinks and print the canonical path in errors; do not silently follow a symlink outside the expected prefix.
   - Zero-byte SQLite files are rejected with `is empty; expected a SQLite database`, not a driver error.

3. **Per-tool path semantics.**
   - Document and test a table: SQLite-file tools (OpenCode `opencode.db`, MiMoCode `mimocode.db`, Devin `sessions.db`, ZCode `db.sqlite`, Nous Hermes `state.db`, Zed `threads.db`, Codex `state_5.sqlite`, Muse `session-index.db`) and directory tools (Claude, command-code, Agy, Droid, codewhale, Codex rollout, DSH, Grok, Muse sessions).
   - `--db` semantics must be identical for the same tool across `thermal <tool>`, `thermal <tool> daily`, `live`, `statusline`, and `share`.
   - Do not change the default path registry; only the override resolution.

4. **Error translation.**
   - Never print `cannot acquire sqlite connection:` with a raw driver code to a non-verbose user.
   - Translate at the loader boundary:
     - `file is not a database` → `<path> is not a SQLite database for <Tool>`.
     - `disk I/O error` on a directory → `<path> is a directory; <Tool> expects <filename>`.
     - `unable to open database file` → `cannot open <path>; check permissions and that the file exists`.
     - `database is locked` → `<path> is locked by another process; Thermal opens databases read-only, so retry when the writer is idle`.
     - `no such table` → `<Tool> database at <path> has no <expected> table; it may be from an unsupported version`.
   - Keep the raw error available with `--verbose` on stderr, prefixed `thermal: verbose:`.
   - Exit `1` for all path/open failures; exit `2` only for the scope/usage error in requirement 1.
   - `--json` output stays empty on stdout for these errors; the message is stderr only.

5. **No silent fallback.**
   - If `--db` is set, the loader must never fall back to `info.DBPath` or a config override. `LoadToolData` already prefers the provided path but returns an error when missing; the aggregate path must stop passing `""`.
   - Audit the user-config override path (`internal/loaders/registry.go:143-196`): a configured path that is missing currently leaves the default in place. Add a warning (at normal verbosity) naming the config key and the fallback, or fail closed; document the choice. This is an adjacent silent fallback and must be covered by a test.
   - Add a test with a fixture HOME containing a valid default DB: `--db /nonexistent` must not produce any rows/report from that default DB.

6. **Help, docs, and migration.**
   - Help: `--db <path>  Read one tool's data from a file or directory (requires --tool)` plus one file example and one directory example.
   - README troubleshooting: three cases (wrong scope, directory vs file, non-SQLite file) with exact messages.
   - MIGRATION six-part entry: `--db` without a single tool now exits `2`; invalid paths now exit `1`; raw SQLite text is replaced; detection and fix commands; escape hatch is to omit `--db`.
   - `project.jsonc` updated.
   - `scripts/simulated_user_gate.sh`: assert AC-001 through AC-006 and AC-009.

## Non-Goals
- Changing default data-directory discovery (`platform-native-data-directories`).
- Adding support for remote databases or `sqlite://` URIs.
- Changing the read-only connection mode or `mmap_size` pragma.
- Adding a new `doctor` command (owned by the first-run goal; this goal exposes reusable path metadata).
- Changing the JSON envelopes.

## Acceptance Criteria
- AC-001: `thermal --db /nonexistent daily` exits `2`, stdout empty, stderr contains `--db applies to one tool` and `--tool`; it does not contain a `Thermal · daily usage` header.
- AC-002: `thermal --tool opencode --db /nonexistent daily` exits `1`, contains `--db path does not exist`, and prints no report.
- AC-003: `thermal --tool opencode --db /tmp daily` exits `1`, contains `is a directory` and `opencode.db`, and does not contain `disk I/O error`.
- AC-004: `thermal --tool opencode --db /etc/passwd daily` exits `1`, contains `is not a SQLite database`, and does not contain `(26)`.
- AC-005: `thermal --tool opencode --db <fixture>/opencode.db daily` exits `0` and reports the fixture's rows.
- AC-006: `thermal --tool claude --db <fixture-jsonl-dir> daily` exits `0`; `--db <fixture.jsonl>` exits `1` with the expected directory shape.
- AC-007: `thermal --tool opencode --db <claude-dir> daily` exits `1` with the expected filename, not a panic.
- AC-008: `thermal --tool opencode --db /etc/passwd daily --verbose` includes the raw driver text on stderr prefixed `thermal: verbose:`; without `--verbose`, the raw text is absent.
- AC-009: With a fixture HOME containing a valid default OpenCode DB, `thermal --tool opencode --db /nonexistent daily` prints no report and exits `1`; the default DB is not read (assert via a sentinel row that must not appear).
- AC-010: `thermal --db <valid> --tool all daily` and `thermal --db <valid> daily` both exit `2`; `thermal --tool opencode --db <valid> daily` exits `0`.
- AC-011: `go test -race ./cmd/thermal/... ./internal/loaders/...` exits `0`, including `TestDBScopeMatrix`, `TestDBPathValidation`, `TestDBErrorTranslation`, and `TestDBNeverFallsBack`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the `--db` section.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `--db /nonexistent daily` | exit 2, scope error |
| `--db <valid> --tool all daily` | exit 2, scope error |
| `--tool opencode --db /nonexistent` | exit 1, path does not exist |
| `--tool opencode --db /tmp` | exit 1, directory + expected filename |
| `--tool opencode --db /etc/passwd` | exit 1, not a SQLite database |
| `--tool opencode --db <empty file>` | exit 1, empty database |
| `--tool claude --db <file.jsonl>` | exit 1, expects a directory |
| `--tool claude --db <empty dir>` | exit 0, zero activity (empty state) |
| `--tool muse --db <sessions dir>` | exit 0 |
| `--tool muse --db <session-index.db>` | exit 0 (file accepted) |
| `--db <path>` with no tool | exit 2, scope error |
| Config override path missing | warning naming the key and fallback, or fail closed |

## Metrics & Regression Guards
- Zero default reads when `--db` is set: a sentinel row in the default fixture must never appear in output.
- Error voice: no `sqlite`, `(10)`, `(26)`, `cannot acquire` in non-verbose stderr; a grep test enforces it.
- Scope: every aggregate command (`daily`, `weekly`, `monthly`, `projects`, `models`, `stats`, `mix`, `trend`, `share`, `statusline`) rejects `--db` without a single tool.
- Path matrix: one test case per supported tool, file and directory shapes.
- Gate: the simulated user gate asserts AC-001, AC-003, AC-004, AC-005, and AC-010.

## Risks & Mitigations
- **Risk:** Rejecting `--db` in aggregate mode breaks a user who relied on the silent behavior. **Mitigation:** that behavior returned unrelated data; document in MIGRATION and suggest per-tool invocations.
- **Risk:** SQLite header pre-flight rejects valid encrypted/WAL files. **Mitigation:** only pre-flight the header when the loader expects plain SQLite; allow `--verbose` to bypass? No — validate with a fixture for every supported format; if a supported variant lacks the header, document and exempt it.
- **Risk:** Path validation duplicates loader behavior and drifts. **Mitigation:** one `ValidateDBPath` used by all callers; loaders still perform their own open and error mapping.
- **Risk:** Symlink resolution changes the path shown in output. **Mitigation:** show both given and resolved paths when they differ.
- **Risk:** Config override warnings become noisy. **Mitigation:** warn once per tool per invocation, only when the configured path is missing.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "--db override ignored aggregate"` and `sila recall "sqlite disk I/O error directory"`.
- **Writes**: record a `case` for the "scope-gated path override with per-tool shape validation" mechanism, grounded to `internal/loaders/pathcheck.go`; record a `lesson` for the silent aggregate fallback.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "--db scope and path errors shipped" --open="doctor source inventory"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-db-override-scope-and-errors/goal.md)` under a `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the scope/path/error matrix.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
