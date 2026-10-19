# Goal: Separate live baseline carry-in from real-time activity and session counters

## Goal Description
The live monitor treats its first poll as a session seed: it copies the current day's persisted totals into `SessionTokens`, `SessionTurns`, and `SessionCost`, sets a warm flame, and emits one event per tool timestamped at the current clock. That is why a 531.5M tok row appears at 18:06:26 and why TODAY'S USAGE and LIVE SESSION show the same 537.1M tok / 12.0K step / $4.09 values. The first read is carried-in history, not a burst.

Change the contract so the first poll records opening balances as **carried in**, excludes them from session counters, rolling buckets, flame, Burn/Hr, and peak, and makes LIVE SESSION launch-relative. TODAY'S USAGE remains the persisted local-day total. The UI must show the carried-in amount separately and must never render two identical Today/Session panels.

Priority: **P1**. Derived from the 2026-10-19 live-monitor feedback (bugs 1-2; improvements "Exclude or flag backfill" and "Split the panels meaningfully"). The closed `live-metric-parity` goal established canonical token/turn filtering but deliberately kept `seedToday` as a display convenience; this goal removes that exception.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 100
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
- Closed contract: `goals/live-metric-parity/goal.md`
- Existing migration entry: `docs/MIGRATION.md` "Changed: Live Monitor Metric Parity and Burn Accounting"

## Requirements
1. **Baseline contract**: Add `CarriedInTokens int64`, `CarriedInTurns int`, `CarriedInCost float64`, and `CarriedInAt time.Time` to `LiveSnapshot` with JSON tags `carriedInTokens`, `carriedInTurns`, `carriedInCost`, and `carriedInAt`. Add `IsBaseline bool` with `json:"isBaseline,omitempty"` to `LiveEvent` for any baseline display row. Values are captured from the first poll's persisted local-day totals after canonical token-only filtering.
2. **First poll is inert**: On initialization, store the carried-in values but do not add them to `SessionTokens`, `SessionTurns`, `SessionCost`, `burnBuckets`, `turnBuckets`, `FlameIntensity`, `ActivityIntensity`, `BurnTokensPerMin`, `BurnTokensPerSec`, `BurnTurnsPerMin`, `BurnTurnsPerSec`, `BurnCostPerHr`, `RollingTokens`, `RollingTurns`, `PeakRollingTok`, or `PeakRollingTurn`. `Poll` must return zero new events on the first call. `RecentEvents` must not contain a synthetic `+531.5M tok` burst.
3. **Launch-relative session**: Remove the `seedToday`/`SetSeedToday` session-seeding path. After baseline initialization, only observed deltas increment session counters. `ResetSession` continues to zero session counters without touching carried-in values, Today, or baselines. `thermal live` and `thermal live --json` both use the launch-relative behavior.
4. **`--fresh` semantics**: `--fresh` must not re-enable seeded session totals. Keep the flag accepted for `thermal live` compatibility (it remains valid for `thermal statusline`), update its live help text to state that live sessions are always launch-relative, and document the changed behavior in `docs/MIGRATION.md`. Do not remove the flag in this goal.
5. **Today vs Session**: TODAY'S USAGE stays the persisted local-day total (token-only volume, turns/steps, cost), but it must not read as live burn when carried-in dominates it. Add a separate `Carried in:` line below the KPI grid when any carried-in value is non-zero, formatted with the compact number helpers, e.g. `Carried in: 531.5M tok · 12.0K step · ~$3.61 (baseline)`, and make the new-since-launch portion visible through the launch-relative Live Session panel (or an explicit `new since launch` annotation). The carried-in line must never be rendered as a `+` burst row. If session counters are all zero, the LIVE SESSION panel shows an explicit `Since launch: no new activity` state (or is merged into Today) rather than duplicating Today's values. If `TodayTokens == SessionTokens && TodayTurns == SessionTurns && TodayCost == SessionCost`, render one panel, not two identical panels.
6. **Midnight and reset**: On a local-day rollover, recompute carried-in from the new day's persisted totals (normally zero), keep session counters continuous, and do not replay the previous day as a burst. `ResetSession` (`[c]`) must not recreate the baseline row or warm the flame.
7. **JSON and stream**: `thermal live --json` exposes carried-in fields and zero session counters at startup. `thermal live --json --stream` emits the initial `snapshot` line with carried-in fields and no baseline `event` lines. Any future baseline display event must carry `isBaseline: true` and must not be counted in rate/flame/session math.
8. **Diagnostics and safety**: Keep ingestion read-only, keep diagnostics on stderr via `Summary.Warnings`, and never pollute JSON stdout. No new dependencies, no background daemon, and preserve the sub-10ms cached invocation target.

## Done Condition
1. A fixed-clock first poll with 531.5M tok / 12.0K step / $3.61 today yields `SessionTokens == 0`, `SessionTurns == 0`, `SessionCost == 0`, zero burn/peak/flame, `CarriedInTokens == 531.5M`, and no new event.
2. A second poll adding +10K tok produces only a +10K event, `SessionTokens == 10K`, and `CarriedInTokens == 531.5M`; Today equals the persisted sum.
3. Activity-only and hybrid inputs preserve token/step disjointness and produce no token burn from steps.
4. TUI at 80x24 and 120x40 renders the carried-in line, does not render a synthetic `+531.5M` ticker row, and does not render identical Today/Session panels.
5. `thermal live --json` and `--json --stream` match the new contract, with a six-part migration entry and synchronized `project.jsonc`.
6. `go test -race ./...`, `go build -o /tmp/thermal-test ./cmd/thermal`, and `./scripts/simulated_user_gate.sh` pass; the attribution commit uses a `feat`/`fix`/`refactor`/`perf` subject ending in `(goals/live-baseline-separation/goal.md)`.

## Scope Boundary
No one-second bucket redistribution or Peak renaming (owned by `live-second-bucket-accounting`), no ticker unit/cost/outlier redesign (owned by `live-ticker-unit-and-cost-clarity`), no flame ceiling/configuration (owned by `live-flame-scale-contract`), and no ticker rollup or sort toggle (owned by `live-ticker-rollup-and-cost-sort`). This goal may add the `IsBaseline`/carried-in fields those goals consume, and must set baseline flame to zero.
