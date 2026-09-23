# Goal: Declared Model Identity, Router Alias Normalization & Canonical Catalog

## Goal Description
Implement declared model identity and alias normalization in Go (clean-room). Map noisy router prefixes (openrouter/, deepseek/), vendor-qualified slugs, date-tagged model variants, and bare harness strings into canonical model identifiers without fuzzy matching, ensuring 100% sole copyright ownership and unencumbered dual-licensing.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- internal/loaders/modelname.go
- internal/loaders/modelname_test.go
- internal/pricing/models.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/model-identity-and-alias-resolution/facts.md`](facts.md)
- **Execution Plan**: [`goals/model-identity-and-alias-resolution/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
