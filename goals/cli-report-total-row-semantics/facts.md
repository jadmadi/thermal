# Facts: Make report total rows contain only additive, dimension-consistent values

## Observed Evidence (2026-10-29, current source build `thermal dev`)
- `thermal projects --no-color` total row:
  ```
  │        Total                         Agy · CMD +10             23.3B   ~$503.59     505  2026-10-29 │
  ```
- `thermal models --no-color` total row:
  ```
  │       Total                           CMD · Claude +9           22.0B   ~$614.41     190  2026-10-29 │
  ```
- `thermal` leaderboard footer reports `1.3B tokens have no model recorded` and 96 active days in the daily report; the projects total says 505 P-Days and models says 190 M-Days.
- `thermal daily --last 7` total row: `22.3M` input + `1.8M` output + `194.7K` reason + `785.0M` cache = `809.3M` tokens; additive and correct.
- `thermal mix` total row: volumes sum and percentages are recomputed; correct.
- `thermal projects` footer: `Total = $242.09 recorded + ~$261.50 estimated from pricing data.` — correct arithmetic.
- `thermal models` footer: `Cost is estimated from models.dev list prices; recorded session cost is not attributable to one model.` plus the classified/unclassified split.

## Root Cause
- `internal/thermal/types.go:295-300`: `ProjectReport.Totals` is a `ProjectRow`, so it inherits `Project`, `Tools`, `ActiveDays`, `FirstDay`, and `LastDay`, which are row-shaped fields.
- `internal/thermal/projects.go:197-229` (`sumProjects`):
  ```go
  total.ActiveDays += row.ActiveDays
  ...
  for _, tool := range row.Tools {
      if !containsString(total.Tools, tool) { total.Tools = append(total.Tools, tool) }
  }
  ```
  and min `FirstDay`, max `LastDay`.
- `internal/render/projects.go:99-110` builds the total cells with:
  ```go
  padRight(toolsCell(totals.Tools, toolsWidth), toolsWidth)
  padLeft(fmt.Sprintf("%d", totals.ActiveDays), daysWidth)
  ```
  so the union list and the summed day count are printed as if they were ordinary row values.
- `internal/render/models.go` mirrors this for models.
- `internal/thermal/report.go:322` already sets `total.ActiveDays = len(totalActive)` for daily/weekly/monthly, so the period reports deduplicate correctly; the project/model aggregations do not.

## Architectural Invariants & Constraints
- A total row is a summary of the report window, not a sum of every cell. Additive cells must equal the row sum; distinct/min/max/count cells must be labeled and computed independently.
- A total row must not invent a new entity (a tool union is not a tool).
- Blank is not `—`: the em dash means "unpriced" in the cost vocabulary.
- Totals are independent of `--top`; capping rows must not change totals.
- JSON keys are additive; the `activeDays` semantic change must be documented.
- The width contract applies to the total row as part of the table; dropped columns are dropped from header, body, and total together.

## File & Interface Contracts
- `internal/thermal/types.go`:
  - add `ProjectDays int \`json:"projectDays,omitempty"\`` and `ToolCount int \`json:"toolCount,omitempty"\`` to `ProjectRow` (or a dedicated `ProjectTotals` struct);
  - add `ModelDays int \`json:"modelDays,omitempty"\`` and `ToolCount int \`json:"toolCount,omitempty"\`` to `ModelRow`.
- `internal/thermal/projects.go`:
  - `sumProjects(rows []ProjectRow) ProjectRow` computes:
    ```go
    days := map[string]struct{}{}
    // for each row, the row must expose its day set or the aggregate must receive it
    total.ActiveDays = len(days)
    total.ProjectDays += row.ActiveDays
    total.ToolCount = len(total.Tools)
    ```
  - If `ProjectRow` does not carry a day set, extend the aggregation input with one; do not approximate from `ActiveDays`.
