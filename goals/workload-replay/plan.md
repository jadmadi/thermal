# Implementation Plan: Workload Replay Simulation (`thermal replay`)

- [x] **Phase 1: Subscription & Pricing Profiles (`internal/pricing/subscription.go`)**:
  - `SubscriptionPlan` struct with ID, Name, MonthlyFee, DailyTokenCapacity, BurstTokenLimit, Model, IsPayAsYouGo.
  - Seed canonical profiles based on observed real-world rate limits.
  - Add helper to look up plans by name or list all available simulation targets.

- [x] **Phase 2: Simulation Engine (`internal/thermal/replay.go`)**:
  - `ReplayResult` and `ReplayComparison` definitions.
  - Implement `SimulateModelReplay` with disjoint token pricing from `models.dev`.
  - Implement `SimulateSubscriptionReplay` with capacity checks.

- [x] **Phase 3: Terminal UI & JSON Presentation (`internal/render/replay.go`)**:
  - Render Executive Replay Table.
  - Render Actionable Recommendation Block.
  - Support `--json` flag to emit raw simulation metrics.

- [x] **Phase 4: CLI Wiring & Tests (`cmd/thermal/main.go`, `*_test.go`)**:
  - Add `replay` to reserved command verbs.
  - Wire `--against` and `--compare` flags.
  - Comprehensive unit tests for mathematical correctness and rate-limiting heuristics.
  - Verify complete build and test suite with `-race`.
