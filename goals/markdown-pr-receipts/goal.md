# Goal: Zero-Leak Markdown PR Contribution Receipts

## Goal Description
Add --format md (and --format markdown) to thermal receipt and thermal yield to generate sanitized, privacy-safe Markdown badges and summary tables ready to embed in GitHub PR descriptions without leaking proprietary prompts.

## Dependencies & Execution Order
- **Mode**: Independent (disjoint, parallelizable)
- **Depends On**: none
- **Sequence**: 74
- **Shape**: ship
- **Tier**: roadmap

## Files to Touch
- cmd/thermal/main.go
- internal/render/receipt.go
- internal/render/yield.go
- internal/render/receipt_test.go
- internal/render/yield_test.go
- docs/MIGRATION.md

## References
- **Shared Understanding & Fact Sheet**: [`goals/markdown-pr-receipts/facts.md`](facts.md)
- **Execution Plan**: [`goals/markdown-pr-receipts/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
