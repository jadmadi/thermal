# Facts: Separate live baseline carry-in from real-time activity and session counters

## Observed Evidence
- `internal/tui/live.go:64-84` defaults `seedToday` to true in `NewLiveModel`; `cmd/thermal/main.go:2005` and `:2073` pass `!opts.Fresh` into both the JSON tracker and the TUI model. The `--fresh` help text at `cmd/thermal/main.go:107` promises "Start live session counters from 0 instead of today's total."
- `internal/thermal/live.go:257-477` is the first-poll branch. When `seedToday` is true, lines 370-373 copy `todayTok`, `todayTurns`, and `todayCost` into `sessionTokens`, `sessionTurns`, and `sessionCost`; lines 374-379 set flame/activity intensity to 0.3; lines 381-472 emit one `LiveEvent` per tool with today's totals and `Timestamp: now` (line 461).
- That path renders the reported 531.5M tok row at 18:06:26 and makes TODAY'S USAGE and LIVE SESSION identical. It is a display baseline, not a delta.
- `internal/thermal/live.go:212-245` computes `todayTok` from daily rows and filters activity-only rows; `internal/thermal/live.go:248-252` computes the cache-hit rate. These Today values are correct persisted totals and should remain.
- `internal/thermal/live.go:125-131` only zeroes session counters in `ResetSession`; it does not clear the seeded event history or carried-in values.
- `internal/tui/live.go:328-342` renders Today from `TodayTokens/TodayTurns/TodayCost` and Session from `SessionTokens/SessionTurns/SessionCost` without any carried-in field or empty-session state.
- `internal/thermal/live.go:15-51` defines the live JSON schema. There is no carried-in field and no baseline marker.
- `cmd/thermal/main.go:2003-2059` writes the JSON snapshot and streams NDJSON events. The initial snapshot is emitted once; subsequent `newEvents` lines include whatever the first poll produced.
- The current tests encode the bug: `internal/thermal/live_test.go:730-892` (`TestLiveTracker_SeedTodayBaselineAndDelta`) expects session counters seeded with today's totals, a 0.3 flame, and a baseline event; `internal/tui/live_test.go:422-482` (`TestLiveModel_SeedTodayAndActivityUnits`) expects the default model to seed 40K session tokens and 150 session turns.
- Existing `live-metric-parity` (closed) established canonical token-only totals and turn tracking but did not remove `seedToday`; its migration entry at `docs/MIGRATION.md:478-503` says nothing about a launch-relative session.

## Interpretation
The 531.5M row is not a burst and should not be treated as one. It is the opening balance of the local day at process start. The correct split is:
- **Today**: persisted local-day total, including the opening balance.
- **Session**: deltas observed after process start, starting at zero.
- **Carried in**: the opening balance, displayed once and excluded from all rates.

The duplicate panels are a direct consequence of copying Today into Session on the first poll. Removing that copy fixes both symptoms without changing loader data.

## Ownership and Constraints
- `internal/thermal/live.go` and `internal/thermal/live_test.go` own tracker semantics.
- `internal/tui/live.go` and `internal/tui/live_test.go` own rendering and key flow.
- `cmd/thermal/main.go` and `cmd/thermal/simulated_user_test.go` own CLI/JSON wiring and end-to-end checks.
- `docs/MIGRATION.md` and `project.jsonc` own the public contract record.

Preserve unrelated working-tree changes. Do not modify tool databases or loader read paths. Keep aggregation in `internal/thermal`; the TUI must not re-derive totals. Keep diagnostics on stderr and JSON stdout clean. No new dependencies. Maintain the bounded collector and sub-10ms cached invocation target.

## Verification Contract
```bash
go test -race ./internal/thermal ./internal/tui ./cmd/thermal -run 'Live|Baseline|Sim'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

Use an isolated synthetic home and fixed clocks. The JSON/stream check must assert:
- `carriedInTokens == 531500000` (or the synthetic equivalent)
- `sessionTokens == 0`, `sessionTurns == 0`, `sessionCost == 0`
- `burnTokensPerMin == 0`, `peakRollingTok == 0`, `flameIntensity == 0`
- the first stream line is `type: snapshot`; no `type: event` baseline line follows
