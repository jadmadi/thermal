# Implementation Plan: Workload Replay Simulation (`thermal replay`)

## Phase 1: Subscription & Pricing Profiles (`internal/pricing/subscription.go`)
1. Define `SubscriptionPlan` struct:
   - `ID` (e.g. `claude-pro`, `claude-max`, `chatgpt-plus`, `chatgpt-pro`, `cursor-pro`, `deepseek-api`)
   - `Name` (human-readable title)
   - `MonthlyFee` (USD)
   - `DailyTokenCapacity` (estimated token ceiling per day before throttling)
   - `BurstTokenLimit` (max single-session burst ceiling)
   - `Model` (underlying frontier or default model)
   - `IsPayAsYouGo` (boolean)
2. Seed canonical profiles based on observed real-world rate limits.
3. Add helper to look up plans by name or list all available simulation targets.

## Phase 2: Simulation Engine (`internal/thermal/replay.go`)
1. Define `ReplayResult` and `ReplayComparison`:
   - `ActualSpend`: Recorded spend across tools in the selected time window.
   - `TotalTokens`: Disjoint breakdown of tokens replayed.
   - `SimulatedCost`: Projected cost under the target plan or model.
   - `CostDelta`: Savings or excess ($ and %).
   - `ThrottledDays`: Count and list of days where token volume exceeded plan capacity.
   - `ThrottleRate`: Percentage of active days throttled.
   - `CapacityVerdict`: `Pass (0% throttle)`, `Degraded (1-20%)`, `Fail (>20% throttle)`.
2. Implement `SimulateModelReplay`: Multiplies `UncachedInput`, `Output`, `Reasoning`, `CacheRead`, and `CacheWrite` by target model rates from `models.dev`.
3. Implement `SimulateSubscriptionReplay`: Compares daily token distributions against plan capacities.

## Phase 3: Terminal UI & JSON Presentation (`internal/render/replay.go`)
1. Render Executive Replay Table:
   - Columns: Plan / Target Model, Type, Monthly Cost, Cost Delta, Capacity Verdict.
2. Render Actionable Recommendation Block:
   - Highlight the optimal plan based on cost vs. capacity verdict.
3. Support `--json` flag to emit raw simulation metrics.

## Phase 4: CLI Wiring & Tests (`cmd/thermal/main.go`, `*_test.go`)
1. Add `replay` to reserved command verbs.
2. Wire `--against` and `--compare` flags.
3. Write comprehensive unit tests for mathematical correctness and rate-limiting heuristics.
4. Verify complete build and test suite with `-race`.
