# Goal: Name denominators and intervals and remove run-on or zero-noise report prose

## Goal Description
Priority: P2 (output clarity). The analytics commands are numerically correct but their prose hides the denominator. `stats` divides by active days; `trend` divides by every calendar day in the window, including zero days. Both label the result "Mean", four times apart in the same dataset, with no explanation. The audit's "biggest win" sentence splices a severity label into the recommendation, and the stats histogram prints empty bins.

Observed on the current source build (2026-10-29):

- `thermal stats`:
  ```
  Active days  96
  Total        23.3B
  Mean         242.7M
  ```
  `Mean` is `23.3B / 96` (active days).
- `thermal trend`:
  ```
  Trajectory: [RISING ↗ +720.9K (1.2%/day)]  Calendar window: [399 days · 2025-09-26 to 2026-10-29]  Daily Mean: ≈58.4M
  ```
  `Mean` is `23.3B / 399` (calendar days, including 303 zero days).
- `thermal stats` outlier line:
  `Outliers above 225.1M (median + 2 MAD): all 5 top days above, plus 22 more`
- `thermal stats` distribution includes a zero-count bin:
  `807.3M — 1.0B  ························  0`
- `thermal trend` projection card:
  ```
  Projection to 2026-10-31 (2 days)
  Expected     203.3M
  Band         0 to 609.1M
  ```
  The band is `expected ± 1.96 * residualStd` (`internal/thermal/trend.go:167`), a 95% normal-approximation prediction band, with the lower bound clamped to 0. The label "Band" states none of that.
- `thermal audit`:
  ```
  Score components: rules -15 · MCP -21 · readiness -5. Largest: MCP density. Biggest win:
  Critical: AGENTS.md (~/.agents) is 343 lines (~5641 tok). Split into specialized skills or trim
  to <150 lines.
  ```
  The `Biggest win:` sentence is split across the `Largest:` sentence, and the recommendation starts with `Critical:` (a severity) rather than an action.
- The audit rule-file table legend defines `[CRIT]` only; `PASS` and `WARN` are unexplained.
- The stats footnote about denominators exists only for `--metric cost`:
  `Mean, median, p90, max, weekday profile, and distribution use costed days; active days include zero-cost days.`

The outcome of this goal is **precise analytics prose**: every statistic names its denominator, the trend interval is labeled with its confidence and method, the outlier sentence states a count instead of a self-reference, empty histogram bins are omitted, and the audit recommendation is one actionable sentence with its path and without a spliced severity label.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 253
- **Shape**: ship
- **Tier**: roadmap
- **Memory Protocol**: required
- **Wave**: 4 (after the width and totals goals settle the layout)
- **Related implemented goals (do not regress)**: `audit-path-and-score-clarity` owns audit paths and score explanation; this goal owns the *sentence shape* and the denominator/interval labels. `output-report-totals-reconciliation` owns numeric reconciliation; this goal must not change numbers. `output-formatting-primitives` owns symbol semantics.

## Files to Touch
- `internal/render/analytics.go`
- `internal/render/audit.go`
- `internal/thermal/stats.go`
- `internal/thermal/trend.go`
- `internal/render/analytics_test.go`
- `internal/render/audit_test.go`
- `cmd/thermal/analytics_test.go`
- `scripts/simulated_user_gate.sh`
- `README.md`
- `docs/MIGRATION.md`
- `project.jsonc`

## References
- **Shared Understanding & Fact Sheet**: [`goals/cli-analytics-and-audit-prose-clarity/facts.md`](facts.md)
- **Execution Plan**: [`goals/cli-analytics-and-audit-prose-clarity/plan.md`](plan.md)

## Requirements

1. **Denominator labels in `stats`.**
   - Replace `Mean` with `Mean / active day`; keep the value.
   - Add a line: `Basis: 96 active days (days with activity in the window).`
   - `Median`, `p90`, and `Max` are per active day; label the block `Per active day`.
   - For `--metric cost`, keep the existing costed-day distinction and add `Costed days: N (M zero/unpriced excluded)`.
   - Do not change the arithmetic.
   - Add `meanBasis: "active-day"` to the stats JSON (additive).

2. **Denominator labels in `trend`.**
   - Replace `Daily Mean` and the card `Mean` with `Mean / calendar day` and add `Basis: 399 calendar days (96 active, 303 zero).`
   - The `Calendar window` label stays; it explains the denominator.
   - Do not change the arithmetic.
   - Add `meanBasis: "calendar-day"` to the trend JSON (additive).
   - If the active-day count is not currently in the trend report, compute it from points with `Value > 0` (or a day with turns/tokens > 0) and expose it as `activeDays` in JSON.