- `internal/thermal/models.go`: the model equivalent.
- `internal/render/totals.go` (new):
  ```go
  type TotalKind int // totalSum, totalDistinct, totalCount, totalMax, totalBlank, totalRecomputed
  type TotalCell struct { Text string; Kind TotalKind; AlignRight bool }
  func ProjectTotalCells(t ProjectRow, widths []int) []TotalCell
  func ModelTotalCells(t ModelRow, widths []int) []TotalCell
  ```
- `internal/render/projects.go`, `internal/render/models.go`: render the cells from the helper, not ad hoc.

## Total Semantics (frozen)
| Report | Cell | Semantics | Example |
| --- | --- | --- | --- |
| projects | Project | blank | ` ` |
| projects | Tools | count | `13 tools` |
| projects | Tokens | sum | `23.3B` |
| projects | Cost | sum | `~$503.59` |
| projects | P-Days | distinct calendar days | `96` |
| projects | Last | max | `2026-10-29` |
| models | Model | blank | ` ` |
| models | Tools | count | `10 tools` |
| models | Tokens | sum | `22.0B` |
| models | Cost | sum | `~$614.41` |
| models | M-Days | distinct calendar days | `96` |
| models | Last | max | `2026-10-29` |
| daily/weekly/monthly | numeric | sum | — |
| daily/weekly/monthly | Active days | distinct | — |
| mix | shares | recomputed | — |

## Footer Contract
```
P-Days are per project; the total uses distinct calendar days
(505 project-days across 13 projects).
```
```
M-Days are per model; the total uses distinct calendar days
(190 model-days across 10 models).
```
- Wrap to the width; keep the existing cost footer line.
- Do not use `—` for a blank name cell.

## JSON Contract (additive)
```json
"totals": {
  "activeDays": 96,
  "projectDays": 505,
  "toolCount": 13,
  "totalTokens": 23300000000,
  "cost": 503.59
}
```
```json
"totals": {
  "activeDays": 96,
  "modelDays": 190,
  "toolCount": 10,
  "totalTokens": 22000000000,
  "estimatedCost": 614.41
}
```
- `activeDays` changes meaning for totals from "sum of per-row days" to "distinct calendar days"; document it.
- Per-row `activeDays` is unchanged.
- `tools` may remain as the union for compatibility; the renderer uses `toolCount`.
- Totals are unaffected by `--top`.

## Verification Contract
```bash
# Union tools no longer printed
thermal projects --no-color | grep 'Total' | grep -q '·' && exit 1
thermal models --no-color | grep 'Total' | grep -q '·' && exit 1

# Overlap fixtures
# project A and B both active on 2026-10-01
thermal projects --json | jq -e '.totals.activeDays==1 and .totals.projectDays==2 and .totals.toolCount==2'
thermal models --json | jq -e '.totals.activeDays==1 and .totals.modelDays==2 and .totals.toolCount==2'

# Additivity (daily/weekly/monthly)
# parse the table and compare each numeric column to the sum of rows
CGO_ENABLED=0 go test -race ./internal/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- Two projects active on the same day, different tools.
- Two models active on the same day, different tools.
- One project / one model (activeDays == projectDays/modelDays).
- A project active on three days and another on two, with one overlapping day.
- Cost fixtures: recorded-only, estimated-only, mixed.
- `--top 3` on a 20-row report to prove totals are full-window.
- Golden total rows at 80 and 120 columns.

## Open Questions
- Should the total P-Days be distinct calendar days or remain the sum? Recommendation: distinct calendar days, with `projectDays` in JSON and a footer explaining the sum. This matches the leaderboard/daily.
- Should the Tools cell be blank or `N tools`? Recommendation: `N tools`; it is informative and unambiguous.
- Should `totals.tools` remain? Recommendation: keep it for compatibility but document that it is a union; add `toolCount`.
- Should the `Last` total be labeled `Latest`? Recommendation: keep the header and add a footer note; avoid a one-off label in the total row.
- Should models show the distinct-day total even though recorded cost is not attributable? Recommendation: yes; days are a usage dimension, independent of cost.
