# Facts: Make every surface fit the terminal width or refuse with a clear minimum-width message

## Observed Evidence (2026-10-29, current source build `thermal dev`)
### Measured max line widths (`COLUMNS=60`, `--no-color` where supported)
| Command | Max width |
| --- | --- |
| `thermal` | 85 |
| `thermal --help` | 94 |
| `thermal daily` | 104 |
| `thermal weekly` | 104 |
| `thermal monthly` | 104 |
| `thermal projects` | 105 |
| `thermal models` | 106 |
| `thermal mix` | 101 |
| `thermal stats` | 113 |
| `thermal trend` | 119 |
| `thermal audit` | 98 |
| `thermal changelog` | 98 |
| `thermal license` | 77 |
| `thermal opencode` (heatmap) | 100 |

### Real PTY behavior
- 50x20 `thermal live` (tmux capture):
  ```
   ╭─ 🔥 THERMAL LIVE ─────────────────────────── ●
   │ BURN VELOCITY           TODAY'S USAGE
   ...
   │   TIME      AGE   TOOL               TOK      S
   ...
    • Controls: [Space] Pause  ·  [c] Reset  ·  [s]
  ```
  No right border, header cut at `●`, `STEPS` cut to `S`, controls cut mid-word.
- 60x20 / 60x24 `thermal live`: `● LIVE (1.0s` (missing `)`), `LIVE SESS` (missing `ION`), `TOOLTOK` (columns collide).
- 80x24 `thermal live`: clean frame, right border aligned, controls complete.
- 120x40 `thermal live`: clean, expanded KPI columns.
- `COLUMNS=50 script -qec "stty cols 50; thermal"` → still 85-character lines.
- `COLUMNS=40` and `COLUMNS=80` produce identical static output widths.

## Root Cause
- `cmd/thermal/main.go:1073-1084` (`reportWidth`) resolves a width only from `term.GetSize(os.Stdout)` / `term.GetSize(os.Stdin)` with a fallback of `100`; it never reads `COLUMNS`, and it is used only by chart renderers.
- `internal/render/projects.go:70-76`, `internal/render/leaderboard.go:156/254`, and the other report renderers build fixed-width column slices (`projectWidth`, `toolsWidth`, `numberWidth`, `daysWidth`, `lastWidth`) with no width parameter.
- `internal/render/card.go` / `BoundedCardWidth` cap the card, but the table rows inside can exceed the cap; the card border is not recomputed from the row width.
- `internal/tui` uses `tea.WindowSizeMsg` but the layout has only two tested sizes (80x24 and 120x40); below 80 it clips rather than reflowing.
- `internal/render/help.go` renders fixed-width sections (94 columns) with no wrapping.
- The heatmap renderer computes its width from the number of weeks (52 by default), not from the terminal.

## Architectural Invariants & Constraints
- No rendered text line may exceed the resolved width; box borders must be complete.
- Width must be measured in display cells (runes), not bytes; CJK and emoji count as two cells where appropriate.
- `--json` is width-independent and must be byte-identical across widths.
- The live TUI must never hide `q Quit`; the user must always be able to leave.
- `--no-color` output must satisfy the same geometry; color is never used to convey a dropped column.
- The default 80x24 and 120x40 live frames are the compatibility baseline; changes require an intentional golden update.
- `--width` is a test/automation seam and must not alter data.

## File & Interface Contracts
- `internal/render/width.go` (new):
  ```go
  const MinReportWidth = 40
  func ResolveWidth(explicit int, env string, out, in *os.File) int
  func DisplayWidth(s string) int
  func TruncateToWidth(s string, width int) string
  type Layout struct { Width int; ShowModels, ShowTools, ShowDays, ShowCache bool }
  func LayoutFor(report string, width int) Layout
  ```
- `cmd/thermal/main.go`: resolve width once after flag parsing; store it in `thermal.Options.Width`; pass it to every renderer. Add `flag.IntVar(&opts.Width, "width", 0, "Render width in columns (default: terminal width)")`.
- `internal/render/*`: every `Render*` function gains a width parameter or a `RenderOptions{Width: int}`.
- `internal/tui`: `NewLiveModel(..., width, height)` or handle `WindowSizeMsg` with a documented breakpoint function.
- `internal/render/help.go`: `RenderHelp(noColor bool, width int)` and `RenderCommandHelp(..., width int)`.

