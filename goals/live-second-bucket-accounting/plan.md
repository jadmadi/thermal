# Plan: Bucket live deltas into explicit one-second bins so Peak, Rate, and Speed share one window

## Execution Steps
- [x] **Task 1 — Define the bucket allocator**: Add a helper in `internal/thermal/live.go` that advances the ring by whole seconds and distributes a token/turn delta across the seconds it spans, with a documented remainder rule. Reject non-positive elapsed time without creating buckets.
  Acceptance and verification: `go test -race -run TestLiveTracker_SecondBucketDistribution ./internal/thermal`; a 2s +20K delta yields two 10K buckets and a 10K peak.

- [x] **Task 2 — Cap the rolling window correctly**: When elapsed exceeds 60 seconds, advance/fill only the most recent 60 seconds and keep older activity in session/event totals only. Ensure the 60-second sum equals the sum of the ring.
  Acceptance and verification: `go test -race -run 'TestLiveTracker_LongIntervalWindow|TestLiveTracker_RingSumInvariant' ./internal/thermal`; a 90s +90K delta leaves 60K in the ring and 90K in session.

- [x] **Task 3 — Rename Peak and add units**: Replace `PeakRollingTok`/`PeakRollingTurn` with `PeakTokensPerSec`/`PeakStepsPerSec`, add `BucketSeconds` and `WindowSeconds` to `LiveSnapshot`, and update all TUI/JSON consumers. Document the key rename and semantic clarification in `docs/MIGRATION.md`.
  Acceptance and verification: `go test -race -run TestLiveSnapshot_PeakUnits ./internal/thermal`; JSON contains the new keys and no ambiguous `peakRolling*` keys.

- [x] **Task 4 — Unify Rate and Speed formatting**: Refactor `formatRate`/`formatSpeed` in `internal/tui/live.go` to consume the same canonical floats and a shared precision helper. Preserve compact output above 10 and fractional output below 10; no `int64()` truncation for positive sub-10 values.
  Acceptance and verification: `go test -race -run 'TestLiveTUI_RateSpeedParity|TestLiveTUI_FormatSpeed_Fractional' ./internal/tui`; for turns, `abs(speed*60-rate) <= 0.05` across 1..999 values.

- [x] **Task 5 — Label the window and bucket size**: Update the sparkline section title and Peak caption to state `60s window · 1s buckets` and `Peak: X tok/s (1s bucket)`/`Peak: X step/s (1s bucket)`. Keep the 60s sum display.
  Acceptance and verification: `go test -race -run TestLiveModel_BucketLabels ./internal/tui` at 80x24 and 120x40; both views contain the bucket/window labels and no bare `Peak:` without a unit.

- [x] **Task 6 — Regression and edge cases**: Cover sub-second ticks, backward clock jumps, activity-only turns, hybrid token/turn data, and `--interval=2s` end-to-end. Ensure the prerequisite baseline fields never enter a bucket.
  Acceptance and verification: `go test -race -run 'TestLiveTracker_.*Bucket|TestLiveTracker_ClockJump' ./internal/thermal`; `thermal live --json --interval=2s` with synthetic deltas reports a one-second peak.

- [x] **Task 7 — Migration, metadata, and attribution commit**: Update the migration entry for the JSON key rename and `--interval` semantics; synchronize `project.jsonc`. Run the full gate and create an impact-first commit ending in `(goals/live-second-bucket-accounting/goal.md)`. Run `sila goals` and do not push.

## Dependencies and Handoff
Start only after `live-baseline-separation` closes. Read its `CarriedIn*` contract before editing the first-poll path. Preserve the carried-in exclusion; baseline volume must never be distributed into buckets.

## Verification
```bash
go test -race ./internal/thermal ./internal/tui -run 'Live|Bucket|Peak|Rate'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
