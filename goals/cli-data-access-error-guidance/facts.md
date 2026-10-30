# Facts: Turn every data-access failure into a named condition and a runnable next step

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

All runs used a fixture HOME with Claude and codewhale data but no OpenCode,
Codex, or Hermes data, plus `--offline --no-update-check`.

### Missing default path (tool not installed)
```
$ thermal opencode weekly --chart
exit 1; stdout empty
thermal: database not found: <home>/.local/share/opencode/opencode.db

$ thermal --tool codex
exit 1; stdout empty
thermal: data directory not found: <home>/.codex

$ thermal nous
exit 1; stdout empty
thermal: warning: nous is deprecated and will be removed in v0.14.0;
  use hermes (see docs/MIGRATION.md#deprecated-tool-nous)
thermal: database not found: <home>/.hermes/state.db
```
No inventory, available-tool list, report alternative, or `--db` hint.

### Explicit `--db` path, missing
```
$ thermal opencode --db <missing>.db
exit 1; stdout empty
thermal: --db path does not exist: <path>
  check the path or omit --db
```
This one is good and must not regress.

### Explicit `--db`, wrong shape
```
$ thermal claude --db <regular-file>
exit 1; stdout empty
thermal: <path> is a file; Claude expects a directory
  pass the data directory, not a single file
```
This one is good.

### Filesystem conditions collapsed
```
$ thermal claude --db <loop-a>      # loop-a -> loop-b -> loop-a
exit 1; stdout empty
thermal: cannot read <loop-a>; check permissions

$ thermal claude --db <mode-000-dir>
exit 1; stdout empty
thermal: cannot read <noperm>; check permissions
```
Two different errno values (`ELOOP`, `EACCES`) produce one sentence.

## Root Cause
- `internal/loaders/registry.go:386` returns
  `fmt.Errorf("database not found: %s", p)` for a missing file-shaped default.
- `internal/loaders/registry.go:433` returns
  `fmt.Errorf("data directory not found: %s", dir)` for a missing
  directory-shaped default.
- These are plain errors: no tool display name, no condition enum, no next
  commands, and no `Unwrap` for `--verbose`.
- `internal/loaders/pathcheck.go:32-47` defines `DBPathError` with
  `Tool`, `Path`, `Message`, `Fix`, and `Err`; explicit `--db` validation uses
  it, which is why that path is better.
- `internal/loaders/pathcheck.go:139-143`, `:227-231`, and `:291-295` return
  `"cannot read %s; check permissions"` for every `os.Stat` failure, so
  `ELOOP`, `EACCES`, and `ENOTDIR` collapse into one remedy.
- `cmd/thermal/diagnostics.go:27-36` (`printLoadError`) prints the message and,
  under `--verbose`, unwraps only `*DBPathError`. Default-path errors have no
  raw error to show.
- `cmd/thermal/main.go:749-752` returns the registry error unchanged for a
  named tool.
- `cmd/thermal/help.go` examples include `thermal opencode weekly --chart`
  without a prerequisite line.

## Architectural Invariants & Constraints
- Human errors go to stderr; stdout stays empty unless `--json` is requested.
- Exit `1` for a data/IO failure; `2` stays reserved for usage errors.
- The error must name the tool, the condition, and a display path; the raw
  absolute path is for `--verbose` and JSON.
- Never open a database to decide whether it exists; use `stat`, `ReadDir`,
  `Lstat`, and `EvalSymlinks` only.
- The error path must not trigger a pricing fetch, update check, or cache write.
- The inventory must respect the same environment overrides as the loaders
  (`HERMES_HOME`, `DSH_HOME`, platform paths).
- A symlink loop must not hang; bound resolution and report `ELOOP`.
- Do not print the 15-tool inventory on every failure; offer `sources`.
- The help example must be honest: either it runs or it states the
  prerequisite in the same block.

## File & Interface Contracts
- `internal/loaders/pathcheck.go`:
  ```go
  type DataCondition string

  const (
      ConditionMissing    DataCondition = "missing"
      ConditionUnreadable DataCondition = "unreadable"
      ConditionWrongShape DataCondition = "wrong-shape"
      ConditionNotDir     DataCondition = "not-dir"
      ConditionLoop       DataCondition = "loop"
      ConditionOther      DataCondition = "other"
  )

  type DataAccessError struct {
      Tool        string
      Path        string
      DisplayPath string
      Condition   DataCondition
      Message     string
      Fix         string
      Err         error
  }

  func classifyDataAccess(tool, path string, err error) *DataAccessError
  ```
  `Error()` returns `Message` plus `"\n  " + Fix`; `Unwrap()` returns `Err`.
- `internal/loaders/registry.go`: replace the two `fmt.Errorf` sites with
  `classifyDataAccess` and the tool display name.
