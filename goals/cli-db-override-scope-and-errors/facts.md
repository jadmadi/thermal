# Facts: Make --db scope explicit and turn raw SQLite failures into actionable errors

## Observed Evidence (2026-10-29, current source build `thermal dev`)
- `thermal --db /nonexistent daily` → full report:
  ```
    Thermal · daily usage · all tools
    Window: [all time (96 active days) · all tools]  Total: [23.3B tok · ~$503.44]  Cache: [97.8% hit]
  ```
  Exit `0`, stderr empty.
- `thermal --db /tmp daily` → the same full report, exit `0`.
- `thermal --db /etc/passwd daily` → the same full report, exit `0`.
- `thermal opencode --db /nonexistent` → `thermal: database not found: /nonexistent`, exit `1`.
- `thermal opencode --db /tmp` → `thermal: cannot acquire sqlite connection: disk I/O error (10)`, exit `1`.
- `thermal opencode --db /etc/passwd` → `thermal: cannot acquire sqlite connection: file is not a database (26)`, exit `1`.
- `thermal --help` describes `--db <path>  Override database/data directory path` with no scope note.

## Root Cause
- `cmd/thermal/main.go:700-706` (leaderboard aggregate) and `:890-900` (report aggregate) call:
  ```go
  data, err := loaders.LoadToolData(t, info, "")
  ```
  The third argument is the DB override; it is hard-coded to the empty string.
- `cmd/thermal/main.go:785` and `:929` (single-tool paths) and `:1586` (live) pass `opts.DBPath`, so the flag works only when exactly one tool is selected.
- `internal/loaders/registry.go:352-367` (`LoadToolData`) uses the supplied path when non-empty and returns `database not found` when it is missing; it does not validate the shape.
- `internal/loaders/bounded.go:62` wraps the driver error verbatim: `fmt.Errorf("cannot acquire sqlite connection: %w", err)`. SQLite's `(10)` and `(26)` codes reach the terminal.
- `internal/loaders/registry.go:143-196` applies user-config overrides by probing a directory for a known filename; if the file is absent, `info.DBPath` stays at the default. That is a second silent fallback, independent of `--db`.

## Architectural Invariants & Constraints
- All SQLite connections are read-only (`?mode=ro`) and memory-mapped (`PRAGMA mmap_size`); this goal must not change either.
- A user-supplied path must never silently resolve to a default path.
- Error messages must name the tool, the path, and the expected shape; raw driver text is diagnostic, not user-facing.
- `--db` is an override, not a search path; it must not scan siblings or parent directories.
- Aggregate commands must remain able to read every tool from defaults when `--db` is absent.
- Exit `2` for scope/usage errors; exit `1` for open/validation errors.
- `--json` stdout remains empty on errors; stderr carries the message.

## Per-Tool Path Shapes (from `internal/loaders/registry.go`)
| Tool | Default | `--db` accepts |
| --- | --- | --- |
| OpenCode / MiMoCode | `opencode.db` / `mimocode.db` | SQLite file or directory containing it |
| Devin | `sessions.db` | SQLite file or directory containing it |
| ZCode | `db.sqlite` | SQLite file or directory containing it |
| Nous Hermes | `state.db` | SQLite file or directory containing it |
| Zed | `threads.db` | SQLite file or directory containing it |
| Codex | `state_5.sqlite` + rollout JSONL | SQLite file or data directory |
| Muse | `session-index.db` + `sessions/` | file or directory |
| Claude | `~/.claude/projects/*/*.jsonl` | directory |
| command-code | `~/.commandcode/projects/*/*.jsonl` | directory |
| Agy | `.../brain/*/.../transcript.jsonl` | directory |
| Droid | `~/.factory/sessions/*/*.jsonl` | directory |
| codewhale | `~/.codewhale/sessions/*.json` | directory |
| DeepSeek DSH | `~/.dsh/storages/...` | directory |
| Grok | `<grok-home>/sessions/*/*/updates.jsonl` | directory |

