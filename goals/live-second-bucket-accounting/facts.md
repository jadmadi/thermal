# Facts: Bucket live deltas into explicit one-second bins so Peak, Rate, and Speed share one window

## Observed Evidence
- `internal/thermal/live.go:163-190` advances the rolling ring by `int(elapsed.Seconds())` and zeroes each crossed bucket. If more than 60 seconds elapse, it caps the advance at 60.
- `internal/thermal/live.go:840-847` adds the **entire** `deltaTok`/`deltaTurns` observed since the previous poll to the single current bucket, after the ring has already advanced. With a 2-second poll interval, a 20K delta lands in one bucket and is reported as 20K tok/s.
- `internal/thermal/live.go:881-902` computes `sumRolling` and `peakRolling` by scanning the 60 buckets, then `burnPerMin = sumRolling` and `burnPerSec = burnPerMin/60`. The math is internally consistent only if each bucket is one second of activity; the write path violates that invariant.
- `internal/thermal/live.go:914-938` exposes `peakRollingTok` and `peakRollingTurn` without a unit or window description.
- `internal/tui/live.go:814-865` labels the peak as `Peak: X tok/s` / `Peak: X step/s` and the sparkline as `ROLLING ACTIVITY (60s)`.
- `internal/tui/live.go:568-588` formats `turnRate` with `thermal.CompactNumber(int64(turnRate))`, truncating fractional per-minute values; `internal/tui/live.go:590-630` formats `turnSpeed` as a float. The same 60-second sum can therefore display as `8 step/m` and `0.1 step/s` (6/min).
- `cmd/thermal/main.go:103` defaults `--interval` to 1s, but `cmd/thermal/main.go:282-288` accepts any duration or float seconds, and `:1993-1996` uses it directly. The bug is observable at the default only when polling is delayed; it is deterministic at `--interval=2s` or greater.
- Existing test `internal/thermal/live_test.go:49-121` polls at a 2-second interval and asserts only that burn is positive; it never asserts the bucket distribution or peak value. `internal/tui/live_test.go:564-588` tests fractional speeds but not Rate/Speed parity.

## Interpretation
A bucket is currently a poll-sized accumulator despite being labeled per-second. There are two honest fixes: label peak as "largest poll delta" (rejected because the UI claims tok/s), or distribute each delta across the seconds it spans. This goal chooses distribution, which preserves the existing 60-bucket ring, keeps the 60-second sum equal to observed volume, and makes Peak a true one-second maximum. For intervals longer than 60 seconds, tokens older than the window are intentionally excluded from Rate/Peak but remain in session totals.

The Rate/Speed contradiction is a formatting bug, not a second data source: both values derive from `sumTurns`. Use one canonical float and render it with sufficient precision on both lines.

## Ownership and Constraints
- Tracker math and JSON fields: `internal/thermal/live.go`, `internal/thermal/live_test.go`.
- Presentation: `internal/tui/live.go`, `internal/tui/live_test.go`.
- Public contract: `docs/MIGRATION.md`, `project.jsonc`.
- Do not modify loader ingestion or tool databases.
- Preserve the prerequisite baseline exclusion: carried-in volume never enters buckets.
- No new dependencies; keep the ring fixed-size and allocation-free on idle ticks.

## Verification Contract
```bash
go test -race ./internal/thermal ./internal/tui -run 'Live|Bucket|Peak|Rate'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

Deterministic cases:
- t0 baseline; t0+2s +20K -> buckets `[10K, 10K]`, peak 10K/s, sum 20K, speed 333.3 tok/s.
- t0+90s +90K -> only the most recent 60K in the rolling window, peak <= 1K/s if evenly distributed, session +90K.
- t0+500ms +1K; t0+1s +1K -> the first 1K flushes into the first second, the second 1K lands in the next; no bucket exceeds one second of activity.
- `8 turns/min` -> `8.0 step/m` (or `8 step/m`) and `0.13 step/s`, never `0.1 step/s`.
