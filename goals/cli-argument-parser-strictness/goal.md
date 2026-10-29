# Goal: Reject malformed and contradictory arguments instead of silently dropping them

## Goal Description
Priority: P0 (trust). The CLI has two parsers with different strictness. Before the first positional, Go's `flag` package validates values. After the first positional, a hand-rolled scanner (`cmd/thermal/main.go:123-273`) reads strings, calls `strconv.Atoi` or `time.ParseDuration`, and **silently keeps the default when parsing fails**. The result is not an error: it is a plausible report over a different window than the one requested.

Observed on the current source build (2026-10-29):

- `thermal daily --last abc` renders the full all-time report (`Window: [all time (96 active days)]`) and exits `0`; stderr is empty.
- `thermal daily --last 1.5` does the same.
- `thermal daily --weeks abc` and `thermal daily --top xyz` render the full report and exit `0`.
- `thermal live --interval abc` accepts the default interval and exits `0` (non-TTY pointer).
- `thermal --weeks abc` (flag before positional) exits `2` with `invalid value "abc" for flag -weeks: parse error` followed by the entire 84-line manual.
- `thermal opencode --tool codex` renders the OpenCode report, exit `0`; the explicit `--tool` is silently discarded.
- `thermal --tool=` exits `1` with `thermal: unknown tool: ` (empty name).
- `thermal daily --since 2026-10-29 --until 2026-10-01` prints `No activity in the selected window.` and exits `0`.
- `thermal daily --last 0` is accepted as "all time".

