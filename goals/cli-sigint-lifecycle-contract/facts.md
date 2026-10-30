# Facts: Treat SIGINT as a normal stop for live and stream

## Observed Evidence (2026-10-30, real PTY, `/tmp/thermal-cli-review`)

The harness used `pty.fork`, started the process, slept 1.5 s, sent one SIGINT,
and read until exit.

### Interactive `live`
```
$ thermal live --offline --no-update-check --no-color
<PTY; SIGINT after 1.5s>
exit 1
tail (escapes shown as ^[):
  ^[[?2026$p^[[?2027$p^[[>4m^[[?1049h^[[?25l^[[?5W^[[?2004h
  ...
  ^[[?1049l^[[?25h^[[?2004l
  thermal: live monitor failed: program was killed: program was interrupted
```
The alternate screen, cursor, and bracketed paste are restored before the
failure line. The exit code and the sentence are wrong, not the cleanup.

### Stream
```
$ thermal live --json --stream --offline --no-update-check
<PTY; SIGINT after 1.5s>
exit 0
last line: {"snapshot":{...},"timestamp":"...","type":"snapshot"}
```
No error text; the last line parses.

### State after interruption
No `*.lock`, `*.pid`, or `*.tmp.*` was found under the fixture HOME's cache.

## Root Cause
- `cmd/thermal/main.go:1604-1608`:
  ```go
  p := tea.NewProgram(model, tea.WithFPS(10))
  if _, err := p.Run(); err != nil {
      fmt.Fprintf(os.Stderr, "thermal: live monitor failed: %v\n", err)
      os.Exit(1)
  }
  ```
  Bubble Tea returns an error for an interrupted program; the code treats every
  error as a render failure.
- There is no `signal.NotifyContext` around the live program; the TUI's own
  Ctrl+C handling produces the interruption error.
- The stream loop (`main.go:1550-1586`) returns cleanly when stdout closes, but
  it does not install a signal context; the PTY test happens to exit `0`
  because the TUI is not involved.
- `cmd/thermal/sigpipe_unix.go` ignores `SIGPIPE` only; it does not own
  `SIGINT`/`SIGTERM`.
- Cache writers (`internal/loaders/dircache.go`) use atomic temp+rename, but a
  scan is not context-aware; a signal can leave work in progress until the
  process dies.

## Architectural Invariants & Constraints
- SIGINT is a user control, not a failure. The first one is a normal stop.
- The terminal must be restored before any exit, including a forced second
  signal.
- A JSON/NDJSON stream never emits a partial object; the last byte is `\n`.
- Exit `0` is reserved for a requested stop; a real renderer error stays `1`.
- Non-TTY text `live` stays exit `2`; this goal does not change it.
- No stale lock, pid, or temp file after a stop.
- A signal during a scan cancels the scan; it does not commit a partial delta.
- Signals are handled once per process; do not spawn a goroutine per signal.

## File & Interface Contracts
- `cmd/thermal/sigint.go` (new, build-tagged for Unix/Windows):
  ```go
  type liveSignalState struct {
      First  <-chan os.Signal
      Forced bool
  }

  func notifyLiveSignals() (context.Context, context.CancelFunc, func())
  func isNormalStop(err error) bool
  ```
  `isNormalStop` matches `tea.ErrProgramKilled`,
  `tea.ErrProgramInterrupted`, `context.Canceled`, and an error string
  containing `program was interrupted`.
- `cmd/thermal/main.go`:
  ```go
  ctx, cancel, restore := notifyLiveSignals()
  defer restore()
  p := tea.NewProgram(model, tea.WithFPS(10), tea.WithContext(ctx))
  if _, err := p.Run(); err != nil && !isNormalStop(err) {
      fmt.Fprintf(os.Stderr, "thermal: live monitor failed: %v\n", err)
      os.Exit(1)
  }
  os.Exit(0)
  ```
  The second-signal handler calls `restore()` then `os.Exit(130)`.
- `internal/tui/live.go`: accept a `context.Context`; return `tea.Quit` on
  `ctx.Done()`.
- `internal/loaders/dircache.go`: check `ctx.Err()` before writing the delta;
  return the context error without persisting.
- `cmd/thermal/stream_test.go`: assert the line-boundary rule.

## Exit Matrix
| Event | Exit | stdout | stderr |
| --- | --- | --- | --- |
| `q` | 0 | normal | empty |
| first SIGINT | 0 | restored screen | empty |
| second SIGINT | 130 | restored screen | empty |
| SIGTERM | 143 | restored screen | empty |
| TUI error | 1 | restored screen | failure line |
| non-TTY text | 2 | empty | guidance |
| stream SIGINT | 0 | complete NDJSON | empty |
| stream broken pipe | 0 | complete lines | empty |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-home>

# PTY SIGINT: expect exit 0, no failure text, restore sequences
python3 - <<'PY'
import os, pty, signal, time, select
B="/tmp/thermal-cli-review"; H=os.environ["H"]
pid, fd = pty.fork()
if pid == 0:
    os.execvpe(B, [B,"live","--offline","--no-update-check","--no-color"],
               dict(os.environ, HOME=H, TERM="xterm-256color"))
time.sleep(1.5); os.kill(pid, signal.SIGINT)
buf=b""; deadline=time.time()+5
while time.time()<deadline:
    r,_,_=select.select([fd],[],[],0.2)
    if r:
        try: buf+=os.read(fd,65536)
        except OSError: break
    w,st=os.waitpid(pid,os.WNOHANG)
    if w==pid: break
else:
    os.kill(pid, signal.SIGKILL); os.waitpid(pid,0); raise SystemExit(9)
assert os.waitstatus_to_exitcode(st)==0, st
assert b"failed" not in buf and b"killed" not in buf
for seq in (b"\x1b[?1049l", b"\x1b[?25h", b"\x1b[?2004l"):
    assert seq in buf, seq
PY

# Second signal -> 130
# (same harness, send two SIGINTs 100ms apart, assert 130)

# Stream line boundary
python3 - <<'PY'
# start `live --json --stream`, send SIGINT, assert exit 0 and last line parses
PY

# No stale state
find "$H/.cache/thermal" \( -name '*.lock' -o -name '*.pid' -o -name '*.tmp.*' \) | grep -q . && exit 1

# TUI error injection (test-only hook)
# assert exit 1 and one failure line

# Non-TTY control
$B live </dev/null >/tmp/o 2>/tmp/e; test $? -eq 2
test ! -s /tmp/o
grep -q 'live needs a terminal' /tmp/e

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/tui/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- PTY harness with a 5 s deadline and SIGKILL fallback.
- One-SIGINT and two-SIGINT cases.
- Stream case with a 1 s interval.
- SIGTERM case on Unix.
- A scan fixture large enough to interrupt.
- A cache directory with a stale temp file to prove cleanup.
- A test-only TUI error hook.

## Open Questions
- Should the first SIGINT exit `0` or `130`?
  Recommendation: `0`; the user asked to stop. Reserve `130` for the forced
  second signal.
- Should SIGTERM exit `0` or `143`?
  Resolved: `143` on Unix, because it was not a keyboard stop. Document it and
  test it; Windows has no SIGTERM.
- Should a dim `live stopped` line be printed?
  Recommendation: no by default; `--verbose` may print one stderr line.
- Should the stream emit a final summary event on shutdown?
  Recommendation: no; it would change the NDJSON contract. Just stop at a
  line boundary.
- Should a signal during a scan cancel or finish?
  Recommendation: cancel; a partial scan is worse than a short delay.
