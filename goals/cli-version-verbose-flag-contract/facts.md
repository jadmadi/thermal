# Facts: Give -v, -V, --version, and --verbose one unambiguous contract

## Observed Evidence (2026-10-29, current source build `thermal dev`)
| Invocation | stdout | Exit |
| --- | --- | --- |
| `thermal -v` | `thermal dev (built from source)` | 0 |
| `thermal --version` | same | 0 |
| `thermal version` | same | 0 |
| `thermal -v daily` | daily report, no version | 0 |
| `thermal -v --json` | leaderboard JSON, no version | 0 |
| `thermal -V` | stderr `flag provided but not defined: -V` + full help | 2 |
| `thermal --help` | `--verbose, -v   Output diagnostic warnings to stderr`; no `--version` | 0 |

## Root Cause
- `cmd/thermal/main.go:48-63`:
  ```go
  if len(os.Args) == 2 {
      arg := os.Args[1]
      ...
      if arg == "-v" || arg == "--version" || arg == "version" {
          fmt.Printf("thermal %s\n", version.String())
          ...
          os.Exit(0)
      }
  }
  ```
  The meaning of `-v` depends on `len(os.Args) == 2`.
- `cmd/thermal/main.go:107`: `flag.BoolVar(&opts.Verbose, "v", false, ...)` registers `-v` as verbose for every other argv shape.
- `cmd/thermal/main.go:174`: the post-positional scanner also maps `"-v"` to `opts.Verbose = true`.
- `cmd/thermal/main.go:585` dispatch has a `version` case; `--version` is not registered in the flag set, so outside the `len==2` branch it would be an unknown flag.
- `internal/render/help.go` lists only `--verbose, -v`; there is no `-V`/`--version` line and no precedence note.
- `version.String()` is the text source; `version --json` is handled by the JSON path, but `-V --json` does not exist.

## Architectural Invariants & Constraints
- One spelling, one meaning. A flag's semantics must not depend on argv length or position.
- Version is a read-only metadata command: no loader, pricing, cache write, update check, or network.
- Help > version > validation > execution (owned by `cli-help-interception-and-scoped-help`); this goal supplies the version half of that table.
- `-v`/`--verbose` must remain available for diagnostics in every command, including version-adjacent paths.
- Version output must be deterministic apart from an optional timestamp; prefer no timestamp for testability.
- Unknown-flag rejection from `cli-argument-parser-strictness` must accept `-V` as a known flag.

## File & Interface Contracts
- `cmd/thermal/main.go`:
  - remove the `len(os.Args)==2` version branch;
  - add `versionRequested(args []string) bool` for `-V`, `--version`, and the `version` command;
  - run the help resolver first, then the version resolver, then parse/validate, then dispatch.
- `cmd/thermal/args.go`: declare `version` as a boolean flag with alias `V`; ensure it is not consumed as a value.
- `cmd/thermal/help.go`: `commandSpec` for `version` with usage, flags, examples, JSON keys.
- `internal/render/help.go`: add the two flag lines and the precedence note.
- `cmd/thermal/version_test.go`: matrix, JSON parity, no-side-effect.
- `scripts/simulated_user_gate.sh`: version/verbose section.

## Output Contract
Text:
```
thermal dev (built from source)
  commit: 5f652d4
  built:  2026-10-29T10:19:25Z
```
- Omit `commit`/`built` when the value is `unknown` (current behavior).
- Exit `0`, stdout only, no stderr, no ANSI.

JSON:
```json
{
  "type": "version",
  "version": "dev (built from source)",
  "commit": "5f652d4",
  "date": "2026-10-29T10:19:25Z"
}
```
- The same object for `-V --json`, `--version --json`, and `version --json`.
- If a `generatedAt` field is retained, inject a clock in tests or exclude it from parity comparisons.
- No data, pricing, or network fields.

## Precedence Matrix
| argv | Result |
| --- | --- |
| `-h` / `--help` | global help |
| `-h -V` / `-V -h` | help wins (help > version) |
| `-V` / `--version` / `version` | version |
| `-V daily` / `daily -V` | version; command ignored |
| `-v -V` | version; verbose flag ignored for output |
| `-V --bogus` | version; validation skipped |
| `-v` | verbose normal execution |
| `-v daily` / `daily -v` | verbose daily |
| `-v --json` | command JSON |
| `-V` empty HOME | version, exit 0 |
| `-v` empty HOME | empty-state error, exit 1 |

## Verification Contract
```bash
# -v is verbose, not version
thermal -v >/tmp/o 2>/tmp/e; test $? -eq 0
grep -q 'Token Warriors' /tmp/o
! grep -q 'thermal dev' /tmp/o

thermal -v daily >/tmp/o; grep -q 'Thermal · daily usage' /tmp/o

# -V is version
thermal -V >/tmp/o; test $? -eq 0; grep -q 'thermal dev' /tmp/o
diff <(thermal -V) <(thermal --version)
diff <(thermal -V) <(thermal version)

# JSON parity
thermal -V --json | jq -e '.type=="version"' >/dev/null
diff <(thermal -V --json | jq -S 'del(.generatedAt)') <(thermal version --json | jq -S 'del(.generatedAt)')

# Precedence
thermal -v -V | grep -q 'thermal dev'
thermal -V daily | grep -q 'thermal dev'
! thermal -V daily | grep -q 'Daily Usage'
thermal -V --bogus >/dev/null; test $? -eq 0

# Help truth
thermal --help | grep -q -- '-V, --version'
thermal --help | grep -q -- '--verbose, -v'
! thermal --help | grep -q -- '-v, --version'

# Empty HOME
home=$(mktemp -d)
HOME=$home thermal -V >/dev/null; test $? -eq 0
HOME=$home thermal -v >/tmp/o 2>/tmp/e; test $? -eq 1
! grep -q 'thermal dev' /tmp/o

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A version fixture with known commit/date, injected via ldflags or a test seam.
- A fixed clock for JSON parity.
- An empty temp HOME.
- A matrix of argv positions for `-V`, `-v`, `--version`, `--verbose`, and `version`.
- A help snapshot with both flag lines.

## Open Questions
- Should `-V` win over `-h`? Recommendation: no; help > version, consistent with the help goal.
- Should `-V --bogus` exit `0`? Recommendation: yes; version > validation, documented. If the team prefers strict flags, make version > validation only when no unknown flag is present; pick one and test it.
- Should `-v` remain an alias for verbose or be removed? Recommendation: keep `-v` as verbose; it matches the help and common convention.
- Should version JSON include `generatedAt`? Recommendation: no; deterministic output is more useful.
- Should `thermal version daily` print version or reject the extra argument? Recommendation: version wins; document it.