The outcome of this goal is a **single strict argument contract**: every value-taking flag is parsed exactly once, every malformed or contradictory invocation exits non-zero with one actionable line naming the offending flag/value, and no invocation can return a report over a window or tool other than the one requested.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 244
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 1
- **Downstream goals**: `cli-unknown-token-diagnostics` (uses the typed parse errors and suggestions) and `cli-top-flag-contract` (uses the validated `--top` value and scope) depend on this goal.
- **Related implemented goals (do not regress)**: `output-gallery-regression-contract` already notes that `thermal daily --since` exits `0` with a duplicate report; this goal is the strict-parser fix behind that observation.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/args.go`
- `cmd/thermal/args_test.go`
- `internal/render/help.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-argument-parser-strictness/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-argument-parser-strictness/plan.md`](plan.md)

## Requirements

1. **One parser, one truth.**
   - Parse flags and positionals in a single pass that permits flags before, between, and after positionals (interspersed parsing), without adding a third-party dependency.
   - A flag token is `--name`, `--name=value`, `-h`, `-v`, `-d`, `-f`; a positional is any token that is not a flag and not the value of a preceding value-taking flag.
   - `--` terminates flag parsing; everything after is positional. Document it.
   - Every value-taking flag is declared once in a metadata table with: name, short alias, value kind (string, int, float, duration, date, enum), required/optional, default, valid range/set, and repeatability.
   - No second list may re-declare the same flag. `validateReportFlags` may keep semantic checks but must consume the parsed values, not re-parse strings.

2. **Strict value handling for every value flag.**
   - Cover at minimum: `--tool`, `--db`, `--weeks`, `--since`, `--until`, `--last`, `--order`, `--sort`, `--top`, `--metric`, `--by`, `--grain`, `--start-of-week`, `--interval`, `--flame-ceiling`, `--format`.
   - Missing value: exit `2`, `thermal: --last needs a value (whole number of days)`.
   - Invalid value: exit `2`, `thermal: --last wants a whole number of days, got "abc"` plus `example: thermal daily --last 7`.
   - Empty value (`--last=`, `--tool=`): exit `2`, `thermal: --last requires a non-empty value`.
   - Out-of-range value: exit `2`, state the range and the received value.
   - Numeric parsing must reject `NaN`, `Inf`, `+`, `-`, whitespace-only, and trailing garbage (`7days`); `strconv` errors must not appear in user output.
   - Duration parsing accepts `1s`, `500ms`, `2m` and a bare number of seconds; zero/negative is rejected.
   - Date parsing accepts `YYYY-MM-DD` and `YYYYMMDD` only; invalid calendar dates (`2026-02-30`) are rejected.
   - Duplicate scalar flags: last occurrence wins, and a `--verbose` line notes the override. Document this; do not reject.

3. **Strict structural handling.**
   - At most two positionals, interpreted as `[tool] [report]` in either order. A third positional exits `2` with `thermal: unexpected argument "extra"`.
   - Empty-string positionals exit `2` with `thermal: empty argument is not a tool or command`.
   - A positional tool that conflicts with `--tool` exits `2`:
     `thermal: positional tool "opencode" conflicts with --tool codex`.
   - A positional tool equal to `--tool` is accepted once.
   - `--tool all` and `--tool auto` remain valid; `--tool=` and `--tool " "` are rejected.
   - Unknown flags exit `2` in every position with `thermal: unknown flag --metirc` and a nearest-flag suggestion; the 84-line manual is not printed for value errors.
   - Unknown single-dash flags (`-x`, `-z`) exit `2` with the same shape.
   - A value that looks like a flag (`--last --json`) exits `2` with `--last needs a value` rather than consuming `--json`.

4. **Semantic validation after parsing.**
   - `--last` must be `>= 1`; `--last 0` is rejected. `--last` with `--since`/`--until` remains rejected.
   - `--weeks` must be `4..104`.
   - `--top` must be `>= 0`; `0` explicitly means "all rows" and is documented; negative is rejected.
   - `--since` must not be after `--until`; include both dates and the swap suggestion.
   - `--order` ∈ `asc|desc`; `--sort` per command (existing rules); `--metric` ∈ `tokens|cost`; `--by` ∈ `tool|model|reasoning`; `--grain` ∈ `day|week|month`; `--start-of-week` ∈ Sunday..Saturday, case-insensitive.
   - `--format` ∈ `text|json`; `md`/`markdown` keeps the retirement error.
   - `--flame-ceiling` is a positive finite number; `--interval` is a positive duration.
   - Report-only flags keep their existing "needs a report command" errors, but with the same error shape and exit `2`.

5. **Error format and exit codes.**
   - Every usage error starts with `thermal: `, is one cause line, optionally one fix line, optionally `run thermal --help`; no stack traces, Go type names, or `flag provided but not defined`.
   - Exit `2` for parse/usage errors; exit `1` for data or semantic errors that require reading data (e.g. unknown tool after resolution); exit `0` only on success.
   - stdout is empty for every usage error; stderr carries the message.
   - `--json` does not change usage errors: they remain stderr text with exit `2` (no JSON error object in this goal).
   - `--verbose` may add detail but must not change the first line or the exit code.

6. **Tests and invariants.**
   - Table-driven matrix over every value flag × {valid, missing, empty, invalid, out-of-range, `=` form, duplicate}.
   - Positional matrix: `[]`, `[tool]`, `[report]`, `[tool report]`, `[report tool]`, `[extra]`, `[tool report extra]`, empty strings.
   - Conflict matrix: positional vs `--tool`, `--last` with `--since`, reversed dates, `--top` with commands that reject it (coordinated with the `--top` goal).
   - A fuzz test (`FuzzParseInvocation`) asserts no panic and that every error is prefixed `thermal: ` and exits `2` or `1`.
   - Golden stderr files for the 12 most common mistakes, max 4 lines each.
   - A property test: for any parsed invocation, the resulting `Options` equals the explicitly requested values; no field retains a default because a value failed to parse.

7. **Documentation and migration.**
   - Help pages show `<value>` placeholders and accepted values (coordinated with `cli-help-interception-and-scoped-help`).
   - `docs/MIGRATION.md` six-part entry: unknown flags and invalid values now fail; `--last 0` now fails; duplicate flags keep the last value; scripts relying on silent fallback must pass valid values or remove the flag.
   - `README.md` and `project.jsonc` updated.

## Non-Goals
- Changing the command taxonomy or adding suggestions for unknown commands (owned by `cli-unknown-token-diagnostics`).
- Changing `--top` applicability (owned by `cli-top-flag-contract`).
- Changing `-v`/`--version` semantics (owned by `cli-version-verbose-flag-contract`).
- Adding a JSON error envelope for usage errors.
- Rewriting report aggregation or rendering.

## Acceptance Criteria
- AC-001: `thermal daily --last abc` exits `2`, stdout is empty, stderr contains `--last`, `abc`, and `whole number`, and does not contain `Thermal · daily`.
- AC-002: `thermal daily --last 1.5`, `--last -1`, `--last 0`, `--weeks abc`, `--top xyz`, and `--interval abc` each exit `2` with a flag-specific message and no report.
- AC-003: `thermal --weeks abc` exits `2`, stderr is at most 4 lines, and contains neither `parse error` nor the full `Thermal · usage profile` manual.
- AC-004: `thermal opencode --tool codex daily` exits `2` with `conflicts with --tool`; `thermal opencode --tool opencode daily` exits `0`.
- AC-005: `thermal --tool= daily` and `thermal --tool "" daily` exit `2` with `requires a non-empty value`.
- AC-006: `thermal daily --since 2026-10-29 --until 2026-10-01` exits `2`, names both dates, and suggests the swap.
- AC-007: `thermal daily extra` exits `2` with `unknown tool "extra"` (two positionals are structurally valid); `thermal daily extra third` exits `2` with `unexpected argument "third"` (three positionals are not).
- AC-008: `thermal daily --metirc cost` exits `2`, names `--metirc` (not `cost`), and suggests `--metric`.
- AC-009: `thermal daily --last --json` exits `2` with `--last needs a value`; `--json` is not consumed as the value.
- AC-010: Every sampled error line matches `^thermal: ` and contains none of `flag provided but not defined`, `parse error`, `strconv.`, `invalid syntax`.
- AC-011: `go test -race ./cmd/thermal/...` exits `0`, including `TestArgumentMatrix`, `TestPositionalMatrix`, `TestConflictMatrix`, `TestErrorVoice`, and `FuzzParseInvocation` (at least 100k executions).
- AC-012: `./scripts/simulated_user_gate.sh` exits `0` and includes the strict-parser section.

## Negative Cases & Failure Modes
| Invocation | Required result |
| --- | --- |
| `daily --last abc` | exit 2, no report |
| `daily --last 1.5` | exit 2, no report |
| `daily --last 0` | exit 2, `--last must be at least 1` |
| `daily --last --json` | exit 2, `--last needs a value` |
| `daily --weeks abc` | exit 2, no report |
| `daily --top xyz` | exit 2, no report |
| `daily --top -1` | exit 2, `--top cannot be negative` |
| `daily --since 2026-13-01` | exit 2, date format message |
| `daily --since 2026-10-29 --until 2026-10-01` | exit 2, swap suggestion |
| `live --interval abc` | exit 2, duration message |
| `--weeks abc` | exit 2, ≤4 lines, no manual |
| `opencode --tool codex` | exit 2, conflict message |
| `--tool=` | exit 2, non-empty message |
| `daily --metirc cost` | exit 2, unknown flag `--metirc` |
| `daily extra` | exit 2, unknown tool "extra" (two positionals are valid) |
| `daily extra third` | exit 2, unexpected argument "third" |
| `daily ""` | exit 2, empty argument |
| `daily --last 7 --last 3` | exit 0, window is 3 days, verbose notes override |

## Metrics & Regression Guards
- Zero silent fallbacks: a property test asserts that no `Options` field changes from `default` to `default` when a supplied value is invalid; invalid input always exits non-zero.
- Error length: every usage error is ≤4 lines in non-verbose mode.
- Manual suppression: a bad value before a positional prints no more than 4 lines, not the 84-line manual.
- Fuzz budget: at least 100,000 executions in CI, no panic, no report on invalid input.
- Gate: the simulated user gate asserts AC-001, AC-003, AC-004, AC-006, AC-008, and AC-009.

## Risks & Mitigations
- **Risk:** Interspersed parsing changes the meaning of a value that begins with `-`. **Mitigation:** a value-taking flag consumes the next token unless it is a known flag; `--last -1` remains a value and is then range-rejected. Add explicit tests.
- **Risk:** Rejecting unknown flags breaks scripts that pass future flags to an older binary. **Mitigation:** that is the point; document in MIGRATION and use `--` only for positionals.
- **Risk:** Duplicate-flag last-wins is invisible. **Mitigation:** `--verbose` warning and a documented rule.
- **Risk:** Error wording changes break golden tests in the simulated user gate. **Mitigation:** update the gate in the same commit and keep stable prefixes.
- **Risk:** The parser refactor touches every command. **Mitigation:** land it behind exhaustive table tests first; keep `validateReportFlags` semantics unchanged except for the documented `--last 0` and reversed-range rules.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "flag package stops at first positional"` and `sila recall "silent numeric parse fallback"`.
- **Writes**: record a `case` for the "single metadata-driven interspersed parser" mechanism, grounded to `cmd/thermal/args.go`; record a `lesson` for the silent-fallback class.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "strict argument parser shipped" --open="JSON usage errors"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-argument-parser-strictness/goal.md)` under a `fix(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the full AC matrix with exit codes.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
