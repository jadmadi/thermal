# Goal: Treat SIGINT as a normal stop for live and stream

## Goal Description
Priority: P1 (interruption trust). Ctrl+C is the documented way out of
`thermal live`, but the process reports it as a crash: the terminal is restored
correctly, then the CLI prints a doubled failure sentence and exits `1`. A
user who runs `thermal live; echo done` is told the exit was abnormal when it
was exactly the expected action. `live --json --stream` already exits `0`,
which is the behavior `live` should match.

Observed on `2026-10-30` with `/tmp/thermal-cli-review` (`thermal dev`) under a
real PTY, sending one SIGINT after 1.5 s:

- `thermal live --offline --no-update-check --no-color` -> exit `1`; tail:
  ```
  ^[[?1049l^[[?25h^[[?2004l
  thermal: live monitor failed: program was killed: program was interrupted
  ```
  The alternate screen, cursor, and bracketed paste are restored correctly.
- `thermal live --json --stream` -> exit `0`; the last line is complete NDJSON,
  no error text.
- No `*.lock`, `*.tmp`, or `*.pid` files were left under HOME.

The outcome of this goal is a **SIGINT lifecycle contract**:

1. The first Ctrl+C in `live` is a normal stop: restore the terminal, print
   nothing (or one dim line), exit `0`.
2. A second Ctrl+C within 2 seconds is a forced stop: exit `130`.
3. `--json --stream` stops at a line boundary, exits `0`, and never emits a
   partial JSON line.
4. SIGTERM follows the same clean-stop path; SIGKILL is out of scope.
5. No cache, lock, or temporary file is left behind; an interrupted scan does
   not persist a partial delta.

## Dependencies & Execution Order
- **Mode**: Dependent (prerequisite goal is implemented)
- **Depends On**: `cli-nontty-and-statusline-output-contract`
- **Sequence**: 259
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 3

Related goals:
- `cli-nontty-and-statusline-output-contract` owns the non-TTY channel and
  exit-code matrix; this goal owns the signal path in the TTY and stream modes.
- `live-ticker-rollup-and-cost-sort`, `live-tracker-delta-clamping`, and the
  other live goals own the tracked numbers; this goal owns only shutdown.
- `directory-scan-incremental-cache` owns cache writes; this goal adds the
  interruption invariant that a partial scan is never committed.

