# Plan: Define and expose the live flame ceiling per metric

## Execution Steps
- [ ] **Task 1 — Replace event bonuses with a rate/ceiling formula**: Remove the `flameIntensity`/`activityIntensity` increments and decays from `LiveTracker.Poll`. Compute one `FlameIntensity` from the canonical 60-second rate and the selected ceiling. Set `FlameMetric` to `tokens` or `steps` with explicit token-first hybrid precedence.
  Acceptance and verification: `go test -race -run 'TestLiveTracker_FlameScale|TestLiveTracker_FlameMetricSelection' ./internal/thermal`; 3.4M/9M and 23/60 both yield ~0.38; zero activity yields 0; hybrid token and step deltas only affect the token gauge.

- [ ] **Task 2 — Publish ceiling and metric in the snapshot**: Remove `ActivityIntensity`, add `FlameMetric` and `FlameCeiling` to `LiveSnapshot`, and update all consumers. Document the removed/changed JSON fields in `docs/MIGRATION.md`.
  Acceptance and verification: `go test -race -run TestLiveSnapshot_FlameFields ./internal/thermal`; JSON has `flameIntensity`, `flameMetric`, `flameCeiling`, and no `activityIntensity`.

- [ ] **Task 3 — Render the ceiling and metric**: Update `formatFlameLine` to show `Flame: [bar] 38% of 9.0M tok/m  WARM` or `38% of 60 step/m  WARM`. Use compact ceiling formatting and keep the existing state thresholds, now explicitly relative to the ceiling. Update `renderHelp` to document the formula.
  Acceptance and verification: `go test -race -run 'TestLiveTUI_FlameCeilingCaption|TestLiveTUI_FlameStates' ./internal/tui` at 80x24 and 120x40; the caption contains the metric, ceiling, percentage, and state; no line shows a bare percentage.

- [ ] **Task 4 — Add and validate `--flame-ceiling`**: Add `FlameCeiling float64` and a raw-string field to `thermal.Options` (or an equivalent local), register a string flag, parse compact `K`/`M`/`B` values after positional processing, and validate positive/finite/command scope. Pass the value to both the TUI tracker and the JSON tracker.
  Acceptance and verification: `go test -race -run 'TestCLI_FlameCeiling|TestLiveJSON_FlameCeiling' ./cmd/thermal`; valid values produce the expected percentages; invalid/misplaced values exit 1 with a clear stderr message.

- [ ] **Task 5 — Rewrite seed/flame tests**: Replace tests that assert the seeded 0.3 flame and dual activity intensity with fixed-clock rate/ceiling cases. Preserve the prerequisite baseline rule that carried-in volume produces zero flame.
  Acceptance and verification: `go test -race -run 'TestLiveTracker_.*Flame|TestLiveModel_.*Flame' ./internal/thermal ./internal/tui`.

- [ ] **Task 6 — End-to-end and width verification**: Add a synthetic live JSON case at 3.4M tok/min and 23 step/min; verify the TUI caption and JSON fields at 80x24/120x40. Check `--no-color` output for ANSI leaks and width overflow.
  Acceptance and verification: `go test -race -run TestLiveFlameEndToEnd ./cmd/thermal`; `thermal live --json` and `thermal live --no-color` snapshots agree on intensity and ceiling.

- [ ] **Task 7 — Migration, metadata, and attribution commit**: Add the six-part migration entry for the new flag and JSON fields, synchronize `project.jsonc`, run the full gate, and create an impact-first commit ending in `(goals/live-flame-scale-contract/goal.md)`. Run `sila goals` and do not push.

## Dependencies and Handoff
Start after `live-baseline-separation` and `live-second-bucket-accounting` close. Use their `CarriedIn*` exclusion and one-second bucket rates; do not reintroduce event-based intensity or a second rate. The later rollup/sort goal must not alter flame semantics.

## Verification
```bash
go test -race ./internal/thermal ./internal/tui ./cmd/thermal -run 'Live|Flame|Ceiling'
go test -race ./...
go build -o /tmp/thermal-test ./cmd/thermal
./scripts/simulated_user_gate.sh
```
