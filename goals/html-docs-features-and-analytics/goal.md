# Goal: GitHub Pages Analytics Suite, Charting & CLI Feature Parity

## Goal Description
Update GitHub Pages documentation (docs/pages/index.html, docs/pages/404.html) to add dedicated showcase sections and interactive examples for analytics commands (mix, stats, trend), plain-text bar charts (--chart), complete flags matrix, and synchronize softwareVersion metadata to v0.13.0.

## Dependencies & Execution Order
- **Mode**: Dependent (requires prerequisite goals)
- **Depends On**: html-docs-licensing-parity
- **Sequence**: 2
- **Shape**: ship

## Files to Touch
- docs/pages/index.html
- docs/pages/404.html

## References
- **Shared Understanding & Fact Sheet**: [`goals/html-docs-features-and-analytics/facts.md`](facts.md)
- **Execution Plan**: [`goals/html-docs-features-and-analytics/plan.md`](plan.md)

## Done Condition
1. All requirements described in the goal and execution plan are implemented.
2. All automated unit and integration tests pass cleanly with `-race` (`go test -race ./...` or stack equivalent).
3. Zero architectural regressions; definitions of done satisfied.
