# Facts: Make window and repeated-flag semantics explicit and non-silent

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

All runs used the fixture HOME and `--offline --no-update-check`.

### One flag, three units
```
$ thermal weekly --last=7
exit 0
Window: [last 7 weeks · all tools]  Total: [1.1M tok · $7.00]

$ thermal monthly --last 7
exit 0
Window: [last 7 months · all tools]  Total: [1.1M tok · $7.00]

$ thermal daily --last 7
exit 0
Window: [last 7 days · all tools]  Total: [1.1M tok · $7.00]
```

### The error contradicts weekly/monthly
```
$ thermal daily --last
exit 2; stdout empty
thermal: --last needs a value (a whole number of days)
  example: thermal daily --last 7
```

### Repeated scalar flag is silent
```
$ thermal --tool claude --tool codewhale
exit 0; stderr empty
  ╭─ codewhale Activity ─────────────── ~/.codewhale/sessions ─╮
  │  1.0M tokens / 52 weeks · 4 active days · 2 day streak      │
```
`--tool claude` is discarded with no warning.

### `--` produces an "unknown tool"
```
$ thermal weekly -- --offline
exit 2; stdout empty
thermal: unknown tool "--offline"
  Valid tools: OpenCode, Devin, Codex, Claude, ZCode, Grok, MiMoCode,
  codewhale, DeepSeek (DSH), Nous Hermes, CMD, Zed, Agy, Muse, Droid
  run thermal --help for supported tools
```

### Placeholder examples break when pasted
```
$ thermal weekly --tool opencode --db ""
exit 2; stdout empty
thermal: --db requires a non-empty value
  example: thermal --tool opencode --db <path> daily
```
In `sh`, `<path` is an input redirect and `>` creates a file named `path`.

## Root Cause
- `cmd/thermal/args.go:72` defines
  `"last": {kindInt, "a whole number of days", "thermal daily --last 7"}`.
  The metadata is global, but the value is interpreted per report grain.
- `cmd/thermal/args.go:393` registers `--last` as an `IntVar`; there is no unit
  field and no suffix parsing.
- `internal/thermal/report.go:193-205` computes `lastStart` from the report
  grain: days for `GrainDay`, `7*(N-1)` days for `GrainWeek`, and `N-1` months
  for `GrainMonth`.
- `internal/thermal/format.go:66-89` (`FormatReportWindow`) prints the grain
  unit, so the output is correct even though the error text is not.
- `cmd/thermal/args.go:88-186` (`normalizeArgs`) appends every occurrence of a
  value flag to `flagArgs`; the stdlib `flag` package keeps the last value.
  `present` is a `map[string]bool`, so repeats are indistinguishable from one
  occurrence.
- `cmd/thermal/args.go:93-96`: after `--`, every token is appended to
  `positionals`; the dispatcher later classifies a flag-like token as an
  unknown tool.
- `cmd/thermal/args.go:67` (`db` example) contains `<path>`.

## Architectural Invariants & Constraints
- A bare `--last N` keeps its current per-report period meaning; this is
  compatibility, not a bug.
- An explicit suffix is absolute: `14d` is 14 days on every report.
- The parser resolves the window; aggregators must not multiply a count by a
  grain afterward.
- A repeated scalar flag never changes scope silently; the last value wins and
  stderr names the discarded value.
- Boolean repeats are idempotent; no warning.
- `--` is POSIX: everything after it is positional. The diagnostic must explain
  the rule, not pretend the token is a tool.
- Every copyable example must be shell-safe. Angle brackets are not allowed in
  `example:` lines.
- JSON stdout is never changed by a warning; warnings go to stderr.
- `--last` and `--since`/`--until` remain mutually exclusive.

## File & Interface Contracts
- `cmd/thermal/args.go`:
  ```go
  type LastSpec struct {
      N        int
      Unit     string // day|week|month|release
      Explicit bool
  }

  func parseLast(raw string, reportUnit string) (LastSpec, error)
  ```
  `parseLast("7", "week")` -> `{7, "week", false}`;
  `parseLast("14d", "week")` -> `{14, "day", true}`;
  `parseLast("3m", "month")` -> `{3, "month", true}`.
