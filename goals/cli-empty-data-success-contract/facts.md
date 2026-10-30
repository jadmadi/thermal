# Facts: Make no-data a successful, schema-stable state across text, JSON, and statusline

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

All runs used a fresh HOME, `--offline --no-update-check`, and captured
stdout/stderr separately.

### Bare invocation, empty HOME
```
$ HOME=<empty> thermal
exit 1
stdout: <empty>
stderr:
  Thermal · sources · 15 supported tools

  TOOL           ALIAS PATH                   STATE        DETAIL
  ---------------------------------------------------------------------
  MiMoCode       mimo  …/mimocode/mimocode.db missing      expected file
  ...
  Summary: 0 found · 0 empty · 15 missing · 0 unreadable
  Legend: found=data · empty=no sessions · missing=path absent
          unreadable=permissions · wrong-shape=unexpected file/directory

  Next: run one prompt in any supported tool, then re-run thermal.
        thermal sources --json   machine-readable inventory
        thermal --help           supported tools and flags
```
29 lines on stderr; zero bytes on stdout.

### JSON, empty HOME
```
$ HOME=<empty> thermal --json
exit 1
stdout:
{"type":"sources","generatedAt":"...","home":"...",
 "summary":{"total":15,"found":0,"empty":0,"missing":15,...},
 "sources":[...15 entries...]}
$ jq -r .type        -> sources
$ jq -c '.ranking'   -> null
```

### Sources, empty HOME
```
$ HOME=<empty> thermal sources        -> exit 0, table on stdout
$ HOME=<empty> thermal sources --json -> exit 0, type:"sources"
```

### Statusline, empty HOME
```
$ HOME=<empty> thermal statusline          -> exit 0, stdout 0 B, stderr 0 B
$ HOME=<empty> thermal statusline --json   -> exit 0, {"state":"no-data",...}
$ HOME=<empty> thermal statusline --plain  -> exit 0, stdout 0 B
```

### Populated HOME
```
$ thermal --json | jq -r .type -> leaderboard
$ thermal statusline           -> "🔥 5d all tools · 0 tok today · top: codewhale"
```

## Root Cause
- `cmd/thermal/sources.go:17` defines `errNoData` as a sentinel error.
- `cmd/thermal/main.go:737-739` returns `errNoData` when all-tools mode has no
  day or project rows.
- `cmd/thermal/main.go:773-780` (`loadUsageOrExit`) calls
  `printEmptyInventory` and `os.Exit` for every foreground report.
- `cmd/thermal/sources.go:41-52` (`printEmptyInventory`) prints the 15-row
  inventory to stderr and returns `1`; for `--json` it prints the `sources`
  envelope to stdout and returns `1`. The requested verb is not consulted.
- `cmd/thermal/main.go:1637-1654` (`runStatusline`) constructs a no-data
  `StatuslineData` and prints `RenderStatusline`; `internal/render/statusline.go`
  returns `""` for `State=="no-data"` in text mode.
- `cmd/thermal/json.go` has no shared no-data envelope; each report marshals its
  own structure, so a generic fallback is currently a schema swap.
- No `--require-data` flag exists in `cmd/thermal/args.go`.

## Architectural Invariants & Constraints
- stdout is the data channel; stderr is the diagnostic channel. A successful
  read-only command writes its requested payload to stdout.
- "No data" is a state of the data, not a failure of the command. The command
  succeeded; it found zero rows.
- The requested envelope type is stable: `thermal --json` is always a
  leaderboard object, `thermal daily --json` always a daily object.
- The inventory is a diagnostic surface (`sources`), not the default payload.
- A shell prompt helper must remain silent in text mode; it may expose state
  through `--json`.
- Exit codes remain distinguishable: `0` success, `1` explicit failure,
  `2` usage, `3` refused.
- `--require-data` is the only way to make "no data" exit non-zero; it must not
  change the payload.
- No network, no pricing fetch, no update check, no telemetry on the no-data
  path.
- The onboarding block must not leak prompts, session titles, project names, or
  file contents.

## File & Interface Contracts
- `internal/thermal` (or `cmd/thermal/json.go`):
  ```go
  type NoDataNext struct {
      Inventory string `json:"inventory"`
      Help      string `json:"help"`
  }
  ```
  Every report envelope gains:
  ```go
  State string      `json:"state,omitempty"` // "no-data" only
  Next  *NoDataNext `json:"next,omitempty"`
  ```
  `sources` never gains either field.
- `cmd/thermal/main.go`:
  ```go
  func renderNoData(verb string, items []thermal.SourceInfo) string
  func noDataEnvelope(verb string) (state string, next *NoDataNext)
  func loadUsageOrExit(opts thermal.Options) usageSet // honors RequireData
  ```
- `cmd/thermal/args.go`: add
  `fs.BoolVar(&opts.RequireData, "require-data", false, "Exit 1 when no data is found")`.
- `cmd/thermal/sources.go`: `printEmptyInventory` is replaced by
  `printNoData`; `--verbose` appends `RenderSources` to stderr.
- `internal/render/statusline.go`: no change to the empty return; add a doc
  comment and a test. `StatuslineOptions` unchanged.

