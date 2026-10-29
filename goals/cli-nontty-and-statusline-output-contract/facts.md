# Facts: Define machine-clean non-TTY stdout, exit codes, and silent statusline degradation

## Observed Evidence (2026-10-29, current source build `thermal dev`)
### `live`
- `thermal live >/tmp/o 2>/tmp/e` (stdout piped to a file):
  - stdout:
    ```
    thermal live is interactive and needs a terminal.
    For scripts, pipelines, and status bars use:

      thermal live --json            instant JSON snapshot
      thermal live --json --stream   stream NDJSON live events
      thermal stats                  daily distribution and token composition
    ```
  - stderr empty; exit `0`.
- `thermal live --json` → one valid JSON snapshot on stdout, exit `0`.
- `thermal live --json --stream | head -1` → one valid NDJSON line, no error text, exit `0`.
- In a 100x30 PTY, `thermal live` launches the TUI and requires `q`.

### `statusline`
- `HOME=$(mktemp -d) thermal statusline` → stdout empty, stderr `thermal: no supported tool data found`, exit `1`.
- `HOME=$(mktemp -d) thermal statusline --json` → stdout empty, same stderr, exit `1`.
- `HOME=$(mktemp -d) thermal statusline --plain` → stdout empty, same stderr, exit `1`.
- With data: stdout `🔥 80d all tools · 134.3M tok today · top: Muse`, exit `0`.
- `--plain` with data: `80d all tools · 134.3M tok today · top: Muse`, exit `0`.
- `--json` with data: `{"streak":80,"scope":"all tools","longestStreak":80,"todayTokens":134254118,"todayTurns":4481,"activeTool":"Muse","day":"2026-10-29","updatedAt":1793300729}`, exit `0`.

### `share`
- `HOME=$(mktemp -d) thermal share` → stdout empty, stderr `thermal: no supported tool data found`, exit `1`.
- With data: a stateless card and URL on stdout, exit `0`.

### Other
- `thermal daily | grep -c $'\x1b'` → `0` (color auto-disabled).
- `NO_COLOR=1 thermal daily | grep -c $'\x1b'` → `0`.
- `thermal --json` errors go to stderr with exit `1`, stdout empty.

## Root Cause
- `cmd/thermal/main.go:1717-1725`:
  ```go
  if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
      fmt.Println("thermal live is interactive and needs a terminal.")
      ...
      return
  }
  ```
  `fmt.Println` writes to stdout, and the bare `return` exits `0`.
- `cmd/thermal/main.go:1740-1744`: `runStatusline` checks the cache first, then calls `loadUsage(opts)`.
- `cmd/thermal/main.go:917` and `:1629`: `loadUsage` prints `thermal: no supported tool data found` and calls `os.Exit(1)` directly, so `runStatusline` cannot choose a quiet degradation.
- `cmd/thermal/main.go:758` has the same pattern for the leaderboard path.
- `internal/render/statusline.go:18-28`: `StatuslineData` has no `state` field; there is no no-data representation.
- The help calls statusline "Instant shell statusline and prompt helper (<3ms, alias: prompt)" but the failure mode is designed for a foreground command.

## Architectural Invariants & Constraints
- stdout is the data channel; stderr is the human/diagnostic channel. A command that produces no requested data must not write guidance to stdout.
- Exit `0` means the requested operation succeeded. A non-TTY `live` without a machine format did not produce live data, so it must not exit `0`.
- A prompt helper must never print an error on every shell redraw; silence is the correct degraded state.
- A deliberate command (`share`, `daily`, `stats`) may fail loudly; a prompt helper (`statusline`, `prompt`) may not.
- JSON and NDJSON must remain valid when stderr carries diagnostics.
- No ANSI on non-TTY streams; no color-only encodings.
- `statusline` must not trigger the update check or network I/O (already true).
- A no-data cache must not mask a newly created session for more than 60 seconds.

