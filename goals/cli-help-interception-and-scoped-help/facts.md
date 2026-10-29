# Facts: Never swallow -h/--help and add command-scoped help

## Observed Evidence (2026-10-29, current source build `thermal dev`)
- `thermal daily --help` prints the standard daily report header (`Thermal · daily usage · all tools`), a window line, and the `Daily Usage` table; exit code `0`.
- `thermal daily -h` behaves identically.
- `thermal live --help` in a 100x30 tmux PTY launches `THERMAL LIVE`, renders the KPI grid, and waits for input; it must be killed with `q` or Ctrl-C.
- `thermal live --help` with stdout piped prints the non-TTY pointer (`thermal live is interactive and needs a terminal.`) and exits `0`.
- `thermal upgrade --help` on a disposable copy of the dev binary performs a real upgrade: output includes `downloading thermal_0.21.0_linux_amd64.tar.gz` and `✓ upgraded to v0.21.0`; md5 changes from `43110b4…` to `ced2913…`.
- `thermal help daily` prints `thermal: unknown tool: help` to stderr and exits `1`.
- `thermal --help` is the only working help form: an 84-line global page, exit `0`, stdout.
- `thermal --help` does not list `-h/--help` precedence rules, per-command pages, or the `help` pseudo-verb.

## Root Cause
- `cmd/thermal/main.go:47-71`: help/version/license are special-cased only when `len(os.Args) == 2`. `arg == "help"` is accepted only in that single-argument branch.
- `cmd/thermal/main.go:109-112`: `flag.Usage` is set, but Go's `flag.Parse()` stops at the first non-flag argument. For `thermal daily --help`, `daily` is the first positional, so `--help` is never parsed as a flag.
- `cmd/thermal/main.go:123-273`: the post-positional scanner handles a fixed switch of long flags; there is no `-h`/`--help` case, so the token is silently ignored and dispatch proceeds.
- `cmd/thermal/main.go:299-311`: any first positional that is not a report word becomes `opts.Tool`; `help` becomes a tool name and fails in `loaders.ResolveTool`.
- `internal/render/help.go:34`: `RenderHelp(noColor bool)` renders one global page and has no command parameter or command registry.
- `cmd/thermal/main.go:585-586`: `case "upgrade": os.Exit(runUpgrade())` has no help guard; `runUpgrade` starts printing and performing network work immediately.

## Architectural Invariants & Constraints
- Help is a read-only, side-effect-free surface: no loader, no pricing catalog load, no update check, no cache write, no TUI, no network, no binary mutation.
- Help must never be a synonym for "run the command with default/ignored arguments".
- The command vocabulary, flag vocabulary, and dispatch must have one source of truth. `isReportWord` (`cmd/thermal/main.go:330-336`) is currently a duplicate list that already disagrees with validation messages (it includes `replay`, `yield`, `receipt`, `web`, `serve`, `changelog`, `live`, `statusline`, `prompt`).
- Help must be usable on a fresh machine with no agent data; it must not fail because scanning failed.
- Existing content guarantees from `help-supported-tools-and-flags` (15 tools, `--sort` values, `--start-of-week <day>`) must be preserved or improved, never regressed.
- `NO_COLOR` and non-TTY stdout always disable ANSI; help is no exception.

## File & Interface Contracts
- `cmd/thermal/main.go`:
  - `parseArgs() thermal.Options` — add a help short-circuit before flag registration/parse.
  - `isReportWord(s string) bool` — replace the literal switch with a registry lookup.
  - `main()` — intercept help before `runUpgrade`, `runLive`, `runStatusline`, and report dispatch.
- `cmd/thermal/help.go` (new):
  - `type commandSpec struct { Name, Aliases []string; Summary, Usage string; Flags []flagSpec; Examples []string; JSONKeys []string; Related []string; AcceptsTool bool }`
  - `func resolveHelp(args []string) (spec commandSpec, ok bool, unknown string)`
  - `func renderCommandHelp(spec commandSpec, noColor bool) string`
- `internal/render/help.go`:
  - keep `RenderHelp(noColor bool) string` for global help;
  - add `RenderCommandHelp(title, usage, description string, sections []Section, noColor bool) string`.
- `cmd/thermal/upgrade.go`:
  - `runUpgrade() int` must not be reachable from a help path; optionally add `runUpgradeWithOptions(opts upgradeOptions) int`.
- `scripts/simulated_user_gate.sh`:
  - add a `Help Precedence & Side-Effect Freedom` section.

## Output Contract (scoped help)
```
  Thermal · daily · daily token volume and cost breakdown

  Usage:
    thermal daily [tool] [options]

  Flags:
    --since <date>       Start date (YYYY-MM-DD or YYYYMMDD)
    --until <date>       End date (YYYY-MM-DD or YYYYMMDD)
    --last <N>           Limit to the last N days
    --breakdown          Add per-model rows
    --chart              Print a bar chart under the table
    --json               Emit the daily JSON envelope
    --no-color           Disable ANSI colors

  Examples:
    thermal daily --last 7
    thermal daily opencode --since 2026-10-01 --json

  JSON:
    keys: type, window, data, totals, costBasis, generatedAt

  Exit codes:
    0 success · 1 validation/data error · 2 flag parse error

  See also: thermal weekly · thermal stats
```
- Every page must end with a single newline.
- No line may exceed 80 columns in the default rendering.
- The page must not contain table borders or date rows that imply real data.

## Verification Contract
```bash
# 1. Interception matrix
for v in daily weekly monthly projects models stats mix trend live statusline share audit changelog upgrade version license; do
  out=$(thermal "$v" --help 2>/dev/null) || { echo "FAIL $v"; exit 1; }
  printf '%s' "$out" | grep -q "Usage:" || { echo "FAIL usage $v"; exit 1; }
done

# 2. Alias parity
cmp <(thermal daily --help) <(thermal help daily)

# 3. No side effects
tmp=$(mktemp -d); marker="$tmp/marker"; touch "$marker"
HOME="$tmp" thermal upgrade --help >/dev/null
[ -z "$(find "$tmp" -newer "$marker" -type f)" ]

# 4. No TUI / no network
HTTPS_PROXY=http://127.0.0.1:9 timeout 2 thermal live --help >/dev/null
! HTTPS_PROXY=http://127.0.0.1:9 timeout 2 thermal live --help | grep -q 'LIVE'

# 5. Unknown context
thermal help nope >/tmp/help-nope.out 2>/tmp/help-nope.err; test $? -eq 1
grep -q 'unknown command' /tmp/help-nope.err

# 6. Unit + gate
CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A temp HOME with zero agent directories.
- A read-only `~/.cache/thermal` directory.
- A disposable binary copy (`/tmp/thermal-help-copy`) for the upgrade-help md5 check.
- A PTY harness (tmux or Python `pty` + `TIOCSWINSZ`) that captures bytes and asserts the absence of `\x1b[?1049h`.
- A table of `(argv, expectedExit, expectedSubstring, forbiddenSubstring)` covering the negative cases in `goal.md`.

## Open Questions
- Should `thermal help daily extra` ignore the extra token or exit `1`? Decision: ignore it; help wins over validation (precedence rule), so `thermal daily --help extra` and `thermal help daily extra` both print the daily page and exit `0`. Document this in the help test.
- Should `thermal --help --bogus` ignore the unknown flag because help wins, or report the flag error first? Recommendation: help wins; document the precedence.
- Should retired verbs (`dashboard`, `serve`, `replay`) get help pages or the retirement message? Recommendation: retirement message plus `See also: thermal live`, exit `1`.
