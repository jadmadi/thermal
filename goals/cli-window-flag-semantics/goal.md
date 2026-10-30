# Goal: Make window and repeated-flag semantics explicit and non-silent

## Goal Description
Priority: P1 (correct numbers). `--last` is one flag with three meanings:
`daily --last 7` is seven days, `weekly --last 7` is seven weeks, and
`monthly --last 7` is seven months. The value error calls it "a whole number of
days", which contradicts the weekly and monthly results. Separately, a repeated
scalar flag silently keeps the last value, `--` produces a confusing
"unknown tool", and the `--db` examples contain a shell-breaking `<path>`
placeholder.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`):

- `thermal weekly --last=7` -> exit `0`, window `last 7 weeks`.
- `thermal monthly --last 7` -> exit `0`, window `last 7 months`.
- `thermal daily --last 7` -> exit `0`, window `last 7 days`.
- `thermal daily --last` -> exit `2`:
  `thermal: --last needs a value (a whole number of days)`
  `  example: thermal daily --last 7`
- `thermal --tool claude --tool codewhale` -> exit `0`, a codewhale report,
  no warning; Claude silently disappears.
- `thermal weekly -- --offline` -> exit `2`:
  `thermal: unknown tool "--offline"` plus the 15-tool list.
- `thermal weekly --tool opencode --db ""` -> exit `2`:
  `example: thermal --tool opencode --db <path> daily`
  Pasting that example into a shell treats `<path` as a redirect.

The outcome of this goal is a **window and flag semantics contract**:

1. `--last` accepts an explicit unit (`7d`, `4w`, `3m`) that means the same
   thing on every report, while a bare integer keeps today's per-report period
   meaning and is labeled with that unit in help and errors.
2. Repeated scalar flags warn on stderr and keep the last value; boolean
   repeats are idempotent and silent.
3. `--` explains that flag parsing stopped and names the positional it
   produced.
4. Every example is shell-safe and runnable, with no `<placeholder>` tokens.

## Dependencies & Execution Order
- **Mode**: Dependent (prerequisite goals are implemented)
- **Depends On**: `cli-argument-parser-strictness`
- **Sequence**: 257
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3

Related goals:
- `cli-report-total-row-semantics` and `output-report-totals-reconciliation`
  own totals for a window; this goal owns how the window is expressed.
- `cli-help-glossary-and-routing` owns unknown-token classification; this goal
  owns the `--` explanation and example runnability.
- `cli-empty-data-success-contract` owns the no-data outcome; this goal keeps
  the window label correct in that state.

## Files to Touch
- `cmd/thermal/args.go`
- `cmd/thermal/args_test.go`
- `cmd/thermal/args_strict_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/help.go`
- `internal/thermal/report.go`
- `internal/thermal/format.go`
- `internal/thermal/format_test.go`
- `internal/thermal/report_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-window-flag-semantics/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-window-flag-semantics/plan.md`](plan.md)
- Prior contract:
  [`goals/cli-argument-parser-strictness/goal.md`](../cli-argument-parser-strictness/goal.md)

## Requirements

1. **Explicit window units.**
   - Accept `--last N` (bare) and `--last Nsuffix` where suffix is `d`, `w`,
     or `m`, case-insensitive.
   - Bare `N` keeps the report's period unit:
     `daily` = days, `weekly` = weeks, `monthly` = months.
   - `Nd` always means N calendar days, even on `weekly`/`monthly`;
     `Nw` always means N weeks; `Nm` always means N calendar months.
   - `--last=7w` and `--last 7w` are equivalent. `--last 7 w` is not accepted.
   - `--last 0`, `--last -1`, `--last 1.5`, and an unknown suffix are usage
     errors with a range and an example.
   - The resolved unit appears in the report window label (`last 7 weeks`,
     `last 14 days`, `last 3 months`).
   - `--last` remains mutually exclusive with `--since`/`--until`; the existing
     error stays.
   - `--last` on `changelog` keeps its "number of releases" meaning and is
     documented as an exception; do not silently reinterpret it as days.

2. **Window model.**
   - Add an explicit unit to the resolved options:
     ```go
     type LastSpec struct {
         N        int
         Unit     string // day|week|month|release
         Explicit bool   // true when the user wrote a suffix
     }
     ```
   - `thermal.Options` carries `LastSpec` (or `Last int` plus `LastUnit`)
     alongside the existing `Last` for compatibility.
   - `AggregateOptions`/`TrendOptions`/`StatsOptions`/`MixOptions` receive a
     resolved start date, not a bare count, or a `LastSpec` they all interpret
     identically. One helper computes the start:
     ```go
     func lastWindowStart(now time.Time, spec LastSpec, grain Grain,
         startOfWeek time.Weekday) time.Time
     ```
   - No report may multiply a bare count by its grain after the parser has
     resolved the unit.
   - `FormatReportWindow` prints the unit from the spec, not from the grain.

3. **Help and errors.**
   - Global help line:
     `--last <N>[d|w|m]   Last N periods; bare N uses the report's unit`.
   - Each report page shows the unit:
     `daily: days · weekly: weeks · monthly: months`.
   - The missing/invalid error names the resolved unit:
     `--last needs a whole number; on weekly, 7 means 7 weeks (or use 7d)`.
   - Examples:
     `thermal weekly --last 4` (last 4 weeks),
     `thermal weekly --last 14d` (last 14 days),
     `thermal monthly --last 3` (last 3 months).
   - `--last` with an explicit suffix and a report whose grain differs must not
     warn; the suffix is unambiguous.

