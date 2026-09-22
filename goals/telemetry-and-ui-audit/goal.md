# Goal: Telemetry Accuracy, Pricing Estimates, and UI Polish (5 Issues)

## Goal Description
Fix 5 issues identified during operational audit of Thermal:
1. **Pricing Estimates on Leaderboard & Single-Tool Dashboard**: Display estimated cost (with `~` prefix) on the leaderboard and in single-tool heatmaps (`thermal <tool>`) when native recorded cost is absent and model usage exists.
2. **dsh in `--help`**: Add `dsh` / `DeepSeek (DSH)` to the supported tools list in `thermal --help` usage output and flag descriptions.
3. **Canonical Tool Name**: Rename tool display name from `dsh` to `DeepSeek (DSH)` across CLI, dashboards, reports, and tables.
4. **Telemetry Invariant in Analytics (`mix`, `stats`, `trend`)**: Filter out `isActivityOnly` tools (steps/messages) when computing token metrics so activity counts never leak into token totals (fixes the 417 "other" tokens bug in `mix` and distorted low-token histogram buckets).
5. **UI Redesign of `thermal stats`**: Replace ugly ASCII `#` histogram with sleek Unicode block bars (`█` / `·`), add mini-bars to weekday profiles, eliminate duplicate outliers/top days, and clean up column alignments.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/telemetry-and-ui-audit/facts.md`](facts.md)
- **Execution Plan**: [`goals/telemetry-and-ui-audit/plan.md`](plan.md)

## Done Condition
1. All 5 issues analyzed, addressed, and verified with unit tests and live execution.
2. `go test -v -race ./...` passes with zero race conditions.
3. `golangci-lint` passes without issues.
4. `thermal --help`, `thermal`, `thermal dsh`, `thermal mix`, and `thermal stats` verified cleanly on terminal output.
5. All docs (`README.md`, `AGENTS.md`, `index.html`, `404.html`) updated with full parity.
