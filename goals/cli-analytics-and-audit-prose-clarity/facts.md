# Facts: Name denominators and intervals and remove run-on or zero-noise report prose

## Observed Evidence (2026-10-29, current source build `thermal dev`)
### `stats`
```
  Thermal · stats · tokens

  Active days  96
  Total        23.3B
  Mean         242.7M
  Median       76.2M
  p90          712.4M
  Max          2.0B
  ...
  Outliers above 225.1M (median + 2 MAD): all 5 top days above, plus 22 more
```
- `Mean` = `23.3B / 96` active days.
- The denominator footnote appears only for `--metric cost`:
  `Mean, median, p90, max, weekday profile, and distribution use costed days; active days include zero-cost days.`
- The distribution prints a zero-count bin: `807.3M — 1.0B  ························  0`.

### `trend`
```
  Trajectory: [RISING ↗ +720.9K (1.2%/day)]  Calendar window: [399 days · 2025-09-26 to 2026-10-29]  Daily Mean: ≈58.4M

  ╭─ Activity Trend ──────────────────────────────────────────────────── tokens ─╮
  │  Range        2025-09-26 to 2026-10-29                                       │
  │  Days         399                                                            │
  │  Mean         58.4M                                                          │
  │  Slope        +720.9K per day                                                │
  │  Direction    rising                                                         │
  │  Projection to 2026-10-31 (2 days)                                           │
  │  Expected     203.3M                                                         │
  │  Band         0 to 609.1M                                                    │
  ╰──────────────────────────────────────────────────────────────────────────────╯
```
- `Mean` = `23.3B / 399` calendar days (96 active + 303 zero).
- `Band` = `expected ± 1.96 × residualStd` (`internal/thermal/trend.go:167`), lower bound clamped to 0; no label states the method or confidence.

### `audit`
```
  Score: [59/100 · Grade D]  Context Tax: ≈21.8K tokens / turn (Rules: ≈8.5K tok, MCP: ≈13.3K tok)
  Score components: rules -15 · MCP -21 · readiness -5. Largest: MCP density. Biggest win:
  Critical: AGENTS.md (~/.agents) is 343 lines (~5641 tok). Split into specialized skills or trim
  to <150 lines.
```
- The `Biggest win:` label is followed by a line break and then `Critical:` (a severity), not an action.
- The rule-file table legend defines `[CRIT]` only; `PASS` and `WARN` are used but undefined.

## Root Cause
- `internal/render/analytics.go:185-195`: the stats rows are `{"Mean", …}`, `{"Median", …}`, etc.; no basis label.
- `internal/thermal/stats.go:199`: `rep.Mean = rep.Total / float64(len(nums))`, where `nums` are active-day values.
- `internal/render/analytics.go:408`: `Daily Mean` in the trend header; `:422` uses `{"Mean", metric(rep.Mean)}` in the card.
- `internal/thermal/trend.go:93-97`: points include every calendar day from first to last, including zeros; `rep.Mean /= float64(len(rep.Points))`.
- `internal/render/analytics.go:337-345`: the outlier sentence builds `all %d top days above, plus %d more`.
- `internal/render/analytics.go:299-310`: the histogram iterates every bin, including `Count == 0`.
- `internal/thermal/trend.go:167`: `band := 1.96 * residualStd(...)`; `:170-172` clamps `low` to 0.
- `internal/render/analytics.go:507`: the projection row is labeled `Band`.
- `internal/render/audit.go:84-87`:
  ```go
  explanation := fmt.Sprintf("Score components: rules -%d · MCP -%d · readiness -%d. Largest: %s.", ...)
  if len(rep.Recommendations) > 0 {
      explanation += " Biggest win: " + rep.Recommendations[0]
  }
  ```
  The recommendation string already begins with a severity (`Critical:`), producing the double label.

## Architectural Invariants & Constraints
- The arithmetic must not change; only labels, sentences, and omitted zero rows.
- Denominators must be explicit and correct: stats = active days (or costed days for cost), trend = calendar days (including zero).
- The trend band is a residual-based prediction band, not a confidence interval for the mean; the label must not overclaim.
- Outlier counts must be computed, not inferred from the top-N list.
- Empty histogram bins may be omitted only with a stated count.
- Audit recommendations must be actionable, path-specific, and free of spliced severity labels.
- All prose must fit the width contract; wrap with a hanging indent.
- JSON additions are additive; `meanBasis`, `activeDays`, `bandMethod`, `residualStd` must not rename existing keys.