4. **Repeated scalar flags.**
   - Maintain a singleton list: `tool`, `db`, `sort`, `order`, `metric`, `by`,
     `grain`, `format`, `width`, `weeks`, `since`, `until`, `last`, `top`,
     `start-of-week`, `interval`, `flame-ceiling`.
   - When a singleton appears more than once, keep the last value (compatible)
     and write one stderr line:
     `thermal: warning: --tool given twice; using "codewhale" (last)`.
   - Boolean repeats (`--json`, `--no-color`, `--verbose`, `--chart`,
     `--breakdown`, `--offline`, `--no-estimate`, `--fresh`, `--plain`,
     `--nerd`, `--stream`) are idempotent and silent.
   - `--json` output is unchanged; the warning goes to stderr only.
   - `--verbose` does not duplicate the warning.
   - The warning is suppressed under `--json` only if it would break a strict
     consumer; recommendation: keep it on stderr, never stdout.

5. **`--` behavior.**
   - After `--`, every token is positional. If the first positional starts with
     `-`, print:
     ```
     thermal: "--offline" was passed after --, so it is a positional, not a flag
       put flags before --, or drop -- to let thermal parse them
     ```
     Exit `2`.
   - If the positional is a valid tool or command, execute it as today.
   - Do not print the 15-tool list for a token that looks like a flag.
   - `--` with no following token is a no-op, exit as the command would.

6. **Shell-safe examples.**
   - Replace `<path>` in `flagMetas` and help with a real path:
     `"$HOME/.local/share/opencode/opencode.db"`.
   - Quote paths with spaces. Do not use angle brackets, `...`, or `{}` in a
     copyable command.
   - If a placeholder is unavoidable, use an env var with a default:
     `--db "${OPENCODE_DB:-$HOME/.local/share/opencode/opencode.db}"`.
   - A test extracts every `example:` line from every usage error and runs it
     with `sh -n` (syntax check) at minimum; a second test runs the help
     examples against fixtures where feasible.

7. **Tests.**
   - Table over `--last` forms × reports: `7`, `7d`, `4w`, `3m`, `0`, `-1`,
     `1.5`, `7x`, `=7w`.
   - Assert the window label and the first/last included day for each.
   - Repeated-flag table for every singleton; assert one warning and the last
     value.
   - Boolean repeat table; assert no warning.
   - `--` table: flag-like positional, tool positional, command positional,
     empty.
   - Example syntax test with `sh -n`.
   - Gate: `simulated_user_gate.sh` covers `weekly --last 7` vs
     `weekly --last 14d` and the repeated `--tool` warning.