- `cmd/thermal/diagnostics.go`: `printLoadError` uses `errors.As` for
  `*DataAccessError`; under `--verbose` prints `thermal: verbose: <Err>`.
- `cmd/thermal/main.go`: `loadUsage` returns the typed error; no change to
  `os.Exit` mapping.
- `cmd/thermal/help.go`: add the prerequisite line or swap the example.
- `internal/loaders/inventory.go`: reuse `loaders.Inventory` for the
  "another tool has data" line; do not add a second path walk.

## Condition Contract
| Condition | errno / trigger | Message stem | Fix |
| --- | --- | --- | --- |
| `missing` default | ENOENT | `<Tool> has no data at <display>` | sources, all-tools report, `--db` |
| `missing` explicit | ENOENT | `--db path does not exist: <path>` | `check the path or omit --db` |
| `unreadable` | EACCES, EPERM | `permission denied reading <display>` | `ls -ld`, `chmod u+rX` |
| `wrong-shape` | file/dir mismatch | `<path> is a file; <Tool> expects a directory` | pass the directory |
| `not-dir` | ENOTDIR | `<path> is not a directory` | check path components |
| `loop` | ELOOP | `too many levels of symbolic links at <path>` | `readlink -f` |
| `other` | anything else | `cannot read <path>: <short>` | `--verbose` |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=$(mktemp -d)
mkdir -p "$H/.claude/projects/p"
printf '%s\n' '{"type":"assistant","timestamp":"2026-10-30T09:00:00Z",
"message":{"id":"m1","model":"claude-3-5-haiku",
"usage":{"input_tokens":100,"output_tokens":50}}}' \
  > "$H/.claude/projects/p/s.jsonl"

# Missing default is enriched
HOME=$H $B opencode weekly --chart >/tmp/o 2>/tmp/e; test $? -eq 1
test ! -s /tmp/o
grep -q 'OpenCode' /tmp/e
grep -q 'thermal sources' /tmp/e
grep -q 'thermal weekly' /tmp/e
grep -q -- '--db' /tmp/e

# Missing --db stays precise
HOME=$H $B opencode --db /nope.db >/tmp/o 2>/tmp/e; test $? -eq 1
grep -q -- '--db path does not exist' /tmp/e
grep -q 'check the path' /tmp/e

# Wrong shape stays precise
: > /tmp/notadir
HOME=$H $B claude --db /tmp/notadir >/tmp/o 2>/tmp/e; test $? -eq 1
grep -q 'expects a directory' /tmp/e

# Loop vs permission
ln -sfn /tmp/loop-b /tmp/loop-a; ln -sfn /tmp/loop-a /tmp/loop-b
HOME=$H $B claude --db /tmp/loop-a >/tmp/o 2>/tmp/e; test $? -eq 1
grep -q 'too many levels of symbolic links' /tmp/e

mkdir -p /tmp/noperm; chmod 000 /tmp/noperm
HOME=$H $B claude --db /tmp/noperm >/tmp/o 2>/tmp/e; test $? -eq 1
grep -q 'permission denied' /tmp/e
chmod 755 /tmp/noperm

# Verbose adds one raw line, no stack
HOME=$H $B opencode --verbose >/tmp/o 2>/tmp/e; test $? -eq 1
grep -c 'thermal: verbose:' /tmp/e | grep -qx 1
! grep -q 'goroutine' /tmp/e

CGO_ENABLED=0 go test -race ./internal/loaders/... ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- HOME with Claude/codewhale data and no OpenCode/Codex/Hermes.
- Missing explicit `--db` file.
- Regular file where a directory is expected.
- Directory where a SQLite file is expected.
- `mode 000` directory (skip as root).
- Symlink loop (`a -> b -> a`).
- Path with an `ENOTDIR` component (`file/child`).
- `HERMES_HOME`/`DSH_HOME` overrides.
- A slow/stat-blocked path to prove the timeout.

## Open Questions
- Should the default-path message use `database not found` or `has no data at`?
  Recommendation: `has no data at` for the default path (state, not a failed
  query); keep `database not found` for an explicit `--db`.
- Should the error list every tool that has data?
  Recommendation: name at most two plus `thermal sources`; a full list belongs
  to the inventory.
- Should `--json` include the human `fix` string?
  Recommendation: no; keep the existing error envelope and put remediation in
  `next` only if the envelope already supports it.
- Should the error path run the inventory?
  Recommendation: yes, stat-only, bounded to 100 ms; if it times out, omit the
  "another tool has data" line rather than block the error.
- Should Windows use `ERROR_CANT_RESOLVE_FILENAME` for loops?
  Recommendation: map it to `loop` in a build-tagged helper; skip POSIX-only
  tests on Windows.
