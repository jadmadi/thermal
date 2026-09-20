# Goal: Pricing catalog coverage for tool-native model names

## Goal Description
Classify each model id the catalog cannot price, then choose a mechanism with the user: explicit prices, a testable alias map, or documentation of the user-override workflow. Devins swe-1-7 is proprietary and will never be in a public catalog, so a price is the only option for it.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: unpriceable-token-sources
- **Sequence**: 11
- **Shape**: ship

## References
- **Shared Understanding & Fact Sheet**: [`goals/pricing-catalog-coverage/facts.md`](facts.md)
- **Execution Plan**: [`goals/pricing-catalog-coverage/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