## Files to Touch
- `cmd/thermal/main.go`
- `cmd/thermal/sigint.go`
- `cmd/thermal/sigint_test.go`
- `cmd/thermal/stream_test.go`
- `cmd/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/loaders/dircache.go`
- `internal/loaders/codexcache.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**:
  [`goals/cli-sigint-lifecycle-contract/facts.md`](facts.md)
- **Execution Plan**:
  [`goals/cli-sigint-lifecycle-contract/plan.md`](plan.md)
- Prior contract:
  [`goals/cli-nontty-and-statusline-output-contract/goal.md`](../cli-nontty-and-statusline-output-contract/goal.md)

## Requirements

1. **Signal handling.**
   - Install `signal.NotifyContext` for `os.Interrupt` (and `SIGTERM` on Unix)
     around the live program and the stream loop.
   - On the first signal:
     - Send `tea.Quit` to the Bubble Tea program (or cancel the context).
     - Wait up to 2 s for `p.Run()` to return.
     - Restore the terminal if the program did not (the current restore
       sequences are correct; keep them).
     - Exit `0`.
   - On a second signal before shutdown completes: exit `130` immediately
     after the terminal restore.
   - SIGTERM follows the first-signal path and exits `143` (128 + 15); it is
     not a keyboard stop.
   - Do not print `live monitor failed` for
     `tea.ErrProgramKilled`/`program was interrupted`; treat those as the
     normal stop.
   - A real TUI error still prints `thermal: live monitor failed: <err>` and
     exits `1`.

2. **Exit-code matrix.**
   | Event | Exit | Output |
   | --- | --- | --- |
   | `q` in the TUI | 0 | normal screen restore |
   | first SIGINT | 0 | restore, no error |
   | second SIGINT within 2 s | 130 | restore, no error |
   | SIGTERM | 143 | restore, no error |
   | TUI init/render error | 1 | `live monitor failed: <err>` |
   | non-TTY text | 2 | guidance on stderr |
   | stream + SIGINT | 0 | complete NDJSON line, no error |
   | stream + broken pipe | 0 | no error |
   - The matrix is frozen in a table test and mirrored in `live --help`.

3. **Stream shutdown.**
   - The stream loop checks `ctx.Done()` before writing each event and after
     each poll.
   - A signal never cuts a JSON object in half. Either finish the current line
     or drop the partial event; the last stdout byte is `\n`.
   - Flush stdout before exit; return `0` even if the signal arrives during a
     poll.
   - `--json --stream | head -n 1` keeps its clean SIGPIPE behavior; do not
     regress it.
   - No diagnostic text on stdout; a shutdown notice is allowed on stderr only
     under `--verbose`.

4. **Terminal restoration.**
   - Assert the restore sequence contains:
     `\x1b[?1049l` (alt screen off), `\x1b[?25h` (cursor shown),
     `\x1b[?2004l` (bracketed paste off).
   - Do not clear the user's scrollback or print a newline that shifts the
     prompt.
   - If the terminal was not in alt-screen mode (e.g. a minimal TERM), the
     restore is a no-op, not an error.
   - `TERM=dumb` on a TTY refuses `live` with the existing width/terminal
     guidance; do not attempt the TUI.

5. **No stale state.**
   - A signal during a directory scan must cancel the scan and not write the
     delta cache.
   - A signal during `live` must not write a partial statusline cache; the
     cache write is atomic (temp + rename) today, keep it.
   - Remove any temp file created by an interrupted cache write on the next
     run; do not leave `*.tmp.*`.
   - After SIGINT, `find "$HOME/.cache/thermal" -name '*.lock' -o -name
     '*.pid'` is empty.
   - A second run after SIGINT produces the same output as a clean run.

6. **Tests.**
   - PTY harness (`pty.fork` or `script`) that starts `live`, sends SIGINT
     after 1.5 s, and asserts exit `0`, no `failed` text, and the three restore
     sequences.
   - Second-SIGINT test: send two signals 100 ms apart; assert exit `130`.
   - Stream test: start `--json --stream`, send SIGINT, assert the last line
     parses with `jq -e` and exit `0`.
   - SIGTERM test on Unix; skip on Windows.
   - Cache test: interrupt a scan with a context and assert no delta file.
   - TUI error test: inject a renderer error and assert exit `1` with the
     failure line.
   - Gate: `simulated_user_gate.sh` includes a PTY SIGINT case (skip if no
     `script`/`python3`).

7. **Documentation.**
   - `thermal live --help` states: `q or Ctrl+C stops the monitor; exit 0`.
   - README live section adds the same sentence.
   - MIGRATION entry for the exit-code change (`1` -> `0`) with detection:
     `timeout -s INT 1 thermal live; echo $?`.
   - `project.jsonc` updated.

## Non-Goals
- Changing the live metrics, polling, or layout.
- Handling SIGKILL or a power loss.
- Adding a crash reporter or telemetry.
- Changing the non-TTY `live` exit `2`.
- Changing SIGPIPE handling for `--stream`.

## Acceptance Criteria
- AC-001: PTY `thermal live`, one SIGINT -> exit `0`, no `failed`/`killed`/
  `interrupted` text, restore sequences present.
- AC-002: A second SIGINT while shutdown is still in progress -> exit `130`
  with the terminal restored. If the first stop already completed, a later
  signal does not override the clean exit `0`.
- AC-003: `thermal live --json --stream`, SIGINT -> exit `0`, last line is
  valid JSON, no partial object.
- AC-004: SIGTERM on Unix -> exit `143`, no failure text.
- AC-005: A real TUI render error still exits `1` with one failure line.
- AC-006: After SIGINT, no `*.lock`/`*.pid`/`*.tmp.*` remains under the cache.
- AC-007: A second run after SIGINT is byte-identical to a clean run.
- AC-008: `thermal live </dev/null` still exits `2` with stderr guidance.
- AC-009: `go test -race ./cmd/thermal/... ./internal/tui/...` exits `0`,
  including `TestLiveSIGINT`, `TestLiveSecondSIGINT`,
  `TestStreamSIGINTLineBoundary`, and `TestInterruptedScanNoCache`.
- AC-010: `./scripts/simulated_user_gate.sh` exits `0` and includes the PTY
  SIGINT case.
- AC-011: `live --help` documents `q or Ctrl+C; exit 0`.
- AC-012: MIGRATION contains the exit-code change and a detection command.

## Negative Cases & Failure Modes
| Event | Required result |
| --- | --- |
| One SIGINT | exit 0, restore, no error |
| Two SIGINTs | exit 130, restore |
| SIGTERM | 143, no error |
| SIGINT during poll | exit 0, complete line |
| SIGINT during stream write | finish or drop whole line |
| SIGINT during scan | cancel, no delta cache |
| TUI error | exit 1, failure line |
| Non-TTY text | exit 2, guidance |
| Broken pipe | exit 0, no error |
| TERM=dumb TTY | refusal, exit 2 |

## Metrics & Regression Guards
- Exit matrix: 100% of rows asserted under a PTY or a pipe.
- Terminal restore: three sequences present in every interactive stop.
- Stream integrity: every stdout line parses; last byte is `\n`.
- State cleanup: zero lock/pid/tmp files after interruption.
- Repeatability: post-SIGINT run equals a clean run.
- Gate: simulated user gate asserts AC-001, AC-003, and AC-006.

## Risks & Mitigations
- **Risk:** Exit `0` hides a real crash. **Mitigation:** only
  `ErrProgramKilled`/interrupted becomes `0`; other errors stay `1`.
- **Risk:** A second SIGINT leaves the terminal dirty.
  **Mitigation:** restore synchronously before `os.Exit(130)`.
- **Risk:** A signal during a cache write corrupts the delta.
  **Mitigation:** temp+rename and a context check before commit.
- **Risk:** Windows lacks SIGTERM.
  **Mitigation:** build-tagged handlers; skip POSIX-only tests on Windows.
- **Risk:** A PTY test hangs in CI.
  **Mitigation:** bound every wait with a 5 s deadline and kill the child.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "live monitor failed interrupted"` and
  `sila recall "SIGINT exit code"`.
- **Writes**: record a `case` for the "first signal normal, second forced"
  lifecycle; record a `lesson` for reporting Ctrl+C as a failure.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "SIGINT lifecycle shipped"
  --open="Windows signal parity"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under
  `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with
   `(goals/cli-sigint-lifecycle-contract/goal.md)` under a `fix(cli):` or
   `feat(cli):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS`
   plus the PTY, stream, and cleanup evidence.
4. `sila goals --json` reports `status: "implemented"`,
   `verification_status: "verified"`, and `unverified_count: 0`, with a clean
   `git status`.