## File & Interface Contracts
- `cmd/thermal/main.go`:
  - `func runLive(opts thermal.Options)`: replace the non-TTY `fmt.Println` block with `fmt.Fprintln(os.Stderr, ...)` and `os.Exit(2)` (or return `2` to the dispatcher).
  - `func loadUsage(opts thermal.Options) usageSet`: change to `(usageSet, error)` so callers can degrade; update all call sites.
  - `func runStatusline(opts thermal.Options)`: on `ErrNoData`, render nothing (text) or the no-data JSON, exit `0`.
- `internal/render/statusline.go`:
  - add `State string \`json:"state"\`` to `StatuslineData`; values `ok`, `no-data`.
  - `RenderStatusline` returns `""` for no-data text, and the JSON object for no-data `--json`.
  - add `SaveStatuslineNoData(home string, ttl time.Duration)` or skip caching the no-data state.
- `cmd/thermal/statusline_test.go`, `cmd/thermal/live_test.go`: pipe/PTY/channel tests.
- `scripts/simulated_user_gate.sh`: channel/exit-code section.

## Channel & Exit Contract
| Surface | TTY | Data | Format | stdout | stderr | Exit |
| --- | --- | --- | --- | --- | --- | --- |
| `live` | no | any | text | empty | guidance | 2 |
| `live` | no | any | `--json` | JSON | empty | 0 (or 1 on load error) |
| `live` | no | any | `--json --stream` | NDJSON | empty | 0 |
| `live` | yes | any | text | TUI | — | 0/1 |
| `statusline` | no | no | text | empty | empty | 0 |
| `statusline` | no | no | `--json` | `state:"no-data"` | empty | 0 |
| `statusline` | no | yes | text | one line | empty | 0 |
| `statusline` | no | yes | `--json` | `state:"ok"` | empty | 0 |
| `share` | no | no | text | empty | message | 1 |
| `daily` | no | yes | text | table | empty | 0 |
| any | no | any | `--json` error | empty | message | 1/2 |

## Verification Contract
```bash
# live non-TTY
thermal live >/tmp/o 2>/tmp/e; test $? -eq 2
test ! -s /tmp/o
grep -q 'live needs a terminal' /tmp/e
grep -q -- '--json --stream' /tmp/e

# live JSON and NDJSON
thermal live --json | jq -e . >/dev/null; test $? -eq 0
thermal live --json --stream | head -1 | jq -e . >/dev/null; test $? -eq 0

# statusline no data
empty=$(mktemp -d)
HOME=$empty thermal statusline >/tmp/o 2>/tmp/e; test $? -eq 0
test ! -s /tmp/o; test ! -s /tmp/e
HOME=$empty thermal statusline --json | jq -e '.state=="no-data"' >/dev/null
HOME=$empty thermal statusline --plain >/tmp/o 2>/tmp/e; test $? -eq 0
test ! -s /tmp/o; test ! -s /tmp/e

# share remains loud
HOME=$empty thermal share >/tmp/o 2>/tmp/e; test $? -eq 1
test ! -s /tmp/o

# Channels and color
NO_COLOR=1 thermal statusline | grep -c $'\x1b'
NO_COLOR=1 thermal live --json | grep -c $'\x1b'

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Empty temp HOME.
- One-tool fixture HOME with a valid OpenCode DB.
- PTY harness (tmux or Python `pty`) for the TTY `live` case.
- Pipe harness with separate `StdoutPipe`/`StderrPipe`.
- A closed-pipe consumer (`head -1`) for SIGPIPE.
- A stale/unreadable statusline cache fixture.
- A no-data cache marker with a short TTL to prove recovery.

## Open Questions
- Should `live` non-TTY text exit `2` or `1`? Recommendation: `2` (usage: wrong invocation for the environment), documented in MIGRATION.
- Should `--stream` imply `--json`, or require both? Recommendation: imply NDJSON; test `thermal live --stream`.
- Should `statusline --json` emit no-data or exit `1`? Recommendation: emit `state:"no-data"` and exit `0`; scripts can inspect the state.
- Should the no-data state be cached? Recommendation: no cache, or a ≤60 s marker; a fresh session must appear on the next prompt.
- Should `statusline` print a placeholder such as `—`? Recommendation: no; empty output is the cleanest prompt integration. Offer `--json` for callers that need a value.
