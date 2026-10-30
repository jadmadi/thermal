# Facts: Make version output and deprecation warnings factually true

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

### Version payload
```
$ thermal version
exit 0; stdout:
thermal dev (built from source)

$ thermal version --json
exit 0; stdout:
{"commit":"unknown","date":"unknown","version":"dev (built from source)"}
```

### Help promise
```
$ thermal --help
version   Print SemVer, commit hash, and build timestamp
```
The output supplies none of the three in a comparable form for a source build.

### Verbose signal
```
$ thermal -v
exit 0; stderr 0 B
stdout: the normal leaderboard (THERMAL — Don't break the streak.)
```
No indication that verbose mode is active.

### Stale deprecation deadlines
```
$ thermal --license
exit 0; stderr:
thermal: warning: --license is deprecated and will be removed in v0.14.0;
  use thermal license (see docs/MIGRATION.md#deprecated-license-flag)

$ thermal nous
exit 1; stderr:
thermal: warning: nous is deprecated and will be removed in v0.14.0;
  use hermes (see docs/MIGRATION.md#deprecated-tool-nous)
thermal: database not found: <home>/.hermes/state.db

$ thermal changelog --top 5
Current: dev · Showing 5 of 25 releases · v0.21.0 is the newest
```
The deadline is seven minor releases in the past, and the remediation path is
repository-relative.

## Root Cause
- `internal/version/version.go:14-18`:
  ```go
  var (
      Version = "dev"
      Commit  = "unknown"
      Date    = "unknown"
  )
  ```
  These are intended ldflags targets for release builds.
- `internal/version/version.go:20-26` (`String`) returns
  `"dev (built from source)"` for `Version == "dev"`.
- `cmd/thermal/version.go:16-40` (`versionOutput`) omits `commit` and `built`
  when they are `"unknown"`, and the JSON always includes the raw unknown
  strings without a `stamped` field.
- `cmd/thermal/help.go:334-349` (version help page) says
  `Summary: "SemVer, commit hash, and build timestamp"` unconditionally.
- `cmd/thermal/main.go:61` emits the `--license` warning with the literal
  `"0.14.0"`.
- `cmd/thermal/main.go:812-832` (`emitDeprecationWarningTo`) formats
  `%s is deprecated and will be removed in v%s; use %s (see
  docs/MIGRATION.md#%s)` and embeds the repo-relative path.
- `cmd/thermal/args.go:420-422` maps `-v`, `--verbose`, and `-d` to
  `opts.Verbose`; `cmd/thermal/main.go` never announces it.
- There is no current-version guard and no single deprecation table.

## Architectural Invariants & Constraints
- Version output is deterministic and read-only; no scan, network, pricing,
  cache, or update check.
- The fields are always present; `unknown` is a value, not an omission.
- A source build must not imply a SemVer it does not have.
- `-v` is verbose, `-V` is version; the meanings never depend on position or
  argument count.
- A warning must be true at the time it is printed. A removal version in the
  past is a bug, not a style issue.
- A remediation must be runnable and resolvable: a command or a `https://`
  URL, never a repo-relative path.
- Warnings go to stderr; JSON stdout stays valid.
- One warning per invocation, even for a repeated flag.
- No colour on a non-TTY; `NO_COLOR` is honored.

## File & Interface Contracts
- `internal/version/version.go`:
  ```go
  type Info struct {
      Version string
      Semver  string
      Commit  string
      Date    string
      Stamped bool
  }

  func Current() Info
  func IsPast(target string) bool // false when Current().Semver == ""
  ```
  `Info.Stamped` is `Version != "dev"` and `Commit != "unknown"`.
- `cmd/thermal/version.go`:
  ```go
  func versionOutput(asJSON bool) string
  ```
  Text always prints two lines; JSON adds `stamped` and `semver`.
