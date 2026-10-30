# Facts: Document exit codes and give upgrade global-flag parity

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

### A known global flag is "unknown" on upgrade
```
$ thermal upgrade --check --offline
exit 2; stdout empty
thermal: unknown flag for upgrade: --offline
  run thermal upgrade --help
  run thermal --help for the flag list

$ thermal upgrade --check --json
exit 0; stdout:
{"type":"upgrade","action":"check","current":"dev","latest":"v0.21.0",
 "asset":"thermal_0.21.0_linux_amd64.tar.gz","target":"/tmp/thermal-cli-review",
 "mutated":false,"migration":""}
```

### The exit-code scheme is real but undocumented
```
$ thermal version              -> exit 0
$ thermal opencode             -> exit 1  (database not found)
$ thermal --tool open          -> exit 2  (unknown tool)
$ thermal upgrade --json       -> exit 3
{"type":"upgrade","action":"declined","current":"dev","latest":"",
 "error":"confirmation required for --json; re-run with --yes or use --check",
 "mutated":false}
```
`thermal --help` has no "Exit codes" section. `thermal upgrade --help` mentions
only `3`.

## Root Cause
- `cmd/thermal/upgrade_flow.go:127-186` (`parseUpgradeArgs`) is a separate
  parser. Its switch accepts `--check`, `--dry-run`, `--json`, `--yes`,
  `--force`, and no-ops `--no-color`, `-v`, `--verbose`, `-d`; everything else
  falls into `unknown flag for upgrade`.
- The global parser in `cmd/thermal/args.go` never sees the `upgrade` argv;
  `upgrade` is intercepted before normal parsing.
- `cmd/thermal/upgrade_flow.go:159` is the no-op list; `--offline` and
  `--no-update-check` are absent.
- Exit codes are literal `0`, `1`, `2`, `3` at call sites:
  `cmd/thermal/upgrade_flow.go:129` (usage `2`), `:190` (refusal `3`),
  `:238`/`:259`/`:269` (declines `3`), `:360`/`:390` (success/error).
- `cmd/thermal/help.go` has no shared exit-code block; only the upgrade page's
  `Note` mentions `3`.

## Architectural Invariants & Constraints
- A known global flag is never "unknown"; it is honored, a no-op, or
  explicitly not applicable with an alternative.
- `--offline` means no network. On upgrade, use a cache or fail closed; never
  silently fetch.
- A non-interactive mutation still requires `--yes`; exit `3` is preserved.
- JSON emits exactly one object on stdout, including on error; human text goes
  to stderr.
- Exit codes are stable: `0` success, `1` data/IO/network, `2` usage,
  `3` refused.
- Documentation and tests share one table; no page invents a fifth code.
- `--width` is not applicable to upgrade; the rejection must name `--json`.
- `--bogus` remains an unknown flag, exit `2`.

## File & Interface Contracts
- `cmd/thermal/exit.go` (new):
  ```go
  const (
      ExitOK      = 0
      ExitFailure = 1
      ExitUsage   = 2
      ExitRefused = 3
  )

  func printExitCodes(w io.Writer)
  ```
- `cmd/thermal/upgrade_flow.go`:
  ```go
  type FlagDisposition string
  const (
      FlagHonored  FlagDisposition = "honored"
      FlagNoOp     FlagDisposition = "no-op"
      FlagRejected FlagDisposition = "not applicable"
  )

  var upgradeFlagRules = map[string]FlagRule{
      "--offline":         {FlagHonored, "use the cached release index", ""},
      "--no-update-check": {FlagNoOp, "upgrade is explicit", ""},
      "--width":           {FlagRejected, "upgrade output is fixed", "--json"},
      "--fresh":           {FlagRejected, "upgrade has no cache to bypass", "--json"},
  }
  ```
- `upgradeOptions` gains `Offline bool` and `Source string`
  (`cache|network|none`); `upgradeResult` gains
  `Source string json:"source,omitempty"`.
- `cmd/thermal/help.go`: add `printExitCodes` to the global and upgrade pages.
- `docs/CLI-CONTRACT.md`: canonical table with one example per code.

## Flag Disposition Matrix
| Flag | upgrade | reports | live | statusline |
| --- | --- | --- | --- | --- |
| `--json` | honored | honored | honored | honored |
| `--no-color` | honored | honored | honored | honored |
| `--verbose` | honored | honored | honored | honored |
| `--offline` | honored (cache) | honored | no-op | no-op |
| `--no-update-check` | no-op | honored | honored | honored |
| `--width` | not applicable (`--json`) | honored | honored | not applicable |
| `--fresh` | not applicable | no-op | honored | honored |

## Exit-Code Table
| Code | Meaning | Example |
| --- | --- | --- |
| `0` | success | `thermal version` |
| `1` | data/IO/network failure | `thermal opencode` with no DB |
| `2` | usage error | `thermal --tool open` |
| `3` | refused non-interactive mutation | `thermal upgrade --json` |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-home>

# Offline with cache
$B upgrade --check --offline --json >/tmp/o 2>/tmp/e
# expect exit 0 and "source":"cache", or exit 1 with the offline message
! grep -q 'unknown flag' /tmp/e

# Dead API proves no network
$B upgrade --check --offline --json 2>/tmp/e | jq -e .
test ! -s /tmp/e

# No-op and rejection
$B upgrade --check --no-update-check >/tmp/o 2>/tmp/e; test $? -eq 0
$B upgrade --width 100 >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'not applicable' /tmp/e
grep -q -- '--json' /tmp/e
! grep -q 'unknown flag' /tmp/e

# Unknown stays unknown
$B upgrade --bogus >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'unknown flag' /tmp/e

# Exit codes
$B version >/dev/null 2>&1; test $? -eq 0
$B opencode >/dev/null 2>&1; test $? -eq 1
$B --tool open >/dev/null 2>&1; test $? -eq 2
$B upgrade --json >/dev/null 2>&1; test $? -eq 3

# Help
$B --help | grep -q 'Exit codes: 0 ok'
$B upgrade --help | grep -q 'Exit codes: 0 ok'
$B upgrade --help | grep -q -- '--offline'

CGO_ENABLED=0 go test -race ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A cache fixture with a release index for `--offline --check`.
- A cache-less HOME for the offline failure.
- A dead API URL (`http://127.0.0.1:1`) to prove zero network attempts.
- A non-interactive stdin (`</dev/null`) for exit `3`.
- A PTY for the interactive upgrade prompt.
- The exit-code matrix commands.

## Open Questions
- Should `--offline` on upgrade use a cache or always fail?
  Recommendation: use a cache and mark `source:"cache"`; failing always makes
  `--offline` useless for the one command that needs the network.
- Should `--width` be a no-op instead of rejected?
  Recommendation: reject with `not applicable` + `--json`; silently ignoring a
  formatting flag is worse.
- Should exit code `3` be `1`?
  Recommendation: keep `3`; it distinguishes refusal from failure and is
  already documented.
- Should JSON include `exitCode`?
  Recommendation: no; the process exit code is the contract.
- Should `--offline --check` with no cache exit `1` or `2`?
  Recommendation: `1` (data/network unavailable), not a usage error.
