# Plan: Charting dependency spike: ntcharts v2 against the stock charm stack

## Execution Steps
- [x] **Task 1 — Scratch module**: Create a module outside the repo (for example `/tmp/opencode/ntcharts-spike`) importing `github.com/NimbleMarkets/ntcharts/v2` v2.2.0, `charm.land/bubbletea/v2` v2.0.9 and `charm.land/lipgloss/v2` v2.0.6. Render one line chart, one bar chart and one sparkline to a string without running a program loop.
- [x] **Task 2 — No-color and fixed width**: Force the color profile to no color, set explicit widths and heights, and capture the rendered strings so `--no-color` behavior is proven before any code lands.
- [x] **Task 3 — Cross-build and size**: Build linux/amd64 and windows/amd64 from the spike module, record the binary size delta against the v0.6.0 release binary, and check UPX compatibility for linux.
- [x] **Task 4 — Decision**: Adopt ntcharts, or fall back to a hand-rolled braille and block canvas on bubbletea plus lipgloss. Record the measured numbers, the API friction found, and the decision in `facts.md`.
- [x] **Task 5 — Attribution commit**: Commit the spike notes with `Goal-Ref: charting-dep-spike` using a `feat:`, `fix:`, `refactor:`, `perf:` or `chore:` subject (docs and test subjects are ignored), then run `sila goals`.
