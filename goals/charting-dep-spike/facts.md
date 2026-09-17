# Facts: Charting dependency spike: ntcharts v2 against the stock charm stack

## Architectural Invariants & Constraints
- ntcharts v2.2.0 carries `replace charm.land/bubbletea/v2 => github.com/neomantra/bubbletea/v2 ...` in its go.mod. Replace directives do not apply to consumers, so thermal builds against stock `charm.land/bubbletea/v2`.
- The spike module lives outside the repo. Nothing lands in thermal until the adopt-or-fallback decision is written down here.
- TUI dependencies are confined to `internal/tui`. Aggregation stays in `internal/thermal`. The CLI path never imports the TUI.
- Charts must never encode meaning in color alone, and must honor NO_COLOR.

## File & Interface Contracts
- Candidate dependencies: `github.com/NimbleMarkets/ntcharts/v2` v2.2.0, `charm.land/bubbletea/v2` v2.0.9, `charm.land/lipgloss/v2` v2.0.6, optionally `github.com/lrstanley/bubblezone/v2`.
- Fallback: hand-rolled primitives under `internal/tui/chart` using eighth-height blocks (▁▂▃▄▅▆▇█), eighth-width blocks (▏▎▍▌▋▊▉█), shade blocks (░▒▓█), and braille (U+2800 to U+28FF).
- Measured facts to record: build result on linux/amd64 and windows/amd64, binary size delta against v0.6.0, UPX compatibility, and any API friction with forcing no-color output.
