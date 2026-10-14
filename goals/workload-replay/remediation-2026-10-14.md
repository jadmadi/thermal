# Replay remediation RPL-2026

## Context

P1. Cwd `/mnt/Jad/github/lab/thermal-streak`. Ready, no prerequisites. Cards R1 then R2.
Why: dropped volume and missing prices currently produce misleading savings recommendations.
Read `internal/thermal/types.go:WorkloadSnapshot`, `ModelTokens.Total`, `internal/thermal/replay.go`,
`internal/loaders/sessions.go`, and the audit probes. All common rules in `tasks/plan.md` apply.
Do not change subscription rates, loader accounting or infer token categories from a model name.

## R1: Preserve volume and expose incomplete price coverage

Allowed: `internal/thermal/replay.go`, `internal/thermal/types.go`,
`internal/thermal/replay_test.go`, plus common companion docs.
Keep public function signatures `BuildWorkloadSnapshot`, `SimulateModel`, `SimulatePlan`,
`AggregateReplay`. Add `WorkloadSnapshot.UnclassifiedTokens int64` with JSON `unclassifiedTokens`.
`TotalTokens` sums canonical in-window token-bearing `DailyRow.Tokens`, independent of model maps.
When model categories cover the full day, use their disjoint cache read/write categories. Otherwise
use canonical day Input/Output/Reasoning/Cache (existing fallback treats Cache as cache read).
Never add day totals to model totals. Any remaining total is UnclassifiedTokens, never invented input.
If component sums exceed canonical totals, preserve the canonical total and mark the entire day's
categories unclassified; do not silently scale bad source data.

Add `PlanReplayRow.PricingStatus string` JSON `pricingStatus`, values `complete`, `partial`, `missing`.
A missing target rate is `missing`; unclassified volume with a known rate is `partial` for PAYG.
Fixed subscription fees remain known, but unknown PAYG comparison cost is not free.
Preserve numeric keys for compatibility; values for incomplete price comparisons are placeholders,
not a savings claim. For partial/missing rows set CostDelta and CostDeltaPercent to zero and
IsRecommended false. Recommendation selection must require complete pricing. Known category
subtotal may remain MonthlyCost for partial rows, with explicit status. Missing rows use zero only
as the existing numeric sentinel. R2 must hide that sentinel behind an unpriced label in text.

- RPL-201: add `TestReplayVolumeConservation`: partial model coverage (100 of 1000) retains total
  1000, peak 1000; complete typed day 800 input/200 output remains priced over all 1000.
  A codewhale-shaped 1000 Unclassified model tokens remains total 1000, unclassified 1000.
  Cover empty maps, activity-only rows, date exclusion and overclassified inconsistent input.
- RPL-202: add `TestReplayMissingPriceNotRecommended`: nil pricer and LookupPrice false never
  recommend a zero-price target or claim 100% savings. Known fully typed priced controls remain
  numerically unchanged. `go test -race ./internal/thermal -run Replay` exits 0.

## R2: Make incomplete estimates readable

Depends on R1. Allowed: `internal/render/replay.go`, `internal/render/replay_test.go`,
`README.md`, plus common companion docs. Preserve unrelated README edits.
Show `Unpriced` for missing rates, `Partial estimate` with the subtotal for partial pricing,
and no numeric savings comparison for either. Show unclassified token coverage in the summary.
Clarify that replay is a scenario comparison; it cannot establish code quality or actual bills.

- RPL-203: add `TestRenderReplayIncompletePricing`, both color-off and JSON fixture expectations;
  missing/partial rows cannot read as `$0.00` complete plans or recommended savings.
  `go test -race ./internal/thermal ./internal/render -run Replay` and common final checks exit 0.

Stop/return/review contract: `tasks/plan.md`. R2 is mandatory before closing this goal.
