# Facts: Define and expose the live flame ceiling per metric

## Observed Evidence
- `internal/tui/live.go:307-320` computes `intensity := math.Max(m.snapshot.FlameIntensity, m.snapshot.ActivityIntensity)`, then switches the burn rate/speed source based on `isActivity`. The gauge therefore mixes a token-event bonus and a step-event bonus.
- `internal/thermal/live.go:840-847` increments `flameIntensity` by 0.5 per token delta and `activityIntensity` by 0.4 per turn delta, capped at 1.0. `internal/thermal/live.go:163-190` decays both by 0.85 per elapsed second. There is no volume denominator.
- `internal/thermal/live.go:914-938` exposes `FlameIntensity` and `ActivityIntensity` without a metric or ceiling. `internal/tui/live.go:752-763` maps the combined 0-1 value to BURNING/WARM/COOLING/IDLE at 0.6/0.2/0.02 thresholds.
- `internal/tui/live.go:765-812` renders only `Flame: [bar] 37% WARM` plus a model name. It never states the ceiling or whether the percentage is tokens or steps.
- The reported contradiction is reproducible from the code: 23 step/min can produce a high intensity from event bonuses, while 3.4M tok/min produces 0.3-0.5 depending on poll count, not on the volume.
- `internal/thermal/live.go:881-902` already provides the canonical 60-second `burnPerMin`/`burnPerSec` values. After `live-second-bucket-accounting`, these are the correct denominator inputs; no separate intensity accumulator is needed.
- `internal/thermal/live_test.go:730-892` asserts the seeded flame is >0 on `seedToday`; `internal/thermal/live_test.go:367-451` asserts activity-only token flame is zero but activity intensity is positive. These tests encode the old event-bonus behavior and must be rewritten.
- The CLI registers flags at `cmd/thermal/main.go:74-121` and then manually parses positional-verb flags at `:140-303`. A compact value like `9M` cannot be registered as `Float64Var` because `flag.Parse()` would reject it before the manual loop; `--interval` is the precedent for dual parsing. Validation for command-scoped flags lives around `:435-453`.

## Interpretation
The flame must be a ratio, not an event counter. The 60-second rolling rate already captures burstiness and decay, so `rate/ceiling` is sufficient and deterministic. The gauge must state both the metric and the ceiling; otherwise a percentage is uninterpretable. A single `FlameIntensity` with `FlameMetric` and `FlameCeiling` fields is clearer than two opaque intensities.

Defaults: 9M tok/min is the ceiling named in the feedback example; 60 step/min is one step per second, a documented and configurable activity ceiling. Both yield ~38% WARM for the reported values, which is the expected honest reading.

## Ownership and Constraints
- Tracker fields/formula: `internal/thermal/live.go`, `internal/thermal/live_test.go`.
- Gauge and badge rendering: `internal/tui/live.go`, `internal/tui/live_test.go`.
- Flag parsing/validation and JSON wiring: `cmd/thermal/main.go`, `cmd/thermal/simulated_user_test.go`.
- Public contract: `docs/MIGRATION.md`, `project.jsonc`.
- Preserve the baseline exclusion and one-second bucket contract. Do not add a second rate calculation.
- No new dependencies; keep `--no-color` and width tests.

## Verification Contract
```bash
go test -race ./internal/thermal ./internal/tui ./cmd/thermal -run 'Live|Flame|Ceiling'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```

Exact cases:
- `3_400_000 tok/min / 9_000_000 tok/min = 0.3778` -> 38%, WARM, caption `9.0M tok/m`.
- `23 step/min / 60 step/min = 0.3833` -> 38%, WARM, caption `60 step/m`.
- `--flame-ceiling=4.5M` on 3.4M tok/min -> 0.7556 -> 76%, BURNING.
- `--flame-ceiling=0`, `--flame-ceiling=abc`, and `thermal daily --flame-ceiling=9M` -> exit 1 with a clear stderr message.
- `thermal live --json` includes `flameIntensity`, `flameMetric`, and `flameCeiling`; `activityIntensity` is absent.