3. **Trend projection labels.**
   - Rename `Expected` to `Point estimate`.
   - Rename `Band` to `95% band (±1.96 residual SD)` or `95% prediction band`; add a footnote: `Least-squares fit; band = ±1.96 × residual standard deviation, lower bound clamped at 0.`
   - State the horizon explicitly (`2 days to 2026-10-31`).
   - If the residual SD is available, print it; if not, print the band width.
   - Do not imply a confidence interval for the mean; call it a prediction band.
   - Keep `low`/`high` JSON keys; add `bandMethod:"residual-sd-1.96"` and `residualStd` if available.

4. **Stats outlier and histogram prose.**
   - Replace `all 5 top days above, plus 22 more` with:
     `Outliers: 27 days above 225.1M (median + 2 MAD); top 5 shown.`
   - Compute the total outlier count, not just the extra count.
   - If there are zero outliers: `Outliers: none above <threshold> (median + 2 MAD).`
   - If there are exactly the top N: `Outliers: 5 days above <threshold>; all shown.`
   - Omit histogram bins with `Count == 0`; if any are omitted, append `N empty bins omitted.`
   - Ensure the remaining bin ranges are contiguous and the last bin includes the max.
   - Keep the 16-cell bar width or the width-goal equivalent.

5. **Audit prose.**
   - Split the score explanation into separate lines:
     ```
     Score components: rules -15 · MCP -21 · readiness -5
     Largest penalty: MCP density (-21)
     Biggest win: trim AGENTS.md (~/.agents, 343 lines) to <150 lines
     ```
   - The `Biggest win:` line must be a single actionable sentence, start with a verb, name the path/scope, and not start with `Critical:`, `Warning:`, `WARN`, or `CRIT`.
   - If the recommendation has a severity, render it at the start of the line in parentheses: `Biggest win (critical): ...`.
   - Add a legend for `PASS`, `WARN`, `CRIT` in the rule-file table: `PASS = within budget · WARN = approaching · CRIT = high context tax`.
   - Keep the existing score formula and component values.
   - Preserve the path/scope work from `audit-path-and-score-clarity`.

6. **Shared prose helpers.**
   - Add `render.LabelValue(label, value string, width int)` and `render.Note(text string, width int)` to keep alignment and wrapping consistent.
   - No sentence may exceed the width contract; wrap with a hanging indent.
   - Do not use a colon inside a colon-led label (`Biggest win: Critical: ...`).
   - Keep `--no-color` plain text.

7. **Tests.**
   - A fixture with 4 active days over 10 calendar days and a total of 100: `stats` mean 25 (`Mean / active day`), `trend` mean 10 (`Mean / calendar day`), both with the correct basis lines.
   - Outlier wording for 0, 5, and 27 outliers.
   - Zero-count bins omitted; count of omitted bins.
   - Trend projection labels and method footnote.
   - Audit biggest-win line: one sentence, starts with a verb, contains the path, no `Critical:` prefix.
   - Audit legend defines all three severities.
   - JSON `meanBasis` fields and `activeDays` in trend.
   - Golden text at 80/120 columns.

8. **Documentation and migration.**
   - README analytics section explains active-day vs calendar-day and the trend band.
   - MIGRATION: label changes are text-only; JSON additions are `meanBasis`, `activeDays`, `bandMethod`, `residualStd`; no numeric changes.
   - `project.jsonc` updated.
   - `scripts/simulated_user_gate.sh`: assert the two basis lines, the outlier sentence, the band label, and the audit recommendation shape.

## Non-Goals
- Changing the mean, median, percentile, slope, or projection math.
- Changing the outlier threshold (median + 2 MAD).
- Changing the audit score formula or penalties.
- Changing the width/layout engine.
- Adding new analytics commands.

