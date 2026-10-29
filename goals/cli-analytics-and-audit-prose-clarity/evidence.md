# Evidence: cli-analytics-and-audit-prose-clarity

Date: 2026-10-30
Project: thermal-streak
Goal: cli-analytics-and-audit-prose-clarity

## Implementation Summary
- Added `meanBasis:"active-day"` to stats JSON and `meanBasis:"calendar-day"` + `activeDays` to trend JSON.
- Stats now labels `Mean / active day` and prints `Basis: N active days`.
- Trend now splits its header into `Trajectory`, `Calendar window`, `Mean / calendar day`, and `Basis: N calendar days`; the card labels `Mean / calendar day`, `Point estimate`, and `Band (95% residual)`.
- Projection metadata adds `bandMethod:"residual-sd-1.96"`; a footnote explains the least-squares fit, ±1.96 residual SD, and zero clamp.
- Outlier prose is one sentence (`Outliers: N days above X (median + 2 MAD); top 5 shown.`) plus `MAD = median absolute deviation.`; the old `all 5 top days above, plus` run-on is gone.
- Zero-count histogram bins are omitted and disclosed with `N empty bin(s) omitted.`.
- Audit prints `Score components:`, `Largest penalty:`, and `Biggest win: Address …` as separate lines; it selects the recommendation mentioning `AGENTS.md` when present and strips `Critical:`/`Warning:` prefixes.
- Added a PASS/WARN/CRIT rule-file legend.
- Added named tests and updated README, migration, project metadata, gallery, and gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...` → PASS.
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (157/157 checks).
- `thermal stats --no-color` → `Mean / active day 245.0M`, `Basis: 97 active days`, `Outliers: 29 days above 227.1M (median + 2 MAD); top 5 shown.`, `MAD = median absolute deviation.`, `1 empty bin omitted.`
- `thermal trend --no-color` → `Trajectory: [RISING ↗ …]`, `Calendar window: 400 days · …`, `Mean / calendar day: ≈59.4M`, `Basis: 400 calendar days`, `Point estimate 205.8M`, `Band (95% residual) 0 to 611.3M`, method footnote.
- `thermal audit --no-color` → separate `Score components:`, `Largest penalty: MCP density.`, `Biggest win: Address AGENTS.md …`, `Legend: PASS = no action; WARN = moderate context tax; CRIT = high …`.
- JSON → stats `meanBasis:"active-day"`; trend `meanBasis:"calendar-day"`, `activeDays` number, `projection.bandMethod:"residual-sd-1.96"`.
- `COLUMNS=80` stats/trend/audit → 0 lines over 80.

## Acceptance Criteria

- AC-001: Stats contains `Mean / active day` and `Basis: N active days`; it does not label the value only `Mean`. **PASS**
- AC-002: Trend contains `Mean / calendar day` and `Basis: N calendar days`. **PASS**
- AC-003: The 4-active/10-calendar fixture (`TestStatsMeanBasis`, `TestTrendMeanBasis`) yields stats mean 25 and trend mean 10 with the basis fields and labels. **PASS**
- AC-004: Trend contains `Point estimate`, `Band (95% residual)`, and a footnote explaining the least-squares method and zero clamp. **PASS**
- AC-005: The outlier line matches `Outliers: N days above …; top 5 shown.`; `all 5 top days above, plus` is absent. **PASS**
- AC-006: Zero-count histogram bins are absent; `N empty bin(s) omitted.` is present. **PASS**
- AC-007: Audit contains `Score components:`, `Largest penalty:`, and `Biggest win: Address …` on separate lines; the biggest win starts with a verb, contains `AGENTS.md` in the fixture, and does not start with `Critical:`. **PASS**
- AC-008: The audit rule-file legend defines PASS, WARN, and CRIT. **PASS**
- AC-009: Stats JSON `.meanBasis=="active-day"`; trend JSON `.meanBasis=="calendar-day"` and `.activeDays` is a number. **PASS**
- AC-010: Trend JSON `.projection.bandMethod=="residual-sd-1.96"`. **PASS**
- AC-011: At `COLUMNS=80`, stats, trend, and audit have no line over 80. **PASS**
- AC-012: `go test -race ./internal/render/... ./internal/thermal/... ./cmd/thermal/...` passes, including `TestStatsBasisLabel`, `TestTrendBasisLabel`, `TestOutlierSentence`, `TestHistogramOmitsEmptyBins`, `TestAuditBiggestWinLine`, `TestStatsMeanBasis`, and `TestTrendMeanBasis`. **PASS**
- AC-013: `./scripts/simulated_user_gate.sh` passes 157/157 and includes the analytics-prose section. **PASS**

## Regression Notes
- The underlying means, totals, and aggregation math are unchanged; only labels and metadata are new.
- JSON additions are additive (`meanBasis`, `activeDays`, `projection.bandMethod`).
- Gate checks for outliers/projection are conditional on the fixture actually containing them.
- At ≥96 columns the established card layout is preserved.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-analytics-and-audit-prose-clarity",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Name analytics denominators and split audit prose", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_24a5d588", "intent": "Name analytics denominators and split audit prose"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
