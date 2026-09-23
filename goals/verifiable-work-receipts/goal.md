# Goal: Verifiable Work Receipts & Session Outcome Correlation ('thermal receipt')

## Goal Description
Implement verifiable work receipts linking agent token spend to concrete verification outcomes ('thermal receipt'). Analyzes session transcripts for test run exit codes (passes vs failures), linter results, and git commit creations to classify session outcomes into Verified, Failed, or Unverified/Claimed tiers, highlighting spend efficiency across real tasks.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: -
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/thermal/receipt.go
- internal/thermal/receipt_test.go
- internal/render/receipt.go

## References
- **Shared Understanding & Fact Sheet**: [`goals/verifiable-work-receipts/facts.md`](facts.md)
- **Execution Plan**: [`goals/verifiable-work-receipts/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