## Column Priority (frozen)
| Report | Priority (drop last) |
| --- | --- |
| Leaderboard | #, Tool, Tokens, Cost, Strk, Best, Days |
| Daily/weekly/monthly | Period, Tokens, Cost, Models, Input, Output, Reason, Cache |
| Projects | Project, Tokens, Cost, P-Days, Tools, Last |
| Models | Model, Tokens, Cost, M-Days, Tools, Last |
| Stats | label, value, bar |
| Trend | Range, Mean, Slope, Activity, Projection |
| Audit | STATUS, path, lines |
| Changelog | version, date, bullet text |
| Help | flag, description |

## Minimum Widths & Refusal
| Surface | Minimum | Refusal |
| --- | --- | --- |
| Leaderboard | 40 | `thermal: leaderboard needs at least 40 columns (terminal is 39)` |
| Daily/weekly/monthly | 56 | `thermal: daily needs at least 56 columns (terminal is 50)` |
| Projects/models | 48 | `thermal: projects needs at least 48 columns (terminal is 40)` |
| Stats/trend | 40/48 | command-specific |
| Audit/changelog/help | 56/48/48 | command-specific |
| Heatmap | 40 | command-specific |
| Live TUI | 60 | `Thermal Live needs at least 60 columns (terminal is 50); resize or use thermal live --json` |
- Refusal goes to stderr, exit `2`, stdout empty, and includes `--width 100` as the escape hatch.
- At exactly the minimum, render (do not refuse).

## Verification Contract
```bash
# Static width matrix
for w in 40 50 56 60 80 100 120 160; do
  for cmd in "" daily weekly monthly projects models stats mix trend audit changelog "--help"; do
    out=$(COLUMNS=$w thermal $cmd --no-color 2>/tmp/e); code=$?
    if [ $code -eq 0 ]; then
      max=$(printf '%s\n' "$out" | awk '{ if (length>m) m=length } END { print m }')
      test "$max" -le "$w" || { echo "FAIL $cmd at $w: $max"; exit 1; }
    else
      grep -q "needs at least" /tmp/e || { echo "FAIL refusal $cmd at $w"; exit 1; }
    fi
  done
done

# Precedence
diff <(COLUMNS=60 thermal daily --no-color) <(thermal daily --width 60 --no-color)
COLUMNS=60 thermal daily --width 100 --no-color | awk 'length>100{exit 1}'

# JSON independence
diff <(COLUMNS=40 thermal daily --json) <(COLUMNS=160 thermal daily --json)

# Long names
# fixture with a 256-char project name
COLUMNS=60 thermal projects --no-color | awk 'length>60{exit 1}'
COLUMNS=60 thermal projects --no-color | grep -q '…'

# Live PTY matrix (tmux)
for size in "50 20" "60 20" "80 24" "120 40"; do ... done

CGO_ENABLED=0 go test -race ./internal/render/... ./internal/tui/... ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Golden captures at 80 and 120 for every report (stored under `internal/render/testdata/`).
- PTY captures at 50x20, 60x20, 80x24, 100x30, 120x40, 160x40.
- A fixture project with a 256-character name, a CJK name, and an emoji name.
- A heatmap fixture with 52 weeks and a narrow terminal.
- `COLUMNS` values `0`, `abc`, `39`, `241`, and unset.
- A `--width` value with `--json` to prove it is ignored.

## Open Questions
- Should the refusal exit `2` or `1`? Recommendation: `2` (usage: the environment cannot render the request).
- Should the leaderboard below 40 columns render a compact one-line-per-tool list instead of refusing? Recommendation: yes if it can keep Tool and Tokens; otherwise refuse.
- Should `--width` accept `auto`? Recommendation: no; omit the flag or use `COLUMNS`.
- Should the heatmap auto-fit or refuse? Recommendation: auto-fit down to 4 weeks, then refuse.
- Should the TUI stack at 60 or refuse? Recommendation: stacked layout at 60–79, refuse below 60; controls must remain visible.