- `thermal.Options` gains `LastUnit string`, `LastExplicit bool`, or a
  `LastSpec`; keep `Last int` for callers that only need the count.
- `internal/thermal/report.go`:
  ```go
  func LastWindowStart(now time.Time, n int, unit string,
      startOfWeek time.Weekday) time.Time
  ```
  Used by `Aggregate`, `FilterDailyRows`, `AggregateStats`, `AggregateTrend`,
  and `AggregateToolMix`; no grain-specific multiplication remains.
- `cmd/thermal/args.go` (`normalizeArgs`): `present map[string]int`; return
  parse warnings so `main` can emit them after the color decision.
- `cmd/thermal/main.go`: `emitParseWarnings(w, present, opts)` writes one line
  per repeated singleton to stderr.
- `cmd/thermal/help.go` and `flagMetas`: suffix-aware help and shell-safe
  examples.

## Unit Matrix
| Invocation | Resolved unit | Window label |
| --- | --- | --- |
| `daily --last 7` | day | `last 7 days` |
| `weekly --last 7` | week | `last 7 weeks` |
| `monthly --last 7` | month | `last 7 months` |
| `daily --last 7w` | week | `last 7 weeks` |
| `weekly --last 14d` | day | `last 14 days` |
| `monthly --last 3m` | month | `last 3 months` |
| `changelog --last 5` | release | `5 of 25 releases` |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-home>

# Bare units
HOME=$H $B weekly --last 7 | grep -q 'last 7 weeks'
HOME=$H $B monthly --last 7 | grep -q 'last 7 months'
HOME=$H $B daily --last 7 | grep -q 'last 7 days'

# Explicit suffixes
HOME=$H $B weekly --last 14d | grep -q 'last 14 days'
HOME=$H $B monthly --last 3m | grep -q 'last 3 months'
HOME=$H $B daily --last 7w | grep -q 'last 7 weeks'

# Invalid forms
HOME=$H $B weekly --last 7x >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'd, w, or m' /tmp/e
HOME=$H $B daily --last >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'weekly, 7 means 7 weeks' /tmp/e

# Repeated scalars
HOME=$H $B --tool claude --tool codewhale >/tmp/o 2>/tmp/e; test $? -eq 0
grep -q 'codewhale' /tmp/o
grep -q -- '--tool given twice' /tmp/e
HOME=$H $B --json --json >/tmp/o 2>/tmp/e; test ! -s /tmp/e

# Double dash
HOME=$H $B weekly -- --offline >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q 'passed after --' /tmp/e
! grep -q 'Valid tools:' /tmp/e

# Examples are shell-safe
HOME=$H $B weekly --tool opencode --db "" 2>&1 | grep -o 'example: .*' | sh -n
! grep -R '<path>' cmd/thermal/help.go cmd/thermal/args.go

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Fixture HOME with four codewhale days and three Claude days.
- Frozen clock (`2026-10-30`) for window boundary assertions.
- A report with one row per day for two months, so `14d` and `2w` coincide.
- Repeated-flag matrix for every singleton and every boolean.
- `--` cases: flag-like token, known tool, known command, empty.
- A shell parser (`sh -n`) for every `example:` line.

## Open Questions
- Should `--last 7m` mean months or minutes? Recommendation: months in this
  flag; `--interval` owns minutes.
- Should repeated scalars be an error instead of a warning?
  Recommendation: warn and keep the last value for compatibility; a future
  major release may make it an error.
- Should the suffix be required? Recommendation: no; bare integers are the
  common case and keep working.
- Should `changelog --last` accept suffixes? Recommendation: no; it counts
  releases, not periods. Document the exception.
- Should the warning appear once per flag or once per run?
  Recommendation: once per repeated flag, with the last value.
