# Evidence: cli-report-total-row-semantics

Date: 2026-10-30
Project: thermal-streak
Goal: cli-report-total-row-semantics

## Implementation Summary
- Added `ProjectDays` and `ToolCount` to `ProjectRow`; `ModelDays` and `ToolCount` to `ModelRow`.
- Project aggregation now tracks a distinct global calendar-day set, sets total `activeDays` to that count, `projectDays` to the sum of per-project active days, and `toolCount` to the distinct tool count.
- Model aggregation does the same with `modelDays`.
- Project/model total rows now print `N tools` (or spaces), `projectDays`/`modelDays`, and the full-window token/cost sums; non-additive cells are blank rather than `—`.
- Added footers naming `per project`/`per model` and the distinct calendar-day denominator.
- Added `TestProjectTotalsDistinctDays`, `TestModelTotalsDistinctDays`, `TestTotalRowAdditivity`, `TestTotalRowBlankCells`, and `TestMixTotals`.
- Updated README, migration guide, project metadata, gallery, and simulated gate.

## Commands Executed
- `go test ./...` → PASS.
- `go test -race ./internal/thermal/... ./internal/render/... ./cmd/thermal/...` → PASS.
- `./scripts/output_gallery_gate.sh` → PASS.
- `./scripts/simulated_user_gate.sh` → PASS (153/153 checks).
- `thermal projects --no-color` total row → `12 tools ... 507` with no `Agy · CMD +10`.
- Projects footer → `Tools are counted per project; P-Days are project-days (507); distinct calendar days are 129.`
- Models footer → `Tools are counted per model; M-Days are model-days (191); distinct calendar days are 78.`
- `thermal projects --json` totals → `activeDays=129`, `projectDays=507`, `toolCount=12`; `models` → `activeDays=78`, `modelDays=191`, `toolCount=11`.
- `projects --top 3 --json .totals` → identical to the uncapped totals.
- `daily --json` → `.data` token sum equals `.totals.totalTokens`; cost sum within $0.01.
- `mix --json` → bucket total sum equals `.total`; series shares sum to 1.

## Acceptance Criteria

- AC-001: The projects total Tools cell is `12 tools` (matches `^[0-9]+ tools$`); it contains no `·` and no `+N`. **PASS**
- AC-002: `TestProjectTotalsDistinctDays` (two projects, same day) yields `activeDays=1`, `projectDays=2`, `toolCount=2`. **PASS**
- AC-003: `TestModelTotalsDistinctDays` (two models, same day) yields `activeDays=1`, `modelDays=2`, `toolCount=2`. **PASS**
- AC-004: `TestTotalRowAdditivity` proves project/model token, cost, and day totals equal the sum of rows; the daily JSON gate proves period additivity. **PASS**
- AC-005: `TestMixTotals` proves bucket volume sums to the total and shares sum to 1 ± 0.01. **PASS**
- AC-006: The projects footer contains `per project`, `distinct calendar days`, and the project-day count; models contains the model equivalent. **PASS**
- AC-007: `TestTotalRowBlankCells` proves blank total cells are spaces; zero cost prints `$0.00`, not `—`. **PASS**
- AC-008: Projects JSON contains `totals.projectDays` and `totals.toolCount`; models contains `totals.modelDays` and `totals.toolCount`. **PASS**
- AC-009: `projects --top 3 --json .totals` equals the uncapped totals. **PASS**
- AC-010: `go test -race ./internal/thermal/... ./internal/render/... ./cmd/thermal/...` passes with the five named tests. **PASS**
- AC-011: `./scripts/simulated_user_gate.sh` passes 153/153 and includes the total-row section. **PASS**

## Regression Notes
- `activeDays` for project/model rows remains per-row; only totals use distinct calendar days.
- The total `Tools` list remains in JSON for compatibility; text uses `toolCount`.
- `--top` still caps rows only; totals are full-window.
- At ≥96 columns the established table layout is preserved.

Verdict: PASS

## Memory Writes

```json
{
  "schema": "sila/memory-writes/v1",
  "goal_slug": "cli-report-total-row-semantics",
  "project": "thermal-streak",
  "prime": {"ran": true, "command": "sila prime"},
  "recall": [{"query": "Make project and model totals additive and distinguish day denominators", "matches": 5}],
  "memos": [],
  "lessons": [],
  "cases": [{"id": "fact_bb5dce75", "intent": "Make project and model totals additive and distinguish day denominators"}],
  "facts_verify": {"exit_code": 0},
  "handoff": {"id": "hoff_6b0f7b33b96c"}
}
```