## No-Data Matrix
| Surface | Text stdout | Text stderr | JSON stdout | Exit | `--require-data` |
| --- | --- | --- | --- | --- | --- |
| bare `thermal` | onboarding, <=12 lines | empty | leaderboard, `state` | 0 | 1 |
| report word | onboarding | empty | verb envelope, `state` | 0 | 1 |
| `sources` | inventory | empty | `type:"sources"` | 0 | 0 |
| `statusline` | empty | empty | `state:"no-data"` | 0 | 0 |
| `share` | empty | message | error object | 1 | 1 |
| `live` non-TTY | empty | guidance | snapshot/NDJSON | 2 | 2 |

## Text Layout (80 columns)
```
  No AI coding sessions found yet.
  Run one prompt in OpenCode, Claude, Codex, or any supported tool, then
  re-run thermal.

  Next: thermal sources   inspect all 15 source paths and states
        thermal --help    supported tools and flags
```
- No leading blank line; no trailing table.
- `--verbose` appends the sources table to stderr, separated by one blank line.
- At `COLUMNS=40`, wrap to <= 40 columns; never truncate a command name.
- Zero ANSI under `--no-color` or `NO_COLOR`.

## JSON Contract
```json
{
  "type": "leaderboard",
  "generatedAt": "2026-10-30T00:00:00Z",
  "state": "no-data",
  "ranking": {"key": "streak", "order": "desc"},
  "results": [],
  "totals": {"tokens": 0, "cost": 0, "sessions": 0},
  "unpricedModels": [],
  "next": {
    "inventory": "thermal sources --json",
    "help": "thermal --help"
  }
}
```
- `type` is the requested verb, never `sources`.
- `state` is absent when data exists; do not emit `"ok"` on normal payloads
  unless the existing schema already has it.
- `next` appears only with `state:"no-data"`.
- `sources --json` remains unchanged and has no `state`/`next`.
- `statusline --json` keeps its existing top-level keys plus `state`.

## Verification Contract
```bash
B=/tmp/thermal-cli-review
empty=$(mktemp -d)
fixture=$(mktemp -d)
mkdir -p "$fixture/.claude/projects/p"
printf '%s\n' '{"type":"assistant","timestamp":"2026-10-30T09:00:00Z",
"message":{"id":"m1","model":"claude-3-5-haiku",
"usage":{"input_tokens":100,"output_tokens":50}}}' \
  > "$fixture/.claude/projects/p/s.jsonl"

# Text onboarding
HOME=$empty $B >/tmp/o 2>/tmp/e; test $? -eq 0
test -s /tmp/o; test ! -s /tmp/e
grep -q 'No AI coding sessions found yet' /tmp/o
grep -q 'thermal sources' /tmp/o
test "$(wc -l </tmp/o)" -le 12

# Stable JSON type
HOME=$empty $B --json >/tmp/j 2>/tmp/e; test $? -eq 0
jq -e '.type=="leaderboard" and .state=="no-data"
       and (.ranking|length)==0 and .next.inventory' /tmp/j
HOME=$empty $B daily --json | jq -e '.type=="daily" and .state=="no-data"'

# Strict mode
HOME=$empty $B --require-data >/tmp/o 2>/tmp/e; test $? -eq 1
grep -q 'No AI coding sessions found yet' /tmp/o
grep -q -- '--require-data' /tmp/e

# Sources stays sources
HOME=$empty $B sources --json | jq -e '.type=="sources" and (has("state")|not)'

# Statusline
HOME=$empty $B statusline >/tmp/o 2>/tmp/e; test $? -eq 0
test ! -s /tmp/o; test ! -s /tmp/e
HOME=$empty $B statusline --json | jq -e '.state=="no-data" and .todayTokens==0'
HOME=$empty $B statusline --verbose >/tmp/o 2>/tmp/e; test $? -eq 0
test ! -s /tmp/o; test "$(wc -l </tmp/e)" -eq 1

# Populated control
HOME=$fixture $B --json | jq -e '.type=="leaderboard" and (has("state")|not)'

# Channel purity and privacy
HOME=$empty $B --json --verbose 2>/tmp/e | jq -e . >/dev/null
grep -q 'SENTINEL' /tmp/o && exit 1

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Empty temp HOME.
- One-tool HOME with a valid Claude JSONL file.
- HOME with an empty `~/.claude` directory (installed, no sessions).
- HOME with an unreadable source directory (skip as root).
- HOME with a sentinel prompt string.
- `--width 40`, `--width 80`, and `COLUMNS=40`.
- A stale statusline cache file.
- An unreadable statusline cache file.

## Open Questions
- Should bare `thermal` exit `0` or `1` when no tool is installed?
  Recommendation: `0`; absence of data is not a command failure. Offer
  `--require-data` for strict callers.
- Should `thermal --json` on empty data keep `type:"sources"` for one release?
  Recommendation: no; the schema swap is the defect. Migrate with a MIGRATION
  entry and `--require-data`.
- Should `statusline` print a placeholder such as `— no data`?
  Recommendation: no; silence is the safest prompt contract. Document it and
  provide `--json` for callers that need a value.
- Should `share` join the success contract?
  Recommendation: no; it is an explicit command that promises a card and
  cannot produce one without data.
- Should `--require-data` also suppress the onboarding block?
  Recommendation: no; keep stdout identical and change only exit/stderr so
  pipelines see the same payload.