## File & Interface Contracts
- `internal/loaders/pathcheck.go` (new):
  ```go
  type PathShape int // shapeSQLiteFile, shapeDirectory, shapeEither
  type DBPathError struct { Tool, Path, Resolved, Expected, Kind string; Err error }
  func (e *DBPathError) Error() string // user-facing, no driver codes
  func (e *DBPathError) Unwrap() error
  func ToolPathShape(t thermal.Tool) PathShape
  func ValidateDBPath(t thermal.Tool, path string) (resolved string, err error)
  ```
- `internal/loaders/bounded.go`: keep the wrapped driver error for `--verbose`, but expose a typed `DBPathError` when the failure is path-related. Use `errors.As`.
- `cmd/thermal/main.go`: stop passing `""`; call `ValidateDBPath` once at the start of `loadUsage` when `opts.DBPath != ""`, and reject when more than one tool is in scope.
- `cmd/thermal/db_override_test.go`: scope matrix, sentinel-row no-fallback test.
- `internal/loaders/pathcheck_test.go`: shape matrix, header pre-flight, error translation.

## Error Messages (canonical)
```
thermal: --db applies to one tool; add --tool <name>
  example: thermal --tool opencode --db ~/.local/share/opencode/opencode.db daily

thermal: --db path does not exist: /nonexistent
  check the path or omit --db

thermal: /tmp is a directory; OpenCode expects opencode.db inside it
  pass the database file directly: --db /tmp/opencode.db

thermal: /etc/passwd is not a SQLite database for OpenCode
  omit --db or point it at opencode.db

thermal: /path/to/db is empty; expected a SQLite database

thermal: /path/to/db is locked by another process
  Thermal opens databases read-only; retry when the writer is idle
```
- `--verbose` appends `thermal: verbose: <raw driver error>` on stderr.
- Never include `cannot acquire sqlite connection`, `(10)`, or `(26)` in non-verbose output.

## Verification Contract
```bash
# Scope rejection
thermal --db /nonexistent daily >/tmp/o 2>/tmp/e; test $? -eq 2
test ! -s /tmp/o
grep -q -- '--db applies to one tool' /tmp/e

# Path validation
thermal --tool opencode --db /nonexistent daily >/dev/null 2>/tmp/e; test $? -eq 1
grep -q 'does not exist' /tmp/e

thermal --tool opencode --db /tmp daily >/dev/null 2>/tmp/e; test $? -eq 1
grep -q 'is a directory' /tmp/e
! grep -q 'disk I/O error' /tmp/e

thermal --tool opencode --db /etc/passwd daily >/dev/null 2>/tmp/e; test $? -eq 1
grep -q 'not a SQLite database' /tmp/e
! grep -q '(26)' /tmp/e

# No fallback: fixture HOME with a sentinel row
HOME="$fixture_home" thermal --tool opencode --db /nonexistent daily >/tmp/o 2>/dev/null
test ! -s /tmp/o
! grep -q 'SENTINEL-PROJECT' /tmp/o

# Tests
CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/loaders/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A minimal OpenCode SQLite fixture with one sentinel row, created in `t.TempDir()`.
- A Claude-style directory with one JSONL transcript.
- A plain-text file, an empty file, a zero-byte file, a directory, a FIFO (skip on Windows), and a permission-denied file.
- A fixture HOME containing the default OpenCode DB with a distinct sentinel row.
- A config-override fixture with a missing path to exercise the warning/fail-closed behavior.

## Open Questions
- Should `--db` with `--tool all` be rejected or applied to the matching tool? Recommendation: reject; silent aggregation is the defect.
- Should a directory for a SQLite tool auto-probe the expected filename? Recommendation: yes, but only the exact known filename; never scan parents.
- Should an empty directory for a directory tool exit `0` with zero activity or exit `1`? Recommendation: exit `0`; emptiness is not a path error, and the first-run goal owns the empty-state message.
- Should config-override fallback fail closed? Recommendation: warn at normal verbosity; fail closed only when `--db` is also set.
