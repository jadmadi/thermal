# Capacity claim remediation CAP-2026

## Context and files

P1. Cwd `/mnt/Jad/github/lab/thermal-streak`. Depends on workload-replay R1 and R2.
Why: a DailyRow has a date, not the intraday events required to calculate a rolling five-hour peak.
Read `internal/thermal/replay.go:SimulatePlan`, `internal/thermal/types.go:DailyRow`,
`internal/pricing/subscription.go` and `tasks/plan.md`.
Edit `internal/thermal/replay.go`, `internal/thermal/replay_test.go`,
`internal/render/replay.go`, `internal/render/replay_test.go`, plus common companion docs.

## Fixed decision and acceptance

Design choice: disclose limits of existing daily telemetry, rather than add timestamp ingestion.
For plans with RollingWindowLimit > 0, do not treat a day's sum as an observed rolling-window peak.
Add `UNKNOWN` to the existing CapacityVerdict vocabulary when daily limits do not already prove
failure but rolling capacity cannot be assessed. Explicitly explain missing intraday evidence.
Do not mark UNKNOWN recommended. A daily-limit exceedance remains a daily failure, with rolling
spikes unasserted (BurstThrottles remains zero for unavailable evidence). No change to plan prices.
A daily total below a rolling limit still cannot exclude a cross-midnight rolling peak.
Text must describe fixed token ceilings as configured scenario assumptions, not verified provider
entitlements. No financial or purchasing recommendation follows from UNKNOWN.

- CAP-201: add `TestReplayDailyDataCannotProveRollingCapacity`: same daily total for hypothetical
  concentrated and spread workloads produces UNKNOWN, not an invented burst count; below-window
  daily totals also remain UNKNOWN; daily ceiling exceeded remains a daily failure.
- CAP-202: add `TestRenderReplayUnknownCapacity`: UNKNOWN visible in monochrome, no recommendation;
  migration entry documents new verdict. `go test -race ./internal/thermal ./internal/render
  -run Replay` exits 0. Then common final checks.

No new dependencies, event store, provider calls, rate refresh, or exact rolling algorithm.
If exact rolling acceptance is required, return blocked for a separate source/coverage design.
Stop/return/review contract: `tasks/plan.md`.

## Scope correction

This current contract supersedes the historical promise of exact five-hour simulation from daily
rows. Exact intraday simulation is deferred until timestamp coverage has its own approved design.
Historical tasks remain recorded; they are not additional contradictory closure requirements.
