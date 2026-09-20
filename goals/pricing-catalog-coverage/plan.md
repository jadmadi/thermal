# Plan: Pricing catalog coverage for tool-native model names

## Execution Steps
- [ ] **Task 1 — Classify the uncovered ids**: for each id in the `No pricing for` list, decide whether it is a missing price, a vendor variant of a catalog model, a provider-qualified path, or a proprietary model that no public catalog will carry. Record the classification in this goal's facts.
- [ ] **Task 2 — Choose the mechanism** with the user: explicit prices, an alias map, or documentation of the user-override workflow. Task 1's classification decides what is even possible for each id.
- [ ] **Task 3 — Implement the chosen mechanism** in `internal/pricing/`, with the alias map testable and case-sensitive, so two distinct models can never collapse into one price.
- [ ] **Task 4 — Verify against real data**: the `No pricing for` line shrinks, the estimated total moves toward the truth, and a test pins the lookup behaviour.
- [ ] **Task 5 — Documentation**: the guide's cost section lists the workflow for a user to add their own price, and AGENTS.md records the rule for adding an alias.
- [ ] **Task 6 — Verification and commit**: `go test -race ./...`, build, and a hand check. Commit with `Goal-Ref: pricing-catalog-coverage` and open a pull request into `main`. Release only with the user's approval.
