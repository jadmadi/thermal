# Facts: Give --top one documented contract and make every remediation hint valid

## Observed Evidence (2026-10-29, current source build `thermal dev`)
| Command | Output | Exit |
| --- | --- | --- |
| `thermal --top 5` | `thermal: --top only applies to the projects and models commands` | 1 |
| `thermal daily --top 5` | same message | 1 |
| `thermal daily --chart --top 100` | same message | 1 |
| `thermal projects --top 3` | 3 project rows | 0 |
| `thermal models --top 3` | 3 model rows | 0 |
| `thermal changelog --top 5` | `Showing 5 of 25 releases` | 0 |
| `thermal changelog` | `Showing 3 of 25 releases`, no affordance | 0 |
| `thermal daily --chart` | footer `… and 84 more rows (raise --top to see them in the table)` | 0 |
| `thermal --top -1` | `thermal: --top cannot be negative` | 1 |

The daily table already shows all 96 active days; the chart caps at 12. The footer tells the user to raise a flag that the command rejects, and claims the hidden rows are in the table (they already are).

## Root Cause
- `cmd/thermal/main.go:450-452` (leaderboard), `:522-524` (trend/mix/stats), and `:553-555` (default report branch) return `--top only applies to the projects and models commands`.
- `cmd/thermal/main.go:471-473` (statusline) and `:487-489` (live) return `--top does not apply to the <command> command`.
- `cmd/thermal/main.go:496-502` (changelog) validates `Breakdown`, `Chart`, and `Sort` but never rejects `Top`; `:602` uses `limit := opts.Top`.
- `cmd/thermal/main.go:674`/`:1110`/`:1149` use `opts.Top` for models/projects/changelog.
- `internal/render/chart.go:68` hard-codes the footer:
  ```go
  fmt.Sprintf("… and %d more rows (raise --top to see them in the table)", omitted)
  ```
  without checking whether the command accepts `--top` or whether the table is already uncapped.
- `internal/render/projects.go:95` and `models.go:120` use a different footer: `… and N more (use --json for the full list)`.
- `internal/render/help.go` describes `--top <N>  Cap table rows to top N items (default: all)`, implying every table.
- `cmd/thermal/main.go:82` declares `flag.IntVar(&opts.Top, "top", 0, "Project or model rows to print, 0 for all")`, a fourth description.

## Architectural Invariants & Constraints
- One flag, one applicability table, one message.
- No remediation hint may recommend a flag the command rejects.
- Table, chart, and JSON must agree on the effective row count for the same invocation.
- Totals always describe the full window, never the capped subset.
- `--top` is presentation only; it must not change aggregation, sorting, or pricing.
- `--top 0` means all rows; the changelog default of 3 requires tracking whether the flag was explicitly set.
- Usage errors exit `2`; data errors exit `1`.

## File & Interface Contracts
- `cmd/thermal/main.go`:
  - `validateReportFlags` `--top` branches replaced by one `topApplies(report) bool` helper and a canonical message.
  - Pass `opts.Top` to the leaderboard and period report renderers.
  - Changelog limit: `if !opts.TopSet { limit = 3 } else { limit = opts.Top }`.
- `cmd/thermal/args.go`: `invocation.Present["top"]` or `opts.TopSet bool` set by the parser.
- `internal/render/top.go` (new):
  ```go
  func EffectiveRows(total, top int) int
  func ChartRows(total, top, defaultCap int) int
  func ChartFooter(total, shown, top int) string
  func TableFooter(total, shown int) string
  ```
- `internal/render/chart.go`: replace the literal footer with `ChartFooter`.
- `internal/render/leaderboard.go`, period renderers, `projects.go`, `models.go`, `changelog.go`: use `EffectiveRows`.
- `internal/render/help.go`: `--top` description and per-command listing.

## Applicability Matrix (frozen)
| Command | Accepted | Default | Meaning |
| --- | --- | --- | --- |
| leaderboard | yes | all | ranked tool rows |
| daily | yes | all | period rows after order |
| weekly | yes | all | week rows |
| monthly | yes | all | month rows |
| projects | yes | all | project rows |
| models | yes | all | model rows |
| changelog | yes | 3 | release cards |
| stats | no | — | — |
| mix | no | — | — |
| trend | no | — | — |
| audit | no | — | — |
| share | no | — | — |
| live | no | — | — |
| statusline/prompt/status | no | — | — |

## Footer Contract
```
# Accepted, chart capped, table uncapped
… and 84 more rows (chart shows 12 of 96; use --top 96 for all)

# Accepted, --top 5, chart capped at 5
chart shows 5 of 96 rows (use --top 96 for all)

# Accepted, --top 0 (all), chart capped at 12
… and 84 more rows (chart shows 12 of 96; use --top 96 for all)

# Not accepted (hypothetical)
… and 84 more rows
```
- Table footer and chart footer must state the same omitted/returned counts.
- The chart cap (12) is named.
- No `raise --top to see them in the table` sentence.

## Verification Contract
```bash
# Acceptance matrix
thermal --top 5 >/dev/null; test $? -eq 0
thermal daily --top 5 >/dev/null; test $? -eq 0
thermal weekly --top 3 >/dev/null; test $? -eq 0
thermal monthly --top 3 >/dev/null; test $? -eq 0
thermal projects --top 3 >/dev/null; test $? -eq 0
thermal models --top 3 >/dev/null; test $? -eq 0
thermal changelog --top 5 >/dev/null; test $? -eq 0

for cmd in stats mix trend audit share live statusline; do
  thermal $cmd --top 5 >/dev/null 2>/tmp/e; test $? -eq 2
  grep -q 'does not apply' /tmp/e
  ! grep -q 'only applies to the projects and models' /tmp/e
done

# Row counts
test "$(thermal --top 5 --json | jq '.data|length')" -eq 5
test "$(thermal daily --top 5 --json | jq '.data|length')" -eq 5
diff <(thermal daily --json | jq '.totals') <(thermal daily --top 5 --json | jq '.totals')

# Chart footer
thermal daily --chart | grep -q 'chart shows 12 of'
! thermal daily --chart | grep -q 'raise --top to see them in the table'
thermal daily --chart --top 5 | grep -q 'chart shows 5 of'

# Changelog
thermal changelog | grep -q 'Showing 3 of 25'
thermal changelog --top 5 | grep -q 'Showing 5 of 25'
thermal changelog --top 0 | grep -q 'Showing 25 of 25'

CGO_ENABLED=0 go test -race ./cmd/thermal/... ./internal/render/...
./scripts/simulated_user_gate.sh
```

## Regression Fixtures
- A 96-day fixture for daily/weekly/monthly so caps are visible.
- A leaderboard with 11 tools and `--sort streak|tokens|cost`.
- A projects fixture with 20 rows, a models fixture with 25 rows.
- A changelog with 25 releases.
- Chart captures for `--top` unset/0/5/12/13/100.
- JSON captures for the same values.

## Open Questions
- Should `--top` cap JSON `data` or only text? Recommendation: cap `data` consistently and keep `totals` full; document it.
- Should the changelog default remain 3? Recommendation: yes, but add the affordance; `--top 0` shows all.
- Should `--top` be accepted on the leaderboard? Recommendation: yes; it removes the asymmetry and makes the chart footer valid.
- Should period reports keep the newest or oldest N with `--order asc`? Recommendation: the first N in the displayed order, so the flag composes with `--order`.
- Should the chart default cap become 12 or `--top`? Recommendation: keep 12 as the chart cap, but allow `--top` to raise it up to the table rows.