- `cmd/thermal/main.go`:
  ```go
  type Deprecation struct {
      Item, Replacement, RemovalVersion, Anchor string
  }

  var deprecations = []Deprecation{...}

  func emitDeprecationWarningTo(w io.Writer, d Deprecation,
      current version.Info, noColor bool)
  func announceVerbose(w io.Writer)
  ```
  `announceVerbose` writes one line and is called once when `opts.Verbose`.
- `cmd/thermal/args.go`: no new flag; `-v`/`--verbose` already exist.
- `cmd/thermal/help.go`: update the version summary and description; no
  unconditional metadata promise.

## Deprecation Table
| Item | Replacement | Removal version | Anchor |
| --- | --- | --- | --- |
| `--license` | `thermal license` | empty (future release) | `https://jadmadi.net/project/thermal/migration#deprecated-license-flag` |
| `nous` | `hermes` | empty (future release) | `https://jadmadi.net/project/thermal/migration#deprecated-tool-nous` |
| `--dense` | `thermal stats` | empty | `...#retired-dense-grid` |
| `--port`/`--host` | `thermal live --json --stream` | empty | `...#retired-web-server` |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-home>

# Unstamped payload
$B version >/tmp/o 2>/tmp/e; test $? -eq 0
grep -q 'version not stamped' /tmp/o
grep -q 'commit: unknown' /tmp/o
grep -q 'built: unknown' /tmp/o
$B version --json | jq -e '.stamped==false and .commit=="unknown"
  and .date=="unknown"' >/dev/null

# Stamped payload (test build with -ldflags)
go build -ldflags "-X github.com/jadmadi/thermal/internal/version.Version=v9.9.9
  -X github.com/jadmadi/thermal/internal/version.Commit=abc123
  -X github.com/jadmadi/thermal/internal/version.Date=2026-01-01T00:00:00Z" \
  -o /tmp/thermal-stamped ./cmd/thermal
/tmp/thermal-stamped version --json | jq -e '.stamped==true
  and .semver=="v9.9.9" and .commit=="abc123"' >/dev/null

# Help truth
$B version --help | grep -q 'source-build fallback'
! $B --help | grep -q 'Print SemVer, commit hash, and build timestamp'

# Verbose signal
$B -v >/tmp/o 2>/tmp/e; test $? -eq 0
grep -c 'verbose diagnostics enabled' /tmp/e | grep -qx 1
$B -v --json 2>/tmp/e | jq -e . >/dev/null
grep -q 'verbose diagnostics enabled' /tmp/e

# Deprecation truth
$B --license >/tmp/o 2>/tmp/e; test $? -eq 0
! grep -q 'v0.14.0' /tmp/e
! grep -q 'docs/MIGRATION.md' /tmp/e
grep -q 'thermal license' /tmp/e
grep -q 'https://' /tmp/e
$B nous >/tmp/o 2>/tmp/e; test $? -eq 1
! grep -q 'v0.14.0' /tmp/e
grep -q 'hermes' /tmp/e

# No expired version in source
! grep -R 'will be removed in v0.14.0' cmd/ internal/

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/version/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Unstamped source build.
- Stamped build with ldflags.
- A fake current version newer than a table entry (unit test) to prove the
  runtime fallback.
- A fake current version older than a table entry to prove the future form.
- `--license`, `nous`, `--dense`, and `--port` warnings.
- `-v`, `-v --json`, `--verbose=false`.

## Open Questions
- Should `commit: unknown` always print? Recommendation: yes; omission hides
  the gap.
- Should the version JSON include `generatedAt`? Recommendation: no; keep it
  deterministic.
- Should the warning keep a future removal version? Recommendation: yes, only
  when a test proves it is still in the future; otherwise omit it.
- Should `dev` be treated as "past" for a dated warning? Recommendation: no;
  it is unknown. The source grep test catches stale literals, and release
  builds are checked against the current SemVer.
- Should the migration URL live in one constant? Recommendation: yes; one
  `migrationURL(anchor)` helper.
