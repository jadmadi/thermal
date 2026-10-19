# Goal: Define and expose the live flame ceiling per metric

## Goal Description
The flame gauge currently mixes metrics and event counts. `intensity := math.Max(FlameIntensity, ActivityIntensity)` (`internal/tui/live.go:307-320`) takes the larger of two event-bonus accumulators, so a step-heavy tick can show `72% BURNING` at 23 step/min while a token-heavy tick shows `37% WARM` at 3.4M tok/min. `FlameIntensity`/`ActivityIntensity` are incremented by +0.5/+0.4 per poll (`internal/thermal/live.go:840-847`) and decay by 0.85 per second (`:175-176`), not by any published volume ceiling. The gauge never states what 37% is a percentage of.

Replace the event-bonus model with a single metric-specific scale: `intensity = clamp(60-second rate / ceiling, 0, 1)`, where the ceiling is `9,000,000 tok/min` for token mode and `60 step/min` for activity mode. Show the ceiling beside the percentage (`38% of 9.0M tok/m`) and make it configurable with `--flame-ceiling=N` in the active metric's units.

Priority: **P1**. Derived from the 2026-10-19 live-monitor feedback (bug 7; improvement "Flame: show the ceiling and make it configurable").

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: live-baseline-separation, live-second-bucket-accounting
- **Sequence**: 103
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `cmd/thermal/main.go`
- `cmd/thermal/simulated_user_test.go`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)
- Prerequisites: live-baseline-separation, live-second-bucket-accounting

## Requirements
1. **One metric per gauge**: Select the flame basis explicitly. Token mode uses `BurnTokensPerMin`; activity-only mode uses `BurnTurnsPerMin`; hybrid mode uses tokens (the spend-relevant metric) and says so. Never take `max()` across token and step intensities, and never let a step delta change a token-mode flame.
2. **Scale formula**: `FlameIntensity = clamp(rate60 / ceiling, 0, 1)`. No per-event bonuses and no exponential decay; the 60-second rolling window naturally decays the rate. A zero rate yields exactly zero. This removes the arbitrary seeded 0.3 flame.
3. **Defaults and states**: Default token ceiling `9_000_000` tok/min and default activity ceiling `60` step/min. Keep the existing state thresholds: BURNING >= 60%, WARM >= 20%, COOLING > 2%, IDLE otherwise, now explicitly defined as percentages of the ceiling. At 3.4M tok/min the gauge reads 38% WARM; at 23 step/min it reads 38% WARM, not 72% BURNING.
4. **Ceiling display**: The flame line reads `Flame: [bar] 38% of 9.0M tok/m  WARM` (using exact rounded percent and compact ceiling), or `38% of 60 step/m  WARM` in activity mode. It must name the metric and the ceiling; `38%` alone is not acceptable. The help screen documents the formula and thresholds.
5. **Configurable flag**: Add `--flame-ceiling=N` to `thermal live`, where `N` is a positive number in the active metric's units (tokens/min or steps/min) and accepts compact `K`/`M`/`B` suffixes (e.g. `9M`, `9000000`, `60`). Because the CLI parses positional verbs before custom flags, register it as a string flag and parse/normalize after positional processing so `--flame-ceiling=9M` and `--flame-ceiling 9M` both survive `flag.Parse()`. Reject non-positive, non-finite, or malformed values with a clear error; reject the flag on commands other than `live`.
6. **Snapshot and JSON**: Replace `ActivityIntensity` with a single `FlameIntensity` plus `FlameMetric` (`tokens` or `steps`) and `FlameCeiling` in `LiveSnapshot` (`flameMetric`, `flameCeiling`). The configured ceiling is reflected in the snapshot and `thermal live --json`. Document the removed field and new fields in `docs/MIGRATION.md`.
7. **Baseline and buckets**: Baseline/carried-in volume never contributes to the flame (prerequisite). Rates come from the one-second buckets (prerequisite), not from a separate accumulator. `ResetSession` clears the session counters but the flame follows the rolling window; it must not jump or retain a seeded value.
8. **Safety**: No new dependencies; keep the gauge width math and ANSI padding intact; no `--no-color` regressions.

## Done Condition
1. A 3.4M tok/min window with the default 9M ceiling renders `38%` and `WARM` (not BURNING), and the line contains `9.0M tok/m`.
2. A 23 step/min activity window with the default 60 step/min ceiling renders `38%` and `WARM`, and the line contains `60 step/m`; it is not influenced by any token intensity.
3. `--flame-ceiling=4.5M` on a 3.4M tok/min window renders `76%` and `BURNING`; `--flame-ceiling=0`, `--flame-ceiling=abc`, and use on a non-live command each return a clear error.
4. A zero-activity snapshot has `flameIntensity == 0`, `flameMetric` set to the active mode, and `IDLE`; no seeded 0.3 flame remains.
5. The TUI at 80x24 and 120x40 shows the bar, percentage, ceiling, metric, and state without clipping; `--json` exposes `flameIntensity`, `flameMetric`, and `flameCeiling`.
6. Focused tests, full race suite, CLI build, and simulated user gate pass; migration documentation and `project.jsonc` are synchronized; the attribution commit ends in `(goals/live-flame-scale-contract/goal.md)`.

## Scope Boundary
No session/baseline changes (prerequisite), no bucket/Peak algorithm changes (prerequisite), no ticker column/cost/outlier changes (`live-ticker-unit-and-cost-clarity`), and no rollup/sort (`live-ticker-rollup-and-cost-sort`). This goal owns only the flame scale, its display, and its configuration.
