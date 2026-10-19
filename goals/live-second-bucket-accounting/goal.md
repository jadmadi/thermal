# Goal: Bucket live deltas into explicit one-second bins so Peak, Rate, and Speed share one window

## Goal Description
`Peak: 567.9K tok/s` is currently the largest per-poll delta, not a per-second rate: the tracker advances empty buckets for elapsed seconds and then dumps the entire delta into the current bucket (`internal/thermal/live.go:163-190`, `:840-847`). A single +567.9K event therefore becomes a 567.9K tok/s peak. Meanwhile Rate shows `8 step/m` while Speed shows `0.1 step/s` because both are derived from the same 60-second sum but formatted with different precision (`internal/tui/live.go:568-630`).

Make the unit of time explicit: every poll delta is allocated across the wall-clock seconds it spans, each bucket represents exactly one second, and Peak is the largest one-second bucket. Rate and Speed must be two displays of the same canonical 60-second window with consistent precision.

Priority: **P1**. Derived from the 2026-10-19 live-monitor feedback (bugs 3-4; improvement "Flame" depends on this window).

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: live-baseline-separation
- **Sequence**: 101
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- `internal/thermal/live.go`
- `internal/thermal/live_test.go`
- `internal/tui/live.go`
- `internal/tui/live_test.go`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- [Evidence and constraints](facts.md)
- [Execution tasks](plan.md)
- Prerequisite: live-baseline-separation

## Requirements
1. **One-second bucket contract**: Each `RollingTokens` and `RollingTurns` element is the activity observed in one distinct wall-clock second. A delta observed at `now` covers `(lastPoll, now]`. Allocate it across the whole seconds in that interval in chronological order; when integer division leaves a remainder, assign the remainder deterministically to the most recent second (or earliest, but document and test one rule). A bucket must never receive the whole multi-second delta.
2. **Sub-second and long intervals**: If less than one second has elapsed, accumulate the delta in the current second and flush when the next second boundary is crossed. If more than 60 seconds have elapsed, only the most recent 60 seconds enter the rolling window; older tokens/turns still count in session totals and events but not in Rate, Speed, Peak, or the sparkline. Clock jumps backward must not produce negative buckets or negative rates.
3. **Peak is per-second**: Rename `PeakRollingTok` to `PeakTokensPerSec` and `PeakRollingTurn` to `PeakStepsPerSec` (JSON `peakTokensPerSec`, `peakStepsPerSec`), and add `bucketSeconds: 1` and `windowSeconds: 60` to the snapshot. Peak equals the maximum one-second bucket, not the maximum event. Update the JSON migration entry for the key rename and semantic clarification.
4. **Single window**: `BurnTokensPerMin` is the sum of the 60 one-second buckets; `BurnTokensPerSec` is that sum divided by 60. Apply the same rule to turns/steps. The TUI must use these canonical values for both Rate and Speed and must not derive a second, differently rounded rate.
5. **Consistent formatting**: Share one value formatter between `formatRate` and `formatSpeed`. For values below 10, Rate shows one decimal (`7.8 step/m`) and Speed shows two decimals (`0.13 step/s`); values at or above 10 use compact numbers. Never truncate a positive sub-10 value through `int64()` to `0` or `0.0`. For turns, a test must assert `abs(speed*60 - rate) <= 0.05` for every displayed pair.
6. **Labels state the unit**: The sparkline caption reads `Peak: X tok/s (1s bucket)` or `Peak: X step/s (1s bucket)`, and the section title states `60s window · 1s buckets`. The 60-second sum remains `60s Burn: X tok` or `60s Steps: X step`.
7. **Events stay per-poll**: The ticker still shows the delta observed between polls; do not fabricate per-second ticker rows. Only the rolling rate/peak series is bucketed.
8. **Safety**: No new dependencies; preserve the fixed-clock test style and the bounded/read-only collector. Keep diagnostic output off JSON stdout.

## Done Condition
1. A 2-second poll adding 20K tok produces two 10K buckets and `PeakTokensPerSec == 10K`, not 20K; `BurnTokensPerMin == 20K`; `BurnTokensPerSec == 333.3`.
2. A 90-second gap adds only the most recent 60 seconds to the rolling window while the full delta remains in session/event totals.
3. A 0.5-second tick accumulates into the current second and does not create a short bucket.
4. `8 step/min` and `0.13 step/s` are displayed together from the same 60-second sum; no `0.1 step/s` contradiction remains.
5. The TUI labels Peak as a one-second bucket and the section as a 60-second window at 80x24 and 120x40.
6. JSON exposes `peakTokensPerSec`, `peakStepsPerSec`, `bucketSeconds`, and `windowSeconds`; migration documentation and `project.jsonc` are synchronized.
7. Focused tests, full race suite, CLI build, and simulated user gate pass; the attribution commit ends in `(goals/live-second-bucket-accounting/goal.md)`.

## Scope Boundary
No session/baseline seeding changes (prerequisite `live-baseline-separation`), no ticker column/cost/outlier changes (`live-ticker-unit-and-cost-clarity`), no flame ceiling policy (`live-flame-scale-contract`), and no rollup/sort (`live-ticker-rollup-and-cost-sort`). This goal may expose `Peak*` values that the flame goal consumes.