8. **Documentation and migration.**
   - README "Reports" shows bare and suffixed forms.
   - MIGRATION entry for the new suffix (additive) and the repeated-flag
     warning (stderr-only). No breaking change for bare integers.
   - `docs/MIGRATION.md` detection: `thermal weekly --last 7` window label.
   - `project.jsonc` updated with the `--last` form.

## Non-Goals
- Changing the default window (all time) or the meaning of an active period.
- Changing `--since`/`--until` parsing.
- Making repeated boolean flags an error.
- Adding a config file for default windows.
- Changing `changelog --last` semantics.

## Acceptance Criteria
- AC-001: `thermal weekly --last 7` exits `0` and the window says
  `last 7 weeks`.
- AC-002: `thermal weekly --last 14d` exits `0` and the window says
  `last 14 days`, with the same rows as `daily --last 14` for the overlapping
  range.
- AC-003: `thermal monthly --last 3m` and `thermal monthly --last 3` produce
  the same window and totals.
- AC-004: `thermal daily --last` exits `2` with a unit-aware message and a
  runnable example.
- AC-005: `thermal weekly --last 7x` exits `2` with the accepted suffixes.
- AC-006: `thermal --tool claude --tool codewhale` exits `0`, reports
  codewhale, and writes exactly one stderr warning naming both values.
- AC-007: Repeating a boolean flag produces no warning.
- AC-008: `thermal weekly -- --offline` exits `2` with the `--` explanation,
  not a 15-tool list.
- AC-009: Every `example:` line from a usage error passes `sh -n`.
- AC-010: No help or error text contains `<path>`.
- AC-011: `go test -race ./cmd/thermal/... ./internal/thermal/...` exits `0`,
  including `TestLastUnitMatrix`, `TestRepeatedScalarFlags`,
  `TestDoubleDashDiagnostics`, and `TestExamplesShellSafe`.
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the
  window and repeated-flag cases.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `weekly --last 7` | 7 weeks |
| `weekly --last 14d` | 14 days |
| `monthly --last 3m` | 3 months |
| `daily --last 7w` | 7 weeks |
| `--last 0` | exit 2, range + unit |
| `--last -1` | exit 2, range + unit |
| `--last 1.5` | exit 2, whole number |
| `--last 7x` | exit 2, accepted suffixes |
| `--last` | exit 2, unit-aware example |
| `--last 7 --since ...` | existing conflict, exit 2 |
| repeated `--tool` | one stderr warning, last wins |
| repeated `--json` | silent, idempotent |
| `weekly -- --offline` | `--` explanation, exit 2 |
| `--` alone | no-op |
| `<path>` in any example | test fails |

## Metrics & Regression Guards
- Unit correctness: every `--last` form has a fixed expected first/last day in
  a table test with a frozen clock.
- Singleton coverage: one test per flag in the singleton list.
- Warning count: exactly one per repeated singleton, zero for booleans.
- Example safety: 100% of `example:` lines pass `sh -n`; no angle brackets.
- Gate: simulated user gate asserts AC-001, AC-002, AC-006, and AC-008.

## Risks & Mitigations
- **Risk:** Adding suffixes changes a de facto numeric contract.
  **Mitigation:** bare integers keep their exact behavior; suffixes are
  additive.
- **Risk:** Repeated-flag warnings pollute stderr in scripts.
  **Mitigation:** stderr is the diagnostic channel; `--json` stdout is
  untouched. Do not warn for boolean repeats.
- **Risk:** `--last 7m` is ambiguous with minutes.
  **Mitigation:** document `m` as month in this flag; minutes are only
  `--interval`.
- **Risk:** The window model refactor touches every report.
  **Mitigation:** resolve to a start date once, then pass the same date to all
  aggregators; one table test compares totals across forms.
- **Risk:** `changelog --last` is a different concept.
  **Mitigation:** document it as a release count and keep its behavior; do not
  route it through `LastSpec`.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "--last weeks months"` and
  `sila recall "repeated flag last wins"`.
- **Writes**: record a `case` for the "explicit window spec with per-report
  default unit" mechanism; record a `lesson` for the silent repeated scalar.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "window and flag semantics shipped"
  --open="suffix UX feedback"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-window-flag-semantics/goal.md)` under a `feat(cli):` or
   `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the unit matrix and repeated-flag evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