## File & Interface Contracts
- `internal/render/analytics.go`:
  - `RenderStats(rep thermal.StatsReport, noColor bool) string`
  - `RenderTrend(rep thermal.TrendReport, noColor bool) string`
  - add `basisLine(metric string, active, calendar, costed int) string`
  - add `outlierSentence(threshold float64, shown, total int, metric string) string`
  - filter `rep.Histogram` bins with `Count == 0` and track the omitted count.
- `internal/render/audit.go`:
  - split `explanation` into `scoreComponents` and `biggestWin`;
  - strip a leading severity token from the recommendation and render it as `(critical)`;
  - add the PASS/WARN/CRIT legend.
- `internal/thermal/stats.go`: expose `MeanBasis string` and `ActiveDays`/`CostedDays` if not already present.
- `internal/thermal/trend.go`: expose `MeanBasis`, `ActiveDays`, `BandMethod`, `ResidualStd` (additive JSON fields).
- `internal/render/format.go`: `LabelValue` and `Note` helpers.

## Text Contract
```
  Per active day (96 days with activity)
  Mean / active day   242.7M
  Median              76.2M
  p90                 712.4M
  Max                 2.0B
```
```
  Mean / calendar day  ≈58.4M
  Basis: 399 calendar days (96 active, 303 zero)
  ...
  Projection to 2026-10-31 (2 days)
  Point estimate       203.3M
  95% band (±1.96 residual SD)  0 to 609.1M
  · Least-squares fit; band = ±1.96 × residual SD, lower bound clamped at 0.
```
```
  Outliers: 27 days above 225.1M (median + 2 MAD); top 5 shown.
```
```
  Score components: rules -15 · MCP -21 · readiness -5
  Largest penalty: MCP density (-21)
  Biggest win: trim AGENTS.md (~/.agents, 343 lines) to <150 lines
  Legend: PASS = within budget · WARN = approaching · CRIT = high context tax
```
- No `Biggest win: Critical:` sequence.
- No `all 5 top days above, plus`.
- No zero-count histogram row without an omission note.

## JSON Contract (additive)
```json
{"type":"stats","metric":"tokens","meanBasis":"active-day","activeDays":96,...}
{"type":"trend","metric":"tokens","meanBasis":"calendar-day","activeDays":96,
 "projection":{"expected":203300000,"low":0,"high":609100000,
               "bandMethod":"residual-sd-1.96","residualStd":310000000,...},...}
```
- Existing `mean` values and keys remain.
- `bandMethod` and `residualStd` are optional if the math does not expose them; document the chosen names.

## Verification Contract
```bash
stats=$(thermal stats --no-color)
printf '%s' "$stats" | grep -q 'Mean / active day'
printf '%s' "$stats" | grep -q 'Basis: .* active days'
printf '%s' "$stats" | grep -q 'Outliers: .*; top 5 shown'
! printf '%s' "$stats" | grep -q 'all 5 top days above, plus'

trend=$(thermal trend --no-color)
printf '%s' "$trend" | grep -q 'Mean / calendar day'
printf '%s' "$trend" | grep -q 'Basis: .* calendar days'
printf '%s' "$trend" | grep -q 'Point estimate'
printf '%s' "$trend" | grep -q '95% band'
printf '%s' "$trend" | grep -q 'residual SD'

audit=$(thermal audit --no-color)
printf '%s' "$audit" | grep -q 'Score components:'
printf '%s' "$audit" | grep -q 'Largest penalty:'
printf '%s' "$audit" | grep -q 'Biggest win:'
! printf '%s' "$audit" | grep -q 'Biggest win: Critical:'

thermal stats --json | jq -e '.meanBasis=="active-day"' >/dev/null
thermal trend --json | jq -e '.meanBasis=="calendar-day" and (.activeDays|type)=="number"' >/dev/null

CGO_ENABLED=0 go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- 4 active days over 10 calendar days with total 100: stats mean 25, trend mean 10.
- 0, 5, and 27 outlier days.
- A histogram with 3 empty bins.
- A projection with a negative lower bound (clamped to 0).
- Audit recommendations beginning with `Critical:`, `Warning:`, and no severity.
- Audit rule files with PASS, WARN, and CRIT rows.
- Golden text at 80 and 120 columns.

## Open Questions
- Should `Mean` become `Mean / active day`, or keep `Mean` and add a `Basis:` line? Recommendation: both — the label and the basis line.
- Should the trend band be called `95% prediction band` or `95% band (±1.96 residual SD)`? Recommendation: the latter, with a footnote; it is precise and does not overclaim.
- Should the stats block be titled `Per active day` and repeat the basis once? Recommendation: one block title plus the explicit basis line.
- Should the audit severity stay in parentheses? Recommendation: yes: `Biggest win (critical): ...`; never splice it after the label.
- Are `activeDays` in trend already available? If not, count points with `Value > 0` and document that zero-value days are the complement.
