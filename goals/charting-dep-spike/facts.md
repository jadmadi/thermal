# Facts: Charting dependency spike: ntcharts v2 against the stock charm stack

## Architectural Invariants & Constraints
- ntcharts v2.2.0 carries `replace charm.land/bubbletea/v2 => github.com/neomantra/bubbletea/v2 ...` in its go.mod. Replace directives do not apply to consumers, so thermal builds against stock `charm.land/bubbletea/v2`.
- Decision: adopt ntcharts v2.2.0 with stock charm modules. Verified by build and render on 2026-09-17, see `spike-notes.md` for the raw runs.
- TUI dependencies are confined to `internal/tui`. Aggregation stays in `internal/thermal`. The CLI path never imports the TUI.
- Charts must never encode meaning in color alone, and must honor NO_COLOR.

## Decided Dependency Set
- `github.com/NimbleMarkets/ntcharts/v2` v2.2.0 (MIT, Neomantra Corp)
- `charm.land/bubbletea/v2` v2.0.9
- `charm.land/lipgloss/v2` v2.0.6
- `github.com/charmbracelet/x/ansi` for `ansi.Strip`, the no-color path
- `github.com/lrstanley/bubblezone/v2` v2.0.0 only if mouse support is wanted later

## Measured Facts
- Cross-builds green: linux/amd64, windows/amd64, darwin/arm64.
- Marginal binary cost, stripped with `-trimpath -ldflags="-s -w"`: +2.1 MB (baseline 1.4 MB, with deps 3.5 MB).
- UPX still works on the linux build: 3.5 MB down to 1.4 MB.
- Chart types verified: braille time series at 64x12 and 30x8, horizontal ranked bars, sparkline strip.

## No-Color Contract
- `lipgloss.Style.Render` always emits truecolor escapes. `lipgloss.Writer` and `colorprofile` downgrade the color parameters but leave bare `ESC[m` resets behind.
- The working no-color path is a full strip: render with ntcharts, then `ansi.Strip(view)`. Measured: 496 escapes become 0 while all 240 block runes survive.
- THERMAL RULE: any chart string that leaves the render layer passes through one helper that either strips ANSI (no-color) or downgrades the profile (256 or 16 color terminals).

## API Friction Found
- `barchart` draws each segment with `BarValue.Style`; bars need an explicit style with a foreground color.
- `barchart.WithBarWidth` plus a canvas too short for bars times width produces an empty chart. Size the canvas from the bar count, or leave bar width to the default.
- `timeserieslinechart` draws axis labels inside `View()`; at 30 columns labels dominate. Narrow layouts need a smaller label formatter.