## Acceptance Criteria
- AC-001: `thermal stats --no-color` contains `Mean / active day` and `Basis: 96 active days` (or the fixture equivalent); it does not label the value only `Mean`.
- AC-002: `thermal trend --no-color` contains `Mean / calendar day` and `Basis: 399 calendar days` (or the fixture equivalent).
- AC-003: In a 4-active/10-calendar-day fixture with total 100, `stats` shows `25` and `trend` shows `10`, each with its basis label.
- AC-004: The trend card contains `Point estimate` and a band label containing `95%` and `residual`; a footnote explains the method and the 0 clamp.
- AC-005: The stats outlier line matches `Outliers: 27 days above ...; top 5 shown.` and does not contain `all 5 top days above, plus`.
- AC-006: Zero-count histogram bins are absent; if any are omitted, the line `N empty bins omitted.` is present.
- AC-007: The audit output contains `Score components:`, `Largest penalty:`, and `Biggest win:` as separate lines; the biggest-win line starts with a verb, contains `AGENTS.md`, and does not start with `Critical:`.
- AC-008: The audit rule-file table legend defines `PASS`, `WARN`, and `CRIT`.
- AC-009: `thermal stats --json | jq -e '.meanBasis=="active-day"'` and `thermal trend --json | jq -e '.meanBasis=="calendar-day" and (.activeDays|type)=="number"'` exit `0`.
- AC-010: `thermal trend --json | jq -e '.projection.bandMethod=="residual-sd-1.96"'` exits `0` (or the documented key).
- AC-011: At `COLUMNS=80`, `thermal stats`, `thermal trend`, and `thermal audit` produce no line >80 (coordinated with the width goal).
- AC-012: `go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...` exits `0`, including `TestStatsBasisLabel`, `TestTrendBasisLabel`, `TestOutlierSentence`, `TestHistogramOmitsEmptyBins`, and `TestAuditBiggestWinLine`.
- AC-013: `./scripts/simulated_user_gate.sh` exits `0` and includes the analytics-prose section.

## Negative Cases & Failure Modes
| Condition | Required result |
| --- | --- |
| Tokens metric | `Mean / active day`, basis line |
| Cost metric | costed-day basis + zero/unpriced note |
| Trend | `Mean / calendar day`, basis line |
| 4 active / 10 calendar | stats 25, trend 10 |
| 0 outliers | `Outliers: none above ...` |
| 5 outliers, all shown | `Outliers: 5 ...; all shown.` |
| 27 outliers | `Outliers: 27 ...; top 5 shown.` |
| Empty histogram bin | omitted + count |
| Projection | `Point estimate`, `95% ... residual`, footnote |
| Audit biggest win | one actionable sentence, path, no severity prefix |
| Audit legend | PASS/WARN/CRIT defined |
| Narrow width | wrapped with hanging indent, no line > width |

## Metrics & Regression Guards
- Denominator labels present on 100% of stats/trend renders.
- Fixture math: stats mean = total/active days; trend mean = total/calendar days; a test would fail if the denominators swap.
- Outlier sentence: no self-referential phrase; count equals the actual number of outliers.
- Histogram: no zero-count row; omitted count accurate.
- Audit: biggest-win line is one sentence and contains a path.
- Gate: simulated user gate asserts AC-001, AC-002, AC-004, AC-005, and AC-007.

## Risks & Mitigations
- **Risk:** Longer labels push the layout past 80 columns. **Mitigation:** width goal coordination; label on its own line, value aligned; wrap the basis line.
- **Risk:** `meanBasis` adds a JSON key that consumers ignore. **Mitigation:** additive and documented; no existing key changes.
- **Risk:** Renaming "Band" to "95% prediction band" overstates the statistics. **Mitigation:** state the method precisely (`±1.96 × residual SD`, lower bound clamped); do not call it a confidence interval.
- **Risk:** Omitting empty bins hides distribution gaps. **Mitigation:** print `N empty bins omitted` and keep the bin ranges contiguous.
- **Risk:** Rewriting the audit recommendation loses the severity. **Mitigation:** render severity in parentheses at the start of the line, not spliced after the label.

## Memory Plan
- **Prime**: run `sila prime` at task start.
- **Recall**: run `sila recall "stats mean active days trend calendar days"` and `sila recall "audit biggest win sentence"`.
- **Writes**: record a `case` for the "explicit denominator/interval labels" mechanism, grounded to `internal/render/analytics.go`; record a `lesson` for the two-means confusion if a user-reported issue is found.
- **Facts**: run `sila facts verify` before certification.
- **Handoff**: run `sila handoff "analytics prose clarity shipped" --open="statistical method docs"`.
- **Evidence**: embed a `sila/memory-writes/v1` block under `## Memory Writes`.

## Done Condition
1. All tasks in `plan.md` are checked (`- [x]`).
2. Implementation is committed with `(goals/cli-analytics-and-audit-prose-clarity/goal.md)` under a `fix(render):` subject.
3. Goal is certified and `evidence.md` contains a standalone `Verdict: PASS` plus the fixture math and prose captures.
4. `sila goals --json` reports `status: "implemented"`, `verification_status: "verified"`, and `unverified_count: 0`, with a clean `git status`.
