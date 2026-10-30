# Facts: Enforce one width, colour, and sanitisation policy across every renderer

## Observed Evidence (2026-10-30, `/tmp/thermal-cli-review`, `thermal dev`)

### COLUMNS on a pipe
```
$ COLUMNS=40 thermal daily >/tmp/o 2>/tmp/e
exit 2
stdout: 0 bytes
stderr:
thermal: daily needs at least 56 columns (terminal is 40)
  widen the terminal, pipe to a file, or set --width 100
```
stdout was already a pipe/file. The advice names a remedy the caller already
performed.

### Legend splits marker from gloss
```
$ COLUMNS=60 thermal --no-color
Legend: Strk = current streak · Best = longest streak ·
Days = active days
Cost: $ recorded · ~$ estimated · ≥$ partial known · —
unpriced · ≈ approximate · $0.00 recorded zero
```

### Sources truncates the useful part
```
$ COLUMNS=80 thermal sources --no-color
OpenCode       oc    …/opencode/opencode.db missing      expected file
Codex          —     ~/.codex               missing      expected dire…
Zed            —     …ed/threads/threads.db missing      expected file
```
The head of the path is gone and `DETAIL` truncates its own vocabulary.

### Statusline colour policy
```
$ NO_COLOR=1 thermal statusline | grep -c $'\x1b'
10
$ thermal statusline --plain | grep -c $'\x1b'
10
$ thermal statusline --no-color | grep -c $'\x1b'
0
```
`--plain` removes the emoji but keeps 24-bit colour.

### Control characters tear the frame
```
$ thermal claude --db "<dir containing a newline>"
  ╭─ Claude Activity ─────────────── <path>/new
line/projects ─╮
  │  1.5K tokens / 52 weeks · 1 active days · 1 day streak │
```
The top border and the title split across two physical lines.

## Root Cause
- `internal/render/wrap.go:27-43` (`TerminalWidth`) checks the resolved
  `--width`, then `COLUMNS`, then `term.GetSize(stdout)`, then stdin, then
  `DefaultTerminalWidth`. `COLUMNS` therefore wins even when stdout is not a
  TTY.
- `cmd/thermal/main.go:47-48` applies `--width` early; `main.go:114-123`
  (`requireCommandWidth`) refuses when `TerminalWidth() < min`. The refusal
  text in `render.WidthRefusal` is static and cannot name the source.
- `internal/render/leaderboard.go:410-415` calls `WrapText` on the whole legend
  string, so the wrapper may break between `—` and `unpriced`.
- `internal/render/sources.go:24-40` uses fixed widths (14/5/22/12/14) and
  `truncateLeft`; `truncateLeft` (`:85-94`) keeps the tail and discards the
  head. `truncateSourceRight` (`:96-105`) clips `expected directory` to
  `expected dire…`.
- `internal/render/statusline.go:96` sets `colors := !opts.NoColor`; it never
  reads `NO_COLOR`, `TERM=dumb`, or the output stream, and `Plain` only
  controls the icon.
- There is no `SanitizeText` helper. Strings from `--db`, project names, and
  model names are interpolated directly into card titles and tables.
- `internal/render/card.go` measures `ansi.StringWidth` per line, but a newline
  inside a title creates two logical lines before measurement.

## Architectural Invariants & Constraints
- Width resolves once per invocation; renderers never read `COLUMNS` or
  `os.Getenv` directly.
- `--width` is explicit intent and always wins; `COLUMNS` is ambient and only
  applies to a TTY.
- JSON ignores width, colour, and sanitisation; it carries the original string.
- A legend pair is atomic; a marker is never separated from its gloss.
- `State`, `Tool`, and cost markers are words or glyphs, never colour-only.
- Untrusted strings are data: escape control characters before rendering.
- Colour policy is one function, applied to every surface.
- `--plain` means no glyphs and no ANSI.
- A refusal must name the source of the width (`--width`, `COLUMNS`, or the
  terminal) and a runnable fix.
- No line may exceed the resolved width; the gallery gate enforces it.

## File & Interface Contracts
- `internal/render/wrap.go`:
  ```go
  type WidthSource string

  const (
      WidthFlag    WidthSource = "--width"
      WidthEnv     WidthSource = "COLUMNS"
      WidthTTY     WidthSource = "terminal"
      WidthDefault WidthSource = "default"
  )

  func ResolveWidth(widthOpt int, stdout io.Writer) (int, WidthSource)
  func TerminalWidth() int // uses the resolved value
  ```
  Non-TTY + no `--width` -> 100; TTY -> `term.GetSize` then `COLUMNS`; a
  non-TTY `COLUMNS` is ignored.
- `internal/render/leaderboard.go`:
  ```go
  type LegendItem struct {
      Marker string
      Gloss  string
  }

  func WrapLegend(items []LegendItem, width int) []string
  ```
  Joins with ` · ` and wraps only between items; continuation indent is the
  legend's first column.
