# Spike Notes: ntcharts v2 against the stock charm stack

Run on 2026-09-17 in `/tmp/opencode/ntcharts-spike`, a scratch module outside
the repo. Nothing from the spike ships except these findings.

## Setup

```
go mod init spike
go get github.com/NimbleMarkets/ntcharts/v2@v2.2.0
go get charm.land/bubbletea/v2@v2.0.9 charm.land/lipgloss/v2@v2.0.6
```

Resolved without touching the fork pin, because a dependency's `replace` is
ignored by Go. ntcharts v2.2.0 compiled and ran against the stock modules.

## What rendered

| Chart | Size | Result |
|---|---|---|
| Time series, braille | 64x12 | clean, axis labels legible |
| Time series, braille | 30x8 | usable, labels crowd the columns |
| Ranked horizontal bars | 56x14 | renders, label column plus sub-cell blocks |
| Sparkline | 48x3 | renders |

## No-color path, measured

| Pass | Bytes | Escape sequences | Block runes |
|---|---|---|---|
| Raw `View()` | 6253 | 496 | 240 |
| Through `colorprofile` ASCII writer | 2781 | 496 | 240 |
| Then `ansi.Strip` | 1293 | 0 | 240 |

`lipgloss.Style.Render` always emits truecolor. `colorprofile` drops the color
parameters but leaves bare `ESC[m` resets, so the no-color path must strip.

## Build and size

| Build | Result |
|---|---|
| `GOOS=linux GOARCH=amd64` | OK, 3.5 MB |
| `GOOS=windows GOARCH=amd64` | OK, 3.4 MB |
| `GOOS=darwin GOARCH=arm64` | OK, 3.4 MB |
| `upx -9` on the linux binary | OK, 1.4 MB |
| Baseline, `fmt` only, same flags | 1.4 MB |
| Marginal dependency cost | +2.1 MB stripped |

Builds used `-buildvcs=false -trimpath -ldflags="-s -w"` because the scratch
module is not a git checkout.

## Probe sources

Three focused programs were used: a general render pass, a profile comparison,
and a strip check. The essential pattern for thermal is:

```go
view := chart.View()            // truecolor ANSI
if noColor {
    view = ansi.Strip(view)     // 0 escapes, structure preserved
}
fmt.Print(view)
```

## Residual risk

ntcharts v2.2.0 was tested upstream against a Bubble Tea fork whose fixes have
not landed upstream. Compilation and the chart types above behave correctly on
stock modules. A fork-only fix would surface as a rendering defect while
building the TUI views; the fallback remains a hand-rolled braille canvas.
