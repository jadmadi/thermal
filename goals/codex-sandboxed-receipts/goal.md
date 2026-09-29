# Goal: Codex Sandboxed Execution Verification Receipts

## Goal Description
Extract local test runner and linter execution commands and exit codes from Codex rollout tool calls (bwrap/sandbox runs) to certify factual work outcomes as Tier 1 Verified without prompt leakage in thermal receipt.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 73
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- internal/loaders/codex.go
- internal/loaders/codex_test.go
- internal/thermal/receipt.go
- internal/thermal/receipt_test.go
- internal/render/receipt.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/codex-sandboxed-receipts/facts.md`](facts.md)
- **Execution Plan**: [`goals/codex-sandboxed-receipts/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