- `internal/render/sources.go`:
  ```go
  func FitPath(display string, width int) string
  func FitSourcesRow(item thermal.SourceInfo, width int) string
  ```
  `FitPath` elides the middle: keep `~/` or the first component and the last
  two components; insert `…`.
- `internal/render/sanitize.go` (new):
  ```go
  func SanitizeText(s string) string
  ```
  Escapes `\n`, `\r`, `\t`, ESC, and C0/C1 controls; preserves valid UTF-8;
  replaces invalid bytes.
- `internal/render/color.go` (or `wrap.go`):
  ```go
  type ColorOptions struct {
      NoColor bool
      Plain   bool
  }

  func ColorEnabled(opts ColorOptions, w io.Writer) bool
  ```
  false for `NoColor`, `Plain`, `NO_COLOR != ""`, `TERM=dumb`, or `w` not being
  a TTY.
- `internal/render/statusline.go`: replace `colors := !opts.NoColor` with
  `ColorEnabled`; `Plain` disables colour and glyphs.
- `internal/render/card.go`: `SanitizeText` the title and right label before
  width measurement.
- `cmd/thermal/main.go`: pass the width source into `WidthRefusal`.

## Width Policy
| Condition | Resolved width | Refusal? |
| --- | --- | --- |
| `--width 100` | 100 | no |
| `--width 40`, min 56 | 40 | yes, names `--width 40` |
| TTY 120, no `--width` | 120 | no |
| TTY 40, no `--width` | 40 | yes, names the terminal |
| pipe, `COLUMNS=40`, no `--width` | 100 | no |
| pipe, no `COLUMNS` | 100 | no |
| JSON, any | ignored | no |

## Verification Contract
```bash
B=/tmp/thermal-cli-review
H=<fixture-home>

# Non-TTY COLUMNS is ignored
COLUMNS=40 $B daily >/tmp/o 2>/tmp/e; test $? -eq 0
test -s /tmp/o
awk 'length>100' /tmp/o | grep -q . && exit 1

# Explicit --width still refuses below the min
$B daily --width 40 >/tmp/o 2>/tmp/e; test $? -eq 2
grep -q -- '--width 40' /tmp/e

# Legend atomic
COLUMNS=60 $B --no-color >/tmp/o 2>/dev/null
! grep -E '· —$' /tmp/o
! grep -E '^unpriced' /tmp/o

# Sources fit and keep the head
COLUMNS=80 $B sources --no-color >/tmp/o
! grep -q 'expected dire' /tmp/o
grep -q '~/.claude' /tmp/o
awk 'length>80' /tmp/o | grep -q . && exit 1

COLUMNS=40 $B sources --no-color >/tmp/o
awk 'length>40' /tmp/o | grep -q . && exit 1

# Sanitisation
mkdir -p "/tmp/x
y/projects/p"
# run thermal claude --db "/tmp/x\ny" with a fixture record
# assert the border closes on one line and no raw newline appears in a title

# Colour matrix
NO_COLOR=1 $B statusline | grep -c $'\x1b' | grep -qx 0
$B statusline --plain | grep -c $'\x1b' | grep -qx 0
$B statusline --no-color | grep -c $'\x1b' | grep -qx 0
TERM=dumb $B --no-color >/dev/null

# JSON is width-independent
$B --json --width 40 >/tmp/a
$B --json --width 200 >/tmp/b
# compare after removing generatedAt

CGO_ENABLED=0 go test -race ./internal/render/... ./cmd/thermal/...
./scripts/output_gallery_gate.sh
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- PTY harness for TTY width and colour.
- Pipe harness with `COLUMNS=40`, 60, 80, 120.
- Fixture HOME with long paths, `~` paths, and a path containing `\n`.
- A source row with `expected directory`.
- `NO_COLOR`, `TERM=dumb`, `CLICOLOR_FORCE=1`, `--plain`, `--no-color`.
- A 40-column golden for `sources`, `thermal`, and `daily`.
- An invalid UTF-8 byte fixture.

## Open Questions
- Should `COLUMNS` be honored on a pipe when it is explicitly exported?
  Recommendation: no; `--width` is the explicit control. Document it.
- Should JSON keep raw control characters?
  Recommendation: yes; JSON escaping handles them, and consumers expect the
  original path.
- Should `--plain` remain a statusline-only flag?
  Recommendation: yes for now, but its contract is "no glyphs and no ANSI";
  applying it to reports would be a no-op.
- Should `CLICOLOR_FORCE=1` force colour on a pipe?
  Recommendation: no; it only forces on a TTY. A pipe stays clean.
- Should the sources table drop `Alias` before `Detail`?
  Recommendation: drop `Detail` first (state + path are more useful), then
  `Alias`.
