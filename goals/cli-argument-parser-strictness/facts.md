# Facts: Reject malformed and contradictory arguments instead of silently dropping them

## Observed Evidence (2026-10-29, current source build `thermal dev`)
- `thermal daily --last abc` → full all-time report (`Window: [all time (96 active days) · all tools]`), exit `0`, stderr empty.
- `thermal daily --last 1.5` → same all-time report, exit `0`.
- `thermal daily --weeks abc` → full daily report, exit `0`.
- `thermal daily --top xyz` → full daily report, exit `0`.
- `thermal live --interval abc` → non-TTY pointer, exit `0`; the interval value is ignored.
- `thermal --weeks abc` → `invalid value "abc" for flag -weeks: parse error`, then the full 84-line help, exit `2`.
- `thermal --top` (no value) → `flag needs an argument: -top`, then the full help, exit `2`.
- `thermal opencode --tool codex` → `Thermal · OpenCode Activity` heatmap, exit `0`; the positional wins.
- `thermal --tool=` → `thermal: unknown tool: `, exit `1`.
- `thermal daily --since 2026-10-29 --until 2026-10-01` → `No activity in the selected window.`, exit `0`.
- `thermal daily --last 0` → all-time report, exit `0`.
- `thermal daily --order sideways` → `thermal: --order must be asc or desc`, exit `1` (positive: semantic validation works when the value is actually parsed).
- `thermal daily --start-of-week funday` → `thermal: --start-of-week must be a weekday name, sunday through saturday`, exit `1` (positive).

## Root Cause
- `cmd/thermal/main.go:112`: `flag.Parse()` uses the stdlib `flag` package, which stops at the first non-flag argument.
- `cmd/thermal/main.go:123-273`: the post-positional scanner handles a fixed switch. Numeric cases use:
  ```go
  if n, err := strconv.Atoi(v); err == nil {
      opts.Last = n
  }
  ```
  and silently ignore the error. The same pattern exists for `--weeks`, `--top`, `--interval`, and the float fallback for `--interval`.
- The scanner has no `default:` case, so unknown flags are ignored in post-positional positions.
- `cmd/thermal/main.go:299-311`: positional assignment is unconditional; `opts.Tool` is overwritten by the positional even when `--tool` was supplied.
- `cmd/thermal/main.go:406-429`: `validateReportFlags` validates parsed values but cannot recover a value that was dropped by the scanner. It checks `opts.Last < 0` but not `--last 0`.
- `cmd/thermal/main.go:322-325`: `--weeks` range is checked only if parsing succeeded.

## Architectural Invariants & Constraints
- A CLI invocation must either do exactly what was asked or fail; it must never do something else with exit `0`.
- Every value-taking flag has one parser and one error message.
- Error messages are user-facing: no Go stdlib text, no flag-package jargon, no full manual for a value error.
- Exit `2` is reserved for usage/parse errors; exit `1` for data/semantic errors after a successful parse.
- stdout stays empty on usage errors; `--json` does not change that contract.
- The parser must remain dependency-free (stdlib only) and CGO-free.
- Existing valid invocations (`daily --last 7`, `opencode weekly --chart`, `--since 20261001`) must keep working byte-for-byte.

## File & Interface Contracts
- `cmd/thermal/args.go` (new):
  - `type valueKind int` (`stringKind`, `intKind`, `floatKind`, `durationKind`, `dateKind`, `enumKind`)
  - `type flagSpec struct { Name string; Alias string; Kind valueKind; Default string; Min, Max int; Values []string; Placeholder string; Help string }`
  - `type invocation struct { Positionals []string; Values map[string]string; Present map[string]bool }`
  - `func parseInvocation(args []string, specs []flagSpec) (invocation, error)`
  - `type usageError struct { Flag, Value, Message, Example string }` implementing `error` with `thermal: …` formatting.
- `cmd/thermal/main.go`: replace the `flag.*Var` block and the manual scanner with `parseInvocation`; map the result into `thermal.Options`; call semantic validation.
- `cmd/thermal/args_test.go`: table matrix, positional matrix, conflict matrix, fuzz.
- `internal/render/help.go`: placeholders come from the same `flagSpec` table.

## Error Messages (canonical)
```
thermal: --last wants a whole number of days, got "abc"
  example: thermal daily --last 7

thermal: --weeks wants a number between 4 and 104, got "abc"
  example: thermal daily --weeks 12

thermal: --since 2026-10-29 is after --until 2026-10-01
  did you mean --since 2026-10-01 --until 2026-10-29?

thermal: positional tool "opencode" conflicts with --tool codex
  remove one of them, or make both "opencode"

thermal: unknown flag --metirc
  did you mean --metric?
  run thermal --help for the flag list

thermal: --tool requires a non-empty value
  example: thermal --tool opencode daily
```
- Maximum four lines per error in non-verbose mode.
- All messages go to stderr; stdout is empty.

## Verification Contract
```bash
# Silent-fallback matrix: every one must exit 2 and print no report
for args in "daily --last abc" "daily --last 1.5" "daily --weeks abc" \
            "daily --top xyz" "live --interval abc" "daily --last 0" \
            "daily --top -1" "daily --since 2026-10-29 --until 2026-10-01"; do
  out=$(thermal $args 2>/tmp/e); code=$?
  test $code -eq 2 || { echo "FAIL exit $code: $args"; exit 1; }
  test -z "$out" || { echo "FAIL stdout: $args"; exit 1; }
done

# Unknown flag after positional is attributed to the flag, not the value
thermal daily --metirc cost >/dev/null 2>/tmp/e; test $? -eq 2
grep -q -- '--metirc' /tmp/e
! grep -q 'unknown tool: cost' /tmp/e

# Conflict
thermal opencode --tool codex daily >/dev/null 2>/tmp/e; test $? -eq 2
grep -q 'conflicts with --tool' /tmp/e

# No stdlib text
for f in /tmp/e; do
  ! grep -Eq 'flag provided but not defined|parse error|strconv\.' "$f"
done

# Tests
CGO_ENABLED=0 go test -race ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A table of all value flags with valid, missing, empty, invalid, out-of-range, and `=`-form examples.
- Positional permutations: none, tool, report, tool+report, report+tool, extra, empty string.
- A fuzz corpus with `--`, `-`, `=`, unicode, very long tokens, and repeated flags.
- Golden stderr files for the canonical messages above.
- A synthetic invocation table asserting the final `Options` equals the requested values.

## Open Questions
- Should `--last 0` mean "all time" or be rejected? Recommendation: reject; `--all` or omitting the flag is the explicit way to request all time.
- Should duplicate flags be last-wins or an error? Recommendation: last-wins with a `--verbose` note, because scripts often compose flags.
- Should unknown flags be rejected even when a later `--help` is present? The help goal defines help > validation; this goal must not override that precedence.
- Should `daily extra` (one unknown positional) exit `2` (usage) or `1` (unknown tool after resolution)? Recommendation: `cli-unknown-token-diagnostics` owns classification; this goal exposes a typed error that can carry either exit code.
